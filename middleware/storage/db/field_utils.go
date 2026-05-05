package db

import (
	"reflect"
	"time"

	"github.com/lazygophers/utils/stringx"
	"gorm.io/gorm/schema"
)

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
	if field.AutoCreateTime > 0 && fieldValue.IsZero() {
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
	if !updated && field.AutoUpdateTime > 0 && fieldValue.IsZero() {
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

// setValue sets the Id field value based on its type
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
