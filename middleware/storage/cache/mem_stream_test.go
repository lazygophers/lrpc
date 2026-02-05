package cache

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCacheMem_Stream_Basic(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	streamName := "test_stream"

	// 添加消息
	id1, err := cache.XAdd(streamName, map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	})
	assert.NilError(t, err)
	assert.Equal(t, len(id1) > 0, true)

	_, err = cache.XAdd(streamName, map[string]interface{}{
		"field3": "value3",
	})
	assert.NilError(t, err)

	// 获取流长度
	length, err := cache.XLen(streamName)
	assert.NilError(t, err)
	assert.Equal(t, length, int64(2))

	// 范围查询
	messages, err := cache.XRange(streamName, "-", "+")
	assert.NilError(t, err)
	assert.Equal(t, len(messages), 2)

	// 删除消息
	deleted, err := cache.XDel(streamName, id1)
	assert.NilError(t, err)
	assert.Equal(t, deleted, int64(1))

	// 再次获取长度
	length, err = cache.XLen(streamName)
	assert.NilError(t, err)
	assert.Equal(t, length, int64(1))
}

func TestCacheMem_Stream_ConsumerGroup(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	streamName := "test_stream_group"
	groupName := "test_group"

	// 添加消息
	_, err := cache.XAdd(streamName, map[string]interface{}{
		"data": "message1",
	})
	assert.NilError(t, err)

	// 创建消费者组
	err = cache.XGroupCreate(streamName, groupName, "0")
	assert.NilError(t, err)

	// 设置消费者组 ID
	err = cache.XGroupSetID(streamName, groupName, "0")
	assert.NilError(t, err)

	// 查询待处理消息（应该为 0）
	pending, err := cache.XPending(streamName, groupName)
	assert.NilError(t, err)
	assert.Equal(t, pending, int64(0))

	// 销毁消费者组
	err = cache.XGroupDestroy(streamName, groupName)
	assert.NilError(t, err)

	// 再次查询应该返回 0（组不存在）
	pending, err = cache.XPending(streamName, groupName)
	assert.NilError(t, err)
	assert.Equal(t, pending, int64(0))
}

func TestCacheMem_Stream_Trim(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	streamName := "test_stream_trim"

	// 添加多条消息
	for i := 0; i < 10; i++ {
		_, err := cache.XAdd(streamName, map[string]interface{}{
			"index": i,
		})
		assert.NilError(t, err)
	}

	// 裁剪到最大长度 5
	trimmed, err := cache.XTrim(streamName, 5)
	assert.NilError(t, err)
	assert.Equal(t, trimmed, int64(5))

	// 验证长度
	length, err := cache.XLen(streamName)
	assert.NilError(t, err)
	assert.Equal(t, length, int64(5))
}

func TestCacheMem_Stream_Range(t *testing.T) {
	cache := NewMem()
	defer cache.Close()

	streamName := "test_stream_range"

	// 添加消息
	id1, _ := cache.XAdd(streamName, map[string]interface{}{"value": "1"})
	id2, _ := cache.XAdd(streamName, map[string]interface{}{"value": "2"})
	id3, _ := cache.XAdd(streamName, map[string]interface{}{"value": "3"})

	// 正向范围查询
	messages, err := cache.XRange(streamName, "-", "+")
	assert.NilError(t, err)
	assert.Equal(t, len(messages), 3)

	// 反向范围查询
	messages, err = cache.XRevRange(streamName, "+", "-")
	assert.NilError(t, err)
	assert.Equal(t, len(messages), 3)

	// 带数量限制的范围查询
	messages, err = cache.XRange(streamName, "-", "+", 2)
	assert.NilError(t, err)
	assert.Equal(t, len(messages), 2)

	// 删除测试消息
	cache.XDel(streamName, id1, id2, id3)
}

func TestCacheMem_MockMode(t *testing.T) {
	// 测试 mock 模式下返回 mem 缓存
	config := &Config{
		Type: Redis,
		Mock: true,
	}

	cache, err := New(config)
	assert.NilError(t, err)
	assert.Assert(t, cache != nil)

	// 验证是 mem 缓存（通过测试基本操作）
	err = cache.Set("test_key", "test_value")
	assert.NilError(t, err)

	value, err := cache.Get("test_key")
	assert.NilError(t, err)
	assert.Equal(t, value, "test_value")

	cache.Close()
}
