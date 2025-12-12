package cache

import (
	"strconv"
	"time"

	"github.com/echovault/sugardb/sugardb"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
)

type CacheSugarDB struct {
	cli *sugardb.SugarDB
}

func (p *CacheSugarDB) Clean() error {
	p.cli.Flush(-1)
	return nil
}

func (p *CacheSugarDB) SetPrefix(prefix string) {
}

func (p *CacheSugarDB) Get(key string) (string, error) {
	value, err := p.cli.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return "", ErrNotFound
	}
	if value == "" {
		return "", ErrNotFound
	}

	return value, nil
}

func (p *CacheSugarDB) Set(key string, value any) error {
	_, _, err := p.cli.Set(key, candy.ToString(value), sugardb.SETOptions{})
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheSugarDB) SetEx(key string, value any, timeout time.Duration) error {
	_, _, err := p.cli.Set(key, candy.ToString(value), sugardb.SETOptions{
		ExpireOpt:  sugardb.SETEX,
		ExpireTime: int(time.Now().Add(timeout).Unix()),
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheSugarDB) SetNx(key string, value interface{}) (bool, error) {
	_, ok, err := p.cli.Set(key, candy.ToString(value), sugardb.SETOptions{
		WriteOpt: sugardb.SETNX,
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return ok, err
	}
	return ok, nil
}

func (p *CacheSugarDB) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
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

func (p *CacheSugarDB) Ttl(key string) (time.Duration, error) {
	ttl, err := p.cli.TTL(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}

	if ttl == -1 {
		return -1, ErrNotFound
	}
	if ttl == -2 {
		return -2, ErrNotFound
	}

	return time.Second * time.Duration(ttl), nil
}

func (p *CacheSugarDB) Expire(key string, timeout time.Duration) (bool, error) {
	ok, err := p.cli.Expire(key, int(timeout.Seconds()))
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return ok, nil
}

func (p *CacheSugarDB) Incr(key string) (int64, error) {
	value, err := p.cli.Incr(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) Decr(key string) (int64, error) {
	value, err := p.cli.Decr(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) IncrBy(key string, val int64) (int64, error) {
	value, err := p.cli.IncrBy(key, strconv.FormatInt(val, 10))
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) DecrBy(key string, val int64) (int64, error) {
	value, err := p.cli.DecrBy(key, strconv.FormatInt(val, 10))
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) Exists(keys ...string) (bool, error) {
	count, err := p.cli.Exists(keys...)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return count > 0, nil
}

func (p *CacheSugarDB) HSet(key string, field string, value interface{}) (bool, error) {
	val, err := p.cli.HSet(key, map[string]string{
		field: candy.ToString(value),
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return val > 0, nil
}

func (p *CacheSugarDB) HGet(key, field string) (string, error) {
	values, err := p.cli.HGet(key, field)
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	if len(values) == 0 {
		return "", ErrNotFound
	}

	return values[0], nil
}

func (p *CacheSugarDB) HDel(key string, fields ...string) (int64, error) {
	value, err := p.cli.HDel(key, fields...)
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) HKeys(key string) ([]string, error) {
	values, err := p.cli.HKeys(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return values, nil
}

func (p *CacheSugarDB) HGetAll(key string) (map[string]string, error) {
	values, err := p.cli.HGetAll(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	result := make(map[string]string)
	for i := 0; i < len(values); i += 2 {
		if i+1 < len(values) {
			result[values[i]] = values[i+1]
		}
	}

	return result, nil
}

func (p *CacheSugarDB) HExists(key string, field string) (bool, error) {
	ok, err := p.cli.HExists(key, field)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return ok, nil
}

func (p *CacheSugarDB) HIncr(key string, subKey string) (int64, error) {
	value, err := p.cli.HIncrBy(key, subKey, 1)
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) HIncrBy(key string, field string, increment int64) (int64, error) {
	value, err := p.cli.HIncrBy(key, field, int(increment))
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) HDecr(key string, field string) (int64, error) {
	val, err := p.HDecrBy(key, field, 1)
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return val, nil
}

func (p *CacheSugarDB) HDecrBy(key string, field string, increment int64) (int64, error) {
	value, err := p.cli.HIncrBy(key, field, -int(increment))
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) SAdd(key string, members ...string) (int64, error) {
	value, err := p.cli.SAdd(key, members...)
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) SMembers(key string) ([]string, error) {
	values, err := p.cli.SMembers(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return values, nil
}

func (p *CacheSugarDB) SRem(key string, members ...string) (int64, error) {
	value, err := p.cli.SRem(key, members...)
	if err != nil {
		log.Errorf("err:%v", err)
		return -1, err
	}
	return int64(value), nil
}

func (p *CacheSugarDB) SRandMember(key string, count ...int64) ([]string, error) {
	cnt := 1
	if len(count) > 0 {
		cnt = int(count[0])
	}

	values, err := p.cli.SRandMember(key, cnt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return values, nil
}

func (p *CacheSugarDB) SPop(key string) (string, error) {
	values, err := p.cli.SPop(key, 1)
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}

	if len(values) == 0 {
		return "", ErrNotFound
	}
	return values[0], nil
}

func (p *CacheSugarDB) SisMember(key, field string) (bool, error) {
	ok, err := p.cli.SisMember(key, field)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return ok, nil
}

func (p *CacheSugarDB) Del(key ...string) error {
	_, err := p.cli.Del(key...)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheSugarDB) Close() error {
	p.cli.ShutDown()
	return nil
}

func (p *CacheSugarDB) Ping() error {
	// SugarDB doesn't have Ping method, use DBSize as health check
	_, err := p.cli.DBSize()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func NewSugarDB(c *Config) (Cache, error) {
	ec := sugardb.DefaultConfig()
	if c.DataDir != "" {
		ec.DataDir = c.DataDir
	}

	if c.Password != "" {
		ec.Password = c.Password
	}

	ec.RestoreSnapshot = true
	ec.RestoreAOF = true

	ec.SnapShotThreshold = 100
	ec.SnapshotInterval = time.Minute

	cli, err := sugardb.NewSugarDB(sugardb.WithConfig(ec))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	p := &CacheSugarDB{
		cli: cli,
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
