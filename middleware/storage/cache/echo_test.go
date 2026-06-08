//go:build sugardb

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
	defer cache.Close()

	// Test basic operations
	err = cache.Set("key1", "value1")
	assert.NoError(t, err)

	value, err := cache.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Test SetEx
	err = cache.SetEx("key2", "value2", time.Second*10)
	assert.NoError(t, err)

	// Test TTL
	ttl, err := cache.Ttl("key2")
	assert.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))

	// Test TTL for non-existent key
	_, err = cache.Ttl("nonexistent")
	assert.Equal(t, ErrNotFound, err)

	// Test SetNx
	ok, err := cache.SetNx("nxkey1", "nxvalue1")
	assert.NoError(t, err)
	assert.True(t, ok)

	// Test Exists
	exists, err := cache.Exists("key1")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = cache.Exists("nonexistent")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test Expire
	ok, err = cache.Expire("key1", time.Second*10)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Test increment/decrement operations
	val, err := cache.Incr("counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = cache.Decr("counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), val)

	// Test hash operations
	ok, err = cache.HSet("hash1", "field1", "value1")
	assert.NoError(t, err)
	assert.True(t, ok)

	value, err = cache.HGet("hash1", "field1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Test Clean
	err = cache.Set("cleankey", "cleanvalue")
	assert.NoError(t, err)

	err = cache.Clean()
	assert.NoError(t, err)

	// Verify key is cleaned
	_, err = cache.Get("cleankey")
	assert.Equal(t, err, ErrNotFound)
}

func TestCacheSugarDB_ErrorHandling(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "sugardb_error_test")
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
	defer cache.Close()

	// Test operations on non-existent keys/fields
	_, err = cache.HGet("nonexistent", "field")
	assert.Error(t, err)

	members, err := cache.SMembers("nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(members))

	members, err = cache.SRandMember("nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(members))
}
