package webapp

import (
	"context"
	"fmt"
	"reverseproxy/internal/model/policy"
	ssl "reverseproxy/internal/model/ssl"
	repository "reverseproxy/internal/repository"
	"reverseproxy/internal/response/webapp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository    *repository.MongoRepository[WebApp]
	sslService    *ssl.Service
	policyService *policy.Service
}

func NewService(repo *repository.MongoRepository[WebApp], sslService *ssl.Service) *Service {
	return &Service{repository: repo, sslService: sslService}
}

func (s *Service) FindAll(ctx context.Context) ([]WebApp, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) List(ctx context.Context) ([]webapp.WebAppResponse, error) {
	webapps, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]webapp.WebAppResponse, 0, len(webapps))
	for _, w := range webapps {
		policy, err := s.policyService.FindById(ctx, w.PolicyId)
		if err != nil {
			return nil, fmt.Errorf("failed to find policy for webapp %v: %s", w.Name, err)
		}
		ssl, err := s.sslService.FindByID(ctx, w.SSLId)
		if err != nil {
			return nil, fmt.Errorf("failed to find ssl config for webapp %v: %s", w.Name, err)
		}
		responses = append(responses, webapp.WebAppResponse{
			ID:         w.ID.Hex(),
			Name:       w.Name,
			PolicyId:   policy.ID.Hex(),
			PolicyName: policy.Name,
			SSLId:      ssl.ID.Hex(),
			SSLName:    ssl.Name,
		})
	}
	return responses, nil
}

func (s *Service) Insert(ctx context.Context, app WebApp) (primitive.ObjectID, error) {
	// находим ssl конфигурацию чтобы по ней создать nginx файлы
	ssl, err := s.sslService.FindByID(ctx, app.SSLId)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to find ssl for web app %w", err)
	}
	id, err := s.repository.Insert(ctx, app)
	// TODO а если не primitive
	if err != nil {
		return primitive.NilObjectID, err
	}
	app.ID = id
	nginxConfig := generateNginxConfig(app, ssl.CertFileName, ssl.KeyFileName)
	createNginxFiles(app, nginxConfig)
	return id, nil
}

func (s *Service) Delete(ctx context.Context, app *WebApp) error {
	deleteNginxFiles(*app)
	return s.repository.Delete(ctx, app)
}

func (s *Service) Edit(ctx context.Context, app *WebApp) error {
	ssl, err := s.sslService.FindByID(ctx, app.SSLId)
	if err != nil {
		return fmt.Errorf("failed to find ssl for web app %w", err)
	}
	nginxConfig := generateNginxConfig(*app, ssl.CertFileName, ssl.KeyFileName)
	editNginxFiles(*app, nginxConfig)
	return s.repository.Update(ctx, app)
}

func (s *Service) GetWebAppForHost(ctx context.Context, host string) (*WebApp, error) {
	return s.repository.FindOne(ctx, bson.M{"hosts": host})
}

func (s *Service) FindById(ctx context.Context, id primitive.ObjectID) (*WebApp, error) {
	return s.repository.FindById(ctx, id)
}
