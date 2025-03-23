package cache

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/lazygophers/utils/app"
	"github.com/lazygophers/utils/json"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"time"
)

var NotFound = errors.New("key not found")

type BaseCache interface {
	SetPrefix(prefix string)

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

	Clean() error
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

	SetPb(key string, j proto.Message) error
	SetPbEx(key string, j proto.Message, timeout time.Duration) error
	GetPb(key string, j proto.Message) error

	HGetJson(key, field string, j interface{}) error

	Limit(key string, limit int64, timeout time.Duration) (bool, error)
	LimitUpdateOnCheck(key string, limit int64, timeout time.Duration) (bool, error)
}

type Config struct {
	// Cache type, support mem, redis, bbolt, default mem
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Cache address
	// mem: empty
	// redis: redis address, default 127.0.0.1:6379
	// bbolt: bbolt file path, default ./ice.cache
	Address string `yaml:"address,omitempty" json:"address,omitempty"`

	// Cache password
	// mem: empty
	// redis: redis password
	// bbolt: empty
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Cache db
	// mem: empty
	// redis: redis db, default 0
	// bbolt: empty
	Db int `yaml:"db,omitempty" json:"db,omitempty"`

	// Cache data dir
	// mem: empty
	// redis: empty
	// bbolt: empty
	// echo: DataDir, default .
	DataDir string `yaml:"data_dir,omitempty" json:"data_dir,omitempty"`
}

func (c *Config) apply() {
	if c.Type == "" {
		c.Type = "mem"
	}

	switch c.Type {
	case "bbolt":
		if c.Address == "" {
			c.Address, _ = os.Executable()
			c.Address = filepath.Join(c.Address, app.Name+".cache")
		}
	case "echo":
		if c.DataDir == "" {
			c.DataDir, _ = os.Executable()
			c.DataDir = filepath.Join(c.DataDir, app.Name+".cache")
		}
	case "bitcask":
		if c.DataDir == "" {
			c.DataDir, _ = os.Executable()
			c.DataDir = filepath.Join(c.DataDir, app.Name+".cache")
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
			redis.DialDatabase(c.Db),
			redis.DialConnectTimeout(time.Second*3),
			redis.DialReadTimeout(time.Minute),
			redis.DialWriteTimeout(time.Minute),
			redis.DialKeepAlive(time.Minute),
			redis.DialPassword(c.Password),
		)

	case "mem":
		return NewMem(), nil

	case "echo":
		return NewEcho(c)

	case "bitcask":
		return NewBitcask(c)

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
