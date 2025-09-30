package db

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/lazygophers/log"
)

type Cond struct {
	conds       []string
	isOr        bool
	isTopLevel  bool
	tablePrefix string

	// 标记跳过请求，用于一些逻辑上就不需要进行请求的场景
	skip bool

	// clientType for database-specific quoting
	clientType string
}

// quoteFieldName quotes a field name with the appropriate quote character for the database type.
// MySQL uses backticks (`), PostgreSQL and SQLite use double quotes (").
func quoteFieldName(name string, clientType string) string {
	// Check if already quoted
	if strings.HasPrefix(name, "`") || strings.HasPrefix(name, "\"") {
		return name
	}

	switch clientType {
	case MySQL:
		return fmt.Sprintf("`%s`", name)
	case Postgres, Sqlite:
		return fmt.Sprintf("\"%s\"", name)
	default:
		// Default to double quotes for unknown types
		return fmt.Sprintf("\"%s\"", name)
	}
}

func quoteStr(s string) string {
	return strconv.Quote(s)
}

func simpleTypeToStr(value interface{}, quoteSlice bool) string {
	if value == nil {
		return "NULL"
	}
	vo := reflect.ValueOf(value)
	for vo.Kind() == reflect.Ptr || vo.Kind() == reflect.Interface {
		if vo.IsNil() {
			return "NULL"
		}
		vo = vo.Elem()
	}
	value = vo.Interface()
	switch v := value.(type) {
	case string:
		v = EscapeMysqlString(v)
		return quoteStr(v)
	case []byte:
		s := EscapeMysqlString(string(v))
		return quoteStr(s)
	case bool:
		if v {
			return "1"
		}
		return "0"
	}
	// 容器单独处理
	switch vo.Kind() {
	case reflect.Slice, reflect.Array:
		count := vo.Len()
		elList := make([]string, 0, count)
		for x := 0; x < count; x++ {
			el := vo.Index(x)
			elList = append(elList, simpleTypeToStr(el.Interface(), quoteSlice))
		}
		res := strings.Join(elList, ",")
		if quoteSlice {
			res = fmt.Sprintf("(%s)", res)
		}
		return res
	case reflect.Uint32, reflect.Uint64, reflect.Uint16, reflect.Uint8, reflect.Uint:
		return strconv.FormatUint(vo.Uint(), 10)
	case reflect.Int32, reflect.Int64, reflect.Int16, reflect.Int8, reflect.Int:
		return strconv.FormatInt(vo.Int(), 10)
	case reflect.String:
		return quoteStr(vo.String())
	default:
		return quoteStr(fmt.Sprintf("%v", value))
	}
}

func (p *Cond) whereRaw(cond string, values ...interface{}) {
	var res string
	if len(values) == 0 {
		res = cond
	} else {
		list := strings.Split(cond, "?")
		if len(list)-1 != len(values) {
			log.Warnf("invalid number of values, q %d, v %d", len(list)-1, len(values))
		}
		out := make([]string, 0, len(list)+len(values))
		for i := 0; i < len(list); i++ {
			out = append(out, list[i])
			if i < len(list)-1 && i < len(values) {
				value := values[i]
				out = append(out, simpleTypeToStr(value, false))
			}
		}
		res = strings.Join(out, "")
	}
	if res != "" {
		p.conds = append(p.conds, res)
	}
}

func (p *Cond) addCond(fieldName, op string, val interface{}) {
	if fieldName == "" {
		panic("fieldName empty")
	}
	if op == "" {
		panic(fmt.Sprintf("empty op for field %s", fieldName))
	}

	valStr := simpleTypeToStr(val, true)
	var cond string
	if p.tablePrefix == "" {
		cond = fmt.Sprintf("(%s %s %s)", quoteFieldName(fieldName, p.clientType), op, valStr)
	} else {
		cond = fmt.Sprintf("(%s.%s %s %s)", p.tablePrefix, fieldName, op, valStr)
	}
	p.conds = append(p.conds, cond)
}

func getFirstInvalidFieldNameCharIndex(s string) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Valid characters: alphanumeric, underscore, space, dot, backtick
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '_' || c == ' ' || c == '.' || c == '`' {
			continue
		}
		return i
	}
	return -1
}

func (p *Cond) addSubWhere(isOr bool, args ...interface{}) {
	subCond := &Cond{
		isOr:        isOr,
		tablePrefix: p.tablePrefix,
	}
	subCond.where(args...)
	c := subCond.ToString()
	if c == "" {
		return
	}

	p.conds = append(p.conds, c)
}

func (p *Cond) addCmdCond(cmd string, cond interface{}) {
	if strings.HasPrefix(cmd, "or") {
		p.addSubWhere(true, cond)
	} else if strings.HasPrefix(cmd, "and") {
		p.addSubWhere(false, cond)
	} else if strings.HasPrefix(cmd, "raw") {
		vo := reflect.ValueOf(cond)
		vk := vo.Kind()
		var list []interface{}
		if vk == reflect.Slice || vk == reflect.Array {
			n := vo.Len()
			list = make([]interface{}, 0, n)
			for i := 0; i < n; i++ {
				vv := vo.Index(i)
				if !vv.CanInterface() {
					panic("$raw slice element can't convert to interface")
				}
				list = append(list, vv.Interface())
			}
		} else {
			list = make([]interface{}, 1)
			list[0] = cond
		}
		if len(list) == 0 {
			panic("$raw list empty")
		}
		list0 := reflect.ValueOf(list[0])
		if list0.Kind() != reflect.String {
			panic("$raw cond required string type")
		}
		p.whereRaw(list0.String(), list[1:]...)
	} else {
		panic(fmt.Sprintf("invalid cmd %s", cmd))
	}
}

func toInterfaces(v reflect.Value) []interface{} {
	list := make([]interface{}, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		vv := v.Index(i)
		if !vv.CanInterface() {
			panic("slice element can't convert to interface")
		}
		list = append(list, vv.Interface())
	}
	return list
}

func getOp(fieldName string) (newFieldName, op string) {
	op = "="
	newFieldName = fieldName
	idx := getFirstInvalidFieldNameCharIndex(fieldName)
	if idx > 0 {
		o := strings.TrimSpace(fieldName[idx:])
		newFieldName = fieldName[:idx]
		if o != "" {
			op = o
		}
	}
	return
}

// where 支持多种调用形式：
//   - map[string]interface{} 多个条件
//     key -> field name [op], op 选填，可以这样写, fieldName> 表示 >, fieldName Like 表示 like 操作
//     val -> 任意类型
//   - []interface{}
//     interface{} 可以是:
//     - []string, 可以写成 {"fieldName", "op"?, value}
//     - map[string]interface{}
//   - fieldName 'op'? arg, op 不填，也就是只有两个入参时，表示是相等操作 =
//   - 自己构造的sql条件，比如: a = ? or (c = ? and d = ?), x, y, z
func (p *Cond) where(args ...interface{}) {
	if len(args) == 0 {
		return
	}

	if cond, ok := args[0].(*Cond); ok {
		p.whereRaw(cond.ToString())
		p.where(args[1:]...)
		return
	}

	arg0 := reflect.ValueOf(args[0])
	for arg0.Kind() == reflect.Interface || arg0.Kind() == reflect.Ptr {
		arg0 = arg0.Elem()
	}

	switch arg0.Kind() {
	case reflect.Bool:
		v := arg0.Bool()
		if v {
			p.conds = append(p.conds, "(1=1)")
		} else {
			p.conds = append(p.conds, "(1=0)")
			p.skip = true
		}

	case reflect.String:
		fieldName := arg0.String()
		if strings.HasPrefix(fieldName, "$") {
			if len(args) != 2 {
				panic(fmt.Sprintf("invalid number of args %d for $... cond, expected 2", len(args)))
			}
			p.addCmdCond(fieldName[1:], args[1])
			break
		}
		if strings.IndexByte(fieldName, '?') >= 0 {
			p.whereRaw(fieldName, args[1:]...)
			break
		}
		var op string
		var val interface{}
		if len(args) == 2 {
			fieldName, op = getOp(fieldName)
			val = args[1]
			p.addCond(fieldName, op, val)
		} else if len(args) == 3 {
			vo := reflect.ValueOf(args[1])
			if vo.Kind() == reflect.String {
				op = vo.String()
			} else if vo.Kind() == reflect.Int32 {
				// 可以支持 '>' 单括号写法
				op = strings.TrimSpace(fmt.Sprintf("%c", int(vo.Int())))
				if op == "" {
					panic(fmt.Sprintf("invalid op type with int %d", vo.Int()))
				}
			} else {
				panic(fmt.Sprintf("invalid op type %v", vo.Type()))
			}
			val = args[2]
			p.addCond(fieldName, op, val)
		} else if len(args) == 1 {
			p.conds = append(p.conds, fieldName)
		} else {
			panic(fmt.Sprintf("invalid number of where args %d by `string` prefix", len(args)))
		}

	case reflect.Map:
		typ := arg0.Type()
		if typ.Key().Kind() != reflect.String {
			panic(fmt.Sprintf("map key type required string, but got %v", typ.Key()))
		}
		for _, k := range arg0.MapKeys() {
			fieldName := k.String()
			val := arg0.MapIndex(k)
			if !val.IsValid() || !val.CanInterface() {
				panic(fmt.Sprintf("invalid map val for field %s", fieldName))
			}
			if strings.HasPrefix(fieldName, "$") {
				p.addCmdCond(fieldName[1:], val.Interface())
				continue
			}
			var op string
			fieldName, op = getOp(fieldName)
			p.addCond(fieldName, op, val.Interface())
		}
		if len(args) > 1 {
			p.where(args[1:]...)
		}

	case reflect.Slice, reflect.Array:
		n := arg0.Len()
		if n == 0 {
			break
		}
		// 检查下第1项，是不是 string，是的话，表示这是一条条件
		{
			v := arg0.Index(0)
			if v.Kind() == reflect.String {
				p.where(toInterfaces(arg0)...)
				if len(args) > 1 {
					p.where(args[1:]...)
				}
				break
			}
		}
		for i := 0; i < n; i++ {
			v := arg0.Index(i)
			for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			vk := v.Kind()
			if vk == reflect.Map {
				p.addSubWhere(false, v.Interface())
			} else {
				var list []interface{}
				if vk == reflect.Slice || vk == reflect.Array {
					vLen := v.Len()
					list = make([]interface{}, 0, vLen)
					for ii := 0; ii < vLen; ii++ {
						vv := v.Index(ii)
						if !vv.CanInterface() {
							panic("slice element can't convert to interface")
						}
						list = append(list, vv.Interface())
					}
				} else {
					if !v.CanInterface() {
						panic("slice element can't convert to interface")
					}
					list = make([]interface{}, 1)
					list[0] = v.Interface()
				}
				p.where(list...)
			}
		}
	default:
		panic("unhandled default case")
	}
}

func (p *Cond) ToString() string {
	n := len(p.conds)
	if n == 0 {
		return ""
	} else if n == 1 {
		return p.conds[0]
	}
	var s string
	if p.isOr {
		s = strings.Join(p.conds, " OR ")
	} else {
		s = strings.Join(p.conds, " AND ")
	}
	if !p.isTopLevel {
		s = fmt.Sprintf("(%s)", s)
	}
	return s
}

func (p *Cond) String() string {
	return p.ToString()
}

func (p *Cond) GoString() string {
	return p.ToString()
}

func (p *Cond) Where(args ...interface{}) *Cond {
	p.addSubWhere(false, args...)
	return p
}

func (p *Cond) OrWhere(args ...interface{}) *Cond {
	p.addSubWhere(true, args...)
	return p
}

func (p *Cond) Or(args ...interface{}) *Cond {
	p.addSubWhere(true, args...)
	return p
}

func (p *Cond) Like(column string, value string) *Cond {
	if value == "" {
		return p
	}
	p.where(column, "LIKE", "%"+value+"%")
	return p
}

func (p *Cond) LeftLike(column string, value string) *Cond {
	if value == "" {
		return p
	}
	p.where(column, "LIKE", value+"%")
	return p
}

func (p *Cond) RightLike(column string, value string) *Cond {
	if value == "" {
		return p
	}
	p.where(column, "LIKE", "%"+value)
	return p
}

func (p *Cond) NotLike(column string, value string) *Cond {
	if value == "" {
		return p
	}
	p.where(column, "NOT LIKE", "%"+value+"%")
	return p
}

func (p *Cond) NotLeftLike(column string, value string) *Cond {
	if value == "" {
		return p
	}
	p.where(column, "NOT LIKE", value+"%")
	return p
}

func (p *Cond) NotRightLike(column string, value string) *Cond {
	if value == "" {
		return p
	}
	p.where(column, "NOT LIKE", "%"+value)
	return p
}

func (p *Cond) Between(column string, min, max interface{}) *Cond {
	p.whereRaw(quoteFieldName(column, p.clientType)+" BETWEEN ? AND ?", min, max)
	return p
}

func (p *Cond) NotBetween(column string, min, max interface{}) *Cond {
	p.whereRaw(quoteFieldName(column, p.clientType)+" NOT BETWEEN ? AND ?", min, max)
	return p
}

// Reset clears all conditions and resets the Cond to its initial state.
// This includes clearing conditions, resetting OR/AND mode, top-level flag,
// table prefix, and skip flag. Use this when you want to reuse a Cond object.
func (p *Cond) Reset() *Cond {
	p.conds = p.conds[:0]
	p.isOr = false
	p.isTopLevel = true
	p.tablePrefix = ""
	p.skip = false
	return p
}
