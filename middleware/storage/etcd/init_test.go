package etcd_test

import (
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/etcd"
	"sync"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	err := etcd.Connect(&etcd.Config{
		Address: []string{"http://127.0.0.1:2379"},
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}

	var w sync.WaitGroup
	etcd.WatchPrefix("test", &w, func(event *etcd.Event) {
		log.Info(event.Key)
		log.Info(event.Type)
	})

	for i := 0; i < 10; i++ {
		etcd.Set(fmt.Sprintf("test-%d", i), time.Now().String())
		time.Sleep(time.Second * 3)
	}
}
