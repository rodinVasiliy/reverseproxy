package initialization

import (
	"fmt"
	ssl "reverseproxy/model/ssl_config"
	webapp "reverseproxy/model/web_app"
	service "reverseproxy/service"
)

func NewTestWebApp(ps *service.PolicyService, sslS *service.SSLConfigurationService, ws *service.WebAppService) error {
	policy, err := ps.FindByName(service.DEFAULT_POLICY_NAME)
	if err != nil {
		return fmt.Errorf("failed to get default policy %w", err)
	}

	// это пока не нужно, ssl в nginx еще не пробрасывается
	sslConfig := ssl.SSLConfiguration{Name: "myproxytest.site",
		CertPath: "/myproxytest.site/fullchain.pem", KeyPath: "/myproxytest.site/privkey.pem"}
	sslId, err := sslS.Add(&sslConfig)
	if err != nil {
		return fmt.Errorf("failed to add test ssl config %w", err)
	}

	webApp := webapp.WebApp{Name: "test", SSLId: sslId, PolicyId: policy.ID, Upstream: "http://localhost:9091"}
	_, err = ws.Add(&webApp)
	if err != nil {
		return fmt.Errorf("failed to add test webapp %w", err)
	}
	fmt.Println("New test WebApp successfully added")
	return nil
}
