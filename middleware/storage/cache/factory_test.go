package cache

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCacheFactory(t *testing.T) {
	// Test New() function with different cache types

	// Test memory cache creation
	memCache, err := New(&Config{
		Type: Mem,
	})
	assert.NilError(t, err)
	assert.Assert(t, memCache != nil)
	memCache.Close()

	// Test with invalid type (should return error)
	_, err = New(&Config{
		Type: "unknown",
	})
	assert.Assert(t, err != nil, "Should return error for unknown cache type")

	// Test with empty config
	emptyCache, err := New(&Config{})
	assert.NilError(t, err)
	assert.Assert(t, emptyCache != nil)
	emptyCache.Close()
}

func TestNotFoundError(t *testing.T) {
	// Test NotFound error type
	assert.Equal(t, NotFound.Error(), "key not found")

	// Test that it's comparable
	err := NotFound
	assert.Assert(t, err == NotFound)
}
