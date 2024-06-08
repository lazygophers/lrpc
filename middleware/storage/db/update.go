package db

import (
	"bytes"
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/json"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
)

// CASE WHEN (`a` = 1) THEN "1" WHEN (`b` = 2) THEN 2 ELSE 10 END
func UpdateCase[M string | *Cond](caseMap map[M]any, def ...interface{}) clause.Expr {
	var b bytes.Buffer
	b.WriteString("CASE")

	appendAny := func(v any) {
		switch x := v.(type) {
		case string:
			b.WriteString(strconv.Quote(x))
		case []byte:
			b.WriteString(strconv.Quote(string(x)))
		case int:
			b.WriteString(strconv.FormatInt(int64(x), 10))
		case int8:
			b.WriteString(strconv.FormatInt(int64(x), 10))
		case int16:
			b.WriteString(strconv.FormatInt(int64(x), 10))
		case int32:
			b.WriteString(strconv.FormatInt(int64(x), 10))
		case int64:
			b.WriteString(strconv.FormatInt(x, 10))
		case uint:
			b.WriteString(strconv.FormatUint(uint64(x), 10))
		case uint8:
			b.WriteString(strconv.FormatUint(uint64(x), 10))
		case uint16:
			b.WriteString(strconv.FormatUint(uint64(x), 10))
		case uint32:
			b.WriteString(strconv.FormatUint(uint64(x), 10))
		case uint64:
			b.WriteString(strconv.FormatUint(x, 10))
		case float32:
			b.WriteString(strconv.FormatFloat(float64(x), 'f', -1, 32))
		case float64:
			b.WriteString(strconv.FormatFloat(x, 'f', -1, 64))
		case bool:
			if x {
				b.WriteString("1")
			} else {
				b.WriteString("0")
			}
		default:
			buf, err := json.MarshalString(x)
			if err != nil {
				// NOTE: should not happen
				log.Errorf("err:%v", err)
			} else {
				b.WriteString(strconv.Quote(buf))
			}
		}
	}

	for k, v := range caseMap {
		b.WriteString(" WHEN ")
		b.WriteString(fmt.Sprintf("%s", k))
		b.WriteString(" THEN ")
		appendAny(v)
	}

	if len(def) > 0 {
		b.WriteString(" ELSE ")
		appendAny(def[0])
	}

	b.WriteString(" END")

	return gorm.Expr(b.String())
}

// CASE `a` WHEN 1 THEN "1" WHEN 2 THEN 2 ELSE 10 END
func UpdateCaseOneField(field string, caseMap map[any]any, def ...interface{}) clause.Expr {
	var b bytes.Buffer

	appendAny := func(v any) {
		switch x := v.(type) {
		case string:
			b.WriteString(strconv.Quote(x))
		case []byte:
			b.WriteString(strconv.Quote(string(x)))
		case int:
			b.WriteString(strconv.FormatInt(int64(x), 10))
		case int8:
			b.WriteString(strconv.FormatInt(int64(x), 10))
		case int16:
			b.WriteString(strconv.FormatInt(int64(x), 10))
		case int32:
			b.WriteString(strconv.FormatInt(int64(x), 10))
		case int64:
			b.WriteString(strconv.FormatInt(x, 10))
		case uint:
			b.WriteString(strconv.FormatUint(uint64(x), 10))
		case uint8:
			b.WriteString(strconv.FormatUint(uint64(x), 10))
		case uint16:
			b.WriteString(strconv.FormatUint(uint64(x), 10))
		case uint32:
			b.WriteString(strconv.FormatUint(uint64(x), 10))
		case uint64:
			b.WriteString(strconv.FormatUint(x, 10))
		case float32:
			b.WriteString(strconv.FormatFloat(float64(x), 'f', -1, 32))
		case float64:
			b.WriteString(strconv.FormatFloat(x, 'f', -1, 64))
		case bool:
			if x {
				b.WriteString("1")
			} else {
				b.WriteString("0")
			}
		default:
			buf, err := json.MarshalString(x)
			if err != nil {
				// NOTE: should not happen
				log.Errorf("err:%v", err)
			} else {
				b.WriteString(strconv.Quote(buf))
			}
		}
	}

	b.WriteString("CASE `")
	b.WriteString(field)
	b.WriteString("`")

	for k, v := range caseMap {
		b.WriteString(" WHEN ")
		appendAny(k)
		b.WriteString(" THEN ")
		appendAny(v)
	}

	if len(def) > 0 {
		b.WriteString(" ELSE ")
		appendAny(def[0])
	}

	b.WriteString(" END")

	return gorm.Expr(b.String())
}
