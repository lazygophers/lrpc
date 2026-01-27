package queue

import (
	"time"

	"github.com/lazygophers/utils/cryptox"
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

// NewMessage 创建新消息
func NewMessage[T any](body T) *Message[T] {
	return &Message[T]{
		Id:        GenerateMessageID(),
		Body:      body,
		Timestamp: time.Now().Unix(),
		Attempts:  0,
	}
}

// NewMessageWithID 使用指定 ID 创建消息
func NewMessageWithID[T any](id string, body T) *Message[T] {
	return &Message[T]{
		Id:        id,
		Body:      body,
		Timestamp: time.Now().Unix(),
		Attempts:  0,
	}
}

// GenerateMessageID 生成消息唯一标识
func GenerateMessageID() string {
	return cryptox.ULID()
}

// Clone 创建消息的副本（用于不同 Channel）
func (m *Message[T]) Clone() *Message[T] {
	return &Message[T]{
		Id:        m.Id,
		Body:      m.Body,
		Timestamp: m.Timestamp,
		Attempts:  m.Attempts,
		Channel:   m.Channel,
	}
}

// ResetAttempts 重置尝试次数
func (m *Message[T]) ResetAttempts() {
	m.Attempts = 0
}

// IncrementAttempts 增加尝试次数
func (m *Message[T]) IncrementAttempts() {
	m.Attempts++
}
