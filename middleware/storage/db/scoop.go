package db

import (
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"
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
	return &Scoop{
		depth:      3,
		clientType: clientType,
		_db: db.Session(&gorm.Session{
			//NewDB: true,
			Initialized: true,
		}),
	}
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
	p.cond.whereRaw(quoteFieldName(column)+" BETWEEN ? AND ?", min, max)
	return p
}

func (p *Scoop) NotBetween(column string, min, max interface{}) *Scoop {
	p.cond.whereRaw(quoteFieldName(column)+" NOT BETWEEN ? AND ?", min, max)
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
		panic("invalid out type, not ptr")
	}
	vv = vv.Elem()
	if vv.Type().Kind() != reflect.Slice {
		panic("invalid out type, not slice")
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

	cols, err := rows.Columns()
	if err != nil {
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, -1
		}, err)
		return &FindResult{
			Error: err,
		}
	}

	values := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var rawsAffected int64
	// 把数据写回到out
	for rows.Next() {
		rawsAffected++

		err = rows.Scan(scanArgs...)
		if err != nil {
			GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
				return sqlRaw, rawsAffected
			}, err)
			return &FindResult{
				Error: err,
			}
		}
		var v reflect.Value
		if elem.Elem().Kind() == reflect.Ptr {
			v = reflect.New(elem.Elem().Elem())
		} else {
			v = reflect.New(elem.Elem())
		}

		for i, col := range values {
			if col == nil {
				continue
			}
			field := v.Elem().FieldByName(stringx.Snake2Camel(cols[i]))
			if !field.IsValid() {
				log.Warnf("invalid field: %s", stringx.Snake2Camel(cols[i]))
				continue
			}

			err = decode(field, col)
			if err != nil {
				GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
					return sqlRaw, rawsAffected
				}, err)
				return &FindResult{
					Error: err,
				}
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
		panic("invalid out type, not ptr")
	}

	vv = vv.Elem()
	if vv.Type().Kind() != reflect.Slice {
		panic("invalid out type, not slice")
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
		panic("invalid out type, not ptr")
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

	cols, err := rows.Columns()
	if err != nil {
		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return sqlRaw, -1
		}, err)
		return &FirstResult{
			Error: err,
		}
	}

	values := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// 把数据写回到out
	var rowAffected int64
	for rows.Next() {
		rowAffected++
		err = rows.Scan(scanArgs...)
		if err != nil {
			GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
				return sqlRaw, 1
			}, err)
			return &FirstResult{
				Error: err,
			}
		}

		if rowAffected != 1 {
			continue
		}

		for i, col := range values {
			if col == nil {
				continue
			}
			field := vv.Elem().FieldByName(stringx.Snake2Camel(cols[i]))
			if !field.IsValid() {
				log.Debugf("invalid field: %s", stringx.Snake2Camel(cols[i]))
				continue
			}
			err = decode(field, col)
			if err != nil {
				GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
					return sqlRaw, 1
				}, err)
				return &FirstResult{
					Error: err,
				}
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
			Error: errors.New("database connection is nil"),
		}
	}

	vv := reflect.ValueOf(value)
	for vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	if vv.Kind() != reflect.Struct {
		panic("value is not struct")
	}

	elem := vv.Type()
	if p.table == "" {
		p.table = getTableName(elem)
	}

	// Parse struct to get fields and values
	stmt := &gorm.Statement{
		DB:    p._db,
		Table: p.table,
		Model: value,
	}
	if p._db.Statement != nil {
		stmt.TableExpr = p._db.Statement.TableExpr
	}

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
		if field.AutoCreateTime != 0 && vv.FieldByName(field.Name).IsZero() {
			// Set auto create time for CreatedAt/UpdatedAt
			if field.DataType == "int64" || field.DataType == "uint64" {
				vv.FieldByName(field.Name).SetInt(time.Now().Unix())
			}
		}

		// Skip auto increment primary key if it's zero
		if field.AutoIncrement && vv.FieldByName(field.Name).IsZero() {
			continue
		}

		fieldValue := vv.FieldByName(field.Name)

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
	var insertSQL string
	if p.ignore {
		switch p.clientType {
		case MySQL:
			insertSQL = "INSERT IGNORE INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
		case Sqlite:
			insertSQL = "INSERT OR IGNORE INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
		case Postgres:
			insertSQL = "INSERT INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ") ON CONFLICT DO NOTHING"
		default:
			insertSQL = "INSERT INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
		}
	} else {
		insertSQL = "INSERT INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	}

	start := time.Now()
	// Use Session with PrepareStmt disabled for raw SQL
	session := p._db.Session(&gorm.Session{
		PrepareStmt: false,
	})
	res := session.Exec(insertSQL, values...)

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
			// Get the last insert ID using a single connection query
			// This ensures we get the ID from the same connection that performed the INSERT
			var lastInsertID int64
			var queryErr error

			switch p.clientType {
			case Sqlite:
				queryErr = session.Raw("SELECT last_insert_rowid()").Scan(&lastInsertID).Error
			case MySQL:
				queryErr = session.Raw("SELECT LAST_INSERT_ID()").Scan(&lastInsertID).Error
			case Postgres:
				// For Postgres, query the current value of the sequence
				// Assumes the sequence name follows the pattern: tablename_id_seq
				sequenceName := p.table + "_id_seq"
				queryErr = session.Raw("SELECT currval(?)", sequenceName).Scan(&lastInsertID).Error
			}

			if queryErr != nil {
				log.Errorf("err:%v", queryErr)
			} else if lastInsertID > 0 {
				field.SetInt(lastInsertID)
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
			Error: errors.New("database connection is nil"),
		}
	}

	// value should be a slice
	vv := reflect.ValueOf(value)
	for vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	if vv.Kind() != reflect.Slice {
		panic("value is not slice")
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
	}

	// Get first element to parse schema
	firstElem := vv.Index(0)
	for firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	stmt := &gorm.Statement{
		DB:    p._db,
		Table: p.table,
		Model: firstElem.Interface(),
	}
	if p._db.Statement != nil {
		stmt.TableExpr = p._db.Statement.TableExpr
	}

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
				if info.isAutoTime && fieldValue.IsZero() {
					if info.field.DataType == "int64" || info.field.DataType == "uint64" {
						fieldValue.SetInt(time.Now().Unix())
					}
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
		var insertSQL string
		if p.ignore {
			switch p.clientType {
			case MySQL:
				insertSQL = "INSERT IGNORE INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES " + strings.Join(allPlaceholders, ", ")
			case Sqlite:
				insertSQL = "INSERT OR IGNORE INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES " + strings.Join(allPlaceholders, ", ")
			case Postgres:
				insertSQL = "INSERT INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES " + strings.Join(allPlaceholders, ", ") + " ON CONFLICT DO NOTHING"
			default:
				insertSQL = "INSERT INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES " + strings.Join(allPlaceholders, ", ")
			}
		} else {
			insertSQL = "INSERT INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES " + strings.Join(allPlaceholders, ", ")
		}

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
		panic("table name is empty")
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
			Error: errors.New("updateMap is empty"),
		}
	}

	if p.table == "" {
		panic("table name is empty")
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
		sqlRaw.WriteString(quoteFieldName(k))
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
			Error: errors.New("m must be map or struct"),
		}
	}
	fieldNum := mType.NumField()
	valMap := make(map[string]interface{})
	for i := 0; i < fieldNum; i++ {
		fieldType := mType.Field(i)
		fieldVal := mVal.Field(i)

		if !fieldVal.IsValid() {
			continue
		}

		if !fieldVal.CanInterface() {
			continue
		}

		if fieldVal.IsZero() {
			continue
		}

		// 判断一下 gorm tags 的配置
		// TODO 添加解析的缓存
		gormTag := fieldType.Tag.Get("gorm")
		if gormTag == "-" {
			continue
		}

		var fieldName string
		if gormTag != "" {
			// 判断是否为主键
			if gormTag == "primaryKey" {
				continue
			}

			if strings.HasPrefix(gormTag, "primaryKey;") {
				continue
			}

			if strings.Contains(gormTag, ";primaryKey") {
				continue
			}

			// 判断是否是自动更新的字段
			if strings.HasPrefix(gormTag, "autoCreateTime") {
				continue
			}

			if strings.Contains(gormTag, ";autoUpdateTime") {
				continue
			}

			if strings.HasPrefix(gormTag, "autoUpdateTime") {
				continue
			}

			if strings.Contains(gormTag, ";autoUpdateTime") {
				continue
			}

			// 获取字段名
			idx := strings.Index(gormTag, "column:")
			if idx > 0 {
				tag := gormTag[idx+7:]
				idx = strings.Index(tag, ";")
				if idx > 0 {
					fieldName = tag[:idx]
				} else {
					fieldName = tag
				}
			}
		}

		if fieldName == "" {
			fieldName = Camel2UnderScore(fieldType.Name)
		}

		switch fieldName {
		case fieldCreatedAt, fieldUpdatedAt:
			continue
		}

		valMap[fieldName] = fieldVal.Interface()
	}
	if len(valMap) == 0 {
		return &UpdateResult{
			Error: errors.New("no field need to update"),
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
		panic("table name is empty")
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
		panic("table name is empty")
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
