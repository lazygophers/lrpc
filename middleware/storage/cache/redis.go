package cache

import (
	"errors"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/shomali11/xredis"
)

type CacheRedis struct {
	cli *xredis.Client

	prefix string
}

func (p *CacheRedis) Clean() error {
	conn := p.cli.GetConnection()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", p.prefix+"*"))
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	_, err = conn.Do("DEL", args...)
	return err
}

func (p *CacheRedis) SetPrefix(prefix string) {
	p.prefix = prefix
}

func NewRedis(address string, opts ...redis.DialOption) (Cache, error) {
	p := &CacheRedis{
		cli: xredis.NewClient(&redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", address, opts...)
			},
			MaxIdle:     1000,
			MaxActive:   1000,
			IdleTimeout: time.Second * 5,
			Wait:        true,
		}),
	}

	atexit.Register(func() {
		err := p.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	pong, err := p.cli.Ping()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	log.Infof("ping:%v", pong)

	return newBaseCache(p), nil
}

func (p *CacheRedis) Incr(key string) (int64, error) {
	return p.cli.Incr(p.prefix + key)
}

func (p *CacheRedis) Decr(key string) (int64, error) {
	return p.cli.Decr(p.prefix + key)
}

func (p *CacheRedis) IncrBy(key string, value int64) (int64, error) {
	return p.cli.IncrBy(p.prefix+key, value)
}

func (p *CacheRedis) IncrByFloat(key string, increment float64) (float64, error) {
	return p.cli.IncrByFloat(p.prefix+key, increment)
}

func (p *CacheRedis) DecrBy(key string, value int64) (int64, error) {
	return p.cli.DecrBy(p.prefix+key, value)
}

func (p *CacheRedis) Get(key string) (string, error) {
	log.Debugf("get %s", key)

	val, ok, err := p.cli.Get(p.prefix + key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", NotFound
	}

	return val, nil
}

func (p *CacheRedis) Exists(keys ...string) (bool, error) {
	ok, err := p.cli.Exists(candy.Map(keys, func(key string) string {
		return p.prefix + key
	})...)
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return false, nil
		}

		return false, err
	}

	return ok, nil
}

func (p *CacheRedis) SetNx(key string, value interface{}) (bool, error) {
	log.Debugf("set nx %s", key)

	ok, err := p.cli.SetNx(p.prefix+key, candy.ToString(value))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}

		return false, err
	}

	return ok, nil
}

func (p *CacheRedis) Expire(key string, timeout time.Duration) (bool, error) {
	ok, err := p.cli.Expire(p.prefix+key, int64(timeout.Seconds()))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}

		return false, err
	}

	return ok, nil
}

func (p *CacheRedis) Ttl(key string) (time.Duration, error) {
	connection := p.cli.GetConnection()

	ttl, err := redis.Int64(connection.Do("TTL", p.prefix+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return time.Duration(ttl) * time.Second, nil
}

func (p *CacheRedis) Set(key string, value interface{}) (err error) {
	log.Debugf("set %s", key)

	_, err = p.cli.Set(p.prefix+key, candy.ToString(value))
	return err
}

func (p *CacheRedis) SetEx(key string, value interface{}, timeout time.Duration) error {
	log.Debugf("set ex %s", key)

	_, err := p.cli.SetEx(p.prefix+key, candy.ToString(value), int64(timeout.Seconds()))
	return err
}

func (p *CacheRedis) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
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

func (p *CacheRedis) Del(keys ...string) (err error) {
	_, err = p.cli.Del(candy.Map(keys, func(key string) string {
		return p.prefix + key
	})...)
	return
}

func (p *CacheRedis) HSet(key string, field string, value interface{}) (bool, error) {
	return p.cli.HSet(p.prefix+key, field, candy.ToString(value))
}

func (p *CacheRedis) HGet(key, field string) (string, error) {
	val, ok, err := p.cli.HGet(p.prefix+key, field)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", NotFound
	}

	return val, nil
}

func (p *CacheRedis) HGetAll(key string) (map[string]string, error) {
	return p.cli.HGetAll(p.prefix + key)
}

func (p *CacheRedis) HKeys(key string) ([]string, error) {
	return p.cli.HKeys(p.prefix + key)
}

func (p *CacheRedis) HVals(key string) ([]string, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	return redis.Strings(conn.Do("HVALS", p.prefix+key))
}

func (p *CacheRedis) HDel(key string, fields ...string) (int64, error) {
	return p.cli.HDel(p.prefix+key, fields...)
}

func (p *CacheRedis) SAdd(key string, members ...string) (int64, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	args := make([]interface{}, 0, len(members)+1)
	args = append(args, p.prefix+key)
	for _, member := range members {
		args = append(args, member)
	}

	return redis.Int64(conn.Do("SADD", args...))
}

func (p *CacheRedis) SMembers(key string) ([]string, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	return redis.Strings(conn.Do("SMEMBERS", p.prefix+key))
}

func (p *CacheRedis) SRem(key string, members ...string) (int64, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	args := make([]interface{}, 0, len(members)+1)
	args = append(args, p.prefix+key)
	for _, member := range members {
		args = append(args, member)
	}

	return redis.Int64(conn.Do("SREM", args...))
}

func (p *CacheRedis) SRandMember(key string, count ...int64) ([]string, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	args := make([]interface{}, 0)
	args = append(args, p.prefix+key)
	if len(count) > 0 {
		args = append(args, count[0])
	} else {
		args = append(args, 1)
	}

	return redis.Strings(conn.Do("SRANDMEMBER", args...))
}

func (p *CacheRedis) SPop(key string) (string, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	return redis.String(conn.Do("SPOP", p.prefix+key))
}

func (p *CacheRedis) HIncr(key string, subKey string) (int64, error) {
	return p.cli.HIncr(p.prefix+key, subKey)
}

func (p *CacheRedis) HIncrBy(key string, field string, increment int64) (int64, error) {
	return p.cli.HIncrBy(p.prefix+key, field, increment)
}

func (p *CacheRedis) HIncrByFloat(key string, field string, increment float64) (float64, error) {
	return p.cli.HIncrByFloat(p.prefix+key, field, increment)
}

func (p *CacheRedis) HDecr(key string, field string) (int64, error) {
	return p.cli.HDecr(p.prefix+key, field)
}

func (p *CacheRedis) HDecrBy(key string, field string, increment int64) (int64, error) {
	return p.cli.HDecrBy(p.prefix+key, field, increment)
}

func (p *CacheRedis) HExists(key, field string) (bool, error) {
	return p.cli.HExists(p.prefix+key, field)
}

func (p *CacheRedis) SisMember(key, field string) (bool, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	args := make([]interface{}, 0, 2)
	args = append(args, p.prefix+key, field)

	return redis.Bool(conn.Do("SISMEMBER", args...))
}

func (p *CacheRedis) Close() error {
	return p.cli.Close()
}
