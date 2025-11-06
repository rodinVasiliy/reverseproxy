package wafconfig

import (
	"fmt"
	"net/http/httputil"

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

}
