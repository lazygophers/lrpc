package cache

import (
	"encoding/json"
	"strconv"
	"time"

	"gorm.io/gorm/utils"
)

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
