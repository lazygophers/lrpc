package db

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/lazygophers/log"
	"gorm.io/gorm"
)

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
//
//	user := &User{Name: "John", Age: 30}
//	result := scoop.Table("users").Create(user)
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
		// 跳过不可创建的字段（如 gorm:"-" 标记的字段）
		if !field.Creatable {
			continue
		}

		// 跳过没有数据库字段名的字段（防止生成空字段名导致 SQL 语法错误）
		if field.DBName == "" {
			continue
		}

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
			// Use GORM's field value method to apply serializers if configured
			var fieldVal interface{}
			if field.Serializer != nil {
				// Call serializer's Value method
				serializedValue, err := field.Serializer.Value(stmt.Context, field, vv, fieldValue.Interface())
				if err != nil {
					log.Errorf("err:%v", err)
					return &CreateResult{
						Error: fmt.Errorf("failed to serialize field %s: %w", field.Name, err),
					}
				}
				fieldVal = serializedValue
			} else {
				fieldVal = fieldValue.Interface()
			}
			values = append(values, fieldVal)
		} else {
			values = append(values, nil)
		}
	}

	// Build INSERT statement with IGNORE support for different databases
	insertSQL := p.buildInsertSQL(columns, "("+strings.Join(placeholders, ", ")+")")

	start := time.Now()

	// Check if we need auto-increment ID
	idInfo := getIdFieldInfo(vv)

	// Build column map for efficient lookup
	columnMap := make(map[string]bool, len(columns))
	for _, col := range columns {
		columnMap[col] = true
	}

	// Determine if we should use RETURNING clause for auto-increment ID (PostgreSQL/GaussDB)
	useReturning := (p.clientType == Postgres || p.clientType == GaussDB) &&
		idInfo.needsAutoIncrement() &&
		!columnMap["id"]

	var lastInsertID int64
	var rowsAffected int64
	var execErr error

	if useReturning {
		// PostgreSQL/GaussDB: Use RETURNING clause to get the ID in one query
		insertSQL += " RETURNING id"
		session := p._db.Session(&gorm.Session{PrepareStmt: false})
		res := session.Raw(insertSQL, values...).Scan(&lastInsertID)
		rowsAffected = res.RowsAffected
		execErr = res.Error

		GetDefaultLogger().Log(p.depth, start, func() (sql string, rowsAffected int64) {
			return FormatSql(insertSQL, values...), res.RowsAffected
		}, res.Error)
	} else {
		// MySQL/TiDB/SQLite/Others: Use ConnPool.ExecContext to get sql.Result
		// This avoids the race condition of querying LAST_INSERT_ID() on a different connection
		session := p._db.Session(&gorm.Session{PrepareStmt: false})

		// Get connection pool and context from GORM
		connPool := session.Statement.ConnPool
		ctx := session.Statement.Context

		// Execute INSERT using ConnPool.ExecContext
		result, err := connPool.ExecContext(ctx, insertSQL, values...)

		GetDefaultLogger().Log(p.depth, start, func() (sql string, affectedRows int64) {
			if result != nil {
				affectedRows, _ = result.RowsAffected()
			}
			return FormatSql(insertSQL, values...), affectedRows
		}, err)

		if err != nil {
			execErr = err
		} else {
			// Get rows affected from sql.Result
			rowsAffected, _ = result.RowsAffected()

			// Get LastInsertId directly from sql.Result (no additional query needed)
			// This is thread-safe because sql.Result contains the ID from the execution
			if idInfo.needsAutoIncrement() && rowsAffected > 0 {
				lastInsertID, err = result.LastInsertId()
				if err != nil {
					log.Errorf("err:%v", err)
				}
			}
		}
	}

	// Handle execution errors
	if execErr != nil {
		log.Errorf("err:%v", execErr)
		if p.IsDuplicatedKeyError(execErr) {
			return &CreateResult{
				RowsAffected: rowsAffected,
				Error:        p.getDuplicatedKeyError(),
			}
		}
		return &CreateResult{
			Error: execErr,
		}
	}

	// Set the auto-generated ID back to the struct if applicable
	if lastInsertID > 0 && idInfo.needsAutoIncrement() {
		idInfo.setValue(lastInsertID)
	}

	return &CreateResult{
		RowsAffected: rowsAffected,
		Error:        nil,
	}
}

