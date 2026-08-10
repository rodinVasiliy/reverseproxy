package initialization

import (
	"context"
	"fmt"
	policy "reverseproxy/internal/domain/policy"
	ssl "reverseproxy/internal/domain/ssl"
	webapp "reverseproxy/internal/domain/webapp"
	"reverseproxy/internal/waf/proxy"
)

func NewTestWebApp(policyService *policy.Service, sslService *ssl.Service, webappService *webapp.Service, manager *proxy.Manager) error {
	p, err := policyService.FindByName(context.Background(), DEFAULT_POLICY_NAME)
	if err != nil {
		return fmt.Errorf("failed to get default policy %w", err)
	}

	certFileName := "fullchain.pem"
	keyFileName := "privkey.pem"
	sslConfig := ssl.SSL{Name: "myproxytest.site",
		CertFileName: certFileName, KeyFileName: keyFileName}
	sslId, err := sslService.Insert(context.Background(), sslConfig)
	if err != nil {
		return fmt.Errorf("failed to add test ssl config %w", err)
	}
	host := "myproxytest.site"
	port := 443
	webApp := webapp.WebApp{Name: "test", SSLId: sslId, PolicyId: p.ID, Port: port, Upstream: "http://92.168.11.202:9091", Hosts: []string{host}}
	id, err := webappService.Insert(context.Background(), webApp)
	if err != nil {
		return fmt.Errorf("failed to add test webapp %w", err)
	}
	webApp.ID = id
	err = manager.SetProxyToManager(&webApp)
	if err != nil {
		return fmt.Errorf("failed to set proxy for manager: %v", err)
	}
	fmt.Println("New test WebApp successfully added")
	return nil
}
