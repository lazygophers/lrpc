package cache

import (
	"errors"
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

	receivedChan := make(chan bool, 1)
	doneChan := make(chan error, 1)

	go func() {
		err := cache.Subscribe(func(ch string, msg []byte) error {
			assert.Equal(t, ch, channel)
			assert.Equal(t, string(msg), message)
			receivedChan <- true
			return errors.New("test done")
		}, channel)
		doneChan <- err
	}()

	time.Sleep(500 * time.Millisecond)

	count, err := cache.Publish(channel, message)
	assert.NilError(t, err)
	assert.Assert(t, count >= 0)

	select {
	case <-receivedChan:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	select {
	case err := <-doneChan:
		assert.Error(t, err, "test done")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscribe to finish")
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

func TestRedisPubSubPanicRecovery(t *testing.T) {
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

	channel := "test_panic_channel"

	panicCount := 0
	normalCount := 0
	doneChan := make(chan error, 1)

	go func() {
		err := cache.Subscribe(func(ch string, msg []byte) error {
			message := string(msg)

			if message == "panic" {
				panicCount++
				panic("test panic")
			}

			if message == "normal" {
				normalCount++
			}

			if normalCount >= 2 {
				return errors.New("test done")
			}

			return nil
		}, channel)
		doneChan <- err
	}()

	time.Sleep(500 * time.Millisecond)

	_, err = cache.Publish(channel, "panic")
	assert.NilError(t, err)

	time.Sleep(100 * time.Millisecond)

	_, err = cache.Publish(channel, "normal")
	assert.NilError(t, err)

	time.Sleep(100 * time.Millisecond)

	_, err = cache.Publish(channel, "normal")
	assert.NilError(t, err)

	select {
	case err := <-doneChan:
		assert.Error(t, err, "test done")
		assert.Equal(t, panicCount, 1)
		assert.Equal(t, normalCount, 2)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for subscribe to finish")
	}
}
