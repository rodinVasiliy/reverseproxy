package config

import (
	"net"
	"net/http"
)

func getIpFromRequest(r *http.Request) net.IP {
	xrip := r.Header.Get("X-Real-IP")
	return net.ParseIP(xrip)
}

func checkInList(list []*net.IPNet, ip net.IP) bool {
	for _, net := range list {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}
