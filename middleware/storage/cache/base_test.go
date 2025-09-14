package cache

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestBaseCacheExtensions(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Setup test data
	cache.Set("string_key", "123")
	cache.Set("bool_key", "true")
	cache.Set("int_key", "42")
	cache.Set("float_key", "3.14")
	cache.Set("slice_key", "[1,2,3]")

	// Test GetBool
	boolVal, err := cache.GetBool("bool_key")
	assert.NilError(t, err)
	assert.Equal(t, boolVal, true)

	// Test GetBool with invalid value (candy.ToBool doesn't return errors, uses defaults)
	cache.Set("invalid_bool", "not_a_bool")
	boolVal, err = cache.GetBool("invalid_bool")
	assert.NilError(t, err)
	assert.Equal(t, boolVal, true) // candy.ToBool("not_a_bool") returns true for non-empty strings

	// Test GetInt
	intVal, err := cache.GetInt("int_key")
	assert.NilError(t, err)
	assert.Equal(t, intVal, 42)

	// Test GetUint
	uintVal, err := cache.GetUint("int_key")
	assert.NilError(t, err)
	assert.Equal(t, uintVal, uint(42))

	// Test GetInt32
	int32Val, err := cache.GetInt32("int_key")
	assert.NilError(t, err)
	assert.Equal(t, int32Val, int32(42))

	// Test GetUint32
	uint32Val, err := cache.GetUint32("int_key")
	assert.NilError(t, err)
	assert.Equal(t, uint32Val, uint32(42))

	// Test GetInt64
	int64Val, err := cache.GetInt64("int_key")
	assert.NilError(t, err)
	assert.Equal(t, int64Val, int64(42))

	// Test GetUint64
	uint64Val, err := cache.GetUint64("int_key")
	assert.NilError(t, err)
	assert.Equal(t, uint64Val, uint64(42))

	// Test GetFloat32
	float32Val, err := cache.GetFloat32("float_key")
	assert.NilError(t, err)
	assert.Equal(t, float32Val, float32(3.14))

	// Test GetFloat64
	float64Val, err := cache.GetFloat64("float_key")
	assert.NilError(t, err)
	assert.Equal(t, float64Val, 3.14)

	// Test GetIntSlice
	intSlice, err := cache.GetIntSlice("slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(intSlice), 3)
	assert.Equal(t, intSlice[0], 1)
	assert.Equal(t, intSlice[1], 2)
	assert.Equal(t, intSlice[2], 3)

	// Test GetUintSlice
	uintSlice, err := cache.GetUintSlice("slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(uintSlice), 3)
	assert.Equal(t, uintSlice[0], uint(1))

	// Test GetInt32Slice
	int32Slice, err := cache.GetInt32Slice("slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(int32Slice), 3)
	assert.Equal(t, int32Slice[0], int32(1))

	// Test GetUint32Slice
	uint32Slice, err := cache.GetUint32Slice("slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(uint32Slice), 3)
	assert.Equal(t, uint32Slice[0], uint32(1))

	// Test GetInt64Slice
	int64Slice, err := cache.GetInt64Slice("slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(int64Slice), 3)
	assert.Equal(t, int64Slice[0], int64(1))

	// Test GetUint64Slice
	uint64Slice, err := cache.GetUint64Slice("slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(uint64Slice), 3)
	assert.Equal(t, uint64Slice[0], uint64(1))

	// Test GetFloat32Slice
	cache.Set("float_slice_key", "[1.1,2.2,3.3]")
	float32Slice, err := cache.GetFloat32Slice("float_slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(float32Slice), 3)
	assert.Equal(t, float32Slice[0], float32(1.1))

	// Test GetFloat64Slice
	float64Slice, err := cache.GetFloat64Slice("float_slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(float64Slice), 3)
	assert.Equal(t, float64Slice[0], 1.1)

	// Test GetBoolSlice
	cache.Set("bool_slice_key", "[true,false,true]")
	boolSlice, err := cache.GetBoolSlice("bool_slice_key")
	assert.NilError(t, err)
	assert.Equal(t, len(boolSlice), 3)
	assert.Equal(t, boolSlice[0], true)
	assert.Equal(t, boolSlice[1], false)
	assert.Equal(t, boolSlice[2], true)
}

func TestBaseCacheErrorCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test with non-existent keys
	_, err := cache.GetBool("nonexistent")
	assert.Equal(t, err, NotFound)

	_, err = cache.GetInt("nonexistent")
	assert.Equal(t, err, NotFound)

	_, err = cache.GetFloat32("nonexistent")
	assert.Equal(t, err, NotFound)

	_, err = cache.GetIntSlice("nonexistent")
	assert.Equal(t, err, NotFound)

	// Test with invalid numeric values (candy functions use defaults, don't return errors)
	cache.Set("invalid_int", "not_a_number")
	intVal, err := cache.GetInt("invalid_int")
	assert.NilError(t, err)
	assert.Equal(t, intVal, 0) // Default value

	floatVal, err := cache.GetFloat32("invalid_int")
	assert.NilError(t, err)
	assert.Equal(t, floatVal, float32(0)) // Default value

	// Test with invalid slice values (expects JSON format, not comma-separated)
	cache.Set("invalid_slice", "[1,\"not_a_number\",3]")
	intSlice, err := cache.GetIntSlice("invalid_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(intSlice), 3)
	assert.Equal(t, intSlice[0], 1)
	assert.Equal(t, intSlice[1], 0) // Default value for invalid number
	assert.Equal(t, intSlice[2], 3)
}

func TestBaseCacheSliceOperations(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test GetSlice (expects JSON format)
	cache.Set("slice_test", `["a","b","c"]`)
	slice, err := cache.GetSlice("slice_test")
	assert.NilError(t, err)
	assert.Equal(t, len(slice), 3)
	assert.Equal(t, slice[0], "a")
	assert.Equal(t, slice[1], "b")
	assert.Equal(t, slice[2], "c")

	// Test GetSlice with empty value
	cache.Set("empty_slice", "")
	emptySlice, err := cache.GetSlice("empty_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(emptySlice), 0) // Empty string returns nil slice

	// Test GetSlice with single value JSON
	cache.Set("single_slice", `["single"]`)
	singleSlice, err := cache.GetSlice("single_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(singleSlice), 1)
	assert.Equal(t, singleSlice[0], "single")
}

func TestBaseCacheJSON(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test struct for JSON serialization
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Test GetJson and SetJson would require protobuf support
	// These are currently not implemented in base.go and would return errors

	// Test HGetJson
	jsonStr := `{"name":"Jane","age":25}`
	cache.HSet("json_hash", "user", jsonStr)

	var result TestStruct
	err := cache.HGetJson("json_hash", "user", &result)
	assert.NilError(t, err)
	assert.Equal(t, result.Name, "Jane")
	assert.Equal(t, result.Age, 25)

	// Test HGetJson with non-existent key
	err = cache.HGetJson("nonexistent", "field", &result)
	assert.Equal(t, err, NotFound)

	// Test HGetJson with invalid JSON
	cache.HSet("bad_json", "field", "not valid json")
	err = cache.HGetJson("bad_json", "field", &result)
	assert.Assert(t, err != nil, "Should return an error for invalid JSON")
}

func TestBaseCacheLimit(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test rate limiting (Limit increments counter then checks)
	allowed, err := cache.Limit("test_limit", 2, 1*time.Second) // 2 requests per window
	assert.NilError(t, err)
	assert.Equal(t, allowed, true) // Count = 1, allowed

	allowed, err = cache.Limit("test_limit", 2, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true) // Count = 2, still allowed

	// Third request should be rate limited
	allowed, err = cache.Limit("test_limit", 2, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false) // Count = 3, exceeds limit of 2
}

func TestBaseCacheLimitUpdateOnCheck(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test LimitUpdateOnCheck
	allowed, err := cache.LimitUpdateOnCheck("check_limit", 1, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Second call should be limited
	allowed, err = cache.LimitUpdateOnCheck("check_limit", 1, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false)
}
