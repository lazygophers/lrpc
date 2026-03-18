package cache

import (
	"strconv"
	"sync"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"gorm.io/gorm/utils"
)

// PubSubSubscription 订阅信息
type PubSubSubscription struct {
	channel string
	handler func(channel string, message []byte) error
	quit    chan struct{}
}

// StreamMessage Stream 消息
type StreamMessage struct {
	ID        string
	Values    map[string]interface{}
	CreatedAt time.Time
	Acked     bool // 是否已确认
}

// ConsumerGroup 消费者组
type ConsumerGroup struct {
	Name      string
	LastID    string
	Pending   map[string]*StreamMessage // 待处理的消息 ID -> 消息
	Consumers map[string]string         // consumer -> 最后投递的消息 ID
}

// Stream 数据流
type Stream struct {
	Messages []*StreamMessage
	Groups   map[string]*ConsumerGroup
	mu       sync.RWMutex
}

type CacheMem struct {
	sync.RWMutex

	data map[string]*Item

	// Pub/Sub
	pubsubMu      sync.RWMutex
	subscriptions map[string][]*PubSubSubscription // channel -> subscriptions

	// Stream
	streamsMu sync.RWMutex
	streams   map[string]*Stream // stream name -> stream

	// ZSet
	zsetsMu sync.RWMutex
	zsets   map[string]*ZSet // zset name -> zset

	// 全局消息 ID 生成器
	streamID int64
}

func (p *CacheMem) Clean() error {
	p.Lock()
	defer p.Unlock()

	p.data = make(map[string]*Item)

	return nil
}

func (p *CacheMem) SetPrefix(prefix string) {
}

func (p *CacheMem) IncrBy(key string, value int64) (int64, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		p.data[key] = &Item{Data: "0"}
		item = p.data[key]
	}

	current, err := strconv.ParseInt(item.Data, 10, 64)
	if err != nil {
		current = 0
	}

	newVal := current + value
	item.Data = strconv.FormatInt(newVal, 10)
	return newVal, nil
}

func (p *CacheMem) DecrBy(key string, value int64) (int64, error) {
	val, err := p.IncrBy(key, -value)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (p *CacheMem) Expire(key string, timeout time.Duration) (bool, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists {
		return false, nil
	}

	if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
		delete(p.data, key)
		return false, nil
	}

	item.ExpireAt = time.Now().Add(timeout)
	return true, nil
}

func (p *CacheMem) Ttl(key string) (time.Duration, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists {
		return -2 * time.Second, nil // Key does not exist
	}

	if item.ExpireAt.IsZero() {
		return -1 * time.Second, nil // Key has no expiration
	}

	if time.Now().After(item.ExpireAt) {
		return -2 * time.Second, nil // Key expired
	}

	return item.ExpireAt.Sub(time.Now()), nil
}

func (p *CacheMem) Incr(key string) (int64, error) {
	val, err := p.IncrBy(key, 1)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (p *CacheMem) Decr(key string) (int64, error) {
	val, err := p.IncrBy(key, -1)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (p *CacheMem) Exists(keys ...string) (bool, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	for _, key := range keys {
		item, exists := p.data[key]
		if !exists {
			return false, nil
		}

		if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
			return false, nil
		}
	}

	return true, nil
}

func (p *CacheMem) SetEx(key string, value any, timeout time.Duration) error {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	p.data[key] = &Item{
		Data:     utils.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}

	return nil
}

func (p *CacheMem) autoClear() {
	p.clear()
}

func (p *CacheMem) clear() {
	p.Lock()
	defer p.Unlock()

	data := make(map[string]*Item)

	for k, v := range p.data {
		if v.ExpireAt.IsZero() {
			data[k] = v
			continue
		}

		if time.Now().Before(v.ExpireAt) {
			data[k] = v
		}
	}

	p.data = data
}

func (p *CacheMem) SetNx(key string, value interface{}) (bool, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, ok := p.data[key]
	if ok && (item.ExpireAt.IsZero() || time.Now().Before(item.ExpireAt)) {
		return false, nil
	}

	p.data[key] = &Item{
		Data: utils.ToString(value),
	}

	return true, nil
}

func (p *CacheMem) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, ok := p.data[key]
	if ok && (item.ExpireAt.IsZero() || time.Now().Before(item.ExpireAt)) {
		return false, nil
	}

	p.data[key] = &Item{
		Data:     utils.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}

	return true, nil
}

func (p *CacheMem) Get(key string) (string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	val, ok := p.data[key]
	if !ok {
		return "", ErrNotFound
	}

	if !val.ExpireAt.IsZero() && time.Now().After(val.ExpireAt) {
		return "", ErrNotFound
	}

	return val.Data, nil
}

func (p *CacheMem) Set(key string, val any) error {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	p.data[key] = &Item{
		Data: utils.ToString(val),
	}

	return nil
}

func (p *CacheMem) Del(key ...string) error {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	for _, k := range key {
		delete(p.data, k)
	}

	return nil
}

func (p *CacheMem) Close() error {
	p.Lock()
	defer p.Unlock()

	p.data = make(map[string]*Item)

	return nil
}

func (p *CacheMem) Client() any {
	return nil
}

func (p *CacheMem) Reset() error {
	p.Lock()
	defer p.Unlock()

	p.data = make(map[string]*Item)

	return nil
}

func (p *CacheMem) Ping() error {
	return nil
}

func NewMem() Cache {
	p := &CacheMem{
		data:     make(map[string]*Item),
		streams:  make(map[string]*Stream),
		streamID: 0,
	}

	atexit.Register(func() {
		err := p.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	return newBaseCache(p)
}
