package db

import (
	"fmt"
	"reflect"
	"testing"
)

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