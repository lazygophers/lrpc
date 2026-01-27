package queue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lazygophers/log"
	"github.com/redis/go-redis/v9"
)

// getRedisClient 获取测试用 Redis 客户端
// 通过环境变量 REDIS_ADDR 指定地址，默认 localhost:6379
func getRedisClient(t *testing.T) *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0, // 使用 DB 0 进行测试
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// 测试连接
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis 连接失败 (跳过测试): %v", err)
		return nil
	}

	return client
}

// cleanTestStream 清理测试用的 Stream
func cleanTestStream(cli *redis.Client, stream string) {
	ctx := context.Background()
	cli.Del(ctx, stream)
}

func TestRedisTopic_NewRedisTopic(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	tests := []struct {
		name   string
		config *TopicConfig
		prefix string
	}{
		{
			name:   "默认配置",
			config: &TopicConfig{},
			prefix: "test:",
		},
		{
			name: "自定义配置",
			config: &TopicConfig{
				MaxRetries:  10,
				MessageTTL:  time.Hour,
				MaxBodySize: 1024,
			},
			prefix: "custom:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic := NewRedisTopic[TestMsg](cli, "test-topic", tt.config, tt.prefix)
			if topic == nil {
				t.Fatal("NewRedisTopic() = nil, want non-nil")
			}

			err := topic.Close()
			if err != nil {
				t.Errorf("Topic.Close() error = %v", err)
			}
		})
	}
}

func TestRedisTopic_Pub(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "pub-test", &TopicConfig{}, "test:")
	defer topic.Close()

	// 创建 Channel
	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	// 清理测试数据
	cleanTestStream(cli, "test:pub-test:ch1")
	defer cleanTestStream(cli, "test:pub-test:ch1")

	// 发布消息
	msg := TestMsg{Content: "hello"}
	err = topic.Pub(msg)
	if err != nil {
		t.Errorf("Pub() error = %v", err)
	}

	// 验证消息已发送
	ctx := context.Background()
	length, err := cli.XLen(ctx, "test:pub-test:ch1").Result()
	if err != nil {
		t.Errorf("XLen() error = %v", err)
	}
	if length != 1 {
		t.Errorf("消息数量 = %d, want 1", length)
	}
}

func TestRedisTopic_PubBatch(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "pub-batch-test", &TopicConfig{}, "test:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "test:pub-batch-test:ch1")
	defer cleanTestStream(cli, "test:pub-batch-test:ch1")

	// 批量发布
	msgs := []TestMsg{
		{Content: "msg1"},
		{Content: "msg2"},
		{Content: "msg3"},
	}
	err = topic.PubBatch(msgs)
	if err != nil {
		t.Errorf("PubBatch() error = %v", err)
	}

	// 验证
	ctx := context.Background()
	length, err := cli.XLen(ctx, "test:pub-batch-test:ch1").Result()
	if err != nil {
		t.Errorf("XLen() error = %v", err)
	}
	if length != 3 {
		t.Errorf("消息数量 = %d, want 3", length)
	}
}

func TestRedisTopic_GetOrAddChannel(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct{}

	topic := NewRedisTopic[TestMsg](cli, "channel-test", &TopicConfig{}, "test:")
	defer topic.Close()

	cleanTestStream(cli, "test:channel-test:ch1")
	defer cleanTestStream(cli, "test:channel-test:ch1")

	// 创建新 Channel
	ch1, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}

	// 获取已存在的 Channel
	ch2, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Errorf("第二次 GetOrAddChannel() error = %v", err)
	}
	if ch1 != ch2 {
		t.Error("应该返回同一个 Channel 实例")
	}

	// GetChannel
	ch3, err := topic.GetChannel("ch1")
	if err != nil {
		t.Errorf("GetChannel() error = %v", err)
	}
	if ch1 != ch3 {
		t.Error("GetChannel 应该返回同一个实例")
	}

	// ChannelList
	list := topic.ChannelList()
	if len(list) != 1 {
		t.Errorf("ChannelList() length = %d, want 1", len(list))
	}
	if list[0] != "ch1" {
		t.Errorf("ChannelList()[0] = %s, want ch1", list[0])
	}
}

func TestRedisChannel_Subscribe(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "subscribe-test", &TopicConfig{}, "test:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{
		MaxInFlight: 2,
	})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "test:subscribe-test:ch1")
	defer cleanTestStream(cli, "test:subscribe-test:ch1")

	// 订阅消息
	received := make(chan *Message[TestMsg], 10)
	handler := func(msg *Message[TestMsg]) (ProcessRsp, error) {
		received <- msg
		return ProcessRsp{Retry: false}, nil
	}

	ch.Subscribe(handler)

	// 发布消息
	msg := TestMsg{Content: "test message"}
	err = topic.Pub(msg)
	if err != nil {
		t.Errorf("Pub() error = %v", err)
	}

	// 等待接收消息
	select {
	case got := <-received:
		if got.Body.Content != "test message" {
			t.Errorf("接收到消息内容 = %s, want 'test message'", got.Body.Content)
		}
	case <-time.After(5 * time.Second):
		t.Error("未在超时时间内接收到消息")
	}
}

func TestRedisChannel_Next(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "next-test", &TopicConfig{}, "test:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "test:next-test:ch1")
	defer cleanTestStream(cli, "test:next-test:ch1")

	// 先发布一条消息
	msg := TestMsg{Content: "next test"}
	err = topic.Pub(msg)
	if err != nil {
		t.Errorf("Pub() error = %v", err)
	}

	// Next 获取消息
	got, err := ch.Next()
	if err != nil {
		t.Errorf("Next() error = %v", err)
	}
	if got.Body.Content != "next test" {
		t.Errorf("Next() 内容 = %s, want 'next test'", got.Body.Content)
	}

	// Ack 消息
	err = ch.Ack(got.Id)
	if err != nil {
		t.Errorf("Ack() error = %v", err)
	}
}

func TestRedisChannel_TryNext(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "trynext-test", &TopicConfig{}, "test:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "test:trynext-test:ch1")
	defer cleanTestStream(cli, "test:trynext-test:ch1")

	// TryNext 无消息（超时为 0）
	_, err = ch.TryNext(0)
	if err == nil {
		t.Error("TryNext(0) 应该返回错误")
	}

	// 发布消息
	msg := TestMsg{Content: "trynext test"}
	err = topic.Pub(msg)
	if err != nil {
		t.Errorf("Pub() error = %v", err)
	}

	// TryNext 有消息
	got, err := ch.TryNext(5 * time.Second)
	if err != nil {
		t.Errorf("TryNext() error = %v", err)
	}
	if got.Body.Content != "trynext test" {
		t.Errorf("TryNext() 内容 = %s, want 'trynext test'", got.Body.Content)
	}

	err = ch.Ack(got.Id)
	if err != nil {
		t.Errorf("Ack() error = %v", err)
	}
}

func TestRedisChannel_Ack(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "ack-test", &TopicConfig{}, "test:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "test:ack-test:ch1")
	defer cleanTestStream(cli, "test:ack-test:ch1")

	// 发布并获取消息
	msg := TestMsg{Content: "ack test"}
	err = topic.Pub(msg)
	if err != nil {
		t.Errorf("Pub() error = %v", err)
	}

	got, err := ch.Next()
	if err != nil {
		t.Errorf("Next() error = %v", err)
	}

	// Ack 消息
	err = ch.Ack(got.Id)
	if err != nil {
		t.Errorf("Ack() error = %v", err)
	}

	// 检查 pending 数量应该为 0
	ctx := context.Background()
	pending, err := cli.XPending(ctx, "test:ack-test:ch1", "ack-test:ch1").Result()
	if err != nil {
		t.Errorf("XPending() error = %v", err)
	}
	if pending.Count != 0 {
		t.Errorf("Pending 数量 = %d, want 0", pending.Count)
	}
}

func TestRedisChannel_Nack(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "nack-test", &TopicConfig{}, "test:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "test:nack-test:ch1")
	defer cleanTestStream(cli, "test:nack-test:ch1")

	// 发布并获取消息
	msg := TestMsg{Content: "nack test"}
	err = topic.Pub(msg)
	if err != nil {
		t.Errorf("Pub() error = %v", err)
	}

	got, err := ch.Next()
	if err != nil {
		t.Errorf("Next() error = %v", err)
	}

	// Nack 消息（应该重新入队）
	err = ch.Nack(got.Id)
	if err != nil {
		t.Errorf("Nack() error = %v", err)
	}

	// 等待消息重新入队
	time.Sleep(100 * time.Millisecond)

	// 再次获取消息，应该是重新入队的消息
	got2, err := ch.Next()
	if err != nil {
		t.Errorf("Next() error = %v", err)
	}
	if got2.Body.Content != "nack test" {
		t.Errorf("Next() 内容 = %s, want 'nack test'", got2.Body.Content)
	}
	if got2.Attempts != 1 {
		t.Errorf("Attempts = %d, want 1", got2.Attempts)
	}

	// 清理
	err = ch.Ack(got2.Id)
	if err != nil {
		t.Errorf("Ack() error = %v", err)
	}
}

func TestRedisChannel_Depth(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "depth-test", &TopicConfig{}, "test:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "test:depth-test:ch1")
	defer cleanTestStream(cli, "test:depth-test:ch1")

	// 初始深度
	depth, err := ch.Depth()
	if err != nil {
		t.Errorf("Depth() error = %v", err)
	}
	if depth != 0 {
		t.Errorf("初始 Depth = %d, want 0", depth)
	}

	// 发布消息
	for i := 0; i < 5; i++ {
		msg := TestMsg{Content: "test"}
		err = topic.Pub(msg)
		if err != nil {
			t.Errorf("Pub() error = %v", err)
		}
	}

	// 深度应该增加
	depth, err = ch.Depth()
	if err != nil {
		t.Errorf("Depth() error = %v", err)
	}
	if depth != 5 {
		t.Errorf("Depth = %d, want 5", depth)
	}

	// 消费一条消息
	got, err := ch.Next()
	if err != nil {
		t.Errorf("Next() error = %v", err)
	}
	_ = got

	// 深度不变（消息在 pending 中）
	depth, err = ch.Depth()
	if err != nil {
		t.Errorf("Depth() error = %v", err)
	}
	if depth != 5 {
		t.Errorf("Depth = %d, want 5", depth)
	}

	// Ack 后深度应该减少
	err = ch.Ack(got.Id)
	if err != nil {
		t.Errorf("Ack() error = %v", err)
	}

	// 给 Redis 一些时间处理
	time.Sleep(100 * time.Millisecond)

	depth, err = ch.Depth()
	if err != nil {
		t.Errorf("Depth() error = %v", err)
	}
	// 深度应该为 4（一条已确认）
	if depth != 4 {
		t.Errorf("Depth = %d, want 4", depth)
	}
}

func TestRedisTopic_MultipleChannels(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "multi-test", &TopicConfig{}, "test:")
	defer topic.Close()

	ch1, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel(ch1) error = %v", err)
	}
	defer ch1.Close()

	ch2, err := topic.GetOrAddChannel("ch2", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel(ch2) error = %v", err)
	}
	defer ch2.Close()

	cleanTestStream(cli, "test:multi-test:ch1")
	defer cleanTestStream(cli, "test:multi-test:ch1")
	cleanTestStream(cli, "test:multi-test:ch2")
	defer cleanTestStream(cli, "test:multi-test:ch2")

	// 发布消息
	msg := TestMsg{Content: "broadcast"}
	err = topic.Pub(msg)
	if err != nil {
		t.Errorf("Pub() error = %v", err)
	}

	// 两个 Channel 都应该收到消息
	ctx := context.Background()
	len1, err := cli.XLen(ctx, "test:multi-test:ch1").Result()
	if err != nil {
		t.Errorf("XLen(ch1) error = %v", err)
	}
	if len1 != 1 {
		t.Errorf("ch1 消息数 = %d, want 1", len1)
	}

	len2, err := cli.XLen(ctx, "test:multi-test:ch2").Result()
	if err != nil {
		t.Errorf("XLen(ch2) error = %v", err)
	}
	if len2 != 1 {
		t.Errorf("ch2 消息数 = %d, want 1", len2)
	}
}

func TestRedisTopic_Close(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct{}

	topic := NewRedisTopic[TestMsg](cli, "close-test", &TopicConfig{}, "test:")
	defer func() {
		// 即使手动关闭，defer 也应该安全
		_ = topic.Close()
	}()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		t.Fatalf("GetOrAddChannel() error = %v", err)
	}

	// 关闭 Topic
	err = topic.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 关闭后应该无法创建新 Channel
	_, err = topic.GetOrAddChannel("ch2", &ChannelConfig{})
	if err == nil {
		t.Error("关闭后应该无法创建新 Channel")
	}

	// Channel 也应该被关闭
	err = ch.Close()
	if err != nil {
		t.Errorf("Channel.Close() error = %v", err)
	}

	// 再次关闭 Topic 应该安全
	err = topic.Close()
	if err != nil {
		t.Errorf("再次 Close() 应该安全，但返回 error = %v", err)
	}
}

func TestNewTopic_WithRedis(t *testing.T) {
	cli := getRedisClient(t)
	if cli == nil {
		return
	}
	defer cli.Close()

	type TestMsg struct{}

	// 使用 Redis 创建 Queue
	config := &Config{
		StorageType: StorageRedis,
		RedisConfig: &RedisConfig{
			KeyPrefix: "test:",
		},
		RedisClient: cli,
	}

	queue := NewQueue(config)
	defer func() {
		err := queue.Close()
		if err != nil {
			log.Errorf("Queue.Close() error = %v", err)
		}
	}()

	topic := NewTopic[TestMsg](queue, "factory-test", &TopicConfig{})
	if topic == nil {
		t.Fatal("NewTopic() = nil, want non-nil")
	}
	defer topic.Close()

	// 验证 Topic 类型
	_, ok := topic.(*redisTopic[TestMsg])
	if !ok {
		t.Error("NewTopic() 应该返回 *redisTopic 类型")
	}
}

// TestRedisConfig_Defaults 测试 Redis 配置默认值
func TestRedisConfig_Defaults(t *testing.T) {
	config := &RedisConfig{}
	config.apply()

	if config.Addr != "localhost:6379" {
		t.Errorf("默认 Addr = %s, want localhost:6379", config.Addr)
	}
	if config.DB != 0 {
		t.Errorf("默认 DB = %d, want 0", config.DB)
	}
	if config.KeyPrefix != "lrpc:queue:" {
		t.Errorf("默认 KeyPrefix = %s, want lrpc:queue:", config.KeyPrefix)
	}
	if config.PoolSize != 10 {
		t.Errorf("默认 PoolSize = %d, want 10", config.PoolSize)
	}
	if config.MinIdleConns != 5 {
		t.Errorf("默认 MinIdleConns = %d, want 5", config.MinIdleConns)
	}
	if config.MaxRetries != 3 {
		t.Errorf("默认 MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.DialTimeout != 5*time.Second {
		t.Errorf("默认 DialTimeout = %v, want 5s", config.DialTimeout)
	}
}

// BenchmarkRedisTopic_Pub 基准测试：发布消息
func BenchmarkRedisTopic_Pub(b *testing.B) {
	cli := getRedisClient(&testing.T{})
	if cli == nil {
		b.Skip("Redis 不可用")
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "bench-pub", &TopicConfig{}, "bench:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		b.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "bench:bench-pub:ch1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := TestMsg{Content: "benchmark"}
		err = topic.Pub(msg)
		if err != nil {
			b.Errorf("Pub() error = %v", err)
		}
	}
}

// BenchmarkRedisChannel_Next 基准测试：获取消息
func BenchmarkRedisChannel_Next(b *testing.B) {
	cli := getRedisClient(&testing.T{})
	if cli == nil {
		b.Skip("Redis 不可用")
		return
	}
	defer cli.Close()

	type TestMsg struct {
		Content string
	}

	topic := NewRedisTopic[TestMsg](cli, "bench-next", &TopicConfig{}, "bench:")
	defer topic.Close()

	ch, err := topic.GetOrAddChannel("ch1", &ChannelConfig{})
	if err != nil {
		b.Fatalf("GetOrAddChannel() error = %v", err)
	}
	defer ch.Close()

	cleanTestStream(cli, "bench:bench-next:ch1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 先发布一条消息
		msg := TestMsg{Content: "benchmark"}
		err = topic.Pub(msg)
		if err != nil {
			b.Errorf("Pub() error = %v", err)
		}

		// 获取消息
		_, err := ch.Next()
		if err != nil {
			b.Errorf("Next() error = %v", err)
		}
	}
}
