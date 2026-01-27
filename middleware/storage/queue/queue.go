package queue

import (
	"context"
	"fmt"
	"time"

	kafka "github.com/segmentio/kafka-go"

	"github.com/lazygophers/log"
	"github.com/redis/go-redis/v9"
)

type Queue struct {
	c           *Config
	redisClient *redis.Client
	kafkaWriter *kafka.Writer
}

func NewQueue(c *Config) *Queue {
	c.apply()

	p := &Queue{
		c: c,
	}

	// 根据存储类型初始化客户端
	switch c.StorageType {
	case StorageRedis:
		p.initRedisClient()
	case StorageKafka:
		p.initKafkaClient()
	}

	return p
}

// initRedisClient 初始化 Redis 客户端
func (q *Queue) initRedisClient() {
	// 优先使用外部传入的客户端
	if q.c.RedisClient != nil {
		q.redisClient = q.c.RedisClient
		return
	}

	// 使用配置创建客户端
	if q.c.RedisConfig == nil {
		q.c.RedisConfig = &RedisConfig{}
	}
	q.c.RedisConfig.apply()

	config := q.c.RedisConfig
	q.redisClient = redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
	})

	// 测试连接
	ctx := context.Background()
	_, err := q.redisClient.Ping(ctx).Result()
	if err != nil {
		log.Errorf("redis ping failed: %v", err)
		panic(fmt.Sprintf("redis connection failed: %v", err))
	}
}

// initKafkaClient 初始化 Kafka Writer
func (q *Queue) initKafkaClient() {
	if q.c.KafkaConfig == nil {
		q.c.KafkaConfig = &KafkaConfig{}
	}
	q.c.KafkaConfig.apply()

	config := q.c.KafkaConfig

	// 创建 Kafka Writer
	// 注意：Writer 的 topic 是动态的，在创建 Topic 时指定
	q.kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		BatchBytes:   1048576,
		ReadTimeout:  config.ReadBatchTimeout,
		WriteTimeout: config.WriteTimeout,
		RequiredAcks: kafka.RequiredAcks(config.RequiredAcks),
		MaxAttempts:  config.MaxAttempts,
	}
}

func NewTopic[T any](queue *Queue, name string, topic *TopicConfig) Topic[T] {
	switch queue.c.StorageType {
	case StorageMemory:
		return NewMemoryTopic[T](name, topic)
	case StorageRedis:
		if queue.redisClient == nil {
			panic("redis client not initialized")
		}
		prefix := ""
		if queue.c.RedisConfig != nil {
			prefix = queue.c.RedisConfig.KeyPrefix
		}
		return NewRedisTopic[T](queue.redisClient, name, topic, prefix)
	case StorageKafka:
		if queue.kafkaWriter == nil {
			panic("kafka writer not initialized")
		}
		prefix := ""
		if queue.c.KafkaConfig != nil {
			prefix = queue.c.KafkaConfig.TopicPrefix
		}
		// 为每个 topic 创建专用的 writer
		writer := NewKafkaWriter(queue.c.KafkaConfig.Brokers, prefix+name, queue.c.KafkaConfig)
		return NewKafkaTopic[T](writer, name, topic, prefix, queue.c.KafkaConfig.Brokers, queue.c.KafkaConfig)
	default:
		panic(fmt.Sprintf("storage type %s not supported", queue.c.StorageType))
	}
}

// Close 关闭队列
func (q *Queue) Close() error {
	if q.redisClient != nil && q.c.RedisClient == nil {
		// 只有当客户端是由队列创建时才关闭
		_ = q.redisClient.Close()
	}

	if q.kafkaWriter != nil {
		_ = q.kafkaWriter.Close()
	}

	return nil
}
