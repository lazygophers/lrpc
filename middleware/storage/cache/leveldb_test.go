package cache

import (
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestLevelDBBasicOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test Set and Get
	err = cache.Set("test_key", "test_value")
	assert.NilError(t, err)

	value, err := cache.Get("test_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "test_value")

	// Test non-existent key
	_, err = cache.Get("non_existent")
	assert.Equal(t, err, ErrNotFound)

	// Test Exists
	exists, err := cache.Exists("test_key")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	exists, err = cache.Exists("non_existent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Test Del
	err = cache.Del("test_key")
	assert.NilError(t, err)

	_, err = cache.Get("test_key")
	assert.Equal(t, err, ErrNotFound)
}

func TestLevelDBSetEx(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test SetEx with expiration
	err = cache.SetEx("expire_key", "expire_value", 100*time.Millisecond)
	assert.NilError(t, err)

	// Should exist initially
	value, err := cache.Get("expire_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "expire_value")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = cache.Get("expire_key")
	assert.Equal(t, err, ErrNotFound)
}

func TestLevelDBTtl(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test TTL for non-existent key
	ttl, err := cache.Ttl("non_existent")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -2*time.Second)

	// Test TTL for key without expiration
	err = cache.Set("no_expire", "value")
	assert.NilError(t, err)

	ttl, err = cache.Ttl("no_expire")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -1*time.Second)

	// Test TTL for key with expiration
	err = cache.SetEx("with_expire", "value", 1*time.Hour)
	assert.NilError(t, err)

	ttl, err = cache.Ttl("with_expire")
	assert.NilError(t, err)
	assert.Assert(t, ttl > 59*time.Minute && ttl <= 1*time.Hour)
}

func TestLevelDBExpire(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test Expire on non-existent key
	success, err := cache.Expire("non_existent", 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Test Expire on existing key
	err = cache.Set("expire_test", "value")
	assert.NilError(t, err)

	success, err = cache.Expire("expire_test", 100*time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, success, true)

	// Should exist initially
	_, err = cache.Get("expire_test")
	assert.NilError(t, err)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = cache.Get("expire_test")
	assert.Equal(t, err, ErrNotFound)
}

func TestLevelDBSetNx(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test SetNx on non-existent key
	success, err := cache.SetNx("new_key", "new_value")
	assert.NilError(t, err)
	assert.Equal(t, success, true)

	value, err := cache.Get("new_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "new_value")

	// Test SetNx on existing key
	success, err = cache.SetNx("new_key", "another_value")
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Value should remain unchanged
	value, err = cache.Get("new_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "new_value")
}

func TestLevelDBSetNxWithTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test SetNxWithTimeout on non-existent key
	success, err := cache.SetNxWithTimeout("timeout_key", "timeout_value", 100*time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, success, true)

	// Should exist initially
	value, err := cache.Get("timeout_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "timeout_value")

	// Test SetNxWithTimeout on existing key
	success, err = cache.SetNxWithTimeout("timeout_key", "new_value", 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = cache.Get("timeout_key")
	assert.Equal(t, err, ErrNotFound)
}

func TestLevelDBIncrDecr(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test Incr on non-existent key
	result, err := cache.Incr("counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1))

	// Test Incr again
	result, err = cache.Incr("counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(2))

	// Test Decr
	result, err = cache.Decr("counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1))

	// Test IncrBy
	result, err = cache.IncrBy("counter", 10)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(11))

	// Test DecrBy
	result, err = cache.DecrBy("counter", 5)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(6))
}

func TestLevelDBHashOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test HSet
	isNew, err := cache.HSet("hash_key", "field1", "value1")
	assert.NilError(t, err)
	assert.Equal(t, isNew, true)

	isNew, err = cache.HSet("hash_key", "field1", "new_value1")
	assert.NilError(t, err)
	assert.Equal(t, isNew, false)

	// Test HGet
	value, err := cache.HGet("hash_key", "field1")
	assert.NilError(t, err)
	assert.Equal(t, value, "new_value1")

	// Test HGet non-existent field
	_, err = cache.HGet("hash_key", "non_existent")
	assert.Equal(t, err, ErrNotFound)

	// Test HExists
	exists, err := cache.HExists("hash_key", "field1")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	exists, err = cache.HExists("hash_key", "non_existent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Add more fields
	cache.HSet("hash_key", "field2", "value2")
	cache.HSet("hash_key", "field3", "value3")

	// Test HKeys
	keys, err := cache.HKeys("hash_key")
	assert.NilError(t, err)
	assert.Equal(t, len(keys), 3)
	assert.Assert(t, containsString(keys, "field1"))
	assert.Assert(t, containsString(keys, "field2"))
	assert.Assert(t, containsString(keys, "field3"))

	// Test HGetAll
	all, err := cache.HGetAll("hash_key")
	assert.NilError(t, err)
	assert.Equal(t, len(all), 3)
	assert.Equal(t, all["field1"], "new_value1")
	assert.Equal(t, all["field2"], "value2")
	assert.Equal(t, all["field3"], "value3")

	// Test HDel
	deleted, err := cache.HDel("hash_key", "field1", "field2", "non_existent")
	assert.NilError(t, err)
	assert.Equal(t, deleted, int64(2))

	// Verify deletion
	exists, err = cache.HExists("hash_key", "field1")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	exists, err = cache.HExists("hash_key", "field3")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)
}

func TestLevelDBHashIncr(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test HIncr on non-existent field
	result, err := cache.HIncr("hash_counter", "counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1))

	// Test HIncrBy
	result, err = cache.HIncrBy("hash_counter", "counter", 10)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(11))

	// Test HDecr
	result, err = cache.HDecr("hash_counter", "counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(10))

	// Test HDecrBy
	result, err = cache.HDecrBy("hash_counter", "counter", 5)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(5))
}

func TestLevelDBSetOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test SAdd
	added, err := cache.SAdd("set_key", "member1", "member2", "member3")
	assert.NilError(t, err)
	assert.Equal(t, added, int64(3))

	// Test adding duplicate members
	added, err = cache.SAdd("set_key", "member1", "member4")
	assert.NilError(t, err)
	assert.Equal(t, added, int64(1))

	// Test SMembers
	members, err := cache.SMembers("set_key")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 4)
	assert.Assert(t, containsString(members, "member1"))
	assert.Assert(t, containsString(members, "member2"))
	assert.Assert(t, containsString(members, "member3"))
	assert.Assert(t, containsString(members, "member4"))

	// Test SisMember
	isMember, err := cache.SisMember("set_key", "member1")
	assert.NilError(t, err)
	assert.Equal(t, isMember, true)

	isMember, err = cache.SisMember("set_key", "non_member")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)

	// Test SRem
	removed, err := cache.SRem("set_key", "member1", "member2", "non_member")
	assert.NilError(t, err)
	assert.Equal(t, removed, int64(2))

	// Test SRandMember
	randomMembers, err := cache.SRandMember("set_key", 1)
	assert.NilError(t, err)
	assert.Equal(t, len(randomMembers), 1)

	randomMembers, err = cache.SRandMember("set_key", 10)
	assert.NilError(t, err)
	assert.Equal(t, len(randomMembers), 2) // Only 2 members left

	// Test SPop
	member, err := cache.SPop("set_key")
	assert.NilError(t, err)
	assert.Assert(t, member == "member3" || member == "member4")

	// Verify member was removed
	members, err = cache.SMembers("set_key")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1)

	// Test SPop on empty set
	cache.SPop("set_key") // Remove last member
	member, err = cache.SPop("set_key")
	assert.NilError(t, err)
	assert.Equal(t, member, "")
}

func TestLevelDBClean(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Add some data
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.HSet("hash", "field", "value")
	cache.SAdd("set", "member1", "member2")

	// Verify data exists
	_, err = cache.Get("key1")
	assert.NilError(t, err)

	// Clean all data
	err = cache.Clean()
	assert.NilError(t, err)

	// Verify all data is gone
	_, err = cache.Get("key1")
	assert.Equal(t, err, ErrNotFound)

	_, err = cache.Get("key2")
	assert.Equal(t, err, ErrNotFound)

	_, err = cache.HGet("hash", "field")
	assert.Equal(t, err, ErrNotFound)

	members, err := cache.SMembers("set")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)
}

func TestLevelDBSetPrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)
	defer cache.Close()

	// Test SetPrefix
	cache.SetPrefix("test_prefix")

	// The prefix functionality is mainly for identifying cached data
	// but doesn't affect the actual key storage in leveldb
	err = cache.Set("key", "value")
	assert.NilError(t, err)

	value, err := cache.Get("key")
	assert.NilError(t, err)
	assert.Equal(t, value, "value")
}

func TestLevelDBErrorCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_leveldb_*")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)

	cache, err := NewLevelDB(&Config{
		DataDir: tmpDir,
	})
	assert.NilError(t, err)

	// Test operations on closed cache
	cache.Close()

	err = cache.Set("key", "value")
	assert.Assert(t, err != nil)

	_, err = cache.Get("key")
	assert.Assert(t, err != nil)
}
