package config

import (
	"net"
	"net/http"
)

type RequestParams struct {
	ip      net.IP
	host    string
	path    string
	ua      string
	cookies []*http.Cookie
}

type ActionParams struct {
	rule string
	rp   *RequestParams
}

func ParseRequest(r *http.Request) *RequestParams {
	rp := RequestParams{
		ip:      getIpFromRequest(r),
		host:    r.Host,
		path:    r.URL.RawPath,
		ua:      r.Header.Get("User-Agent"),
		cookies: r.Cookies(),
	}
	return &rp
}
