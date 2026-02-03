package db

import (
	"fmt"
	"reflect"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/candy"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UpdateResult struct {
	RowsAffected int64
	Error        error
}

// updateWithOrder performs an UPDATE operation with ordered fields.
// updateFields is a slice of [key, value, key, value, ...] pairs.
// This preserves the order of fields in the UPDATE SET clause.
func (p *Scoop) updateWithOrder(updateFields []interface{}) *UpdateResult {
	if p.cond.skip {
		return &UpdateResult{}
	}
	if len(updateFields) == 0 {
		return &UpdateResult{
			Error: fmt.Errorf("Update failed: updateFields is empty, no values to update"),
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

	// Build ordered map for fields
	// Using a slice of key-value pairs to preserve order
	type orderedField struct {
		key   string
		value interface{}
	}
	fields := make([]orderedField, 0, len(updateFields)/2+1)

	// Add update fields in order
	for i := 0; i < len(updateFields); i += 2 {
		key := updateFields[i].(string)
		value := updateFields[i+1]
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

	p.inc()
	defer p.dec()

	// 获取schema以检查serializer
	stmt := getStatement(p._db, p.table, nil)
	defer putStatement(stmt)

	sqlRaw := log.GetBuffer()
	defer log.PutBuffer(sqlRaw)

	sqlRaw.WriteString("UPDATE ")
	sqlRaw.WriteString(p.table)

	sqlRaw.WriteString(" SET ")
	// Pre-allocate values slice with capacity based on fields length
	values := make([]interface{}, 0, len(fields))
	for i, orderedField := range fields {
		if i > 0 {
			sqlRaw.WriteString(", ")
		}
		sqlRaw.WriteString(quoteFieldName(orderedField.key, p.clientType))
		sqlRaw.WriteString("=")

		// Handle clause.Expr with proper type assertion
		switch x := orderedField.value.(type) {
		case clause.Expr:
			// For expressions, use the SQL directly instead of placeholder
			sqlRaw.WriteString(x.SQL)
			// Append expression variables to values
			values = append(values, x.Vars...)
		default:
			// For normal values, use placeholder
			sqlRaw.WriteString("?")

			// Check if this field has a serializer
			var finalValue interface{}
			if stmt.Schema != nil {
				if field, ok := stmt.Schema.FieldsByDBName[orderedField.key]; ok && field.Serializer != nil {
					// Apply serializer if configured
					serializedValue, serErr := field.Serializer.Value(stmt.Context, field, reflect.Value{}, orderedField.value)
					if serErr != nil {
						log.Errorf("serializer failed for field %s: %v", orderedField.key, serErr)
						finalValue = orderedField.value
					} else {
						finalValue = serializedValue
					}
				} else {
					finalValue = candy.ToString(x)
				}
			} else {
				finalValue = candy.ToString(x)
			}
			values = append(values, finalValue)
		}
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

	// 获取schema以检查serializer
	stmt := getStatement(p._db, p.table, nil)
	defer putStatement(stmt)

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

			// Check if this field has a serializer
			var finalValue interface{}
			if stmt.Schema != nil {
				if field, ok := stmt.Schema.FieldsByDBName[k]; ok && field.Serializer != nil {
					// Apply serializer if configured
					serializedValue, serErr := field.Serializer.Value(stmt.Context, field, reflect.Value{}, v)
					if serErr != nil {
						log.Errorf("serializer failed for field %s: %v", k, serErr)
						finalValue = v
					} else {
						finalValue = serializedValue
					}
				} else {
					finalValue = candy.ToString(x)
				}
			} else {
				finalValue = candy.ToString(x)
			}
			values = append(values, finalValue)
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
// The parameters can be:
//   - map[string]interface{}: key-value pairs for update (order not preserved)
//   - struct: non-zero fields are updated (order not preserved)
//   - variadic args: key, value, key, value, ... pairs (order preserved)
//   - argument count must be even
//   - keys at even positions must be strings
//   - values at odd positions can be any type
//
// Only non-zero fields are updated for structs. Use clause.Expr for SQL expressions.
// Automatically updates UpdatedAt field if present.
// Returns UpdateResult with RowsAffected count.
//
// Example:
//
//	// Using map (order not preserved)
//	result := scoop.Table("users").Where("id = ?", 1).Updates(map[string]interface{}{"age": 25, "name": "John"})
//	// Using struct (order not preserved)
//	result := scoop.Table("users").Where("id = ?", 1).Updates(&User{Age: 25})
//	// Using variadic args (order preserved)
//	result := scoop.Table("users").Where("id = ?", 1).Updates("name", "John", "age", 25)
func (p *Scoop) Updates(m interface{}, args ...interface{}) *UpdateResult {
	if p.cond.skip {
		return &UpdateResult{}
	}

	p.inc()
	defer p.dec()

	// Check if using variadic args (key, value, key, value, ...)
	if len(args) > 0 {
		// First argument m is the first key, args contains the rest
		// Combine m and args into a single slice
		allArgs := make([]interface{}, 0, len(args)+1)
		allArgs = append(allArgs, m)
		allArgs = append(allArgs, args...)

		// Validate argument count is even
		if len(allArgs)%2 != 0 {
			return &UpdateResult{
				Error: fmt.Errorf("Updates failed: argument count must be even (key-value pairs), got %d arguments", len(allArgs)),
			}
		}

		// Validate keys are strings
		for i := 0; i < len(allArgs); i += 2 {
			if _, ok := allArgs[i].(string); !ok {
				return &UpdateResult{
					Error: fmt.Errorf("Updates failed: key at position %d must be a string, got %T", i, allArgs[i]),
				}
			}
		}

		return p.updateWithOrder(allArgs)
	}

	// Check if m is a map
	if v, ok := m.(map[string]interface{}); ok {
		return p.update(v)
	}

	mVal := reflect.ValueOf(m)
	mType := mVal.Type()

	// Check if m is a slice (for ordered updates)
	if mType.Kind() == reflect.Slice {
		sliceLen := mVal.Len()

		// Validate slice length is even
		if sliceLen%2 != 0 {
			return &UpdateResult{
				Error: fmt.Errorf("Updates failed: slice length must be even (key-value pairs), got length %d", sliceLen),
			}
		}

		// Validate and convert to []interface{}
		updateFields := make([]interface{}, sliceLen)
		for i := 0; i < sliceLen; i++ {
			elem := mVal.Index(i)
			if !elem.CanInterface() {
				return &UpdateResult{
					Error: fmt.Errorf("Updates failed: cannot access slice element at index %d", i),
				}
			}

			// Check that keys (even indices) are strings
			if i%2 == 0 {
				if elem.Kind() != reflect.String {
					return &UpdateResult{
						Error: fmt.Errorf("Updates failed: key at index %d must be a string, got %v", i, elem.Kind()),
					}
				}
			}
			updateFields[i] = elem.Interface()
		}

		return p.updateWithOrder(updateFields)
	}

	// Handle struct
	if mType.Kind() == reflect.Ptr {
		mVal = mVal.Elem()
		mType = mVal.Type()
	}
	if mType.Kind() != reflect.Struct {
		return &UpdateResult{
			Error: fmt.Errorf("Updates failed: expected map, struct, or variadic args, got %v", mType.Kind()),
		}
	}

	// Get GORM schema to check for serializers
	stmt := getStatement(p._db, p.table, m)
	defer putStatement(stmt)

	err := stmt.ParseWithSpecialTableName(m, p.table)
	if err != nil {
		log.Errorf("err:%v", err)
		return &UpdateResult{
			Error: fmt.Errorf("Updates failed: unable to parse schema: %w", err),
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

		// Check if field has a serializer
		var finalValue interface{}
		if field, ok := stmt.Schema.FieldsByDBName[dbName]; ok && field.Serializer != nil {
			serializedValue, serErr := field.Serializer.Value(stmt.Context, field, mVal, fieldVal.Interface())
			if serErr != nil {
				log.Errorf("err:%v", serErr)
				return &UpdateResult{
					Error: fmt.Errorf("Updates failed: failed to serialize field %s: %w", fieldName, serErr),
				}
			}
			finalValue = serializedValue
		} else {
			finalValue = fieldVal.Interface()
		}

		valMap[dbName] = finalValue
	}

	if len(valMap) == 0 {
		return &UpdateResult{
			Error: fmt.Errorf("Updates failed: no non-zero fields found in struct %v to update", mType),
		}
	}
	return p.update(valMap)
}
