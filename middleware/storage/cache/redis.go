package cache

import (
	"errors"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/anyx"
	"github.com/lazygophers/utils/app"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/shomali11/xredis"
)

type Redis struct {
	cli *xredis.Client
}

func NewRedis(address string, opts ...redis.DialOption) (Cache, error) {
	p := &Redis{
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

func (p *Redis) Incr(key string) (int64, error) {
	return p.cli.Incr(app.Name + ":" + key)
}

func (p *Redis) Decr(key string) (int64, error) {
	return p.cli.Decr(app.Name + ":" + key)
}

func (p *Redis) IncrBy(key string, value int64) (int64, error) {
	return p.cli.IncrBy(app.Name+":"+key, value)
}

func (p *Redis) IncrByFloat(key string, increment float64) (float64, error) {
	return p.cli.IncrByFloat(app.Name+":"+key, increment)
}

func (p *Redis) DecrBy(key string, value int64) (int64, error) {
	return p.cli.DecrBy(app.Name+":"+key, value)
}

func (p *Redis) Get(key string) (string, error) {
	log.Debugf("get %s", key)

	val, ok, err := p.cli.Get(app.Name + ":" + key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", NotFound
	}

	return val, nil
}

func (p *Redis) Exists(keys ...string) (bool, error) {
	ok, err := p.cli.Exists(candy.Map(keys, func(key string) string {
		return app.Name + ":" + key
	})...)
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return false, nil
		}

		return false, err
	}

	return ok, nil
}

func (p *Redis) SetNx(key string, value interface{}) (bool, error) {
	log.Debugf("set nx %s", key)

	ok, err := p.cli.SetNx(app.Name+":"+key, anyx.ToString(value))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}

		return false, err
	}

	return ok, nil
}

func (p *Redis) Expire(key string, timeout time.Duration) (bool, error) {
	ok, err := p.cli.Expire(app.Name+":"+key, int64(timeout.Seconds()))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}

		return false, err
	}

	return ok, nil
}

func (p *Redis) Ttl(key string) (time.Duration, error) {
	connection := p.cli.GetConnection()

	ttl, err := redis.Int64(connection.Do("TTL", app.Name+":"+key))
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return time.Duration(ttl) * time.Second, nil
}

func (p *Redis) Set(key string, value interface{}) (err error) {
	log.Debugf("set %s", app.Name+":"+key)

	_, err = p.cli.Set(app.Name+":"+key, anyx.ToString(value))
	return err
}

func (p *Redis) SetEx(key string, value interface{}, timeout time.Duration) error {
	log.Debugf("set ex %s", app.Name+":"+key)

	_, err := p.cli.SetEx(app.Name+":"+key, anyx.ToString(value), int64(timeout.Seconds()))
	return err
}

func (p *Redis) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	log.Debugf("set nx ex %s", app.Name+":"+key)

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

func (p *Redis) Del(keys ...string) (err error) {
	_, err = p.cli.Del(candy.Map(keys, func(key string) string {
		return app.Name + ":" + key
	})...)
	return
}

func (p *Redis) HSet(key string, field string, value interface{}) (bool, error) {
	return p.cli.HSet(app.Name+":"+key, field, anyx.ToString(value))
}

func (p *Redis) HGet(key, field string) (string, error) {
	val, ok, err := p.cli.HGet(app.Name+":"+key, field)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", NotFound
	}

	return val, nil
}

func (p *Redis) HGetAll(key string) (map[string]string, error) {
	return p.cli.HGetAll(app.Name + ":" + key)
}

func (p *Redis) HKeys(key string) ([]string, error) {
	return p.cli.HKeys(app.Name + ":" + key)
}

func (p *Redis) HVals(key string) ([]string, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	return redis.Strings(conn.Do("HVALS", app.Name+":"+key))
}

func (p *Redis) HDel(key string, fields ...string) (int64, error) {
	return p.cli.HDel(app.Name+":"+key, fields...)
}

func (p *Redis) SAdd(key string, members ...string) (int64, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	args := make([]interface{}, 0, len(members)+1)
	args = append(args, app.Name+":"+key)
	for _, member := range members {
		args = append(args, member)
	}

	return redis.Int64(conn.Do("SADD", args...))
}

func (p *Redis) SMembers(key string) ([]string, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	return redis.Strings(conn.Do("SMEMBERS", app.Name+":"+key))
}

func (p *Redis) SRem(key string, members ...string) (int64, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	args := make([]interface{}, 0, len(members)+1)
	args = append(args, app.Name+":"+key)
	for _, member := range members {
		args = append(args, member)
	}

	return redis.Int64(conn.Do("SREM", args...))
}

func (p *Redis) SRandMember(key string, count ...int64) ([]string, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	args := make([]interface{}, 0)
	args = append(args, app.Name+":"+key)
	if len(count) > 0 {
		args = append(args, count[0])
	} else {
		args = append(args, 1)
	}

	return redis.Strings(conn.Do("SRANDMEMBER", args...))
}

func (p *Redis) SPop(key string) (string, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	return redis.String(conn.Do("SPOP", app.Name+":"+key))
}

func (p *Redis) HIncr(key string, subKey string) (int64, error) {
	return p.cli.HIncr(app.Name+":"+key, subKey)
}

func (p *Redis) HIncrBy(key string, field string, increment int64) (int64, error) {
	return p.cli.HIncrBy(app.Name+":"+key, field, increment)
}

func (p *Redis) HIncrByFloat(key string, field string, increment float64) (float64, error) {
	return p.cli.HIncrByFloat(app.Name+":"+key, field, increment)
}

func (p *Redis) HDecr(key string, field string) (int64, error) {
	return p.cli.HDecr(app.Name+":"+key, field)
}

func (p *Redis) HDecrBy(key string, field string, increment int64) (int64, error) {
	return p.cli.HDecrBy(app.Name+":"+key, field, increment)
}

func (p *Redis) HExists(key, field string) (bool, error) {
	return p.cli.HExists(app.Name+":"+key, field)
}

func (p *Redis) SisMember(key, field string) (bool, error) {
	conn := p.cli.GetConnection()
	defer conn.Close()

	args := make([]interface{}, 0, 2)
	args = append(args, app.Name+":"+key, field)

	return redis.Bool(conn.Do("SISMEMBER", args...))
}

func (p *Redis) Close() error {
	return p.cli.Close()
}
