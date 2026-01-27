package queue

import (
	"time"
)

type StorageType string

const (
	StorageMemory StorageType = "memory"
	StorageRedis  StorageType = "redis"
)

// Message 泛型消息结构
type Message[T any] struct {
	// Id 消息唯一标识
	Id string `json:"id,omitempty" yaml:"id,omitempty" toml:"id,omitempty"`
	// Body 消息体，使用泛型支持任意类型
	Body T `json:"body,omitempty" yaml:"body,omitempty" toml:"body,omitempty"`
	// Timestamp 消息产生时间戳
	Timestamp int64 `json:"timestamp,omitempty" yaml:"timestamp,omitempty" toml:"timestamp,omitempty"`
	// Attempts 消费尝试次数
	Attempts int `json:"attempts,omitempty" yaml:"attempts,omitempty" toml:"attempts,omitempty"`
	// Channel 所属 Channel
	Channel string `json:"channel,omitempty" yaml:"channel,omitempty" toml:"channel,omitempty"`
}

// Topic Topic 接口，负责消息的生产和分发
type Topic[T any] interface {
	// Pub 发布消息到 Topic
	Pub(msg T) error

	// PubBatch 批量发布消息
	PubBatch(msgs []T) error

	// PubMsg 发布完整消息（包含元数据）
	PubMsg(msg *Message[T]) error

	// PubMsgBatch 批量发布完整消息
	PubMsgBatch(msgs []*Message[T]) error

	// GetOrAddChannel 获取或创建一个 Channel
	GetOrAddChannel(name string, config *ChannelConfig) (Channel[T], error)

	// GetChannel 获取已存在的 Channel
	GetChannel(name string) (Channel[T], error)

	// ChannelList 返回所有 Channel 名称
	ChannelList() []string

	// Close 关闭 Topic
	Close() error
}

// ProcessRsp 消息处理响应
type ProcessRsp struct {
	// Retry 是否需要重试
	Retry bool
	// SkipAttempts 跳过记录重试次数
	SkipAttempts bool
}

// Handler 消息处理函数类型
type Handler[T any] func(msg *Message[T]) (ProcessRsp, error)

// Channel Channel 接口，负责消息的消费
type Channel[T any] interface {
	// Name 返回 Channel 名称
	Name() string

	// Subscribe 订阅消息，使用回调函数处理
	Subscribe(handler Handler[T])

	// Next 获取下一条消息（阻塞）
	Next() (*Message[T], error)

	// TryNext 尝试获取下一条消息（可设置超时）
	TryNext(timeout time.Duration) (*Message[T], error)

	// Ack 确认消息已成功处理
	Ack(msgId string) error

	// Nack 消息处理失败，重新入队
	Nack(msgId string) error

	// Depth 返回 Channel 深度（包括待处理和未确认）
	Depth() (int64, error)

	// Close 关闭 Channel
	Close() error
}

// Config 队列配置
type Config struct {
	// StorageType 存储类型
	StorageType StorageType `json:"storage_type,omitempty" yaml:"storage_type,omitempty" toml:"storage_type,omitempty" default:"memory"`
	// MaxRetries 最大重试次数
	MaxRetries int `json:"max_retries,omitempty" yaml:"max_retries,omitempty" toml:"max_retries,omitempty" default:"5"`
	// RetryDelay 重试延迟
	RetryDelay time.Duration `json:"retry_delay,omitempty" yaml:"retry_delay,omitempty" toml:"retry_delay,omitempty" default:"1s"`
	// MessageTTL 消息过期时间
	MessageTTL time.Duration `json:"message_ttl,omitempty" yaml:"message_ttl,omitempty" toml:"message_ttl,omitempty" default:"24h"`
	// MaxBodySize 最大消息体大小
	MaxBodySize int64 `json:"max_body_size,omitempty" yaml:"max_body_size,omitempty" toml:"max_body_size,omitempty" default:"1048576"`
	// MaxMsgSize 最大消息数量
	MaxMsgSize int64 `json:"max_msg_size,omitempty" yaml:"max_msg_size,omitempty" toml:"max_msg_size,omitempty" default:"1000000"`
}

func (p *Config) apply() {
	if p.StorageType == "" {
		p.StorageType = StorageMemory
	}
	if p.MaxRetries <= 0 {
		p.MaxRetries = 5
	}
	if p.RetryDelay <= 0 {
		p.RetryDelay = time.Second
	}
	if p.MessageTTL <= 0 {
		p.MessageTTL = 24 * time.Hour
	}
	if p.MaxBodySize <= 0 {
		p.MaxBodySize = 1024 * 1024
	}
	if p.MaxMsgSize <= 0 {
		p.MaxMsgSize = 1000000
	}
}

// TopicConfig Topic 配置
type TopicConfig struct {
	// MaxRetries 最大重试次数
	MaxRetries int `json:"max_retries,omitempty" yaml:"max_retries,omitempty" toml:"max_retries,omitempty" default:"5"`
	// RetryDelay 重试延迟
	RetryDelay time.Duration `json:"retry_delay,omitempty" yaml:"retry_delay,omitempty" toml:"retry_delay,omitempty" default:"1s"`
	// MessageTTL 消息过期时间
	MessageTTL time.Duration `json:"message_ttl,omitempty" yaml:"message_ttl,omitempty" toml:"message_ttl,omitempty" default:"24h"`
	// MaxBodySize 最大消息体大小
	MaxBodySize int64 `json:"max_body_size,omitempty" yaml:"max_body_size,omitempty" toml:"max_body_size,omitempty" default:"1048576"`
	// MaxMsgSize 最大消息数量
	MaxMsgSize int64 `json:"max_msg_size,omitempty" yaml:"max_msg_size,omitempty" toml:"max_msg_size,omitempty" default:"1000000"`
}

func (p *TopicConfig) apply() {
	if p.MaxRetries <= 0 {
		p.MaxRetries = 5
	}
	if p.RetryDelay <= 0 {
		p.RetryDelay = time.Second
	}
	if p.MessageTTL <= 0 {
		p.MessageTTL = 24 * time.Hour
	}
	if p.MaxBodySize <= 0 {
		p.MaxBodySize = 1024 * 1024
	}
	if p.MaxMsgSize <= 0 {
		p.MaxMsgSize = 1000000
	}
}

// ChannelConfig Channel 配置
type ChannelConfig struct {
	// MaxRetries 最大重试次数
	MaxRetries int `json:"max_retries,omitempty" yaml:"max_retries,omitempty" toml:"max_retries,omitempty" default:"5"`
	// RetryDelay 重试延迟
	RetryDelay time.Duration `json:"retry_delay,omitempty" yaml:"retry_delay,omitempty" toml:"retry_delay,omitempty" default:"1s"`
	// MessageTTL 消息过期时间
	MessageTTL time.Duration `json:"message_ttl,omitempty" yaml:"message_ttl,omitempty" toml:"message_ttl,omitempty" default:"24h"`
	// MaxInFlight 最大飞行消息数（同时消费的消息数）
	MaxInFlight int `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty" toml:"max_in_flight,omitempty" default:"10"`
	// AckTimeout 超时自动重新入队
	AckTimeout time.Duration `json:"ack_timeout,omitempty" yaml:"ack_timeout,omitempty" toml:"ack_timeout,omitempty" default:"30s"`
}

func (p *ChannelConfig) apply() {
	if p.MaxRetries <= 0 {
		p.MaxRetries = 5
	}
	if p.RetryDelay <= 0 {
		p.RetryDelay = time.Second
	}
	if p.MessageTTL <= 0 {
		p.MessageTTL = 24 * time.Hour
	}
	if p.MaxInFlight <= 0 {
		p.MaxInFlight = 10
	}
	if p.AckTimeout <= 0 {
		p.AckTimeout = 30 * time.Second
	}
}

// 错误码定义
const (
	// ErrQueueClosed 队列已关闭
	ErrQueueClosed = 20001
	// ErrTopicClosed Topic 已关闭
	ErrTopicClosed = 20002
	// ErrChannelClosed Channel 已关闭
	ErrChannelClosed = 20003
	// ErrChannelNotFound Channel 不存在
	ErrChannelNotFound = 20004
	// ErrNoMessage 没有可用消息
	ErrNoMessage = 20005
)
