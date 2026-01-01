package cache

import (
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"gotest.tools/v3/assert"
)

func TestRedisPubSub(t *testing.T) {
	cache, err := NewRedis("localhost:6379",
		redis.DialDatabase(0),
		redis.DialConnectTimeout(time.Second*3),
		redis.DialReadTimeout(time.Minute),
		redis.DialWriteTimeout(time.Minute),
	)
	if err != nil {
		t.Skipf("skip test: redis not available: %v", err)
		return
	}
	defer cache.Close()

	cache.SetPrefix("test:")

	channel := "test_channel"
	message := "hello world"

	msgChan, errChan, err := cache.Subscribe(channel)
	assert.NilError(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		count, err := cache.Publish(channel, message)
		assert.NilError(t, err)
		assert.Assert(t, count >= 0)
	}()

	select {
	case msg := <-msgChan:
		assert.Equal(t, string(msg), message)
	case err := <-errChan:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestRedisStream(t *testing.T) {
	cache, err := NewRedis("localhost:6379",
		redis.DialDatabase(0),
		redis.DialConnectTimeout(time.Second*3),
		redis.DialReadTimeout(time.Minute),
		redis.DialWriteTimeout(time.Minute),
	)
	if err != nil {
		t.Skipf("skip test: redis not available: %v", err)
		return
	}
	defer cache.Close()

	cache.SetPrefix("test:")

	streamKey := "test_stream"

	err = cache.Del(streamKey)
	assert.NilError(t, err)

	id, err := cache.XAdd(streamKey, map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	})
	assert.NilError(t, err)
	assert.Assert(t, id != "")

	length, err := cache.XLen(streamKey)
	assert.NilError(t, err)
	assert.Equal(t, length, int64(1))

	entries, err := cache.XRange(streamKey, "-", "+")
	assert.NilError(t, err)
	assert.Equal(t, len(entries), 1)
	assert.Equal(t, entries[0]["field1"], "value1")
	assert.Equal(t, entries[0]["field2"], "value2")

	revEntries, err := cache.XRevRange(streamKey, "+", "-")
	assert.NilError(t, err)
	assert.Equal(t, len(revEntries), 1)

	deletedCount, err := cache.XDel(streamKey, id)
	assert.NilError(t, err)
	assert.Equal(t, deletedCount, int64(1))

	length, err = cache.XLen(streamKey)
	assert.NilError(t, err)
	assert.Equal(t, length, int64(0))

	err = cache.Del(streamKey)
	assert.NilError(t, err)
}

func TestRedisStreamTrim(t *testing.T) {
	cache, err := NewRedis("localhost:6379",
		redis.DialDatabase(0),
		redis.DialConnectTimeout(time.Second*3),
		redis.DialReadTimeout(time.Minute),
		redis.DialWriteTimeout(time.Minute),
	)
	if err != nil {
		t.Skipf("skip test: redis not available: %v", err)
		return
	}
	defer cache.Close()

	cache.SetPrefix("test:")

	streamKey := "test_stream_trim"

	err = cache.Del(streamKey)
	assert.NilError(t, err)

	for i := 0; i < 10; i++ {
		_, err := cache.XAdd(streamKey, map[string]interface{}{
			"index": i,
		})
		assert.NilError(t, err)
	}

	length, err := cache.XLen(streamKey)
	assert.NilError(t, err)
	assert.Equal(t, length, int64(10))

	trimmed, err := cache.XTrim(streamKey, 5)
	assert.NilError(t, err)
	assert.Assert(t, trimmed >= 0)

	length, err = cache.XLen(streamKey)
	assert.NilError(t, err)
	assert.Assert(t, length <= 5)

	err = cache.Del(streamKey)
	assert.NilError(t, err)
}
