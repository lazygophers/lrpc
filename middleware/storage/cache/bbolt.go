package cache

import (
	"fmt"
	"strconv"
	"time"

	"github.com/beefsack/go-rate"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/app"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/json"
	"go.etcd.io/bbolt"
	"gorm.io/gorm/utils"
)

type CacheBbolt struct {
	conn *bbolt.DB

	rt *rate.RateLimiter

	prefix string
}

// getBucket is a helper method to get or create the bucket
func (p *CacheBbolt) getBucket(tx *bbolt.Tx, write bool) (*bbolt.Bucket, error) {
	if write {
		return tx.CreateBucketIfNotExists([]byte(p.prefix))
	}
	b := tx.Bucket([]byte(p.prefix))
	if b == nil {
		return nil, ErrNotFound
	}
	return b, nil
}

func (p *CacheBbolt) Clean() error {
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b, err := p.getBucket(tx, false)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		err = b.ForEach(func(k, v []byte) error {
			err := b.Delete(k)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
			return nil
		})
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheBbolt) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *CacheBbolt) IncrBy(key string, value int64) (int64, error) {
	p.autoClear()

	var result int64
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b, err := p.getBucket(tx, true)
		if err != nil {
			return err
		}
		v := b.Get([]byte(key))

		var current int64
		if v != nil {
			var item Item
			err := json.Unmarshal(v, &item)
			if err == nil {
				if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
					current = 0 // Expired, treat as 0
				} else {
					current, _ = strconv.ParseInt(item.Data, 10, 64)
				}
			} else {
				log.Errorf("err:%v", err)
			}
		}

		result = current + value
		item := &Item{
			Data: strconv.FormatInt(result, 10),
		}

		return b.Put([]byte(key), item.Bytes())
	})

	return result, err
}

func (p *CacheBbolt) DecrBy(key string, value int64) (int64, error) {
	result, err := p.IncrBy(key, -value)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return result, nil
}

func (p *CacheBbolt) Expire(key string, timeout time.Duration) (bool, error) {
	var found bool
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		v := b.Get([]byte(key))
		if v == nil {
			return nil
		}

		found = true
		var item Item
		err := json.Unmarshal(v, &item)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		item.ExpireAt = time.Now().Add(timeout)
		err = b.Put([]byte(key), item.Bytes())
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		return nil
	})

	return found, err
}

func (p *CacheBbolt) Ttl(key string) (time.Duration, error) {
	var ttl time.Duration = -2 * time.Second // Key not found
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		v := b.Get([]byte(key))
		if v == nil {
			return nil
		}

		var item Item
		err := json.Unmarshal(v, &item)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if item.ExpireAt.IsZero() {
			ttl = -1 * time.Second // No expiration
		} else if time.Now().After(item.ExpireAt) {
			ttl = -2 * time.Second // Expired
		} else {
			ttl = item.ExpireAt.Sub(time.Now())
		}
		return nil
	})

	return ttl, err
}

func (p *CacheBbolt) Incr(key string) (int64, error) {
	return p.IncrBy(key, 1)
}

func (p *CacheBbolt) Decr(key string) (int64, error) {
	return p.IncrBy(key, -1)
}

func (p *CacheBbolt) Exists(keys ...string) (bool, error) {
	p.autoClear()

	var allExist bool = true
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		for _, key := range keys {
			v := b.Get([]byte(key))
			if v == nil {
				allExist = false
				return nil
			}

			var item Item
			err := json.Unmarshal(v, &item)
			if err == nil {
				// Only process if JSON is valid
				if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
					allExist = false
					return nil
				}
			} else {
				// Treat invalid JSON as non-existent (graceful handling)
				log.Errorf("err:%v", err)
				allExist = false
				return nil
			}
		}
		return nil
	})

	return allExist, err
}

func (p *CacheBbolt) HIncr(key string, subKey string) (int64, error) {
	return p.HIncrBy(key, subKey, 1)
}

func (p *CacheBbolt) HIncrBy(key string, field string, increment int64) (int64, error) {
	p.autoClear()

	var result int64
	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		v := b.Get([]byte(hashKey))

		var current int64
		if v != nil {
			var item Item
			err := json.Unmarshal(v, &item)
			if err == nil {
				if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
					current = 0
				} else {
					current, _ = strconv.ParseInt(item.Data, 10, 64)
				}
			} else {
				log.Errorf("err:%v", err)
			}
		}

		result = current + increment
		item := &Item{
			Data: strconv.FormatInt(result, 10),
		}

		return b.Put([]byte(hashKey), item.Bytes())
	})

	return result, err
}

func (p *CacheBbolt) HDecr(key string, field string) (int64, error) {
	return p.HIncrBy(key, field, -1)
}

func (p *CacheBbolt) HDecrBy(key string, field string, increment int64) (int64, error) {
	return p.HIncrBy(key, field, -increment)
}

func (p *CacheBbolt) SAdd(key string, members ...string) (int64, error) {
	p.autoClear()

	var added int64
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))

		// Get existing set
		setKey := fmt.Sprintf("%s:set", key)
		v := b.Get([]byte(setKey))
		var existingSet map[string]bool
		if v != nil {
			var item Item
			err := json.Unmarshal(v, &item)
			if err == nil {
				if item.ExpireAt.IsZero() || time.Now().Before(item.ExpireAt) {
					err = json.UnmarshalString(item.Data, &existingSet)
					if err != nil {
						log.Errorf("err:%v", err)
					}
				}
			} else {
				log.Errorf("err:%v", err)
			}
		}

		if existingSet == nil {
			existingSet = make(map[string]bool)
		}

		// Add new members
		for _, member := range members {
			if !existingSet[member] {
				existingSet[member] = true
				added++
			}
		}

		// Save updated set
		data, err := json.MarshalString(existingSet)
		if err != nil {
			return err
		}
		item := &Item{
			Data: data,
		}

		return b.Put([]byte(setKey), item.Bytes())
	})

	return added, err
}

func (p *CacheBbolt) SMembers(key string) ([]string, error) {
	p.autoClear()

	var members []string
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		setKey := fmt.Sprintf("%s:set", key)
		v := b.Get([]byte(setKey))
		if v == nil {
			return nil
		}

		var item Item
		err := json.Unmarshal(v, &item)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
			return nil
		}

		var setData map[string]bool
		err = json.UnmarshalString(item.Data, &setData)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		for member := range setData {
			members = append(members, member)
		}

		return nil
	})

	return members, err
}

func (p *CacheBbolt) SRem(key string, members ...string) (int64, error) {
	p.autoClear()

	var removed int64
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		setKey := fmt.Sprintf("%s:set", key)
		v := b.Get([]byte(setKey))
		if v == nil {
			return nil
		}

		var item Item
		err := json.Unmarshal(v, &item)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
			return nil
		}

		var setData map[string]bool
		err = json.UnmarshalString(item.Data, &setData)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		for _, member := range members {
			if setData[member] {
				delete(setData, member)
				removed++
			}
		}

		if len(setData) == 0 {
			return b.Delete([]byte(setKey))
		}

		data, err := json.MarshalString(setData)
		if err != nil {
			return err
		}
		item.Data = data
		return b.Put([]byte(setKey), item.Bytes())
	})

	return removed, err
}

func (p *CacheBbolt) SRandMember(key string, count ...int64) ([]string, error) {
	members, err := p.SMembers(key)
	if err != nil || len(members) == 0 {
		return members, err
	}

	n := int64(1)
	if len(count) > 0 && count[0] > 0 {
		n = count[0]
	}

	if n >= int64(len(members)) {
		return members, nil
	}

	// Simple random selection (not cryptographically secure)
	result := make([]string, 0, n)
	used := make(map[int]bool)
	for int64(len(result)) < n {
		idx := int(time.Now().UnixNano()) % len(members)
		if !used[idx] {
			used[idx] = true
			result = append(result, members[idx])
		}
	}

	return result, nil
}

func (p *CacheBbolt) SPop(key string) (string, error) {
	members, err := p.SMembers(key)
	if err != nil || len(members) == 0 {
		return "", err
	}

	// Pop the first member (could be randomized)
	member := members[0]
	_, err = p.SRem(key, member)
	if err != nil {
		return "", err
	}
	return member, nil
}

func (p *CacheBbolt) SisMember(key, field string) (bool, error) {
	p.autoClear()

	var isMember bool
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		setKey := fmt.Sprintf("%s:set", key)
		v := b.Get([]byte(setKey))
		if v == nil {
			return nil
		}

		var item Item
		err := json.Unmarshal(v, &item)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
			return nil
		}

		var setData map[string]bool
		err = json.UnmarshalString(item.Data, &setData)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		isMember = setData[field]
		return nil
	})

	return isMember, err
}

func (p *CacheBbolt) HExists(key string, field string) (bool, error) {
	_, err := p.HGet(key, field)
	if err != nil {
		if err == ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *CacheBbolt) HKeys(key string) ([]string, error) {
	p.autoClear()

	var keys []string
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		prefix := fmt.Sprintf("%s:hash:", key)

		c := b.Cursor()
		for k, v := c.Seek([]byte(prefix)); k != nil && string(k[:len(prefix)]) == prefix; k, v = c.Next() {
			var item Item
			err := json.Unmarshal(v, &item)
			if err != nil {
				log.Errorf("err:%v", err)
				continue
			}

			if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
				continue
			}

			// Extract field name from key
			fieldName := string(k[len(prefix):])
			keys = append(keys, fieldName)
		}
		return nil
	})

	return keys, err
}

func (p *CacheBbolt) HSet(key string, field string, value interface{}) (bool, error) {
	p.autoClear()

	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	var isNew bool

	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		v := b.Get([]byte(hashKey))
		isNew = v == nil

		item := &Item{
			Data: candy.ToString(value),
		}

		return b.Put([]byte(hashKey), item.Bytes())
	})

	return isNew, err
}

func (p *CacheBbolt) HGet(key, field string) (string, error) {
	p.autoClear()

	var value string
	hashKey := fmt.Sprintf("%s:hash:%s", key, field)
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		v := b.Get([]byte(hashKey))
		if v == nil {
			return ErrNotFound
		}

		var item Item
		err := json.Unmarshal(v, &item)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
			return ErrNotFound
		}

		value = item.Data
		return nil
	})

	return value, err
}

func (p *CacheBbolt) HDel(key string, fields ...string) (int64, error) {
	p.autoClear()

	var deleted int64
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))

		for _, field := range fields {
			hashKey := fmt.Sprintf("%s:hash:%s", key, field)
			if b.Get([]byte(hashKey)) != nil {
				err := b.Delete([]byte(hashKey))
				if err != nil {
					log.Errorf("err:%v", err)
					return err
				}
				deleted++
			}
		}

		return nil
	})

	return deleted, err
}

func (p *CacheBbolt) HGetAll(key string) (map[string]string, error) {
	p.autoClear()

	result := make(map[string]string)
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		prefix := fmt.Sprintf("%s:hash:", key)

		c := b.Cursor()
		for k, v := c.Seek([]byte(prefix)); k != nil && string(k[:len(prefix)]) == prefix; k, v = c.Next() {
			var item Item
			err := json.Unmarshal(v, &item)
			if err != nil {
				log.Errorf("err:%v", err)
				continue
			}

			if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
				continue
			}

			// Extract field name from key
			fieldName := string(k[len(prefix):])
			result[fieldName] = item.Data
		}
		return nil
	})

	return result, err
}

func (p *CacheBbolt) autoClear() {
	ok, _ := p.rt.Try()
	if !ok {
		return
	}

	p.clear()
}

func (p *CacheBbolt) clear() {
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))

		var item Item
		return b.ForEach(func(k, v []byte) error {
			err := json.Unmarshal(v, &item)
			if err != nil {
				return err
			}

			if item.ExpireAt.IsZero() {
				return nil
			}

			if time.Now().After(item.ExpireAt) {
				return b.Delete(k)
			}

			return nil
		})
	})
	if err != nil {
		log.Errorf("err:%v", err)
	}
}

func (p *CacheBbolt) SetEx(key string, value any, timeout time.Duration) error {
	item := &Item{
		Data:     utils.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}

	return p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		return b.Put([]byte(key), item.Bytes())
	})
}

func (p *CacheBbolt) Get(key string) (string, error) {
	p.autoClear()

	var value string
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b, err := p.getBucket(tx, false)
		if err != nil {
			return err
		}
		v := b.Get([]byte(key))
		if v == nil {
			return ErrNotFound
		}

		var item Item
		err = json.Unmarshal(v, &item)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
			return ErrNotFound
		}

		value = item.Data
		return nil
	})

	return value, err
}

func (p *CacheBbolt) Set(key string, value any) error {
	item := &Item{
		Data: utils.ToString(value),
	}

	return p.conn.Update(func(tx *bbolt.Tx) error {
		b, err := p.getBucket(tx, true)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), item.Bytes())
	})
}

func (p *CacheBbolt) SetNx(key string, value interface{}) (bool, error) {
	item := &Item{
		Data: utils.ToString(value),
	}

	var ok bool
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))

		value := b.Get([]byte(key))
		if value != nil {
			return nil
		}

		ok = true

		return b.Put([]byte(key), item.Bytes())
	})
	return ok, err
}

func (p *CacheBbolt) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	item := &Item{
		Data:     utils.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}

	var ok bool
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))

		value := b.Get([]byte(key))
		if value != nil {
			return nil
		}

		ok = true

		return b.Put([]byte(key), item.Bytes())
	})
	return ok, err
}

func (p *CacheBbolt) Del(key ...string) error {
	return p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		for _, k := range key {
			err := b.Delete([]byte(k))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *CacheBbolt) Close() error {
	return p.conn.Close()
}

func (p *CacheBbolt) Reset() error {
	return p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		return b.ForEach(func(k, v []byte) error {
			return b.Delete(k)
		})
	})
}

func (p *CacheBbolt) Ping() error {
	err := p.conn.View(func(tx *bbolt.Tx) error {
		return nil
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func NewBbolt(addr string, options *bbolt.Options) (Cache, error) {
	prefix := app.Name
	if prefix == "" {
		prefix = "lrpc_cache"
	}
	p := &CacheBbolt{
		rt:     rate.New(2, time.Minute*10),
		prefix: prefix,
	}

	conn, err := bbolt.Open(addr, 0o666, options)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	err = conn.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(p.prefix))
		return err
	})

	p.conn = conn

	atexit.Register(func() {
		err := p.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	return newBaseCache(p), nil
}
