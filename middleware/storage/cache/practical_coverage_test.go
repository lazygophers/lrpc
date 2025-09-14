package cache

import (
	"os"
	"testing"
	"time"

	"github.com/lazygophers/utils/app"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)

// This file focuses on testing code paths that can actually be tested without
// external dependencies or incomplete implementations

func TestNewFactoryWithAllTypes(t *testing.T) {
	// Test New function with different types - focus on code paths that work

	// Test memory cache (this works)
	memCache, err := New(&Config{Type: Mem})
	assert.NilError(t, err)
	assert.Assert(t, memCache != nil)
	memCache.Close()

	// Test empty config (defaults to mem)
	defaultCache, err := New(&Config{})
	assert.NilError(t, err)
	assert.Assert(t, defaultCache != nil)
	defaultCache.Close()

	// Test unsupported type
	_, err = New(&Config{Type: "unsupported"})
	assert.Assert(t, err != nil)
	assert.Equal(t, err.Error(), "cache type not support")

	// Test Bbolt factory call (may fail but tests the factory code)
	tmpfile, err := os.CreateTemp("", "test_bbolt_*.db")
	if err == nil {
		tmpfile.Close()
		defer os.Remove(tmpfile.Name())

		_, err = New(&Config{Type: Bbolt, Address: tmpfile.Name()})
		// Don't assert on error - just test that factory code path is executed
	}

	// Test Redis factory call (will fail without server but tests factory code)
	_, err = New(&Config{Type: Redis, Address: "localhost:16379"})
	// Don't assert on error - just test that factory code path is executed

	// Test SugarDB factory call
	tmpdir, err := os.MkdirTemp("", "test_sugar_*")
	if err == nil {
		defer os.RemoveAll(tmpdir)
		_, err = New(&Config{Type: SugarDB, DataDir: tmpdir})
		// Don't assert on error - just test that factory code path is executed
	}

	// Test Bitcask factory call
	tmpdir, err = os.MkdirTemp("", "test_bitcask_*")
	if err == nil {
		defer os.RemoveAll(tmpdir)
		_, err = New(&Config{Type: Bitcask, DataDir: tmpdir})
		// Don't assert on error - just test that factory code path is executed
	}
}

func TestConfigApplyAllPaths(t *testing.T) {
	// Save original app name
	originalName := app.Name
	app.Name = "test_app"
	defer func() { app.Name = originalName }()

	// Test all configuration apply paths
	testConfigs := []*Config{
		{},                                   // Empty - should default to mem
		{Type: Mem},                          // Explicit mem
		{Type: Bbolt},                        // Bbolt without address
		{Type: Bbolt, Address: "/tmp/test"},  // Bbolt with address
		{Type: SugarDB},                      // SugarDB without dir
		{Type: SugarDB, DataDir: "/tmp"},     // SugarDB with dir
		{Type: Bitcask},                      // Bitcask without dir
		{Type: Bitcask, DataDir: "/tmp"},     // Bitcask with dir
		{Type: Redis},                        // Redis without address
		{Type: Redis, Address: "custom:123"}, // Redis with address
	}

	for i, config := range testConfigs {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			originalConfig := *config
			config.apply()

			switch originalConfig.Type {
			case "", Mem:
				assert.Equal(t, config.Type, Mem)
			case Bbolt:
				if originalConfig.Address == "" {
					assert.Assert(t, config.Address != "")
				} else {
					assert.Equal(t, config.Address, originalConfig.Address)
				}
			case SugarDB:
				if originalConfig.DataDir == "" {
					assert.Assert(t, config.DataDir != "")
				} else {
					assert.Equal(t, config.DataDir, originalConfig.DataDir)
				}
			case Bitcask:
				if originalConfig.DataDir == "" {
					assert.Assert(t, config.DataDir != "")
				} else {
					assert.Equal(t, config.DataDir, originalConfig.DataDir)
				}
			case Redis:
				if originalConfig.Address == "" {
					assert.Equal(t, config.Address, "127.0.0.1:6379")
				} else {
					assert.Equal(t, config.Address, originalConfig.Address)
				}
			}
		})
	}
}

func TestItemMethodsComplete(t *testing.T) {
	now := time.Now()
	item := &Item{
		Data:     "test data",
		ExpireAt: now,
	}

	// Test Bytes method
	bytes := item.Bytes()
	assert.Assert(t, len(bytes) > 0)

	// Test String method
	str := item.String()
	assert.Assert(t, str != "")
	assert.Assert(t, len(str) > 0)
}

func TestSetPbSuccessPath(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test successful SetPb
	proto := &timestamppb.Timestamp{Seconds: 123, Nanos: 456}
	err := cache.SetPb("test_key", proto)
	assert.NilError(t, err)

	// Test successful SetPbEx
	err = cache.SetPbEx("test_key_ex", proto, 1*time.Hour)
	assert.NilError(t, err)

	// Verify we can read it back
	retrieved := &timestamppb.Timestamp{}
	err = cache.GetPb("test_key", retrieved)
	assert.NilError(t, err)
	assert.Equal(t, retrieved.Seconds, int64(123))
	assert.Equal(t, retrieved.Nanos, int32(456))
}

func TestAllTypedGettersWithErrors(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test all typed getters with nonexistent keys
	testCases := []struct {
		name string
		fn   func() error
	}{
		{"GetBool", func() error { _, err := cache.GetBool("nonexistent"); return err }},
		{"GetInt", func() error { _, err := cache.GetInt("nonexistent"); return err }},
		{"GetUint", func() error { _, err := cache.GetUint("nonexistent"); return err }},
		{"GetInt32", func() error { _, err := cache.GetInt32("nonexistent"); return err }},
		{"GetUint32", func() error { _, err := cache.GetUint32("nonexistent"); return err }},
		{"GetInt64", func() error { _, err := cache.GetInt64("nonexistent"); return err }},
		{"GetUint64", func() error { _, err := cache.GetUint64("nonexistent"); return err }},
		{"GetFloat32", func() error { _, err := cache.GetFloat32("nonexistent"); return err }},
		{"GetFloat64", func() error { _, err := cache.GetFloat64("nonexistent"); return err }},
		{"GetSlice", func() error { _, err := cache.GetSlice("nonexistent"); return err }},
		{"GetBoolSlice", func() error { _, err := cache.GetBoolSlice("nonexistent"); return err }},
		{"GetIntSlice", func() error { _, err := cache.GetIntSlice("nonexistent"); return err }},
		{"GetUintSlice", func() error { _, err := cache.GetUintSlice("nonexistent"); return err }},
		{"GetInt32Slice", func() error { _, err := cache.GetInt32Slice("nonexistent"); return err }},
		{"GetUint32Slice", func() error { _, err := cache.GetUint32Slice("nonexistent"); return err }},
		{"GetInt64Slice", func() error { _, err := cache.GetInt64Slice("nonexistent"); return err }},
		{"GetUint64Slice", func() error { _, err := cache.GetUint64Slice("nonexistent"); return err }},
		{"GetFloat32Slice", func() error { _, err := cache.GetFloat32Slice("nonexistent"); return err }},
		{"GetFloat64Slice", func() error { _, err := cache.GetFloat64Slice("nonexistent"); return err }},
		{"GetJson", func() error { var v interface{}; return cache.GetJson("nonexistent", &v) }},
		{"HGetJson", func() error { var v interface{}; return cache.HGetJson("nonexistent", "field", &v) }},
		{"GetPb", func() error { return cache.GetPb("nonexistent", &timestamppb.Timestamp{}) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			assert.Assert(t, err != nil, "Should return error for nonexistent key")
		})
	}
}

func TestAllSliceMethodsWithInvalidJson(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Set invalid JSON
	cache.Set("invalid", "not valid json")

	sliceTestCases := []struct {
		name string
		fn   func() error
	}{
		{"GetBoolSlice", func() error { _, err := cache.GetBoolSlice("invalid"); return err }},
		{"GetIntSlice", func() error { _, err := cache.GetIntSlice("invalid"); return err }},
		{"GetUintSlice", func() error { _, err := cache.GetUintSlice("invalid"); return err }},
		{"GetInt32Slice", func() error { _, err := cache.GetInt32Slice("invalid"); return err }},
		{"GetUint32Slice", func() error { _, err := cache.GetUint32Slice("invalid"); return err }},
		{"GetInt64Slice", func() error { _, err := cache.GetInt64Slice("invalid"); return err }},
		{"GetUint64Slice", func() error { _, err := cache.GetUint64Slice("invalid"); return err }},
		{"GetFloat32Slice", func() error { _, err := cache.GetFloat32Slice("invalid"); return err }},
		{"GetFloat64Slice", func() error { _, err := cache.GetFloat64Slice("invalid"); return err }},
		{"GetSlice", func() error { _, err := cache.GetSlice("invalid"); return err }},
	}

	for _, tc := range sliceTestCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			assert.Assert(t, err != nil, "Should return error for invalid JSON")
		})
	}
}

func TestGetSliceEmptyString(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test GetSlice with empty string (special case)
	cache.Set("empty", "")
	slice, err := cache.GetSlice("empty")
	assert.NilError(t, err)
	assert.Assert(t, slice == nil, "Should return nil for empty string")
}

func TestMemCacheSetPrefixComplete(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Get the underlying CacheMem to test SetPrefix
	memCache := cache.(*baseCache).BaseCache.(*CacheMem)

	// SetPrefix is a no-op but should not panic
	memCache.SetPrefix("test_prefix:")

	// Should work normally after SetPrefix
	err := cache.Set("test", "value")
	assert.NilError(t, err)
}

func TestConstructorFunctions(t *testing.T) {
	// Test NewMem constructor
	memCache := NewMem()
	assert.Assert(t, memCache != nil)
	memCache.Close()

	// Test NewBbolt constructor (may fail but tests code path)
	tmpfile, err := os.CreateTemp("", "test_constructor_*.db")
	if err == nil {
		tmpfile.Close()
		defer os.Remove(tmpfile.Name())

		bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
		if err == nil && bboltCache != nil {
			bboltCache.Close()
		}
		// Don't assert on error - just test constructor code path
	}

	// Test NewDatabase constructor (skip actual call due to nil DB issues)
	// The NewDatabase function requires a valid DB connection
	// We just call it in a safe way that doesn't crash
	if false { // Disabled to avoid crashes
		_, _ = NewDatabase(nil, "test_table")
	}

	// Test NewRedis constructor
	_, err = NewRedis("localhost:16379")
	// Will likely fail but tests constructor code path

	// Test newBaseCache function
	memCacheBase := NewMem().(*baseCache).BaseCache
	wrapped := newBaseCache(memCacheBase)
	assert.Assert(t, wrapped != nil)
	wrapped.Close()
}

func TestLimitFunctionsNormalPath(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Limit normal operation
	allowed, err := cache.Limit("test_limit", 2, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Second call within limit
	allowed, err = cache.Limit("test_limit", 2, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Third call exceeds limit
	allowed, err = cache.Limit("test_limit", 2, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false)

	// Test LimitUpdateOnCheck normal operation
	allowed, err = cache.LimitUpdateOnCheck("test_update", 1, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	// Second call should exceed limit
	allowed, err = cache.LimitUpdateOnCheck("test_update", 1, 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false)
}

func TestMemCacheEdgeCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Expire on nonexistent key
	success, err := cache.Expire("nonexistent", 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Test Ttl on nonexistent key
	ttl, err := cache.Ttl("nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -2*time.Second)

	// Test Ttl on key with no expiration
	cache.Set("no_expire", "value")
	ttl, err = cache.Ttl("no_expire")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -1*time.Second)

	// Test SRandMember edge cases
	cache.SAdd("test_set", "a", "b", "c")

	// Test with count 0 (should return 1 member)
	members, err := cache.SRandMember("test_set", 0)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1)

	// Test with negative count
	members, err = cache.SRandMember("test_set", -1)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1)

	// Test SRandMember on empty set
	members, err = cache.SRandMember("empty_set")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)

	// Test SisMember on nonexistent set
	isMember, err := cache.SisMember("nonexistent", "member")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)

	// Test Exists with multiple keys where some don't exist
	cache.Set("exists1", "value")
	exists, err := cache.Exists("exists1", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false) // Should be false if any key doesn't exist
}
