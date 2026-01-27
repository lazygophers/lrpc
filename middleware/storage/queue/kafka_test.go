package queue

import (
	"os"
	"testing"
	"time"

	"github.com/lazygophers/log"
	"gotest.tools/v3/assert"
)

const (
	envKafkaAddr = "KAFKA_ADDR"
)

func getKafkaAddr() string {
	addr := os.Getenv(envKafkaAddr)
	if addr == "" {
		addr = "localhost:9092"
	}
	return addr
}

func setupKafkaQueue(t *testing.T) *Queue {
	addr := getKafkaAddr()

	q := NewQueue(&Config{
		StorageType: StorageKafka,
		KafkaConfig: &KafkaConfig{
			Brokers:           []string{addr},
			TopicPrefix:       "test-queue-",
			AutoCreateTopics:  true,
			Partition:         1,
			ReplicationFactor: 1,
		},
	})

	return q
}

func TestKafkaPubSub(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct {
		Type string
		Data string
	}

	topic := NewTopic[Event](q, "test-pubsub", &TopicConfig{})
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("handler", &ChannelConfig{
		MaxInFlight: 1,
	})
	assert.NilError(t, err)
	defer ch.Close()

	// 订阅消息
	received := make(chan *Message[Event], 1)
	ch.Subscribe(func(msg *Message[Event]) (ProcessRsp, error) {
		received <- msg
		return ProcessRsp{Retry: false}, nil
	})

	// 等待消费者准备
	time.Sleep(500 * time.Millisecond)

	// 发布消息
	event := Event{Type: "test", Data: "hello"}
	err = topic.Pub(event)
	assert.NilError(t, err)

	// 等待接收消息
	select {
	case msg := <-received:
		assert.Equal(t, msg.Body.Type, "test")
		assert.Equal(t, msg.Body.Data, "hello")
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestKafkaPubBatch(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct {
		Index int
	}

	topic := NewTopic[Event](q, "test-batch", &TopicConfig{})
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("handler", &ChannelConfig{
		MaxInFlight: 5,
	})
	assert.NilError(t, err)
	defer ch.Close()

	receivedCount := 0
	receivedChan := make(chan bool, 1)

	ch.Subscribe(func(msg *Message[Event]) (ProcessRsp, error) {
		receivedCount++
		if receivedCount >= 3 {
			receivedChan <- true
		}
		return ProcessRsp{Retry: false}, nil
	})

	// 等待消费者准备
	time.Sleep(500 * time.Millisecond)

	// 批量发布消息
	events := []Event{
		{Index: 1},
		{Index: 2},
		{Index: 3},
	}
	err = topic.PubBatch(events)
	assert.NilError(t, err)

	// 等待接收所有消息
	select {
	case <-receivedChan:
		assert.Equal(t, receivedCount, 3)
	case <-time.After(10 * time.Second):
		t.Fatalf("timeout waiting for messages, received %d", receivedCount)
	}
}

func TestKafkaPubMsg(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct {
		Content string
	}

	topic := NewTopic[Event](q, "test-pubmsg", &TopicConfig{})
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("handler", &ChannelConfig{
		MaxInFlight: 1,
	})
	assert.NilError(t, err)
	defer ch.Close()

	received := make(chan *Message[Event], 1)

	ch.Subscribe(func(msg *Message[Event]) (ProcessRsp, error) {
		received <- msg
		return ProcessRsp{Retry: false}, nil
	})

	// 等待消费者准备
	time.Sleep(500 * time.Millisecond)

	// 发布带元数据的消息
	msg := NewMessage(Event{Content: "test message"})
	msg.SetExpires(1 * time.Hour)

	err = topic.PubMsg(msg)
	assert.NilError(t, err)

	// 接收并验证消息
	select {
	case receivedMsg := <-received:
		assert.Equal(t, receivedMsg.Body.Content, "test message")
		assert.Equal(t, receivedMsg.Id, msg.Id)
		assert.Assert(t, !receivedMsg.IsExpired())
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestKafkaRetry(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct {
		ShouldFail bool
	}

	topic := NewTopic[Event](q, "test-retry", &TopicConfig{
		MaxRetries: 3,
	})
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("handler", &ChannelConfig{
		MaxInFlight: 1,
		MaxRetries:  3,
	})
	assert.NilError(t, err)
	defer ch.Close()

	attempts := 0
	receivedChan := make(chan bool, 1)

	ch.Subscribe(func(msg *Message[Event]) (ProcessRsp, error) {
		attempts++
		log.Debugf("message attempt: %d", msg.Attempts)

		if msg.Attempts < 2 {
			// 前两次重试
			return ProcessRsp{Retry: true}, nil
		}

		receivedChan <- true
		return ProcessRsp{Retry: false}, nil
	})

	// 等待消费者准备
	time.Sleep(500 * time.Millisecond)

	// 发布需要重试的消息
	err = topic.Pub(Event{ShouldFail: true})
	assert.NilError(t, err)

	// 等待重试完成
	select {
	case <-receivedChan:
		assert.Assert(t, attempts >= 2)
	case <-time.After(15 * time.Second):
		t.Fatalf("timeout waiting for retry, attempts: %d", attempts)
	}
}

func TestKafkaMessageExpiration(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct {
		Data string
	}

	topic := NewTopic[Event](q, "test-expiration", &TopicConfig{})
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("handler", &ChannelConfig{
		MaxInFlight: 1,
	})
	assert.NilError(t, err)
	defer ch.Close()

	received := make(chan *Message[Event], 1)

	ch.Subscribe(func(msg *Message[Event]) (ProcessRsp, error) {
		received <- msg
		return ProcessRsp{Retry: false}, nil
	})

	// 等待消费者准备
	time.Sleep(500 * time.Millisecond)

	// 发布已过期的消息
	msg := NewMessage(Event{Data: "expired"})
	msg.SetExpiresAt(time.Now().Unix() - 100) // 设置为过去的时间

	err = topic.PubMsg(msg)
	assert.NilError(t, err)

	// 消息应该被过滤掉，不应该收到
	select {
	case <-received:
		t.Fatal("should not receive expired message")
	case <-time.After(2 * time.Second):
		// 预期超时，消息被过滤
	}
}

func TestKafkaMultipleChannels(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct {
		Content string
	}

	topic := NewTopic[Event](q, "test-multichannel", &TopicConfig{})
	defer topic.Close()

	ch1, err := topic.GetOrAddChannel("handler1", &ChannelConfig{MaxInFlight: 1})
	assert.NilError(t, err)
	defer ch1.Close()

	ch2, err := topic.GetOrAddChannel("handler2", &ChannelConfig{MaxInFlight: 1})
	assert.NilError(t, err)
	defer ch2.Close()

	received1 := make(chan bool, 1)
	received2 := make(chan bool, 1)

	ch1.Subscribe(func(msg *Message[Event]) (ProcessRsp, error) {
		received1 <- true
		return ProcessRsp{Retry: false}, nil
	})

	ch2.Subscribe(func(msg *Message[Event]) (ProcessRsp, error) {
		received2 <- true
		return ProcessRsp{Retry: false}, nil
	})

	// 等待消费者准备
	time.Sleep(500 * time.Millisecond)

	// 发布消息，两个 channel 都应该收到
	err = topic.Pub(Event{Content: "test"})
	assert.NilError(t, err)

	// 两个 channel 都应该接收到消息
	<-received1
	<-received2
}

func TestKafkaMaxInFlight(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct {
		Index int
	}

	topic := NewTopic[Event](q, "test-maxinflight", &TopicConfig{})
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("handler", &ChannelConfig{
		MaxInFlight: 2, // 限制并发数为 2
	})
	assert.NilError(t, err)
	defer ch.Close()

	activeCount := 0
	maxActiveCount := 0
	doneChan := make(chan bool, 1)

	ch.Subscribe(func(msg *Message[Event]) (ProcessRsp, error) {
		activeCount++
		if activeCount > maxActiveCount {
			maxActiveCount = activeCount
		}

		// 模拟处理耗时
		time.Sleep(500 * time.Millisecond)

		activeCount--

		if msg.Body.Index >= 5 {
			doneChan <- true
		}

		return ProcessRsp{Retry: false}, nil
	})

	// 等待消费者准备
	time.Sleep(500 * time.Millisecond)

	// 批量发布消息
	for i := 1; i <= 5; i++ {
		err = topic.Pub(Event{Index: i})
		assert.NilError(t, err)
	}

	// 等待所有消息处理完成
	select {
	case <-doneChan:
		// 验证最大并发数不超过 MaxInFlight
		assert.Assert(t, maxActiveCount <= 2)
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for messages")
	}
}

func TestKafkaTopicChannelList(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct{}

	topic := NewTopic[Event](q, "test-channellist", &TopicConfig{})
	defer topic.Close()

	// 添加多个 channel
	_, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	assert.NilError(t, err)

	_, err = topic.GetOrAddChannel("ch2", &ChannelConfig{})
	assert.NilError(t, err)

	_, err = topic.GetOrAddChannel("ch3", &ChannelConfig{})
	assert.NilError(t, err)

	// 获取 channel 列表
	list := topic.ChannelList()
	assert.Equal(t, len(list), 3)

	// 验证 channel 名称
	channels := make(map[string]bool)
	for _, name := range list {
		channels[name] = true
	}

	assert.Assert(t, channels["ch1"])
	assert.Assert(t, channels["ch2"])
	assert.Assert(t, channels["ch3"])
}

func TestKafkaGetChannel(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct{}

	topic := NewTopic[Event](q, "test-getchannel", &TopicConfig{})
	defer topic.Close()

	// 添加 channel
	ch1, err := topic.GetOrAddChannel("handler1", &ChannelConfig{})
	assert.NilError(t, err)

	// 获取已存在的 channel
	ch2, err := topic.GetChannel("handler1")
	assert.NilError(t, err)

	// 应该是同一个 channel
	assert.Equal(t, ch1.Name(), ch2.Name())

	// 获取不存在的 channel
	_, err = topic.GetChannel("notexist")
	assert.Assert(t, err != nil, "expected error for non-existent channel")
}

func TestKafkaTopicClose(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct{}

	topic := NewTopic[Event](q, "test-close", &TopicConfig{})

	// 添加 channel
	_, err := topic.GetOrAddChannel("handler", &ChannelConfig{})
	assert.NilError(t, err)

	// 关闭 topic
	err = topic.Close()
	assert.NilError(t, err)

	// 关闭后不能添加 channel
	_, err = topic.GetOrAddChannel("handler2", &ChannelConfig{})
	assert.Assert(t, err != nil, "expected error when adding channel to closed topic")

	// 关闭后不能发布消息
	err = topic.Pub(Event{})
	assert.Assert(t, err != nil, "expected error when publishing to closed topic")
}

func TestKafkaChannelClose(t *testing.T) {
	q := setupKafkaQueue(t)
	defer q.Close()

	type Event struct{}

	topic := NewTopic[Event](q, "test-channel-close", &TopicConfig{})
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("handler", &ChannelConfig{})
	assert.NilError(t, err)

	// 关闭 channel
	err = ch.Close()
	assert.NilError(t, err)

	// 再次关闭应该安全
	err = ch.Close()
	assert.NilError(t, err)
}

func TestNewKafkaWriter(t *testing.T) {
	addr := getKafkaAddr()

	config := &KafkaConfig{
		Brokers:          []string{addr},
		CompressionType:  "gzip",
		RequiredAcks:     1,
		ReadBatchTimeout: 10 * time.Second,
		WriteTimeout:     10 * time.Second,
		MaxAttempts:      5,
	}

	writer := NewKafkaWriter(config.Brokers, "test-topic", config)
	assert.Assert(t, writer != nil)

	_ = writer.Close()
}

func BenchmarkKafkaPub(b *testing.B) {
	addr := getKafkaAddr()

	q := NewQueue(&Config{
		StorageType: StorageKafka,
		KafkaConfig: &KafkaConfig{
			Brokers:           []string{addr},
			TopicPrefix:       "bench-queue-",
			AutoCreateTopics:  true,
			Partition:         1,
			ReplicationFactor: 1,
		},
	})
	defer q.Close()

	type Event struct {
		Data string
	}

	topic := NewTopic[Event](q, "bench-pub", &TopicConfig{})
	defer topic.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := topic.Pub(Event{Data: "benchmark"})
		if err != nil {
			b.Fatalf("publish failed: %v", err)
		}
	}

	// 等待消息发送完成
	time.Sleep(1 * time.Second)
}

func BenchmarkKafkaPubBatch(b *testing.B) {
	addr := getKafkaAddr()

	q := NewQueue(&Config{
		StorageType: StorageKafka,
		KafkaConfig: &KafkaConfig{
			Brokers:           []string{addr},
			TopicPrefix:       "bench-queue-",
			AutoCreateTopics:  true,
			Partition:         1,
			ReplicationFactor: 1,
		},
	})
	defer q.Close()

	type Event struct {
		Data string
	}

	topic := NewTopic[Event](q, "bench-pub-batch", &TopicConfig{})
	defer topic.Close()

	batchSize := 100

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i += batchSize {
		events := make([]Event, batchSize)
		for j := range events {
			events[j] = Event{Data: "benchmark"}
		}

		err := topic.PubBatch(events)
		if err != nil {
			b.Fatalf("publish batch failed: %v", err)
		}
	}

	// 等待消息发送完成
	time.Sleep(1 * time.Second)
}
