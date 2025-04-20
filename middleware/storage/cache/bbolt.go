package cache

import (
	"github.com/beefsack/go-rate"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/app"
	"github.com/lazygophers/utils/json"
	"go.etcd.io/bbolt"
	"gorm.io/gorm/utils"
	"time"
)

type CacheBbolt struct {
	conn *bbolt.DB

	rt *rate.RateLimiter

	prefix string
}

func (p *CacheBbolt) Clean() error {
	return p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		return b.ForEach(func(k, v []byte) error {
			return b.Delete(k)
		})
	})
}

func (p *CacheBbolt) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *CacheBbolt) IncrBy(key string, value int64) (int64, error) {

	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) DecrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) Expire(key string, timeout time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) Ttl(key string) (time.Duration, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) Incr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) Decr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HIncr(key string, subKey string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HIncrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HExists(key string, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HKeys(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HSet(key string, field string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HGet(key, field string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HDel(key string, fields ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBbolt) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
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
	var value string
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(p.prefix))
		v := b.Get([]byte(key))
		if v == nil {
			return NotFound
		}

		var item Item
		err := json.Unmarshal(v, &value)
		if err != nil {
			log.Error(err)
			return err
		}

		if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
			return NotFound
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
		b := tx.Bucket([]byte(p.prefix))
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

func NewBbolt(addr string, options *bbolt.Options) (Cache, error) {
	p := &CacheBbolt{
		rt:     rate.New(2, time.Minute*10),
		prefix: app.Name,
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

	return newBaseCache(p), nil
}
