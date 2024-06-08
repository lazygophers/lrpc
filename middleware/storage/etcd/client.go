package etcd

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/lazygophers/utils/anyx"
	"github.com/lazygophers/utils/routine"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Client struct {
	cli *clientv3.Client

	eventPool *sync.Pool
}

func NewClient(c *Config) (*Client, error) {
	if c == nil {
		c = &Config{}
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            c.Address,
		AutoSyncInterval:     time.Second,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    time.Second,
		DialKeepAliveTimeout: time.Second,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  nil,
		Username:             "",
		Password:             "",
		RejectOldCluster:     true,
		DialOptions:          nil,
		Context:              nil,
		Logger:               zap.NewNop(),
		LogConfig:            nil,
		PermitWithoutStream:  true,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		cli: cli,
		eventPool: &sync.Pool{
			New: func() interface{} {
				return new(Event)
			},
		},
	}, nil
}

func (p *Client) Close() error {
	return p.cli.Close()
}

func (p *Client) Watch(key string, logic func(event *Event)) {
	p.watch(p.cli.Watch(context.Background(), key, clientv3.WithKeysOnly()), logic)
}

func (p *Client) WatchPrefix(key string, logic func(event *Event)) {
	p.watch(p.cli.Watch(context.Background(), key, clientv3.WithPrefix()), logic)
}

func (p *Client) watch(wc clientv3.WatchChan, logic func(event *Event)) {
	routine.GoWithRecover(func() (err error) {
		for {
			select {
			case v := <-wc:
				for _, event := range v.Events {
					e := p.eventPool.Get().(*Event)
					e.Key = string(event.Kv.Key)
					e.Value = string(event.Kv.Value)

					e.Version = event.Kv.Version

					switch event.Type {
					case mvccpb.PUT:
						e.Type = Changed
					case mvccpb.DELETE:
						e.Type = Deleted
					}

					logic(e)
					p.eventPool.Put(e)
				}
			}
		}
	})
}

func (p *Client) Del(key string) error {
	_, err := p.cli.Delete(context.TODO(), key)
	if err != nil {
		return err
	}

	return nil
}

func (p *Client) Set(key string, val interface{}) error {
	_, err := p.cli.Put(context.TODO(), key, anyx.ToString(val))
	if err != nil {
		return err
	}

	return nil
}

func (p *Client) SetWithLock(key string, val interface{}) error {
	lock := NewMutexWithClient(cli, "lock/"+key)
	return lock.WhenLocked(func() error {
		return p.Set(key, val)
	})
}

func (p *Client) WhenLocked(key string, logic func(client *Client, key string) error) error {
	lock := NewMutexWithClient(cli, "lock/"+key)
	return lock.WhenLocked(func() error {
		return logic(p, key)
	})
}

func (p *Client) SetEx(key string, val interface{}, timeout time.Duration) error {
	lease := clientv3.NewLease(p.cli)
	defer lease.Close()

	leaseRsp, err := lease.Grant(context.TODO(), int64(timeout.Seconds()))
	if err != nil {
		return err
	}

	_, err = p.cli.Put(context.TODO(), key, anyx.ToString(val), clientv3.WithLease(leaseRsp.ID))
	if err != nil {
		return err
	}

	return nil
}

func (p *Client) Get(key string) ([]byte, error) {
	resp, err := p.cli.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	for _, kv := range resp.Kvs {
		if string(kv.Key) != key {
			continue
		}
		return kv.Value, nil
	}

	return nil, ErrNotFound
}

func (p *Client) Prefix(key string) (map[string][]byte, error) {
	resp, err := p.cli.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	m := make(map[string][]byte, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		m[string(kv.Key)] = kv.Value
	}

	return m, ErrNotFound
}

func (p *Client) GetJson(key string, j interface{}) error {
	val, err := p.Get(key)
	if err != nil {
		return err
	}

	err = sonic.Unmarshal(val, j)
	if err != nil {
		return err
	}

	return nil
}

func (p *Client) Exist(key string) (bool, error) {
	resp, err := p.cli.Get(context.TODO(), key)
	if err != nil {
		return false, err
	}

	for _, kv := range resp.Kvs {
		if string(kv.Key) != key {
			continue
		}
		return true, nil
	}

	return false, nil
}
