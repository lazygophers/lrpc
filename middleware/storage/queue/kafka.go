package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/json"
	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// kafkaTopic Kafka 实现 of Topic 接口
type kafkaTopic[T any] struct {
	name        string
	config      *TopicConfig
	cli         *kafka.Writer
	topic       string // Kafka topic 名称
	prefix      string
	brokers     []string
	kafkaConfig *KafkaConfig
	channels    map[string]*kafkaChannel[T]
	mu          sync.RWMutex
	closed      bool
}

// kafkaChannel Kafka 实现 of Channel 接口
type kafkaChannel[T any] struct {
	topic    *kafkaTopic[T]
	name     string
	config   *ChannelConfig
	reader   *kafka.Reader
	groupID  string
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
	once     sync.Once
	closed   bool
}

// NewKafkaTopic 创建 Kafka Topic
func NewKafkaTopic[T any](writer *kafka.Writer, name string, config *TopicConfig, prefix string, brokers []string, kafkaConfig *KafkaConfig) Topic[T] {
	if config == nil {
		config = &TopicConfig{}
	}
	config.apply()

	t := &kafkaTopic[T]{
		name:        name,
		config:      config,
		cli:         writer,
		topic:       prefix + name,
		prefix:      prefix,
		brokers:     brokers,
		kafkaConfig: kafkaConfig,
		channels:    make(map[string]*kafkaChannel[T]),
		closed:      false,
	}

	// 自动创建 topic（如果配置要求）
	if kafkaConfig != nil && kafkaConfig.AutoCreateTopics {
		t.createTopic(brokers, kafkaConfig)
	}

	return t
}

// createTopic 创建 Kafka topic
func (t *kafkaTopic[T]) createTopic(brokers []string, config *KafkaConfig) {
	conn, err := kafka.DialLeader(context.Background(), "tcp", brokers[0], t.topic, 0)
	if err != nil {
		log.Warnf("failed to connect to kafka leader for topic %s: %v", t.topic, err)
		return
	}
	defer conn.Close()

	// 获取 controller
	controller, err := conn.Controller()
	if err != nil {
		log.Warnf("failed to get kafka controller: %v", err)
		return
	}

	// 连接到 controller
	controllerConn, err := kafka.DialContext(context.Background(), "tcp", controller.Host)
	if err != nil {
		log.Warnf("failed to connect to kafka controller: %v", err)
		return
	}
	defer controllerConn.Close()

	// 创建 topic
	topicConfigs := kafka.TopicConfig{
		Topic:             t.topic,
		NumPartitions:     config.Partition,
		ReplicationFactor: config.ReplicationFactor,
	}

	err = controllerConn.CreateTopics(topicConfigs)
	if err != nil {
		// Topic 可能已存在，忽略错误
		log.Debugf("topic %s may already exist: %v", t.topic, err)
	}
}

// Pub 发布消息
func (t *kafkaTopic[T]) Pub(msg T) error {
	return t.PubMsg(NewMessage(msg))
}

// PubBatch 批量发布消息
func (t *kafkaTopic[T]) PubBatch(msgs []T) error {
	msgList := make([]*Message[T], len(msgs))
	for i, msg := range msgs {
		msgList[i] = NewMessage(msg)
	}
	return t.PubMsgBatch(msgList)
}

// PubMsg 发布完整消息
func (t *kafkaTopic[T]) PubMsg(msg *Message[T]) error {
	if t.closed {
		return fmt.Errorf("topic closed")
	}

	// 序列化消息体
	body, err := json.Marshal(msg.Body)
	if err != nil {
		log.Errorf("marshal message body failed: %v", err)
		return err
	}

	// 构建完整的消息元数据
	meta := map[string]string{
		"id":         msg.Id,
		"timestamp":  fmt.Sprintf("%d", msg.Timestamp),
		"expires_at": fmt.Sprintf("%d", msg.ExpiresAt),
		"attempts":   fmt.Sprintf("%d", msg.Attempts),
		"channel":    msg.Channel,
	}

	// 创建 Kafka 消息
	kafkaMsg := kafka.Message{
		Topic:   t.topic,
		Key:     []byte(msg.Id),
		Value:   body,
		Headers: make([]kafka.Header, 0, len(meta)),
	}

	for k, v := range meta {
		kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	// 写入消息
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = t.cli.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		log.Errorf("write kafka message failed: %v", err)
		return err
	}

	return nil
}

// PubMsgBatch 批量发布完整消息
func (t *kafkaTopic[T]) PubMsgBatch(msgs []*Message[T]) error {
	if t.closed {
		return fmt.Errorf("topic closed")
	}

	kafkaMsgs := make([]kafka.Message, len(msgs))
	for i, msg := range msgs {
		body, err := json.Marshal(msg.Body)
		if err != nil {
			log.Errorf("marshal message body failed: %v", err)
			return err
		}

		meta := map[string]string{
			"id":         msg.Id,
			"timestamp":  fmt.Sprintf("%d", msg.Timestamp),
			"expires_at": fmt.Sprintf("%d", msg.ExpiresAt),
			"attempts":   fmt.Sprintf("%d", msg.Attempts),
			"channel":    msg.Channel,
		}

		kafkaMsg := kafka.Message{
			Topic:   t.topic,
			Key:     []byte(msg.Id),
			Value:   body,
			Headers: make([]kafka.Header, 0, len(meta)),
		}

		for k, v := range meta {
			kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}

		kafkaMsgs[i] = kafkaMsg
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := t.cli.WriteMessages(ctx, kafkaMsgs...)
	if err != nil {
		log.Errorf("write kafka messages failed: %v", err)
		return err
	}

	return nil
}

// GetOrAddChannel 获取或创建 Channel
func (t *kafkaTopic[T]) GetOrAddChannel(name string, config *ChannelConfig) (Channel[T], error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, fmt.Errorf("topic closed")
	}

	if ch, ok := t.channels[name]; ok {
		return ch, nil
	}

	if config == nil {
		config = &ChannelConfig{}
	}
	config.apply()

	ch := newKafkaChannel[T](t, name, config, t.brokers, t.kafkaConfig)
	t.channels[name] = ch

	return ch, nil
}

// GetChannel 获取已存在的 Channel
func (t *kafkaTopic[T]) GetChannel(name string) (Channel[T], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return nil, fmt.Errorf("topic closed")
	}

	ch, ok := t.channels[name]
	if !ok {
		return nil, fmt.Errorf("channel not found: %s", name)
	}

	return ch, nil
}

// ChannelList 返回所有 Channel 名称
func (t *kafkaTopic[T]) ChannelList() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	list := make([]string, 0, len(t.channels))
	for name := range t.channels {
		list = append(list, name)
	}

	return list
}

// Close 关闭 Topic
func (t *kafkaTopic[T]) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	// 关闭所有 channels
	for _, ch := range t.channels {
		_ = ch.Close()
	}

	// 关闭 writer
	if t.cli != nil {
		_ = t.cli.Close()
	}

	return nil
}

// newKafkaChannel 创建 Kafka Channel
func newKafkaChannel[T any](topic *kafkaTopic[T], name string, config *ChannelConfig, brokers []string, kafkaConfig *KafkaConfig) *kafkaChannel[T] {
	// 为每个 channel 创建独立的 consumer group
	groupID := fmt.Sprintf("%s%s-%s", topic.prefix, topic.name, name)

	// 创建 reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic.topic,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.FirstOffset,
	})

	ctx, cancel := context.WithCancel(context.Background())

	ch := &kafkaChannel[T]{
		topic:   topic,
		name:    name,
		config:  config,
		reader:  reader,
		groupID: groupID,
		ctx:     ctx,
		cancel:  cancel,
		closed:  false,
	}

	return ch
}

// Name 返回 Channel 名称
func (ch *kafkaChannel[T]) Name() string {
	return ch.name
}

// Subscribe 订阅消息
func (ch *kafkaChannel[T]) Subscribe(handler Handler[T]) {
	ch.once.Do(func() {
		ch.wg.Add(1)
		go ch.consumeLoop(handler)
	})
}

// consumeLoop 消费消息循环
func (ch *kafkaChannel[T]) consumeLoop(handler Handler[T]) {
	defer ch.wg.Done()

	semaphore := make(chan struct{}, ch.config.MaxInFlight)

	for {
		select {
		case <-ch.ctx.Done():
			return
		default:
		}

		// 获取信号量
		select {
		case semaphore <- struct{}{}:
			// 继续
		case <-ch.ctx.Done():
			return
		case <-time.After(5 * time.Second):
			// 超时继续
			continue
		}

		// 读取消息
		ctx, cancel := context.WithTimeout(ch.ctx, ch.config.AckTimeout)
		kafkaMsg, err := ch.reader.FetchMessage(ctx)
		cancel()

		if err != nil {
			<-semaphore
			if ch.ctx.Err() != nil {
				return
			}
			log.Warnf("fetch kafka message failed: %v", err)
			continue
		}

		// 在 goroutine 中处理消息
		go func(km kafka.Message) {
			defer func() { <-semaphore }()

			// 解析消息
			msg := ch.parseKafkaMessage(&km)
			if msg == nil {
				// 无法解析的消息，直接确认
				_ = ch.commitMessage(&km)
				return
			}

			// 检查消息是否过期
			if msg.IsExpired() {
				log.Debugf("message expired: %s", msg.Id)
				_ = ch.commitMessage(&km)
				return
			}

			// 增加尝试次数
			msg.IncrementAttempts()

			// 处理消息
			rsp, err := handler(msg)
			if err != nil {
				log.Errorf("handle message failed: %v", err)
			}

			// 根据 ProcessRsp 决定是否确认
			if rsp.Retry && msg.Attempts < ch.config.MaxRetries {
				// 重试：不确认消息，Kafka 会重新投递
				time.Sleep(ch.config.RetryDelay)
			} else {
				// 确认消息
				_ = ch.commitMessage(&km)
			}
		}(kafkaMsg)
	}
}

// parseKafkaMessage 解析 Kafka 消息
func (ch *kafkaChannel[T]) parseKafkaMessage(km *kafka.Message) *Message[T] {
	msg := &Message[T]{
		Id:        string(km.Key),
		Timestamp: time.Now().Unix(),
		Channel:   ch.name,
		Attempts:  0,
	}

	// 从 headers 解析元数据
	for _, h := range km.Headers {
		switch h.Key {
		case "timestamp":
			fmt.Sscanf(string(h.Value), "%d", &msg.Timestamp)
		case "expires_at":
			fmt.Sscanf(string(h.Value), "%d", &msg.ExpiresAt)
		case "attempts":
			fmt.Sscanf(string(h.Value), "%d", &msg.Attempts)
		case "channel":
			msg.Channel = string(h.Value)
		}
	}

	// 解析消息体
	var body T
	if err := json.Unmarshal(km.Value, &body); err != nil {
		log.Errorf("unmarshal message body failed: %v", err)
		return nil
	}
	msg.Body = body

	return msg
}

// commitMessage 确认消息
func (ch *kafkaChannel[T]) commitMessage(km *kafka.Message) error {
	ctx, cancel := context.WithTimeout(ch.ctx, 5*time.Second)
	defer cancel()

	if err := ch.reader.CommitMessages(ctx, *km); err != nil {
		log.Errorf("commit kafka message failed: %v", err)
		return err
	}

	return nil
}

// Next 获取下一条消息（阻塞）
func (ch *kafkaChannel[T]) Next() (*Message[T], error) {
	ctx := context.Background()
	kafkaMsg, err := ch.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	msg := ch.parseKafkaMessage(&kafkaMsg)
	if msg == nil {
		_ = ch.commitMessage(&kafkaMsg)
		return nil, fmt.Errorf("parse message failed")
	}

	msg.IncrementAttempts()

	// 保存原始消息用于 Ack
	ch.mu.Lock()
	// 这里需要某种机制来存储消息引用
	// 简化实现：直接返回消息，Ack 时需要额外处理
	ch.mu.Unlock()

	return msg, nil
}

// TryNext 尝试获取下一条消息
func (ch *kafkaChannel[T]) TryNext(timeout time.Duration) (*Message[T], error) {
	ctx, cancel := context.WithTimeout(ch.ctx, timeout)
	defer cancel()

	kafkaMsg, err := ch.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	msg := ch.parseKafkaMessage(&kafkaMsg)
	if msg == nil {
		_ = ch.commitMessage(&kafkaMsg)
		return nil, fmt.Errorf("parse message failed")
	}

	msg.IncrementAttempts()

	return msg, nil
}

// Ack 确认消息
func (ch *kafkaChannel[T]) Ack(msgId string) error {
	// Kafka 的 Ack 需要原始消息，这里简化处理
	// 实际使用中应该在 Next 时保存消息引用
	// 由于 Kafka 的 offset 提交机制，我们依赖 handler 回调中的自动提交
	return nil
}

// Nack 消息处理失败
func (ch *kafkaChannel[T]) Nack(msgId string) error {
	// Kafka 中 Nack 的实现是不提交 offset，消息会被重新投递
	// 这里简化处理，依赖 handler 的返回值和自动重试机制
	return nil
}

// Depth 返回 Channel 深度
func (ch *kafkaChannel[T]) Depth() (int64, error) {
	// Kafka 不直接提供消费者 lag
	// 需要通过连接到 broker 查询
	// 这里返回 -1 表示不支持
	return -1, nil
}

// Close 关闭 Channel
func (ch *kafkaChannel[T]) Close() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return nil
	}

	ch.closed = true

	// 取消上下文
	ch.cancel()

	// 等待消费循环结束
	ch.wg.Wait()

	// 关闭 reader
	if ch.reader != nil {
		_ = ch.reader.Close()
	}

	return nil
}

// NewKafkaWriter 创建 Kafka Writer
func NewKafkaWriter(brokers []string, topic string, config *KafkaConfig) *kafka.Writer {
	if config == nil {
		config = &KafkaConfig{}
	}
	config.apply()

	// 解析压缩类型
	var compression compress.Compression
	switch config.CompressionType {
	case "gzip":
		compression = compress.Gzip
	case "snappy":
		compression = compress.Snappy
	case "lz4":
		compression = compress.Lz4
	case "zstd":
		compression = compress.Zstd
	default:
		compression = compress.None
		// 检查无效的压缩类型
		if config.CompressionType != "none" && config.CompressionType != "" {
			log.Warnf("unknown compression type: %s, using none", config.CompressionType)
		}
	}

	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		Compression:  compression,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		BatchBytes:   1048576, // 1MB
		ReadTimeout:  config.ReadBatchTimeout,
		WriteTimeout: config.WriteTimeout,
		RequiredAcks: kafka.RequiredAcks(config.RequiredAcks),
		MaxAttempts:  config.MaxAttempts,
	}
}
