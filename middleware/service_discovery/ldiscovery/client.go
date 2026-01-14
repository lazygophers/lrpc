package ldiscovery

import (
	"math/rand"
	"net"
	"net/url"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/lazygophers/utils/app"
	"github.com/valyala/fasthttp"
)

func GetServer(name string) (*core.ServiceDiscoveryService, error) {
	return defaultDiscovery.GetServer(name)
}

func GetAllNode(name string) ([]*core.ServiceDiscoveryNode, error) {
	server, err := GetServer(name)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return server.NodeList, nil
}

func ChooseNode(name string) (*core.ServiceDiscoveryNode, error) {
	nodeList, err := GetAllNode(name)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if len(nodeList) == 0 {
		return nil, xerror.New(int32(core.ErrCode_ServerNodeNotFound))
	}

	if len(nodeList) == 1 {
		return nodeList[0], nil
	}

	seed := rand.Int()
	l := len(nodeList)
	for i := 0; i < l; i++ {
		node := nodeList[(i+seed)%l]
		if !node.Alive {
			continue
		}

		return node, nil
	}

	return nil, xerror.New(int32(core.ErrCode_ServerAliveNodeNotFound))
}

func DiscoveryClient(c *core.ServiceDiscoveryClient) (lrpc.Client, *fasthttp.Request) {
	client := &fasthttp.HostClient{
		Addr:                     c.ServiceName,
		Name:                     app.Name,
		NoDefaultUserAgentHeader: true,
		Dial: func(addr string) (net.Conn, error) {
			node, err := ChooseNode(addr)
			if err != nil {
				log.Errorf("err:%v", err)
				return nil, err
			}

			return net.Dial("tcp", net.JoinHostPort(node.Host, node.Port))
		},
		Transport: nil,
	}

	u := &url.URL{
		Scheme: "lrpc",
		Host:   c.ServiceName,
		Path:   c.ServicePath,
	}

	reqeust := &fasthttp.Request{}
	reqeust.SetRequestURI(u.String())
	if c.Method == "" {
		reqeust.Header.SetMethod("POST")
	} else {
		reqeust.Header.SetMethod(c.Method)
	}

	if c.Timeout != nil {
		reqeust.SetTimeout(c.Timeout.AsDuration())
	} else {
		reqeust.SetTimeout(time.Second * 30)
	}

	return client, reqeust
}
