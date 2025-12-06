package ssl

import (
	"context"
	repository "reverseproxy/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository *repository.MongoRepository[SSLConfiguration]
}

func NewService(repo *repository.MongoRepository[SSLConfiguration]) *Service {
	return &Service{repository: repo}
}

func (s *Service) GetByID(ctx context.Context, id primitive.ObjectID) (*SSLConfiguration, error) {
	return s.repository.FindById(ctx, id)
}

func (s *Service) Insert(ctx context.Context, ssl SSLConfiguration) (primitive.ObjectID, error) {
	return s.repository.Insert(ctx, ssl)
}
