package config

import (
	"net"
)

func checkInList(list []*net.IPNet, ip net.IP) bool {
	for _, net := range list {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}
