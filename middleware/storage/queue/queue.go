package queue

import (
	"context"
	"fmt"

	"github.com/lazygophers/log"
	"github.com/redis/go-redis/v9"
)

type Queue struct {
	c           *Config
	redisClient *redis.Client
}

func NewQueue(c *Config) *Queue {
	c.apply()

	p := &Queue{
		c: c,
	}

	// 初始化 Redis 客户端（如果需要）
	if c.StorageType == StorageRedis {
		p.initRedisClient()
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
	default:
		panic(fmt.Sprintf("storage type %s not supported", queue.c.StorageType))
	}
}

// Close 关闭队列
func (q *Queue) Close() error {
	if q.redisClient != nil && q.c.RedisClient == nil {
		// 只有当客户端是由队列创建时才关闭
		return q.redisClient.Close()
	}
	return nil
}
