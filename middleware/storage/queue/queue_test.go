package queue

import (
	"testing"
)

func TestNewQueue(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "默认配置",
			config: &Config{},
		},
		{
			name: "指定配置",
			config: &Config{
				StorageType: StorageMemory,
				MaxRetries:  10,
			},
		},
		{
			name: "Redis配置",
			config: &Config{
				StorageType: StorageRedis,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQueue(tt.config)
			if q == nil {
				t.Fatal("NewQueue() = nil, want non-nil")
			}
			if q.c == nil {
				t.Error("Queue.c = nil, want non-nil")
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
