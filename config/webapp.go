package config

import (
	"fmt"
	"net/http/httputil"
	"net/url"
)

type WebApp struct {
	pol              *Policy
	upstream         *url.URL
	sslConfiguration *SSLConfiguration
	proxy            *httputil.ReverseProxy
}

func initWebApp() (*WebApp, error) {

	upstream, err := url.Parse("http://localhost:9091")
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream %s", err)
	}

	policy, err := DefaultPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to get policy %s", err)
	}

	// мб экранировать
	// TODO добавить экспорт этого конфига в nginx.conf
	ssl := SSLConfiguration{
		certPath: "/etc/letsencrypt/live/myproxytest.site/fullchain.pem",
		keyPath:  "/etc/letsencrypt/live/myproxytest.site/privkey.pem",
	}

	return &WebApp{policy, upstream, &ssl, httputil.NewSingleHostReverseProxy(upstream)}, nil
}
