package db_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

func TestCond(t *testing.T) {
	assert.Equal(t, db.OrWhere(map[string]any{
		"a": 1,
	}, map[string]any{
		"a": 2,
	}, map[string]any{
		"a": 3,
	}).ToString(), "((a = 1) OR (a = 2) OR (a = 3))")

	assert.Equal(t, db.Where("a", 1).ToString(), "(a = 1)")

	assert.Equal(t, db.Or(db.Where("a", 1), db.Where("a", 2)).ToString(), "((a = 1) OR (a = 2))")

	// Test OrWhere with multiple conditions - order may vary due to map iteration
	result := db.OrWhere(db.Where(map[string]any{
		"a": 1,
		"b": 2,
	}), db.Where(map[string]any{
		"a": 2,
		"b": 3,
	})).ToString()

	// Check that result contains both expected condition groups in any order
	if !(strings.Contains(result, "(a = 1)") && strings.Contains(result, "(b = 2)")) {
		t.Errorf("Result should contain (a = 1) and (b = 2) conditions: %s", result)
	}
	if !(strings.Contains(result, "(a = 2)") && strings.Contains(result, "(b = 3)")) {
		t.Errorf("Result should contain (a = 2) and (b = 3) conditions: %s", result)
	}
	if !strings.Contains(result, " OR ") {
		t.Errorf("Result should contain OR operator: %s", result)
	}
}

func TestLike(t *testing.T) {
	t.Log(db.Where("name", "like", "%a%").ToString())
}

func TestIn(t *testing.T) {
	t.Log(db.Where("id", "in", []int{1, 2, 3}).ToString())
}

func TestQuote(t *testing.T) {
	t.Log(strconv.Quote("a"))
}

func TestGormTag(t *testing.T) {
	//tag := "column:id;primaryKey;autoIncrement;not null"
	//tag := "primaryKey;autoIncrement;not null"
	tag := "primaryKey"

	idx := strings.Index(tag, "primaryKey")
	t.Log(idx)
}

// TestToInterfacesFunction tests the toInterfaces function through condition building
func TestToInterfacesFunction(t *testing.T) {
	t.Run("slice condition that triggers toInterfaces", func(t *testing.T) {
		// Create a condition that will call toInterfaces internally
		// This happens when we pass a slice to Where()

		// Test with slice of strings (should trigger toInterfaces)
		cond := db.Where([]interface{}{"name", "=", "John"})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)

		// Test with nested slices
		cond2 := db.Where([]interface{}{
			[]interface{}{"name", "John"},
			[]interface{}{"age", 25},
		})
		result2 := cond2.ToString()
		assert.Assert(t, len(result2) > 0)
	})

	t.Run("array conditions", func(t *testing.T) {
		// Test with array that triggers slice handling
		arr := [3]interface{}{"field", "=", "value"}
		cond := db.Where(arr[:])
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})
}

// TestConditionCoverage tests more condition building scenarios
func TestConditionCoverage(t *testing.T) {
	t.Run("$raw command conditions", func(t *testing.T) {
		// Test $raw command which calls addCmdCond
		// $raw only accepts 2 arguments: command and condition
		cond := db.Where("$raw", "field = 123")
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)

		// Test $raw with slice (SQL with parameters)
		cond2 := db.Where("$raw", []interface{}{"field IN (1, 2, 3)"})
		result2 := cond2.ToString()
		assert.Assert(t, len(result2) > 0)
	})

	t.Run("$and and $or commands", func(t *testing.T) {
		// Test $and command
		cond := db.Where("$and", map[string]interface{}{
			"name": "John",
			"age":  25,
		})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)

		// Test $or command
		cond2 := db.Where("$or", map[string]interface{}{
			"name":  "John",
			"name2": "Jane",
		})
		result2 := cond2.ToString()
		assert.Assert(t, len(result2) > 0)
	})

	t.Run("boolean conditions", func(t *testing.T) {
		// Test boolean true condition (should add 1=1)
		cond := db.Where(true)
		result := cond.ToString()
		assert.Equal(t, "(1=1)", result)

		// Test boolean false condition (should add 1=0 and set skip)
		cond2 := db.Where(false)
		result2 := cond2.ToString()
		assert.Equal(t, "(1=0)", result2)
	})

	t.Run("character operators", func(t *testing.T) {
		// Test character operators (int32 to string conversion)
		cond := db.Where("field", '>', 100) // '>' as rune/int32
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)

		cond2 := db.Where("field", '<', 100) // '<' as rune/int32
		result2 := cond2.ToString()
		assert.Assert(t, len(result2) > 0)
	})
}

// TestSimpleTypeToStrCoverage tests different data types for simpleTypeToStr function
func TestSimpleTypeToStrCoverage(t *testing.T) {
	t.Run("different value types in conditions", func(t *testing.T) {
		// Test different types that call simpleTypeToStr
		testCases := map[string]interface{}{
			"string_field":  "test string",
			"bytes_field":   []byte("test bytes"),
			"bool_true":     true,
			"bool_false":    false,
			"int_field":     123,
			"int8_field":    int8(123),
			"int16_field":   int16(123),
			"int32_field":   int32(123),
			"int64_field":   int64(123),
			"uint_field":    uint(123),
			"uint8_field":   uint8(123),
			"uint16_field":  uint16(123),
			"uint32_field":  uint32(123),
			"uint64_field":  uint64(123),
			"float32_field": float32(1.23),
			"float64_field": float64(1.23),
			"slice_field":   []int{1, 2, 3},
			"array_field":   [3]int{1, 2, 3},
		}

		for field, value := range testCases {
			cond := db.Where(field, value)
			result := cond.ToString()
			assert.Assert(t, len(result) > 0, "Failed for field: %s", field)
		}
	})

	t.Run("slice with and without quotes", func(t *testing.T) {
		// Test slices that are quoted vs not quoted
		cond1 := db.Where("field", "IN", []int{1, 2, 3}) // Should be quoted
		result1 := cond1.ToString()
		assert.Assert(t, len(result1) > 0)

		cond2 := db.Where("field =", []int{1, 2, 3}) // Different context
		result2 := cond2.ToString()
		assert.Assert(t, len(result2) > 0)
	})
}

// TestFieldNameValidation tests field name character validation
func TestFieldNameValidation(t *testing.T) {
	t.Run("field names with special characters", func(t *testing.T) {
		// Test field names that trigger getFirstInvalidFieldNameCharIndex
		specialFields := []string{
			"field>value",      // Has '>' operator
			"field<value",      // Has '<' operator
			"field>=value",     // Has '>=' operator
			"field<=value",     // Has '<=' operator
			"field!=value",     // Has '!=' operator
			"field LIKE value", // Has 'LIKE' operator with space
			"field@invalid",    // Has invalid character '@'
			"field#invalid",    // Has invalid character '#'
		}

		for _, fieldName := range specialFields {
			cond := db.Where(fieldName, "test")
			result := cond.ToString()
			assert.Assert(t, len(result) > 0, "Failed for field: %s", fieldName)
		}
	})
}

// TestTablePrefixAndQuoting tests table prefix and field quoting
func TestTablePrefixAndQuoting(t *testing.T) {
	t.Run("field names with table prefixes", func(t *testing.T) {
		// Test conditions that should trigger table prefix logic
		cond := db.Where("users.name", "John")
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)

		cond2 := db.Where("`users`.`name`", "John") // Already quoted
		result2 := cond2.ToString()
		assert.Assert(t, len(result2) > 0)
	})
}

// TestCond_String 测试 String 方法
func TestCond_String(t *testing.T) {
	cond := db.Where("name", "John")
	str := cond.String()
	assert.Assert(t, len(str) > 0)
	assert.Equal(t, cond.ToString(), str)
}

// TestCond_GoString 测试 GoString 方法
func TestCond_GoString(t *testing.T) {
	cond := db.Where("name", "John")
	goStr := cond.GoString()
	assert.Assert(t, len(goStr) > 0)
	assert.Equal(t, cond.ToString(), goStr)
}

// TestCond_Or 测试 Or 方法
func TestCond_Or(t *testing.T) {
	// 使用 db.Or 创建 OR 条件
	cond := db.Or(map[string]interface{}{
		"name": "John",
		"age":  25,
	})
	result := cond.ToString()
	assert.Assert(t, len(result) > 0)
	assert.Assert(t, strings.Contains(result, "OR"))

	// 测试链式 Or 方法（会添加为 AND，因为是添加子条件）
	cond2 := db.Where("name", "John").Or("age", 25)
	result2 := cond2.ToString()
	assert.Assert(t, len(result2) > 0)
	// 链式 Or 实际上添加的是子条件，会用 AND 连接
	assert.Assert(t, strings.Contains(result2, "AND"))
}

// TestCond_Reset 测试 Reset 方法
func TestCond_Reset(t *testing.T) {
	cond := db.Where("name", "John").Where("age", 25)
	assert.Assert(t, len(cond.ToString()) > 0)

	cond.Reset()
	assert.Equal(t, "", cond.ToString())

	// 重置后应该可以重新使用
	cond.Where("email", "test@example.com")
	result := cond.ToString()
	assert.Assert(t, len(result) > 0)
	assert.Assert(t, strings.Contains(result, "email"))
}

// TestCond_ToInterfaces 测试 toInterfaces 函数的覆盖
func TestCond_ToInterfaces(t *testing.T) {
	t.Run("nested slice conditions", func(t *testing.T) {
		// 这会触发 toInterfaces 函数
		cond := db.Where([]interface{}{
			[]string{"name", "John"},
			[]string{"age", "=", "25"},
		})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("mixed slice conditions", func(t *testing.T) {
		cond := db.Where([]interface{}{
			[]interface{}{"field1", "value1"},
			[]interface{}{"field2", "=", "value2"},
			[]interface{}{"field3", ">", 100},
		})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})
}

// TestSimpleTypeToStrEdgeCases 测试 simpleTypeToStr 的边界情况
func TestSimpleTypeToStrEdgeCases(t *testing.T) {
	t.Run("float32 and float64 types", func(t *testing.T) {
		// 测试 float32 和 float64 类型
		// 这些类型会走到 default 分支，使用 fmt.Sprintf
		cond := db.Where("float32_field", float32(3.14))
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
		assert.Assert(t, strings.Contains(result, "float32_field"))

		cond2 := db.Where("float64_field", float64(2.718))
		result2 := cond2.ToString()
		assert.Assert(t, len(result2) > 0)
		assert.Assert(t, strings.Contains(result2, "float64_field"))
	})

	t.Run("interface type with string value", func(t *testing.T) {
		// 测试接口类型的具体值是字符串的情况
		var iface interface{} = "test_string"
		cond := db.Where("field", iface)
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("slice without quotes (quoteSlice=false)", func(t *testing.T) {
		// 在某些上下文中，slice 不应该被括号包裹
		// 这会在 addCond 中调用 simpleTypeToStr(val, true)
		// quoteSlice 参数为 true
		cond := db.Where("id", "IN", []int{1, 2, 3})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("slice with mixed types", func(t *testing.T) {
		// 测试包含混合类型的 slice
		cond := db.Where("field", []interface{}{1, "two", 3.0, true})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("array type", func(t *testing.T) {
		// 测试数组类型（不是 slice）
		cond := db.Where("field", [3]int{1, 2, 3})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("empty slice", func(t *testing.T) {
		// 测试空 slice
		cond := db.Where("field", []int{})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("pointer to basic type", func(t *testing.T) {
		// 测试指向基本类型的指针
		value := 42
		cond := db.Where("field", &value)
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("nil pointer", func(t *testing.T) {
		// 测试 nil 指针
		var ptr *int
		cond := db.Where("field", ptr)
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
		assert.Assert(t, strings.Contains(result, "NULL"))
	})

	t.Run("nested interface with nil", func(t *testing.T) {
		// 测试 interface{} 包含 nil 值
		var iface interface{} = nil
		cond := db.Where("field", iface)
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("uint types coverage", func(t *testing.T) {
		// 确保所有 uint 类型都被覆盖
		testCases := []struct {
			name  string
			value interface{}
		}{
			{"uint", uint(42)},
			{"uint8", uint8(255)},
			{"uint16", uint16(65535)},
			{"uint32", uint32(4294967295)},
			{"uint64", uint64(18446744073709551615)},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cond := db.Where("field", tc.value)
				result := cond.ToString()
				assert.Assert(t, len(result) > 0)
			})
		}
	})

	t.Run("int types coverage", func(t *testing.T) {
		// 确保所有 int 类型都被覆盖
		testCases := []struct {
			name  string
			value interface{}
		}{
			{"int", int(-42)},
			{"int8", int8(-128)},
			{"int16", int16(-32768)},
			{"int32", int32(-2147483648)},
			{"int64", int64(-9223372036854775808)},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cond := db.Where("field", tc.value)
				result := cond.ToString()
				assert.Assert(t, len(result) > 0)
			})
		}
	})
}

// TestWhereRawEdgeCases 测试 whereRaw 的边界情况
func TestWhereRawEdgeCases(t *testing.T) {
	t.Run("whereRaw with more placeholders than values", func(t *testing.T) {
		// 这会触发 whereRaw 中的警告日志
		cond := db.Where("field1 = ? AND field2 = ?", "value1")
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("whereRaw with more values than placeholders", func(t *testing.T) {
		// 多余的值会被忽略
		cond := db.Where("field = ?", "value1", "value2", "value3")
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})

	t.Run("whereRaw with no placeholders", func(t *testing.T) {
		// 纯 SQL 条件，没有参数
		cond := db.Where("field1 IS NOT NULL")
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
		assert.Assert(t, strings.Contains(result, "field1 IS NOT NULL"))
	})

	t.Run("whereRaw with complex values", func(t *testing.T) {
		// 测试复杂类型的值转换
		cond := db.Where("field IN (?)", []int{1, 2, 3})
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})
}

// TestExpr 测试 Expr 函数
func TestExpr(t *testing.T) {
	t.Run("create expression with args", func(t *testing.T) {
		// 测试带参数的表达式
		expr := db.Expr("age > ?", 18)
		sql := expr.SQL
		assert.Assert(t, len(sql) > 0)
	})

	t.Run("create expression without args", func(t *testing.T) {
		// 测试不带参数的表达式
		expr := db.Expr("NOW()")
		sql := expr.SQL
		assert.Assert(t, len(sql) > 0)
	})

	t.Run("create expression with multiple args", func(t *testing.T) {
		// 测试多个参数的表达式
		expr := db.Expr("field BETWEEN ? AND ?", 1, 100)
		sql := expr.SQL
		assert.Assert(t, len(sql) > 0)
	})
}
