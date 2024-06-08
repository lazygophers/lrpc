package db

import (
	"errors"
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils"
	"github.com/lazygophers/utils/anyx"
	"github.com/lazygophers/utils/stringx"
	"gorm.io/gorm/clause"
	"reflect"
	"strconv"
	"strings"
)

func EnsureIsSliceOrArray(obj interface{}) (res reflect.Value) {
	vo := reflect.ValueOf(obj)
	for vo.Kind() == reflect.Ptr || vo.Kind() == reflect.Interface {
		vo = vo.Elem()
	}
	k := vo.Kind()
	if k != reflect.Slice && k != reflect.Array {
		panic(fmt.Sprintf("obj required slice or array type, but got %v", vo.Type()))
	}
	res = vo
	return
}

func EscapeMysqlString(sql string) string {
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
	m := map[interface{}]struct{}{}
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

func decode(field reflect.Value, col []byte) error {
	switch field.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		val, err := strconv.ParseInt(string(col), 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(val)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		val, err := strconv.ParseUint(string(col), 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Float32,
		reflect.Float64:
		val, err := strconv.ParseFloat(string(col), 64)
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.String:
		field.SetString(string(col))
	case reflect.Bool:
		switch strings.ToLower(string(col)) {
		case "true", "1":
			field.SetBool(true)
		case "false", "0":
			field.SetBool(false)
		default:
			return fmt.Errorf("invalid bool value: %s", string(col))
		}
	case reflect.Struct:
		val := reflect.New(field.Type())
		err := utils.Scan(col, val.Interface())
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		field.Set(val.Elem())
	case reflect.Slice:
		val := reflect.New(field.Type())
		err := utils.Scan(col, val.Interface())
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		field.Set(val.Elem())
	case reflect.Map:
		val := reflect.New(field.Type())
		err := utils.Scan(col, val.Interface())
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		field.Set(val.Elem())
	case reflect.Ptr:
		val := reflect.New(field.Type().Elem())
		err := utils.Scan(col, val.Interface())
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		field.Set(val)
	default:
		log.Info(string(col))
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
	idx := strings.Index(tableName, "/")
	if idx > 0 {
		tableName = tableName[idx+1:]
		idx = strings.Index(tableName, "/")
		if idx > 0 {
			tableName = tableName[idx+1:]
			idx = strings.Index(tableName, "/")
			if idx > 0 {
				tableName = tableName[:idx]
			}
		}
	}

	return stringx.Camel2Snake(tableName + strings.TrimPrefix(elem.Elem().Name(), "Model"))
}

func hasDeleted(elem reflect.Type) bool {
	for elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	_, ok := elem.FieldByName("DeletedAt")
	return ok
}

func hasId(elem reflect.Type) bool {
	for elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	_, ok := elem.FieldByName("Id")
	return ok
}

func Camel2UnderScore(name string) string {
	var posList []int
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
	b := strings.Builder{}
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
	for {
		idx := strings.Index(sql, "?")
		if idx < 0 {
			break
		}

		out.WriteString(sql[:idx])
		sql = sql[idx+1:]

		if i >= len(values) {
			out.WriteString("?")
			continue
		}

		switch x := values[i].(type) {
		case clause.Expr:
			out.WriteString(x.SQL)
			for _, v := range x.Vars {
				out.WriteString(anyx.ToString(v))
			}
		default:
			out.WriteString(anyx.ToString(values[i]))
		}

	}

	out.WriteString(sql)
	return out.String()
}

func IsUniqueIndexConflictErr(err error) bool {
	return strings.Contains(err.Error(), "Error 1062: Duplicate entry") || strings.Contains(err.Error(), "Duplicate entry")
}

var ErrBatchesStop = errors.New("batches stop")
