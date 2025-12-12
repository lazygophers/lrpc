package db

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm/clause"
	"gotest.tools/v3/assert"
)

// ============================================================================
// Unit Tests
// ============================================================================

func TestEnsureIsSliceOrArray(t *testing.T) {
	t.Run("valid slice", func(t *testing.T) {
		slice := []int{1, 2, 3}
		result := EnsureIsSliceOrArray(slice)
		assert.Equal(t, reflect.Slice, result.Kind())
		assert.Equal(t, 3, result.Len())
	})

	t.Run("valid array", func(t *testing.T) {
		array := [3]string{"a", "b", "c"}
		result := EnsureIsSliceOrArray(array)
		assert.Equal(t, reflect.Array, result.Kind())
		assert.Equal(t, 3, result.Len())
	})

	t.Run("pointer to slice", func(t *testing.T) {
		slice := []int{1, 2, 3}
		ptr := &slice
		result := EnsureIsSliceOrArray(ptr)
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
		EnsureIsSliceOrArray("not a slice")
	})

	t.Run("invalid type - map", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				assert.Assert(t, r != nil)
			} else {
				t.Error("Expected panic for map type")
			}
		}()
		EnsureIsSliceOrArray(map[string]int{"key": 1})
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
		result := EscapeMysqlString(tc.input)
		assert.Equal(t, tc.expected, result, fmt.Sprintf("Input: %q", tc.input))
	}
}

func TestUniqueSlice(t *testing.T) {
	t.Run("int slice with duplicates", func(t *testing.T) {
		input := []int{1, 2, 3, 2, 4, 1, 5}
		result := UniqueSlice(input)

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
		result := UniqueSlice(input)

		resultSlice := result.([]string)
		assert.Equal(t, 3, len(resultSlice))
	})

	t.Run("slice with less than 2 elements", func(t *testing.T) {
		input := []int{1}
		result := UniqueSlice(input)

		resultSlice := result.([]int)
		assert.Equal(t, 1, len(resultSlice))
		assert.Equal(t, 1, resultSlice[0])
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []int{}
		result := UniqueSlice(input)

		resultSlice := result.([]int)
		assert.Equal(t, 0, len(resultSlice))
	})

	t.Run("slice with no duplicates", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := UniqueSlice(input)

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
		UniqueSlice("not a slice")
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
		result := Camel2UnderScore(tc.input)
		assert.Equal(t, tc.expected, result, fmt.Sprintf("Input: %s", tc.input))
	}
}

func TestFormatSql(t *testing.T) {
	t.Run("no placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users"
		result := FormatSql(sql)
		assert.Equal(t, sql, result)
	})

	t.Run("with placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ? AND name = ?"
		result := FormatSql(sql, 123, "John")
		expected := "SELECT * FROM users WHERE id = 123 AND name = John"
		assert.Equal(t, expected, result)
	})

	t.Run("more placeholders than values", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ? AND name = ? AND age = ?"
		result := FormatSql(sql, 123)
		expected := "SELECT * FROM users WHERE id = 123 AND name = ? AND age = ?"
		assert.Equal(t, expected, result)
	})

	t.Run("more values than placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ?"
		result := FormatSql(sql, 123, "extra", "values")
		expected := "SELECT * FROM users WHERE id = 123"
		assert.Equal(t, expected, result)
	})

	t.Run("empty sql", func(t *testing.T) {
		result := FormatSql("", 123)
		assert.Equal(t, "", result)
	})

	t.Run("with clause.Expr", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id IN (?)"
		expr := clause.Expr{SQL: "1, 2, 3", Vars: []interface{}{}}
		result := FormatSql(sql, expr)
		expected := "SELECT * FROM users WHERE id IN (1, 2, 3)"
		assert.Equal(t, expected, result)
	})

	t.Run("with clause.Expr with vars", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ?"
		expr := clause.Expr{SQL: "subquery_value", Vars: []interface{}{100, 200}}
		result := FormatSql(sql, expr)
		expected := "SELECT * FROM users WHERE id = subquery_value100200"
		assert.Equal(t, expected, result)
	})
}

func TestIsUniqueIndexConflictErr(t *testing.T) {
	t.Run("mysql duplicate entry error", func(t *testing.T) {
		err := fmt.Errorf("Error 1062: Duplicate entry 'test' for key 'unique_key'")
		result := IsUniqueIndexConflictErr(err)
		assert.Assert(t, result)
	})

	t.Run("mysql duplicate entry error without error code", func(t *testing.T) {
		err := fmt.Errorf("Duplicate entry 'test' for key 'unique_key'")
		result := IsUniqueIndexConflictErr(err)
		assert.Assert(t, result)
	})

	t.Run("different error", func(t *testing.T) {
		err := fmt.Errorf("Table 'test.users' doesn't exist")
		result := IsUniqueIndexConflictErr(err)
		assert.Assert(t, !result)
	})

	t.Run("empty error", func(t *testing.T) {
		err := fmt.Errorf("")
		result := IsUniqueIndexConflictErr(err)
		assert.Assert(t, !result)
	})

	t.Run("nil error", func(t *testing.T) {
		result := IsUniqueIndexConflictErr(nil)
		assert.Assert(t, !result)
	})
}

func TestErrBatchesStop(t *testing.T) {
	assert.Equal(t, "batches stop", ErrBatchesStop.Error())
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
		err := decode(field, []byte("123"))
		assert.NilError(t, err)
		assert.Equal(t, 123, result)
	})

	t.Run("uint type", func(t *testing.T) {
		var result uint
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("456"))
		assert.NilError(t, err)
		assert.Equal(t, uint(456), result)
	})

	t.Run("float type", func(t *testing.T) {
		var result float64
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("3.14"))
		assert.NilError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("string type", func(t *testing.T) {
		var result string
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("hello"))
		assert.NilError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("bool type - true", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("true"))
		assert.NilError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("bool type - 1", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("1"))
		assert.NilError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("bool type - false", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("false"))
		assert.NilError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("bool type - 0", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("0"))
		assert.NilError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("bool type - invalid", func(t *testing.T) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("invalid"))
		assert.Error(t, err, "invalid bool value: invalid")
	})

	t.Run("int parse error", func(t *testing.T) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("not_a_number"))
		assert.Assert(t, err != nil)
	})

	t.Run("uint parse error", func(t *testing.T) {
		var result uint
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("-123"))
		assert.Assert(t, err != nil)
	})

	t.Run("float parse error", func(t *testing.T) {
		var result float64
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("not_a_float"))
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
		err := scanComplexType(field, jsonData, false)
		assert.NilError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("slice type", func(t *testing.T) {
		var result []int
		field := reflect.ValueOf(&result).Elem()
		jsonData := []byte(`[1,2,3,4,5]`)
		err := scanComplexType(field, jsonData, false)
		assert.NilError(t, err)
		assert.Equal(t, 5, len(result))
		assert.Equal(t, 1, result[0])
	})

	t.Run("map type", func(t *testing.T) {
		var result map[string]interface{}
		field := reflect.ValueOf(&result).Elem()
		jsonData := []byte(`{"key1":"value1","key2":123}`)
		err := scanComplexType(field, jsonData, false)
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
		err := scanComplexType(field, jsonData, true)
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
		err := scanComplexType(field, jsonData, false)
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
		tableName := getTableName(typ)
		// Should return snake_case of struct name
		assert.Assert(t, tableName != "")
	})

	t.Run("struct with Tabler interface", func(t *testing.T) {
		typ := reflect.TypeOf(TestModel{})
		tableName := getTableName(typ)
		assert.Equal(t, "test_models", tableName)
	})

	t.Run("pointer type", func(t *testing.T) {
		typ := reflect.TypeOf(&TestModel{})
		tableName := getTableName(typ)
		assert.Equal(t, "test_models", tableName)
	})

	t.Run("cache hit", func(t *testing.T) {
		typ := reflect.TypeOf(TestModel{})
		// First call
		tableName1 := getTableName(typ)
		// Second call should use cache
		tableName2 := getTableName(typ)
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
		result := hasField(typ, "Name")
		assert.Assert(t, result)
	})

	t.Run("field does not exist", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasField(typ, "NonExistent")
		assert.Assert(t, !result)
	})

	t.Run("pointer type", func(t *testing.T) {
		typ := reflect.TypeOf(&TestStruct{})
		result := hasField(typ, "Name")
		assert.Assert(t, result)
	})

	t.Run("DeletedAt field", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasField(typ, "DeletedAt")
		assert.Assert(t, result)
	})

	t.Run("CreatedAt field", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasField(typ, "CreatedAt")
		assert.Assert(t, result)
	})
}

func TestDecodeAllTypes(t *testing.T) {
	// Test all int types
	t.Run("int8", func(t *testing.T) {
		var result int8
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("127"))
		assert.NilError(t, err)
		assert.Equal(t, int8(127), result)
	})

	t.Run("int16", func(t *testing.T) {
		var result int16
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("32000"))
		assert.NilError(t, err)
		assert.Equal(t, int16(32000), result)
	})

	t.Run("int32", func(t *testing.T) {
		var result int32
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("2147483647"))
		assert.NilError(t, err)
		assert.Equal(t, int32(2147483647), result)
	})

	t.Run("int64", func(t *testing.T) {
		var result int64
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("9223372036854775807"))
		assert.NilError(t, err)
		assert.Equal(t, int64(9223372036854775807), result)
	})

	// Test all uint types
	t.Run("uint8", func(t *testing.T) {
		var result uint8
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("255"))
		assert.NilError(t, err)
		assert.Equal(t, uint8(255), result)
	})

	t.Run("uint16", func(t *testing.T) {
		var result uint16
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("65535"))
		assert.NilError(t, err)
		assert.Equal(t, uint16(65535), result)
	})

	t.Run("uint32", func(t *testing.T) {
		var result uint32
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("4294967295"))
		assert.NilError(t, err)
		assert.Equal(t, uint32(4294967295), result)
	})

	t.Run("uint64", func(t *testing.T) {
		var result uint64
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("18446744073709551615"))
		assert.NilError(t, err)
		assert.Equal(t, uint64(18446744073709551615), result)
	})

	// Test float types
	t.Run("float32", func(t *testing.T) {
		var result float32
		field := reflect.ValueOf(&result).Elem()
		err := decode(field, []byte("3.14"))
		assert.NilError(t, err)
		assert.Assert(t, result > 3.13 && result < 3.15)
	})
}

// Test wrapper functions that are one-line calls to hasField

func TestHasDeletedAt(t *testing.T) {
	type TestStruct struct {
		DeletedAt int64
	}

	t.Run("has DeletedAt", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasDeletedAt(typ)
		assert.Assert(t, result)
	})

	t.Run("no DeletedAt", func(t *testing.T) {
		type NoDeletedAt struct {
			ID int64
		}
		typ := reflect.TypeOf(NoDeletedAt{})
		result := hasDeletedAt(typ)
		assert.Assert(t, !result)
	})
}

func TestHasCreatedAt(t *testing.T) {
	type TestStruct struct {
		CreatedAt int64
	}

	t.Run("has CreatedAt", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasCreatedAt(typ)
		assert.Assert(t, result)
	})

	t.Run("no CreatedAt", func(t *testing.T) {
		type NoCreatedAt struct {
			ID int64
		}
		typ := reflect.TypeOf(NoCreatedAt{})
		result := hasCreatedAt(typ)
		assert.Assert(t, !result)
	})
}

func TestHasUpdatedAt(t *testing.T) {
	type TestStruct struct {
		UpdatedAt int64
	}

	t.Run("has UpdatedAt", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasUpdatedAt(typ)
		assert.Assert(t, result)
	})

	t.Run("no UpdatedAt", func(t *testing.T) {
		type NoUpdatedAt struct {
			ID int64
		}
		typ := reflect.TypeOf(NoUpdatedAt{})
		result := hasUpdatedAt(typ)
		assert.Assert(t, !result)
	})
}

func TestHasId(t *testing.T) {
	type TestStruct struct {
		Id int64
	}

	t.Run("has Id", func(t *testing.T) {
		typ := reflect.TypeOf(TestStruct{})
		result := hasId(typ)
		assert.Assert(t, result)
	})

	t.Run("no Id", func(t *testing.T) {
		type NoId struct {
			Name string
		}
		typ := reflect.TypeOf(NoId{})
		result := hasId(typ)
		assert.Assert(t, !result)
	})
}

// ============================================================================
// Concurrent Tests
// ============================================================================

// TestConcurrentGetTableName tests thread safety of getTableName with caching
func TestConcurrentGetTableName(t *testing.T) {
	type TestModel1 struct {
		ID   int64
		Name string
	}
	type TestModel2 struct {
		ID    int64
		Email string
	}

	typ1 := reflect.TypeOf(TestModel1{})
	typ2 := reflect.TypeOf(TestModel2{})

	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Test concurrent access to cached types
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if id%2 == 0 {
					_ = getTableName(typ1)
				} else {
					_ = getTableName(typ2)
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentHasField tests thread safety of hasField with caching
func TestConcurrentHasField(t *testing.T) {
	type TestStruct struct {
		ID        int64
		Name      string
		DeletedAt int64
		CreatedAt int64
		UpdatedAt int64
	}

	typ := reflect.TypeOf(TestStruct{})

	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	fields := []string{"ID", "Name", "DeletedAt", "CreatedAt", "UpdatedAt", "NonExistent"}

	// Test concurrent access to cached field lookups
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				fieldName := fields[j%len(fields)]
				_ = hasField(typ, fieldName)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentDecode tests thread safety of decode function
func TestConcurrentDecode(t *testing.T) {
	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Test concurrent decoding of different types
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				switch j % 5 {
				case 0: // int
					var result int
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("123"))
				case 1: // string
					var result string
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("test"))
				case 2: // bool
					var result bool
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("1"))
				case 3: // float64
					var result float64
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("3.14"))
				case 4: // uint64
					var result uint64
					field := reflect.ValueOf(&result).Elem()
					_ = decode(field, []byte("999"))
				}
			}
		}(i)
	}

	wg.Wait()
}

// Test race conditions with -race flag
func TestRaceConditionGetTableName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	type RaceModel struct {
		ID int64
	}
	typ := reflect.TypeOf(RaceModel{})

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = getTableName(typ)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRaceConditionHasField(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	type RaceStruct struct {
		ID   int64
		Name string
	}
	typ := reflect.TypeOf(RaceStruct{})

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = hasField(typ, "Name")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Stress test with high concurrency
func TestHighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	type StressModel struct {
		ID        int64
		Name      string
		DeletedAt int64
		CreatedAt int64
		UpdatedAt int64
	}
	typ := reflect.TypeOf(StressModel{})

	const goroutines = 1000
	const iterations = 10000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Mix different operations
				switch (id + j) % 5 {
				case 0:
					_ = getTableName(typ)
				case 1:
					_ = hasField(typ, "DeletedAt")
				case 2:
					_ = hasDeletedAt(typ)
				case 3:
					_ = hasCreatedAt(typ)
				case 4:
					_ = hasUpdatedAt(typ)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		t.Fatalf("Encountered %d errors during stress test", len(errors))
	}
}

// Verify cache effectiveness
func TestCacheEffectiveness(t *testing.T) {
	type CacheTestModel struct {
		ID int64
	}
	typ := reflect.TypeOf(CacheTestModel{})

	// Clear caches
	tableNameCacheMu.Lock()
	tableNameCache = make(map[reflect.Type]string)
	tableNameCacheMu.Unlock()

	hasFieldCacheMu.Lock()
	hasFieldCache = make(map[fieldCacheKey]bool)
	hasFieldCacheMu.Unlock()

	// First call - cache miss
	name1 := getTableName(typ)

	// Second call - should hit cache
	name2 := getTableName(typ)

	if name1 != name2 {
		t.Errorf("Cache returned different values: %s vs %s", name1, name2)
	}

	// Verify cache was populated
	tableNameCacheMu.RLock()
	_, exists := tableNameCache[typ]
	tableNameCacheMu.RUnlock()

	if !exists {
		t.Error("Cache was not populated after getTableName call")
	}

	// Test hasField cache
	has1 := hasField(typ, "ID")
	has2 := hasField(typ, "ID")

	if has1 != has2 {
		t.Errorf("hasField cache returned different values: %v vs %v", has1, has2)
	}

	key := fieldCacheKey{typ: typ, fieldName: "ID"}
	hasFieldCacheMu.RLock()
	_, exists = hasFieldCache[key]
	hasFieldCacheMu.RUnlock()

	if !exists {
		t.Error("hasField cache was not populated")
	}
}

// Test concurrent cache growth
func TestConcurrentCacheGrowth(t *testing.T) {
	const numTypes = 100
	const goroutines = 50

	// Create many different types
	types := make([]reflect.Type, numTypes)
	for i := 0; i < numTypes; i++ {
		// Create a struct type with a unique field
		structFields := []reflect.StructField{
			{Name: fmt.Sprintf("Field%d", i), Type: reflect.TypeOf(int64(0))},
		}
		types[i] = reflect.StructOf(structFields)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numTypes; j++ {
				typ := types[(id+j)%numTypes]
				_ = getTableName(typ)
				_ = hasField(typ, fmt.Sprintf("Field%d", (id+j)%numTypes))
			}
		}(i)
	}

	wg.Wait()

	// Verify all types were cached
	tableNameCacheMu.RLock()
	cacheSize := len(tableNameCache)
	tableNameCacheMu.RUnlock()

	if cacheSize == 0 {
		t.Error("Cache should contain entries")
	}

	t.Logf("Cache contains %d entries after concurrent operations", cacheSize)
}

// ============================================================================
// Benchmark Tests
// ============================================================================

// Benchmark for EscapeMysqlString - tests the optimized lookup table approach

func BenchmarkEscapeMysqlString_NoEscape(b *testing.B) {
	sql := "SELECT * FROM users WHERE id = 123 AND name = test"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EscapeMysqlString(sql)
	}
}

func BenchmarkEscapeMysqlString_WithEscape(b *testing.B) {
	sql := "SELECT * FROM users WHERE name = 'test' AND desc = \"description\""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EscapeMysqlString(sql)
	}
}

func BenchmarkEscapeMysqlString_MixedContent(b *testing.B) {
	sql := "INSERT INTO logs VALUES ('error', \"message\nwith\nnewlines\", 'quote\\'test')"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EscapeMysqlString(sql)
	}
}

func BenchmarkEscapeMysqlString_LargeString(b *testing.B) {
	// Simulate large SQL query with 1000 characters
	sql := ""
	for i := 0; i < 20; i++ {
		sql += "SELECT * FROM table WHERE field = 'value' AND "
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EscapeMysqlString(sql)
	}
}

// Benchmark for UniqueSlice - tests reflection-based deduplication

func BenchmarkUniqueSlice_IntSmall(b *testing.B) {
	input := []int{1, 2, 3, 2, 4, 1, 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UniqueSlice(input)
	}
}

func BenchmarkUniqueSlice_IntLarge(b *testing.B) {
	input := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i % 100 // 10x duplicates
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UniqueSlice(input)
	}
}

func BenchmarkUniqueSlice_StringSmall(b *testing.B) {
	input := []string{"a", "b", "c", "b", "d", "a", "e"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UniqueSlice(input)
	}
}

func BenchmarkUniqueSlice_StringLarge(b *testing.B) {
	input := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = string(rune('a' + (i % 26))) // 26 unique values
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UniqueSlice(input)
	}
}

func BenchmarkUniqueSlice_NoDuplicates(b *testing.B) {
	input := make([]int, 100)
	for i := 0; i < 100; i++ {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UniqueSlice(input)
	}
}

// Benchmark for Camel2UnderScore - tests string transformation

func BenchmarkCamel2UnderScore_Simple(b *testing.B) {
	input := "CamelCase"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Camel2UnderScore(input)
	}
}

func BenchmarkCamel2UnderScore_NoUpperCase(b *testing.B) {
	input := "alllowercase"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Camel2UnderScore(input)
	}
}

func BenchmarkCamel2UnderScore_Consecutive(b *testing.B) {
	input := "HTTPSConnection"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Camel2UnderScore(input)
	}
}

func BenchmarkCamel2UnderScore_Long(b *testing.B) {
	input := "VeryLongCamelCaseStringWithManyWordsToConvert"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Camel2UnderScore(input)
	}
}

// Benchmark for FormatSql - tests SQL placeholder replacement

func BenchmarkFormatSql_NoPlaceholders(b *testing.B) {
	sql := "SELECT * FROM users WHERE id = 123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatSql(sql)
	}
}

func BenchmarkFormatSql_SinglePlaceholder(b *testing.B) {
	sql := "SELECT * FROM users WHERE id = ?"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatSql(sql, 123)
	}
}

func BenchmarkFormatSql_MultiplePlaceholders(b *testing.B) {
	sql := "SELECT * FROM users WHERE id = ? AND name = ? AND age = ? AND city = ?"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatSql(sql, 123, "John", 30, "NYC")
	}
}

func BenchmarkFormatSql_ManyPlaceholders(b *testing.B) {
	sql := "INSERT INTO table VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	values := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatSql(sql, values...)
	}
}

// Benchmark for decode - tests type conversion performance

func BenchmarkDecode_Int(b *testing.B) {
	var result int
	field := reflect.ValueOf(&result).Elem()
	data := []byte("12345")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decode(field, data)
	}
}

func BenchmarkDecode_Int64(b *testing.B) {
	var result int64
	field := reflect.ValueOf(&result).Elem()
	data := []byte("9223372036854775807")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decode(field, data)
	}
}

func BenchmarkDecode_String(b *testing.B) {
	var result string
	field := reflect.ValueOf(&result).Elem()
	data := []byte("test string value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decode(field, data)
	}
}

func BenchmarkDecode_Bool(b *testing.B) {
	var result bool
	field := reflect.ValueOf(&result).Elem()
	data := []byte("true")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decode(field, data)
	}
}

func BenchmarkDecode_Float64(b *testing.B) {
	var result float64
	field := reflect.ValueOf(&result).Elem()
	data := []byte("3.141592653589793")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decode(field, data)
	}
}

func BenchmarkDecode_Uint64(b *testing.B) {
	var result uint64
	field := reflect.ValueOf(&result).Elem()
	data := []byte("18446744073709551615")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decode(field, data)
	}
}

// Benchmark for scanComplexType - tests JSON deserialization

func BenchmarkScanComplexType_Struct(b *testing.B) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var result TestStruct
	field := reflect.ValueOf(&result).Elem()
	data := []byte(`{"name":"John","age":30}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scanComplexType(field, data, false)
	}
}

func BenchmarkScanComplexType_Slice(b *testing.B) {
	var result []int
	field := reflect.ValueOf(&result).Elem()
	data := []byte(`[1,2,3,4,5,6,7,8,9,10]`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scanComplexType(field, data, false)
	}
}

func BenchmarkScanComplexType_Map(b *testing.B) {
	var result map[string]interface{}
	field := reflect.ValueOf(&result).Elem()
	data := []byte(`{"key1":"value1","key2":123,"key3":true}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scanComplexType(field, data, false)
	}
}

func BenchmarkScanComplexType_LargeStruct(b *testing.B) {
	type LargeStruct struct {
		Field1  string `json:"field1"`
		Field2  int    `json:"field2"`
		Field3  bool   `json:"field3"`
		Field4  string `json:"field4"`
		Field5  int    `json:"field5"`
		Field6  bool   `json:"field6"`
		Field7  string `json:"field7"`
		Field8  int    `json:"field8"`
		Field9  bool   `json:"field9"`
		Field10 string `json:"field10"`
	}
	var result LargeStruct
	field := reflect.ValueOf(&result).Elem()
	data := []byte(`{"field1":"value1","field2":123,"field3":true,"field4":"value4","field5":456,"field6":false,"field7":"value7","field8":789,"field9":true,"field10":"value10"}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scanComplexType(field, data, false)
	}
}

// Benchmark for getTableName - tests caching effectiveness

func BenchmarkGetTableName_FirstCall(b *testing.B) {
	type BenchModel struct {
		ID   int64
		Name string
	}
	typ := reflect.TypeOf(BenchModel{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache to simulate first call
		tableNameCache = make(map[reflect.Type]string)
		_ = getTableName(typ)
	}
}

func BenchmarkGetTableName_CachedCall(b *testing.B) {
	type CachedModel struct {
		ID   int64
		Name string
	}
	typ := reflect.TypeOf(CachedModel{})

	// Warm up cache
	_ = getTableName(typ)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getTableName(typ)
	}
}

func BenchmarkGetTableName_WithTabler(b *testing.B) {
	typ := reflect.TypeOf(TestModel{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tableNameCache = make(map[reflect.Type]string)
		_ = getTableName(typ)
	}
}

// Benchmark for hasField - tests field existence checking

func BenchmarkHasField_Exists(b *testing.B) {
	type TestStruct struct {
		ID        int64
		Name      string
		DeletedAt int64
		CreatedAt int64
		UpdatedAt int64
	}
	typ := reflect.TypeOf(TestStruct{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasField(typ, "Name")
	}
}

func BenchmarkHasField_NotExists(b *testing.B) {
	type TestStruct struct {
		ID   int64
		Name string
	}
	typ := reflect.TypeOf(TestStruct{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasField(typ, "NonExistent")
	}
}

func BenchmarkHasDeletedAt(b *testing.B) {
	type TestStruct struct {
		DeletedAt int64
	}
	typ := reflect.TypeOf(TestStruct{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasDeletedAt(typ)
	}
}

func BenchmarkHasCreatedAt(b *testing.B) {
	type TestStruct struct {
		CreatedAt int64
	}
	typ := reflect.TypeOf(TestStruct{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasCreatedAt(typ)
	}
}

func BenchmarkHasUpdatedAt(b *testing.B) {
	type TestStruct struct {
		UpdatedAt int64
	}
	typ := reflect.TypeOf(TestStruct{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasUpdatedAt(typ)
	}
}

func BenchmarkHasId(b *testing.B) {
	type TestStruct struct {
		Id int64
	}
	typ := reflect.TypeOf(TestStruct{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasId(typ)
	}
}

// Benchmark for EnsureIsSliceOrArray - tests reflection validation

func BenchmarkEnsureIsSliceOrArray_Slice(b *testing.B) {
	slice := []int{1, 2, 3, 4, 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EnsureIsSliceOrArray(slice)
	}
}

func BenchmarkEnsureIsSliceOrArray_Array(b *testing.B) {
	array := [5]int{1, 2, 3, 4, 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EnsureIsSliceOrArray(array)
	}
}

func BenchmarkEnsureIsSliceOrArray_PointerToSlice(b *testing.B) {
	slice := []int{1, 2, 3, 4, 5}
	ptr := &slice
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EnsureIsSliceOrArray(ptr)
	}
}

// Benchmark for IsUniqueIndexConflictErr - tests error checking

func BenchmarkIsUniqueIndexConflictErr_True(b *testing.B) {
	err := fmt.Errorf("Error 1062: Duplicate entry 'test' for key 'unique_key'")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsUniqueIndexConflictErr(err)
	}
}

func BenchmarkIsUniqueIndexConflictErr_False(b *testing.B) {
	err := fmt.Errorf("Table 'test.users' doesn't exist")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsUniqueIndexConflictErr(err)
	}
}

func BenchmarkIsUniqueIndexConflictErr_Nil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsUniqueIndexConflictErr(nil)
	}
}

// Benchmark concurrent getTableName
func BenchmarkConcurrentGetTableName(b *testing.B) {
	type BenchModel struct {
		ID   int64
		Name string
	}
	typ := reflect.TypeOf(BenchModel{})

	// Warm up cache
	_ = getTableName(typ)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = getTableName(typ)
		}
	})
}

// Benchmark concurrent getTableName with multiple types
func BenchmarkConcurrentGetTableName_MultipleTypes(b *testing.B) {
	types := make([]reflect.Type, 10)
	for i := 0; i < 10; i++ {
		// Create distinct types dynamically
		switch i {
		case 0:
			type T0 struct{ ID int64 }
			types[i] = reflect.TypeOf(T0{})
		case 1:
			type T1 struct{ ID int64 }
			types[i] = reflect.TypeOf(T1{})
		case 2:
			type T2 struct{ ID int64 }
			types[i] = reflect.TypeOf(T2{})
		case 3:
			type T3 struct{ ID int64 }
			types[i] = reflect.TypeOf(T3{})
		case 4:
			type T4 struct{ ID int64 }
			types[i] = reflect.TypeOf(T4{})
		case 5:
			type T5 struct{ ID int64 }
			types[i] = reflect.TypeOf(T5{})
		case 6:
			type T6 struct{ ID int64 }
			types[i] = reflect.TypeOf(T6{})
		case 7:
			type T7 struct{ ID int64 }
			types[i] = reflect.TypeOf(T7{})
		case 8:
			type T8 struct{ ID int64 }
			types[i] = reflect.TypeOf(T8{})
		case 9:
			type T9 struct{ ID int64 }
			types[i] = reflect.TypeOf(T9{})
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = getTableName(types[i%len(types)])
			i++
		}
	})
}

// Benchmark concurrent hasField
func BenchmarkConcurrentHasField(b *testing.B) {
	type TestStruct struct {
		ID        int64
		Name      string
		DeletedAt int64
	}
	typ := reflect.TypeOf(TestStruct{})

	// Warm up cache
	_ = hasField(typ, "DeletedAt")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = hasField(typ, "DeletedAt")
		}
	})
}

// Benchmark concurrent hasField with cache misses
func BenchmarkConcurrentHasField_MultipleFields(b *testing.B) {
	type TestStruct struct {
		Field1  string
		Field2  int
		Field3  bool
		Field4  float64
		Field5  uint64
		Field6  string
		Field7  int
		Field8  bool
		Field9  float64
		Field10 uint64
	}
	typ := reflect.TypeOf(TestStruct{})

	fields := []string{"Field1", "Field2", "Field3", "Field4", "Field5",
		"Field6", "Field7", "Field8", "Field9", "Field10"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = hasField(typ, fields[i%len(fields)])
			i++
		}
	})
}

// Benchmark concurrent decode
func BenchmarkConcurrentDecode_Int(b *testing.B) {
	data := []byte("12345")

	b.RunParallel(func(pb *testing.PB) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		for pb.Next() {
			_ = decode(field, data)
		}
	})
}

func BenchmarkConcurrentDecode_String(b *testing.B) {
	data := []byte("test string value")

	b.RunParallel(func(pb *testing.PB) {
		var result string
		field := reflect.ValueOf(&result).Elem()
		for pb.Next() {
			_ = decode(field, data)
		}
	})
}

func BenchmarkConcurrentDecode_Bool(b *testing.B) {
	data := []byte("1")

	b.RunParallel(func(pb *testing.PB) {
		var result bool
		field := reflect.ValueOf(&result).Elem()
		for pb.Next() {
			_ = decode(field, data)
		}
	})
}

func BenchmarkConcurrentDecode_Float64(b *testing.B) {
	data := []byte("3.141592653589793")

	b.RunParallel(func(pb *testing.PB) {
		var result float64
		field := reflect.ValueOf(&result).Elem()
		for pb.Next() {
			_ = decode(field, data)
		}
	})
}

// Benchmark concurrent operations with mixed types
func BenchmarkConcurrentMixedOperations(b *testing.B) {
	type TestModel struct {
		ID        int64
		Name      string
		DeletedAt int64
	}
	typ := reflect.TypeOf(TestModel{})

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 3 {
			case 0:
				_ = getTableName(typ)
			case 1:
				_ = hasField(typ, "DeletedAt")
			case 2:
				var result int
				field := reflect.ValueOf(&result).Elem()
				_ = decode(field, []byte("123"))
			}
			i++
		}
	})
}

// Benchmark to compare before/after optimization
func BenchmarkGetTableName_Serial(b *testing.B) {
	type SerialModel struct {
		ID int64
	}
	typ := reflect.TypeOf(SerialModel{})

	// Warm up
	_ = getTableName(typ)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getTableName(typ)
	}
}

func BenchmarkHasField_Serial(b *testing.B) {
	type SerialStruct struct {
		ID        int64
		DeletedAt int64
	}
	typ := reflect.TypeOf(SerialStruct{})

	// Warm up
	_ = hasField(typ, "DeletedAt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasField(typ, "DeletedAt")
	}
}

func BenchmarkUnsafeString(b *testing.B) {
	data := []byte("test string for unsafe conversion")

	b.Run("unsafeString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = unsafeString(data)
		}
	})

	b.Run("string()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = string(data)
		}
	})
}

func BenchmarkDecode_WithAndWithoutUnsafe(b *testing.B) {
	data := []byte("12345")

	b.Run("decode_int_optimized", func(b *testing.B) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = decode(field, data)
		}
	})
}

// Benchmark memory allocations
func BenchmarkDecodeAllocations(b *testing.B) {
	data := []byte("12345")

	b.Run("int_with_unsafe", func(b *testing.B) {
		var result int
		field := reflect.ValueOf(&result).Elem()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = decode(field, data)
		}
	})

	b.Run("string_with_copy", func(b *testing.B) {
		var result string
		field := reflect.ValueOf(&result).Elem()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = decode(field, data)
		}
	})

	b.Run("bool_optimized", func(b *testing.B) {
		data := []byte("1")
		var result bool
		field := reflect.ValueOf(&result).Elem()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = decode(field, data)
		}
	})
}
