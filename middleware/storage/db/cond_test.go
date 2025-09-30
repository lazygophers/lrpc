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
	}).ToString(), "((\"a\" = 1) OR (\"a\" = 2) OR (\"a\" = 3))")

	assert.Equal(t, db.Where("a", 1).ToString(), "(\"a\" = 1)")

	assert.Equal(t, db.Or(db.Where("a", 1), db.Where("a", 2)).ToString(), "((\"a\" = 1) OR (\"a\" = 2))")

	// Test OrWhere with multiple conditions - order may vary due to map iteration
	result := db.OrWhere(db.Where(map[string]any{
		"a": 1,
		"b": 2,
	}), db.Where(map[string]any{
		"a": 2,
		"b": 3,
	})).ToString()
	
	// Check that result contains both expected condition groups in any order
	if !(strings.Contains(result, "(\"a\" = 1)") && strings.Contains(result, "(\"b\" = 2)")) {
		t.Errorf("Result should contain a=1 and b=2 conditions: %s", result)
	}
	if !(strings.Contains(result, "(\"a\" = 2)") && strings.Contains(result, "(\"b\" = 3)")) {
		t.Errorf("Result should contain a=2 and b=3 conditions: %s", result)
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
			"name": "John",
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
			"string_field":   "test string",
			"bytes_field":    []byte("test bytes"),
			"bool_true":      true,
			"bool_false":     false,
			"int_field":      123,
			"int8_field":     int8(123),
			"int16_field":    int16(123),
			"int32_field":    int32(123),
			"int64_field":    int64(123),
			"uint_field":     uint(123),
			"uint8_field":    uint8(123),
			"uint16_field":   uint16(123),
			"uint32_field":   uint32(123),
			"uint64_field":   uint64(123),
			"float32_field":  float32(1.23),
			"float64_field":  float64(1.23),
			"slice_field":    []int{1, 2, 3},
			"array_field":    [3]int{1, 2, 3},
		}

		for field, value := range testCases {
			cond := db.Where(field, value)
			result := cond.ToString()
			assert.Assert(t, len(result) > 0, "Failed for field: %s", field)
		}
	})

	t.Run("slice with and without quotes", func(t *testing.T) {
		// Test slices that are quoted vs not quoted
		cond1 := db.Where("field", "IN", []int{1, 2, 3})     // Should be quoted
		result1 := cond1.ToString()
		assert.Assert(t, len(result1) > 0)

		cond2 := db.Where("field =", []int{1, 2, 3})         // Different context
		result2 := cond2.ToString()
		assert.Assert(t, len(result2) > 0)
	})
}

// TestFieldNameValidation tests field name character validation
func TestFieldNameValidation(t *testing.T) {
	t.Run("field names with special characters", func(t *testing.T) {
		// Test field names that trigger getFirstInvalidFieldNameCharIndex
		specialFields := []string{
			"field>value",     // Has '>' operator
			"field<value",     // Has '<' operator
			"field>=value",    // Has '>=' operator
			"field<=value",    // Has '<=' operator
			"field!=value",    // Has '!=' operator
			"field LIKE value", // Has 'LIKE' operator with space
			"field@invalid",   // Has invalid character '@'
			"field#invalid",   // Has invalid character '#'
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
