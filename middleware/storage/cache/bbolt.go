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

var (
	bboltBucket = []byte(app.Name)
)

type Bbolt struct {
	conn *bbolt.DB
	rt   *rate.RateLimiter
}

func (p *Bbolt) SetPrefix(prefix string) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) IncrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) DecrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) Expire(key string, timeout time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) Ttl(key string) (time.Duration, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) Incr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) Decr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HIncr(key string, subKey string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HIncrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HExists(key string, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HKeys(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HSet(key string, field string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HGet(key, field string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HDel(key string, fields ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bbolt) autoClear() {
	ok, _ := p.rt.Try()
	if !ok {
		return
	}

	p.clear()
}

func (p *Bbolt) clear() {
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bboltBucket)

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

func (p *Bbolt) SetEx(key string, value any, timeout time.Duration) error {
	item := &Item{
		Data:     utils.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}

	return p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bboltBucket)
		return b.Put([]byte(key), item.Bytes())
	})
}

func (p *Bbolt) Get(key string) (string, error) {
	var value string
	err := p.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bboltBucket)
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

func (p *Bbolt) Set(key string, value any) error {
	item := &Item{
		Data: utils.ToString(value),
	}

	return p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bboltBucket)
		return b.Put([]byte(key), item.Bytes())
	})
}

func (p *Bbolt) SetNx(key string, value interface{}) (bool, error) {
	item := &Item{
		Data: utils.ToString(value),
	}

	var ok bool
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bboltBucket)

		value := b.Get([]byte(key))
		if value != nil {
			return nil
		}

		ok = true

		return b.Put([]byte(key), item.Bytes())
	})
	return ok, err
}

func (p *Bbolt) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	item := &Item{
		Data:     utils.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}

	var ok bool
	err := p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bboltBucket)

		value := b.Get([]byte(key))
		if value != nil {
			return nil
		}

		ok = true

		return b.Put([]byte(key), item.Bytes())
	})
	return ok, err
}

func (p *Bbolt) Del(key ...string) error {
	return p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bboltBucket)
		for _, k := range key {
			err := b.Delete([]byte(k))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *Bbolt) Close() error {
	return p.conn.Close()
}

func (p *Bbolt) Reset() error {
	return p.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bboltBucket)
		return b.ForEach(func(k, v []byte) error {
			return b.Delete(k)
		})
	})
}

func NewBbolt(addr string, options *bbolt.Options) (Cache, error) {
	p := &Bbolt{
		rt: rate.New(2, time.Minute*10),
	}

	conn, err := bbolt.Open(addr, 0666, options)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	err = conn.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bboltBucket)
		return err
	})

	p.conn = conn

	return newBaseCache(p), nil
}
