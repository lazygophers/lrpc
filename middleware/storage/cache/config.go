package cache

import (
	"github.com/lazygophers/utils/app"
	"os"
	"path/filepath"
)

const (
	Mem     string = "mem"
	Redis   string = "redis"
	Bbolt   string = "bbolt"
	SugarDB string = "sugardb"
	Bitcask string = "bitcask"
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
	case Redis:
		if c.Address == "" {
			c.Address = "127.0.0.1:6379"
		}
	}
}
