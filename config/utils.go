package config

import (
	"net"
	"net/http"
)

func getIpFromRequest(r *http.Request) net.IP {
	ipStr, _, _ := net.SplitHostPort(r.RemoteAddr)
	return net.ParseIP(ipStr)
}

func checkInList(list []*net.IPNet, ip net.IP) bool {
	for _, net := range list {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}
