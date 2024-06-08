package etcd

import (
	"context"
	"errors"
	"github.com/lazygophers/log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Mutex struct {
	key     string
	lock    *concurrency.Mutex
	leaseId clientv3.LeaseID

	cli *Client
}

func NewMutex(key string) *Mutex {
	return NewMutexWithClient(cli, key)
}

func NewMutexWithClient(cli *Client, key string) *Mutex {
	p := Mutex{
		key: key,
		cli: cli,
	}

	return &p
}

func (p *Mutex) Lock() error {
	lease := clientv3.NewLease(p.cli.cli)
	defer lease.Close()

	leaseRsp, err := lease.Grant(context.TODO(), 3)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.leaseId = leaseRsp.ID

	session, err := concurrency.NewSession(p.cli.cli, concurrency.WithLease(p.leaseId))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.lock = concurrency.NewMutex(session, p.key)

	err = p.lock.Lock(context.TODO())
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *Mutex) LockWithTimeout(timeout time.Duration) error {
	lease := clientv3.NewLease(p.cli.cli)
	defer lease.Close()

	leaseRsp, err := lease.Grant(context.TODO(), 3)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.leaseId = leaseRsp.ID

	session, err := concurrency.NewSession(p.cli.cli, concurrency.WithLease(p.leaseId))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.lock = concurrency.NewMutex(session, p.key)

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		err = p.lock.TryLock(context.TODO())
		if err != nil {
			if err != concurrency.ErrLocked {
				log.Errorf("err:%v", err)
				return err
			}

		} else {
			return nil
		}

		select {
		case <-time.After(time.Millisecond * 100):
			break
		case <-timer.C:
			return errors.New("lock timeout")
		}
	}
}

func (p *Mutex) TryLock() error {
	lease := clientv3.NewLease(p.cli.cli)
	defer lease.Close()

	leaseRsp, err := lease.Grant(context.TODO(), 3)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.leaseId = leaseRsp.ID

	session, err := concurrency.NewSession(p.cli.cli, concurrency.WithLease(p.leaseId))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	p.lock = concurrency.NewMutex(session, p.key)

	err = p.lock.TryLock(context.TODO())
	if err != nil {
		if err != concurrency.ErrLocked {
			log.Errorf("err:%v", err)
			return err
		}
	} else {
		return nil
	}

	return errors.New("try lock timeout")
}

func (p *Mutex) Unlock() error {
	lease := clientv3.NewLease(p.cli.cli)
	defer lease.Close()

	_, err := lease.Revoke(context.TODO(), p.leaseId)
	if err != nil {
		log.Errorf("err:%v", err)
	}

	err = p.lock.Unlock(context.TODO())
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *Mutex) WhenLocked(logic func() error) error {
	err := p.Lock()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	e := logic()
	err = p.Unlock()
	if e != nil {
		log.Errorf("err:%v", e)
		return e
	}
	if err != nil {
		log.Errorf("err:%v", err)
	}

	return nil
}
