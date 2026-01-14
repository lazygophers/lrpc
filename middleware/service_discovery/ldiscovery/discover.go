package ldiscovery

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/storage/etcd"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/lazygophers/utils/app"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/routine"
	"github.com/lazygophers/utils/runtime"
	"google.golang.org/protobuf/proto"

	"github.com/fsnotify/fsnotify"
)

type Discovery struct {
	client  *etcd.Client
	watcher *fsnotify.Watcher

	cacheLock sync.RWMutex
	cache     map[string]*core.ServiceDiscoveryService
}

func (p *Discovery) AddNode(n *core.ServiceDiscoveryService) error {
	log.Infof("try add node to %s", n.ServiceName)
	etcdKey := fmt.Sprintf("/%s/%s/%s", app.Organization, "discovery", n.ServiceName)

	log.Infof("try add node to %s", etcdKey)

	for i := 0; i < 5; i++ {
		var service core.ServiceDiscoveryService
		value, version, err := p.client.GetWithVersion(etcdKey)
		if err != nil {
			if !errors.Is(err, etcd.ErrNotFound) {
				log.Errorf("err:%v", err)
				return err
			}
			service = core.ServiceDiscoveryService{
				ServiceName: n.ServiceName,
			}
		} else {
			err = proto.Unmarshal(value, &service)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}

		for _, node := range n.NodeList {
			(&service).MergeNode(node)
		}

		ok, err := p.client.SetPbVersion(etcdKey, &service, version)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if ok {
			return nil
		}
	}

	return errors.New("duplicate too times")
}

func (p *Discovery) RemoveNode(n *core.ServiceDiscoveryService) error {
	etcdKey := fmt.Sprintf("/%s/%s/%s", app.Organization, "discovery", n.ServiceName)

	log.Infof("try del node to %s", etcdKey)

	for i := 0; i < 5; i++ {
		value, version, err := p.client.GetWithVersion(etcdKey)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		var service core.ServiceDiscoveryService
		err = proto.Unmarshal(value, &service)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		for _, node := range n.NodeList {
			(&service).RemoveNode(node)
		}

		ok, err := p.client.SetPbVersion(etcdKey, &service, version)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		if ok {
			return nil
		}
	}

	return errors.New("duplicate too times")
}

func (p *Discovery) GetServer(name string) (*core.ServiceDiscoveryService, error) {
	service, ok := p.getServerLocal(name)
	if ok {
		return service, nil
	}

	service, err := p.getServerDisk(name)
	if err == nil {
		return service, nil
	} else if os.IsNotExist(err) {
		return nil, xerror.New(int32(core.ErrCode_ServerNodeNotFound))
	}

	return nil, err
}

func (p *Discovery) removeNodeLocal(name string) error {
	p.cacheLock.Lock()
	defer p.cacheLock.Unlock()

	delete(p.cache, name)

	return nil
}

func (p *Discovery) getServerLocal(name string) (*core.ServiceDiscoveryService, bool) {
	p.cacheLock.RLock()
	defer p.cacheLock.RUnlock()

	service, ok := p.cache[name]
	return service, ok
}

func (p *Discovery) getServerDisk(name string) (*core.ServiceDiscoveryService, error) {
	p.cacheLock.Lock()
	defer p.cacheLock.Unlock()

	buffer, err := os.ReadFile(filepath.Join(runtime.LazyConfigDir(), "discovery", name))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	var service core.ServiceDiscoveryService
	err = proto.Unmarshal(buffer, &service)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	p.cache[name] = &service

	_ = p.watchServer(name)

	return &service, nil
}

func (p *Discovery) reloadCache(name string) error {
	service, err := p.getServerDisk(name)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.cache[name] = service

	return nil
}

func (p *Discovery) watchServer(name string) error {
	name = filepath.Join(runtime.LazyConfigDir(), "discovery", name)

	if candy.Contains(p.watcher.WatchList(), name) {
		return nil
	}

	err := p.watcher.Add(name)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func NewDiscover(client *etcd.Client) *Discovery {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("err:%v", err)
		return nil
	}

	p := &Discovery{
		client:    client,
		watcher:   watcher,
		cacheLock: sync.RWMutex{},
		cache:     map[string]*core.ServiceDiscoveryService{},
	}

	routine.Go(func() (err error) {
		for event := range watcher.Events {
			name := filepath.Base(event.Name)

			switch event.Op {
			case fsnotify.Create, fsnotify.Write:
				log.Warnf("%s service changed", name)
				err = p.reloadCache(name)
				if err != nil {
					log.Errorf("err:%v", err)
				}

			case fsnotify.Remove:
				log.Warnf("%s service removed", name)
				err = p.removeNodeLocal(name)
				if err != nil {
					log.Errorf("err:%v", err)
				}

			default:
				// do nothing
			}
		}

		return nil
	})

	return p
}

// 考虑全局适用的场景
var defaultDiscovery = NewDiscover(nil)

func SetEtcd(client *etcd.Client) {
	if defaultDiscovery == nil {
		defaultDiscovery = NewDiscover(client)
	} else {
		defaultDiscovery.client = client
	}
}
