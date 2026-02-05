package cache

import (
	"github.com/lazygophers/log"
	"gorm.io/gorm/utils"
)

// Publish 发布消息到指定频道
func (p *CacheMem) Publish(channel string, message interface{}) (int64, error) {
	p.pubsubMu.RLock()
	defer p.pubsubMu.RUnlock()

	subs, exists := p.subscriptions[channel]
	if !exists || len(subs) == 0 {
		return 0, nil
	}

	// 将消息转换为字节
	var msgBytes []byte
	switch v := message.(type) {
	case []byte:
		msgBytes = v
	case string:
		msgBytes = []byte(v)
	default:
		msgBytes = []byte(utils.ToString(message))
	}

	// 异步发送给所有订阅者
	sentCount := int64(0)
	for _, sub := range subs {
		go func(s *PubSubSubscription) {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("panic in pub/sub handler: %v", r)
				}
			}()

			err := s.handler(s.channel, msgBytes)
			if err != nil {
				log.Errorf("err:%v", err)
			}
		}(sub)
		sentCount++
	}

	return sentCount, nil
}

// Subscribe 订阅一个或多个频道
func (p *CacheMem) Subscribe(handler func(channel string, message []byte) error, channels ...string) error {
	p.pubsubMu.Lock()
	defer p.pubsubMu.Unlock()

	if p.subscriptions == nil {
		p.subscriptions = make(map[string][]*PubSubSubscription)
	}

	for _, channel := range channels {
		sub := &PubSubSubscription{
			channel: channel,
			handler: handler,
			quit:    make(chan struct{}),
		}
		p.subscriptions[channel] = append(p.subscriptions[channel], sub)
	}

	return nil
}
