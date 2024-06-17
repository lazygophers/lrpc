package etcd

import (
	"errors"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/runtime"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var (
	ErrNotFound = errors.New("key not found")
)

func Connect(c *Config) (*Client, error) {
	cli, err := NewClient(c)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	log.Infof("connnect etcd successfully")

	return cli, nil
}

func ConnectWithLazy(lazypaths ...string) (*Client, error) {
	lazypath := filepath.Join(runtime.LazyConfigDir(), "etcd.yaml")
	if len(lazypaths) > 0 {
		lazypath = lazypaths[0]
	}

	log.Infof("load etcd configuration from %s", lazypath)

	file, err := os.Open(lazypath)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	defer file.Close()

	var c Config
	err = yaml.NewDecoder(file).Decode(&c)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return Connect(&c)
}

type EventType uint8

const (
	Changed EventType = iota + 1
	Deleted
)

func (e EventType) String() string {
	switch e {
	case Changed:
		return "changed"
	case Deleted:
		return "delete"
	default:
		return "known"
	}
}

type Event struct {
	Key     string
	Type    EventType
	Value   []byte
	Version int64
}
