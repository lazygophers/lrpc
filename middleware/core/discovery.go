package core

import (
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/candy"
)

func (p *ServiceDiscoveryNode) key() string {
	switch p.Type {
	case ServiceType_Service:
		return fmt.Sprintf("%s:%s", p.Host, p.Port)

	default:
		log.Panicf("service type %d not supported", p.Type)
		return ""
	}
}

func (p *ServiceDiscoveryService) MergeNode(node *ServiceDiscoveryNode) {
	key := node.key()

	if candy.ContainsUsing(p.NodeList, func(v *ServiceDiscoveryNode) bool {
		return key == v.key()
	}) {
		return
	}

	p.NodeList = append(p.NodeList, node)
	return
}

func (p *ServiceDiscoveryService) RemoveNode(node *ServiceDiscoveryNode) {
	key := node.key()

	p.NodeList = candy.Filter(p.NodeList, func(v *ServiceDiscoveryNode) bool {
		return key == v.key()
	})

	return
}
