package db_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
	"gorm.io/gorm/clause"
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

	t.Run("with clause.Expr", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id IN (?)"
		expr := clause.Expr{SQL: "1, 2, 3", Vars: []interface{}{}}
		result := db.FormatSql(sql, expr)
		expected := "SELECT * FROM users WHERE id IN (1, 2, 3)"
		assert.Equal(t, expected, result)
	})

	t.Run("with clause.Expr with vars", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ?"
		expr := clause.Expr{SQL: "subquery_value", Vars: []interface{}{100, 200}}
		result := db.FormatSql(sql, expr)
		expected := "SELECT * FROM users WHERE id = subquery_value100200"
		assert.Equal(t, expected, result)
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

	t.Run("nil error", func(t *testing.T) {
		result := db.IsUniqueIndexConflictErr(nil)
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

func TestDecode(t *testing.T) {
	t.Run("int type", func(t *testing.T) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("123"))
		assert.NilError(t, err)
		assert.Equal(t, 123, result)
	})

	t.Run("uint type", func(t *testing.T) {
		var result uint
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("456"))
		assert.NilError(t, err)
		assert.Equal(t, uint(456), result)
	})

	t.Run("float type", func(t *testing.T) {
		var result float64
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("3.14"))
		assert.NilError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("string type", func(t *testing.T) {
		var result string
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("hello"))
		assert.NilError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("bool type - true", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("true"))
		assert.NilError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("bool type - 1", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("1"))
		assert.NilError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("bool type - false", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("false"))
		assert.NilError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("bool type - 0", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("0"))
		assert.NilError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("bool type - invalid", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("invalid"))
		assert.Error(t, err, "invalid bool value: invalid")
	})

	t.Run("int parse error", func(t *testing.T) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("not_a_number"))
		assert.Assert(t, err != nil)
	})

	t.Run("uint parse error", func(t *testing.T) {
		var result uint
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("-123"))
		assert.Assert(t, err != nil)
	})

	t.Run("float parse error", func(t *testing.T) {
		var result float64
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("not_a_float"))
		assert.Assert(t, err != nil)
	})
}

func TestScanComplexType(t *testing.T) {
	t.Run("struct type", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		var result TestStruct
		field := reflect.ValueOf(&result).Elem()
		jsonData := []byte(`{"name":"John","age":30}`)
		err := db.ExportScanComplexType(field, jsonData, false)
		assert.NilError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("slice type", func(t *testing.T) {
		var result []int
		field := reflect.ValueOf(&result).Elem()
		jsonData := []byte(`[1,2,3,4,5]`)
		err := db.ExportScanComplexType(field, jsonData, false)
		assert.NilError(t, err)
		assert.Equal(t, 5, len(result))
		assert.Equal(t, 1, result[0])
	})

	t.Run("map type", func(t *testing.T) {
		var result map[string]interface{}
		field := reflect.ValueOf(&result).Elem()
		jsonData := []byte(`{"key1":"value1","key2":123}`)
		err := db.ExportScanComplexType(field, jsonData, false)
		assert.NilError(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "value1", result["key1"])
	})

	t.Run("pointer type", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
		}
		var result *TestStruct
		field := reflect.ValueOf(&result).Elem()
		jsonData := []byte(`{"name":"Alice"}`)
		err := db.ExportScanComplexType(field, jsonData, true)
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Equal(t, "Alice", result.Name)
	})

	t.Run("invalid json", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
		}
		var result TestStruct
		field := reflect.ValueOf(&result).Elem()
		jsonData := []byte(`{invalid json`)
		err := db.ExportScanComplexType(field, jsonData, false)
		assert.Assert(t, err != nil)
	})
}

func TestGetTableName(t *testing.T) {
	t.Run("struct without Tabler interface", func(t *testing.T) {
		type SimpleStruct struct {
			ID   int64
			Name string
		}
		typ := reflect.TypeOf(SimpleStruct{})
		tableName := db.ExportGetTableName(typ)
		// Should return snake_case of struct name
		assert.Assert(t, tableName != "")
	})

	t.Run("struct with Tabler interface", func(t *testing.T) {
		typ := reflect.TypeOf(TestModel{})
		tableName := db.ExportGetTableName(typ)
		assert.Equal(t, "test_models", tableName)
	})

	t.Run("pointer type", func(t *testing.T) {
		typ := reflect.TypeOf(&TestModel{})
		tableName := db.ExportGetTableName(typ)
		assert.Equal(t, "test_models", tableName)
	})

	t.Run("cache hit", func(t *testing.T) {
		typ := reflect.TypeOf(TestModel{})
		// First call
		tableName1 := db.ExportGetTableName(typ)
		// Second call should use cache
		tableName2 := db.ExportGetTableName(typ)
		assert.Equal(t, tableName1, tableName2)
	})
}

func TestHasField(t *testing.T) {
	type TestStruct struct {
		ID        int64
		Name      string
		DeletedAt int64
		CreatedAt int64
	}

	t.Run("field exists", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := db.ExportHasField(typ, "Name")
		assert.Assert(t, result)
	})

	t.Run("field does not exist", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := db.ExportHasField(typ, "NonExistent")
		assert.Assert(t, !result)
	})

	t.Run("pointer type", func(t *testing.T) {
		typ := reflect.TypeOf(&TestStruct{})
		result := db.ExportHasField(typ, "Name")
		assert.Assert(t, result)
	})

	t.Run("DeletedAt field", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := db.ExportHasField(typ, "DeletedAt")
		assert.Assert(t, result)
	})

	t.Run("CreatedAt field", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := db.ExportHasField(typ, "CreatedAt")
		assert.Assert(t, result)
	})
}

func TestDecodeAllTypes(t *testing.T) {
	// Test all int types
	t.Run("int8", func(t *testing.T) {
		var result int8
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("127"))
		assert.NilError(t, err)
		assert.Equal(t, int8(127), result)
	})

	t.Run("int16", func(t *testing.T) {
		var result int16
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("32000"))
		assert.NilError(t, err)
		assert.Equal(t, int16(32000), result)
	})

	t.Run("int32", func(t *testing.T) {
		var result int32
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("2147483647"))
		assert.NilError(t, err)
		assert.Equal(t, int32(2147483647), result)
	})

	t.Run("int64", func(t *testing.T) {
		var result int64
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("9223372036854775807"))
		assert.NilError(t, err)
		assert.Equal(t, int64(9223372036854775807), result)
	})

	// Test all uint types
	t.Run("uint8", func(t *testing.T) {
		var result uint8
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("255"))
		assert.NilError(t, err)
		assert.Equal(t, uint8(255), result)
	})

	t.Run("uint16", func(t *testing.T) {
		var result uint16
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("65535"))
		assert.NilError(t, err)
		assert.Equal(t, uint16(65535), result)
	})

	t.Run("uint32", func(t *testing.T) {
		var result uint32
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("4294967295"))
		assert.NilError(t, err)
		assert.Equal(t, uint32(4294967295), result)
	})

	t.Run("uint64", func(t *testing.T) {
		var result uint64
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("18446744073709551615"))
		assert.NilError(t, err)
		assert.Equal(t, uint64(18446744073709551615), result)
	})

	// Test float types
	t.Run("float32", func(t *testing.T) {
		var result float32
		field := reflect.ValueOf(&result).Elem()
		err := db.ExportDecode(field, []byte("3.14"))
		assert.NilError(t, err)
		assert.Assert(t, result > 3.13 && result < 3.15)
	})
}