package ldiscovery

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/utils/app"
	"github.com/lazygophers/utils/runtime"
)

// 对于服务端来说，只需要支持服务注册
func OnListen(listen lrpc.ListenData) error {
	// 想看看全局指定的注册 IP
	host, _ := getHostByGlobalDefault()

	if host == "" {
		host = listen.Host
	}

	err := defaultDiscovery.AddNode(&core.ServiceDiscoveryService{
		ServiceName: app.Name,
		NodeList: []*core.ServiceDiscoveryNode{
			{
				Type: core.ServiceType_Service,
				Host: host,
				Port: listen.Port,
			},
		},
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func OnShutdown(listen lrpc.ListenData) {
	host, _ := getHostByGlobalDefault()

	if host == "" {
		host = listen.Host
	}

	err := defaultDiscovery.RemoveNode(&core.ServiceDiscoveryService{
		ServiceName: app.Name,
		NodeList: []*core.ServiceDiscoveryNode{
			{
				Type: core.ServiceType_Service,
				Host: host,
				Port: listen.Port,
			},
		},
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}

	return
}

func getHostByGlobalDefault() (string, error) {
	buf, err := os.ReadFile(filepath.Join(runtime.LazyConfigDir(), "host"))
	if err != nil {
		log.Debugf("err:%v", err)
		return "", err
	}

	buf = bytes.TrimSpace(buf)
	buf = bytes.ReplaceAll(buf, []byte("\n"), []byte(""))
	buf = bytes.ReplaceAll(buf, []byte("\t"), []byte(""))
	buf = bytes.ReplaceAll(buf, []byte("\r"), []byte(""))

	return string(buf), nil
}
