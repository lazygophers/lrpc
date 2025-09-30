package db

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/stringx"
	"gorm.io/gorm/clause"
)

func EnsureIsSliceOrArray(obj interface{}) reflect.Value {
	vo := reflect.ValueOf(obj)
	for vo.Kind() == reflect.Ptr || vo.Kind() == reflect.Interface {
		vo = vo.Elem()
	}
	k := vo.Kind()
	if k != reflect.Slice && k != reflect.Array {
		panic(fmt.Sprintf("obj required slice or array type, but got %v", vo.Type()))
	}
	return vo
}

func EscapeMysqlString(sql string) string {
	// Fast path: check if escaping is needed
	needsEscape := false
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		if c == 0 || c == '\n' || c == '\r' || c == '\\' || c == '\'' || c == '"' || c == '\032' {
			needsEscape = true
			break
		}
	}

	// Return original string if no escaping needed
	if !needsEscape {
		return sql
	}

	// Slow path: perform escaping
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		escape = 0
		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
		case '\n': /* Must be escaped for logs */
			escape = 'n'
		case '\r':
			escape = 'r'
		case '\\':
			escape = '\\'
		case '\'':
			escape = '\''
		case '"': /* Better safe than sorry */
			escape = '"'
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}
		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}
	return string(dest)
}

// UniqueSlice removes duplicate elements from a slice while preserving order.
// It uses a map for O(n) time complexity but requires elements to be comparable.
//
// Performance characteristics:
//   - Time: O(n) where n is the slice length
//   - Space: O(n) for the map storage
//   - Returns original slice unchanged if length < 2
//   - Note: Uses reflection, so has overhead. Consider type-specific implementations
//     for performance-critical paths.
//
// Example:
//
//	input := []int{1, 2, 2, 3, 1, 4}
//	output := UniqueSlice(input).([]int) // []int{1, 2, 3, 4}
func UniqueSlice(s interface{}) interface{} {
	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Slice {
		panic(fmt.Sprintf("s required slice, but got %v", t))
	}

	vo := reflect.ValueOf(s)

	if vo.Len() < 2 {
		return s
	}

	res := reflect.MakeSlice(t, 0, vo.Len())
	m := make(map[interface{}]struct{}, vo.Len())
	for i := 0; i < vo.Len(); i++ {
		el := vo.Index(i)
		eli := el.Interface()
		if _, ok := m[eli]; !ok {
			res = reflect.Append(res, el)
			m[eli] = struct{}{}
		}
	}

	return res.Interface()
}

// scanComplexType handles scanning for Struct, Slice, Map, and Ptr types using utils.Scan
func scanComplexType(field reflect.Value, col []byte, isPtr bool) error {
	var val reflect.Value
	if isPtr {
		val = reflect.New(field.Type().Elem())
	} else {
		val = reflect.New(field.Type())
	}

	err := utils.Scan(col, val.Interface())
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	if isPtr {
		field.Set(val)
	} else {
		field.Set(val.Elem())
	}
	return nil
}

func decode(field reflect.Value, col []byte) error {
	// Convert []byte to string once for types that need it
	colStr := string(col)

	switch field.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		val, err := strconv.ParseInt(colStr, 10, 64)
		if err != nil {
			log.Errorf("parse %s err:%s", col, err)
			return err
		}
		field.SetInt(val)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		val, err := strconv.ParseUint(colStr, 10, 64)
		if err != nil {
			log.Errorf("err:%s", err)
			return err
		}
		field.SetUint(val)
	case reflect.Float32,
		reflect.Float64:
		val, err := strconv.ParseFloat(colStr, 64)
		if err != nil {
			log.Errorf("err:%s", err)
			return err
		}
		field.SetFloat(val)
	case reflect.String:
		field.SetString(colStr)
	case reflect.Bool:
		switch strings.ToLower(colStr) {
		case "true", "1":
			field.SetBool(true)
		case "false", "0":
			field.SetBool(false)
		default:
			return fmt.Errorf("invalid bool value: %s", colStr)
		}
	case reflect.Struct, reflect.Slice, reflect.Map:
		return scanComplexType(field, col, false)
	case reflect.Ptr:
		return scanComplexType(field, col, true)
	default:
		log.Errorf("unsupported column: %s", colStr)
		return fmt.Errorf("invalid type: %s", field.Kind().String())
	}

	return nil
}

func getTableName(elem reflect.Type) string {
	for elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	if x, ok := reflect.New(elem).Interface().(Tabler); ok {
		return x.TableName()
	}

	tableName := elem.PkgPath()
	// Extract the third level component from path like "github.com/user/project/package"
	// Use SplitN to limit splits and avoid unnecessary work
	parts := strings.SplitN(tableName, "/", 4)
	if len(parts) >= 3 {
		tableName = parts[2]
	}

	// Get element name safely without calling Elem() on non-pointer types
	elemName := elem.Name()
	return stringx.Camel2Snake(tableName + strings.TrimPrefix(elemName, "Model"))
}

// hasField checks if the given type has a field with the specified name.
// It unwraps pointer types before checking.
func hasField(elem reflect.Type, fieldName string) bool {
	for elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	_, ok := elem.FieldByName(fieldName)
	return ok
}

func hasDeletedAt(elem reflect.Type) bool {
	return hasField(elem, "DeletedAt")
}

func hasCreatedAt(elem reflect.Type) bool {
	return hasField(elem, "CreatedAt")
}

func hasUpdatedAt(elem reflect.Type) bool {
	return hasField(elem, "UpdatedAt")
}

func hasId(elem reflect.Type) bool {
	return hasField(elem, "Id")
}

func Camel2UnderScore(name string) string {
	if name == "" {
		return ""
	}

	// Preallocate posList with estimated capacity (typically 1/4 of string length)
	posList := make([]int, 0, len(name)/4+1)
	i := 1
	for i < len(name) {
		if name[i] >= 'A' && name[i] <= 'Z' {
			posList = append(posList, i)
			i++
			for i < len(name) && name[i] >= 'A' && name[i] <= 'Z' {
				i++
			}
		} else {
			i++
		}
	}
	lower := strings.ToLower(name)
	if len(posList) == 0 {
		return lower
	}
	// Preallocate Builder capacity: original length + underscores
	var b strings.Builder
	b.Grow(len(name) + len(posList))
	left := 0
	for _, right := range posList {
		b.WriteString(lower[left:right])
		b.WriteByte('_')
		left = right
	}
	b.WriteString(lower[left:])
	return b.String()
}

func FormatSql(sql string, values ...interface{}) string {
	out := log.GetBuffer()
	defer log.PutBuffer(out)

	var i int
	lastIdx := 0
	for {
		idx := strings.IndexByte(sql[lastIdx:], '?')
		if idx < 0 {
			break
		}
		idx += lastIdx

		out.WriteString(sql[lastIdx:idx])
		lastIdx = idx + 1

		if i >= len(values) {
			out.WriteString("?")
			continue
		}

		switch x := values[i].(type) {
		case clause.Expr:
			out.WriteString(x.SQL)
			for _, v := range x.Vars {
				out.WriteString(candy.ToString(v))
			}
		default:
			out.WriteString(candy.ToString(values[i]))
		}
		i++
	}

	out.WriteString(sql[lastIdx:])
	return out.String()
}

func IsUniqueIndexConflictErr(err error) bool {
	if err == nil {
		return false
	}
	// Check for "Duplicate entry" which covers both:
	// - "Error 1062: Duplicate entry" (MySQL error format)
	// - "Duplicate entry" (shorter format)
	return strings.Contains(err.Error(), "Duplicate entry")
}

var ErrBatchesStop = errors.New("batches stop")
