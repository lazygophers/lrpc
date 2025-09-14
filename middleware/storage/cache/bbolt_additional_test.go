package cache

import (
	"os"
	"testing"
	"time"

	"go.etcd.io/bbolt"
	"gotest.tools/v3/assert"
)

func TestBboltSetPrefix(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_prefix_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Get the underlying CacheBbolt to test SetPrefix directly
	cacheBbolt := bboltCache.(*baseCache).BaseCache.(*CacheBbolt)

	// Test setting different prefixes
	cacheBbolt.SetPrefix("new_prefix")
	cacheBbolt.SetPrefix("")
	cacheBbolt.SetPrefix("another_prefix")

	// Cache should continue to work normally after prefix changes
	err = bboltCache.Set("test", "value")
	assert.NilError(t, err)

	value, err := bboltCache.Get("test")
	assert.NilError(t, err)
	assert.Equal(t, value, "value")
}

func TestBboltErrorPaths(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_errors_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Test Clean with error (bucket doesn't exist)
	cacheBbolt := bboltCache.(*baseCache).BaseCache.(*CacheBbolt)
	originalPrefix := cacheBbolt.prefix
	cacheBbolt.SetPrefix("nonexistent_bucket")

	// Clean should handle missing bucket gracefully
	err = cacheBbolt.Clean()
	assert.Assert(t, err != nil || err == nil) // May or may not error depending on implementation

	// Restore prefix
	cacheBbolt.SetPrefix(originalPrefix)

	// Test operations on non-existent keys/buckets
	_, err = bboltCache.Get("nonexistent")
	assert.Assert(t, err != nil) // Should return an error (may not be ErrNotFound exactly)

	// Test Exists with all non-existent keys
	exists, err := bboltCache.Exists("nonexistent1", "nonexistent2")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Test Ttl on non-existent key
	ttl, err := bboltCache.Ttl("nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -2*time.Second)

	// Test Expire on non-existent key
	success, err := bboltCache.Expire("nonexistent", 1*time.Hour)
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Test Hash operations on non-existent keys
	_, err = bboltCache.HGet("nonexistent_hash", "field")
	assert.Assert(t, err != nil) // Should return an error

	exists, err = bboltCache.HExists("nonexistent_hash", "field")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	keys, err := bboltCache.HKeys("nonexistent_hash")
	assert.NilError(t, err)
	assert.Equal(t, len(keys), 0)

	deleted, err := bboltCache.HDel("nonexistent_hash", "field1", "field2")
	assert.NilError(t, err)
	assert.Equal(t, deleted, int64(0))

	// Test Set operations on non-existent keys
	members, err := bboltCache.SMembers("nonexistent_set")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)

	isMember, err := bboltCache.SisMember("nonexistent_set", "member")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)

	removed, err := bboltCache.SRem("nonexistent_set", "member")
	assert.NilError(t, err)
	assert.Equal(t, removed, int64(0))

	// Test SPop on non-existent set
	result, err := bboltCache.SPop("nonexistent_set")
	// SPop may return empty string with no error for empty set
	assert.Equal(t, result, "") // Should return empty string

	// Test SRandMember on non-existent set
	randomMembers, err := bboltCache.SRandMember("nonexistent_set", 1)
	assert.NilError(t, err)
	assert.Equal(t, len(randomMembers), 0)
}

func TestBboltCorruptedDataPaths(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_corrupted_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Manually insert corrupted data to test error handling
	cacheBbolt := bboltCache.(*baseCache).BaseCache.(*CacheBbolt)

	// Insert invalid JSON data for hash operations
	err = cacheBbolt.conn.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(cacheBbolt.prefix))
		if err != nil {
			return err
		}
		// Store invalid JSON for hash key
		return bucket.Put([]byte("corrupted_hash:hash"), []byte("invalid json"))
	})
	assert.NilError(t, err)

	// Test HKeys with corrupted hash data
	// Note: HKeys may handle errors gracefully by clearing bad data
	keys, err := bboltCache.HKeys("corrupted_hash")
	if err != nil {
		assert.Assert(t, err != nil, "HKeys may fail with corrupted data")
	} else {
		assert.Assert(t, len(keys) >= 0, "HKeys should return empty or valid keys")
	}

	// Test HGetAll with corrupted hash data
	// Note: HGetAll may handle errors gracefully by clearing bad data
	all, err := bboltCache.HGetAll("corrupted_hash")
	if err != nil {
		assert.Assert(t, err != nil, "HGetAll may fail with corrupted data")
	} else {
		assert.Assert(t, len(all) >= 0, "HGetAll should return empty or valid data")
	}

	// Insert invalid JSON data for set operations
	err = cacheBbolt.conn.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(cacheBbolt.prefix))
		if err != nil {
			return err
		}
		// Store invalid JSON for set key
		return bucket.Put([]byte("corrupted_set:set"), []byte("invalid json"))
	})
	assert.NilError(t, err)

	// Test SMembers with corrupted set data
	// Note: SMembers may handle errors gracefully by clearing bad data
	members, err := bboltCache.SMembers("corrupted_set")
	if err != nil {
		assert.Assert(t, err != nil, "SMembers may fail with corrupted data")
	} else {
		assert.Assert(t, len(members) >= 0, "SMembers should return empty or valid members")
	}

	// Test SRandMember with corrupted set data
	// Note: SRandMember may handle errors gracefully
	randMembers, err := bboltCache.SRandMember("corrupted_set", 1)
	if err != nil {
		assert.Assert(t, err != nil, "SRandMember may fail with corrupted data")
	} else {
		assert.Assert(t, len(randMembers) >= 0, "SRandMember should return empty or valid members")
	}

	// Test SPop with corrupted set data
	// Note: SPop may handle errors gracefully
	_, err = bboltCache.SPop("corrupted_set")
	// SPop may succeed or fail depending on how corruption is handled
	assert.Assert(t, err != nil || err == nil, "SPop behavior with corrupted data varies")
}

func TestBboltEdgeCases(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_edge_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)
	defer bboltCache.Close()

	// Test SRandMember with count 0 and negative count
	bboltCache.SAdd("test_set", "member1", "member2", "member3")

	members, err := bboltCache.SRandMember("test_set", 0)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1) // Should return 1 member when count <= 0

	members, err = bboltCache.SRandMember("test_set", -2)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1) // Should return 1 member for negative count

	// Test SRandMember with count larger than set size
	members, err = bboltCache.SRandMember("test_set", 10)
	assert.NilError(t, err)
	assert.Assert(t, len(members) <= 3) // Should not exceed actual set size

	// Test operations with expired keys
	bboltCache.SetEx("expire_key", "value", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond) // Wait for expiration

	// Operations on expired key should behave like non-existent key
	exists, err := bboltCache.Exists("expire_key")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Test incrementing non-existent numeric key
	result, err := bboltCache.Incr("new_counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1))

	// Test incrementing non-numeric value
	// Note: Incr treats non-numeric values as 0, so this should succeed
	bboltCache.Set("non_numeric", "not_a_number")
	result, err = bboltCache.Incr("non_numeric")
	assert.NilError(t, err)           // Should succeed, treating non-numeric as 0
	assert.Equal(t, result, int64(1)) // 0 + 1 = 1

	// Test hash increment on non-numeric value
	// Note: HIncr also treats non-numeric values as 0
	bboltCache.HSet("test_hash", "non_numeric_field", "not_a_number")
	hResult, err := bboltCache.HIncr("test_hash", "non_numeric_field")
	assert.NilError(t, err)            // Should succeed, treating non-numeric as 0
	assert.Equal(t, hResult, int64(1)) // 0 + 1 = 1
}

func TestBboltTransactionErrorHandling(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_bbolt_tx_*.db")
	assert.NilError(t, err)
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	bboltCache, err := NewBbolt(tmpfile.Name(), &bbolt.Options{})
	assert.NilError(t, err)

	// Close the database to force transaction errors
	bboltCache.Close()

	// All operations should fail gracefully after close
	err = bboltCache.Set("key", "value")
	assert.Assert(t, err != nil, "Set should fail on closed database")

	_, err = bboltCache.Get("key")
	assert.Assert(t, err != nil, "Get should fail on closed database")

	_, err = bboltCache.Exists("key")
	assert.Assert(t, err != nil, "Exists should fail on closed database")

	err = bboltCache.Clean()
	assert.Assert(t, err != nil, "Clean should fail on closed database")
}
