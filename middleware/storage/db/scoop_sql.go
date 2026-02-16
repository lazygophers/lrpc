package db

import (
	"reflect"
	"strconv"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/candy"
	"gorm.io/gorm/clause"
)

// SQLOperationType represents the type of SQL operation
type SQLOperationType string

const (
	SQLOperationSelect SQLOperationType = "SELECT"
	SQLOperationInsert SQLOperationType = "INSERT"
	SQLOperationUpdate SQLOperationType = "UPDATE"
	SQLOperationDelete SQLOperationType = "DELETE"
)

// ToSQL returns the generated SQL query string for debugging/testing purposes
// operation: the type of SQL operation (SELECT, INSERT, UPDATE, DELETE)
// value: optional value parameter (required for INSERT and UPDATE operations)
// This method is primarily for testing and should not be used in production code
func (p *Scoop) ToSQL(operation SQLOperationType, value ...interface{}) string {
	switch operation {
	case SQLOperationSelect:
		return p.findSql()
	case SQLOperationInsert:
		if len(value) == 0 {
			return "-- ERROR: INSERT requires a value parameter"
		}
		return p.insertSql(value[0])
	case SQLOperationUpdate:
		if len(value) == 0 {
			return "-- ERROR: UPDATE requires a value parameter"
		}
		return p.updateSql(value[0])
	case SQLOperationDelete:
		return p.deleteSql()
	default:
		return "-- ERROR: unknown operation type"
	}
}

// deleteSql generates DELETE SQL statement for debugging/testing
func (p *Scoop) deleteSql() string {
	if p.table == "" {
		return "-- ERROR: table name is empty"
	}

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	// Soft delete
	if !p.unscoped && p.hasDeletedAt {
		sqlRaw.WriteString("UPDATE ")
		sqlRaw.WriteString(p.table)
		sqlRaw.WriteString(" SET deleted_at = ")
		sqlRaw.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	} else {
		sqlRaw.WriteString("DELETE FROM ")
		sqlRaw.WriteString(p.table)
	}

	// Build WHERE conditions
	conds := make([]string, 0, len(p.cond.conds))
	if !p.unscoped && p.hasDeletedAt {
		conds = append(conds, "deleted_at = 0")
	}
	conds = append(conds, p.cond.conds...)

	if len(conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(conds[0])
		for _, c := range conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	return sqlRaw.String()
}

// updateSql generates UPDATE SQL statement for debugging/testing
func (p *Scoop) updateSql(m interface{}) string {
	if p.table == "" {
		return "-- ERROR: table name is empty"
	}

	mVal := reflect.ValueOf(m)
	mType := mVal.Type()

	// Check if m is a slice (for ordered updates)
	if mType.Kind() == reflect.Slice {
		sliceLen := mVal.Len()

		// Validate slice length is even
		if sliceLen%2 != 0 {
			return "-- ERROR: slice length must be even (key-value pairs)"
		}

		if sliceLen == 0 {
			return "-- ERROR: no fields to update"
		}

		// Build ordered fields list
		type orderedField struct {
			key   string
			value interface{}
		}
		fields := make([]orderedField, 0, sliceLen/2)

		for i := 0; i < sliceLen; i += 2 {
			elem := mVal.Index(i)
			if !elem.CanInterface() {
				return "-- ERROR: cannot access slice element"
			}

			// Check that keys (even indices) are strings
			if elem.Kind() != reflect.String {
				return "-- ERROR: key must be a string"
			}

			key := elem.Interface().(string)
			valueElem := mVal.Index(i + 1)
			if !valueElem.CanInterface() {
				return "-- ERROR: cannot access value element"
			}
			value := valueElem.Interface()

			fields = append(fields, orderedField{key: key, value: value})
		}

		// Add updated_at if needed and not already present
		if p.hasUpdatedAt {
			hasUpdatedAt := false
			for _, f := range fields {
				if f.key == fieldUpdatedAt {
					hasUpdatedAt = true
					break
				}
			}
			if !hasUpdatedAt {
				fields = append(fields, orderedField{key: fieldUpdatedAt, value: time.Now().Unix()})
			}
		}

		sqlRaw := log.GetBuffer()
		defer log.PutBuffer(sqlRaw)

		sqlRaw.WriteString("UPDATE ")
		sqlRaw.WriteString(p.table)
		sqlRaw.WriteString(" SET ")

		for i, field := range fields {
			if i > 0 {
				sqlRaw.WriteString(", ")
			}
			sqlRaw.WriteString(quoteFieldName(field.key, p.clientType))
			sqlRaw.WriteString(" = ")

			switch x := field.value.(type) {
			case clause.Expr:
				sqlRaw.WriteString(x.SQL)
			default:
				sqlRaw.WriteString(candy.ToString(x))
			}
		}

		// Build WHERE conditions
		conds := make([]string, 0, len(p.cond.conds))
		if !p.unscoped && p.hasDeletedAt {
			conds = append(conds, "deleted_at = 0")
		}
		conds = append(conds, p.cond.conds...)

		if len(conds) > 0 {
			sqlRaw.WriteString(" WHERE ")
			sqlRaw.WriteString(conds[0])
			for _, c := range conds[1:] {
				sqlRaw.WriteString(" AND ")
				sqlRaw.WriteString(c)
			}
		}

		return sqlRaw.String()
	}

	// Convert input to map
	var updateMap map[string]interface{}
	if v, ok := m.(map[string]interface{}); ok {
		updateMap = v
	} else {
		if mVal.Type().Kind() == reflect.Ptr {
			mVal = mVal.Elem()
		}
		mType := mVal.Type()
		if mType.Kind() != reflect.Struct {
			return "-- ERROR: expected struct, map, or slice"
		}

		tagInfo := parseGormTags(mType)
		updateMap = make(map[string]interface{}, len(tagInfo.updatableFields))

		for fieldName, dbName := range tagInfo.updatableFields {
			fieldVal := mVal.FieldByName(fieldName)
			if !fieldVal.IsValid() || !fieldVal.CanInterface() || fieldVal.IsZero() {
				continue
			}
			updateMap[dbName] = fieldVal.Interface()
		}
	}

	if len(updateMap) == 0 {
		return "-- ERROR: no fields to update"
	}

	// Add updated_at if needed
	if p.hasUpdatedAt {
		updateMap[fieldUpdatedAt] = time.Now().Unix()
	}

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	sqlRaw.WriteString("UPDATE ")
	sqlRaw.WriteString(p.table)
	sqlRaw.WriteString(" SET ")

	var i int
	for k, v := range updateMap {
		if i > 0 {
			sqlRaw.WriteString(", ")
		}
		sqlRaw.WriteString(quoteFieldName(k, p.clientType))
		sqlRaw.WriteString(" = ")

		switch x := v.(type) {
		case clause.Expr:
			sqlRaw.WriteString(x.SQL)
		default:
			sqlRaw.WriteString(candy.ToString(x))
		}
		i++
	}

	// Build WHERE conditions
	conds := make([]string, 0, len(p.cond.conds))
	if !p.unscoped && p.hasDeletedAt {
		conds = append(conds, "deleted_at = 0")
	}
	conds = append(conds, p.cond.conds...)

	if len(conds) > 0 {
		sqlRaw.WriteString(" WHERE ")
		sqlRaw.WriteString(conds[0])
		for _, c := range conds[1:] {
			sqlRaw.WriteString(" AND ")
			sqlRaw.WriteString(c)
		}
	}

	return sqlRaw.String()
}

// insertSql generates INSERT SQL statement for debugging/testing
func (p *Scoop) insertSql(value interface{}) string {
	if p.table == "" {
		return "-- ERROR: table name is empty"
	}

	val := reflect.ValueOf(value)
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var isBatch bool
	var elemType reflect.Type

	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		if val.Len() == 0 {
			return "-- ERROR: empty slice"
		}
		isBatch = true
		elemType = val.Index(0).Type()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
	} else {
		elemType = val.Type()
	}

	if elemType.Kind() != reflect.Struct {
		return "-- ERROR: expected struct or slice of structs"
	}

	// Get statement to parse schema
	stmt := getStatement(p._db, p.table, reflect.New(elemType).Interface())
	defer putStatement(stmt)

	err := stmt.Parse(reflect.New(elemType).Interface())
	if err != nil {
		return "-- ERROR: failed to parse schema"
	}

	// Collect columns and values
	var columns []string
	var valueSets [][]string

	if isBatch {
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i)
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}

			if i == 0 {
				// Collect columns from first item
				for _, field := range stmt.Schema.Fields {
					if field.AutoCreateTime != 0 || field.AutoUpdateTime != 0 || field.PrimaryKey {
						continue
					}
					columns = append(columns, field.DBName)
				}
			}

			var values []string
			for _, field := range stmt.Schema.Fields {
				if field.AutoCreateTime != 0 || field.AutoUpdateTime != 0 || field.PrimaryKey {
					continue
				}
				fieldValue := item.FieldByName(field.Name)
				if fieldValue.IsValid() {
					values = append(values, candy.ToString(fieldValue.Interface()))
				} else {
					values = append(values, "NULL")
				}
			}
			valueSets = append(valueSets, values)
		}
	} else {
		// Single insert
		for _, field := range stmt.Schema.Fields {
			if field.AutoCreateTime != 0 || field.AutoUpdateTime != 0 || field.PrimaryKey {
				continue
			}
			columns = append(columns, field.DBName)
		}

		var values []string
		for _, field := range stmt.Schema.Fields {
			if field.AutoCreateTime != 0 || field.AutoUpdateTime != 0 || field.PrimaryKey {
				continue
			}
			fieldValue := val.FieldByName(field.Name)
			if fieldValue.IsValid() {
				values = append(values, candy.ToString(fieldValue.Interface()))
			} else {
				values = append(values, "NULL")
			}
		}
		valueSets = append(valueSets, values)
	}

	// Build INSERT SQL
	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	if p.ignore {
		switch p.clientType {
		case MySQL, TiDB:
			sqlRaw.WriteString("INSERT IGNORE INTO ")
		case Sqlite:
			sqlRaw.WriteString("INSERT OR IGNORE INTO ")
		case Postgres, GaussDB:
			sqlRaw.WriteString("INSERT INTO ")
		default:
			sqlRaw.WriteString("INSERT INTO ")
		}
	} else {
		sqlRaw.WriteString("INSERT INTO ")
	}

	sqlRaw.WriteString(p.table)
	sqlRaw.WriteString(" (")
	for i, col := range columns {
		if i > 0 {
			sqlRaw.WriteString(", ")
		}
		sqlRaw.WriteString(col)
	}
	sqlRaw.WriteString(") VALUES ")

	for i, values := range valueSets {
		if i > 0 {
			sqlRaw.WriteString(", ")
		}
		sqlRaw.WriteString("(")
		for j, v := range values {
			if j > 0 {
				sqlRaw.WriteString(", ")
			}
			sqlRaw.WriteString(v)
		}
		sqlRaw.WriteString(")")
	}

	if p.ignore && (p.clientType == Postgres || p.clientType == GaussDB) {
		sqlRaw.WriteString(" ON CONFLICT DO NOTHING")
	}

	return sqlRaw.String()
}

func (p *Scoop) findSql() string {
	b := log.GetBuffer()
	defer log.PutBuffer(b)

	b.WriteString("SELECT ")
	if len(p.selects) > 0 {
		b.WriteString(quoteFieldName(p.selects[0], p.clientType))

		for _, s := range p.selects[1:] {
			b.WriteString(", ")
			b.WriteString(quoteFieldName(s, p.clientType))
		}
	} else {
		b.WriteString("*")
	}

	b.WriteString(" FROM ")
	b.WriteString(p.table)

	// Add JOIN clauses
	if len(p.joins) > 0 {
		for _, join := range p.joins {
			b.WriteString(" ")
			b.WriteString(join.joinType)
			b.WriteString(" JOIN ")
			b.WriteString(join.table)
			if join.condition != "" {
				b.WriteString(" ON ")
				b.WriteString(join.condition)
			}
		}
	}

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

	// Add HAVING clause
	if p.havingCond != nil && len(p.havingCond.conds) > 0 {
		b.WriteString(" HAVING ")
		b.WriteString(p.havingCond.conds[0])
		for _, c := range p.havingCond.conds[1:] {
			b.WriteString(" AND ")
			b.WriteString(c)
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
