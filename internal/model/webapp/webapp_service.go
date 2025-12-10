package webapp

import (
	"context"
	"fmt"
	ssl "reverseproxy/internal/model/ssl"
	repository "reverseproxy/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository *repository.MongoRepository[WebApp]
	sslService *ssl.Service
}

func NewService(repo *repository.MongoRepository[WebApp]) *Service {
	return &Service{repository: repo}
}

func (s *Service) FindAll(ctx context.Context) ([]WebApp, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) Insert(ctx context.Context, app WebApp) (primitive.ObjectID, error) {
	ssl, err := s.sslService.GetByID(ctx, app.ID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to find ssl for web app %w", err)
	}
	id, err := s.repository.Insert(ctx, app)
	// TODO а если не primitive
	if err != nil {
		return primitive.NilObjectID, err
	}
	app.ID = id
	nginxConfig := generateNginxConfig(app, ssl.CertPath, ssl.KeyPath)
	createNginxFiles(app, nginxConfig)
	return id, nil
}

func (s *Service) Delete(ctx context.Context, app *WebApp) error {
	deleteNginxFiles(*app)
	return s.repository.Delete(ctx, app)
}

func (s *Service) Edit(ctx context.Context, app *WebApp) error {
	ssl, err := s.sslService.GetByID(ctx, app.ID)
	if err != nil {
		return fmt.Errorf("failed to find ssl for web app %w", err)
	}
	nginxConfig := generateNginxConfig(*app, ssl.CertPath, ssl.KeyPath)
	editNginxFiles(*app, nginxConfig)
	return s.repository.Update(ctx, app)
}

func (s *Service) GetWebAppForHost(ctx context.Context, host string) (*WebApp, error) {
	return s.repository.FindOne(ctx, bson.M{"hosts": host})
}
