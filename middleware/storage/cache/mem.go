package cache

import (
	"encoding/json"
	"math/rand"
	"strconv"

	"github.com/beefsack/go-rate"
	"gorm.io/gorm/utils"

	"sync"
	"time"
)

type CacheMem struct {
	sync.RWMutex

	data map[string]*Item
	rt   *rate.RateLimiter
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
	hashMap[field] = utils.ToString(value)

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
	ok, _ := p.rt.Try()
	if !ok {
		return
	}

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
		data: make(map[string]*Item),
		rt:   rate.New(2, time.Minute),
	}

	return newBaseCache(p)
}
