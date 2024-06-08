package etcd

import (
	"errors"
	"github.com/lazygophers/log"
	"sync"
	"time"
)

var (
	cli *Client

	once sync.Once
)

var (
	ErrNotFound = errors.New("key not found")
)

func Connect(c *Config) error {
	if cli != nil {
		return nil
	}

	var err error

	once.Do(func() {
		cli, err = NewClient(c)
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	return err
}

func DefaultCli() *Client {
	return cli
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
	Value   string
	Version int64
}

func Watch(key string, logic func(event *Event)) {
	cli.Watch(key, logic)
}

func WatchPrefix(key string, logic func(event *Event)) {
	cli.WatchPrefix(key, logic)
}

func Set(key string, val interface{}) error {
	return cli.Set(key, val)
}

func SetWithLock(key string, val interface{}) error {
	return cli.SetWithLock(key, val)
}

func WhenLocked(key string, logic func(client *Client, key string) (err error)) error {
	return cli.WhenLocked(key, logic)
}

func SetEx(key string, val interface{}, timeout time.Duration) error {
	return cli.SetEx(key, val, timeout)
}

func Get(key string) ([]byte, error) {
	return cli.Get(key)
}

func Prefix(key string) (map[string][]byte, error) {
	return cli.Prefix(key)
}

func GetJson(key string, j interface{}) error {
	return cli.GetJson(key, j)
}
