package cache

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/lazygophers/utils/json"
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"time"
)

var NotFound = errors.New("key not found")

type BaseCache interface {
	Get(key string) (string, error)

	Set(key string, value any) error
	SetEx(key string, value any, timeout time.Duration) error
	SetNx(key string, value interface{}) (bool, error)
	SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error)

	Ttl(key string) (time.Duration, error)
	Expire(key string, timeout time.Duration) (bool, error)

	Incr(key string) (int64, error)
	Decr(key string) (int64, error)
	IncrBy(key string, value int64) (int64, error)
	DecrBy(key string, value int64) (int64, error)

	Exists(keys ...string) (bool, error)

	HSet(key string, field string, value interface{}) (bool, error)
	HGet(key, field string) (string, error)
	HDel(key string, fields ...string) (int64, error)
	HKeys(key string) ([]string, error)
	HGetAll(key string) (map[string]string, error)
	HExists(key string, field string) (bool, error)
	HIncr(key string, subKey string) (int64, error)
	HIncrBy(key string, field string, increment int64) (int64, error)
	HDecr(key string, field string) (int64, error)
	HDecrBy(key string, field string, increment int64) (int64, error)

	SAdd(key string, members ...string) (int64, error)
	SMembers(key string) ([]string, error)
	SRem(key string, members ...string) (int64, error)
	SRandMember(key string, count ...int64) ([]string, error)
	SPop(key string) (string, error)
	SisMember(key, field string) (bool, error) // 成员是否存在

	Del(key ...string) error

	//Reset() error

	Close() error
}

type Cache interface {
	BaseCache

	GetBool(key string) (bool, error)
	GetInt(key string) (int, error)
	GetUint(key string) (uint, error)
	GetInt32(key string) (int32, error)
	GetUint32(key string) (uint32, error)
	GetInt64(key string) (int64, error)
	GetUint64(key string) (uint64, error)
	GetFloat32(key string) (float32, error)
	GetFloat64(key string) (float64, error)

	GetSlice(key string) ([]string, error)
	GetBoolSlice(key string) ([]bool, error)
	GetIntSlice(key string) ([]int, error)
	GetUintSlice(key string) ([]uint, error)
	GetInt32Slice(key string) ([]int32, error)
	GetUint32Slice(key string) ([]uint32, error)
	GetInt64Slice(key string) ([]int64, error)
	GetUint64Slice(key string) ([]uint64, error)
	GetFloat32Slice(key string) ([]float32, error)
	GetFloat64Slice(key string) ([]float64, error)

	GetJson(key string, j interface{}) error

	HGetJson(key, field string, j interface{}) error

	Limit(key string, limit int64, timeout time.Duration) (bool, error)
}

type Config struct {
	// Cache type, support mem, redis, bbolt, default mem
	Type string `yaml:"type"`

	// Cache address
	// mem: empty
	// redis: redis address, default 127.0.0.1:6379
	// bbolt: bbolt file path, default ./ice.cache
	Address string `yaml:"address"`

	// Cache password
	// mem: empty
	// redis: redis password
	// bbolt: empty
	Password string `yaml:"password"`
}

func (c *Config) apply() {
	if c.Type == "" {
		c.Type = "mem"
	}

	switch c.Type {
	case "bbolt":
		if c.Address == "" {
			c.Address, _ = os.Executable()
			c.Address = filepath.Join(c.Address, "ice.cache")
		}
	case "redis":
		if c.Address == "" {
			c.Address = "127.0.0.1:6379"
		}
	}
}

func New(c *Config) (Cache, error) {
	c.apply()

	switch c.Type {
	case "bbolt":
		return NewBbolt(c.Address, &bbolt.Options{
			Timeout:  time.Second * 5,
			ReadOnly: false,
		})

	case "redis":
		return NewRedis(c.Address,
			redis.DialDatabase(0),
			redis.DialConnectTimeout(time.Second*3),
			redis.DialReadTimeout(time.Second*3),
			redis.DialWriteTimeout(time.Second*3),
			redis.DialKeepAlive(time.Minute),
			redis.DialPassword(c.Password),
		)

	case "mem":
		return NewMem(), nil

	default:
		return nil, errors.New("cache type not support")
	}
}

type Item struct {
	Data string `json:"data,omitempty"`

	ExpireAt time.Time `json:"expire_at,omitempty"`
}

func (p *Item) Bytes() []byte {
	buf, _ := json.Marshal(p)
	return buf
}

func (p *Item) String() string {
	str, _ := json.MarshalString(p)
	return str
}
