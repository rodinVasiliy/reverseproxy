package proxy

import (
	"context"
	"fmt"
	"net/http/httputil"
	"net/url"
	webApp "reverseproxy/internal/domain/webapp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Manager key - webapp id, value - proxy with upstream
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
	for _, wa := range webApps {
		proxy, err := newProxyForUpstream(wa.Upstream)
		if err != nil {
			return nil, fmt.Errorf("failed to get Proxy for webApp %s", err)
		}
		resultMap[wa.ID] = proxy
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

func (manager *Manager) GetProxyForWebApp(webApp *webApp.WebApp) (*httputil.ReverseProxy, bool) {
	val, ok := manager.proxies[webApp.ID]
	return val, ok
}

func (manager *Manager) SetProxyToManager(webapp *webApp.WebApp) error {
	proxy, err := newProxyForUpstream(webapp.Upstream)
	if err != nil {
		return fmt.Errorf("failed to set proxy for upstream %s", err)
	}
	manager.proxies[webapp.ID] = proxy
	return nil
}

func (manager *Manager) DeleteProxyFromManager(webapp *webApp.WebApp) {
	manager.proxies[webapp.ID] = nil
}
