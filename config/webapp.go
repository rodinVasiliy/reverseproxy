package config

import (
	"net/http/httputil"
	"net/url"
)

type WebApp struct {
	pol              *Policy
	upstream         *url.URL
	sslConfiguration *SSLConfiguration
	proxy            *httputil.ReverseProxy
}
