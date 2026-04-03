package cache

import (
	"os"
	"path/filepath"
	"time"

	"github.com/lazygophers/utils/app"
)

const (
	Mem     string = "mem"
	Redis   string = "redis"
	Bbolt   string = "bbolt"
	SugarDB string = "sugardb"
	Bitcask string = "bitcask"
	LevelDB string = "leveldb"
)

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

	Mock bool `yaml:"mock,omitempty" json:"mock,omitempty"`

	// Redis connection pool configuration
	// PoolSize is the maximum number of socket connections.
	// Default: 1 (significantly reduces memory usage)
	PoolSize int `yaml:"pool_size,omitempty" json:"pool_size,omitempty"`

	// MinIdleConns is the minimum number of idle connections.
	// Default: 5 (maintain warm connections for high performance)
	MinIdleConns int `yaml:"min_idle_conns,omitempty" json:"min_idle_conns,omitempty"`

	// MaxIdleConns is the maximum number of idle connections.
	// Default: 20 (balance between resource usage and performance)
	MaxIdleConns int `yaml:"max_idle_conns,omitempty" json:"max_idle_conns,omitempty"`

	// ConnMaxLifetime is the maximum connection lifetime.
	// Default: 1 hour (avoid long-lived connection issues)
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime,omitempty" json:"conn_max_lifetime,omitempty"`

	// ConnMaxIdleTime is the maximum idle time for connections.
	// Default: 10 minutes (release idle connections promptly)
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time,omitempty" json:"conn_max_idle_time,omitempty"`

	// MaxRetries is the maximum number of retries.
	// Default: 3 (resilient to transient network failures)
	MaxRetries int `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`

	// MinRetryBackoff is the minimum backoff between retries.
	// Default: 8ms
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff,omitempty" json:"min_retry_backoff,omitempty"`

	// MaxRetryBackoff is the maximum backoff between retries.
	// Default: 512ms
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff,omitempty" json:"max_retry_backoff,omitempty"`
}

func (c *Config) apply() {
	if c.Type == "" {
		c.Type = "mem"
	}

	switch c.Type {
	case Bbolt:
		if c.Address == "" {
			c.Address, _ = os.Executable()
			c.Address = filepath.Join(c.Address, app.Name+".cache")
		}
	case SugarDB:
		if c.DataDir == "" {
			c.DataDir, _ = os.Executable()
			c.DataDir = filepath.Join(c.DataDir, app.Name+".cache")
		}
	case Bitcask:
		if c.DataDir == "" {
			c.DataDir, _ = os.Executable()
			c.DataDir = filepath.Join(c.DataDir, app.Name+".cache")
		}
	case LevelDB:
		if c.DataDir == "" {
			c.DataDir, _ = os.Executable()
			c.DataDir = filepath.Join(c.DataDir, app.Name+".cache")
		}
	case Redis:
		if c.Address == "" {
			c.Address = "127.0.0.1:6379"
		}
		// Set default PoolSize to 1 to reduce memory usage
		if c.PoolSize == 0 {
			c.PoolSize = 1
		}
		// Set high-performance defaults for connection pool
		if c.MinIdleConns == 0 {
			c.MinIdleConns = 5 // Maintain 5 minimum idle connections
		}
		if c.MaxIdleConns == 0 {
			c.MaxIdleConns = 20 // Maximum 20 idle connections
		}
		if c.ConnMaxLifetime == 0 {
			c.ConnMaxLifetime = time.Hour // Connection lifetime: 1 hour
		}
		if c.ConnMaxIdleTime == 0 {
			c.ConnMaxIdleTime = 10 * time.Minute // Idle timeout: 10 minutes
		}
		if c.MaxRetries == 0 {
			c.MaxRetries = 3 // Retry 3 times on network failure
		}
		if c.MinRetryBackoff == 0 {
			c.MinRetryBackoff = 8 * time.Millisecond // Minimum backoff: 8ms
		}
		if c.MaxRetryBackoff == 0 {
			c.MaxRetryBackoff = 512 * time.Millisecond // Maximum backoff: 512ms
		}
	}
}
