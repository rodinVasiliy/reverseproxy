package webapp

import (
	"context"
	"fmt"
	"reverseproxy/internal/domain/ssl"
	"reverseproxy/internal/domain/webapp"
)

type AppWebappSyncService struct {
	sslService *ssl.Service
}

func NewWebappSyncService(sslService *ssl.Service) *AppWebappSyncService {
	return &AppWebappSyncService{
		sslService: sslService,
	}
}

func (s *AppWebappSyncService) OnCreate(ctx context.Context, app webapp.WebApp) {
	sslConfiguration, err := s.sslService.FindByID(ctx, app.SSLId)
	if err != nil {
		fmt.Println(err)
		return
	}

	config := generateNginxConfig(app, sslConfiguration.CertFileName, sslConfiguration.KeyFileName)
	createNginxFiles(app, config)
}

func (s *AppWebappSyncService) OnUpdate(ctx context.Context, app webapp.WebApp) {
	sslConfiguration, err := s.sslService.FindByID(ctx, app.SSLId)
	if err != nil {
		fmt.Println(err)
		return
	}

	config := generateNginxConfig(app, sslConfiguration.CertFileName, sslConfiguration.KeyFileName)
	editNginxFiles(app, config)
}

func (s *AppWebappSyncService) OnDelete(ctx context.Context, app webapp.WebApp) {
	deleteNginxFiles(app)
}
