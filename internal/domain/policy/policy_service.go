package policy

import (
	"context"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository *repository.MongoRepository[Policy]
}

func NewService(repo *repository.MongoRepository[Policy]) *Service {
	return &Service{repository: repo}
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
