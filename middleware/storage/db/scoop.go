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

func (p *Scoop) Table(m string) *Scoop {
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
		p.cond.whereRaw("deleted_at = 0")
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
	defer rows.Close()
	
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
	defer rows.Close()
	
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

func (p *Scoop) Create(value interface{}) *CreateResult {
	p.inc()
	defer p.dec()

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
	var columns []string
	var placeholders []string
	var values []interface{}

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
		if field.Name == "DeletedAt" && fieldValue.Kind() == reflect.Ptr {
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

	insertSQL := "INSERT "
	if p.ignore {
		if p.clientType == MySQL {
			insertSQL += "IGNORE "
		}
	}
	insertSQL += "INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"

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

			if p.clientType == Sqlite {
				queryErr = session.Raw("SELECT last_insert_rowid()").Scan(&lastInsertID).Error
			} else if p.clientType == MySQL {
				queryErr = session.Raw("SELECT LAST_INSERT_ID()").Scan(&lastInsertID).Error
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

func (p *Scoop) CreateInBatches(value interface{}, batchSize int) *CreateInBatchesResult {
	p.inc()
	defer p.dec()

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
		var columns []string
		var allPlaceholders []string
		var allValues []interface{}

		// Build columns from first row
		firstRowInBatch := vv.Index(i)
		for firstRowInBatch.Kind() == reflect.Ptr {
			firstRowInBatch = firstRowInBatch.Elem()
		}

		for _, field := range stmt.Schema.Fields {
			if field.AutoCreateTime != 0 && firstRowInBatch.FieldByName(field.Name).IsZero() {
				// Set auto create time for CreatedAt/UpdatedAt
				if field.DataType == "int64" || field.DataType == "uint64" {
					// Will set for each row below
				}
			}

			// Skip auto increment primary key if it's zero
			if field.AutoIncrement && firstRowInBatch.FieldByName(field.Name).IsZero() {
				continue
			}

			// Skip DeletedAt in column list if it's nil (will add 0 value per row below)
			fieldValue := firstRowInBatch.FieldByName(field.Name)
			if field.Name == "DeletedAt" && fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				// Will include column but use 0 value
			}

			columns = append(columns, field.DBName)
		}

		// Build values for each row in batch
		for j := i; j < end; j++ {
			rowValue := vv.Index(j)
			for rowValue.Kind() == reflect.Ptr {
				rowValue = rowValue.Elem()
			}

			var rowPlaceholders []string
			for _, field := range stmt.Schema.Fields {
				if field.AutoCreateTime != 0 && rowValue.FieldByName(field.Name).IsZero() {
					// Set auto create time for CreatedAt/UpdatedAt
					if field.DataType == "int64" || field.DataType == "uint64" {
						rowValue.FieldByName(field.Name).SetInt(time.Now().Unix())
					}
				}

				// Skip auto increment primary key if it's zero
				if field.AutoIncrement && rowValue.FieldByName(field.Name).IsZero() {
					continue
				}

				fieldValue := rowValue.FieldByName(field.Name)

				// Handle soft delete field - insert 0 for nil *time.Time
				if field.Name == "DeletedAt" && fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
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

		insertSQL := "INSERT "
		if p.ignore {
			if p.clientType == MySQL {
				insertSQL += "IGNORE "
			}
		}
		insertSQL += "INTO " + p.table + " (" + strings.Join(columns, ", ") + ") VALUES " + strings.Join(allPlaceholders, ", ")

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
		updateMap["updated_at"] = time.Now().Unix()
	}

	//if p.hasCreatedAt {
	//	switch p.clientType {
	//	case Sqlite:
	//		updateMap["created_at"] = gorm.Expr("IIF(created_at > 0,created_at,?)", time.Now().Unix())
	//	default:
	//		updateMap["created_at"] = gorm.Expr("IF(created_at > 0,created_at,?)", time.Now().Unix())
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
	var values []interface{}
	for k, v := range updateMap {
		if i > 0 {
			sqlRaw.WriteString(", ")
		}
		sqlRaw.WriteString(quoteFieldName(k))
		sqlRaw.WriteString("=")
		sqlRaw.WriteString("?")
		switch x := v.(type) {
		case clause.Expr:
			values = append(values, x)
		default:
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
	if res.Error != nil && res.Error == gorm.ErrDuplicatedKey {
		log.Errorf("err:%v", res.Error)
		return &UpdateResult{
			RowsAffected: res.RowsAffected,
			Error:        p.getDuplicatedKeyError(),
		}
	}

	return &UpdateResult{
		RowsAffected: res.RowsAffected,
		Error:        res.Error,
	}
}

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
		case "created_at", "updated_at":
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
