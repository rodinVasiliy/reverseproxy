package ssl

import (
	"context"
	"reverseproxy/internal/domain/ssl"
	"reverseproxy/internal/domain/webapp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppSSLService struct {
	sslService    *ssl.Service
	webappService *webapp.Service
}

func NewAppSSLConfiguration(sslS *ssl.Service, webappS *webapp.Service) *AppSSLService {
	return &AppSSLService{
		sslService:    sslS,
		webappService: webappS,
	}
}

func (s *AppSSLService) CanDeleteSSL(ctx context.Context, id primitive.ObjectID) error {
	webapps, err := s.webappService.FindBySSLId(id, ctx)
	if err != nil {
		return err
	}
	if len(webapps) > 0 {
		return &ssl.SSLInUseError{Webapps: webapps}
	}
	return nil
}
