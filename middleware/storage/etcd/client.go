package etcd

import (
	"context"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/anyx"
	"github.com/lazygophers/utils/json"
	"github.com/lazygophers/utils/routine"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

type Client struct {
	cli *clientv3.Client

	eventPool *sync.Pool

	watchOnce sync.Once
}

func NewClient(c *Config) (*Client, error) {
	if c == nil {
		c = &Config{}
	}

	log.Infof("connecting etcd...")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:             c.Address,
		AutoSyncInterval:      time.Second,
		DialTimeout:           5 * time.Second,
		DialKeepAliveTime:     time.Second,
		DialKeepAliveTimeout:  time.Second,
		MaxCallSendMsgSize:    0,
		MaxCallRecvMsgSize:    0,
		TLS:                   nil,
		Username:              c.Username,
		Password:              c.Password,
		RejectOldCluster:      true,
		DialOptions:           nil,
		Context:               nil,
		Logger:                zap.NewNop(),
		LogConfig:             nil,
		PermitWithoutStream:   true,
		MaxUnaryRetries:       0,
		BackoffWaitBetween:    0,
		BackoffJitterFraction: 0,
	})
	if err != nil {
		log.Errorf("err:%v", err)
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
					e.Value = event.Kv.Value

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

func (p *Client) SetWithVersion(key string, val interface{}, version int64) (bool, error) {
	res, err := p.cli.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.Version(key), "=", version)).
		Then(clientv3.OpPut(key, anyx.ToString(val))).
		Commit()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return res.Succeeded, nil
}

func (p *Client) SetPb(key string, val proto.Message) error {
	buffer, err := proto.Marshal(val)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	_, err = p.cli.Put(context.TODO(), key, string(buffer))
	if err != nil {
		return err
	}

	return nil
}

func (p *Client) SetPbVersion(key string, val proto.Message, version int64) (bool, error) {
	buffer, err := proto.Marshal(val)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	res, err := p.cli.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.Version(key), "=", version)).
		Then(clientv3.OpPut(key, string(buffer))).
		Commit()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return res.Succeeded, nil
}

func (p *Client) SetWithLock(key string, val interface{}) error {
	lock := NewMutex(p, "lock/"+key)
	return lock.WhenLocked(func() error {
		return p.Set(key, val)
	})
}

func (p *Client) WhenLocked(key string, logic func(client *Client, key string) error) error {
	lock := NewMutex(p, "lock/"+key)
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

func (p *Client) GetWithVersion(key string) ([]byte, int64, error) {
	resp, err := p.cli.Get(context.TODO(), key)
	if err != nil {
		return nil, 0, err
	}

	for _, kv := range resp.Kvs {
		if string(kv.Key) != key {
			continue
		}
		return kv.Value, kv.Version, nil
	}

	return nil, 0, ErrNotFound
}

func (p *Client) Prefix(key string) (map[string][]byte, error) {
	resp, err := p.cli.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	m := make(map[string][]byte, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		m[string(kv.Key)] = kv.Value
	}

	return m, nil
}

func (p *Client) ListPrefix(key string) ([]string, error) {
	resp, err := p.cli.Get(context.TODO(), key,
		clientv3.WithPrefix(),
		clientv3.WithIgnoreValue(),
	)
	if err != nil {
		return nil, err
	}

	keys := make([]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		keys = append(keys, string(kv.Key))
	}

	return keys, nil
}

func (p *Client) GetJson(key string, j interface{}) error {
	val, err := p.Get(key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(val, j)
	if err != nil {
		return err
	}

	return nil
}

func (p *Client) GetPb(key string, j proto.Message) error {
	val, err := p.Get(key)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(val, j)
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
