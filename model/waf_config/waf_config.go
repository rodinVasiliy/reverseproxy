package wafconfig

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"reverseproxy/model/webApp"
	service "reverseproxy/service"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// key - domain value - webapplication
type WafConfig struct {
	proxies map[primitive.ObjectID]*httputil.ReverseProxy
}

func NewWafConfig(webAppService *service.WebAppService) (*WafConfig, error) {
	webApps, err := webAppService.FindAllWebApps()
	if err != nil {
		return nil, fmt.Errorf("failed to load waf config %s", err)
	}
	var resultMap map[primitive.ObjectID]*httputil.ReverseProxy
	for _, webApp := range *webApps {
		proxy, err := newProxyForUpstream(webApp.Upstream)
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
