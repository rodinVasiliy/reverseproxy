package initialization

import (
	"context"
	"fmt"
	policy "reverseproxy/internal/model/policy"
	ssl "reverseproxy/internal/model/ssl"
	webapp "reverseproxy/internal/model/webapp"
)

func NewTestWebApp(ps *policy.Service, sslS *ssl.Service, ws *webapp.Service) error {
	policy, err := ps.FindByName(context.Background(), DEFAULT_POLICY_NAME)
	if err != nil {
		return fmt.Errorf("failed to get default policy %w", err)
	}

	certFileName := "fullchain.pem"
	keyFileName := "privkey.pem"
	sslConfig := ssl.SSLConfiguration{Name: "myproxytest.site",
		CertFileName: certFileName, KeyFileName: keyFileName}
	sslId, err := sslS.Insert(context.Background(), sslConfig)
	if err != nil {
		return fmt.Errorf("failed to add test ssl config %w", err)
	}
	host := "myproxytest.site"
	port := 4443
	webApp := webapp.WebApp{Name: "test", SSLId: sslId, PolicyId: policy.ID, Port: port, Upstream: "http://localhost:9091", Hosts: []string{host}}
	_, err = ws.Insert(context.Background(), webApp)
	if err != nil {
		return fmt.Errorf("failed to add test webapp %w", err)
	}
	fmt.Println("New test WebApp successfully added")
	return nil
}
