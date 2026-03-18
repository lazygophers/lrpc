package cache

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestCacheMem_BasicOperations(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Set/Get
	err := cache.Set("key1", "value1")
	assert.NilError(t, err)

	value, err := cache.Get("key1")
	assert.NilError(t, err)
	assert.Equal(t, value, "value1")

	// Test non-existent key
	_, err = cache.Get("nonexistent")
	assert.Equal(t, err, ErrNotFound)
}

func TestCacheMem_SetEx(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test SetEx with short expiration
	err := cache.SetEx("expiring_key", "value", 100*time.Millisecond)
	assert.NilError(t, err)

	// Should exist initially
	value, err := cache.Get("expiring_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "value")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after expiration
	_, err = cache.Get("expiring_key")
	assert.Equal(t, err, ErrNotFound)
}

func TestCacheMem_SetNx(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// First SetNx should succeed
	success, err := cache.SetNx("nx_key", "value1")
	assert.NilError(t, err)
	assert.Equal(t, success, true)

	// Second SetNx should fail
	success, err = cache.SetNx("nx_key", "value2")
	assert.NilError(t, err)
	assert.Equal(t, success, false)

	// Verify original value is preserved
	value, err := cache.Get("nx_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "value1")
}

func TestCacheMem_IncrDecr(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test Incr on non-existent key (should start from 0)
	result, err := cache.Incr("counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1))

	// Test IncrBy
	result, err = cache.IncrBy("counter", 5)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(6))

	// Test Decr
	result, err = cache.Decr("counter")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(5))

	// Test DecrBy
	result, err = cache.DecrBy("counter", 3)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(2))
}

func TestCacheMem_Exists(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test non-existent key
	exists, err := cache.Exists("nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Set a key
	err = cache.Set("exists_key", "value")
	assert.NilError(t, err)

	// Test existing key
	exists, err = cache.Exists("exists_key")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	// Test multiple keys
	exists, err = cache.Exists("exists_key", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false) // Should be false if any key doesn't exist
}

func TestCacheMem_Expire(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Set a key
	err := cache.Set("expire_test", "value")
	assert.NilError(t, err)

	// Set expiration
	success, err := cache.Expire("expire_test", 100*time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, success, true)

	// Key should still exist
	value, err := cache.Get("expire_test")
	assert.NilError(t, err)
	assert.Equal(t, value, "value")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Key should be expired
	_, err = cache.Get("expire_test")
	assert.Equal(t, err, ErrNotFound)
}

func TestCacheMem_Ttl(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test non-existent key
	ttl, err := cache.Ttl("nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -2*time.Second)

	// Set a key without expiration
	err = cache.Set("no_expire", "value")
	assert.NilError(t, err)

	ttl, err = cache.Ttl("no_expire")
	assert.NilError(t, err)
	assert.Equal(t, ttl, -1*time.Second)

	// Set a key with expiration
	err = cache.SetEx("with_expire", "value", 1*time.Second)
	assert.NilError(t, err)

	ttl, err = cache.Ttl("with_expire")
	assert.NilError(t, err)
	assert.Assert(t, ttl > 0 && ttl <= 1*time.Second)
}

func TestCacheMem_HashOperations(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test HSet
	isNew, err := cache.HSet("hash_key", "field1", "value1")
	assert.NilError(t, err)
	assert.Equal(t, isNew, true)

	// Test HSet existing field
	isNew, err = cache.HSet("hash_key", "field1", "new_value1")
	assert.NilError(t, err)
	assert.Equal(t, isNew, false)

	// Test HGet
	value, err := cache.HGet("hash_key", "field1")
	assert.NilError(t, err)
	assert.Equal(t, value, "new_value1")

	// Test HGet non-existent field
	_, err = cache.HGet("hash_key", "nonexistent")
	assert.Equal(t, err, ErrNotFound)

	// Test HExists
	exists, err := cache.HExists("hash_key", "field1")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)

	exists, err = cache.HExists("hash_key", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	// Add more fields
	cache.HSet("hash_key", "field2", "value2")
	cache.HSet("hash_key", "field3", "value3")

	// Test HKeys
	keys, err := cache.HKeys("hash_key")
	assert.NilError(t, err)
	assert.Equal(t, len(keys), 3)

	// Test HGetAll
	allValues, err := cache.HGetAll("hash_key")
	assert.NilError(t, err)
	assert.Equal(t, len(allValues), 3)
	assert.Equal(t, allValues["field1"], "new_value1")
	assert.Equal(t, allValues["field2"], "value2")
	assert.Equal(t, allValues["field3"], "value3")

	// Test HDel
	deletedCount, err := cache.HDel("hash_key", "field2", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, deletedCount, int64(1))

	// Verify deletion
	exists, err = cache.HExists("hash_key", "field2")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)
}

func TestCacheMem_HashIncr(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test HIncr on non-existent hash/field
	result, err := cache.HIncr("hash_counter", "counter_field")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(1))

	// Test HIncrBy
	result, err = cache.HIncrBy("hash_counter", "counter_field", 5)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(6))

	// Test HDecr
	result, err = cache.HDecr("hash_counter", "counter_field")
	assert.NilError(t, err)
	assert.Equal(t, result, int64(5))

	// Test HDecrBy
	result, err = cache.HDecrBy("hash_counter", "counter_field", 2)
	assert.NilError(t, err)
	assert.Equal(t, result, int64(3))
}

func TestCacheMem_SetOperations(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test SAdd
	addedCount, err := cache.SAdd("set_key", "member1", "member2", "member3")
	assert.NilError(t, err)
	assert.Equal(t, addedCount, int64(3))

	// Test SAdd existing members
	addedCount, err = cache.SAdd("set_key", "member2", "member4")
	assert.NilError(t, err)
	assert.Equal(t, addedCount, int64(1)) // Only member4 is new

	// Test SMembers
	members, err := cache.SMembers("set_key")
	assert.NilError(t, err)
	assert.Equal(t, len(members), 4)

	// Test SisMember
	isMember, err := cache.SisMember("set_key", "member1")
	assert.NilError(t, err)
	assert.Equal(t, isMember, true)

	isMember, err = cache.SisMember("set_key", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)

	// Test SRem
	removedCount, err := cache.SRem("set_key", "member2", "nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, removedCount, int64(1)) // Only member2 existed

	// Verify removal
	isMember, err = cache.SisMember("set_key", "member2")
	assert.NilError(t, err)
	assert.Equal(t, isMember, false)

	// Test SRandMember
	randomMembers, err := cache.SRandMember("set_key", 2)
	assert.NilError(t, err)
	assert.Equal(t, len(randomMembers), 2)

	// Test SPop
	currentMembers, _ := cache.SMembers("set_key")
	popped, err := cache.SPop("set_key")
	assert.NilError(t, err)
	assert.Assert(t, popped != "")

	// Verify set size decreased
	newMembers, err := cache.SMembers("set_key")
	assert.NilError(t, err)
	assert.Equal(t, len(newMembers), len(currentMembers)-1)
}

func TestCacheMem_Del(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Set multiple keys
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Delete multiple keys
	err := cache.Del("key1", "key2", "nonexistent")
	assert.NilError(t, err)

	// Verify deletion
	_, err = cache.Get("key1")
	assert.Equal(t, err, ErrNotFound)

	_, err = cache.Get("key2")
	assert.Equal(t, err, ErrNotFound)

	// key3 should still exist
	value, err := cache.Get("key3")
	assert.NilError(t, err)
	assert.Equal(t, value, "value3")
}

func TestCacheMem_Clean(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Set multiple keys
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Clean all data
	err := cache.Clean()
	assert.NilError(t, err)

	// All keys should be gone
	_, err = cache.Get("key1")
	assert.Equal(t, err, ErrNotFound)

	_, err = cache.Get("key2")
	assert.Equal(t, err, ErrNotFound)
}

func TestCacheMem_ConcurrentAccess(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	// Goroutine 1: Write data
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("concurrent_key", "value")
			cache.Incr("concurrent_counter")
		}
		done <- true
	}()

	// Goroutine 2: Read data
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get("concurrent_key")
			cache.Get("concurrent_counter")
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	value, err := cache.Get("concurrent_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "value")

	counter, err := cache.Get("concurrent_counter")
	assert.NilError(t, err)
	assert.Equal(t, counter, "100")
}

func TestCacheMem_AutoClear(t *testing.T) {
	// Create cache
	memCache := &CacheMem{
		data:     make(map[string]*Item),
		streams:  make(map[string]*Stream),
		streamID: 0,
	}
	cache := newBaseCache(memCache)
	defer cache.Close()

	// Set a key with expiration
	err := cache.SetEx("expired_key", "value", 50*time.Millisecond)
	assert.NilError(t, err)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Trigger autoClear by performing an operation
	cache.Set("new_key", "new_value")

	// The expired key should be cleaned up
	_, err = cache.Get("expired_key")
	assert.Equal(t, err, ErrNotFound)

	// The new key should still exist
	value, err := cache.Get("new_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "new_value")
}

// TestCacheMem_ZSet_Comprehensive ZSet综合测试
func TestCacheMem_ZSet_Comprehensive(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	key := "leaderboard"

	// 测试ZAdd
	count, err := cache.ZAdd(key, 100.0, "user1", 95.0, "user2", 90.0, "user3")
	assert.NilError(t, err)
	assert.Equal(t, count, int64(3))

	// 测试ZCard
	card, err := cache.ZCard(key)
	assert.NilError(t, err)
	assert.Equal(t, card, int64(3))

	// 测试ZScore
	score, err := cache.ZScore(key, "user1")
	assert.NilError(t, err)
	assert.Equal(t, score, 100.0)

	// 测试ZRange（升序）
	members, err := cache.ZRange(key, 0, -1)
	assert.NilError(t, err)
	assert.DeepEqual(t, members, []string{"user3", "user2", "user1"})

	// 测试ZRevRange（降序）
	members, err = cache.ZRevRange(key, 0, -1)
	assert.NilError(t, err)
	assert.DeepEqual(t, members, []string{"user1", "user2", "user3"})

	// 测试ZRank
	rank, err := cache.ZRank(key, "user3")
	assert.NilError(t, err)
	assert.Equal(t, rank, int64(0)) // 最低分，排名第0

	// 测试ZRevRank
	revRank, err := cache.ZRevRank(key, "user1")
	assert.NilError(t, err)
	assert.Equal(t, revRank, int64(0)) // 最高分，降序排名第0

	// 测试ZIncrBy
	newScore, err := cache.ZIncrBy(key, 5.0, "user3")
	assert.NilError(t, err)
	assert.Equal(t, newScore, 95.0)

	// 测试ZRangeWithScores
	withScores, err := cache.ZRangeWithScores(key, 0, -1)
	assert.NilError(t, err)
	assert.Equal(t, len(withScores), 3)

	// 测试ZCount
	count, err = cache.ZCount(key, "90", "100")
	assert.NilError(t, err)
	assert.Equal(t, count, int64(3))

	// 测试ZRangeByScore
	members, err = cache.ZRangeByScore(key, "95", "100", 0, 0)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 3)

	// 测试ZRem
	removed, err := cache.ZRem(key, "user2", "user3")
	assert.NilError(t, err)
	assert.Equal(t, removed, int64(2))

	// 验证删除后的结果
	card, err = cache.ZCard(key)
	assert.NilError(t, err)
	assert.Equal(t, card, int64(1))

	// 测试ZRemRangeByRank
	cache.ZAdd(key, 80.0, "user4", 85.0, "user5")
	removed, err = cache.ZRemRangeByRank(key, 0, 1)
	assert.NilError(t, err)
	assert.Equal(t, removed, int64(2))

	// 测试ZRemRangeByScore
	cache.ZAdd(key, 70.0, "user6", 75.0, "user7")
	removed, err = cache.ZRemRangeByScore(key, "70", "75")
	assert.NilError(t, err)
	assert.Equal(t, removed, int64(2))

	// 测试ZUnionStore和ZInterStore
	cache.ZAdd("set1", 1.0, "a", 2.0, "b")
	cache.ZAdd("set2", 2.0, "b", 3.0, "c")

	unionCount, err := cache.ZUnionStore("union", "set1", "set2")
	assert.NilError(t, err)
	assert.Equal(t, unionCount, int64(3))

	interCount, err := cache.ZInterStore("inter", "set1", "set2")
	assert.NilError(t, err)
	assert.Equal(t, interCount, int64(1)) // 只有b在两个集合中

	// 测试错误情况
	_, err = cache.ZScore("nonexistent", "user1")
	assert.Equal(t, err, ErrNotFound)

	_, err = cache.ZRank("nonexistent", "user1")
	assert.Equal(t, err, ErrNotFound)
}

// TestCacheMem_ZSet_Concurrent 并发安全性测试
func TestCacheMem_ZSet_Concurrent(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	key := "concurrent_zset"
	done := make(chan bool, 3)

	// Goroutine 1: 并发ZAdd
	go func() {
		for i := 0; i < 100; i++ {
			cache.ZAdd(key, float64(i), fmt.Sprintf("member%d", i))
		}
		done <- true
	}()

	// Goroutine 2: 并发ZRange
	go func() {
		for i := 0; i < 100; i++ {
			cache.ZRange(key, 0, -1)
		}
		done <- true
	}()

	// Goroutine 3: 并发ZIncrBy
	go func() {
		for i := 0; i < 100; i++ {
			cache.ZIncrBy(key, 1.0, "member50")
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// Verify final state
	card, err := cache.ZCard(key)
	assert.NilError(t, err)
	assert.Assert(t, card > 0)
}

// TestCacheMem_ZSet_EdgeCases ZSet边界测试
func TestCacheMem_ZSet_EdgeCases(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// 空集合测试
	members, err := cache.ZRange("empty", 0, -1)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 0)

	card, err := cache.ZCard("empty")
	assert.NilError(t, err)
	assert.Equal(t, card, int64(0))

	// 单元素测试
	cache.ZAdd("single", 100.0, "only")
	members, err = cache.ZRange("single", 0, -1)
	assert.NilError(t, err)
	assert.DeepEqual(t, members, []string{"only"})

	// 负索引测试
	cache.ZAdd("neg", 1.0, "a", 2.0, "b", 3.0, "c")
	members, err = cache.ZRange("neg", -2, -1)
	assert.NilError(t, err)
	assert.DeepEqual(t, members, []string{"b", "c"})

	// 无穷大测试
	cache.ZAdd("inf", 50.0, "mid")
	members, err = cache.ZRangeByScore("inf", "-inf", "+inf", 0, 0)
	assert.NilError(t, err)
	assert.Equal(t, len(members), 1)
}
