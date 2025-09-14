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
		// Try to find a network interface (eth0 on Linux, en0 on macOS, etc.)
		interfaceNames := []string{"eth0", "en0", "ens33", "enp0s3"}
		found := false
		for _, name := range interfaceNames {
			address, err := net.InterfaceByName(name)
			if err == nil {
				t.Log(address.Name)
				addrs, _ := address.Addrs()
				t.Log(addrs)
				found = true
				break
			}
		}
		if !found {
			t.Log("No common network interface found, skipping interface-specific test")
		}
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			t.Log("ip:", ipnet.IP.String())
		}
	}
}
