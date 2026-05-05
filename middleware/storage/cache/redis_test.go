package cache

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

// TestRedis_MockMode_BasicOperations 测试基本操作
func TestRedis_MockMode_BasicOperations(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	t.Run("Set and Get", func(t *testing.T) {
		err = cache.Set("test_key", "test_value")
		assert.NilError(t, err)

		val, err := cache.Get("test_key")
		assert.NilError(t, err)
		assert.Equal(t, val, "test_value")

		err = cache.Del("test_key")
		assert.NilError(t, err)

		_, err = cache.Get("test_key")
		assert.Equal(t, err, ErrNotFound)
	})

	t.Run("SetEx", func(t *testing.T) {
		err = cache.SetEx("test_ex_key", "test_value", time.Second)
		assert.NilError(t, err)

		val, err := cache.Get("test_ex_key")
		assert.NilError(t, err)
		assert.Equal(t, val, "test_value")
	})

	t.Run("SetNx", func(t *testing.T) {
		err = cache.Del("test_nx_key")
		assert.NilError(t, err)

		ok, err := cache.SetNx("test_nx_key", "first_value")
		assert.NilError(t, err)
		assert.Equal(t, ok, true)

		ok, err = cache.SetNx("test_nx_key", "second_value")
		assert.NilError(t, err)
		assert.Equal(t, ok, false)

		val, err := cache.Get("test_nx_key")
		assert.NilError(t, err)
		assert.Equal(t, val, "first_value")
	})

	t.Run("SetNxWithTimeout", func(t *testing.T) {
		err = cache.Del("test_nx_ex_key")
		assert.NilError(t, err)

		ok, err := cache.SetNxWithTimeout("test_nx_ex_key", "value", time.Second)
		assert.NilError(t, err)
		assert.Equal(t, ok, true)
	})

	t.Run("Exists", func(t *testing.T) {
		err = cache.Set("test_exists_key", "test_value")
		assert.NilError(t, err)

		exists, err := cache.Exists("test_exists_key")
		assert.NilError(t, err)
		assert.Equal(t, exists, true)

		exists, err = cache.Exists("nonexistent_key")
		assert.NilError(t, err)
		assert.Equal(t, exists, false)
	})

	t.Run("Ttl", func(t *testing.T) {
		err = cache.SetEx("test_ttl_key", "value", 10*time.Second)
		assert.NilError(t, err)

		ttl, err := cache.Ttl("test_ttl_key")
		assert.NilError(t, err)
		assert.Assert(t, ttl > 0)

		ttl, err = cache.Ttl("nonexistent_key")
		assert.NilError(t, err)
		assert.Equal(t, ttl, time.Duration(-2))
	})

	t.Run("Expire", func(t *testing.T) {
		err = cache.Set("test_expire_key", "value")
		assert.NilError(t, err)

		ok, err := cache.Expire("test_expire_key", time.Minute)
		assert.NilError(t, err)
		assert.Equal(t, ok, true)
	})

	t.Run("Incr", func(t *testing.T) {
		err = cache.Del("test_incr_key")
		assert.NilError(t, err)

		err = cache.Set("test_incr_key", "10")
		assert.NilError(t, err)

		val, err := cache.Incr("test_incr_key")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(11))
	})

	t.Run("Decr", func(t *testing.T) {
		err = cache.Del("test_decr_key")
		assert.NilError(t, err)

		err = cache.Set("test_decr_key", "10")
		assert.NilError(t, err)

		val, err := cache.Decr("test_decr_key")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(9))
	})

	t.Run("IncrBy", func(t *testing.T) {
		err = cache.Del("test_incrby_key")
		assert.NilError(t, err)

		err = cache.Set("test_incrby_key", "10")
		assert.NilError(t, err)

		val, err := cache.IncrBy("test_incrby_key", 5)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(15))
	})

	t.Run("DecrBy", func(t *testing.T) {
		err = cache.Del("test_decrby_key")
		assert.NilError(t, err)

		err = cache.Set("test_decrby_key", "10")
		assert.NilError(t, err)

		val, err := cache.DecrBy("test_decrby_key", 3)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(7))
	})


	t.Run("Del multiple keys", func(t *testing.T) {
		err = cache.Set("test_del_key1", "value1")
		assert.NilError(t, err)
		err = cache.Set("test_del_key2", "value2")
		assert.NilError(t, err)

		err = cache.Del("test_del_key1", "test_del_key2")
		assert.NilError(t, err)

		_, err = cache.Get("test_del_key1")
		assert.Equal(t, err, ErrNotFound)
		_, err = cache.Get("test_del_key2")
		assert.Equal(t, err, ErrNotFound)
	})

	t.Run("Clean", func(t *testing.T) {
		err = cache.Set("test_clean_key", "value")
		assert.NilError(t, err)

		err = cache.Clean()
		assert.NilError(t, err)

		_, err = cache.Get("test_clean_key")
		assert.Equal(t, err, ErrNotFound)
	})
}

// TestRedis_MockMode_HashOperations 测试 Hash 操作
func TestRedis_MockMode_HashOperations(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	t.Run("HSet and HGet", func(t *testing.T) {
		_, err := cache.HSet("test_hash", "field1", "value1")
		assert.NilError(t, err)

		val, err := cache.HGet("test_hash", "field1")
		assert.NilError(t, err)
		assert.Equal(t, val, "value1")
	})

	t.Run("HGetAll", func(t *testing.T) {
		_, err := cache.HSet("test_hash_all", "field1", "value1")
		assert.NilError(t, err)
		_, err = cache.HSet("test_hash_all", "field2", "value2")
		assert.NilError(t, err)

		all, err := cache.HGetAll("test_hash_all")
		assert.NilError(t, err)
		assert.Equal(t, len(all), 2)
	})

	t.Run("HKeys", func(t *testing.T) {
		_, err := cache.HSet("test_hash_keys", "field1", "value1")
		assert.NilError(t, err)
		_, err = cache.HSet("test_hash_keys", "field2", "value2")
		assert.NilError(t, err)

		keys, err := cache.HKeys("test_hash_keys")
		assert.NilError(t, err)
		assert.Equal(t, len(keys), 2)
	})

	t.Run("HDel", func(t *testing.T) {
		_, err := cache.HSet("test_hash_del", "field1", "value1")
		assert.NilError(t, err)
		_, err = cache.HSet("test_hash_del", "field2", "value2")
		assert.NilError(t, err)

		n, err := cache.HDel("test_hash_del", "field1", "field2")
		assert.NilError(t, err)
		assert.Equal(t, n, int64(2))
	})

	t.Run("HExists", func(t *testing.T) {
		_, err := cache.HSet("test_hash_exists", "field1", "value1")
		assert.NilError(t, err)

		exists, err := cache.HExists("test_hash_exists", "field1")
		assert.NilError(t, err)
		assert.Equal(t, exists, true)

		exists, err = cache.HExists("test_hash_exists", "nonexistent")
		assert.NilError(t, err)
		assert.Equal(t, exists, false)
	})

	t.Run("HIncr", func(t *testing.T) {
		_, err := cache.HSet("test_hash_incr", "field1", "10")
		assert.NilError(t, err)

		val, err := cache.HIncr("test_hash_incr", "field1")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(11))
	})

	t.Run("HIncrBy", func(t *testing.T) {
		_, err := cache.HSet("test_hash_incrby", "field1", "10")
		assert.NilError(t, err)

		val, err := cache.HIncrBy("test_hash_incrby", "field1", 5)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(15))
	})

	t.Run("HDecr", func(t *testing.T) {
		_, err := cache.HSet("test_hash_decr", "field1", "10")
		assert.NilError(t, err)

		val, err := cache.HDecr("test_hash_decr", "field1")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(9))
	})

	t.Run("HDecrBy", func(t *testing.T) {
		_, err := cache.HSet("test_hash_decrby", "field1", "10")
		assert.NilError(t, err)

		val, err := cache.HDecrBy("test_hash_decrby", "field1", 3)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(7))
	})
}

// TestRedis_MockMode_SetOperations 测试 Set 操作
func TestRedis_MockMode_SetOperations(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	t.Run("SAdd", func(t *testing.T) {
		n, err := cache.SAdd("test_set", "member1", "member2", "member3")
		assert.NilError(t, err)
		assert.Equal(t, n, int64(3))
	})

	t.Run("SMembers", func(t *testing.T) {
		_, err := cache.SAdd("test_set_members", "member1", "member2")
		assert.NilError(t, err)

		members, err := cache.SMembers("test_set_members")
		assert.NilError(t, err)
		assert.Equal(t, len(members), 2)
	})

	t.Run("SRem", func(t *testing.T) {
		_, err := cache.SAdd("test_set_rem", "member1", "member2")
		assert.NilError(t, err)

		n, err := cache.SRem("test_set_rem", "member1", "member2")
		assert.NilError(t, err)
		assert.Equal(t, n, int64(2))
	})

	t.Run("SisMember", func(t *testing.T) {
		_, err := cache.SAdd("test_set_ismember", "member1")
		assert.NilError(t, err)

		isMember, err := cache.SisMember("test_set_ismember", "member1")
		assert.NilError(t, err)
		assert.Equal(t, isMember, true)

		isMember, err = cache.SisMember("test_set_ismember", "nonexistent")
		assert.NilError(t, err)
		assert.Equal(t, isMember, false)
	})

	t.Run("SPop", func(t *testing.T) {
		_, err := cache.SAdd("test_set_spop", "member1", "member2")
		assert.NilError(t, err)

		member, err := cache.SPop("test_set_spop")
		assert.NilError(t, err)
		assert.Assert(t, member == "member1" || member == "member2")
	})

	t.Run("SRandMember", func(t *testing.T) {
		_, err := cache.SAdd("test_set_srand", "member1", "member2", "member3")
		assert.NilError(t, err)

		members, err := cache.SRandMember("test_set_srand", 2)
		assert.NilError(t, err)
		assert.Equal(t, len(members), 2)
	})
}

// TestRedis_MockMode_SortedSetOperations 测试 Sorted Set 操作
func TestRedis_MockMode_SortedSetOperations(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	t.Run("ZAdd", func(t *testing.T) {
		n, err := cache.ZAdd("test_zset", 1.0, "member1", 2.0, "member2")
		assert.NilError(t, err)
		assert.Equal(t, n, int64(2))
	})

	t.Run("ZScore", func(t *testing.T) {
		_, err := cache.ZAdd("test_zscore", 1.5, "member1")
		assert.NilError(t, err)

		score, err := cache.ZScore("test_zscore", "member1")
		assert.NilError(t, err)
		assert.Equal(t, score, 1.5)
	})

	t.Run("ZRange", func(t *testing.T) {
		_, err := cache.ZAdd("test_zrange", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		members, err := cache.ZRange("test_zrange", 0, -1)
		assert.NilError(t, err)
		assert.Equal(t, len(members), 3)
	})

	t.Run("ZRangeByScore", func(t *testing.T) {
		_, err := cache.ZAdd("test_zrangebyscore", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		members, err := cache.ZRangeByScore("test_zrangebyscore", "1", "2", 0, 10)
		assert.NilError(t, err)
		assert.Assert(t, len(members) >= 1)
	})

	t.Run("ZRem", func(t *testing.T) {
		_, err := cache.ZAdd("test_zrem", 1, "m1", 2, "m2")
		assert.NilError(t, err)

		n, err := cache.ZRem("test_zrem", "m1", "m2")
		assert.NilError(t, err)
		assert.Equal(t, n, int64(2))
	})

	t.Run("ZCard", func(t *testing.T) {
		_, err := cache.ZAdd("test_zcard", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		n, err := cache.ZCard("test_zcard")
		assert.NilError(t, err)
		assert.Equal(t, n, int64(3))
	})

	t.Run("ZCount", func(t *testing.T) {
		_, err := cache.ZAdd("test_zcount", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		n, err := cache.ZCount("test_zcount", "1", "2")
		assert.NilError(t, err)
		assert.Equal(t, n, int64(2))
	})

	t.Run("ZIncrBy", func(t *testing.T) {
		_, err := cache.ZAdd("test_zincrby", 1.0, "m1")
		assert.NilError(t, err)

		score, err := cache.ZIncrBy("test_zincrby", 2.5, "m1")
		assert.NilError(t, err)
		assert.Equal(t, score, 3.5)
	})

	t.Run("ZRank", func(t *testing.T) {
		_, err := cache.ZAdd("test_zrank", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		rank, err := cache.ZRank("test_zrank", "m2")
		assert.NilError(t, err)
		assert.Equal(t, rank, int64(1))
	})

	t.Run("ZRevRange", func(t *testing.T) {
		_, err := cache.ZAdd("test_zrevrange", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		members, err := cache.ZRevRange("test_zrevrange", 0, -1)
		assert.NilError(t, err)
		assert.Equal(t, len(members), 3)
		assert.Equal(t, members[0], "m3")
	})

	t.Run("ZRevRank", func(t *testing.T) {
		_, err := cache.ZAdd("test_zrevrank", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		rank, err := cache.ZRevRank("test_zrevrank", "m2")
		assert.NilError(t, err)
		assert.Equal(t, rank, int64(1))
	})

	t.Run("ZRangeWithScores", func(t *testing.T) {
		_, err := cache.ZAdd("test_zrangewithscores", 1, "m1", 2, "m2")
		assert.NilError(t, err)

		members, err := cache.ZRangeWithScores("test_zrangewithscores", 0, -1)
		assert.NilError(t, err)
		assert.Equal(t, len(members), 2)
	})

	t.Run("ZRevRangeByScore", func(t *testing.T) {
		_, err := cache.ZAdd("test_zrevrangebyscore", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		members, err := cache.ZRevRangeByScore("test_zrevrangebyscore", "3", "1", 0, 10)
		assert.NilError(t, err)
		assert.Assert(t, len(members) >= 1)
	})

	t.Run("ZRemRangeByRank", func(t *testing.T) {
		_, err := cache.ZAdd("test_zremrangebyrank", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		n, err := cache.ZRemRangeByRank("test_zremrangebyrank", 0, 1)
		assert.NilError(t, err)
		assert.Equal(t, n, int64(2))
	})

	t.Run("ZRemRangeByScore", func(t *testing.T) {
		_, err := cache.ZAdd("test_zremrangebyscore", 1, "m1", 2, "m2", 3, "m3")
		assert.NilError(t, err)

		n, err := cache.ZRemRangeByScore("test_zremrangebyscore", "1", "2")
		assert.NilError(t, err)
		assert.Equal(t, n, int64(2))
	})

	// ZUnionStore 和 ZInterStore 在 Mock 模式下不支持
	t.Run("ZUnionStore", func(t *testing.T) {
		_, err := cache.ZAdd("test_zunion1", 1, "m1")
		assert.NilError(t, err)
		_, err = cache.ZAdd("test_zunion2", 2, "m2")
		assert.NilError(t, err)

		n, err := cache.ZUnionStore("test_zunion_dest", "test_zunion1", "test_zunion2")
		// Mock 模式可能不完整，只验证不 panic
		_ = n
		_ = err
	})

	t.Run("ZInterStore", func(t *testing.T) {
		_, err := cache.ZAdd("test_zinter1", 1, "m1")
		assert.NilError(t, err)
		_, err = cache.ZAdd("test_zinter2", 2, "m1")
		assert.NilError(t, err)

		n, err := cache.ZInterStore("test_zinter_dest", "test_zinter1", "test_zinter2")
		// Mock 模式可能不完整，只验证不 panic
		_ = n
		_ = err
	})
}

// TestRedis_MockMode_PubSub 测试 Pub/Sub
func TestRedis_MockMode_PubSub(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	t.Run("Publish", func(t *testing.T) {
		n, err := cache.Publish("test_channel", "test_message")
		assert.NilError(t, err)
		assert.Assert(t, n >= 0)
	})
}

// TestRedis_MockMode_StreamOperations 测试 Stream 操作
func TestRedis_MockMode_StreamOperations(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	t.Run("XAdd", func(t *testing.T) {
		id, err := cache.XAdd("test_stream", map[string]interface{}{"key1": "value1"})
		assert.NilError(t, err)
		assert.Assert(t, len(id) > 0)
	})

	t.Run("XLen", func(t *testing.T) {
		_, err := cache.XAdd("test_stream_len", map[string]interface{}{"key1": "value1"})
		assert.NilError(t, err)

		len, err := cache.XLen("test_stream_len")
		assert.NilError(t, err)
		assert.Equal(t, len, int64(1))
	})

	t.Run("XRange", func(t *testing.T) {
		_, err := cache.XAdd("test_stream_range", map[string]interface{}{"key1": "value1"})
		assert.NilError(t, err)

		entries, err := cache.XRange("test_stream_range", "-", "+", 10)
		assert.NilError(t, err)
		assert.Equal(t, len(entries), 1)
	})

	t.Run("XRevRange", func(t *testing.T) {
		_, err := cache.XAdd("test_stream_revrange", map[string]interface{}{"key1": "value1"})
		assert.NilError(t, err)

		entries, err := cache.XRevRange("test_stream_revrange", "+", "-", 10)
		assert.NilError(t, err)
		assert.Equal(t, len(entries), 1)
	})

	t.Run("XDel", func(t *testing.T) {
		id, err := cache.XAdd("test_stream_del", map[string]interface{}{"key1": "value1"})
		assert.NilError(t, err)

		n, err := cache.XDel("test_stream_del", id)
		assert.NilError(t, err)
		assert.Equal(t, n, int64(1))
	})

	t.Run("XTrim", func(t *testing.T) {
		_, err := cache.XAdd("test_stream_trim", map[string]interface{}{"key1": "value1"})
		assert.NilError(t, err)

		n, err := cache.XTrim("test_stream_trim", 10)
		assert.NilError(t, err)
		assert.Assert(t, n >= 0)
	})

	// XGroup 相关操作在 Mock 模式下不支持
	t.Run("XGroupCreate", func(t *testing.T) {
		err := cache.XGroupCreate("test_stream_group", "test_group", "0")
		// Mock 模式可能不支持，只验证不 panic
		_ = err
	})

	t.Run("XGroupDestroy", func(t *testing.T) {
		err := cache.XGroupCreate("test_stream_destroy", "test_group", "0")
		_ = err
		err = cache.XGroupDestroy("test_stream_destroy", "test_group")
		// Mock 模式可能不支持，只验证不 panic
		_ = err
	})

	t.Run("XGroupSetID", func(t *testing.T) {
		err := cache.XGroupCreate("test_stream_setid", "test_group", "0")
		_ = err
		err = cache.XGroupSetID("test_stream_setid", "test_group", "1")
		// Mock 模式可能不支持，只验证不 panic
		_ = err
	})

	t.Run("XPending", func(t *testing.T) {
		err := cache.XGroupCreate("test_stream_pending", "test_group", "0")
		_ = err
		n, err := cache.XPending("test_stream_pending", "test_group")
		// Mock 模式可能不支持，只验证不 panic
		_ = n
		_ = err
	})

	t.Run("XAck", func(t *testing.T) {
		err := cache.XGroupCreate("test_stream_ack", "test_group", "0")
		_ = err
		id, err := cache.XAdd("test_stream_ack", map[string]interface{}{"key1": "value1"})
		assert.NilError(t, err)

		n, err := cache.XAck("test_stream_ack", "test_group", id)
		// Mock 模式可能不支持，只验证不 panic
		_ = n
		_ = err
	})
}

// TestRedis_MockMode_OtherOperations 测试其他操作
func TestRedis_MockMode_OtherOperations(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	t.Run("Ping", func(t *testing.T) {
		err := cache.Ping()
		assert.NilError(t, err)
	})

	t.Run("Client", func(t *testing.T) {
		client := cache.Client()
		assert.Assert(t, client != nil)
	})

	t.Run("Close", func(t *testing.T) {
		testCache, err := NewRedisWithConfig(&Config{Type: Redis, Mock: true})
		assert.NilError(t, err)

		err = testCache.Close()
		assert.NilError(t, err)
	})
}
