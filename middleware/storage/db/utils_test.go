package db_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

func TestEnsureIsSliceOrArray(t *testing.T) {
	t.Run("valid slice", func(t *testing.T) {
		slice := []int{1, 2, 3}
		result := db.EnsureIsSliceOrArray(slice)
		assert.Equal(t, reflect.Slice, result.Kind())
		assert.Equal(t, 3, result.Len())
	})
	
	t.Run("valid array", func(t *testing.T) {
		array := [3]string{"a", "b", "c"}
		result := db.EnsureIsSliceOrArray(array)
		assert.Equal(t, reflect.Array, result.Kind())
		assert.Equal(t, 3, result.Len())
	})
	
	t.Run("pointer to slice", func(t *testing.T) {
		slice := []int{1, 2, 3}
		ptr := &slice
		result := db.EnsureIsSliceOrArray(ptr)
		assert.Equal(t, reflect.Slice, result.Kind())
		assert.Equal(t, 3, result.Len())
	})
	
	t.Run("invalid type - string", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				assert.Assert(t, r != nil)
			} else {
				t.Error("Expected panic for non-slice/array type")
			}
		}()
		db.EnsureIsSliceOrArray("not a slice")
	})
	
	t.Run("invalid type - map", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				assert.Assert(t, r != nil)
			} else {
				t.Error("Expected panic for map type")
			}
		}()
		db.EnsureIsSliceOrArray(map[string]int{"key": 1})
	})
}

func TestEscapeMysqlString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with'quote", "with\\'quote"},
		{"with\"doublequote", "with\\\"doublequote"},
		{"with\\backslash", "with\\\\backslash"},
		{"with\nnewline", "with\\nnewline"},
		{"with\rcarriagereturn", "with\\rcarriagereturn"},
		{"with\x00null", "with\\0null"},
		{"with\x1asubtitute", "with\\Zsubtitute"},
		{"", ""},
		{"mixed'\"\\content\n", "mixed\\'\\\"\\\\content\\n"},
	}
	
	for _, tc := range testCases {
		result := db.EscapeMysqlString(tc.input)
		assert.Equal(t, tc.expected, result, fmt.Sprintf("Input: %q", tc.input))
	}
}

func TestUniqueSlice(t *testing.T) {
	t.Run("int slice with duplicates", func(t *testing.T) {
		input := []int{1, 2, 3, 2, 4, 1, 5}
		result := db.UniqueSlice(input)
		
		resultSlice := result.([]int)
		assert.Equal(t, 5, len(resultSlice))
		
		// Check no duplicates
		seen := make(map[int]bool)
		for _, v := range resultSlice {
			assert.Assert(t, !seen[v], fmt.Sprintf("Duplicate value found: %d", v))
			seen[v] = true
		}
	})
	
	t.Run("string slice with duplicates", func(t *testing.T) {
		input := []string{"a", "b", "a", "c", "b"}
		result := db.UniqueSlice(input)
		
		resultSlice := result.([]string)
		assert.Equal(t, 3, len(resultSlice))
	})
	
	t.Run("slice with less than 2 elements", func(t *testing.T) {
		input := []int{1}
		result := db.UniqueSlice(input)
		
		resultSlice := result.([]int)
		assert.Equal(t, 1, len(resultSlice))
		assert.Equal(t, 1, resultSlice[0])
	})
	
	t.Run("empty slice", func(t *testing.T) {
		input := []int{}
		result := db.UniqueSlice(input)
		
		resultSlice := result.([]int)
		assert.Equal(t, 0, len(resultSlice))
	})
	
	t.Run("slice with no duplicates", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := db.UniqueSlice(input)
		
		resultSlice := result.([]int)
		assert.Equal(t, 5, len(resultSlice))
	})
	
	t.Run("invalid type - not slice", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				assert.Assert(t, r != nil)
			} else {
				t.Error("Expected panic for non-slice type")
			}
		}()
		db.UniqueSlice("not a slice")
	})
}

func TestCamel2UnderScore(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"CamelCase", "camel_case"},
		{"XMLParser", "x_mlparser"},
		{"HTTPSConnection", "h_ttpsconnection"},
		{"ID", "i_d"},
		{"UserID", "user_id"},
		{"HTMLToText", "h_tmlto_text"},
		{"simple", "simple"},
		{"", ""},
		{"A", "a"},
		{"AB", "a_b"},
		{"ABC", "a_bc"},
		{"AbC", "ab_c"},
		{"APIKey", "a_pikey"},
	}
	
	for _, tc := range testCases {
		result := db.Camel2UnderScore(tc.input)
		assert.Equal(t, tc.expected, result, fmt.Sprintf("Input: %s", tc.input))
	}
}

func TestFormatSql(t *testing.T) {
	t.Run("no placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users"
		result := db.FormatSql(sql)
		assert.Equal(t, sql, result)
	})
	
	t.Run("with placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ? AND name = ?"
		result := db.FormatSql(sql, 123, "John")
		expected := "SELECT * FROM users WHERE id = 123 AND name = John"
		assert.Equal(t, expected, result)
	})
	
	t.Run("more placeholders than values", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ? AND name = ? AND age = ?"
		result := db.FormatSql(sql, 123)
		expected := "SELECT * FROM users WHERE id = 123 AND name = ? AND age = ?"
		assert.Equal(t, expected, result)
	})
	
	t.Run("more values than placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ?"
		result := db.FormatSql(sql, 123, "extra", "values")
		expected := "SELECT * FROM users WHERE id = 123"
		assert.Equal(t, expected, result)
	})
	
	t.Run("empty sql", func(t *testing.T) {
		result := db.FormatSql("", 123)
		assert.Equal(t, "", result)
	})
}

func TestIsUniqueIndexConflictErr(t *testing.T) {
	t.Run("mysql duplicate entry error", func(t *testing.T) {
		err := fmt.Errorf("Error 1062: Duplicate entry 'test' for key 'unique_key'")
		result := db.IsUniqueIndexConflictErr(err)
		assert.Assert(t, result)
	})
	
	t.Run("mysql duplicate entry error without error code", func(t *testing.T) {
		err := fmt.Errorf("Duplicate entry 'test' for key 'unique_key'")
		result := db.IsUniqueIndexConflictErr(err)
		assert.Assert(t, result)
	})
	
	t.Run("different error", func(t *testing.T) {
		err := fmt.Errorf("Table 'test.users' doesn't exist")
		result := db.IsUniqueIndexConflictErr(err)
		assert.Assert(t, !result)
	})
	
	t.Run("empty error", func(t *testing.T) {
		err := fmt.Errorf("")
		result := db.IsUniqueIndexConflictErr(err)
		assert.Assert(t, !result)
	})
}

func TestErrBatchesStop(t *testing.T) {
	assert.Equal(t, "batches stop", db.ErrBatchesStop.Error())
}

// Test type definitions and interfaces
type TestModel struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"column:name"`
}

func (TestModel) TableName() string {
	return "test_models"
}

// Note: getTableName and hasXXX functions are private and cannot be tested directly