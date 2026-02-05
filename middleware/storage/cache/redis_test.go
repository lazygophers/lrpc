package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gotest.tools/v3/assert"
)

// TestRedis_AllFunctionsCoverage 全面测试 Redis 所有功能
func TestRedis_AllFunctionsCoverage(t *testing.T) {
	// 创建 miniredis 实例
	mr := miniredis.RunT(t)
	defer mr.Close()

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 直接创建 CacheRedis 实例
	redisCache := &CacheRedis{
		cli:       client,
		prefix:    "test:",
		ctx:       context.Background(),
		miniRedis: mr,
	}

	// 测试连接
	err := redisCache.Ping()
	assert.NilError(t, err)

	defer redisCache.Close()

	// 同时创建 baseCache 用于测试接口方法
	cache := newBaseCache(redisCache)
	defer cache.Close()

	t.Run("Clean", func(t *testing.T) {
		_ = cache.Set("key1", "value1")
		_ = cache.Set("key2", "value2")

		err := redisCache.Clean()
		assert.NilError(t, err)

		exists, _ := cache.Exists("key1")
		assert.Equal(t, exists, false)
	})

	t.Run("SetPrefix", func(t *testing.T) {
		redisCache.SetPrefix("newprefix:")

		err := cache.Set("test_key", "test_value")
		assert.NilError(t, err)

		val, err := cache.Get("test_key")
		assert.NilError(t, err)
		assert.Equal(t, val, "test_value")

		redisCache.SetPrefix("test:")
	})

	t.Run("Incr", func(t *testing.T) {
		val, err := cache.Incr("counter")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(1))

		val, err = cache.Incr("counter")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(2))
	})

	t.Run("Decr", func(t *testing.T) {
		_, _ = cache.Incr("decr_counter")

		val, err := cache.Decr("decr_counter")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(0))
	})

	t.Run("IncrBy", func(t *testing.T) {
		val, err := cache.IncrBy("incrby_counter", 10)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(10))

		val, err = cache.IncrBy("incrby_counter", 5)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(15))
	})

	t.Run("IncrByFloat", func(t *testing.T) {
		val, err := redisCache.IncrByFloat("float_counter", 10.5)
		assert.NilError(t, err)
		assert.Equal(t, val, 10.5)

		val, err = redisCache.IncrByFloat("float_counter", 0.5)
		assert.NilError(t, err)
		assert.Equal(t, val, 11.0)
	})

	t.Run("DecrBy", func(t *testing.T) {
		val, err := cache.DecrBy("decrby_counter", 5)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(-5))
	})

	t.Run("Get", func(t *testing.T) {
		err := cache.Set("get_key", "get_value")
		assert.NilError(t, err)

		val, err := cache.Get("get_key")
		assert.NilError(t, err)
		assert.Equal(t, val, "get_value")

		_, err = cache.Get("nonexistent")
		assert.Equal(t, err, ErrNotFound)
	})

	t.Run("Exists", func(t *testing.T) {
		_ = cache.Set("exists_key", "value")

		exists, err := cache.Exists("exists_key")
		assert.NilError(t, err)
		assert.Equal(t, exists, true)

		exists, err = cache.Exists("nonexistent")
		assert.NilError(t, err)
		assert.Equal(t, exists, false)

		exists, err = cache.Exists("exists_key", "nonexistent")
		assert.NilError(t, err)
		assert.Equal(t, exists, true)
	})

	t.Run("SetNx", func(t *testing.T) {
		ok, err := cache.SetNx("nx_key", "value1")
		assert.NilError(t, err)
		assert.Equal(t, ok, true)

		ok, err = cache.SetNx("nx_key", "value2")
		assert.NilError(t, err)
		assert.Equal(t, ok, false)

		val, _ := cache.Get("nx_key")
		assert.Equal(t, val, "value1")
	})

	t.Run("Expire", func(t *testing.T) {
		_ = cache.Set("expire_key", "value")

		ok, err := cache.Expire("expire_key", 1*time.Minute)
		assert.NilError(t, err)
		assert.Equal(t, ok, true)

		ttl, err := cache.Ttl("expire_key")
		assert.NilError(t, err)
		assert.Assert(t, ttl > 0)
	})

	t.Run("Ttl", func(t *testing.T) {
		_ = cache.SetEx("ttl_key", "value", 2*time.Minute)

		ttl, err := cache.Ttl("ttl_key")
		assert.NilError(t, err)
		assert.Assert(t, ttl > 0)

		ttl, err = cache.Ttl("nonexistent")
		assert.NilError(t, err)
		assert.Equal(t, ttl, ttlDuration(-2))
	})

	t.Run("Set", func(t *testing.T) {
		err := cache.Set("set_key", "set_value")
		assert.NilError(t, err)

		val, err := cache.Get("set_key")
		assert.NilError(t, err)
		assert.Equal(t, val, "set_value")
	})

	t.Run("SetEx", func(t *testing.T) {
		err := cache.SetEx("setex_key", "setex_value", 5*time.Minute)
		assert.NilError(t, err)

		val, err := cache.Get("setex_key")
		assert.NilError(t, err)
		assert.Equal(t, val, "setex_value")

		ttl, _ := cache.Ttl("setex_key")
		assert.Assert(t, ttl > 0)
	})

	t.Run("SetNxWithTimeout", func(t *testing.T) {
		ok, err := cache.SetNxWithTimeout("nx_ex_key", "value1", 1*time.Minute)
		assert.NilError(t, err)
		assert.Equal(t, ok, true)

		ok, err = cache.SetNxWithTimeout("nx_ex_key", "value2", 1*time.Minute)
		assert.NilError(t, err)
		assert.Equal(t, ok, false)

		val, _ := cache.Get("nx_ex_key")
		assert.Equal(t, val, "value1")

		ttl, _ := cache.Ttl("nx_ex_key")
		assert.Assert(t, ttl > 0)
	})

	t.Run("Del", func(t *testing.T) {
		_ = cache.Set("del_key1", "value1")
		_ = cache.Set("del_key2", "value2")

		err := cache.Del("del_key1")
		assert.NilError(t, err)

		exists, _ := cache.Exists("del_key1")
		assert.Equal(t, exists, false)

		err = cache.Del("del_key2")
		assert.NilError(t, err)
	})

	t.Run("HSet", func(t *testing.T) {
		ok, err := cache.HSet("hash_key", "field1", "value1")
		assert.NilError(t, err)
		assert.Equal(t, ok, true)

		ok, err = cache.HSet("hash_key", "field2", "value2")
		assert.NilError(t, err)
	})

	t.Run("HGet", func(t *testing.T) {
		_, _ = cache.HSet("hget_key", "field1", "value1")

		val, err := cache.HGet("hget_key", "field1")
		assert.NilError(t, err)
		assert.Equal(t, val, "value1")

		_, err = cache.HGet("hget_key", "nonexistent_field")
		assert.Equal(t, err, ErrNotFound)
	})

	t.Run("HGetAll", func(t *testing.T) {
		_, _ = cache.HSet("hgetall_key", "field1", "value1")
		_, _ = cache.HSet("hgetall_key", "field2", "value2")

		all, err := cache.HGetAll("hgetall_key")
		assert.NilError(t, err)
		assert.Equal(t, len(all), 2)
		assert.Equal(t, all["field1"], "value1")
		assert.Equal(t, all["field2"], "value2")
	})

	t.Run("HKeys", func(t *testing.T) {
		_, _ = cache.HSet("hkeys_key", "field1", "value1")
		_, _ = cache.HSet("hkeys_key", "field2", "value2")

		keys, err := cache.HKeys("hkeys_key")
		assert.NilError(t, err)
		assert.Equal(t, len(keys), 2)
	})

	t.Run("HVals", func(t *testing.T) {
		_, _ = cache.HSet("hvals_key", "field1", "value1")
		_, _ = cache.HSet("hvals_key", "field2", "value2")

		vals, err := redisCache.HVals("hvals_key")
		assert.NilError(t, err)
		assert.Equal(t, len(vals), 2)
	})

	t.Run("HDel", func(t *testing.T) {
		_, _ = cache.HSet("hdel_key", "field1", "value1")
		_, _ = cache.HSet("hdel_key", "field2", "value2")

		count, err := cache.HDel("hdel_key", "field1", "field2")
		assert.NilError(t, err)
		assert.Equal(t, count, int64(2))
	})

	t.Run("SAdd", func(t *testing.T) {
		count, err := cache.SAdd("set_key", "member1", "member2", "member3")
		assert.NilError(t, err)
		assert.Equal(t, count, int64(3))

		count, err = cache.SAdd("set_key", "member1", "member4")
		assert.NilError(t, err)
		assert.Equal(t, count, int64(1))
	})

	t.Run("SMembers", func(t *testing.T) {
		_, _ = cache.SAdd("smembers_key", "member1", "member2")

		members, err := cache.SMembers("smembers_key")
		assert.NilError(t, err)
		assert.Equal(t, len(members), 2)
	})

	t.Run("SRem", func(t *testing.T) {
		_, _ = cache.SAdd("srem_key", "member1", "member2", "member3")

		count, err := cache.SRem("srem_key", "member1", "member2")
		assert.NilError(t, err)
		assert.Equal(t, count, int64(2))

		members, _ := cache.SMembers("srem_key")
		assert.Equal(t, len(members), 1)
	})

	t.Run("SRandMember", func(t *testing.T) {
		_, _ = cache.SAdd("srand_key", "member1", "member2", "member3")

		members, err := cache.SRandMember("srand_key")
		assert.NilError(t, err)
		assert.Equal(t, len(members), 1)

		members, err = cache.SRandMember("srand_key", 2)
		assert.NilError(t, err)
		assert.Assert(t, len(members) <= 2)
	})

	t.Run("SPop", func(t *testing.T) {
		_, _ = cache.SAdd("spop_key", "member1", "member2")

		member, err := cache.SPop("spop_key")
		assert.NilError(t, err)
		assert.Assert(t, member != "")

		members, _ := cache.SMembers("spop_key")
		assert.Equal(t, len(members), 1)
	})

	t.Run("HIncr", func(t *testing.T) {
		val, err := cache.HIncr("hincr_key", "field1")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(1))

		val, err = cache.HIncr("hincr_key", "field1")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(2))
	})

	t.Run("HIncrBy", func(t *testing.T) {
		val, err := cache.HIncrBy("hincrby_key", "field1", 10)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(10))

		val, err = cache.HIncrBy("hincrby_key", "field1", 5)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(15))
	})

	t.Run("HIncrByFloat", func(t *testing.T) {
		val, err := redisCache.HIncrByFloat("hincrbyfloat_key", "field1", 10.5)
		assert.NilError(t, err)
		assert.Equal(t, val, 10.5)

		val, err = redisCache.HIncrByFloat("hincrbyfloat_key", "field1", 0.5)
		assert.NilError(t, err)
		assert.Equal(t, val, 11.0)
	})

	t.Run("HDecr", func(t *testing.T) {
		_, _ = cache.HIncr("hdecr_key", "field1")

		val, err := cache.HDecr("hdecr_key", "field1")
		assert.NilError(t, err)
		assert.Equal(t, val, int64(0))
	})

	t.Run("HDecrBy", func(t *testing.T) {
		val, err := cache.HDecrBy("hdecrby_key", "field1", 5)
		assert.NilError(t, err)
		assert.Equal(t, val, int64(-5))
	})

	t.Run("HExists", func(t *testing.T) {
		_, _ = cache.HSet("hexists_key", "field1", "value1")

		exists, err := cache.HExists("hexists_key", "field1")
		assert.NilError(t, err)
		assert.Equal(t, exists, true)

		exists, err = cache.HExists("hexists_key", "nonexistent_field")
		assert.NilError(t, err)
		assert.Equal(t, exists, false)
	})

	t.Run("SisMember", func(t *testing.T) {
		_, _ = cache.SAdd("sismember_key", "member1")

		isMember, err := cache.SisMember("sismember_key", "member1")
		assert.NilError(t, err)
		assert.Equal(t, isMember, true)

		isMember, err = cache.SisMember("sismember_key", "nonexistent")
		assert.NilError(t, err)
		assert.Equal(t, isMember, false)
	})

	t.Run("Close", func(t *testing.T) {
		testMr := miniredis.RunT(t)
		testClient := redis.NewClient(&redis.Options{
			Addr: testMr.Addr(),
		})

		testCache := &CacheRedis{
			cli:       testClient,
			prefix:    "close:",
			ctx:       context.Background(),
			miniRedis: testMr,
		}

		err := testCache.Close()
		assert.NilError(t, err)
	})

	t.Run("Ping", func(t *testing.T) {
		err := cache.Ping()
		assert.NilError(t, err)
	})

	t.Run("XAdd", func(t *testing.T) {
		values := map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		}

		id, err := cache.XAdd("stream1", values)
		assert.NilError(t, err)
		assert.Assert(t, id != "")
	})

	t.Run("XLen", func(t *testing.T) {
		_, _ = cache.XAdd("xlen_stream", map[string]interface{}{"data": "test1"})
		_, _ = cache.XAdd("xlen_stream", map[string]interface{}{"data": "test2"})

		length, err := cache.XLen("xlen_stream")
		assert.NilError(t, err)
		assert.Equal(t, length, int64(2))
	})

	t.Run("XRange", func(t *testing.T) {
		_, _ = cache.XAdd("xrange_stream", map[string]interface{}{"data": "value1"})
		_, _ = cache.XAdd("xrange_stream", map[string]interface{}{"data": "value2"})

		messages, err := cache.XRange("xrange_stream", "-", "+")
		assert.NilError(t, err)
		assert.Equal(t, len(messages), 2)

		messages, err = cache.XRange("xrange_stream", "-", "+", 1)
		assert.NilError(t, err)
		assert.Equal(t, len(messages), 1)
	})

	t.Run("XRevRange", func(t *testing.T) {
		_, _ = cache.XAdd("xrevrange_stream", map[string]interface{}{"data": "value1"})
		_, _ = cache.XAdd("xrevrange_stream", map[string]interface{}{"data": "value2"})

		messages, err := cache.XRevRange("xrevrange_stream", "+", "-")
		assert.NilError(t, err)
		assert.Equal(t, len(messages), 2)

		messages, err = cache.XRevRange("xrevrange_stream", "+", "-", 1)
		assert.NilError(t, err)
		assert.Equal(t, len(messages), 1)
	})

	t.Run("XDel", func(t *testing.T) {
		id1, _ := cache.XAdd("xdel_stream", map[string]interface{}{"data": "value1"})
		_, _ = cache.XAdd("xdel_stream", map[string]interface{}{"data": "value2"})

		count, err := cache.XDel("xdel_stream", id1)
		assert.NilError(t, err)
		assert.Equal(t, count, int64(1))

		length, _ := cache.XLen("xdel_stream")
		assert.Equal(t, length, int64(1))

		count, err = cache.XDel("xdel_stream", "999999-0")
		assert.NilError(t, err)
		assert.Equal(t, count, int64(0))
	})

	t.Run("XTrim", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			_, _ = cache.XAdd("xtrim_stream", map[string]interface{}{"data": fmt.Sprintf("value%d", i)})
		}

		count, err := cache.XTrim("xtrim_stream", 5)
		assert.NilError(t, err)
		assert.Assert(t, count > 0)

		length, _ := cache.XLen("xtrim_stream")
		assert.Equal(t, length, int64(5))
	})

	t.Run("XGroupCreate", func(t *testing.T) {
		_, _ = cache.XAdd("xgroup_stream", map[string]interface{}{"data": "value1"})

		err := cache.XGroupCreate("xgroup_stream", "group1", "0")
		assert.NilError(t, err)

		err = cache.XGroupCreate("xgroup_stream", "group1", "0")
		assert.Assert(t, err != nil)
	})

	t.Run("XGroupDestroy", func(t *testing.T) {
		_, _ = cache.XAdd("xgroup_destroy_stream", map[string]interface{}{"data": "value1"})
		_ = cache.XGroupCreate("xgroup_destroy_stream", "group1", "0")

		err := cache.XGroupDestroy("xgroup_destroy_stream", "group1")
		assert.NilError(t, err)

		_ = cache.XGroupDestroy("xgroup_destroy_stream", "nonexistent_group")
	})

	t.Run("XGroupSetID", func(t *testing.T) {
		_, _ = cache.XAdd("xgroup_setid_stream", map[string]interface{}{"data": "value1"})
		_ = cache.XGroupCreate("xgroup_setid_stream", "group1", "0")

		err := cache.XGroupSetID("xgroup_setid_stream", "group1", "$")
		assert.NilError(t, err)
	})

	t.Run("XReadGroup", func(t *testing.T) {
		_, _ = cache.XAdd("xreadgroup_stream", map[string]interface{}{"data": "test"})
		_ = cache.XGroupCreate("xreadgroup_stream", "testgroup", "0")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		go func() {
			redisCache.ctx = ctx
			err := redisCache.XReadGroup(
				func(stream string, id string, body []byte) error {
					assert.Equal(t, stream, "xreadgroup_stream")
					assert.Assert(t, id != "")
					assert.Assert(t, len(body) > 0)
					return nil
				},
				"testgroup",
				"consumer1",
				"xreadgroup_stream",
			)
			_ = err
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("XReadGroup did not complete in time")
		}
	})

	t.Run("XAck", func(t *testing.T) {
		id, _ := cache.XAdd("xack_stream", map[string]interface{}{"data": "value1"})
		_ = cache.XGroupCreate("xack_stream", "group1", "0")

		streams, err := redisCache.cli.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    "group1",
			Consumer: "consumer1",
			Streams:  []string{"test:xack_stream", ">"},
			Count:    1,
		}).Result()
		assert.NilError(t, err)

		if len(streams) > 0 && len(streams[0].Messages) > 0 {
			msgID := streams[0].Messages[0].ID

			count, err := cache.XAck("xack_stream", "group1", msgID)
			assert.NilError(t, err)
			assert.Equal(t, count, int64(1))

			count, err = cache.XAck("xack_stream", "group1", msgID)
			assert.NilError(t, err)
			assert.Equal(t, count, int64(0))
		} else {
			count, err := cache.XAck("xack_stream", "group1", id)
			assert.NilError(t, err)
			assert.Assert(t, count >= 0 && count <= int64(1))
		}
	})

	t.Run("XPending", func(t *testing.T) {
		_, _ = cache.XAdd("xpending_stream", map[string]interface{}{"data": "value1"})
		_ = cache.XGroupCreate("xpending_stream", "group1", "0")

		_, _ = redisCache.cli.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    "group1",
			Consumer: "consumer1",
			Streams:  []string{"test:xpending_stream", ">"},
			Count:    1,
		}).Result()

		count, err := cache.XPending("xpending_stream", "group1")
		assert.NilError(t, err)
		assert.Assert(t, count >= 0)
	})

	t.Run("Publish", func(t *testing.T) {
		count, err := cache.Publish("channel1", "message1")
		assert.NilError(t, err)
		assert.Equal(t, count, int64(0))
	})

	t.Run("Subscribe", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		done := make(chan bool)
		received := make(chan bool)

		go func() {
			redisCache.ctx = ctx
			err := cache.Subscribe(
				func(channel string, message []byte) error {
					assert.Equal(t, channel, "test_channel")
					assert.Equal(t, string(message), "test_message")
					received <- true
					return nil
				},
				"test_channel",
			)
			_ = err
			done <- true
		}()

		time.Sleep(10 * time.Millisecond)

		_, _ = cache.Publish("test_channel", "test_message")

		select {
		case <-received:
		case <-time.After(50 * time.Millisecond):
		}

		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("Subscribe did not complete in time")
		}
	})

	t.Run("Client", func(t *testing.T) {
		client := cache.Client()
		assert.Assert(t, client != nil)

		_, ok := client.(*redis.Client)
		assert.Equal(t, ok, true)
	})
}

// TestRedis_ErrorCases 测试关键的错误情况
func TestRedis_ErrorCases(t *testing.T) {
	t.Run("Get_NotFound", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		_, err := redisCache.Get("nonexistent")
		assert.Equal(t, err, ErrNotFound)
	})

	t.Run("HGet_NotFound", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		_, err := redisCache.HGet("test", "field")
		assert.Equal(t, err, ErrNotFound)
	})

	t.Run("SPop_EmptySet", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		_, err := redisCache.SPop("empty_set")
		assert.Assert(t, err != nil)
	})

	t.Run("XGroupCreate_AlreadyExists", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		_, _ = redisCache.XAdd("stream", map[string]interface{}{"data": "test"})
		_ = redisCache.XGroupCreate("stream", "group", "0")

		err := redisCache.XGroupCreate("stream", "group", "0")
		assert.Assert(t, err != nil)
	})

	t.Run("Ttl_KeyNotFound", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		ttl, err := redisCache.Ttl("nonexistent")
		assert.NilError(t, err)
		assert.Assert(t, ttl < 0)
	})

	t.Run("SRandMember_EmptySet", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		members, err := redisCache.SRandMember("empty_set")
		assert.NilError(t, err)
		assert.Equal(t, len(members), 0)
	})

	t.Run("HGetAll_EmptyHash", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		all, err := redisCache.HGetAll("empty_hash")
		assert.NilError(t, err)
		assert.Equal(t, len(all), 0)
	})

	t.Run("HKeys_EmptyHash", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		keys, err := redisCache.HKeys("empty_hash")
		assert.NilError(t, err)
		assert.Equal(t, len(keys), 0)
	})

	t.Run("HVals_EmptyHash", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		vals, err := redisCache.HVals("empty_hash")
		assert.NilError(t, err)
		assert.Equal(t, len(vals), 0)
	})

	t.Run("SMembers_EmptySet", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		members, err := redisCache.SMembers("empty_set")
		assert.NilError(t, err)
		assert.Equal(t, len(members), 0)
	})

	t.Run("XLen_EmptyStream", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		length, err := redisCache.XLen("empty_stream")
		assert.NilError(t, err)
		assert.Equal(t, length, int64(0))
	})

	t.Run("XRange_EmptyStream", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		messages, err := redisCache.XRange("empty_stream", "-", "+")
		assert.NilError(t, err)
		assert.Equal(t, len(messages), 0)
	})

	t.Run("XDel_NonExistentID", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		count, err := redisCache.XDel("stream", "999999-0")
		assert.NilError(t, err)
		assert.Equal(t, count, int64(0))
	})

	t.Run("XTrim_EmptyStream", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		redisCache := &CacheRedis{
			cli:       client,
			prefix:    "test:",
			ctx:       context.Background(),
			miniRedis: mr,
		}

		count, err := redisCache.XTrim("empty_stream", 10)
		assert.NilError(t, err)
		assert.Equal(t, count, int64(0))
	})
}

// TestRedis_ConnectionErrors 测试连接错误情况
func TestRedis_ConnectionErrors(t *testing.T) {
	testMethods := []struct {
		name string
		test func(*CacheRedis) error
	}{
		{"Incr", func(c *CacheRedis) error { _, err := c.Incr("test"); return err }},
		{"Decr", func(c *CacheRedis) error { _, err := c.Decr("test"); return err }},
		{"IncrBy", func(c *CacheRedis) error { _, err := c.IncrBy("test", 5); return err }},
		{"DecrBy", func(c *CacheRedis) error { _, err := c.DecrBy("test", 5); return err }},
		{"IncrByFloat", func(c *CacheRedis) error { _, err := c.IncrByFloat("test", 1.5); return err }},
		{"Get", func(c *CacheRedis) error { _, err := c.Get("test"); return err }},
		{"Set", func(c *CacheRedis) error { return c.Set("test", "value") }},
		{"SetEx", func(c *CacheRedis) error { return c.SetEx("test", "value", 1000) }},
		{"SetNx", func(c *CacheRedis) error { _, err := c.SetNx("test", "value"); return err }},
		{"SetNxWithTimeout", func(c *CacheRedis) error { _, err := c.SetNxWithTimeout("test", "value", 1000); return err }},
		{"Del", func(c *CacheRedis) error { return c.Del("test") }},
		{"Expire", func(c *CacheRedis) error { _, err := c.Expire("test", 1000); return err }},
		{"Ttl", func(c *CacheRedis) error { _, err := c.Ttl("test"); return err }},
		{"Exists", func(c *CacheRedis) error { _, err := c.Exists("test"); return err }},
		{"HSet", func(c *CacheRedis) error { _, err := c.HSet("test", "field", "value"); return err }},
		{"HGet", func(c *CacheRedis) error { _, err := c.HGet("test", "field"); return err }},
		{"HGetAll", func(c *CacheRedis) error { _, err := c.HGetAll("test"); return err }},
		{"HKeys", func(c *CacheRedis) error { _, err := c.HKeys("test"); return err }},
		{"HVals", func(c *CacheRedis) error { _, err := c.HVals("test"); return err }},
		{"HDel", func(c *CacheRedis) error { _, err := c.HDel("test", "field"); return err }},
		{"HIncr", func(c *CacheRedis) error { _, err := c.HIncr("test", "field"); return err }},
		{"HIncrBy", func(c *CacheRedis) error { _, err := c.HIncrBy("test", "field", 5); return err }},
		{"HIncrByFloat", func(c *CacheRedis) error { _, err := c.HIncrByFloat("test", "field", 1.5); return err }},
		{"HDecr", func(c *CacheRedis) error { _, err := c.HDecr("test", "field"); return err }},
		{"HDecrBy", func(c *CacheRedis) error { _, err := c.HDecrBy("test", "field", 5); return err }},
		{"HExists", func(c *CacheRedis) error { _, err := c.HExists("test", "field"); return err }},
		{"SAdd", func(c *CacheRedis) error { _, err := c.SAdd("test", "member"); return err }},
		{"SMembers", func(c *CacheRedis) error { _, err := c.SMembers("test"); return err }},
		{"SRem", func(c *CacheRedis) error { _, err := c.SRem("test", "member"); return err }},
		{"SRandMember", func(c *CacheRedis) error { _, err := c.SRandMember("test"); return err }},
		{"SPop", func(c *CacheRedis) error { _, err := c.SPop("test"); return err }},
		{"SisMember", func(c *CacheRedis) error { _, err := c.SisMember("test", "member"); return err }},
		{"Ping", func(c *CacheRedis) error { return c.Ping() }},
		{"Publish", func(c *CacheRedis) error { _, err := c.Publish("channel", "message"); return err }},
		{"XAdd", func(c *CacheRedis) error {
			_, err := c.XAdd("stream", map[string]interface{}{"data": "test"})
			return err
		}},
		{"XLen", func(c *CacheRedis) error { _, err := c.XLen("stream"); return err }},
		{"XRange", func(c *CacheRedis) error { _, err := c.XRange("stream", "-", "+"); return err }},
		{"XRevRange", func(c *CacheRedis) error { _, err := c.XRevRange("stream", "+", "-"); return err }},
		{"XDel", func(c *CacheRedis) error { _, err := c.XDel("stream", "0-0"); return err }},
		{"XTrim", func(c *CacheRedis) error { _, err := c.XTrim("stream", 10); return err }},
		{"XGroupCreate", func(c *CacheRedis) error { return c.XGroupCreate("stream", "group", "0") }},
		{"XGroupDestroy", func(c *CacheRedis) error { return c.XGroupDestroy("stream", "group") }},
		{"XGroupSetID", func(c *CacheRedis) error { return c.XGroupSetID("stream", "group", "$") }},
		{"XAck", func(c *CacheRedis) error { _, err := c.XAck("stream", "group", "0-0"); return err }},
		{"XPending", func(c *CacheRedis) error { _, err := c.XPending("stream", "group"); return err }},
	}

	for _, tm := range testMethods {
		t.Run(tm.name, func(t *testing.T) {
			mr := miniredis.RunT(t)
			addr := mr.Addr()
			mr.Close()

			client := redis.NewClient(&redis.Options{Addr: addr})
			redisCache := &CacheRedis{
				cli:       client,
				prefix:    "test:",
				ctx:       context.Background(),
				miniRedis: mr,
			}

			err := tm.test(redisCache)
			assert.Assert(t, err != nil, tm.name+" should return error")
		})
	}
}

// TestRedis_MockMode 测试 Redis Mock 模式
func TestRedis_MockMode(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	err = cache.Set("test_key", "test_value")
	assert.NilError(t, err)

	val, err := cache.Get("test_key")
	assert.NilError(t, err)
	assert.Equal(t, val, "test_value")

	err = cache.Del("test_key")
	assert.NilError(t, err)

	_, err = cache.Get("test_key")
	assert.Equal(t, err, ErrNotFound)
}

// TestRedis_MockMode_Ping 测试 Mock 模式的 Ping
func TestRedis_MockMode_Ping(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	err = cache.Ping()
	assert.NilError(t, err)
}

// TestRedis_MockMode_Exists 测试 Mock 模式的 Exists
func TestRedis_MockMode_Exists(t *testing.T) {
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := NewRedisWithConfig(config)
	assert.NilError(t, err)
	defer cache.Close()

	exists, err := cache.Exists("nonexistent")
	assert.NilError(t, err)
	assert.Equal(t, exists, false)

	err = cache.Set("exists_key", "value")
	assert.NilError(t, err)

	exists, err = cache.Exists("exists_key")
	assert.NilError(t, err)
	assert.Equal(t, exists, true)
}

// TestRedis_NewRedisWithConfig 测试使用 Config 创建
func TestRedis_NewRedisWithConfig(t *testing.T) {
	t.Run("MockMode", func(t *testing.T) {
		config := &Config{
			Type: Redis,
			Mock: true,
		}

		cache, err := NewRedisWithConfig(config)
		assert.NilError(t, err)
		defer cache.Close()

		err = cache.Set("test", "value")
		assert.NilError(t, err)

		val, err := cache.Get("test")
		assert.NilError(t, err)
		assert.Equal(t, val, "value")
	})

	t.Run("NilConfig", func(t *testing.T) {
		cache, err := NewRedisWithConfig(nil)
		assert.Assert(t, err != nil || cache != nil)
	})
}

// TestRedis_NewRedis 测试旧的构造函数（兼容性）
func TestRedis_NewRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	cache, err := NewRedis(mr.Addr(), map[string]interface{}{
		"db":       int64(0),
		"password": "",
	})
	assert.NilError(t, err)
	defer cache.Close()

	err = cache.Set("test", "value")
	assert.NilError(t, err)

	val, err := cache.Get("test")
	assert.NilError(t, err)
	assert.Equal(t, val, "value")
}

// TestRedis_NewRedisWithClient 测试使用已有 client 创建
func TestRedis_NewRedisWithClient(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cache, err := NewRedisWithClient(client, "custom:")
	assert.NilError(t, err)
	defer cache.Close()

	err = cache.Set("test", "value")
	assert.NilError(t, err)

	exists := mr.Exists("custom:test")
	assert.Equal(t, exists, true)
}

// TestRedis_ErrorHandling 测试错误处理
func TestRedis_ErrorHandling(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:9999",
	})

	_, err := NewRedisWithClient(client, "")
	assert.Assert(t, err != nil)
}

func ttlDuration(d int64) time.Duration {
	return time.Duration(d)
}
