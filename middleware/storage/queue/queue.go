package queue

import "fmt"

type Queue struct {
	c *Config
}

func NewQueue(c *Config) *Queue {
	c.apply()

	p := &Queue{
		c: c,
	}

	return p
}

func NewTopic[T any](queue *Queue, name string, topic *TopicConfig) Topic[T] {
	switch queue.c.StorageType {
	case StorageMemory:
		return NewMemoryTopic[T](name, topic)
	default:
		panic(fmt.Sprintf("storage type %s not supported", queue.c.StorageType))
	}
}
