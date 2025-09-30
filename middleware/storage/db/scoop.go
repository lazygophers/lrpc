package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/stringx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Field name constants for common database fields
const (
	fieldDeletedAt = "deleted_at"
	fieldCreatedAt = "created_at"
	fieldUpdatedAt = "updated_at"
	fieldID        = "id"
)

// SQL condition constants
const (
	condNotDeleted = "deleted_at = 0"
)

// Field name constants for Go struct fields
const (
	structFieldDeletedAt = "DeletedAt"
	structFieldCreatedAt = "CreatedAt"
	structFieldUpdatedAt = "UpdatedAt"
)

// gormTagCache caches parsed GORM tag information for struct types
// This significantly improves performance for repeated Updates operations
var (
	gormTagCache = make(map[reflect.Type]*gormTagInfo)
	gormTagMutex sync.RWMutex

	// Cache for field name conversions from snake_case to CamelCase
	fieldNameCache = make(map[string]string)
	fieldNameMutex sync.RWMutex

	// Pool for gorm.Statement objects to reduce allocations
	statementPool = sync.Pool{
		New: func() interface{} {
			return &gorm.Statement{}
		},
	}
)

// gormTagInfo stores parsed information about updatable fields
type gormTagInfo struct {
	updatableFields map[string]string // fieldName -> dbName mapping
}

// parseGormTags parses GORM tags for a struct type and caches the result
// Returns information about which fields are updatable and their DB names
func parseGormTags(t reflect.Type) *gormTagInfo {
	// Fast path: read lock for cache lookup
	gormTagMutex.RLock()
	if info, ok := gormTagCache[t]; ok {
		gormTagMutex.RUnlock()
		return info
	}
	gormTagMutex.RUnlock()

	// Slow path: write lock for cache update
	gormTagMutex.Lock()
	defer gormTagMutex.Unlock()

	// Double-check in case another goroutine already cached it
	if info, ok := gormTagCache[t]; ok {
		return info
	}

	// Parse struct tags
	info := &gormTagInfo{
		updatableFields: make(map[string]string),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		gormTag := field.Tag.Get("gorm")

		// Skip ignored fields
		if gormTag == "-" {
			continue
		}

		// Skip fields that should not be updated
		if isSkippableField(gormTag, field.Name) {
			continue
		}

		// Extract DB column name
		dbName := extractDBName(gormTag, field.Name)
		info.updatableFields[field.Name] = dbName
	}

	gormTagCache[t] = info
	return info
}

// isSkippableField checks if a field should be skipped during updates
func isSkippableField(gormTag, fieldName string) bool {
	// Check for special GORM tags
	if strings.Contains(gormTag, "primaryKey") ||
		strings.Contains(gormTag, "autoCreateTime") ||
		strings.Contains(gormTag, "autoUpdateTime") {
		return true
	}

	// Check for time tracking fields by name
	if fieldName == structFieldCreatedAt || fieldName == structFieldUpdatedAt {
		return true
	}

	return false
}

// extractDBName extracts the database column name from GORM tag
func extractDBName(gormTag, fieldName string) string {
	if gormTag == "" {
		return Camel2UnderScore(fieldName)
	}

	// Parse column name from tag
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}

	return Camel2UnderScore(fieldName)
}

type Scoop struct {
	clientType string
	_db        *gorm.DB

	notFoundError      error
	duplicatedKeyError error

	hasCreatedAt, hasUpdatedAt, hasDeletedAt bool

	hasId bool
	table string

	cond          Cond
	limit, offset uint64
	selects       []string
	groups        []string
	orders        []string
	unscoped      bool

	ignore bool

	depth int
}

func NewScoop(db *gorm.DB, clientType string) *Scoop {
	s := &Scoop{
		depth:      3,
		clientType: clientType,
		_db: db.Session(&gorm.Session{
			//NewDB: true,
			Initialized: true,
		}),
	}
	// Set clientType in cond for proper field quoting
	s.cond.clientType = clientType
	return s
}

func (p *Scoop) getNotFoundError() error {
	if p.notFoundError != nil {
		return p.notFoundError
	}

	return gorm.ErrRecordNotFound
}

func (p *Scoop) getDuplicatedKeyError() error {
	if p.duplicatedKeyError != nil {
		return p.duplicatedKeyError
	}

	return gorm.ErrDuplicatedKey
}

func (p *Scoop) IsNotFound(err error) bool {
	return err == p.getNotFoundError() || err == gorm.ErrRecordNotFound
}

func (p *Scoop) IsDuplicatedKeyError(err error) bool {
	return err == p.getDuplicatedKeyError() || err == gorm.ErrDuplicatedKey
}

func (p *Scoop) AutoMigrate(dst ...interface{}) error {
	return p._db.AutoMigrate(dst...)
}

// isValidTableName validates table name to prevent SQL injection
// Only allows alphanumeric characters, underscores, and dots (for schema.table)
func isValidTableName(tableName string) bool {
	if tableName == "" {
		return false
	}
	for _, char := range tableName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '.') {
			return false
		}
	}
	return true
}

func (p *Scoop) inc() {
	p.depth++
}

func (p *Scoop) dec() {
	p.depth--
}

//func (p *Scoop) Session(config ...*gorm.Session) *Scoop {
//	if len(config) == 0 {
//		return NewScoop(p._db.Session(&gorm.Session{
//			NewDB: true,
//		}))
//	}
//	return NewScoop(p._db.Session(config[0]))
//}

func (p *Scoop) Model(m any) *Scoop {
	rt := reflect.ValueOf(m).Type()
	p.table = getTableName(rt)

	p.hasCreatedAt = hasCreatedAt(rt)
	p.hasUpdatedAt = hasUpdatedAt(rt)
	p.hasDeletedAt = hasDeletedAt(rt)

	p.hasId = hasId(rt)

	return p
}

// Table sets the table name for the query
// Validates table name to prevent SQL injection
func (p *Scoop) Table(m string) *Scoop {
	if !isValidTableName(m) {
		log.Warnf("invalid table name: %s, table name must contain only alphanumeric characters, underscores and dots", m)
	}
	p.table = m
	return p
}

// ——————————条件——————————

func (p *Scoop) Select(fields ...string) *Scoop {
	p.selects = append(p.selects, fields...)
	return p
}

func (p *Scoop) Where(args ...interface{}) *Scoop {
	p.cond.Where(args...)
	return p
}

func (p *Scoop) Equal(column string, value interface{}) *Scoop {
	p.cond.where(column, value)
	return p
}

func (p *Scoop) NotEqual(column string, value interface{}) *Scoop {
	p.cond.where(column, "!= ", value)
	return p
}

func (p *Scoop) In(column string, values interface{}) *Scoop {
	vo := EnsureIsSliceOrArray(values)
	if vo.Len() == 0 {
		p.cond.where(false)
		return p
	}
	p.cond.where(column, "IN", UniqueSlice(vo.Interface()))
	return p
}

func (p *Scoop) NotIn(column string, values interface{}) *Scoop {
	vo := EnsureIsSliceOrArray(values)
	if vo.Len() == 0 {
		return p
	}
	p.cond.where(column, "NOT IN", UniqueSlice(vo.Interface()))
	return p
}

func (p *Scoop) Like(column string, value string) *Scoop {
	p.cond.Like(column, value)
	return p
}

func (p *Scoop) LeftLike(column string, value string) *Scoop {
	p.cond.where(column, "LIKE", "%"+value)
	return p
}

func (p *Scoop) RightLike(column string, value string) *Scoop {
	p.cond.where(column, "LIKE", value+"%")
	return p
}

func (p *Scoop) NotLike(column string, value string) *Scoop {
	p.cond.where(column, "NOT LIKE", "%"+value+"%")
	return p
}

func (p *Scoop) NotLeftLike(column string, value string) *Scoop {
	p.cond.where(column, "NOT LIKE", "%"+value)
	return p
}

func (p *Scoop) NotRightLike(column string, value string) *Scoop {
	p.cond.where(column, "NOT LIKE", value+"%")
	return p
}

func (p *Scoop) Between(column string, min, max interface{}) *Scoop {
	p.cond.whereRaw(quoteFieldName(column, p.clientType)+" BETWEEN ? AND ?", min, max)
	return p
}

func (p *Scoop) NotBetween(column string, min, max interface{}) *Scoop {
	p.cond.whereRaw(quoteFieldName(column, p.clientType)+" NOT BETWEEN ? AND ?", min, max)
	return p
}

func (p *Scoop) Unscoped(b ...bool) *Scoop {
	if len(b) == 0 {
		p.unscoped = true
		return p
	}
	p.unscoped = b[0]
	return p
}

func (p *Scoop) Limit(limit uint64) *Scoop {
	p.limit = limit
	return p
}

func (p *Scoop) Offset(offset uint64) *Scoop {
	p.offset = offset
	return p
}

func (p *Scoop) Group(fields ...string) *Scoop {
	p.groups = append(p.groups, fields...)
	return p
}

func (p *Scoop) Order(fields ...string) *Scoop {
	p.orders = append(p.orders, fields...)
	return p
}

func (p *Scoop) Ignore(b ...bool) *Scoop {
	if len(b) == 0 {
		p.ignore = true
		return p
	}
	p.ignore = b[0]
	return p
}

// ——————————操作——————————

// getStatement gets a Statement from the pool and initializes it
func getStatement(db *gorm.DB, table string, model interface{}) *gorm.Statement {
	stmt := statementPool.Get().(*gorm.Statement)
	stmt.DB = db
	stmt.Table = table
	stmt.Model = model
	if db.Statement != nil {
		stmt.TableExpr = db.Statement.TableExpr
	}
	return stmt
}

// putStatement resets and returns a Statement to the pool
func putStatement(stmt *gorm.Statement) {
	// Reset statement fields to avoid memory leaks
	stmt.DB = nil
	stmt.Table = ""
	stmt.Model = nil
	stmt.Schema = nil
	stmt.TableExpr = nil
	statementPool.Put(stmt)
}

// getCachedFieldName converts snake_case to CamelCase with caching.
// This significantly improves performance when scanning database rows.
func getCachedFieldName(dbName string) string {
	// Fast path: read lock
	fieldNameMutex.RLock()
	if camelName, ok := fieldNameCache[dbName]; ok {
		fieldNameMutex.RUnlock()
		return camelName
	}
	fieldNameMutex.RUnlock()

	// Slow path: convert and cache
	fieldNameMutex.Lock()
	defer fieldNameMutex.Unlock()

	// Double-check after acquiring write lock
	if camelName, ok := fieldNameCache[dbName]; ok {
		return camelName
	}

	camelName := stringx.Snake2Camel(dbName)
	fieldNameCache[dbName] = camelName
	return camelName
}

// handleAutoTimeField sets auto timestamp fields (CreatedAt/UpdatedAt) if they are zero.
// Returns true if the field was auto-set, false otherwise.
func handleAutoTimeField(field *schema.Field, fieldValue reflect.Value) bool {
	if field.AutoCreateTime != 0 && fieldValue.IsZero() {
		if field.DataType == "int64" || field.DataType == "uint64" {
			fieldValue.SetInt(time.Now().Unix())
			return true
		}
	}
	return false
}

// buildInsertSQL constructs an INSERT statement with proper IGNORE/ON CONFLICT handling
// based on the database type. It supports single-row and multi-row insertions.
// columns: list of column names
// placeholders: list of value placeholders like "(?)", or "(?), (?)" for batch inserts
func (p *Scoop) buildInsertSQL(columns []string, placeholders string) string {
	columnsStr := strings.Join(columns, ", ")

	if p.ignore {
		switch p.clientType {
		case MySQL:
			return "INSERT IGNORE INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders
		case Sqlite:
			return "INSERT OR IGNORE INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders
		case Postgres:
			return "INSERT INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders + " ON CONFLICT DO NOTHING"
		default:
			return "INSERT INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders
		}
	}
	return "INSERT INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders
}

// scanRowsInto is a helper function that scans SQL rows into a reflect.Value.
// It handles the common logic for both Find and First operations.
// dest should be a valid reflect.Value that can be set.
// For Find, dest should be a slice element; for First, dest should be a struct pointer.
func scanRowsInto(rows *sql.Rows, dest reflect.Value, sqlRaw string, start time.Time, depth int) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	err = rows.Scan(scanArgs...)
	if err != nil {
		return err
	}

	for i, col := range values {
		if col == nil {
			continue
		}
		fieldName := getCachedFieldName(cols[i])
		field := dest.FieldByName(fieldName)
		if !field.IsValid() {
			log.Debugf("invalid field: %s", fieldName)
			continue
		}
		err = decode(field, col)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Scoop) findSql() string {
	b := log.GetBuffer()
	defer log.PutBuffer(b)

	b.WriteString("SELECT ")
	if len(p.selects) > 0 {
		b.WriteString(p.selects[0])

		for _, s := range p.selects[1:] {
			b.WriteString(", ")
			b.WriteString(s)
		}
	} else {
		b.WriteString("*")
	}

	b.WriteString(" FROM ")
	b.WriteString(p.table)

	if len(p.cond.conds) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			b.WriteString(" AND ")
			b.WriteString(c)
		}
	}

	if len(p.groups) > 0 {
		b.WriteString(" GROUP BY ")
		b.WriteString(p.groups[0])
		for _, g := range p.groups[1:] {
			b.WriteString(", ")
			b.WriteString(g)
		}
	}

	if len(p.orders) > 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(p.orders[0])
		for _, o := range p.orders[1:] {
			b.WriteString(", ")
			b.WriteString(o)
		}
	}

	if p.limit > 0 {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.FormatUint(p.limit, 10))
	}

	if p.offset > 0 {
		b.WriteString(" OFFSET ")
		b.WriteString(strconv.FormatUint(p.offset, 10))
	}

	return b.String()
}

type FindResult struct {
	RowsAffected int64
	Error        error
}

// Find executes a SELECT query and scans all matching rows into out.
// The out parameter must be a pointer to a slice of structs.
// Returns FindResult containing any error that occurred.
//
// Example:
//   var users []User
//   result := scoop.Where("age > ?", 18).Find(&users)
func (p *Scoop) Find(out interface{}) *FindResult {
	if p.cond.skip {
		return &FindResult{}
	}

	vv := reflect.ValueOf(out)
	if vv.Type().Kind() != reflect.Ptr {
		return &FindResult{
			Error: fmt.Errorf("Find failed: out parameter must be a pointer, got %v", vv.Type().Kind()),
		}
	}
	vv = vv.Elem()
	if vv.Type().Kind() != reflect.Slice {
		return &FindResult{
			Error: fmt.Errorf("Find failed: out parameter must be a pointer to slice, got pointer to %v", vv.Type().Kind()),
		}
	}

	elem := vv.Type().Elem()

	if p.table == "" {
		p.table = getTableName(elem.Elem())
	}

	if !p.unscoped && (p.hasDeletedAt || hasDeletedAt(elem)) {
		p.cond.whereRaw(condNotDeleted)
	}

	p.inc()
	defer p.dec()

	logBuf := log.GetBuffer()
	defer log.PutBuffer(logBuf)

	sqlRaw := p.findSql()
	start := time.Now()

	scope := p._db.Raw(sqlRaw)
	rows, err := scope.Rows()
	if err != nil {
		return &FindResult{
			Error: err,
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Errorf("err:%v", closeErr)
		}
	}()

	if err = rows.Err(); err != nil {
		return &FindResult{
			Error: err,
		}
	}

	var rawsAffected int64
	// 把数据写回到out
	for rows.Next() {
		rawsAffected++

		var v reflect.Value
		if elem.Elem().Kind() == reflect.Ptr {
			v = reflect.New(elem.Elem().Elem())
		} else {
			v = reflect.New(elem.Elem())
		}

		err = scanRowsInto(rows, v.Elem(), sqlRaw, start, p.depth)
		if err != nil {
			GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
				return sqlRaw, rawsAffected
			}, err)
			return &FindResult{
				Error: err,
			}
		}

		vv.Set(reflect.Append(vv, v))
	}

	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw, rawsAffected
	}, nil)
	return &FindResult{
		RowsAffected: rawsAffected,
	}
}

type ChunkResult struct {
	Error error
}

func (p *Scoop) Chunk(dest interface{}, size uint64, fc func(tx *Scoop, offset uint64) error) *ChunkResult {
	p.offset = 0
	p.limit = size

	vv := reflect.ValueOf(dest)
	if vv.Type().Kind() != reflect.Ptr {
		return &ChunkResult{
			Error: fmt.Errorf("Chunk failed: dest parameter must be a pointer, got %v", vv.Type().Kind()),
		}
	}

	vv = vv.Elem()
	if vv.Type().Kind() != reflect.Slice {
		return &ChunkResult{
			Error: fmt.Errorf("Chunk failed: dest parameter must be a pointer to slice, got pointer to %v", vv.Type().Kind()),
		}
	}

	elem := vv.Type().Elem().Elem()

	if p.table == "" {
		p.table = getTableName(elem)
	}

	p.hasDeletedAt = hasDeletedAt(elem)

	p.inc()
	defer p.dec()

	for {
		// 重置dest内容
		vv.Set(reflect.MakeSlice(vv.Type(), 0, int(size)))

		res := p.Find(dest)
		if res.Error != nil {
			return &ChunkResult{
				Error: res.Error,
			}
		}

		if res.RowsAffected == 0 {
			break
		}

		err := fc(p, p.offset)
		if err != nil {
			return &ChunkResult{
				Error: err,
			}
		}

		p.offset += size
	}

	return &ChunkResult{}
}

type FirstResult struct {
	Error error
}

// First executes a SELECT query with LIMIT 1 and scans the first row into out.
// The out parameter must be a pointer to a struct.
// Returns FirstResult with Error set to ErrRecordNotFound if no row is found.
//
// Example:
//   var user User
//   result := scoop.Table("users").Where("id = ?", 1).First(&user)
func (p *Scoop) First(out interface{}) *FirstResult {
	if p.cond.skip {
		return &FirstResult{
			Error: p.getNotFoundError(),
		}
	}

	vv := reflect.ValueOf(out)
	if vv.Type().Kind() != reflect.Ptr {
		return &FirstResult{
			Error: fmt.Errorf("First failed: out parameter must be a pointer, got %v", vv.Type().Kind()),
		}
	}

	if p.table == "" {
		p.table = getTableName(vv.Type())
	}

	if !p.unscoped && (p.hasDeletedAt || hasDeletedAt(vv.Type())) {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.offset = 0
	p.limit = 1

	p.inc()
	defer p.dec()

	sqlRaw := p.findSql()
	start := time.Now()

	scope := p._db.Raw(sqlRaw)
	rows, err := scope.Rows()
	if err != nil {
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, -1
		}, err)
		return &FirstResult{
			Error: err,
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Errorf("err:%v", closeErr)
		}
	}()

	if err = rows.Err(); err != nil {
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, -1
		}, err)
		return &FirstResult{
			Error: err,
		}
	}

	// 把数据写回到out
	var rowAffected int64
	for rows.Next() {
		rowAffected++

		if rowAffected != 1 {
			continue
		}

		err = scanRowsInto(rows, vv.Elem(), sqlRaw, start, p.depth)
		if err != nil {
			GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
				return sqlRaw, 1
			}, err)
			return &FirstResult{
				Error: err,
			}
		}
	}

	if rowAffected == 0 {
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, 0
		}, p.getNotFoundError())
		return &FirstResult{
			Error: p.getNotFoundError(),
		}
	}

	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw, rowAffected
	}, nil)
	return &FirstResult{}
}

type CreateResult struct {
	RowsAffected int64
	Error        error
}

// Create inserts a single record into the database using raw SQL.
// The value parameter must be a pointer to a struct.
// Auto-increment primary keys and timestamp fields are handled automatically.
// Supports soft delete by inserting 0 for nil DeletedAt fields.
// Returns CreateResult with LastInsertId and RowsAffected.
//
// Example:
//   user := &User{Name: "John", Age: 30}
//   result := scoop.Table("users").Create(user)
func (p *Scoop) Create(value interface{}) *CreateResult {
	p.inc()
	defer p.dec()

	// Check for nil database connection
	if p._db == nil {
		return &CreateResult{
			Error: fmt.Errorf("Create failed: database connection is nil"),
		}
	}

	vv := reflect.ValueOf(value)
	for vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	if vv.Kind() != reflect.Struct {
		return &CreateResult{
			Error: fmt.Errorf("Create failed: value must be a struct, got %v", vv.Kind()),
		}
	}

	elem := vv.Type()
	if p.table == "" {
		p.table = getTableName(elem)
		if p.table == "" {
			return &CreateResult{
				Error: fmt.Errorf("Create failed: unable to determine table name for type %v", elem),
			}
		}
	}

	// Parse struct to get fields and values
	stmt := getStatement(p._db, p.table, value)
	defer putStatement(stmt)

	err := stmt.ParseWithSpecialTableName(value, stmt.Table)
	if err != nil {
		log.Errorf("err:%v", err)
		return &CreateResult{
			Error: err,
		}
	}

	// Build INSERT SQL
	// Pre-allocate slices with estimated capacity based on field count
	fieldCount := len(stmt.Schema.Fields)
	columns := make([]string, 0, fieldCount)
	placeholders := make([]string, 0, fieldCount)
	values := make([]interface{}, 0, fieldCount)

	for _, field := range stmt.Schema.Fields {
		fieldValue := vv.FieldByName(field.Name)

		// Set auto create time for CreatedAt/UpdatedAt if needed
		handleAutoTimeField(field, fieldValue)

		// Skip auto increment primary key if it's zero
		if field.AutoIncrement && fieldValue.IsZero() {
			continue
		}

		// Handle soft delete field - insert 0 for nil *time.Time
		if field.Name == structFieldDeletedAt && fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				columns = append(columns, field.DBName)
				placeholders = append(placeholders, "?")
				values = append(values, 0) // 0 means not deleted in unix timestamp format
				continue
			}
		}

		columns = append(columns, field.DBName)
		placeholders = append(placeholders, "?")

		if fieldValue.IsValid() {
			values = append(values, fieldValue.Interface())
		} else {
			values = append(values, nil)
		}
	}

	// Build INSERT statement with IGNORE support for different databases
	insertSQL := p.buildInsertSQL(columns, "("+strings.Join(placeholders, ", ")+")")

	start := time.Now()
	// Use Session with PrepareStmt disabled for raw SQL
	session := p._db.Session(&gorm.Session{
		PrepareStmt: false,
	})

	// For PostgreSQL with auto-increment ID, use RETURNING clause for better performance
	var lastInsertID int64
	var res *gorm.DB
	if p.clientType == Postgres {
		if field := vv.FieldByName("Id"); field.IsValid() && field.CanSet() && field.Kind() == reflect.Int && field.Int() == 0 {
			// Check if id column was included in the insert
			hasIdColumn := false
			for _, col := range columns {
				if col == "id" {
					hasIdColumn = true
					break
				}
			}
			if !hasIdColumn {
				// Use RETURNING clause to get the ID in one query
				insertSQL += " RETURNING id"
				res = session.Raw(insertSQL, values...).Scan(&lastInsertID)
			} else {
				res = session.Exec(insertSQL, values...)
			}
		} else {
			res = session.Exec(insertSQL, values...)
		}
	} else {
		res = session.Exec(insertSQL, values...)
	}

	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return insertSQL, res.RowsAffected
	}, res.Error)

	if res.Error != nil {
		log.Errorf("err:%v", res.Error)
		if p.IsDuplicatedKeyError(res.Error) {
			return &CreateResult{
				RowsAffected: res.RowsAffected,
				Error:        p.getDuplicatedKeyError(),
			}
		}
		return &CreateResult{
			Error: res.Error,
		}
	}

	// Set the auto-generated ID back to the struct if applicable
	if res.RowsAffected > 0 {
		if field := vv.FieldByName("Id"); field.IsValid() && field.CanSet() && field.Kind() == reflect.Int && field.Int() == 0 {
			// For PostgreSQL, lastInsertID was already set via RETURNING clause
			if p.clientType == Postgres && lastInsertID > 0 {
				field.SetInt(lastInsertID)
			} else {
				// For MySQL and SQLite, query the last insert ID
				var queryErr error

				switch p.clientType {
				case Sqlite:
					queryErr = session.Raw("SELECT last_insert_rowid()").Scan(&lastInsertID).Error
				case MySQL:
					queryErr = session.Raw("SELECT LAST_INSERT_ID()").Scan(&lastInsertID).Error
				}

				if queryErr != nil {
					log.Errorf("err:%v", queryErr)
				} else if lastInsertID > 0 {
					field.SetInt(lastInsertID)
				}
			}
		}
	}

	return &CreateResult{
		RowsAffected: res.RowsAffected,
		Error:        nil,
	}
}

type CreateInBatchesResult struct {
	RowsAffected int64
	Error        error
}

// CreateInBatches inserts multiple records in batches using raw SQL.
// The value parameter must be a slice of structs or pointers to structs.
// Records are inserted in batches of the specified batchSize to optimize performance.
// Auto-increment primary keys and timestamp fields are handled automatically.
// Returns CreateInBatchesResult with total RowsAffected across all batches.
//
// Example:
//   users := []User{{Name: "Alice"}, {Name: "Bob"}, {Name: "Charlie"}}
//   result := scoop.Table("users").CreateInBatches(users, 100)
func (p *Scoop) CreateInBatches(value interface{}, batchSize int) *CreateInBatchesResult {
	p.inc()
	defer p.dec()

	// Check for nil database connection
	if p._db == nil {
		return &CreateInBatchesResult{
			Error: fmt.Errorf("CreateInBatches failed: database connection is nil"),
		}
	}

	// value should be a slice
	vv := reflect.ValueOf(value)
	for vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	if vv.Kind() != reflect.Slice {
		return &CreateInBatchesResult{
			Error: fmt.Errorf("CreateInBatches failed: value must be a slice, got %v", vv.Kind()),
		}
	}

	if vv.Len() == 0 {
		return &CreateInBatchesResult{
			RowsAffected: 0,
			Error:        nil,
		}
	}

	elem := vv.Type().Elem()
	for elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	if p.table == "" {
		p.table = getTableName(elem)
		if p.table == "" {
			return &CreateInBatchesResult{
				Error: fmt.Errorf("CreateInBatches failed: unable to determine table name for type %v", elem),
			}
		}
	}

	// Get first element to parse schema
	firstElem := vv.Index(0)
	for firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	stmt := getStatement(p._db, p.table, firstElem.Interface())
	defer putStatement(stmt)

	err := stmt.ParseWithSpecialTableName(firstElem.Interface(), stmt.Table)
	if err != nil {
		log.Errorf("err:%v", err)
		return &CreateInBatchesResult{
			Error: err,
		}
	}

	var totalRowsAffected int64

	// Process in batches
	for i := 0; i < vv.Len(); i += batchSize {
		end := i + batchSize
		if end > vv.Len() {
			end = vv.Len()
		}

		// Build INSERT SQL for this batch
		// Pre-allocate slices with estimated capacity
		batchRowCount := end - i
		fieldCount := len(stmt.Schema.Fields)
		columns := make([]string, 0, fieldCount)
		allPlaceholders := make([]string, 0, batchRowCount)
		allValues := make([]interface{}, 0, batchRowCount*fieldCount)

		// Analyze first row to determine which fields to include
		// This avoids repeated reflection and field checks for each row
		firstRowInBatch := vv.Index(i)
		for firstRowInBatch.Kind() == reflect.Ptr {
			firstRowInBatch = firstRowInBatch.Elem()
		}

		// Build list of fields to include and their metadata
		type fieldInfo struct {
			field           *schema.Field
			isAutoIncrement bool
			isDeletedAt     bool
			isAutoTime      bool
		}
		includedFields := make([]fieldInfo, 0, fieldCount)

		for _, field := range stmt.Schema.Fields {
			// Skip auto increment primary key if it's zero in first row
			if field.AutoIncrement && firstRowInBatch.FieldByName(field.Name).IsZero() {
				continue
			}

			info := fieldInfo{
				field:           field,
				isAutoIncrement: field.AutoIncrement,
			}

			// Check if this is auto time field
			if field.AutoCreateTime != 0 {
				info.isAutoTime = true
			}

			// Check if this is DeletedAt field
			fieldValue := firstRowInBatch.FieldByName(field.Name)
			if field.Name == structFieldDeletedAt && fieldValue.Kind() == reflect.Ptr {
				info.isDeletedAt = true
			}

			includedFields = append(includedFields, info)
			columns = append(columns, field.DBName)
		}

		// Build values for each row in batch using pre-analyzed field list
		for j := i; j < end; j++ {
			rowValue := vv.Index(j)
			for rowValue.Kind() == reflect.Ptr {
				rowValue = rowValue.Elem()
			}

			rowPlaceholders := make([]string, 0, len(columns))
			for _, info := range includedFields {
				fieldValue := rowValue.FieldByName(info.field.Name)

				// Set auto time if field is zero
				if info.isAutoTime {
					handleAutoTimeField(info.field, fieldValue)
				}

				// Handle soft delete field - insert 0 for nil *time.Time
				if info.isDeletedAt && fieldValue.IsNil() {
					rowPlaceholders = append(rowPlaceholders, "?")
					allValues = append(allValues, 0) // 0 means not deleted
					continue
				}

				rowPlaceholders = append(rowPlaceholders, "?")
				if fieldValue.IsValid() {
					allValues = append(allValues, fieldValue.Interface())
				} else {
					allValues = append(allValues, nil)
				}
			}
			allPlaceholders = append(allPlaceholders, "("+strings.Join(rowPlaceholders, ", ")+")")
		}

		// Build INSERT statement with IGNORE support for different databases
		insertSQL := p.buildInsertSQL(columns, strings.Join(allPlaceholders, ", "))

		start := time.Now()
		// Use Session with PrepareStmt disabled for raw SQL
		session := p._db.Session(&gorm.Session{
			PrepareStmt: false,
		})
		res := session.Exec(insertSQL, allValues...)

		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return insertSQL, res.RowsAffected
		}, res.Error)

		if res.Error != nil {
			log.Errorf("err:%v", res.Error)
			if p.IsDuplicatedKeyError(res.Error) {
				return &CreateInBatchesResult{
					RowsAffected: totalRowsAffected,
					Error:        p.getDuplicatedKeyError(),
				}
			}
			return &CreateInBatchesResult{
				RowsAffected: totalRowsAffected,
				Error:        res.Error,
			}
		}

		totalRowsAffected += res.RowsAffected
	}

	return &CreateInBatchesResult{
		RowsAffected: totalRowsAffected,
		Error:        nil,
	}
}

type DeleteResult struct {
	RowsAffected int64
	Error        error
}

// Delete performs a soft or hard delete on matching rows.
// If the table has a DeletedAt field, performs soft delete by setting timestamp.
// Otherwise performs hard delete by removing rows from the database.
// Use Unscoped() to force hard delete even with DeletedAt field.
// Returns DeleteResult with RowsAffected count.
//
// Example:
//   // Soft delete
//   result := scoop.Table("users").Where("id = ?", 1).Delete()
//   // Hard delete
//   result := scoop.Table("users").Unscoped().Where("id = ?", 1).Delete()
func (p *Scoop) Delete() *DeleteResult {
	if p.cond.skip {
		return &DeleteResult{}
	}

	if p.table == "" {
		return &DeleteResult{
			Error: fmt.Errorf("Delete failed: table name is empty, use Table() to specify table name"),
		}
	}

	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.inc()
	defer p.dec()

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	// 软删除
	if !p.unscoped && p.hasDeletedAt {
		sqlRaw.WriteString("UPDATE")
		sqlRaw.WriteString(" ")
		sqlRaw.WriteString(p.table)
		sqlRaw.WriteString(" SET deleted_at = ")
		sqlRaw.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
		p.cond.addCond("deleted_at", "=", 0)
	} else {
		sqlRaw.WriteString("DELETE FROM")
		sqlRaw.WriteString(" ")
		sqlRaw.WriteString(p.table)
	}

	if len(p.cond.conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	start := time.Now()
	res := p._db.Exec(sqlRaw.String())
	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw.String(), res.RowsAffected
	}, res.Error)

	if res.Error != nil {
		log.Errorf("err:%v", res.Error)
	}

	return &DeleteResult{
		RowsAffected: res.RowsAffected,
		Error:        res.Error,
	}
}

type UpdateResult struct {
	RowsAffected int64
	Error        error
}

func (p *Scoop) update(updateMap map[string]interface{}) *UpdateResult {
	if p.cond.skip {
		return &UpdateResult{}
	}
	if len(updateMap) == 0 {
		return &UpdateResult{
			Error: fmt.Errorf("Update failed: updateMap is empty, no values to update"),
		}
	}

	if p.table == "" {
		return &UpdateResult{
			Error: fmt.Errorf("Update failed: table name is empty, use Table() to specify table name"),
		}
	}

	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw("deleted_at = 0")
	}

	if p.hasUpdatedAt {
		updateMap[fieldUpdatedAt] = time.Now().Unix()
	}

	//if p.hasCreatedAt {
	//	switch p.clientType {
	//	case Sqlite:
	//		updateMap[fieldCreatedAt] = gorm.Expr("IIF(created_at > 0,created_at,?)", time.Now().Unix())
	//	default:
	//		updateMap[fieldCreatedAt] = gorm.Expr("IF(created_at > 0,created_at,?)", time.Now().Unix())
	//	}
	//}

	p.inc()
	defer p.dec()

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	sqlRaw.WriteString("UPDATE ")
	sqlRaw.WriteString(p.table)

	sqlRaw.WriteString(" SET ")
	var i int
	// Pre-allocate values slice with capacity based on updateMap size
	values := make([]interface{}, 0, len(updateMap))
	for k, v := range updateMap {
		if i > 0 {
			sqlRaw.WriteString(", ")
		}
		sqlRaw.WriteString(quoteFieldName(k, p.clientType))
		sqlRaw.WriteString("=")

		// Handle clause.Expr with proper type assertion
		switch x := v.(type) {
		case clause.Expr:
			// For expressions, use the SQL directly instead of placeholder
			sqlRaw.WriteString(x.SQL)
			// Append expression variables to values
			values = append(values, x.Vars...)
		default:
			// For normal values, use placeholder
			sqlRaw.WriteString("?")
			values = append(values, candy.ToString(x))
		}
		i++
	}

	if len(p.cond.conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	start := time.Now()
	res := p._db.Exec(sqlRaw.String(), values...)
	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return FormatSql(sqlRaw.String(), values...), res.RowsAffected
	}, res.Error)

	err := res.Error
	if err != nil {
		log.Errorf("err:%v", err)
		if err == gorm.ErrDuplicatedKey {
			return &UpdateResult{
				RowsAffected: res.RowsAffected,
				Error:        p.getDuplicatedKeyError(),
			}
		}
		return &UpdateResult{
			RowsAffected: res.RowsAffected,
			Error:        err,
		}
	}

	return &UpdateResult{
		RowsAffected: res.RowsAffected,
		Error:        nil,
	}
}

// Updates performs an UPDATE operation on matching rows.
// The m parameter can be either a map[string]interface{} or a struct.
// Only non-zero fields are updated. Use clause.Expr for SQL expressions.
// Automatically updates UpdatedAt field if present.
// Returns UpdateResult with RowsAffected count.
//
// Example:
//   // Using map
//   result := scoop.Table("users").Where("id = ?", 1).Updates(map[string]interface{}{"age": 25})
//   // Using struct
//   result := scoop.Table("users").Where("id = ?", 1).Updates(&User{Age: 25})
func (p *Scoop) Updates(m interface{}) *UpdateResult {
	if p.cond.skip {
		return &UpdateResult{}
	}

	p.inc()
	defer p.dec()

	if v, ok := m.(map[string]interface{}); ok {
		return p.update(v)
	}
	mVal := reflect.ValueOf(m)
	if mVal.Type().Kind() == reflect.Ptr {
		mVal = mVal.Elem()
	}
	mType := mVal.Type()
	if mType.Kind() != reflect.Struct {
		return &UpdateResult{
			Error: fmt.Errorf("Updates failed: expected struct, got %v", mType.Kind()),
		}
	}

	// Use cached tag information for better performance
	tagInfo := parseGormTags(mType)
	valMap := make(map[string]interface{}, len(tagInfo.updatableFields))

	for fieldName, dbName := range tagInfo.updatableFields {
		fieldVal := mVal.FieldByName(fieldName)

		if !fieldVal.IsValid() || !fieldVal.CanInterface() || fieldVal.IsZero() {
			continue
		}

		valMap[dbName] = fieldVal.Interface()
	}

	if len(valMap) == 0 {
		return &UpdateResult{
			Error: fmt.Errorf("Updates failed: no non-zero fields found in struct %v to update", mType),
		}
	}
	return p.update(valMap)
}

// Count returns the number of rows matching the query conditions.
// Respects WHERE clauses and soft delete (DeletedAt) filtering.
// Returns the count and any error that occurred.
//
// Example:
//   count, err := scoop.Table("users").Where("age > ?", 18).Count()
func (p *Scoop) Count() (uint64, error) {
	if p.cond.skip {
		return 0, nil
	}

	if p.table == "" {
		return 0, fmt.Errorf("Count failed: table name is empty, use Table() to specify table name")
	}

	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.inc()
	defer p.dec()

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	sqlRaw.WriteString("SELECT COUNT(*) FROM ")
	sqlRaw.WriteString(p.table)

	if len(p.cond.conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	start := time.Now()
	var count uint64
	err := p._db.Raw(sqlRaw.String()).Scan(&count).Error
	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw.String(), int64(count)
	}, err)

	return count, err
}

func (p *Scoop) Exist() (bool, error) {
	if p.cond.skip {
		return false, nil
	}

	if p.table == "" {
		return false, fmt.Errorf("Exists failed: table name is empty, use Table() to specify table name")
	}

	if !p.unscoped && p.hasDeletedAt {
		p.cond.whereRaw("deleted_at = 0")
	}

	p.limit = 1
	p.offset = 0

	p.inc()
	defer p.dec()

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	sqlRaw.WriteString("SELECT ")
	if p.hasId {
		sqlRaw.WriteString("id")
	} else {
		sqlRaw.WriteString("count(*)")
	}

	sqlRaw.WriteString(" FROM ")
	sqlRaw.WriteString(p.table)

	if len(p.cond.conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(p.cond.conds[0])
		for _, c := range p.cond.conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	sqlRaw.WriteString(" LIMIT 1 OFFSET 0")

	start := time.Now()
	var count int64
	err := p._db.Raw(sqlRaw.String()).Scan(&count).Error
	GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
		return sqlRaw.String(), 0
	}, err)
	return count > 0, err
}

func (p *Scoop) FindByPage(opt *core.ListOption, values any) (*core.Paginate, error) {
	p.Offset(opt.Offset).Limit(opt.Limit)

	page := &core.Paginate{
		Offset: opt.Offset,
		Limit:  opt.Limit,
	}

	p.inc()
	defer p.dec()

	err := p.Find(values).Error
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if opt.ShowTotal {
		page.Total, err = p.Count()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}
	}

	return page, nil
}

// ——————————事务——————————

func (p *Scoop) Begin() *Scoop {
	return NewScoop(p._db.Begin(), p.clientType)
}

func (p *Scoop) Rollback() *Scoop {
	p._db.Rollback()
	return p
}

func (p *Scoop) Commit() *Scoop {
	p._db.Commit()
	return p
}

func (p *Scoop) CommitOrRollback(tx *Scoop, logic func(tx *Scoop) error) error {
	err := logic(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}
