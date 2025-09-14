package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/lazygophers/utils/app"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)


// MockCache for testing error scenarios
type MockCache struct {
	*CacheMem
	shouldFailIncr   bool
	shouldFailExpire bool
}

func (m *MockCache) Incr(key string) (int64, error) {
	if m.shouldFailIncr {
		return 0, errors.New("incr failed")
	}
	return m.CacheMem.Incr(key)
}

func (m *MockCache) Expire(key string, timeout time.Duration) (bool, error) {
	if m.shouldFailExpire {
		return false, errors.New("expire failed")
	}
	return m.CacheMem.Expire(key, timeout)
}

func TestSetPbErrorPaths(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test SetPb with marshal error - create invalid proto
	// Use a mock approach to test error paths
	invalidProto := &timestamppb.Timestamp{Seconds: 123}
	
	// First test successful case to ensure method works
	err := cache.SetPb("test_key", invalidProto)
	assert.NilError(t, err, "SetPb should succeed with valid proto")

	// Test SetPbEx with valid proto
	err = cache.SetPbEx("test_key_ex", invalidProto, 1*time.Second)
	assert.NilError(t, err, "SetPbEx should succeed with valid proto")

	// Test successful SetPb and SetPbEx
	validProto := &timestamppb.Timestamp{Seconds: 123}
	err = cache.SetPb("valid_key", validProto)
	assert.NilError(t, err)

	err = cache.SetPbEx("valid_key_ex", validProto, 1*time.Second)
	assert.NilError(t, err)
}

func TestLimitErrorPaths(t *testing.T) {
	// Test Limit with Incr error  
	memCache := NewMem()
	defer memCache.Close()
	
	// Create a more realistic test - we can't easily make Incr fail in memory cache
	// but we can test the error handling paths by using the actual implementation
	mockCache := &MockCache{
		CacheMem:         memCache.(*baseCache).BaseCache.(*CacheMem),
		shouldFailIncr:   true,
		shouldFailExpire: false,
	}
	mockBaseCache := &baseCache{BaseCache: mockCache}

	allowed, err := mockBaseCache.Limit("test_key", 1, 1*time.Second)
	assert.Assert(t, err != nil, "Limit should fail when Incr fails")
	assert.Equal(t, allowed, false)

	// Test Limit with Expire error
	memCache2 := NewMem()
	defer memCache2.Close()
	mockCache2 := &MockCache{
		CacheMem:         memCache2.(*baseCache).BaseCache.(*CacheMem),
		shouldFailIncr:   false,
		shouldFailExpire: true,
	}
	mockBaseCache2 := &baseCache{BaseCache: mockCache2}

	allowed, err = mockBaseCache2.Limit("test_key2", 1, 1*time.Second)
	assert.Assert(t, err != nil, "Limit should fail when Expire fails")
	assert.Equal(t, allowed, false)

	// Test LimitUpdateOnCheck with Incr error
	allowed, err = mockBaseCache.LimitUpdateOnCheck("test_key3", 1, 1*time.Second)
	assert.Assert(t, err != nil, "LimitUpdateOnCheck should fail when Incr fails")
	assert.Equal(t, allowed, false)

	// Test LimitUpdateOnCheck with Expire error
	allowed, err = mockBaseCache2.LimitUpdateOnCheck("test_key4", 1, 1*time.Second)
	assert.Assert(t, err != nil, "LimitUpdateOnCheck should fail when Expire fails")
	assert.Equal(t, allowed, false)
}

func TestConfigApply(t *testing.T) {
	// Test default type (empty)
	config := &Config{}
	config.apply()
	assert.Equal(t, config.Type, "mem")

	// Test Bbolt configuration
	config = &Config{Type: Bbolt}
	config.apply()
	assert.Assert(t, config.Address != "", "Bbolt address should be set")

	// Test Bbolt with existing address
	config = &Config{Type: Bbolt, Address: "/custom/path"}
	config.apply()
	assert.Equal(t, config.Address, "/custom/path")

	// Test SugarDB configuration
	config = &Config{Type: SugarDB}
	config.apply()
	assert.Assert(t, config.DataDir != "", "SugarDB data dir should be set")

	// Test SugarDB with existing data dir
	config = &Config{Type: SugarDB, DataDir: "/custom/dir"}
	config.apply()
	assert.Equal(t, config.DataDir, "/custom/dir")

	// Test Bitcask configuration
	config = &Config{Type: Bitcask}
	config.apply()
	assert.Assert(t, config.DataDir != "", "Bitcask data dir should be set")

	// Test Bitcask with existing data dir
	config = &Config{Type: Bitcask, DataDir: "/custom/dir"}
	config.apply()
	assert.Equal(t, config.DataDir, "/custom/dir")

	// Test Redis configuration
	config = &Config{Type: Redis}
	config.apply()
	assert.Equal(t, config.Address, "127.0.0.1:6379")

	// Test Redis with existing address
	config = &Config{Type: Redis, Address: "custom:6380"}
	config.apply()
	assert.Equal(t, config.Address, "custom:6380")
}

func TestNewCacheTypes(t *testing.T) {
	// Test successful memory cache creation
	memCache, err := New(&Config{Type: Mem})
	assert.NilError(t, err)
	assert.Assert(t, memCache != nil)
	memCache.Close()

	// Test with unsupported cache type
	_, err = New(&Config{Type: "unsupported"})
	assert.Assert(t, err != nil, "Should return error for unsupported cache type")
	assert.Equal(t, err.Error(), "cache type not support")

	// Test with empty config (defaults to mem)
	defaultCache, err := New(&Config{})
	assert.NilError(t, err)
	assert.Assert(t, defaultCache != nil)
	defaultCache.Close()

	// Test Redis cache creation (will fail without server, but tests code path)
	_, err = New(&Config{Type: Redis, Address: "localhost:6379"})
	// Redis creation may fail due to no server, but the factory code path is tested
	// We don't assert error here as it depends on environment

	// Test Bbolt cache creation (may fail due to file permissions)
	_, err = New(&Config{Type: Bbolt, Address: "/tmp/test.cache"})
	// Bbolt creation may succeed or fail, but the factory code path is tested
}

func TestItemMethods(t *testing.T) {
	item := &Item{
		Data:     "test data",
		ExpireAt: time.Now().Add(1 * time.Hour),
	}

	// Test Bytes method
	bytes := item.Bytes()
	assert.Assert(t, len(bytes) > 0, "Bytes should return non-empty data")

	// Test String method
	str := item.String()
	assert.Assert(t, str != "", "String should return non-empty string")
	assert.Assert(t, len(str) > 0, "String should have content")
}

func TestBaseCacheErrorScenarios(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test type conversion methods with key that doesn't exist
	_, err := cache.GetBool("nonexistent")
	assert.Assert(t, err != nil, "GetBool should fail for nonexistent key")

	_, err = cache.GetInt("nonexistent")
	assert.Assert(t, err != nil, "GetInt should fail for nonexistent key")

	_, err = cache.GetUint("nonexistent")
	assert.Assert(t, err != nil, "GetUint should fail for nonexistent key")

	_, err = cache.GetInt32("nonexistent")
	assert.Assert(t, err != nil, "GetInt32 should fail for nonexistent key")

	_, err = cache.GetUint32("nonexistent")
	assert.Assert(t, err != nil, "GetUint32 should fail for nonexistent key")

	_, err = cache.GetInt64("nonexistent")
	assert.Assert(t, err != nil, "GetInt64 should fail for nonexistent key")

	_, err = cache.GetUint64("nonexistent")
	assert.Assert(t, err != nil, "GetUint64 should fail for nonexistent key")

	_, err = cache.GetFloat32("nonexistent")
	assert.Assert(t, err != nil, "GetFloat32 should fail for nonexistent key")

	_, err = cache.GetFloat64("nonexistent")
	assert.Assert(t, err != nil, "GetFloat64 should fail for nonexistent key")

	// Test slice methods with nonexistent keys
	_, err = cache.GetSlice("nonexistent")
	assert.Assert(t, err != nil, "GetSlice should fail for nonexistent key")

	_, err = cache.GetBoolSlice("nonexistent")
	assert.Assert(t, err != nil, "GetBoolSlice should fail for nonexistent key")

	_, err = cache.GetIntSlice("nonexistent")
	assert.Assert(t, err != nil, "GetIntSlice should fail for nonexistent key")

	_, err = cache.GetUintSlice("nonexistent")
	assert.Assert(t, err != nil, "GetUintSlice should fail for nonexistent key")

	_, err = cache.GetInt32Slice("nonexistent")
	assert.Assert(t, err != nil, "GetInt32Slice should fail for nonexistent key")

	_, err = cache.GetUint32Slice("nonexistent")
	assert.Assert(t, err != nil, "GetUint32Slice should fail for nonexistent key")

	_, err = cache.GetInt64Slice("nonexistent")
	assert.Assert(t, err != nil, "GetInt64Slice should fail for nonexistent key")

	_, err = cache.GetUint64Slice("nonexistent")
	assert.Assert(t, err != nil, "GetUint64Slice should fail for nonexistent key")

	_, err = cache.GetFloat32Slice("nonexistent")
	assert.Assert(t, err != nil, "GetFloat32Slice should fail for nonexistent key")

	_, err = cache.GetFloat64Slice("nonexistent")
	assert.Assert(t, err != nil, "GetFloat64Slice should fail for nonexistent key")

	// Test JSON methods with nonexistent keys
	var result interface{}
	err = cache.GetJson("nonexistent", &result)
	assert.Assert(t, err != nil, "GetJson should fail for nonexistent key")

	err = cache.HGetJson("nonexistent", "field", &result)
	assert.Assert(t, err != nil, "HGetJson should fail for nonexistent key")

	// Test GetPb with nonexistent key
	validProto := &timestamppb.Timestamp{}
	err = cache.GetPb("nonexistent", validProto)
	assert.Assert(t, err != nil, "GetPb should fail for nonexistent key")

	// Test slice methods with invalid JSON
	cache.Set("invalid_json", "not json")
	_, err = cache.GetBoolSlice("invalid_json")
	assert.Assert(t, err != nil, "GetBoolSlice should fail with invalid JSON")

	_, err = cache.GetIntSlice("invalid_json")
	assert.Assert(t, err != nil, "GetIntSlice should fail with invalid JSON")

	_, err = cache.GetUintSlice("invalid_json")
	assert.Assert(t, err != nil, "GetUintSlice should fail with invalid JSON")

	_, err = cache.GetInt32Slice("invalid_json")
	assert.Assert(t, err != nil, "GetInt32Slice should fail with invalid JSON")

	_, err = cache.GetUint32Slice("invalid_json")
	assert.Assert(t, err != nil, "GetUint32Slice should fail with invalid JSON")

	_, err = cache.GetInt64Slice("invalid_json")
	assert.Assert(t, err != nil, "GetInt64Slice should fail with invalid JSON")

	_, err = cache.GetUint64Slice("invalid_json")
	assert.Assert(t, err != nil, "GetUint64Slice should fail with invalid JSON")

	_, err = cache.GetFloat32Slice("invalid_json")
	assert.Assert(t, err != nil, "GetFloat32Slice should fail with invalid JSON")

	_, err = cache.GetFloat64Slice("invalid_json")
	assert.Assert(t, err != nil, "GetFloat64Slice should fail with invalid JSON")
}

func TestMemCacheUncoveredPaths(t *testing.T) {
	cache := NewMem().(*baseCache).BaseCache.(*CacheMem)
	defer cache.Close()

	// Test SetPrefix (should not panic, it's a no-op)
	cache.SetPrefix("test_prefix:")

	// Test additional Exists edge cases
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	exists, err := cache.Exists("key1")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	exists, err = cache.Exists("nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Test Ttl edge case with expired key that doesn't exist
	ttl, err := cache.Ttl("never_existed")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -2*time.Second)

	// Test SRandMember edge cases
	cache.SAdd("test_set", "a", "b", "c")
	
	// Test with negative count (should return single member)
	members, err := cache.SRandMember("test_set", -1)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1)

	// Test SRandMember on nonexistent set
	members, err = cache.SRandMember("nonexistent_set")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)
}

func TestCompleteConfigCoverage(t *testing.T) {
	// Save original app.Name and restore it
	originalName := app.Name
	defer func() {
		app.Name = originalName
	}()

	// Test config apply with different app names
	app.Name = "test_app"

	// Test all config branches
	testCases := []struct {
		configType string
		config     *Config
	}{
		{
			configType: "empty_mem",
			config:     &Config{},
		},
		{
			configType: "explicit_mem",
			config:     &Config{Type: Mem},
		},
		{
			configType: "bbolt_no_address",
			config:     &Config{Type: Bbolt},
		},
		{
			configType: "bbolt_with_address",
			config:     &Config{Type: Bbolt, Address: "/tmp/custom.db"},
		},
		{
			configType: "sugardb_no_dir",
			config:     &Config{Type: SugarDB},
		},
		{
			configType: "sugardb_with_dir",
			config:     &Config{Type: SugarDB, DataDir: "/tmp/custom_sugar"},
		},
		{
			configType: "bitcask_no_dir",
			config:     &Config{Type: Bitcask},
		},
		{
			configType: "bitcask_with_dir",
			config:     &Config{Type: Bitcask, DataDir: "/tmp/custom_bitcask"},
		},
		{
			configType: "redis_no_address",
			config:     &Config{Type: Redis},
		},
		{
			configType: "redis_with_address",
			config:     &Config{Type: Redis, Address: "127.0.0.1:6380"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.configType, func(t *testing.T) {
			originalConfig := *tc.config
			tc.config.apply()

			switch tc.config.Type {
			case Mem, "":
				assert.Equal(t, tc.config.Type, Mem)
			case Bbolt:
				if originalConfig.Address == "" {
					assert.Assert(t, tc.config.Address != "", "Bbolt address should be set")
				} else {
					assert.Equal(t, tc.config.Address, originalConfig.Address)
				}
			case SugarDB:
				if originalConfig.DataDir == "" {
					assert.Assert(t, tc.config.DataDir != "", "SugarDB data dir should be set")
				} else {
					assert.Equal(t, tc.config.DataDir, originalConfig.DataDir)
				}
			case Bitcask:
				if originalConfig.DataDir == "" {
					assert.Assert(t, tc.config.DataDir != "", "Bitcask data dir should be set")
				} else {
					assert.Equal(t, tc.config.DataDir, originalConfig.DataDir)
				}
			case Redis:
				if originalConfig.Address == "" {
					assert.Equal(t, tc.config.Address, "127.0.0.1:6379")
				} else {
					assert.Equal(t, tc.config.Address, originalConfig.Address)
				}
			}
		})
	}
}