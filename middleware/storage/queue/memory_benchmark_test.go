package queue

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// BenchmarkPub 测试单发布者性能
func BenchmarkPub(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
	}
}

// BenchmarkPubMultipleChannels 测试发布到多个 channel 的性能
func BenchmarkPubMultipleChannels(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	// 创建 10 个 channel
	channels := []string{"ch0", "ch1", "ch2", "ch3", "ch4", "ch5", "ch6", "ch7", "ch8", "ch9"}
	for _, name := range channels {
		topic.GetOrAddChannel(name, nil)
	}

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
	}
}

// BenchmarkPubBatch 测试批量发布性能
func BenchmarkPubBatch(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	topic.GetOrAddChannel("ch1", nil)

	// 预创建消息池
	msgs := make([]*BenchmarkMessage, 100)
	for i := range msgs {
		msgs[i] = &BenchmarkMessage{ID: i, Data: make([]byte, 128)}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = topic.PubBatch(msgs)
	}
}

// BenchmarkPubBatch1 到 BenchmarkPubBatch1000 测试不同批量大小
func BenchmarkPubBatch1(b *testing.B)   { benchmarkPubBatch(b, 1) }
func BenchmarkPubBatch10(b *testing.B)  { benchmarkPubBatch(b, 10) }
func BenchmarkPubBatch100(b *testing.B) { benchmarkPubBatch(b, 100) }
func BenchmarkPubBatch1000(b *testing.B) {
	benchmarkPubBatch(b, 1000)
}

func benchmarkPubBatch(b *testing.B, batchSize int) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	topic.GetOrAddChannel("ch1", nil)

	msgs := make([]*BenchmarkMessage, batchSize)
	for i := range msgs {
		msgs[i] = &BenchmarkMessage{ID: i, Data: make([]byte, 128)}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = topic.PubBatch(msgs)
	}
}

// BenchmarkNext 测试获取消息性能（队列有消息）
func BenchmarkNext(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	// 预填充队列
	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = ch.Next()
	}
}

// BenchmarkTryNext 测试 TryNext 性能
func BenchmarkTryNext(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 每次发布一条消息然后立即获取
		_ = topic.Pub(msg)
		_, _ = ch.TryNext(10 * time.Millisecond)
	}
}

// BenchmarkAck 测试消息确认性能
func BenchmarkAck(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	// 获取一条消息用于测试
	_ = topic.Pub(msg)
	testMsg, _ := ch.Next()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ch.Ack(testMsg.Id)
	}
}

// BenchmarkNack 测试消息拒绝性能
func BenchmarkNack(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	// 获取一条消息用于测试
	_ = topic.Pub(msg)
	testMsg, _ := ch.Next()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ch.Nack(testMsg.Id)
	}
}

// BenchmarkDepth 测试获取队列深度性能
func BenchmarkDepth(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = ch.Depth()
	}
}

// BenchmarkConcurrentPub 测试并发发布性能
func BenchmarkConcurrentPub(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = topic.Pub(msg)
		}
	})
}

// BenchmarkConcurrentPubNext 测试并发发布和消费性能
func BenchmarkConcurrentPubNext(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}
	var wg sync.WaitGroup

	b.ResetTimer()

	// 启动多个消费者 goroutine
	consumers := 4
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < b.N/consumers; j++ {
				ch.Next()
			}
		}()
	}

	// 发布消息
	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
	}

	wg.Wait()
}

// BenchmarkSubscribe 测试订阅消费性能
func BenchmarkSubscribe(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	received := make(chan struct{}, 100)
	handler := func(msg *Message[*BenchmarkMessage]) (ProcessRsp, error) {
		received <- struct{}{}
		return ProcessRsp{Retry: false}, nil
	}

	ch.Subscribe(handler)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
		<-received
	}
}

// BenchmarkMessageSize 测试不同消息大小的影响
func BenchmarkMessageSize128(b *testing.B)  { benchmarkMessageSize(b, 128) }
func BenchmarkMessageSize512(b *testing.B)  { benchmarkMessageSize(b, 512) }
func BenchmarkMessageSize1K(b *testing.B)   { benchmarkMessageSize(b, 1024) }
func BenchmarkMessageSize4K(b *testing.B)   { benchmarkMessageSize(b, 4096) }
func BenchmarkMessageSize10K(b *testing.B)  { benchmarkMessageSize(b, 10240) }

func benchmarkMessageSize(b *testing.B, size int) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, size)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
		_, _ = ch.Next()
	}
}

// BenchmarkQueueDepth 测试不同队列深度的影响
func BenchmarkQueueDepth100(b *testing.B)   { benchmarkQueueDepth(b, 100) }
func BenchmarkQueueDepth1000(b *testing.B)  { benchmarkQueueDepth(b, 1000) }
func BenchmarkQueueDepth10000(b *testing.B) { benchmarkQueueDepth(b, 10000) }

func benchmarkQueueDepth(b *testing.B, depth int) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	// 预填充队列到指定深度
	for i := 0; i < depth; i++ {
		_ = topic.Pub(msg)
	}

	b.ResetTimer()
	b.ReportAllocs()

	// 消费所有消息
	for i := 0; i < depth; i++ {
		_, _ = ch.Next()
	}
}

// BenchmarkMultipleProducers 测试多生产者性能
func BenchmarkMultipleProducers1(b *testing.B)   { benchmarkMultipleProducers(b, 1) }
func BenchmarkMultipleProducers2(b *testing.B)   { benchmarkMultipleProducers(b, 2) }
func BenchmarkMultipleProducers4(b *testing.B)   { benchmarkMultipleProducers(b, 4) }
func BenchmarkMultipleProducers8(b *testing.B)   { benchmarkMultipleProducers(b, 8) }
func BenchmarkMultipleProducers16(b *testing.B)  { benchmarkMultipleProducers(b, 16) }
func BenchmarkMultipleProducers32(b *testing.B)  { benchmarkMultipleProducers(b, 32) }

func benchmarkMultipleProducers(b *testing.B, producers int) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()

	var wg sync.WaitGroup
	msgsPerProducer := b.N / producers

	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < msgsPerProducer; j++ {
				_ = topic.Pub(msg)
			}
		}()
	}

	wg.Wait()
}

// BenchmarkMultipleConsumers 测试多消费者性能
func BenchmarkMultipleConsumers1(b *testing.B)  { benchmarkMultipleConsumers(b, 1) }
func BenchmarkMultipleConsumers2(b *testing.B)  { benchmarkMultipleConsumers(b, 2) }
func BenchmarkMultipleConsumers4(b *testing.B)  { benchmarkMultipleConsumers(b, 4) }
func BenchmarkMultipleConsumers8(b *testing.B)  { benchmarkMultipleConsumers(b, 8) }
func BenchmarkMultipleConsumers16(b *testing.B) { benchmarkMultipleConsumers(b, 16) }

func benchmarkMultipleConsumers(b *testing.B, consumers int) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	ch, _ := topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	// 预发布所有消息
	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
	}

	b.ResetTimer()

	var wg sync.WaitGroup
	msgsPerConsumer := b.N / consumers

	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < msgsPerConsumer; j++ {
				_, _ = ch.Next()
			}
		}()
	}

	wg.Wait()
}

// BenchmarkPubContention 测试锁竞争情况下的发布性能
func BenchmarkPubContention(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	// 创建多个 channel 增加锁竞争
	channels := []string{"ch0", "ch1", "ch2", "ch3", "ch4", "ch5", "ch6", "ch7", "ch8", "ch9"}
	for _, name := range channels {
		topic.GetOrAddChannel(name, nil)
	}

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = topic.Pub(msg)
		}
	})
}

// BenchmarkGetOrAddChannel 测试获取或创建 channel 性能
func BenchmarkGetOrAddChannel(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)

	// 预创建 channel
	topic.GetOrAddChannel("ch1", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = topic.GetOrAddChannel("ch1", nil)
	}
}

// BenchmarkChannelList 测试获取 channel 列表性能
func BenchmarkChannelList(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	// 创建 100 个 channel
	channelNames := []string{"ch0", "ch1", "ch2", "ch3", "ch4", "ch5", "ch6", "ch7", "ch8", "ch9"}
	for i := 0; i < 100; i++ {
		name := channelNames[i%10] + "_" + string(rune('0'+i/10))
		topic.GetOrAddChannel(name, nil)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = topic.ChannelList()
	}
}

// BenchmarkMessageCreation 测试消息创建开销
func BenchmarkMessageCreation(b *testing.B) {
	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		message := &Message[*BenchmarkMessage]{
			Id:        uuid.New().String(),
			Body:      msg,
			Timestamp: time.Now().Unix(),
			Attempts:  0,
		}
		_ = message
	}
}

// BenchmarkUUIDGeneration 测试 UUID 生成性能
func BenchmarkUUIDGeneration(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		id := uuid.New().String()
		_ = id
	}
}

// BenchmarkMemoryAllocation 测试内存分配
func BenchmarkMemoryAllocation(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
		ch, _ := topic.GetChannel("ch1")
		msg, _ := ch.Next()
		_ = ch.Ack(msg.Id)
	}
}

// BenchmarkFullCycle 测试完整的发布-消费-确认周期
func BenchmarkFullCycle(b *testing.B) {
	topic := NewMemoryTopic[*BenchmarkMessage]("bench", nil)
	topic.GetOrAddChannel("ch1", nil)

	msg := &BenchmarkMessage{ID: 1, Data: make([]byte, 128)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = topic.Pub(msg)
		ch, _ := topic.GetChannel("ch1")
		received, _ := ch.Next()
		_ = ch.Ack(received.Id)
	}
}

// BenchmarkMessage 类型
type BenchmarkMessage struct {
	ID   int
	Data []byte
}
