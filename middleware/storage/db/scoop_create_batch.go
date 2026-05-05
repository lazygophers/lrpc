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
