package db

import (
	"database/sql"
	"reflect"
	"strings"
	"time"

	"github.com/lazygophers/log"
	"gorm.io/gorm"
)

// queryLastInsertID retrieves the last inserted ID from the database
// Supports different SQL dialects with their specific syntax
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
		quotedColumns[i] = QuoteFieldName(col, p.clientType)
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
