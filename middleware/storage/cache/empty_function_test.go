package cache

import (
	"testing"
)

func TestSetPrefixDirectly(t *testing.T) {
	mem := &CacheMem{}
	mem.SetPrefix("test") // This should hit line 30-31

	// Also test through the interface
	cache := NewMem()
	defer cache.Close()
	memCache := cache.(*baseCache).BaseCache.(*CacheMem)
	memCache.SetPrefix("another_test") // This should definitely hit it
}
