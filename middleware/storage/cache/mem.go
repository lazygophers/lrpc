package cache

import (
	"github.com/beefsack/go-rate"
	"gorm.io/gorm/utils"

	"sync"
	"time"
)

type Mem struct {
	sync.RWMutex

	data map[string]*Item
	rt   *rate.RateLimiter
}

func (p *Mem) IncrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) DecrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) Expire(key string, timeout time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) Ttl(key string) (time.Duration, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) Incr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) Decr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HIncr(key string, subKey string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HIncrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HExists(key string, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HKeys(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HSet(key string, field string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HGet(key, field string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HDel(key string, fields ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Mem) SetEx(key string, value any, timeout time.Duration) error {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	p.data[key] = &Item{
		Data:     utils.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}

	return nil
}

func (p *Mem) autoClear() {
	ok, _ := p.rt.Try()
	if !ok {
		return
	}

	p.clear()
}

func (p *Mem) clear() {
	p.Lock()
	defer p.Unlock()

	data := make(map[string]*Item)

	for k, v := range p.data {
		if v.ExpireAt.IsZero() {
			data[k] = v
			continue
		}

		if time.Now().After(v.ExpireAt) {
			continue
		}
	}

	p.data = data
}

func (p *Mem) SetNx(key string, value interface{}) (bool, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	_, ok := p.data[key]
	if ok {
		return false, nil
	}

	p.data[key] = &Item{
		Data: utils.ToString(value),
	}

	return true, nil
}

func (p *Mem) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	_, ok := p.data[key]
	if ok {
		return false, nil
	}

	p.data[key] = &Item{
		Data:     utils.ToString(value),
		ExpireAt: time.Now().Add(timeout),
	}

	return true, nil
}

func (p *Mem) Get(key string) (string, error) {
	p.autoClear()

	p.RLock()
	defer p.RUnlock()

	val, ok := p.data[key]
	if !ok {
		return "", NotFound
	}

	if !val.ExpireAt.IsZero() && time.Now().After(val.ExpireAt) {
		return "", NotFound
	}

	return val.Data, nil
}

func (p *Mem) Set(key string, val any) error {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	p.data[key] = &Item{
		Data: utils.ToString(val),
	}

	return nil
}

func (p *Mem) Del(key ...string) error {
	p.autoClear()

	p.Lock()
	defer p.Unlock()

	for _, k := range key {
		delete(p.data, k)
	}

	return nil
}

func (p *Mem) Close() error {
	p.Lock()
	defer p.Unlock()

	p.data = make(map[string]*Item)

	return nil
}

func (p *Mem) Reset() error {
	p.Lock()
	defer p.Unlock()

	p.data = make(map[string]*Item)

	return nil
}

func NewMem() Cache {
	p := &Mem{
		data: make(map[string]*Item),
		rt:   rate.New(2, time.Minute),
	}

	return newBaseCache(p)
}
