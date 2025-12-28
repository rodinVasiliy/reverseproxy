package wafconfig

import (
	"context"
	"fmt"
	"net/http/httputil"
	"net/url"
	webApp "reverseproxy/internal/model/webapp"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// key - domain value - webapplication
type WafConfig struct {
	proxies map[primitive.ObjectID]*httputil.ReverseProxy
}

func NewWafConfig(webAppService *webApp.Service) (*WafConfig, error) {
	ctx := context.Background()
	webApps, err := webAppService.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load waf config %s", err)
	}
	resultMap := make(map[primitive.ObjectID]*httputil.ReverseProxy)
	for _, webApp := range webApps {
		proxy, err := newProxyForUpstream(webApp.Upstream + ":" + strconv.Itoa(webApp.Port))
		if err != nil {
			return nil, fmt.Errorf("failed to get Proxy for webApp %s", err)
		}
		resultMap[webApp.ID] = proxy
	}
	return &WafConfig{proxies: resultMap}, nil
}

func newProxyForUpstream(upstream string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(upstream)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream %s to url %s", upstream, err)
	}
	return httputil.NewSingleHostReverseProxy(url), nil
}

func (wafConfig *WafConfig) GetProxyForWebApp(webApp *webApp.WebApp) *httputil.ReverseProxy {
	return wafConfig.proxies[webApp.ID]
}
