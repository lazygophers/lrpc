package db

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/lazygophers/log"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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
//
//	users := []User{{Name: "Alice"}, {Name: "Bob"}, {Name: "Charlie"}}
//	result := scoop.Table("users").CreateInBatches(users, 100)
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
			// 跳过不可创建的字段（如 gorm:"-" 标记的字段）
			if !field.Creatable {
				continue
			}

			// 跳过没有数据库字段名的字段（防止生成空字段名导致 SQL 语法错误）
			if field.DBName == "" {
				continue
			}

			// Skip auto increment primary key if it's zero in first row
			if field.AutoIncrement && firstRowInBatch.FieldByName(field.Name).IsZero() {
				continue
			}

			info := fieldInfo{
				field:           field,
				isAutoIncrement: field.AutoIncrement,
			}

			// Check if this is auto time field
			if field.AutoCreateTime != 0 || field.AutoUpdateTime != 0 {
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
					// Use GORM's field value method to apply serializers if configured
					var fieldVal interface{}
					if info.field.Serializer != nil {
						serializedValue, serErr := info.field.Serializer.Value(stmt.Context, info.field, rowValue, fieldValue.Interface())
						if serErr != nil {
							log.Errorf("err:%v", serErr)
							return &CreateInBatchesResult{
								Error: fmt.Errorf("CreateInBatches failed: failed to serialize field %s: %w", info.field.Name, serErr),
							}
						}
						fieldVal = serializedValue
					} else {
						fieldVal = fieldValue.Interface()
					}
					allValues = append(allValues, fieldVal)
				} else {
					allValues = append(allValues, nil)
				}
			}
			allPlaceholders = append(allPlaceholders, "("+strings.Join(rowPlaceholders, ", ")+")")
		}

		// Build INSERT statement with IGNORE support for different databases
		insertSQL := p.buildInsertSQL(columns, strings.Join(allPlaceholders, ", "))

		start := time.Now()

		// Get session and connection pool
		session := p._db.Session(&gorm.Session{
			PrepareStmt: false,
		})
		connPool := session.Statement.ConnPool
		ctx := session.Statement.Context

		var batchRowsAffected int64
		var insertedIDs []int64
		var execErr error

		// Check if we need to retrieve auto-increment IDs
		firstRowInCurrentBatch := vv.Index(i)
		for firstRowInCurrentBatch.Kind() == reflect.Ptr {
			firstRowInCurrentBatch = firstRowInCurrentBatch.Elem()
		}
		needsIDRetrieval := firstRowInCurrentBatch.FieldByName("Id").IsValid()

		// Use different strategies based on database type for optimal performance
		if needsIDRetrieval && (p.clientType == Postgres || p.clientType == GaussDB) {
			// PostgreSQL/GaussDB: Use RETURNING clause to get all inserted IDs in one query
			// This completely avoids race conditions and is the most efficient approach
			insertSQL += " RETURNING id"

			rows, err := connPool.QueryContext(ctx, insertSQL, allValues...)
			if err != nil {
				execErr = err
				GetDefaultLogger().Log(p.depth, start, func() (sql string, affectedRows int64) {
					return FormatSql(insertSQL, allValues...), 0
				}, err)
			} else {
				defer rows.Close()

				// Collect all returned IDs
				for rows.Next() {
					var id int64
					if err := rows.Scan(&id); err != nil {
						log.Errorf("err:%v", err)
						continue
					}
					insertedIDs = append(insertedIDs, id)
				}

				if err := rows.Err(); err != nil {
					log.Errorf("err:%v", err)
				}

				batchRowsAffected = int64(len(insertedIDs))
				GetDefaultLogger().Log(p.depth, start, func() (sql string, affectedRows int64) {
					return FormatSql(insertSQL, allValues...), batchRowsAffected
				}, nil)
			}
		} else {
			// MySQL/TiDB/SQLite: Use ExecContext and get LastInsertId from sql.Result
			// This avoids the race condition of querying LAST_INSERT_ID() on a different connection
			result, err := connPool.ExecContext(ctx, insertSQL, allValues...)

			GetDefaultLogger().Log(p.depth, start, func() (sql string, affectedRows int64) {
				if result != nil {
					affectedRows, _ = result.RowsAffected()
				}
				return FormatSql(insertSQL, allValues...), affectedRows
			}, err)

			if err != nil {
				execErr = err
			} else {
				batchRowsAffected, _ = result.RowsAffected()

				// Get LastInsertId and calculate IDs for the batch
				if needsIDRetrieval && batchRowsAffected > 0 {
					insertID, err := result.LastInsertId()
					if err == nil && insertID > 0 {
						// Calculate all IDs based on database-specific behavior
						switch p.clientType {
						case MySQL, TiDB:
							// MySQL's LastInsertId() returns the first ID of a batch insert
							for idx := int64(0); idx < batchRowsAffected; idx++ {
								insertedIDs = append(insertedIDs, insertID+idx)
							}
						case Sqlite:
							// SQLite's last_insert_rowid() returns the last rowid inserted
							// Calculate from last ID backwards
							firstID := insertID - batchRowsAffected + 1
							for idx := int64(0); idx < batchRowsAffected; idx++ {
								insertedIDs = append(insertedIDs, firstID+idx)
							}
						case ClickHouse:
							// ClickHouse doesn't support auto-increment IDs
							log.Warnf("ClickHouse does not support auto-increment ID retrieval")
						}
					}
				}
			}
		}

		// Handle execution errors
		if execErr != nil {
			log.Errorf("err:%v", execErr)
			if p.IsDuplicatedKeyError(execErr) {
				return &CreateInBatchesResult{
					RowsAffected: totalRowsAffected,
					Error:        p.getDuplicatedKeyError(),
				}
			}
			return &CreateInBatchesResult{
				RowsAffected: totalRowsAffected,
				Error:        execErr,
			}
		}

		totalRowsAffected += batchRowsAffected

		// Write back auto-generated IDs to the structs
		if len(insertedIDs) > 0 {
			for idx, j := 0, i; j < end && idx < len(insertedIDs); j, idx = j+1, idx+1 {
				rowValue := vv.Index(j)
				for rowValue.Kind() == reflect.Ptr {
					rowValue = rowValue.Elem()
				}

				if field := rowValue.FieldByName("Id"); field.IsValid() && field.CanSet() {
					isIntType := (field.Kind() == reflect.Int || field.Kind() == reflect.Int64) && field.Int() == 0
					isUintType := field.Kind() == reflect.Uint64 && field.Uint() == 0

					if isIntType {
						field.SetInt(insertedIDs[idx])
					} else if isUintType {
						field.SetUint(uint64(insertedIDs[idx]))
					}
				}
			}
		}
	}

	return &CreateInBatchesResult{
		RowsAffected: totalRowsAffected,
		Error:        nil,
	}
}
