package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/lazygophers/utils/routine"
	"github.com/redis/go-redis/v9"
)

// redisTopic Redis Stream Topic 实现
type redisTopic[T any] struct {
	name   string
	config *TopicConfig
	cli    *redis.Client
	prefix string

	channels map[string]*redisChannel[T]
	mu       sync.RWMutex
	closed   bool
}

// redisChannel Redis Stream Channel 实现
type redisChannel[T any] struct {
	name    string
	topic   *redisTopic[T]
	config  *ChannelConfig
	stream  string // 完整的 Redis Stream 键名
	group   string // 消费者组名称

	mu     sync.Mutex
	closed bool
}

// NewRedisTopic 创建 Redis Topic
func NewRedisTopic[T any](cli *redis.Client, name string, config *TopicConfig, prefix string) Topic[T] {
	if config == nil {
		config = &TopicConfig{}
	}
	config.apply()

	t := &redisTopic[T]{
		name:     name,
		config:   config,
		cli:      cli,
		prefix:   prefix,
		channels: make(map[string]*redisChannel[T]),
	}

	return t
}

// Pub 发布消息到 Topic
func (t *redisTopic[T]) Pub(msg T) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return xerror.New(ErrTopicClosed, "topic closed")
	}

	message := NewMessage(msg)

	// 发布到所有 Channel
	for _, ch := range t.channels {
		err := ch.pubMsg(message)
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}

	return nil
}

// PubBatch 批量发布消息
func (t *redisTopic[T]) PubBatch(msgs []T) error {
	for _, msg := range msgs {
		err := t.Pub(msg)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	return nil
}

// PubMsg 发布完整消息
func (t *redisTopic[T]) PubMsg(msg *Message[T]) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return xerror.New(ErrTopicClosed, "topic closed")
	}

	if msg.Id == "" {
		msg.Id = GenerateMessageID()
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}

	// 发布到所有 Channel
	for _, ch := range t.channels {
		err := ch.pubMsg(msg)
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}

	return nil
}

// PubMsgBatch 批量发布完整消息
func (t *redisTopic[T]) PubMsgBatch(msgs []*Message[T]) error {
	for _, msg := range msgs {
		err := t.PubMsg(msg)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	return nil
}

// GetOrAddChannel 获取或创建 Channel
func (t *redisTopic[T]) GetOrAddChannel(name string, config *ChannelConfig) (Channel[T], error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, xerror.New(ErrTopicClosed, "topic closed")
	}

	if ch, exists := t.channels[name]; exists {
		return ch, nil
	}

	if config == nil {
		config = &ChannelConfig{}
	}
	config.apply()

	stream := fmt.Sprintf("%s%s:%s", t.prefix, t.name, name)
	group := fmt.Sprintf("%s:%s", t.name, name)

	ch := &redisChannel[T]{
		name:    name,
		topic:   t,
		config:  config,
		stream:  stream,
		group:   group,
		closed:  false,
	}

	// 创建消费者组（如果不存在）
	ctx := context.Background()
	err := t.cli.XGroupCreate(ctx, stream, group, "0").Err()
	if err != nil && err != redis.Nil {
		// 如果是 BUSYGROUP 错误，表示组已存在，忽略
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			log.Errorf("err:%v", err)
			return nil, err
		}
	}

	t.channels[name] = ch

	return ch, nil
}

// GetChannel 获取已存在的 Channel
func (t *redisTopic[T]) GetChannel(name string) (Channel[T], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ch, exists := t.channels[name]
	if !exists {
		return nil, xerror.New(ErrChannelNotFound, "channel not found")
	}

	return ch, nil
}

// ChannelList 返回所有 Channel 名称
func (t *redisTopic[T]) ChannelList() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	names := make([]string, 0, len(t.channels))
	for name := range t.channels {
		names = append(names, name)
	}

	return names
}

// Close 关闭 Topic
func (t *redisTopic[T]) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	for _, ch := range t.channels {
		err := ch.close()
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}

	return nil
}

// pubMsg 发布消息到 Channel
func (ch *redisChannel[T]) pubMsg(msg *Message[T]) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return xerror.New(ErrChannelClosed, "channel closed")
	}

	// 序列化消息体
	body, err := json.Marshal(msg.Body)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// 构造 Stream 字段
	values := map[string]interface{}{
		"id":        msg.Id,
		"body":      body,
		"timestamp": msg.Timestamp,
		"channel":   ch.name,
	}

	if msg.ExpiresAt > 0 {
		values["expires_at"] = msg.ExpiresAt
	}
	if msg.Attempts > 0 {
		values["attempts"] = msg.Attempts
	}

	// 添加到 Stream
	ctx := context.Background()
	_, err = ch.topic.cli.XAdd(ctx, &redis.XAddArgs{
		Stream: ch.stream,
		MaxLen: ch.topic.config.MaxMsgSize,
		Approx: true,
		ID:     msg.Id,
		Values: values,
	}).Result()

	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

// Name 返回 Channel 名称
func (ch *redisChannel[T]) Name() string {
	return ch.name
}

// Subscribe 订阅消息，使用 worker pool 并发处理
func (ch *redisChannel[T]) Subscribe(handler Handler[T]) {
	routine.GoWithMustSuccess(func() error {
		// 使用 semaphore 限制并发数
		sem := make(chan struct{}, ch.config.MaxInFlight)

		for {
			msg, err := ch.Next()
			if err != nil {
				break
			}

			// 获取 semaphore，如果已满则阻塞
			sem <- struct{}{}

			// 启动 worker 处理消息
			routine.GoWithMustSuccess(func() error {
				defer func() { <-sem }() // 处理完成后释放 semaphore

				rsp, err := Handle(handler, msg)
				if err != nil {
					log.Errorf("err:%v", err)
				}

				if rsp.Retry {
					err = ch.Nack(msg.Id)
					if err != nil {
						log.Errorf("err:%v", err)
					}
				} else {
					err = ch.Ack(msg.Id)
					if err != nil {
						log.Errorf("err:%v", err)
					}
				}

				return nil
			})
		}

		// 等待所有 worker 完成
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}

		return nil
	})
}

// Next 获取下一条消息（阻塞）
func (ch *redisChannel[T]) Next() (*Message[T], error) {
	ctx := context.Background()
	consumer := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	for {
		ch.mu.Lock()
		if ch.closed {
			ch.mu.Unlock()
			return nil, xerror.New(ErrChannelClosed, "channel closed")
		}
		ch.mu.Unlock()

		// 使用 XREADGROUP 阻塞读取消息
		// BLOCK 60000 表示阻塞 60 秒
		// ">" 表示只读取新消息
		result, err := ch.topic.cli.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    ch.group,
			Consumer: consumer,
			Streams:  []string{ch.stream, ">"},
			Count:    1,
			Block:    60 * time.Second,
		}).Result()

		if err != nil && err != redis.Nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		// 检查是否关闭
		ch.mu.Lock()
		if ch.closed {
			ch.mu.Unlock()
			return nil, xerror.New(ErrChannelClosed, "channel closed")
		}
		ch.mu.Unlock()

		// 解析消息
		for _, stream := range result {
			for _, xMsg := range stream.Messages {
				msg, err := ch.parseMessage(&xMsg)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				// 检查消息是否过期
				if msg.IsExpired() {
					// 确认过期消息
					_ = ch.Ack(msg.Id)
					continue
				}

				return msg, nil
			}
		}
	}
}

// TryNext 尝试获取下一条消息（可设置超时）
func (ch *redisChannel[T]) TryNext(timeout time.Duration) (*Message[T], error) {
	ctx := context.Background()
	consumer := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	ch.mu.Lock()
	if ch.closed {
		ch.mu.Unlock()
		return nil, xerror.New(ErrChannelClosed, "channel closed")
	}
	ch.mu.Unlock()

	if timeout == 0 {
		// 不阻塞，只尝试读取一次
		result, err := ch.topic.cli.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    ch.group,
			Consumer: consumer,
			Streams:  []string{ch.stream, ">"},
			Count:    1,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				return nil, xerror.New(ErrNoMessage, "no message available")
			}
			log.Errorf("err:%v", err)
			return nil, err
		}

		for _, stream := range result {
			for _, xMsg := range stream.Messages {
				msg, err := ch.parseMessage(&xMsg)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				if msg.IsExpired() {
					_ = ch.Ack(msg.Id)
					continue
				}

				return msg, nil
			}
		}

		return nil, xerror.New(ErrNoMessage, "no message available")
	}

	// 有超时时间
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		remaining := deadline.Sub(time.Now())
		if remaining > 60*time.Second {
			remaining = 60 * time.Second
		}

		result, err := ch.topic.cli.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    ch.group,
			Consumer: consumer,
			Streams:  []string{ch.stream, ">"},
			Count:    1,
			Block:    remaining,
		}).Result()

		if err != nil && err != redis.Nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		ch.mu.Lock()
		if ch.closed {
			ch.mu.Unlock()
			return nil, xerror.New(ErrChannelClosed, "channel closed")
		}
		ch.mu.Unlock()

		for _, stream := range result {
			for _, xMsg := range stream.Messages {
				msg, err := ch.parseMessage(&xMsg)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				if msg.IsExpired() {
					_ = ch.Ack(msg.Id)
					continue
				}

				return msg, nil
			}
		}
	}

	return nil, xerror.New(ErrNoMessage, "no message available")
}

// Ack 确认消息已成功处理
func (ch *redisChannel[T]) Ack(msgId string) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return xerror.New(ErrChannelClosed, "channel closed")
	}

	ctx := context.Background()
	_, err := ch.topic.cli.XAck(ctx, ch.stream, ch.group, msgId).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

// Nack 消息处理失败，重新入队
func (ch *redisChannel[T]) Nack(msgId string) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return xerror.New(ErrChannelClosed, "channel closed")
	}

	// Nack 在 Redis Streams 中通过不确认消息实现
	// 消息会在 AckTimeout 后自动重新变为可消费
	// 或者我们可以增加 Attempts 计数并重新添加到队列

	ctx := context.Background()

	// 获取消息内容
	msgs, err := ch.topic.cli.XRange(ctx, ch.stream, msgId, msgId).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	if len(msgs) == 0 {
		return xerror.New(ErrNoMessage, "message not found")
	}

	// 解析并更新消息
	xMsg := msgs[0]
	var body T
	err = json.Unmarshal([]byte(xMsg.Values["body"].(string)), &body)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// 创建新消息（增加尝试次数）
	msg := &Message[T]{
		Id:        GenerateMessageID(),
		Body:      body,
		Timestamp: time.Now().Unix(),
		Attempts:  int(xMsg.Values["attempts"].(int64)) + 1,
		Channel:   ch.name,
	}

	// 确认旧消息
	_, err = ch.topic.cli.XAck(ctx, ch.stream, ch.group, msgId).Result()
	if err != nil {
		log.Errorf("err:%v", err)
	}

	// 重新添加到队列
	bodyBytes, _ := json.Marshal(msg.Body)
	values := map[string]interface{}{
		"id":        msg.Id,
		"body":      bodyBytes,
		"timestamp": msg.Timestamp,
		"channel":   ch.name,
		"attempts":  msg.Attempts,
	}

	if msg.ExpiresAt > 0 {
		values["expires_at"] = msg.ExpiresAt
	}

	_, err = ch.topic.cli.XAdd(ctx, &redis.XAddArgs{
		Stream: ch.stream,
		MaxLen: ch.topic.config.MaxMsgSize,
		Approx: true,
		ID:     msg.Id,
		Values: values,
	}).Result()

	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

// Depth 返回 Channel 深度
func (ch *redisChannel[T]) Depth() (int64, error) {
	ctx := context.Background()
	length, err := ch.topic.cli.XLen(ctx, ch.stream).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	// 获取待处理消息数
	pending, err := ch.topic.cli.XPending(ctx, ch.stream, ch.group).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return length, nil
	}

	return length + pending.Count, nil
}

// Close 关闭 Channel
func (ch *redisChannel[T]) Close() error {
	return ch.close()
}

// close 内部关闭方法
func (ch *redisChannel[T]) close() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.closed {
		return xerror.New(ErrChannelClosed, "channel already closed")
	}

	ch.closed = true
	return nil
}

// parseMessage 解析 Redis Stream 消息
func (ch *redisChannel[T]) parseMessage(xMsg *redis.XMessage) (*Message[T], error) {
	msg := &Message[T]{}

	// 解析 ID
	msg.Id = xMsg.ID

	// 解析 Body
	bodyStr, ok := xMsg.Values["body"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid body type")
	}

	var body T
	err := json.Unmarshal([]byte(bodyStr), &body)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	msg.Body = body

	// 解析 Timestamp
	if timestamp, ok := xMsg.Values["timestamp"].(string); ok {
		var ts int64
		_, err := fmt.Sscanf(timestamp, "%d", &ts)
		if err == nil {
			msg.Timestamp = ts
		}
	}

	// 解析 ExpiresAt
	if expiresAt, ok := xMsg.Values["expires_at"].(string); ok {
		var ea int64
		_, err := fmt.Sscanf(expiresAt, "%d", &ea)
		if err == nil {
			msg.ExpiresAt = ea
		}
	}

	// 解析 Attempts
	if attempts, ok := xMsg.Values["attempts"].(int64); ok {
		msg.Attempts = int(attempts)
	}

	// 解析 Channel
	if channel, ok := xMsg.Values["channel"].(string); ok {
		msg.Channel = channel
	}

	return msg, nil
}
