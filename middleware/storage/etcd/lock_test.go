package etcd_test

import (
	"github.com/lazygophers/lrpc/middleware/etcd"
	"sync"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	err := etcd.Connect(&etcd.Config{
		Address: []string{"http://127.0.0.1:2379"},
	})
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	var w sync.WaitGroup
	logic := func(i int) {
		defer w.Done()

		lock := etcd.NewMutex("test")
		err := lock.Lock()
		if err != nil {
			t.Errorf("err:%v", err)
			return
		}

		// if i%2 == 0 {
		// 	return
		// }

		time.Sleep(time.Second * 20)
		t.Log(i)

		err = lock.Unlock()
		if err != nil {
			t.Errorf("err:%v", err)
			return
		}
	}

	for i := 0; i < 5; i++ {
		w.Add(1)
		go logic(i)
	}

	w.Wait()
}

func TestWhenLock(t *testing.T) {
	err := etcd.Connect(&etcd.Config{
		Address: []string{"http://127.0.0.1:2379"},
	})
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	lock := etcd.NewMutex("test")
	err = lock.Lock()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	err = lock.Lock()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Log(time.Now().String())

	err = lock.Lock()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Log(time.Now().String())

	err = lock.Lock()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Log(time.Now().String())
}
