package cache

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestCacheMem_SetNxWithTimeout(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test SetNxWithTimeout success
	success, err := cache.SetNxWithTimeout("timeout_key", "value", 100*time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, success, true)

	// Test SetNxWithTimeout failure (key exists)
	success, err = cache.SetNxWithTimeout("timeout_key", "new_value", 100*time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Wait for expiration and try again
	time.Sleep(150 * time.Millisecond)

	// Try SetNxWithTimeout again - it should succeed because the key has expired
	success, err = cache.SetNxWithTimeout("timeout_key", "new_value", 100*time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, success, true)
}

func TestCacheMem_Reset(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Add some data
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Verify data exists
	_, err := cache.Get("key1")
	assert.NilError(t, err)

	// Reset cache
	memCache := cache.(*baseCache).BaseCache.(*CacheMem)
	err = memCache.Reset()
	assert.NilError(t, err)

	// Verify data is gone
	_, err = cache.Get("key1")
	assert.Equal(t, err, ErrNotFound)
}

func TestCacheMem_SetPrefix(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// SetPrefix is a no-op for memory cache
	memCache := cache.(*baseCache).BaseCache.(*CacheMem)
	memCache.SetPrefix("test_prefix") // Should not panic or error
}

func TestCacheMem_EdgeCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test operations on expired keys
	err := cache.SetEx("expired_key", "value", 1*time.Millisecond)
	assert.NilError(t, err)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Operations on expired key should behave like key doesn't exist
	success, err := cache.Expire("expired_key", 1*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	exists, err := cache.Exists("expired_key")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Test IncrBy with non-numeric existing value
	cache.Set("non_numeric", "not_a_number")
	result, err := cache.IncrBy("non_numeric", 1)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1)) // Should reset to 0 then add 1

	// Test Hash operations with expired key
	cache.SetEx("expired_hash", "{\"field\":\"value\"}", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	// Should create new hash
	newVal, err := cache.HIncrBy("expired_hash", "counter", 5)
	assert.NilError(t, err)
	assert.Equal(t, newVal, int64(5))

	// Test Set operations with expired key
	cache.SetEx("expired_set", "[\"member1\"]", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	// Should create new set
	addedCount, err := cache.SAdd("expired_set", "member2")
	assert.NilError(t, err)
	assert.Equal(t, addedCount, int64(1))
}

func TestCacheMem_HashComplexOperations(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test HIncrBy with invalid JSON
	cache.Set("bad_hash", "not valid json")
	result, err := cache.HIncrBy("bad_hash", "field", 1)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1)) // Should create new hash

	// Test HSet with invalid JSON
	cache.Set("bad_hash2", "not valid json")
	isNew, err := cache.HSet("bad_hash2", "field", "value")
	assert.NilError(t, err)
	assert.Equal(t, isNew, true)

	// Test HGet with invalid JSON
	cache.Set("bad_hash3", "not valid json")
	_, err = cache.HGet("bad_hash3", "field")
	assert.Equal(t, err, ErrNotFound)

	// Test HExists with invalid JSON
	exists, err := cache.HExists("bad_hash3", "field")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Test HKeys with invalid JSON
	keys, err := cache.HKeys("bad_hash3")
	assert.NilError(t, err)
	assert.Equal(t, len(keys), 0)

	// Test HDel with invalid JSON
	deleted, err := cache.HDel("bad_hash3", "field")
	assert.NilError(t, err)
	assert.Equal(t, deleted, int64(0))

	// Test HGetAll with invalid JSON
	allVals, err := cache.HGetAll("bad_hash3")
	assert.NilError(t, err)
	assert.Equal(t, len(allVals), 0)
}

func TestCacheMem_SetComplexOperations(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test operations with invalid JSON
	cache.Set("bad_set", "not valid json")

	// SAdd with invalid JSON - should create new set
	added, err := cache.SAdd("bad_set", "member1")
	assert.NilError(t, err)
	assert.Equal(t, added, int64(1))

	// Reset to invalid JSON
	cache.Set("bad_set2", "not valid json")

	// SMembers with invalid JSON
	members, err := cache.SMembers("bad_set2")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)

	// SRem with invalid JSON
	removed, err := cache.SRem("bad_set2", "member1")
	assert.NilError(t, err)
	assert.Equal(t, removed, int64(0))

	// SRandMember with invalid JSON
	randMembers, err := cache.SRandMember("bad_set2", 1)
	assert.NilError(t, err)
	assert.Equal(t, len(randMembers), 0)

	// SPop with invalid JSON
	popped, err := cache.SPop("bad_set2")
	assert.NilError(t, err)
	assert.Equal(t, popped, "")

	// SisMember with invalid JSON
	isMember, err := cache.SisMember("bad_set2", "member1")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)
}

func TestCacheMem_SetOperationsEdgeCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test SRandMember with empty set
	randMembers, err := cache.SRandMember("empty_set", 5)
	assert.NilError(t, err)
	assert.Equal(t, len(randMembers), 0)

	// Test SRandMember with count larger than set size
	cache.SAdd("small_set", "member1", "member2")
	randMembers, err = cache.SRandMember("small_set", 10)
	assert.NilError(t, err)
	assert.Equal(t, len(randMembers), 2) // Should return all members

	// Test SRandMember with default count (no count parameter)
	randMembers, err = cache.SRandMember("small_set")
	assert.NilError(t, err)
	assert.Equal(t, len(randMembers), 1) // Should return 1 member

	// Test SPop with empty set
	popped, err := cache.SPop("empty_set_for_pop")
	assert.NilError(t, err)
	assert.Equal(t, popped, "")

	// Test operations on non-existent sets
	members, err := cache.SMembers("nonexistent_set")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)

	removed, err := cache.SRem("nonexistent_set", "member1")
	assert.NilError(t, err)
	assert.Equal(t, removed, int64(0))
}

func TestCacheMem_HashOperationsEdgeCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test operations on non-existent hash
	_, err := cache.HGet("nonexistent_hash", "field")
	assert.Equal(t, err, ErrNotFound)

	exists, err := cache.HExists("nonexistent_hash", "field")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	keys, err := cache.HKeys("nonexistent_hash")
	assert.NilError(t, err)
	assert.Equal(t, len(keys), 0)

	deleted, err := cache.HDel("nonexistent_hash", "field")
	assert.NilError(t, err)
	assert.Equal(t, deleted, int64(0))

	allVals, err := cache.HGetAll("nonexistent_hash")
	assert.NilError(t, err)
	assert.Equal(t, len(allVals), 0)

	// Test HIncrBy with invalid existing numeric value
	cache.HSet("hash_with_invalid_num", "field", "not_a_number")
	result, err := cache.HIncrBy("hash_with_invalid_num", "field", 5)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(5)) // Should reset to 0 then add 5
}
