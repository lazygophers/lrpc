package cache

import (
	"testing"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)

func TestBaseCacheProtobufErrorPaths(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test SetPb with marshaling error - using invalid message
	// This is hard to trigger, but let's test normal flow to reach the Set() error path
	// In memory cache, Set() shouldn't fail, so we test the success path
	testMsg := &timestamppb.Timestamp{
		Seconds: 1234567890,
		Nanos:   123456789,
	}
	err := cache.SetPb("proto_test", testMsg)
	assert.NilError(t, err)

	// Test SetPbEx with marshaling error - using invalid message
	// Similar to above, testing normal flow
	err = cache.SetPbEx("proto_test_ex", testMsg, 1)
	assert.NilError(t, err)
}

func TestBaseCacheTypeConversionsErrorPaths(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test all type conversion methods with non-existent keys to cover error paths

	// Test GetUint error path
	_, err := cache.GetUint("nonexistent_uint")
	assert.Equal(t, err, NotFound)

	// Test GetInt32 error path
	_, err = cache.GetInt32("nonexistent_int32")
	assert.Equal(t, err, NotFound)

	// Test GetUint32 error path
	_, err = cache.GetUint32("nonexistent_uint32")
	assert.Equal(t, err, NotFound)

	// Test GetInt64 error path
	_, err = cache.GetInt64("nonexistent_int64")
	assert.Equal(t, err, NotFound)

	// Test GetUint64 error path
	_, err = cache.GetUint64("nonexistent_uint64")
	assert.Equal(t, err, NotFound)

	// Test GetFloat64 error path
	_, err = cache.GetFloat64("nonexistent_float64")
	assert.Equal(t, err, NotFound)
}

func TestBaseCacheSliceErrorPaths(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test slice methods with non-existent keys to cover GetJson error paths

	// Test GetBoolSlice error path
	_, err := cache.GetBoolSlice("nonexistent_bool_slice")
	assert.Equal(t, err, NotFound)

	// Test GetUintSlice error path
	_, err = cache.GetUintSlice("nonexistent_uint_slice")
	assert.Equal(t, err, NotFound)

	// Test GetInt32Slice error path
	_, err = cache.GetInt32Slice("nonexistent_int32_slice")
	assert.Equal(t, err, NotFound)

	// Test GetUint32Slice error path
	_, err = cache.GetUint32Slice("nonexistent_uint32_slice")
	assert.Equal(t, err, NotFound)

	// Test GetInt64Slice error path
	_, err = cache.GetInt64Slice("nonexistent_int64_slice")
	assert.Equal(t, err, NotFound)

	// Test GetUint64Slice error path
	_, err = cache.GetUint64Slice("nonexistent_uint64_slice")
	assert.Equal(t, err, NotFound)

	// Test GetFloat32Slice error path
	_, err = cache.GetFloat32Slice("nonexistent_float32_slice")
	assert.Equal(t, err, NotFound)

	// Test GetFloat64Slice error path
	_, err = cache.GetFloat64Slice("nonexistent_float64_slice")
	assert.Equal(t, err, NotFound)
}

func TestMemCacheCompleteExists(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Exists with all non-existent keys
	exists, err := cache.Exists("nonexistent1", "nonexistent2", "nonexistent3")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)
}

func TestMemCacheSetPrefixCoverage(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test SetPrefix method - it's a no-op but should be tested for coverage
	memCache := cache.(*baseCache).BaseCache.(*CacheMem)
	memCache.SetPrefix("test_prefix_coverage")
}

func TestMemCacheTtlVariations(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test TTL with key that has expiration - use longer duration
	cache.SetEx("ttl_test_key", "value", 10000) // 10 seconds
	ttl, err := cache.Ttl("ttl_test_key")
	assert.NilError(t, err)
	assert.Assert(t, ttl > 0, "TTL should be positive for non-expired key")

	// Clean up expired key to reach the expired key branch
	// We can't easily test the expired key branch in TTL without complex timing,
	// but we can test the normal flow
}

func TestSRandMemberEdgeCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Create a set
	cache.SAdd("edge_set", "a", "b", "c")

	// Test SRandMember with negative count (should still return 1 member)
	members, err := cache.SRandMember("edge_set", -1)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1)
}
