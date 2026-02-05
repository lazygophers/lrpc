package cache

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestCacheMem_PubSub(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	// 测试订阅和发布
	receivedMessages := make(chan string, 2)
	err := cache.Subscribe(func(channel string, message []byte) error {
		receivedMessages <- channel + ":" + string(message)
		return nil
	}, "channel1", "channel2")
	assert.NilError(t, err)

	// 发布消息到 channel1
	count, err := cache.Publish("channel1", "message1")
	assert.NilError(t, err)
	assert.Equal(t, count, int64(1))

	// 发布消息到 channel2
	count, err = cache.Publish("channel2", "message2")
	assert.NilError(t, err)
	assert.Equal(t, count, int64(1))

	// 发布到不存在的频道
	count, err = cache.Publish("channel3", "message3")
	assert.NilError(t, err)
	assert.Equal(t, count, int64(0))

	// 接收消息（异步发送，可能需要短暂等待）
	timeout := time.After(100 * time.Millisecond)
	received := []string{}
	for i := 0; i < 2; i++ {
		select {
		case msg := <-receivedMessages:
			received = append(received, msg)
		case <-timeout:
			break
		}
	}

	// 验证至少收到一条消息
	assert.Equal(t, len(received) > 0, true)
}
