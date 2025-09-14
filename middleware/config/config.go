package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/storage/etcd"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/lazygophers/utils/app"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/json"
	"github.com/lazygophers/utils/routine"
	"github.com/lazygophers/utils/runtime"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	cli *etcd.Client

	cacheLock sync.RWMutex
	cache     map[string]*core.ConfigItem

	watcher    *fsnotify.Watcher
	oncChanged func(item *core.ConfigItem, eventType EventType)
	keyword    string
}

func (p *Config) ListPrefix(prefix string) ([]string, error) {
	return p.cli.ListPrefix(p.etcdKey(prefix))
}

func (p *Config) Set(key string, value []byte) error {
	return p.cli.SetPb(p.etcdKey(key), &core.ConfigItem{
		Key:   key,
		Value: value,
	})
}

func (p *Config) SetJson(key string, value any) error {
	buffer, err := json.Marshal(value)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return p.Set(key, buffer)
}

func (p *Config) SetPb(key string, value proto.Message) error {
	buffer, err := proto.Marshal(value)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return p.Set(p.etcdKey(key), buffer)
}

func (p *Config) Get(key string) ([]byte, error) {
	item, ok := p.getLocal(key)
	if ok {
		return item.Value, nil
	}

	item, err := p.getDisk(key)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, xerror.NewError(int32(core.ErrCode_ConfigNotFound))
		}

		log.Errorf("err:%v", err)
		return nil, err
	}

	return item.Value, nil
}

func (p *Config) GetJson(key string, value interface{}) error {
	buffer, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	if buffer == nil {
		return nil
	}

	return json.Unmarshal(buffer, value)
}

func (p *Config) GetPb(key string, value proto.Message) error {
	buffer, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	if buffer == nil {
		return nil
	}

	return proto.Unmarshal(buffer, value)
}

func (p *Config) Delete(key string) error {
	return p.cli.Del(p.etcdKey(key))
}

func (p *Config) etcdKey(key string) string {
	return fmt.Sprintf("/%s/%s/%s", app.Organization, p.keyword, key)
}

func (p *Config) filePath(key string) string {
	return filepath.Join(runtime.LazyConfigDir(), p.keyword, key)
}

func (p *Config) getLocal(key string) (*core.ConfigItem, bool) {
	p.cacheLock.RLock()
	defer p.cacheLock.RUnlock()

	item, ok := p.cache[key]
	return item, ok
}

func (p *Config) mustGetLocal(key string) *core.ConfigItem {
	p.cacheLock.RLock()
	defer p.cacheLock.RUnlock()

	return p.cache[key]
}

func (p *Config) getDisk(key string) (*core.ConfigItem, error) {
	p.cacheLock.Lock()
	defer p.cacheLock.Unlock()

	log.Infof("read file from %s", p.filePath(key))

	buffer, err := os.ReadFile(p.filePath(key))
	if err != nil {
		if !os.IsNotExist(err) {
			log.Errorf("err:%v", err)
		}
		return nil, err
	}

	if len(buffer) == 0 {
		return &core.ConfigItem{
			Key: key,
		}, nil
	}

	var item core.ConfigItem
	err = proto.Unmarshal(buffer, &item)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	p.cache[key] = &item

	_ = p.watch(key)

	return &item, nil
}

func (p *Config) watch(key string) error {
	key = p.filePath(key)

	if candy.Contains(p.watcher.WatchList(), key) {
		return nil
	}

	err := p.watcher.Add(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *Config) upDisk(key string) error {
	p.cacheLock.Lock()
	defer p.cacheLock.Unlock()

	buffer, err := os.ReadFile(p.filePath(key))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	var item core.ConfigItem
	err = proto.Unmarshal(buffer, &item)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.cache[key] = &item

	return nil
}

func (p *Config) removeLocal(key string) error {
	p.cacheLock.Lock()
	defer p.cacheLock.Unlock()

	delete(p.cache, key)

	return nil
}

type EventType uint8

const (
	Changed EventType = iota + 1
	Deleted
)

func (p *Config) OnChanged(logic func(item *core.ConfigItem, eventType EventType)) {
	p.oncChanged = logic
}

func NewConfig(cli *etcd.Client, keywords ...string) *Config {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("err:%v", err)
		return nil
	}

	p := &Config{
		cli:     cli,
		watcher: watcher,
	}

	if len(keywords) > 0 {
		p.keyword = keywords[0]
	} else {
		p.keyword = "config"
	}

	routine.Go(func() (err error) {
		for event := range watcher.Events {
			name := filepath.Base(event.Name)

			switch event.Op {
			case fsnotify.Create, fsnotify.Write:
				log.Warnf("%s service changed", name)
				err = p.upDisk(name)
				if err != nil {
					log.Errorf("err:%v", err)
				}

				if p.oncChanged != nil {
					p.oncChanged(p.mustGetLocal(name), Deleted)
				}

			case fsnotify.Remove:
				log.Warnf("%s service removed", name)
				err = p.removeLocal(name)
				if err != nil {
					log.Errorf("err:%v", err)
				}

				if p.oncChanged != nil {
					p.oncChanged(&core.ConfigItem{
						Key: name,
					}, Deleted)
				}

			default:
				// do nothing
			}
		}

		return nil
	})

	return p
}
