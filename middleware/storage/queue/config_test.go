package queue

import (
	"testing"
	"time"
)

func TestConfigApply(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   StorageType
	}{
		{
			name:   "默认值",
			config: &Config{},
			want:   StorageMemory,
		},
		{
			name: "指定存储类型",
			config: &Config{
				StorageType: StorageRedis,
			},
			want: StorageRedis,
		},
		{
			name: "负值重试次数",
			config: &Config{
				MaxRetries: -1,
			},
			want: StorageMemory,
		},
		{
			name: "零值重试延迟",
			config: &Config{
				RetryDelay: 0,
			},
			want: StorageMemory,
		},
		{
			name: "零值TTL",
			config: &Config{
				MessageTTL: 0,
			},
			want: StorageMemory,
		},
		{
			name: "负值消息大小限制",
			config: &Config{
				MaxBodySize: -1,
			},
			want: StorageMemory,
		},
		{
			name: "负值消息数量限制",
			config: &Config{
				MaxMsgSize: -1,
			},
			want: StorageMemory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.apply()
			if tt.config.StorageType != tt.want {
				t.Errorf("StorageType = %v, want %v", tt.config.StorageType, tt.want)
			}
			if tt.config.MaxRetries <= 0 {
				t.Errorf("MaxRetries = %d, want > 0", tt.config.MaxRetries)
			}
			if tt.config.RetryDelay <= 0 {
				t.Errorf("RetryDelay = %v, want > 0", tt.config.RetryDelay)
			}
			if tt.config.MessageTTL <= 0 {
				t.Errorf("MessageTTL = %v, want > 0", tt.config.MessageTTL)
			}
			if tt.config.MaxBodySize <= 0 {
				t.Errorf("MaxBodySize = %d, want > 0", tt.config.MaxBodySize)
			}
			if tt.config.MaxMsgSize <= 0 {
				t.Errorf("MaxMsgSize = %d, want > 0", tt.config.MaxMsgSize)
			}
		})
	}
}

func TestTopicConfigApply(t *testing.T) {
	tests := []struct {
		name   string
		config *TopicConfig
	}{
		{
			name:   "默认值",
			config: &TopicConfig{},
		},
		{
			name: "负值重试次数",
			config: &TopicConfig{
				MaxRetries: -1,
			},
		},
		{
			name: "零值重试延迟",
			config: &TopicConfig{
				RetryDelay: 0,
			},
		},
		{
			name: "零值TTL",
			config: &TopicConfig{
				MessageTTL: 0,
			},
		},
		{
			name: "负值消息大小限制",
			config: &TopicConfig{
				MaxBodySize: -1,
			},
		},
		{
			name: "负值消息数量限制",
			config: &TopicConfig{
				MaxMsgSize: -1,
			},
		},
		{
			name: "完整配置",
			config: &TopicConfig{
				MaxRetries:  10,
				RetryDelay:  5 * time.Second,
				MessageTTL:  48 * time.Hour,
				MaxBodySize: 2048 * 1024,
				MaxMsgSize:  2000000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.apply()
			if tt.config.MaxRetries <= 0 {
				t.Errorf("MaxRetries = %d, want > 0", tt.config.MaxRetries)
			}
			if tt.config.RetryDelay <= 0 {
				t.Errorf("RetryDelay = %v, want > 0", tt.config.RetryDelay)
			}
			if tt.config.MessageTTL <= 0 {
				t.Errorf("MessageTTL = %v, want > 0", tt.config.MessageTTL)
			}
			if tt.config.MaxBodySize <= 0 {
				t.Errorf("MaxBodySize = %d, want > 0", tt.config.MaxBodySize)
			}
			if tt.config.MaxMsgSize <= 0 {
				t.Errorf("MaxMsgSize = %d, want > 0", tt.config.MaxMsgSize)
			}
		})
	}
}

func TestChannelConfigApply(t *testing.T) {
	tests := []struct {
		name   string
		config *ChannelConfig
	}{
		{
			name:   "默认值",
			config: &ChannelConfig{},
		},
		{
			name: "负值重试次数",
			config: &ChannelConfig{
				MaxRetries: -1,
			},
		},
		{
			name: "零值重试延迟",
			config: &ChannelConfig{
				RetryDelay: 0,
			},
		},
		{
			name: "零值TTL",
			config: &ChannelConfig{
				MessageTTL: 0,
			},
		},
		{
			name: "零值最大飞行数",
			config: &ChannelConfig{
				MaxInFlight: 0,
			},
		},
		{
			name: "零值ACK超时",
			config: &ChannelConfig{
				AckTimeout: 0,
			},
		},
		{
			name: "完整配置",
			config: &ChannelConfig{
				MaxRetries:  10,
				RetryDelay:  5 * time.Second,
				MessageTTL:  48 * time.Hour,
				MaxInFlight: 20,
				AckTimeout:  60 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.apply()
			if tt.config.MaxRetries <= 0 {
				t.Errorf("MaxRetries = %d, want > 0", tt.config.MaxRetries)
			}
			if tt.config.RetryDelay <= 0 {
				t.Errorf("RetryDelay = %v, want > 0", tt.config.RetryDelay)
			}
			if tt.config.MessageTTL <= 0 {
				t.Errorf("MessageTTL = %v, want > 0", tt.config.MessageTTL)
			}
			if tt.config.MaxInFlight <= 0 {
				t.Errorf("MaxInFlight = %d, want > 0", tt.config.MaxInFlight)
			}
			if tt.config.AckTimeout <= 0 {
				t.Errorf("AckTimeout = %v, want > 0", tt.config.AckTimeout)
			}
		})
	}
}
