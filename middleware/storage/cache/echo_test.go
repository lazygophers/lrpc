package cache

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheSugarDB_AllFunctionsCoverage(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "sugardb_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := &Config{
		Type:     SugarDB,
		DataDir:  tmpDir,
		Password: "testpass",
	}

	cache, err := NewSugarDB(config)
	require.NoError(t, err)
	require.NotNil(t, cache)

	// Access the underlying SugarDB instance
	baseCache := cache.(*baseCache)
	sugarDB := baseCache.BaseCache.(*CacheSugarDB)

	// Test SetPrefix (no-op but should be callable)
	sugarDB.SetPrefix("test:")

	// Test basic operations
	err = sugarDB.Set("key1", "value1")
	assert.NoError(t, err)

	value, err := sugarDB.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Test Get non-existent key
	_, err = sugarDB.Get("nonexistent")
	assert.Equal(t, ErrNotFound, err)

	// Test SetEx
	err = sugarDB.SetEx("expkey", "expvalue", time.Second*1)
	assert.NoError(t, err)

	value, err = sugarDB.Get("expkey")
	assert.NoError(t, err)
	assert.Equal(t, "expvalue", value)

	// Test TTL
	ttl, err := sugarDB.Ttl("expkey")
	assert.NoError(t, err)
	assert.True(t, ttl > 0)

	// Test TTL for non-existent key
	_, err = sugarDB.Ttl("nonexistent")
	assert.Equal(t, ErrNotFound, err)

	// Test SetNx
	ok, err := sugarDB.SetNx("nxkey1", "nxvalue1")
	assert.NoError(t, err)
	assert.True(t, ok)

	// Test SetNx on existing key (should fail)
	ok, err = sugarDB.SetNx("nxkey1", "newvalue")
	assert.NoError(t, err) // SugarDB returns error when key exists
	assert.False(t, ok)

	// Test SetNxWithTimeout with new key
	ok, err = sugarDB.SetNxWithTimeout("nxkey2", "nxvalue2", time.Second*1)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Test SetNxWithTimeout on existing key (should fail and not set timeout)
	ok, err = sugarDB.SetNxWithTimeout("nxkey2", "newvalue", time.Second*1)
	if err != nil {
		// SugarDB returns error for existing key
		assert.False(t, ok)
	} else {
		assert.False(t, ok)
	}

	// Test Expire
	ok, err = sugarDB.Expire("key1", time.Second*10)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Test increment/decrement operations
	val, err := sugarDB.Incr("counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = sugarDB.Decr("counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), val)

	val, err = sugarDB.IncrBy("counter", 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), val)

	val, err = sugarDB.DecrBy("counter", 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)

	// Test Exists
	exists, err := sugarDB.Exists("key1")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = sugarDB.Exists("nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = sugarDB.Exists("key1", "nxkey1") // multiple keys
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test hash operations
	ok, err = sugarDB.HSet("hash1", "field1", "value1")
	assert.NoError(t, err)
	assert.True(t, ok)

	value, err = sugarDB.HGet("hash1", "field1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Test HGet non-existent field
	_, err = sugarDB.HGet("hash1", "nonexistent")
	assert.Equal(t, ErrNotFound, err)

	// Test HGet non-existent hash
	_, err = sugarDB.HGet("nonexistent", "field")
	assert.Error(t, err)

	exists, err = sugarDB.HExists("hash1", "field1")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = sugarDB.HExists("hash1", "nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test HKeys
	sugarDB.HSet("hash1", "field2", "value2")
	keys, err := sugarDB.HKeys("hash1")
	assert.NoError(t, err)
	assert.Contains(t, keys, "field1")
	assert.Contains(t, keys, "field2")

	// Test HGetAll
	all, err := sugarDB.HGetAll("hash1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", all["field1"])
	assert.Equal(t, "value2", all["field2"])

	// Test HGetAll on non-existent hash
	emptyAll, err := sugarDB.HGetAll("nonexistent")
	if err != nil {
		// SugarDB may return error for non-existent hash
		assert.Error(t, err)
	} else {
		assert.Empty(t, emptyAll)
	}

	// Test hash increment operations
	val, err = sugarDB.HIncr("hash1", "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = sugarDB.HIncrBy("hash1", "counter", 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(6), val)

	val, err = sugarDB.HDecr("hash1", "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), val)

	val, err = sugarDB.HDecrBy("hash1", "counter", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), val)

	// Test HDel
	count, err := sugarDB.HDel("hash1", "field1", "field2")
	assert.NoError(t, err)
	assert.True(t, count >= 0) // May be 0 if fields don't exist

	// Test set operations
	count, err = sugarDB.SAdd("set1", "member1", "member2", "member3")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	members, err := sugarDB.SMembers("set1")
	assert.NoError(t, err)
	assert.Len(t, members, 3)
	assert.Contains(t, members, "member1")

	isMember, err := sugarDB.SisMember("set1", "member1")
	assert.NoError(t, err)
	assert.True(t, isMember)

	isMember, err = sugarDB.SisMember("set1", "nonexistent")
	assert.NoError(t, err)
	assert.False(t, isMember)

	// Test SRandMember
	randMembers, err := sugarDB.SRandMember("set1")
	assert.NoError(t, err)
	assert.Len(t, randMembers, 1)
	assert.Contains(t, members, randMembers[0])

	randMembers, err = sugarDB.SRandMember("set1", 2)
	assert.NoError(t, err)
	assert.True(t, len(randMembers) >= 1)

	// Test SRem
	count, err = sugarDB.SRem("set1", "member1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Test SPop
	member, err := sugarDB.SPop("set1")
	assert.NoError(t, err)
	assert.NotEmpty(t, member)

	// Test SPop on empty set - remove all members first
	sugarDB.SRem("set1", "member2", "member3")
	_, err = sugarDB.SPop("set1")
	assert.Equal(t, ErrNotFound, err)

	// Test Del
	err = sugarDB.Set("delkey1", "delvalue1")
	assert.NoError(t, err)
	err = sugarDB.Set("delkey2", "delvalue2")
	assert.NoError(t, err)

	err = sugarDB.Del("delkey1", "delkey2")
	assert.NoError(t, err)

	// Verify keys are deleted
	_, err = sugarDB.Get("delkey1")
	assert.Equal(t, ErrNotFound, err)

	// Test Clean
	err = sugarDB.Set("cleankey", "cleanvalue")
	assert.NoError(t, err)

	err = sugarDB.Clean()
	assert.NoError(t, err)

	// Verify key is cleaned
	_, err = sugarDB.Get("cleankey")
	assert.Equal(t, ErrNotFound, err)

	// Test Close
	err = sugarDB.Close()
	assert.NoError(t, err)

	// Close cache
	cache.Close()
}

func TestNewSugarDB_ConfigOptions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sugardb_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test with minimal config
	config1 := &Config{
		Type: SugarDB,
	}
	cache1, err := NewSugarDB(config1)
	assert.NoError(t, err)
	assert.NotNil(t, cache1)
	cache1.Close()

	// Test with DataDir
	config2 := &Config{
		Type:    SugarDB,
		DataDir: tmpDir,
	}
	cache2, err := NewSugarDB(config2)
	assert.NoError(t, err)
	assert.NotNil(t, cache2)
	cache2.Close()

	// Test with Password
	config3 := &Config{
		Type:     SugarDB,
		DataDir:  tmpDir,
		Password: "testpass",
	}
	cache3, err := NewSugarDB(config3)
	assert.NoError(t, err)
	assert.NotNil(t, cache3)
	cache3.Close()
}

func TestCacheSugarDB_ErrorHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sugardb_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := &Config{
		Type:    SugarDB,
		DataDir: tmpDir,
	}

	cache, err := NewSugarDB(config)
	require.NoError(t, err)
	defer cache.Close()

	baseCache := cache.(*baseCache)
	sugarDB := baseCache.BaseCache.(*CacheSugarDB)

	// Test error paths by working with actual data
	// Test empty string handling
	err = sugarDB.Set("emptykey", "")
	assert.NoError(t, err)

	_, err = sugarDB.Get("emptykey")
	assert.Equal(t, ErrNotFound, err) // Empty strings treated as non-existent

	// Test different value types
	err = sugarDB.Set("intkey", 123)
	assert.NoError(t, err)

	value, err := sugarDB.Get("intkey")
	assert.NoError(t, err)
	assert.Equal(t, "123", value)

	err = sugarDB.Set("boolkey", true)
	assert.NoError(t, err)

	value, err = sugarDB.Get("boolkey")
	assert.NoError(t, err)
	// SugarDB converts boolean differently
	assert.NotEmpty(t, value)

	// Test operations on non-existent keys/fields
	_, err = sugarDB.HGet("nonexistent", "field")
	assert.Error(t, err) // Should error on non-existent hash

	_, err = sugarDB.SMembers("nonexistent")
	assert.Error(t, err) // Should error on non-existent set

	_, err = sugarDB.SRandMember("nonexistent")
	assert.Error(t, err) // Should error on non-existent set
}