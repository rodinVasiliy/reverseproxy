package policy

import (
	"context"
	"log"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository     *repository.MongoRepository[Policy]
	webappProvider WebappProvider
}

func NewService(repo *repository.MongoRepository[Policy]) *Service {
	return &Service{repository: repo}
}

func (s *Service) SetWebappProvider(webappProvider WebappProvider) {
	s.webappProvider = webappProvider
}

func (s *Service) Insert(ctx context.Context, policy Policy) (primitive.ObjectID, error) {
	return s.repository.Insert(ctx, policy)
}

func (s *Service) FindByName(ctx context.Context, name string) (*Policy, error) {
	return s.repository.FindOne(ctx, bson.M{"name": name})
}

func (s *Service) FindById(ctx context.Context, id primitive.ObjectID) (*Policy, error) {
	return s.repository.FindById(ctx, id)
}

func (s *Service) FindAll(ctx context.Context) ([]Policy, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	policies, err := s.repository.FindAll(ctx)
	if err != nil {
		log.Printf("failed to find webapps: %v", err)
		return nil, err
	}
	responses := make([]Response, 0, len(policies))
	for _, policy := range policies {
		webapps, err := s.webappProvider.FindByPolicyId(policy.ID, ctx)
		if err != nil {
			log.Printf("failed to find webapps: %v", err)
			continue
		}
		responses = append(responses, Response{ID: policy.ID.Hex(), Name: policy.Name, Webapps: webapps})
	}
	return responses, nil
}
