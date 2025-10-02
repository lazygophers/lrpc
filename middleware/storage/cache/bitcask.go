package cache

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/json"
	"go.mills.io/bitcask/v2"
)

type CacheBitcask struct {
	prefix string

	cli *bitcask.Bitcask
}

// getItem is a helper method to get an item and check if it's expired
func (p *CacheBitcask) getItem(key string) (*Item, error) {
	value, err := p.cli.Get(bitcask.Key(key))
	if err != nil {
		if err == bitcask.ErrKeyNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var item Item
	err = json.Unmarshal(value, &item)
	if err != nil {
		// Try to treat as plain string for backward compatibility
		item = Item{
			Data: string(value),
		}
	}

	// Check if expired
	if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
		// Delete expired key
		p.cli.Delete(bitcask.Key(key))
		return nil, ErrNotFound
	}

	return &item, nil
}

func (p *CacheBitcask) Clean() error {
	keys := make([]string, 0)
	
	err := p.cli.Scan(nil, func(k bitcask.Key) error {
		keys = append(keys, string(k))
		return nil
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	
	for _, key := range keys {
		err = p.cli.Delete(bitcask.Key(key))
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	
	return nil
}

func (p *CacheBitcask) Get(key string) (string, error) {
	item, err := p.getItem(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return item.Data, nil
}

func (p *CacheBitcask) Set(key string, value any) error {
	item := &Item{
		Data: candy.ToString(value),
	}
	err := p.cli.Put(bitcask.Key(key), bitcask.Value(item.Bytes()))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheBitcask) SetEx(key string, value any, timeout time.Duration) error {
	item := &Item{
		Data:     candy.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}
	err := p.cli.Put(bitcask.Key(key), bitcask.Value(item.Bytes()))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheBitcask) SetNx(key string, value interface{}) (bool, error) {
	_, err := p.getItem(key)
	if err == nil {
		// Key exists
		return false, nil
	}
	if err != ErrNotFound {
		log.Errorf("err:%v", err)
		return false, err
	}
	
	// Key doesn't exist, set it
	err = p.Set(key, value)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return true, nil
}

func (p *CacheBitcask) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	_, err := p.getItem(key)
	if err == nil {
		// Key exists
		return false, nil
	}
	if err != ErrNotFound {
		log.Errorf("err:%v", err)
		return false, err
	}
	
	// Key doesn't exist, set it with timeout
	err = p.SetEx(key, value, timeout)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return true, nil
}

func (p *CacheBitcask) Ttl(key string) (time.Duration, error) {
	item, err := p.getItem(key)
	if err != nil {
		if err == ErrNotFound {
			return -2 * time.Second, nil // Key doesn't exist
		}
		log.Errorf("err:%v", err)
		return 0, err
	}

	if item.ExpireAt.IsZero() {
		return -1 * time.Second, nil // No expiration
	}

	remaining := time.Until(item.ExpireAt)
	if remaining <= 0 {
		return -2 * time.Second, nil // Already expired
	}

	return remaining, nil
}

func (p *CacheBitcask) Expire(key string, timeout time.Duration) (bool, error) {
	item, err := p.getItem(key)
	if err != nil {
		if err == ErrNotFound {
			return false, nil
		}
		log.Errorf("err:%v", err)
		return false, err
	}

	// Update expiration
	item.ExpireAt = time.Now().Add(timeout)
	err = p.cli.Put(bitcask.Key(key), bitcask.Value(item.Bytes()))
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return true, nil
}

func (p *CacheBitcask) Incr(key string) (int64, error) {
	val, err := p.IncrBy(key, 1)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheBitcask) Decr(key string) (int64, error) {
	val, err := p.IncrBy(key, -1)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheBitcask) IncrBy(key string, value int64) (int64, error) {
	item, err := p.getItem(key)
	var current int64
	if err != nil {
		if err != ErrNotFound {
			log.Errorf("err:%v", err)
			return 0, err
		}
		// Key doesn't exist, start from 0
		current = 0
	} else {
		current, _ = strconv.ParseInt(item.Data, 10, 64)
	}

	result := current + value
	err = p.Set(key, strconv.FormatInt(result, 10))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return result, nil
}

func (p *CacheBitcask) DecrBy(key string, value int64) (int64, error) {
	val, err := p.IncrBy(key, -value)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheBitcask) Exists(keys ...string) (bool, error) {
	for _, key := range keys {
		_, err := p.getItem(key)
		if err != nil {
			if err == ErrNotFound {
				return false, nil
			}
			log.Errorf("err:%v", err)
			return false, err
		}
	}
	return true, nil
}

func (p *CacheBitcask) HSet(key string, field string, value interface{}) (bool, error) {
	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	_, err := p.getItem(hashKey)
	isNew := (err == ErrNotFound)
	
	err = p.Set(hashKey, value)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return isNew, nil
}

func (p *CacheBitcask) HGet(key, field string) (string, error) {
	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	value, err := p.Get(hashKey)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (p *CacheBitcask) HDel(key string, fields ...string) (int64, error) {
	var deleted int64
	for _, field := range fields {
		hashKey := fmt.Sprintf("%s:hash:%s", key, field)
		_, err := p.getItem(hashKey)
		if err == nil {
			err = p.cli.Delete(bitcask.Key(hashKey))
			if err != nil {
				log.Errorf("err:%v", err)
				return deleted, err
			}
			deleted++
		} else if err != ErrNotFound {
			log.Errorf("err:%v", err)
			return deleted, err
		}
	}
	return deleted, nil
}

func (p *CacheBitcask) HKeys(key string) ([]string, error) {
	prefix := fmt.Sprintf("%s:hash:", key)
	var keys []string

	err := p.cli.Scan(bitcask.Key(prefix), func(k bitcask.Key) error {
		keyStr := string(k)
		if len(keyStr) > len(prefix) && keyStr[:len(prefix)] == prefix {
			// Extract field name
			field := keyStr[len(prefix):]
			
			// Check if the key is not expired
			_, err := p.getItem(keyStr)
			if err == nil {
				keys = append(keys, field)
			}
		}
		return nil
	})
	
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return keys, nil
}

func (p *CacheBitcask) HGetAll(key string) (map[string]string, error) {
	prefix := fmt.Sprintf("%s:hash:", key)
	result := make(map[string]string)

	err := p.cli.Scan(bitcask.Key(prefix), func(k bitcask.Key) error {
		keyStr := string(k)
		if len(keyStr) > len(prefix) && keyStr[:len(prefix)] == prefix {
			// Extract field name
			field := keyStr[len(prefix):]
			
			// Get value and check expiration
			item, err := p.getItem(keyStr)
			if err == nil {
				result[field] = item.Data
			}
		}
		return nil
	})

	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return result, nil
}

func (p *CacheBitcask) HExists(key string, field string) (bool, error) {
	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	_, err := p.getItem(hashKey)
	if err != nil {
		if err == ErrNotFound {
			return false, nil
		}
		log.Errorf("err:%v", err)
		return false, err
	}
	return true, nil
}

func (p *CacheBitcask) HIncr(key string, subKey string) (int64, error) {
	val, err := p.HIncrBy(key, subKey, 1)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheBitcask) HIncrBy(key string, field string, increment int64) (int64, error) {
	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	val, err := p.IncrBy(hashKey, increment)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheBitcask) HDecr(key string, field string) (int64, error) {
	val, err := p.HIncrBy(key, field, -1)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheBitcask) HDecrBy(key string, field string, increment int64) (int64, error) {
	val, err := p.HIncrBy(key, field, -increment)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheBitcask) SAdd(key string, members ...string) (int64, error) {
	setKey := fmt.Sprintf("%s:set", key)
	
	// Get existing set
	var existingSet map[string]bool
	item, err := p.getItem(setKey)
	if err == nil {
		json.UnmarshalString(item.Data, &existingSet)
	} else if err != ErrNotFound {
		log.Errorf("err:%v", err)
		return 0, err
	}
	if existingSet == nil {
		existingSet = make(map[string]bool)
	}

	// Add new members
	var added int64
	for _, member := range members {
		if !existingSet[member] {
			existingSet[member] = true
			added++
		}
	}

	// Save updated set
	data, _ := json.MarshalString(existingSet)
	err = p.Set(setKey, data)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return added, nil
}

func (p *CacheBitcask) SMembers(key string) ([]string, error) {
	setKey := fmt.Sprintf("%s:set", key)
	item, err := p.getItem(setKey)
	if err != nil {
		if err == ErrNotFound {
			return []string{}, nil
		}
		log.Errorf("err:%v", err)
		return nil, err
	}

	var setData map[string]bool
	err = json.UnmarshalString(item.Data, &setData)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	var members []string
	for member := range setData {
		members = append(members, member)
	}
	return members, nil
}

func (p *CacheBitcask) SRem(key string, members ...string) (int64, error) {
	setKey := fmt.Sprintf("%s:set", key)
	item, err := p.getItem(setKey)
	if err != nil {
		if err == ErrNotFound {
			return 0, nil
		}
		log.Errorf("err:%v", err)
		return 0, err
	}

	var setData map[string]bool
	err = json.UnmarshalString(item.Data, &setData)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	var removed int64
	for _, member := range members {
		if setData[member] {
			delete(setData, member)
			removed++
		}
	}

	// Save updated set
	data, _ := json.MarshalString(setData)
	err = p.Set(setKey, data)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return removed, nil
}

func (p *CacheBitcask) SRandMember(key string, count ...int64) ([]string, error) {
	members, err := p.SMembers(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if len(members) == 0 {
		return []string{}, nil
	}

	var resultCount int64 = 1
	if len(count) > 0 && count[0] > 0 {
		resultCount = count[0]
	}

	if resultCount >= int64(len(members)) {
		return members, nil
	}

	// Shuffle and return random members
	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	return members[:resultCount], nil
}

func (p *CacheBitcask) SPop(key string) (string, error) {
	members, err := p.SMembers(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}

	if len(members) == 0 {
		return "", nil
	}

	// Get random member
	randomIndex := rand.Intn(len(members))
	member := members[randomIndex]

	// Remove it from set
	_, err = p.SRem(key, member)
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}

	return member, nil
}

func (p *CacheBitcask) SisMember(key, field string) (bool, error) {
	members, err := p.SMembers(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	for _, member := range members {
		if member == field {
			return true, nil
		}
	}
	return false, nil
}

func (p *CacheBitcask) Del(key ...string) error {
	for _, k := range key {
		err := p.cli.Delete(bitcask.Key(k))
		if err != nil && err != bitcask.ErrKeyNotFound {
			log.Errorf("err:%v", err)
			return err
		}
	}
	return nil
}

func (p *CacheBitcask) Close() error {
	err := p.cli.Close()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheBitcask) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *CacheBitcask) Ping() error {
	err := p.cli.Sync()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func NewBitcask(c *Config) (Cache, error) {
	var err error
	p := &CacheBitcask{}

	p.cli, err = bitcask.Open(c.DataDir, bitcask.WithAutoRecovery(true), bitcask.WithSyncWrites(false))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	atexit.Register(func() {
		err := p.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	return newBaseCache(p), nil
}