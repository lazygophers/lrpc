package cache

import (
	"os"
	"testing"
	"time"

	"go.etcd.io/bbolt"
	"gotest.tools/v3/assert"
)

func TestBboltCacheImplementation(t *testing.T) {
	// Create a temporary file for bbolt
	tmpfile, err := os.CreateTemp("", "test_bbolt_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{
		Timeout:      time.Second * 5,
		ReadOnly:     false,
		FreelistType: bbolt.FreelistArrayType,
	})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Test basic operations
	err = bboltCache.Set("test_key", "test_value")
	assert.NilError(t, err)

	value, err := bboltCache.Get("test_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "test_value")

	// Test increment operations
	result, err := bboltCache.Incr("counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1))

	result, err = bboltCache.IncrBy("counter", 5)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(6))

	result, err = bboltCache.Decr("counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(5))

	result, err = bboltCache.DecrBy("counter", 2)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(3))

	// Test expiration
	err = bboltCache.SetEx("expire_key", "value", 100*time.Millisecond)
	assert.NilError(t, err)

	// Should exist initially
	exists, err := bboltCache.Exists("expire_key")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	// Should expire
	time.Sleep(150 * time.Millisecond)
	_, err = bboltCache.Get("expire_key")
	assert.Assert(t, err == NotFound)

	// Test TTL
	err = bboltCache.SetEx("ttl_key", "value", 1*time.Hour)
	assert.NilError(t, err)

	ttl, err := bboltCache.Ttl("ttl_key")
	assert.NilError(t, err)
	assert.Assert(t, ttl > 0, "TTL should be positive")
	assert.Assert(t, ttl <= time.Hour, "TTL should be less than or equal to 1 hour")

	// Test TTL for non-existent key
	ttl, err = bboltCache.Ttl("nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -2*time.Second)

	// Test TTL for key without expiration
	err = bboltCache.Set("no_expire", "value")
	assert.NilError(t, err)

	ttl, err = bboltCache.Ttl("no_expire")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -1*time.Second)

	// Test Expire function
	success, err := bboltCache.Expire("no_expire", 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, success, true)

	// Now it should have TTL
	ttl, err = bboltCache.Ttl("no_expire")
	assert.NilError(t, err)
	assert.Assert(t, ttl > 0, "TTL should be positive after setting expiration")
}

func TestBboltHashOperations(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_hash_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Test hash operations
	isNew, err := bboltCache.HSet("hash_key", "field1", "value1")
	assert.NilError(t, err)
	assert.Equal(t, isNew, true)

	isNew, err = bboltCache.HSet("hash_key", "field1", "value1_updated")
	assert.NilError(t, err)
	assert.Equal(t, isNew, false) // Should be false because field already existed

	value, err := bboltCache.HGet("hash_key", "field1")
	assert.NilError(t, err)
	assert.Equal(t, value, "value1_updated")

	// Test HExists
	exists, err := bboltCache.HExists("hash_key", "field1")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	exists, err = bboltCache.HExists("hash_key", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Test HIncr
	bboltCache.HSet("hash_key", "counter", "10")
	result, err := bboltCache.HIncr("hash_key", "counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(11))

	result, err = bboltCache.HIncrBy("hash_key", "counter", 5)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(16))

	result, err = bboltCache.HDecr("hash_key", "counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(15))

	result, err = bboltCache.HDecrBy("hash_key", "counter", 3)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(12))

	// Add more fields
	bboltCache.HSet("hash_key", "field2", "value2")
	bboltCache.HSet("hash_key", "field3", "value3")

	// Test HKeys
	keys, err := bboltCache.HKeys("hash_key")
	assert.NilError(t, err)
	assert.Assert(t, len(keys) >= 3) // At least field1, field2, field3, counter

	// Test HGetAll
	all, err := bboltCache.HGetAll("hash_key")
	assert.NilError(t, err)
	assert.Assert(t, len(all) >= 3)
	assert.Equal(t, all["field2"], "value2")
	assert.Equal(t, all["field3"], "value3")

	// Test HDel
	deleted, err := bboltCache.HDel("hash_key", "field2", "field3")
	assert.NilError(t, err)
	assert.Equal(t, deleted, int64(2))

	// Verify deletion
	_, err = bboltCache.HGet("hash_key", "field2")
	assert.Assert(t, err == NotFound)
}

func TestBboltSetOperations(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_set_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Test set operations
	added, err := bboltCache.SAdd("set_key", "member1", "member2", "member3")
	assert.NilError(t, err)
	assert.Equal(t, added, int64(3))

	// Adding existing members should return 0
	added, err = bboltCache.SAdd("set_key", "member1", "member2")
	assert.NilError(t, err)
	assert.Equal(t, added, int64(0))

	// Test SMembers
	members, err := bboltCache.SMembers("set_key")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 3)

	// Test SisMember
	isMember, err := bboltCache.SisMember("set_key", "member1")
	assert.NilError(t, err)
	assert.Equal(t, isMember, true)

	isMember, err = bboltCache.SisMember("set_key", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)

	// Test SRem
	removed, err := bboltCache.SRem("set_key", "member1")
	assert.NilError(t, err)
	assert.Equal(t, removed, int64(1))

	// Verify removal
	isMember, err = bboltCache.SisMember("set_key", "member1")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)

	// Test SRandMember
	randomMembers, err := bboltCache.SRandMember("set_key", 1)
	assert.NilError(t, err)
	assert.Equal(t, len(randomMembers), 1)

	// Test SPop
	poppedMember, err := bboltCache.SPop("set_key")
	assert.NilError(t, err)
	assert.Assert(t, poppedMember != "")

	// Verify the member was removed
	isMember, err = bboltCache.SisMember("set_key", poppedMember)
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)
}

func TestBboltSetNxOperations(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_setnx_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Test SetNx with non-existent key
	set, err := bboltCache.SetNx("nx_key", "value1")
	assert.NilError(t, err)
	assert.Equal(t, set, true)

	// Test SetNx with existing key
	set, err = bboltCache.SetNx("nx_key", "value2")
	assert.NilError(t, err)
	assert.Equal(t, set, false)

	// Value should remain unchanged
	value, err := bboltCache.Get("nx_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "value1")

	// Test SetNxWithTimeout
	set, err = bboltCache.SetNxWithTimeout("nx_timeout", "value", 100*time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, set, true)

	// Should exist initially
	value, err = bboltCache.Get("nx_timeout")
	assert.NilError(t, err)
	assert.Equal(t, value, "value")

	// Should expire
	time.Sleep(150 * time.Millisecond)
	_, err = bboltCache.Get("nx_timeout")
	assert.Assert(t, err == NotFound)
}

func TestBboltCleanOperation(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_clean_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Add some data
	bboltCache.Set("key1", "value1")
	bboltCache.Set("key2", "value2")
	bboltCache.HSet("hash", "field", "value")
	bboltCache.SAdd("set", "member")

	// Verify data exists
	value, err := bboltCache.Get("key1")
	assert.NilError(t, err)
	assert.Equal(t, value, "value1")

	// Clean all data
	err = bboltCache.Clean()
	assert.NilError(t, err)

	// Verify data is gone
	_, err = bboltCache.Get("key1")
	assert.Assert(t, err == NotFound)

	_, err = bboltCache.HGet("hash", "field")
	assert.Assert(t, err == NotFound)

	members, err := bboltCache.SMembers("set")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)
}
