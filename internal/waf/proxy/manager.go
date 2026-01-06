package wafconfig

import (
	"context"
	"fmt"
	"net/http/httputil"
	"net/url"
	webApp "reverseproxy/internal/domain/webapp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// key - webapp id, value - proxy with upstream
type Manager struct {
	proxies map[primitive.ObjectID]*httputil.ReverseProxy
}

func NewManager(webAppService *webApp.Service) (*Manager, error) {
	ctx := context.Background()
	webApps, err := webAppService.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load waf config %s", err)
	}
	resultMap := make(map[primitive.ObjectID]*httputil.ReverseProxy)
	for _, webApp := range webApps {
		proxy, err := newProxyForUpstream(webApp.Upstream)
		if err != nil {
			return nil, fmt.Errorf("failed to get Proxy for webApp %s", err)
		}
		resultMap[webApp.ID] = proxy
	}
	return &Manager{proxies: resultMap}, nil
}

func newProxyForUpstream(upstream string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(upstream)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream %s to url %s", upstream, err)
	}
	return httputil.NewSingleHostReverseProxy(url), nil
}

func (manager *Manager) GetProxyForWebApp(webApp *webApp.WebApp) *httputil.ReverseProxy {
	return manager.proxies[webApp.ID]
}
