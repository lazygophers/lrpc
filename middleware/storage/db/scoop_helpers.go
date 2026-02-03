package db

import (
	"database/sql"
	"reflect"
	"strings"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/stringx"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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
	if !fieldValue.CanSet() {
		return false
	}

	now := time.Now()
	var updated bool

	// Handle auto create time (only set if zero)
	if field.AutoCreateTime != 0 && fieldValue.IsZero() {
		switch fieldValue.Kind() {
		case reflect.Int64, reflect.Uint64:
			fieldValue.SetInt(now.Unix())
			updated = true
		case reflect.Struct:
			// Check if it's a time.Time type
			if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
				fieldValue.Set(reflect.ValueOf(now))
				updated = true
			}
		}
	}

	// Handle auto update time (always set on create when zero)
	if !updated && field.AutoUpdateTime != 0 && fieldValue.IsZero() {
		switch fieldValue.Kind() {
		case reflect.Int64, reflect.Uint64:
			fieldValue.SetInt(now.Unix())
			updated = true
		case reflect.Struct:
			// Check if it's a time.Time type
			if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
				fieldValue.Set(reflect.ValueOf(now))
				updated = true
			}
		}
	}

	return updated
}

// idFieldInfo contains information about the Id field type
type idFieldInfo struct {
	field      reflect.Value
	isIntType  bool
	isUintType bool
	isZero     bool
}

// getIdFieldInfo extracts Id field information from a struct value
func getIdFieldInfo(vv reflect.Value) *idFieldInfo {
	field := vv.FieldByName("Id")
	if !field.IsValid() || !field.CanSet() {
		return nil
	}

	info := &idFieldInfo{
		field: field,
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		info.isIntType = true
		info.isZero = field.Int() == 0
	case reflect.Uint64:
		info.isUintType = true
		info.isZero = field.Uint() == 0
	default:
		return nil
	}

	return info
}

// setIdValue sets the Id field value based on its type
func (info *idFieldInfo) setValue(id int64) {
	if info.isIntType {
		info.field.SetInt(id)
	} else if info.isUintType {
		info.field.SetUint(uint64(id))
	}
}

// needsAutoIncrement checks if the field needs auto-increment ID retrieval
func (info *idFieldInfo) needsAutoIncrement() bool {
	return info != nil && info.isZero && (info.isIntType || info.isUintType)
}

// queryLastInsertID queries the last insert ID based on database type
func (p *Scoop) queryLastInsertID(session *gorm.DB) (int64, error) {
	var lastInsertID int64
	var err error

	switch p.clientType {
	case Sqlite:
		err = session.Raw("SELECT last_insert_rowid()").Scan(&lastInsertID).Error
	case MySQL, TiDB:
		err = session.Raw("SELECT LAST_INSERT_ID()").Scan(&lastInsertID).Error
	case ClickHouse:
		// ClickHouse doesn't support auto-increment IDs in the traditional sense
		log.Warnf("ClickHouse does not support auto-increment ID retrieval")
		return 0, nil
	default:
		return 0, nil
	}

	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return lastInsertID, nil
}

// buildInsertSQL constructs an INSERT statement with proper IGNORE/ON CONFLICT handling
// based on the database type. It supports single-row and multi-row insertions.
// columns: list of column names
// placeholders: list of value placeholders like "(?)", or "(?), (?)" for batch inserts
func (p *Scoop) buildInsertSQL(columns []string, placeholders string) string {
	// 对列名进行引用处理，防止保留关键字导致的 SQL 语法错误
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = quoteFieldName(col, p.clientType)
	}
	columnsStr := strings.Join(quotedColumns, ", ")

	if p.ignore {
		switch p.clientType {
		case MySQL, TiDB:
			// MySQL and TiDB use INSERT IGNORE
			return "INSERT IGNORE INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders
		case Sqlite:
			// SQLite uses INSERT OR IGNORE
			return "INSERT OR IGNORE INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders
		case Postgres, GaussDB:
			// PostgreSQL and GaussDB use ON CONFLICT DO NOTHING
			return "INSERT INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders + " ON CONFLICT DO NOTHING"
		case ClickHouse:
			// ClickHouse doesn't support INSERT IGNORE, use regular INSERT
			// Duplicate handling should be done via ReplacingMergeTree or deduplication
			return "INSERT INTO " + p.table + " (" + columnsStr + ") VALUES " + placeholders
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

	// Note: This function uses the custom decode logic.
	// For fields with GORM serializers, use GORM's native Scan method instead (see First/Find methods)

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

		// Check if the field has a serializer configured
		// This requires access to GORM schema which we don't have here
		// For now, use decode which handles basic types
		err = decode(field, col)
		if err != nil {
			return err
		}
	}

	return nil
}
