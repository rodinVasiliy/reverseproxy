package config

import (
	"net"
	"net/http"
)

type RequestParams struct {
	ip   net.IP
	host string
	path string
}

type ActionParams struct {
	rule string
	rp   *RequestParams
}

func ParseRequest(r *http.Request) *RequestParams {
	rp := RequestParams{
		ip:   getIpFromRequest(r),
		host: r.Host,
		path: r.URL.RawPath,
	}
	return &rp
}
