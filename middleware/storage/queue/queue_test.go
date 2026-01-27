package queue

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

// checkRedisAvailable 检查 Redis 是否可用
func checkRedisAvailable() bool {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	return err == nil
}

func TestNewQueue(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		skip   bool
	}{
		{
			name:   "默认配置",
			config: &Config{},
			skip:   false,
		},
		{
			name: "指定配置",
			config: &Config{
				StorageType: StorageMemory,
				MaxRetries:  10,
			},
			skip: false,
		},
		{
			name: "Redis配置",
			config: &Config{
				StorageType: StorageRedis,
				RedisConfig: &RedisConfig{
					Addr: "localhost:6379",
				},
			},
			skip: !checkRedisAvailable(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Redis 不可用")
			}

			q := NewQueue(tt.config)
			if q == nil {
				t.Fatal("NewQueue() = nil, want non-nil")
			}
			if q.c == nil {
				t.Error("Queue.c = nil, want non-nil")
			}

			// 清理
			if q.c.StorageType == StorageRedis {
				_ = q.Close()
			}
		})
	}
}

func TestNewTopic(t *testing.T) {
	type TestMsg struct {
		Content string
	}

	tests := []struct {
		name      string
		storage   StorageType
		wantPanic bool
	}{
		{
			name:      "内存存储",
			storage:   StorageMemory,
			wantPanic: false,
		},
		{
			name:      "未支持的存储类型",
			storage:   "unknown",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("期望 panic 但没有发生")
					}
				}()
			}

			q := NewQueue(&Config{StorageType: tt.storage})
			topic := NewTopic[TestMsg](q, "test-topic", &TopicConfig{})

			if !tt.wantPanic && topic == nil {
				t.Error("NewTopic() = nil, want non-nil")
			}
		})
	}
}
