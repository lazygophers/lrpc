package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
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
		Data:     candy.ToString(value),
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
		Data: candy.ToString(value),
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
		Data:     candy.ToString(value),
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
		Data: candy.ToString(val),
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

func init() {
	RegisterBuilder(Mem, func(c *Config) (Cache, error) {
		return NewMem(), nil
	})
}

func (p *CacheMem) HIncr(key string, subKey string) (int64, error) {
	val, err := p.HIncrBy(key, subKey, 1)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (p *CacheMem) HIncrBy(key string, field string, increment int64) (int64, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		p.data[key] = &Item{Data: "{}"}
		item = p.data[key]
	}

	var hashMap map[string]string
	if err := json.Unmarshal([]byte(item.Data), &hashMap); err != nil {
		hashMap = make(map[string]string)
	}

	current, err := strconv.ParseInt(hashMap[field], 10, 64)
	if err != nil {
		current = 0
	}

	newVal := current + increment
	hashMap[field] = strconv.FormatInt(newVal, 10)

	data, _ := json.Marshal(hashMap)
	item.Data = string(data)

	return newVal, nil
}

func (p *CacheMem) HDecr(key string, field string) (int64, error) {
	val, err := p.HIncrBy(key, field, -1)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (p *CacheMem) HDecrBy(key string, field string, increment int64) (int64, error) {
	val, err := p.HIncrBy(key, field, -increment)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (p *CacheMem) HExists(key string, field string) (bool, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return false, nil
	}

	var hashMap map[string]string
	if err := json.Unmarshal([]byte(item.Data), &hashMap); err != nil {
		return false, nil
	}

	_, exists = hashMap[field]
	return exists, nil
}

func (p *CacheMem) HKeys(key string) ([]string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return []string{}, nil
	}

	var hashMap map[string]string
	if err := json.Unmarshal([]byte(item.Data), &hashMap); err != nil {
		return []string{}, nil
	}

	keys := make([]string, 0, len(hashMap))
	for k := range hashMap {
		keys = append(keys, k)
	}

	return keys, nil
}

func (p *CacheMem) HSet(key string, field string, value interface{}) (bool, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		p.data[key] = &Item{Data: "{}"}
		item = p.data[key]
	}

	var hashMap map[string]string
	if err := json.Unmarshal([]byte(item.Data), &hashMap); err != nil {
		hashMap = make(map[string]string)
	}

	_, existed := hashMap[field]
	hashMap[field] = candy.ToString(value)

	data, _ := json.Marshal(hashMap)
	item.Data = string(data)

	return !existed, nil
}

func (p *CacheMem) HGet(key, field string) (string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return "", ErrNotFound
	}

	var hashMap map[string]string
	if err := json.Unmarshal([]byte(item.Data), &hashMap); err != nil {
		return "", ErrNotFound
	}

	value, exists := hashMap[field]
	if !exists {
		return "", ErrNotFound
	}

	return value, nil
}

func (p *CacheMem) HDel(key string, fields ...string) (int64, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return 0, nil
	}

	var hashMap map[string]string
	if err := json.Unmarshal([]byte(item.Data), &hashMap); err != nil {
		return 0, nil
	}

	deletedCount := int64(0)
	for _, field := range fields {
		if _, exists := hashMap[field]; exists {
			delete(hashMap, field)
			deletedCount++
		}
	}

	data, _ := json.Marshal(hashMap)
	item.Data = string(data)

	return deletedCount, nil
}

func (p *CacheMem) HGetAll(key string) (map[string]string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return make(map[string]string), nil
	}

	var hashMap map[string]string
	if err := json.Unmarshal([]byte(item.Data), &hashMap); err != nil {
		return make(map[string]string), nil
	}

	return hashMap, nil
}

func (p *CacheMem) SAdd(key string, members ...string) (int64, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		p.data[key] = &Item{Data: "[]"}
		item = p.data[key]
	}

	var setMembers []string
	if err := json.Unmarshal([]byte(item.Data), &setMembers); err != nil {
		setMembers = make([]string, 0)
	}

	setMap := make(map[string]bool)
	for _, member := range setMembers {
		setMap[member] = true
	}

	addedCount := int64(0)
	for _, member := range members {
		if !setMap[member] {
			setMap[member] = true
			addedCount++
		}
	}

	newMembers := make([]string, 0, len(setMap))
	for member := range setMap {
		newMembers = append(newMembers, member)
	}

	data, _ := json.Marshal(newMembers)
	item.Data = string(data)

	return addedCount, nil
}

func (p *CacheMem) SMembers(key string) ([]string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return []string{}, nil
	}

	var members []string
	if err := json.Unmarshal([]byte(item.Data), &members); err != nil {
		return []string{}, nil
	}

	return members, nil
}

func (p *CacheMem) SRem(key string, members ...string) (int64, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return 0, nil
	}

	var setMembers []string
	if err := json.Unmarshal([]byte(item.Data), &setMembers); err != nil {
		return 0, nil
	}

	removeMap := make(map[string]bool)
	for _, member := range members {
		removeMap[member] = true
	}

	newMembers := make([]string, 0)
	removedCount := int64(0)
	for _, member := range setMembers {
		if removeMap[member] {
			removedCount++
		} else {
			newMembers = append(newMembers, member)
		}
	}

	data, _ := json.Marshal(newMembers)
	item.Data = string(data)

	return removedCount, nil
}

func (p *CacheMem) SRandMember(key string, count ...int64) ([]string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return []string{}, nil
	}

	var members []string
	if err := json.Unmarshal([]byte(item.Data), &members); err != nil {
		return []string{}, nil
	}

	if len(members) == 0 {
		return []string{}, nil
	}

	n := int64(1)
	if len(count) > 0 && count[0] > 0 {
		n = count[0]
	}

	if n >= int64(len(members)) {
		return members, nil
	}

	result := make([]string, 0, n)
	selected := make(map[int]bool)
	for int64(len(result)) < n {
		idx := rand.Intn(len(members))
		if !selected[idx] {
			selected[idx] = true
			result = append(result, members[idx])
		}
	}

	return result, nil
}

func (p *CacheMem) SPop(key string) (string, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return "", nil
	}

	var members []string
	if err := json.Unmarshal([]byte(item.Data), &members); err != nil || len(members) == 0 {
		return "", nil
	}

	idx := rand.Intn(len(members))
	popped := members[idx]
	newMembers := make([]string, 0, len(members)-1)
	for i, member := range members {
		if i != idx {
			newMembers = append(newMembers, member)
		}
	}

	data, _ := json.Marshal(newMembers)
	item.Data = string(data)

	return popped, nil
}

func (p *CacheMem) SisMember(key, field string) (bool, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	item, exists := p.data[key]
	if !exists || (!item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt)) {
		return false, nil
	}

	var members []string
	if err := json.Unmarshal([]byte(item.Data), &members); err != nil {
		return false, nil
	}

	for _, member := range members {
		if member == field {
			return true, nil
		}
	}

	return false, nil
}

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
		msgBytes = []byte(candy.ToString(message))
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

func (p *CacheMem) XAdd(stream string, values map[string]interface{}) (string, error) {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		p.streams = make(map[string]*Stream)
	}

	s, exists := p.streams[stream]
	if !exists {
		s = &Stream{
			Messages: make([]*StreamMessage, 0),
			Groups:   make(map[string]*ConsumerGroup),
		}
		p.streams[stream] = s
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成新的消息 ID (使用时间戳-序列号格式)
	id := fmt.Sprintf("%d-%d", time.Now().UnixMilli(), atomic.AddInt64(&p.streamID, 1))

	msg := &StreamMessage{
		ID:        id,
		Values:    values,
		CreatedAt: time.Now(),
		Acked:     false,
	}

	s.Messages = append(s.Messages, msg)
	return id, nil
}

func (p *CacheMem) XLen(stream string) (int64, error) {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.Messages)), nil
}

func (p *CacheMem) XRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	if p.streams == nil {
		return []map[string]interface{}{}, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return []map[string]interface{}{}, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	startIdx := 0
	endIdx := len(s.Messages)

	// 解析 start 参数
	if start != "-" && start != "0" {
		for i, msg := range s.Messages {
			if msg.ID >= start {
				startIdx = i
				break
			}
		}
	}

	// 解析 stop 参数
	if stop != "+" {
		for i := len(s.Messages) - 1; i >= 0; i-- {
			if s.Messages[i].ID <= stop {
				endIdx = i + 1
				break
			}
		}
	}

	if startIdx >= endIdx {
		return []map[string]interface{}{}, nil
	}

	// 应用 count 限制
	maxCount := int64(-1)
	if len(count) > 0 && count[0] > 0 {
		maxCount = count[0]
	}

	result := make([]map[string]interface{}, 0)
	for i := startIdx; i < endIdx && (maxCount == -1 || int64(len(result)) < maxCount); i++ {
		result = append(result, map[string]interface{}{
			"id":     s.Messages[i].ID,
			"values": s.Messages[i].Values,
		})
	}

	return result, nil
}

func (p *CacheMem) XRevRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	if p.streams == nil {
		return []map[string]interface{}{}, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return []map[string]interface{}{}, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	startIdx := len(s.Messages) - 1
	endIdx := 0

	// 解析 start 参数
	if start != "+" {
		for i := len(s.Messages) - 1; i >= 0; i-- {
			if s.Messages[i].ID <= start {
				startIdx = i
				break
			}
		}
	}

	// 解析 stop 参数
	if stop != "-" && stop != "0" {
		for i, msg := range s.Messages {
			if msg.ID >= stop {
				endIdx = i
				break
			}
		}
	}

	if startIdx < endIdx {
		return []map[string]interface{}{}, nil
	}

	// 应用 count 限制
	maxCount := int64(-1)
	if len(count) > 0 && count[0] > 0 {
		maxCount = count[0]
	}

	result := make([]map[string]interface{}, 0)
	for i := startIdx; i >= endIdx && (maxCount == -1 || int64(len(result)) < maxCount); i-- {
		result = append(result, map[string]interface{}{
			"id":     s.Messages[i].ID,
			"values": s.Messages[i].Values,
		})
	}

	return result, nil
}

func (p *CacheMem) XDel(stream string, ids ...string) (int64, error) {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	deletedCount := int64(0)
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	newMessages := make([]*StreamMessage, 0)
	for _, msg := range s.Messages {
		if idMap[msg.ID] {
			deletedCount++
		} else {
			newMessages = append(newMessages, msg)
		}
	}

	s.Messages = newMessages

	// 同时从所有消费者组的待处理消息中删除
	for _, group := range s.Groups {
		for _, id := range ids {
			delete(group.Pending, id)
		}
	}

	return deletedCount, nil
}

func (p *CacheMem) XTrim(stream string, maxLen int64) (int64, error) {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldLen := len(s.Messages)
	if int64(oldLen) <= maxLen {
		return 0, nil
	}

	// 保留最新的 maxLen 条消息
	s.Messages = s.Messages[oldLen-int(maxLen):]
	return int64(oldLen - len(s.Messages)), nil
}

func (p *CacheMem) XGroupCreate(stream, group, start string) error {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		p.streams = make(map[string]*Stream)
	}

	s, exists := p.streams[stream]
	if !exists {
		s = &Stream{
			Messages: make([]*StreamMessage, 0),
			Groups:   make(map[string]*ConsumerGroup),
		}
		p.streams[stream] = s
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Groups[group]; exists {
		return nil // 组已存在，直接返回
	}

	// 确定起始 ID
	lastID := "0"
	if start == "$" {
		// 从当前最后一条消息开始
		if len(s.Messages) > 0 {
			lastID = s.Messages[len(s.Messages)-1].ID
		}
	} else if start != "0" {
		lastID = start
	}

	s.Groups[group] = &ConsumerGroup{
		Name:      group,
		LastID:    lastID,
		Pending:   make(map[string]*StreamMessage),
		Consumers: make(map[string]string),
	}

	return nil
}

func (p *CacheMem) XGroupDestroy(stream, group string) error {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Groups, group)
	return nil
}

func (p *CacheMem) XGroupSetID(stream, group, id string) error {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return errors.New("stream not found")
	}

	s, exists := p.streams[stream]
	if !exists {
		return errors.New("stream not found")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	g, exists := s.Groups[group]
	if !exists {
		return errors.New("group not found")
	}

	g.LastID = id
	return nil
}

func (p *CacheMem) XReadGroup(handler func(stream string, id string, body []byte) error, group, consumer, stream string) error {
	p.streamsMu.Lock()

	if p.streams == nil {
		p.streamsMu.Unlock()
		return errors.New("stream not found")
	}

	s, exists := p.streams[stream]
	if !exists {
		p.streamsMu.Unlock()
		return errors.New("stream not found")
	}

	p.streamsMu.Unlock()

	// 持续消费消息
	for {
		s.mu.Lock()

		g, exists := s.Groups[group]
		if !exists {
			s.mu.Unlock()
			return errors.New("group not found")
		}

		// 查找新消息（ID > LastID）
		var newMsgs []*StreamMessage
		for _, msg := range s.Messages {
			if msg.ID > g.LastID && !msg.Acked {
				newMsgs = append(newMsgs, msg)
			}
		}

		// 如果有新消息，处理它们
		if len(newMsgs) > 0 {
			for _, msg := range newMsgs {
				// 将消息添加到待处理列表
				g.Pending[msg.ID] = msg
				g.Consumers[consumer] = msg.ID
				g.LastID = msg.ID

				s.mu.Unlock()

				// 提取消息体（假设使用单个字段 "data"）
				var body []byte
				if len(msg.Values) == 1 {
					for _, v := range msg.Values {
						body = []byte(candy.ToString(v))
						break
					}
				} else {
					// 如果有多个字段，序列化为 JSON
					jsonData, err := json.Marshal(msg.Values)
					if err != nil {
						log.Errorf("err:%v", err)
						body = []byte("{}")
					} else {
						body = jsonData
					}
				}

				// 调用处理函数
				err := handler(stream, msg.ID, body)
				if err != nil {
					log.Errorf("err:%v", err)
				}

				s.mu.Lock()
			}
		} else {
			s.mu.Unlock()
		}

		s.mu.Unlock()

		// 短暂休眠避免 CPU 占用过高
		time.Sleep(100 * time.Millisecond)
	}
}

func (p *CacheMem) XAck(stream, group string, ids ...string) (int64, error) {
	p.streamsMu.Lock()
	defer p.streamsMu.Unlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	g, exists := s.Groups[group]
	if !exists {
		return 0, nil
	}

	ackedCount := int64(0)
	for _, id := range ids {
		if msg, exists := g.Pending[id]; exists {
			msg.Acked = true
			delete(g.Pending, id)
			ackedCount++
		}
	}

	return ackedCount, nil
}

func (p *CacheMem) XPending(stream, group string) (int64, error) {
	p.streamsMu.RLock()
	defer p.streamsMu.RUnlock()

	if p.streams == nil {
		return 0, nil
	}

	s, exists := p.streams[stream]
	if !exists {
		return 0, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	g, exists := s.Groups[group]
	if !exists {
		return 0, nil
	}

	return int64(len(g.Pending)), nil
}

// ZSet 有序集合数据结构
type ZSet struct {
	mu      sync.RWMutex
	members map[string]float64 // member -> score映射
}

// ZSetMember 有序集合成员（用于排序）
type ZSetMember struct {
	Member string
	Score  float64
}

// newZSet 创建新的ZSet
func newZSet() *ZSet {
	return &ZSet{
		members: make(map[string]float64),
	}
}

// add 添加成员
func (z *ZSet) add(member string, score float64) bool {
	_, exists := z.members[member]
	z.members[member] = score
	return !exists // 返回是否是新成员
}

// remove 删除成员
func (z *ZSet) remove(member string) bool {
	_, exists := z.members[member]
	if exists {
		delete(z.members, member)
	}
	return exists
}

// score 获取成员分数
func (z *ZSet) score(member string) (float64, bool) {
	score, exists := z.members[member]
	return score, exists
}

// card 获取成员数量
func (z *ZSet) card() int {
	return len(z.members)
}

// getSortedMembers 获取排序后的成员列表
func (z *ZSet) getSortedMembers(ascending bool) []ZSetMember {
	members := make([]ZSetMember, 0, len(z.members))
	for member, score := range z.members {
		members = append(members, ZSetMember{
			Member: member,
			Score:  score,
		})
	}

	sort.Slice(members, func(i, j int) bool {
		if members[i].Score != members[j].Score {
			if ascending {
				return members[i].Score < members[j].Score
			}
			return members[i].Score > members[j].Score
		}
		// 分数相同时按字典序排序
		if ascending {
			return members[i].Member < members[j].Member
		}
		return members[i].Member > members[j].Member
	})

	return members
}

// rank 获取成员排名（升序）
func (z *ZSet) rank(member string, ascending bool) (int64, bool) {
	if _, exists := z.members[member]; !exists {
		return -1, false
	}

	sorted := z.getSortedMembers(ascending)
	for i, m := range sorted {
		if m.Member == member {
			return int64(i), true
		}
	}
	return -1, false
}

// CacheMem ZSet 方法实现

// ZAdd 添加成员到有序集合
func (p *CacheMem) ZAdd(key string, members ...interface{}) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	if p.zsets == nil {
		p.zsets = make(map[string]*ZSet)
	}
	if p.zsets[key] == nil {
		p.zsets[key] = newZSet()
	}

	zset := p.zsets[key]
	zset.mu.Lock()
	defer zset.mu.Unlock()

	// 解析members（score, member pairs）
	var count int64
	for i := 0; i < len(members); i += 2 {
		if i+1 >= len(members) {
			break
		}

		var score float64
		switch v := members[i].(type) {
		case float64:
			score = v
		case float32:
			score = float64(v)
		case int:
			score = float64(v)
		case int64:
			score = float64(v)
		case int32:
			score = float64(v)
		default:
			continue
		}

		member := ""
		switch v := members[i+1].(type) {
		case string:
			member = v
		default:
			continue
		}

		if zset.add(member, score) {
			count++
		}
	}

	return count, nil
}

// ZScore 获取成员分数
func (p *CacheMem) ZScore(key, member string) (float64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, ErrNotFound
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	score, exists := zset.score(member)
	if !exists {
		return 0, ErrNotFound
	}

	return score, nil
}

// ZRange 按索引范围获取成员（升序）
func (p *CacheMem) ZRange(key string, start, stop int64) ([]string, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []string{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	sorted := zset.getSortedMembers(true)
	return p.sliceRange(sorted, start, stop), nil
}

// ZRangeByScore 按分数范围获取成员
func (p *CacheMem) ZRangeByScore(key, min, max string, offset, count int64) ([]string, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []string{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	minScore, minInclusive := p.parseScore(min)
	maxScore, maxInclusive := p.parseScore(max)

	sorted := zset.getSortedMembers(true)
	result := make([]string, 0)

	for _, m := range sorted {
		if (minInclusive && m.Score >= minScore) || (!minInclusive && m.Score > minScore) {
			if (maxInclusive && m.Score <= maxScore) || (!maxInclusive && m.Score < maxScore) {
				result = append(result, m.Member)
			}
		}
	}

	// 应用offset和count
	if offset > 0 {
		if offset >= int64(len(result)) {
			return []string{}, nil
		}
		result = result[offset:]
	}
	if count > 0 && count < int64(len(result)) {
		result = result[:count]
	}

	return result, nil
}

// ZRem 删除成员
func (p *CacheMem) ZRem(key string, members ...string) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.Lock()
	defer zset.mu.Unlock()

	var count int64
	for _, member := range members {
		if zset.remove(member) {
			count++
		}
	}

	return count, nil
}

// ZCard 获取集合成员数
func (p *CacheMem) ZCard(key string) (int64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	return int64(zset.card()), nil
}

// ZCount 统计分数范围内成员数
func (p *CacheMem) ZCount(key, min, max string) (int64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	minScore, minInclusive := p.parseScore(min)
	maxScore, maxInclusive := p.parseScore(max)

	var count int64
	for _, score := range zset.members {
		if (minInclusive && score >= minScore) || (!minInclusive && score > minScore) {
			if (maxInclusive && score <= maxScore) || (!maxInclusive && score < maxScore) {
				count++
			}
		}
	}

	return count, nil
}

// ZIncrBy 增加成员分数
func (p *CacheMem) ZIncrBy(key string, increment float64, member string) (float64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	if p.zsets == nil {
		p.zsets = make(map[string]*ZSet)
	}
	if p.zsets[key] == nil {
		p.zsets[key] = newZSet()
	}

	zset := p.zsets[key]
	zset.mu.Lock()
	defer zset.mu.Unlock()

	score, exists := zset.score(member)
	if !exists {
		score = 0
	}

	newScore := score + increment
	zset.add(member, newScore)

	return newScore, nil
}

// ZRank 获取成员排名（升序，从0开始）
func (p *CacheMem) ZRank(key, member string) (int64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return -1, ErrNotFound
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	rank, exists := zset.rank(member, true)
	if !exists {
		return -1, ErrNotFound
	}

	return rank, nil
}

// ZRevRange 按索引范围获取成员（降序）
func (p *CacheMem) ZRevRange(key string, start, stop int64) ([]string, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []string{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	sorted := zset.getSortedMembers(false)
	return p.sliceRange(sorted, start, stop), nil
}

// ZRevRank 获取成员排名（降序，从0开始）
func (p *CacheMem) ZRevRank(key, member string) (int64, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return -1, ErrNotFound
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	rank, exists := zset.rank(member, false)
	if !exists {
		return -1, ErrNotFound
	}

	return rank, nil
}

// ZRangeWithScores 按索引范围获取成员和分数
func (p *CacheMem) ZRangeWithScores(key string, start, stop int64) ([]Z, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []Z{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	sorted := zset.getSortedMembers(true)
	members := p.sliceRangeWithScore(sorted, start, stop)

	result := make([]Z, len(members))
	for i, m := range members {
		result[i] = Z{
			Member: m.Member,
			Score:  m.Score,
		}
	}

	return result, nil
}

// ZRevRangeByScore 按分数范围获取成员（降序）
func (p *CacheMem) ZRevRangeByScore(key, max, min string, offset, count int64) ([]string, error) {
	p.zsetsMu.RLock()
	defer p.zsetsMu.RUnlock()

	zset, exists := p.zsets[key]
	if !exists {
		return []string{}, nil
	}

	zset.mu.RLock()
	defer zset.mu.RUnlock()

	minScore, minInclusive := p.parseScore(min)
	maxScore, maxInclusive := p.parseScore(max)

	sorted := zset.getSortedMembers(false)
	result := make([]string, 0)

	for _, m := range sorted {
		if (maxInclusive && m.Score <= maxScore) || (!maxInclusive && m.Score < maxScore) {
			if (minInclusive && m.Score >= minScore) || (!minInclusive && m.Score > minScore) {
				result = append(result, m.Member)
			}
		}
	}

	// 应用offset和count
	if offset > 0 {
		if offset >= int64(len(result)) {
			return []string{}, nil
		}
		result = result[offset:]
	}
	if count > 0 && count < int64(len(result)) {
		result = result[:count]
	}

	return result, nil
}

// ZRemRangeByRank 按排名范围删除成员
func (p *CacheMem) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.Lock()
	defer zset.mu.Unlock()

	sorted := zset.getSortedMembers(true)
	toRemove := p.sliceRange(sorted, start, stop)

	var count int64
	for _, member := range toRemove {
		if zset.remove(member) {
			count++
		}
	}

	return count, nil
}

// ZRemRangeByScore 按分数范围删除成员
func (p *CacheMem) ZRemRangeByScore(key, min, max string) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	zset, exists := p.zsets[key]
	if !exists {
		return 0, nil
	}

	zset.mu.Lock()
	defer zset.mu.Unlock()

	minScore, minInclusive := p.parseScore(min)
	maxScore, maxInclusive := p.parseScore(max)

	toRemove := make([]string, 0)
	for member, score := range zset.members {
		if (minInclusive && score >= minScore) || (!minInclusive && score > minScore) {
			if (maxInclusive && score <= maxScore) || (!maxInclusive && score < maxScore) {
				toRemove = append(toRemove, member)
			}
		}
	}

	var count int64
	for _, member := range toRemove {
		if zset.remove(member) {
			count++
		}
	}

	return count, nil
}

// ZUnionStore 并集存储
func (p *CacheMem) ZUnionStore(destination string, keys ...string) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	if p.zsets == nil {
		p.zsets = make(map[string]*ZSet)
	}

	// 创建新的ZSet存储结果
	result := newZSet()

	for _, key := range keys {
		zset, exists := p.zsets[key]
		if !exists {
			continue
		}

		zset.mu.RLock()
		for member, score := range zset.members {
			existingScore, exists := result.members[member]
			if !exists {
				result.members[member] = score
			} else {
				// 并集：分数相加
				result.members[member] = existingScore + score
			}
		}
		zset.mu.RUnlock()
	}

	p.zsets[destination] = result
	return int64(len(result.members)), nil
}

// ZInterStore 交集存储
func (p *CacheMem) ZInterStore(destination string, keys ...string) (int64, error) {
	p.zsetsMu.Lock()
	defer p.zsetsMu.Unlock()

	if p.zsets == nil {
		p.zsets = make(map[string]*ZSet)
	}

	if len(keys) == 0 {
		return 0, nil
	}

	// 创建新的ZSet存储结果
	result := newZSet()

	// 获取第一个集合作为基准
	firstZset, exists := p.zsets[keys[0]]
	if !exists {
		p.zsets[destination] = result
		return 0, nil
	}

	firstZset.mu.RLock()
	defer firstZset.mu.RUnlock()

	// 遍历第一个集合的每个成员
	for member, score := range firstZset.members {
		sumScore := score
		inAll := true

		// 检查是否在其他所有集合中
		for i := 1; i < len(keys); i++ {
			zset, exists := p.zsets[keys[i]]
			if !exists {
				inAll = false
				break
			}

			zset.mu.RLock()
			otherScore, exists := zset.members[member]
			zset.mu.RUnlock()

			if !exists {
				inAll = false
				break
			}

			sumScore += otherScore
		}

		if inAll {
			result.members[member] = sumScore
		}
	}

	p.zsets[destination] = result
	return int64(len(result.members)), nil
}

// 辅助方法

// sliceRange 从sorted列表中提取start到stop范围的成员
func (p *CacheMem) sliceRange(sorted []ZSetMember, start, stop int64) []string {
	length := int64(len(sorted))
	if length == 0 {
		return []string{}
	}

	// 处理负索引
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// 边界检查
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []string{}
	}

	result := make([]string, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, sorted[i].Member)
	}

	return result
}

// sliceRangeWithScore 从sorted列表中提取start到stop范围的成员（包含分数）
func (p *CacheMem) sliceRangeWithScore(sorted []ZSetMember, start, stop int64) []ZSetMember {
	length := int64(len(sorted))
	if length == 0 {
		return []ZSetMember{}
	}

	// 处理负索引
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// 边界检查
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []ZSetMember{}
	}

	result := make([]ZSetMember, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, sorted[i])
	}

	return result
}

// parseScore 解析分数字符串（支持-inf, +inf, (score等Redis语法）
func (p *CacheMem) parseScore(scoreStr string) (float64, bool) {
	if scoreStr == "-inf" {
		return -1e308, true // 使用极小值表示负无穷
	}
	if scoreStr == "+inf" {
		return 1e308, true // 使用极大值表示正无穷
	}

	inclusive := true
	if len(scoreStr) > 0 && scoreStr[0] == '(' {
		inclusive = false
		scoreStr = scoreStr[1:]
	}

	var score float64
	_, err := fmt.Sscanf(scoreStr, "%f", &score)
	if err != nil {
		return 0, true
	}

	return score, inclusive
}

