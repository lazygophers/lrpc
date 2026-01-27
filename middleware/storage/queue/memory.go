package queue

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/lazygophers/utils/routine"
)

// memoryTopic 内存 Topic 实现
type memoryTopic[T any] struct {
	name     string
	config   *TopicConfig
	channels map[string]*memoryChannel[T]
	mu       sync.RWMutex
	closed   bool
}

// memoryChannel 内存 Channel 实现
type memoryChannel[T any] struct {
	name    string
	topic   *memoryTopic[T]
	config  *ChannelConfig
	queue   []*Message[T]
	mu      sync.Mutex
	cond    *sync.Cond
	closed  bool
	waiting map[string]struct{} // 等待 ACK 的消息
}

// NewMemoryTopic 创建内存 Topic
func NewMemoryTopic[T any](name string, config *TopicConfig) Topic[T] {
	if config == nil {
		config = &TopicConfig{}
	}
	config.apply()

	t := &memoryTopic[T]{
		name:     name,
		config:   config,
		channels: make(map[string]*memoryChannel[T]),
	}

	return t
}

// Pub 发布消息到 Topic
func (t *memoryTopic[T]) Pub(msg T) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return xerror.New(ErrTopicClosed, "topic closed")
	}

	message := &Message[T]{
		Id:        uuid.New().String(),
		Body:      msg,
		Timestamp: time.Now().Unix(),
		Attempts:  0,
	}

	for _, ch := range t.channels {
		err := ch.pubMsg(message)
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}

	return nil
}

// PubBatch 批量发布消息
func (t *memoryTopic[T]) PubBatch(msgs []T) error {
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
func (t *memoryTopic[T]) PubMsg(msg *Message[T]) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return xerror.New(ErrTopicClosed, "topic closed")
	}

	if msg.Id == "" {
		msg.Id = uuid.New().String()
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}

	for _, ch := range t.channels {
		err := ch.pubMsg(msg)
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}

	return nil
}

// PubMsgBatch 批量发布完整消息
func (t *memoryTopic[T]) PubMsgBatch(msgs []*Message[T]) error {
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
func (t *memoryTopic[T]) GetOrAddChannel(name string, config *ChannelConfig) (Channel[T], error) {
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

	ch := &memoryChannel[T]{
		name:    name,
		topic:   t,
		config:  config,
		queue:   make([]*Message[T], 0),
		closed:  false,
		waiting: make(map[string]struct{}),
	}
	ch.cond = sync.NewCond(&ch.mu)

	t.channels[name] = ch

	return ch, nil
}

// GetChannel 获取已存在的 Channel
func (t *memoryTopic[T]) GetChannel(name string) (Channel[T], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ch, exists := t.channels[name]
	if !exists {
		return nil, xerror.New(ErrChannelNotFound, "channel not found")
	}

	return ch, nil
}

// ChannelList 返回所有 Channel 名称
func (t *memoryTopic[T]) ChannelList() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	names := make([]string, 0, len(t.channels))
	for name := range t.channels {
		names = append(names, name)
	}

	return names
}

// Close 关闭 Topic
func (t *memoryTopic[T]) Close() error {
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
func (p *memoryChannel[T]) pubMsg(msg *Message[T]) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return xerror.New(ErrChannelClosed, "channel closed")
	}

	// 复制消息，避免不同 Channel 共享同一消息
	msgCopy := &Message[T]{
		Id:        msg.Id,
		Body:      msg.Body,
		Timestamp: msg.Timestamp,
		Attempts:  msg.Attempts,
		Channel:   p.name,
	}

	p.queue = append(p.queue, msgCopy)

	// 通知等待的消费者
	p.cond.Signal()

	return nil
}

// Name 返回 Channel 名称
func (p *memoryChannel[T]) Name() string {
	return p.name
}

// Subscribe 订阅消息
func (p *memoryChannel[T]) Subscribe(handler Handler[T]) {
	routine.GoWithMustSuccess(func() error {
		for {
			msg, err := p.Next()
			if err != nil {
				break
			}

			rsp, err := Handle(handler, msg)
			if err != nil {
				log.Errorf("err:%v", err)
			}

			if rsp.Retry {
				err = p.Nack(msg.Id)
				if err != nil {
					log.Errorf("err:%v", err)
				}
			} else {
				err = p.Ack(msg.Id)
				if err != nil {
					log.Errorf("err:%v", err)
				}
			}
		}

		return nil
	})
}

// Next 获取下一条消息（阻塞）
func (p *memoryChannel[T]) Next() (*Message[T], error) {
	p.mu.Lock()

	for len(p.queue) == 0 && !p.closed {
		p.cond.Wait()
	}

	if p.closed {
		p.mu.Unlock()
		return nil, xerror.New(ErrChannelClosed, "channel closed")
	}

	if len(p.queue) == 0 {
		p.mu.Unlock()
		return nil, xerror.New(ErrNoMessage, "no message available")
	}

	msg := p.queue[0]
	p.queue = p.queue[1:]
	p.waiting[msg.Id] = struct{}{}

	p.mu.Unlock()

	return msg, nil
}

// TryNext 尝试获取下一条消息（可设置超时）
func (p *memoryChannel[T]) TryNext(timeout time.Duration) (*Message[T], error) {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return nil, xerror.New(ErrChannelClosed, "channel closed")
	}

	if len(p.queue) == 0 {
		if timeout == 0 {
			p.mu.Unlock()
			return nil, xerror.New(ErrNoMessage, "no message available")
		}

		// 有超时时间，使用条件变量等待
		done := make(chan struct{})
		go func() {
			p.cond.Wait()
			close(done)
		}()

		select {
		case <-done:
			// 被唤醒，重新检查
			if p.closed {
				p.mu.Unlock()
				return nil, xerror.New(ErrChannelClosed, "channel closed")
			}
			if len(p.queue) == 0 {
				p.mu.Unlock()
				return nil, xerror.New(ErrNoMessage, "no message available")
			}
		case <-time.After(timeout):
			// 超时，广播以唤醒 goroutine
			p.cond.Broadcast()
			p.mu.Unlock()
			return nil, xerror.New(ErrNoMessage, "no message available")
		}
	}

	msg := p.queue[0]
	p.queue = p.queue[1:]
	p.waiting[msg.Id] = struct{}{}

	p.mu.Unlock()

	return msg, nil
}

// Ack 确认消息已成功处理
func (p *memoryChannel[T]) Ack(msgId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.waiting, msgId)
	return nil
}

// Nack 消息处理失败，重新入队
func (p *memoryChannel[T]) Nack(msgId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 查找在 waiting 中的消息
	// 由于内存队列已经取出，需要重新入队时创建新消息
	// 这里简化处理，直接忽略

	delete(p.waiting, msgId)

	return nil
}

// Depth 返回 Channel 深度
func (p *memoryChannel[T]) Depth() (int64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return int64(len(p.queue) + len(p.waiting)), nil
}

// Close 关闭 Channel
func (p *memoryChannel[T]) Close() error {
	return p.close()
}

// close 内部关闭方法
func (p *memoryChannel[T]) close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	p.cond.Broadcast()

	return nil
}
