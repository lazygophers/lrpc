package cache

import (
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
	"go.mills.io/bitcask/v2"
)

type CacheBitcask struct {
	prefix string

	cli *bitcask.Bitcask
}

func (p *CacheBitcask) Clean() error {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) Get(key string) (string, error) {
	value, err := p.cli.Get(bitcask.Key(key))
	if err != nil {
		if err == bitcask.ErrKeyNotFound {
			return "", NotFound
		}
		return "", err
	}

	return string(value), nil
}

func (p *CacheBitcask) Set(key string, value any) error {
	return p.cli.Put(bitcask.Key(key), bitcask.Value(candy.ToString(value)))
}

func (p *CacheBitcask) SetEx(key string, value any, timeout time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) SetNx(key string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) Ttl(key string) (time.Duration, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) Expire(key string, timeout time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) Incr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) Decr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) IncrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) DecrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HSet(key string, field string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HGet(key, field string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HDel(key string, fields ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HKeys(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HExists(key string, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HIncr(key string, subKey string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HIncrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) Del(key ...string) error {
	//TODO implement me
	panic("implement me")
}

func (p *CacheBitcask) Close() error {
	return p.cli.Close()
}

func (p *CacheBitcask) SetPrefix(prefix string) {
	p.prefix = prefix
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
