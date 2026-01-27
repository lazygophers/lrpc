package queue

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type StorageType string

const (
	StorageMemory StorageType = "memory"
	StorageRedis  StorageType = "redis"
	StorageKafka  StorageType = "kafka"
)

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

// RedisConfig Redis 配置
type RedisConfig struct {
	// Addr Redis 服务器地址（格式：host:port）
	Addr string `json:"addr,omitempty" yaml:"addr,omitempty" toml:"addr,omitempty" default:"localhost:6379"`
	// Password Redis 密码
	Password string `json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty"`
	// DB Redis 数据库编号
	DB int `json:"db,omitempty" yaml:"db,omitempty" toml:"db,omitempty" default:"0"`
	// KeyPrefix Redis 键名前缀
	KeyPrefix string `json:"key_prefix,omitempty" yaml:"key_prefix,omitempty" toml:"key_prefix,omitempty" default:"lrpc:queue:"`
	// PoolSize 连接池大小
	PoolSize int `json:"pool_size,omitempty" yaml:"pool_size,omitempty" toml:"pool_size,omitempty" default:"10"`
	// MinIdleConns 最小空闲连接数
	MinIdleConns int `json:"min_idle_conns,omitempty" yaml:"min_idle_conns,omitempty" toml:"min_idle_conns,omitempty" default:"5"`
	// MaxRetries 最大重试次数
	MaxRetries int `json:"max_retries,omitempty" yaml:"max_retries,omitempty" toml:"max_retries,omitempty" default:"3"`
	// DialTimeout 连接超时
	DialTimeout time.Duration `json:"dial_timeout,omitempty" yaml:"dial_timeout,omitempty" toml:"dial_timeout,omitempty" default:"5s"`
	// ReadTimeout 读取超时
	ReadTimeout time.Duration `json:"read_timeout,omitempty" yaml:"read_timeout,omitempty" toml:"read_timeout,omitempty" default:"3s"`
	// WriteTimeout 写入超时
	WriteTimeout time.Duration `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty" toml:"write_timeout,omitempty" default:"3s"`
	// PoolTimeout 连接池超时
	PoolTimeout time.Duration `json:"pool_timeout,omitempty" yaml:"pool_timeout,omitempty" toml:"pool_timeout,omitempty" default:"4s"`
}

func (p *RedisConfig) apply() {
	if p.Addr == "" {
		p.Addr = "localhost:6379"
	}
	if p.DB < 0 {
		p.DB = 0
	}
	if p.KeyPrefix == "" {
		p.KeyPrefix = "lrpc:queue:"
	}
	if p.PoolSize <= 0 {
		p.PoolSize = 10
	}
	if p.MinIdleConns <= 0 {
		p.MinIdleConns = 5
	}
	if p.MaxRetries <= 0 {
		p.MaxRetries = 3
	}
	if p.DialTimeout <= 0 {
		p.DialTimeout = 5 * time.Second
	}
	if p.ReadTimeout <= 0 {
		p.ReadTimeout = 3 * time.Second
	}
	if p.WriteTimeout <= 0 {
		p.WriteTimeout = 3 * time.Second
	}
	if p.PoolTimeout <= 0 {
		p.PoolTimeout = 4 * time.Second
	}
}

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	// Brokers Kafka 服务器地址列表（格式：host:port）
	Brokers []string `json:"brokers,omitempty" yaml:"brokers,omitempty" toml:"brokers,omitempty"`
	// TopicPrefix Topic 名称前缀
	TopicPrefix string `json:"topic_prefix,omitempty" yaml:"topic_prefix,omitempty" toml:"topic_prefix,omitempty" default:"lrpc-queue-"`
	// Partition 使用的分区数（创建 Topic 时使用）
	Partition int `json:"partition,omitempty" yaml:"partition,omitempty" toml:"partition,omitempty" default:"1"`
	// ReplicationFactor 副本因子（创建 Topic 时使用）
	ReplicationFactor int `json:"replication_factor,omitempty" yaml:"replication_factor,omitempty" toml:"replication_factor,omitempty" default:"1"`
	// ConsumerGroupID 消费者组 ID
	ConsumerGroupID string `json:"consumer_group_id,omitempty" yaml:"consumer_group_id,omitempty" toml:"consumer_group_id,omitempty"`
	// AutoCreateTopics 是否自动创建 Topic
	AutoCreateTopics bool `json:"auto_create_topics,omitempty" yaml:"auto_create_topics,omitempty" toml:"auto_create_topics,omitempty" default:"true"`
	// ReadBatchTimeout 批量读取超时
	ReadBatchTimeout time.Duration `json:"read_batch_timeout,omitempty" yaml:"read_batch_timeout,omitempty" toml:"read_batch_timeout,omitempty" default:"10s"`
	// WriteTimeout 写入超时
	WriteTimeout time.Duration `json:"write_timeout,omitempty" yaml:"write_timeout,omitempty" toml:"write_timeout,omitempty" default:"10s"`
	// RequiredAcks 确认级别（0=无需确认，1=leader确认，-1=all确认）
	RequiredAcks int `json:"required_acks,omitempty" yaml:"required_acks,omitempty" toml:"required_acks,omitempty" default:"1"`
	// CompressionType 压缩类型（none, gzip, snappy, lz4, zstd）
	CompressionType string `json:"compression_type,omitempty" yaml:"compression_type,omitempty" toml:"compression_type,omitempty" default:"none"`
	// SessionTimeout 会话超时
	SessionTimeout time.Duration `json:"session_timeout,omitempty" yaml:"session_timeout,omitempty" toml:"session_timeout,omitempty" default:"30s"`
	// RebalanceTimeout 重平衡超时
	RebalanceTimeout time.Duration `json:"rebalance_timeout,omitempty" yaml:"rebalance_timeout,omitempty" toml:"rebalance_timeout,omitempty" default:"60s"`
	// CommitInterval 提交间隔
	CommitInterval time.Duration `json:"commit_interval,omitempty" yaml:"commit_interval,omitempty" toml:"commit_interval,omitempty" default:"1s"`
	// HeartbeatInterval 心跳间隔
	HeartbeatInterval time.Duration `json:"heartbeat_interval,omitempty" yaml:"heartbeat_interval,omitempty" toml:"heartbeat_interval,omitempty" default:"3s"`
	// MaxAttempts 最大消费尝试次数
	MaxAttempts int `json:"max_attempts,omitempty" yaml:"max_attempts,omitempty" toml:"max_attempts,omitempty" default:"5"`
	// DialTimeout 连接超时
	DialTimeout time.Duration `json:"dial_timeout,omitempty" yaml:"dial_timeout,omitempty" toml:"dial_timeout,omitempty" default:"10s"`
}

func (p *KafkaConfig) apply() {
	if len(p.Brokers) == 0 {
		p.Brokers = []string{"localhost:9092"}
	}
	if p.TopicPrefix == "" {
		p.TopicPrefix = "lrpc-queue-"
	}
	if p.Partition <= 0 {
		p.Partition = 1
	}
	if p.ReplicationFactor <= 0 {
		p.ReplicationFactor = 1
	}
	if p.ConsumerGroupID == "" {
		p.ConsumerGroupID = "lrpc-queue"
	}
	if p.ReadBatchTimeout <= 0 {
		p.ReadBatchTimeout = 10 * time.Second
	}
	if p.WriteTimeout <= 0 {
		p.WriteTimeout = 10 * time.Second
	}
	if p.RequiredAcks == 0 {
		p.RequiredAcks = 1
	}
	if p.CompressionType == "" {
		p.CompressionType = "none"
	}
	if p.SessionTimeout <= 0 {
		p.SessionTimeout = 30 * time.Second
	}
	if p.RebalanceTimeout <= 0 {
		p.RebalanceTimeout = 60 * time.Second
	}
	if p.CommitInterval <= 0 {
		p.CommitInterval = 1 * time.Second
	}
	if p.HeartbeatInterval <= 0 {
		p.HeartbeatInterval = 3 * time.Second
	}
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 5
	}
	if p.DialTimeout <= 0 {
		p.DialTimeout = 10 * time.Second
	}
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

	// RedisConfig Redis 配置（当 StorageType 为 redis 时使用）
	RedisConfig *RedisConfig `json:"redis_config,omitempty" yaml:"redis_config,omitempty" toml:"redis_config,omitempty"`

	// RedisClient 外部传入的 Redis 客户端（优先级高于 RedisConfig）
	RedisClient *redis.Client `json:"-" yaml:"-" toml:"-"`

	// KafkaConfig Kafka 配置（当 StorageType 为 kafka 时使用）
	KafkaConfig *KafkaConfig `json:"kafka_config,omitempty" yaml:"kafka_config,omitempty" toml:"kafka_config,omitempty"`
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
