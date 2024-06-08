package lrpc_test

import (
	"net"
	"testing"
)

func TestInnerIp(t *testing.T) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Errorf("err:%s", err)
		return
	}

	{
		address, err := net.InterfaceByName("eth0")
		if err != nil {
			t.Errorf("err:%v", err)
		} else {
			t.Log(address.Name)
			t.Log(address.Addrs())
		}
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			t.Log("ip:", ipnet.IP.String())
		}
	}
}
