package cache

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/json"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type CacheLevelDB struct {
	prefix string

	db *leveldb.DB
}

// getItem is a helper method to get an item and check if it's expired
func (p *CacheLevelDB) getItem(key string) (*Item, error) {
	value, err := p.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
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
		p.db.Delete([]byte(key), nil)
		return nil, ErrNotFound
	}

	return &item, nil
}

func (p *CacheLevelDB) Clean() error {
	iter := p.db.NewIterator(nil, nil)
	defer iter.Release()

	batch := new(leveldb.Batch)
	for iter.Next() {
		batch.Delete(iter.Key())
	}

	err := p.db.Write(batch, nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return iter.Error()
}

func (p *CacheLevelDB) Get(key string) (string, error) {
	item, err := p.getItem(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return item.Data, nil
}

func (p *CacheLevelDB) Set(key string, value any) error {
	item := &Item{
		Data: candy.ToString(value),
	}
	err := p.db.Put([]byte(key), item.Bytes(), nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheLevelDB) SetEx(key string, value any, timeout time.Duration) error {
	item := &Item{
		Data:     candy.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}
	err := p.db.Put([]byte(key), item.Bytes(), nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheLevelDB) SetNx(key string, value interface{}) (bool, error) {
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

func (p *CacheLevelDB) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
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

func (p *CacheLevelDB) Ttl(key string) (time.Duration, error) {
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

func (p *CacheLevelDB) Expire(key string, timeout time.Duration) (bool, error) {
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
	err = p.db.Put([]byte(key), item.Bytes(), nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return true, nil
}

func (p *CacheLevelDB) Incr(key string) (int64, error) {
	val, err := p.IncrBy(key, 1)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheLevelDB) Decr(key string) (int64, error) {
	val, err := p.IncrBy(key, -1)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheLevelDB) IncrBy(key string, value int64) (int64, error) {
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

func (p *CacheLevelDB) DecrBy(key string, value int64) (int64, error) {
	val, err := p.IncrBy(key, -value)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheLevelDB) Exists(keys ...string) (bool, error) {
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

func (p *CacheLevelDB) HSet(key string, field string, value interface{}) (bool, error) {
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

func (p *CacheLevelDB) HGet(key, field string) (string, error) {
	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	value, err := p.Get(hashKey)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (p *CacheLevelDB) HDel(key string, fields ...string) (int64, error) {
	var deleted int64
	for _, field := range fields {
		hashKey := fmt.Sprintf("%s:hash:%s", key, field)
		_, err := p.getItem(hashKey)
		if err == nil {
			err = p.db.Delete([]byte(hashKey), nil)
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

func (p *CacheLevelDB) HKeys(key string) ([]string, error) {
	prefix := fmt.Sprintf("%s:hash:", key)
	var keys []string

	iter := p.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	for iter.Next() {
		keyStr := string(iter.Key())
		if len(keyStr) > len(prefix) && keyStr[:len(prefix)] == prefix {
			// Extract field name
			field := keyStr[len(prefix):]

			// Check if the key is not expired
			_, err := p.getItem(keyStr)
			if err == nil {
				keys = append(keys, field)
			}
		}
	}

	err := iter.Error()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return keys, nil
}

func (p *CacheLevelDB) HGetAll(key string) (map[string]string, error) {
	prefix := fmt.Sprintf("%s:hash:", key)
	result := make(map[string]string)

	iter := p.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	for iter.Next() {
		keyStr := string(iter.Key())
		if len(keyStr) > len(prefix) && keyStr[:len(prefix)] == prefix {
			// Extract field name
			field := keyStr[len(prefix):]

			// Get value and check expiration
			item, err := p.getItem(keyStr)
			if err == nil {
				result[field] = item.Data
			}
		}
	}

	err := iter.Error()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return result, nil
}

func (p *CacheLevelDB) HExists(key string, field string) (bool, error) {
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

func (p *CacheLevelDB) HIncr(key string, subKey string) (int64, error) {
	val, err := p.HIncrBy(key, subKey, 1)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheLevelDB) HIncrBy(key string, field string, increment int64) (int64, error) {
	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	val, err := p.IncrBy(hashKey, increment)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheLevelDB) HDecr(key string, field string) (int64, error) {
	val, err := p.HIncrBy(key, field, -1)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheLevelDB) HDecrBy(key string, field string, increment int64) (int64, error) {
	val, err := p.HIncrBy(key, field, -increment)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (p *CacheLevelDB) SAdd(key string, members ...string) (int64, error) {
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

func (p *CacheLevelDB) SMembers(key string) ([]string, error) {
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

func (p *CacheLevelDB) SRem(key string, members ...string) (int64, error) {
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

func (p *CacheLevelDB) SRandMember(key string, count ...int64) ([]string, error) {
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

func (p *CacheLevelDB) SPop(key string) (string, error) {
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

func (p *CacheLevelDB) SisMember(key, field string) (bool, error) {
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

func (p *CacheLevelDB) Del(key ...string) error {
	batch := new(leveldb.Batch)
	for _, k := range key {
		batch.Delete([]byte(k))
	}
	err := p.db.Write(batch, nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheLevelDB) Close() error {
	err := p.db.Close()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheLevelDB) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *CacheLevelDB) Ping() error {
	_, err := p.db.GetProperty("leveldb.stats")
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheLevelDB) Publish(channel string, message interface{}) (int64, error) {
	return 0, errors.New("leveldb cache does not support pub/sub")
}

func (p *CacheLevelDB) Subscribe(handler func(channel string, message []byte) error, channels ...string) error {
	return errors.New("leveldb cache does not support pub/sub")
}

func (p *CacheLevelDB) XAdd(stream string, values map[string]interface{}) (string, error) {
	return "", errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XLen(stream string) (int64, error) {
	return 0, errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	return nil, errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XRevRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	return nil, errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XDel(stream string, ids ...string) (int64, error) {
	return 0, errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XTrim(stream string, maxLen int64) (int64, error) {
	return 0, errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XGroupCreate(stream, group, start string) error {
	return errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XGroupDestroy(stream, group string) error {
	return errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XGroupSetID(stream, group, id string) error {
	return errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XReadGroup(handler func(stream string, id string, body []byte) error, group, consumer, stream string) error {
	return errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XAck(stream, group string, ids ...string) (int64, error) {
	return 0, errors.New("leveldb cache does not support stream")
}

func (p *CacheLevelDB) XPending(stream, group string) (int64, error) {
	return 0, errors.New("leveldb cache does not support stream")
}

func NewLevelDB(c *Config) (Cache, error) {
	var err error
	p := &CacheLevelDB{}

	p.db, err = leveldb.OpenFile(c.DataDir, &opt.Options{
		ErrorIfMissing: false,
		Compression:    opt.SnappyCompression,
	})
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
