package cache

import (
	"github.com/echovault/echovault/echovault"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/anyx"
	"strconv"
	"time"
)

type Echo struct {
	cli *echovault.EchoVault

	prefix string
}

func (p *Echo) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *Echo) Get(key string) (string, error) {
	value, err := p.cli.Get(p.prefix + key)
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	if value == "" {
		return "", NotFound
	}

	return value, nil
}

func (p *Echo) Set(key string, value any) error {
	_, _, err := p.cli.Set(p.prefix+key, anyx.ToString(value), echovault.SetOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (p *Echo) SetEx(key string, value any, timeout time.Duration) error {
	_, _, err := p.cli.Set(p.prefix+key, anyx.ToString(value), echovault.SetOptions{
		EXAT: int(time.Now().Add(timeout).Unix()),
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Echo) SetNx(key string, value interface{}) (bool, error) {
	_, ok, err := p.cli.Set(p.prefix+key, anyx.ToString(value), echovault.SetOptions{
		NX: true,
	})
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (p *Echo) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	log.Debugf("set nx ex %s", key)

	ok, err := p.SetNx(key, value)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	if ok {
		_, err = p.Expire(key, timeout)
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}

	return ok, nil
}

func (p *Echo) Ttl(key string) (time.Duration, error) {
	ttl, err := p.cli.TTL(p.prefix + key)
	if err != nil {
		return -1, err
	}

	if ttl == -1 {
		return -1, NotFound
	}
	if ttl == -2 {
		return -2, NotFound
	}

	return time.Second * time.Duration(ttl), nil
}

func (p *Echo) Expire(key string, timeout time.Duration) (bool, error) {
	return p.cli.Expire(p.prefix+key, int(timeout.Seconds()), echovault.ExpireOptions{})
}

func (p *Echo) Incr(key string) (int64, error) {
	value, err := p.cli.Incr(p.prefix + key)
	if err != nil {
		return -1, err
	}
	return int64(value), nil
}

func (p *Echo) Decr(key string) (int64, error) {
	value, err := p.cli.Decr(p.prefix + key)
	if err != nil {
		return -1, err
	}
	return int64(value), nil
}

func (p *Echo) IncrBy(key string, val int64) (int64, error) {
	value, err := p.cli.IncrBy(p.prefix+key, strconv.FormatInt(val, 10))
	if err != nil {
		return -1, err
	}
	return int64(value), nil
}

func (p *Echo) DecrBy(key string, val int64) (int64, error) {
	value, err := p.cli.DecrBy(p.prefix+key, strconv.FormatInt(val, 10))
	if err != nil {
		return -1, err
	}
	return int64(value), nil
}

func (p *Echo) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) HSet(key string, field string, value interface{}) (bool, error) {
	val, err := p.cli.HSet(p.prefix+key, map[string]string{
		field: anyx.ToString(value),
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return val > 0, nil
}

func (p *Echo) HGet(key, field string) (string, error) {
	values, err := p.cli.HGet(p.prefix+key, field)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", NotFound
	}

	return values[0], nil
}

func (p *Echo) HDel(key string, fields ...string) (int64, error) {
	value, err := p.cli.HDel(p.prefix+key, fields...)
	if err != nil {
		return -1, err
	}
	return int64(value), nil
}

func (p *Echo) HKeys(key string) ([]string, error) {
	values, err := p.cli.HKeys(p.prefix + key)
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (p *Echo) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) HExists(key string, field string) (bool, error) {
	return p.cli.HExists(p.prefix+key, field)
}

func (p *Echo) HIncr(key string, subKey string) (int64, error) {
	value, err := p.cli.HIncrBy(p.prefix+key, subKey, 1)
	if err != nil {
		return -1, err
	}
	return int64(value), nil
}

func (p *Echo) HIncrBy(key string, field string, increment int64) (int64, error) {
	value, err := p.cli.HIncrBy(p.prefix+key, field, int(increment))
	if err != nil {
		return -1, err
	}
	return int64(value), nil
}

func (p *Echo) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Echo) Del(key ...string) error {
	_, err := p.cli.Del(key...)
	if err != nil {
		return err
	}
	return nil
}

func (p *Echo) Close() error {
	p.cli.ShutDown()
	return nil
}

func NewEcho(c *Config) (Cache, error) {
	ec := echovault.DefaultConfig()
	if c.DataDir != "" {
		ec.DataDir = c.DataDir
	}

	ec.EvictionPolicy = "volatile-lfu"
	ec.EvictionInterval = time.Second
	ec.EvictionSample = 50

	ec.AOFSyncStrategy = "no"

	cli, err := echovault.NewEchoVault()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	p := &Echo{
		cli: cli,
	}

	return newBaseCache(p), nil
}
