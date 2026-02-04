package cache

import (
	"testing"
	"time"
)

// TestRedis_MockMode 测试 Redis Mock 模式
func TestRedis_MockMode(t *testing.T) {
	// 创建 Mock Redis 缓存
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create mock cache: %v", err)
	}
	defer cache.Close()

	// 测试基本操作
	t.Run("BasicOperations", func(t *testing.T) {
		// Set
		err := cache.Set("test_key", "test_value")
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		// Get
		val, err := cache.Get("test_key")
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if val != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", val)
		}

		// Del
		err = cache.Del("test_key")
		if err != nil {
			t.Errorf("Del failed: %v", err)
		}

		// Get after delete (should return ErrNotFound)
		_, err = cache.Get("test_key")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("SetEx", func(t *testing.T) {
		// SetEx
		err := cache.SetEx("expire_key", "expire_value", 10*time.Minute)
		if err != nil {
			t.Errorf("SetEx failed: %v", err)
		}

		// Get immediately
		val, err := cache.Get("expire_key")
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if val != "expire_value" {
			t.Errorf("Expected 'expire_value', got '%s'", val)
		}

		// Check TTL
		ttl, err := cache.Ttl("expire_key")
		if err != nil {
			t.Errorf("Ttl failed: %v", err)
		}
		if ttl <= 0 {
			t.Errorf("Expected positive TTL, got %v", ttl)
		}
	})

	t.Run("Incr", func(t *testing.T) {
		// Incr
		val, err := cache.Incr("counter")
		if err != nil {
			t.Errorf("Incr failed: %v", err)
		}
		if val != 1 {
			t.Errorf("Expected 1, got %d", val)
		}

		// Incr again
		val, err = cache.Incr("counter")
		if err != nil {
			t.Errorf("Incr failed: %v", err)
		}
		if val != 2 {
			t.Errorf("Expected 2, got %d", val)
		}
	})

	t.Run("Hash", func(t *testing.T) {
		// HSet
		ok, err := cache.HSet("hash_key", "field1", "value1")
		if err != nil {
			t.Errorf("HSet failed: %v", err)
		}
		if !ok {
			t.Error("Expected true for HSet")
		}

		// HGet
		val, err := cache.HGet("hash_key", "field1")
		if err != nil {
			t.Errorf("HGet failed: %v", err)
		}
		if val != "value1" {
			t.Errorf("Expected 'value1', got '%s'", val)
		}

		// HGetAll
		all, err := cache.HGetAll("hash_key")
		if err != nil {
			t.Errorf("HGetAll failed: %v", err)
		}
		if len(all) != 1 || all["field1"] != "value1" {
			t.Errorf("Unexpected HGetAll result: %v", all)
		}
	})

	t.Run("Set", func(t *testing.T) {
		// SAdd
		count, err := cache.SAdd("set_key", "member1", "member2")
		if err != nil {
			t.Errorf("SAdd failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected 2, got %d", count)
		}

		// SMembers
		members, err := cache.SMembers("set_key")
		if err != nil {
			t.Errorf("SMembers failed: %v", err)
		}
		if len(members) != 2 {
			t.Errorf("Expected 2 members, got %d", len(members))
		}
	})
}

// TestRedis_MockMode_Ping 测试 Mock 模式的 Ping
func TestRedis_MockMode_Ping(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create mock cache: %v", err)
	}
	defer cache.Close()

	err = cache.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

// TestRedis_MockMode_Exists 测试 Mock 模式的 Exists
func TestRedis_MockMode_Exists(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create mock cache: %v", err)
	}
	defer cache.Close()

	// Key doesn't exist
	exists, err := cache.Exists("nonexistent")
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected false for non-existent key")
	}

	// Create key
	err = cache.Set("exists_key", "value")
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// Key exists
	exists, err = cache.Exists("exists_key")
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected true for existing key")
	}
}
