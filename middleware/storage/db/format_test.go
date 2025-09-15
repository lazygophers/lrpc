package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
	"gorm.io/gorm/clause"
)

// TestUpdateCaseAndFormatFunctions tests update case functions and formatting
func TestUpdateCaseAndFormatFunctions(t *testing.T) {
	t.Run("update case with different data types", func(t *testing.T) {
		// Test UpdateCase with different value types
		result := db.UpdateCase(map[string]interface{}{
			"(`status` = 1)":   true,
			"(`status` = 2)":   false,
			"(`status` = 3)":   123,
			"(`status` = 4)":   "active",
			"(`status` = 5)":   []byte("data"),
			"(`status` = 6)":   float64(1.5),
		}, "default")
		
		assert.Assert(t, result.SQL != "")
		assert.Assert(t, len(result.SQL) > 0)
	})
	
	t.Run("update case one field with different data types", func(t *testing.T) {
		// Test UpdateCaseOneField with different key and value types
		result := db.UpdateCaseOneField("status", map[any]any{
			1:     "active",
			2:     "inactive", 
			"3":   true,
			true:  false,
			1.5:   123,
		}, "unknown")
		
		assert.Assert(t, result.SQL != "")
		assert.Assert(t, len(result.SQL) > 0)
	})
	
	t.Run("format sql with clause expressions", func(t *testing.T) {
		// Test FormatSql with clause.Expr type
		expr := clause.Expr{
			SQL:  "SELECT ?",
			Vars: []interface{}{1},
		}
		
		result := db.FormatSql("WHERE id IN (?)", expr)
		assert.Assert(t, len(result) > 0)
	})
	
	t.Run("format sql with different value types", func(t *testing.T) {
		// Test FormatSql with various data types
		result := db.FormatSql("INSERT INTO test VALUES (?, ?, ?, ?, ?)", 
			123, "test", true, []byte("data"), nil)
		assert.Assert(t, len(result) > 0)
	})
}

// TestConditionEdgeCases tests condition building edge cases  
func TestConditionEdgeCases(t *testing.T) {
	t.Run("simpleTypeToStr with different types", func(t *testing.T) {
		// We can't test simpleTypeToStr directly since it's private,
		// but we can test it through condition building
		
		// Test with various data types that would call simpleTypeToStr
		cond := db.Where(map[string]interface{}{
			"int_field":    123,
			"string_field": "test",
			"bool_field":   true,
			"float_field":  1.23,
			"bytes_field":  []byte("test"),
			"slice_field":  []int{1, 2, 3},
		})
		
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})
	
	t.Run("condition with table prefix", func(t *testing.T) {
		// Test conditions that might use table prefixes
		cond := db.Where("table.field", "=", "value")
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
	})
	
	t.Run("addCond with different operators", func(t *testing.T) {
		// Test different comparison operators
		operators := []string{"=", ">", "<", ">=", "<=", "!=", "LIKE", "IN"}
		
		for _, op := range operators {
			cond := db.Where("field "+op, "value")
			result := cond.ToString()
			assert.Assert(t, len(result) > 0)
		}
	})
}

// TestGetFirstInvalidFieldNameCharIndex tests field name validation
func TestGetFirstInvalidFieldNameCharIndex(t *testing.T) {
	t.Run("field names with various characters", func(t *testing.T) {
		// Test field names that would trigger character validation
		// We test this indirectly through condition building
		
		testCases := []string{
			"simple_field",
			"field with space",
			"field.with.dots",
			"`quoted_field`",
			"field123",
			"FIELD_UPPER",
			"field>", // This should trigger the character index function
			"field<", 
			"field>=",
			"field<=",
		}
		
		for _, fieldName := range testCases {
			// Create conditions that would use these field names
			cond := db.Where(fieldName, "value")
			result := cond.ToString()
			assert.Assert(t, len(result) > 0)
		}
	})
}

// TestLoggerEdgeCases tests logger with different scenarios
func TestLoggerEdgeCases(t *testing.T) {
	t.Run("logger with different log levels", func(t *testing.T) {
		logger := db.NewLogger()
		
		// Test that LogMode returns a non-nil logger (already tested in main tests)
		newLogger := logger.LogMode(1)
		assert.Assert(t, newLogger != nil)
		
		// Test additional logger functionality
		newLogger2 := logger.LogMode(2)
		assert.Assert(t, newLogger2 != nil)
	})
}