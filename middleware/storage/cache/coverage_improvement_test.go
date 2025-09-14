package cache

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestBaseCacheTypesFullCoverage(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test all type conversion methods with actual data
	cache.Set("uint_test", "42")
	cache.Set("int32_test", "42")
	cache.Set("uint32_test", "42")
	cache.Set("int64_test", "42")
	cache.Set("uint64_test", "42")
	cache.Set("float64_test", "3.14")

	// Test GetUint
	uintVal, err := cache.GetUint("uint_test")
	assert.NilError(t, err)
	assert.Equal(t, uintVal, uint(42))

	// Test GetInt32
	int32Val, err := cache.GetInt32("int32_test")
	assert.NilError(t, err)
	assert.Equal(t, int32Val, int32(42))

	// Test GetUint32
	uint32Val, err := cache.GetUint32("uint32_test")
	assert.NilError(t, err)
	assert.Equal(t, uint32Val, uint32(42))

	// Test GetInt64
	int64Val, err := cache.GetInt64("int64_test")
	assert.NilError(t, err)
	assert.Equal(t, int64Val, int64(42))

	// Test GetUint64
	uint64Val, err := cache.GetUint64("uint64_test")
	assert.NilError(t, err)
	assert.Equal(t, uint64Val, uint64(42))

	// Test GetFloat64
	float64Val, err := cache.GetFloat64("float64_test")
	assert.NilError(t, err)
	assert.Equal(t, float64Val, 3.14)
}

func TestBaseCacheSlicesFullCoverage(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test all slice methods with actual data
	cache.Set("uint_slice", "[1,2,3]")
	cache.Set("int32_slice", "[1,2,3]")
	cache.Set("uint32_slice", "[1,2,3]")
	cache.Set("int64_slice", "[1,2,3]")
	cache.Set("uint64_slice", "[1,2,3]")
	cache.Set("float32_slice", "[1.1,2.2,3.3]")
	cache.Set("float64_slice", "[1.1,2.2,3.3]")

	// Test GetUintSlice
	uintSlice, err := cache.GetUintSlice("uint_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(uintSlice), 3)
	assert.Equal(t, uintSlice[0], uint(1))

	// Test GetInt32Slice
	int32Slice, err := cache.GetInt32Slice("int32_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(int32Slice), 3)
	assert.Equal(t, int32Slice[0], int32(1))

	// Test GetUint32Slice
	uint32Slice, err := cache.GetUint32Slice("uint32_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(uint32Slice), 3)
	assert.Equal(t, uint32Slice[0], uint32(1))

	// Test GetInt64Slice
	int64Slice, err := cache.GetInt64Slice("int64_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(int64Slice), 3)
	assert.Equal(t, int64Slice[0], int64(1))

	// Test GetUint64Slice
	uint64Slice, err := cache.GetUint64Slice("uint64_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(uint64Slice), 3)
	assert.Equal(t, uint64Slice[0], uint64(1))

	// Test GetFloat32Slice
	float32Slice, err := cache.GetFloat32Slice("float32_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(float32Slice), 3)
	assert.Equal(t, float32Slice[0], float32(1.1))

	// Test GetFloat64Slice
	float64Slice, err := cache.GetFloat64Slice("float64_slice")
	assert.NilError(t, err)
	assert.Equal(t, len(float64Slice), 3)
	assert.Equal(t, float64Slice[0], 1.1)

	// Test GetBoolSlice with more scenarios  
	cache.Set("bool_slice_errors", "[true,false,true]")
	boolSlice, err := cache.GetBoolSlice("bool_slice_errors")
	assert.NilError(t, err)
	assert.Equal(t, len(boolSlice), 3)
	assert.Equal(t, boolSlice[0], true)
	assert.Equal(t, boolSlice[1], false)
	assert.Equal(t, boolSlice[2], true)
}

func TestCacheMemEdgeCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Exists with multiple keys where some exist
	cache.Set("exists1", "value")
	cache.Set("exists2", "value")
	
	exists, err := cache.Exists("exists1", "exists2")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)
	
	// Test with one non-existent key
	exists, err = cache.Exists("exists1", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Test SRandMember with edge cases
	cache.SAdd("rand_set", "a", "b", "c", "d", "e")
	
	// Test with count = 0 (should return 1 member by default when count <= 0)
	members, err := cache.SRandMember("rand_set", 0)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1)

	// Test SisMember with non-existent set
	isMember, err := cache.SisMember("nonexistent_set", "member")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)
}

func TestMemCacheExpiration(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Expire on non-existent key
	success, err := cache.Expire("nonexistent_expire", 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Test Ttl on key with no expiration
	cache.Set("no_expire_ttl", "value")
	ttl, err := cache.Ttl("no_expire_ttl")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -1*time.Second)
}

func TestLimitRateLimiting(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Limit with increment failure scenario
	// This tests the error handling path in Limit
	// We can't easily trigger an Incr error in memory cache, but we can test normal flow

	// Test normal rate limiting
	allowed, err := cache.Limit("rate_limit_test", 1, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Second call should be limited
	allowed, err = cache.Limit("rate_limit_test", 1, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false)

	// Test LimitUpdateOnCheck
	allowed, err = cache.LimitUpdateOnCheck("update_check_test", 2, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	allowed, err = cache.LimitUpdateOnCheck("update_check_test", 2, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Third call should be limited
	allowed, err = cache.LimitUpdateOnCheck("update_check_test", 2, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false)
}

func TestGetSliceEmptyValue(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test GetSlice with empty value
	cache.Set("empty_slice_test", "")
	slice, err := cache.GetSlice("empty_slice_test")
	assert.NilError(t, err)
	assert.Assert(t, slice == nil)

	// Test GetSlice with JSON parsing error
	cache.Set("invalid_json_slice", "invalid json")
	_, err = cache.GetSlice("invalid_json_slice")
	assert.Assert(t, err != nil)
}

func TestMemCacheSetPrefix(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test SetPrefix (no-op for memory cache but should not panic)
	memCache := cache.(*baseCache).BaseCache.(*CacheMem)
	memCache.SetPrefix("test_prefix") // Should not panic
}