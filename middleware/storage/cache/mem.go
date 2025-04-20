package cache

import (
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
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) SetPrefix(prefix string) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) IncrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) DecrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) Expire(key string, timeout time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) Ttl(key string) (time.Duration, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) Incr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) Decr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HIncr(key string, subKey string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HIncrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HExists(key string, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HKeys(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HSet(key string, field string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HGet(key, field string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HDel(key string, fields ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheMem) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
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

		if time.Now().After(v.ExpireAt) {
			continue
		}
	}

	p.data = data
}

func (p *CacheMem) SetNx(key string, value interface{}) (bool, error) {
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

func (p *CacheMem) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
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

func (p *CacheMem) Get(key string) (string, error) {
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

func NewMem() Cache {
	p := &CacheMem{
		data: make(map[string]*Item),
		rt:   rate.New(2, time.Minute),
	}

	return newBaseCache(p)
}
