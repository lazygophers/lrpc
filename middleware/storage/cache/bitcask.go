package cache

import (
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/anyx"
	"github.com/lazygophers/utils/atexit"
	"go.mills.io/bitcask/v2"
	"time"
)

type Bitcask struct {
	prefix string

	cli *bitcask.Bitcask
}

func (p *Bitcask) Get(key string) (string, error) {
	value, err := p.cli.Get(bitcask.Key(key))
	if err != nil {
		if err == bitcask.ErrKeyNotFound {
			return "", NotFound
		}
		return "", err
	}

	return string(value), nil
}

func (p *Bitcask) Set(key string, value any) error {
	return p.cli.Put(bitcask.Key(key), bitcask.Value(anyx.ToString(value)))
}

func (p *Bitcask) SetEx(key string, value any, timeout time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) SetNx(key string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) Ttl(key string) (time.Duration, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) Expire(key string, timeout time.Duration) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) Incr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) Decr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) IncrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) DecrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HSet(key string, field string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HGet(key, field string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HDel(key string, fields ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HKeys(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HExists(key string, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HIncr(key string, subKey string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HIncrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) Del(key ...string) error {
	//TODO implement me
	panic("implement me")
}

func (p *Bitcask) Close() error {
	return p.cli.Close()
}

func (p *Bitcask) SetPrefix(prefix string) {
	p.prefix = prefix
}

func NewBitcask(c *Config) (Cache, error) {
	var err error
	p := &Bitcask{}

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
