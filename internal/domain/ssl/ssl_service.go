package ssl

import (
	"context"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository *repository.MongoRepository[SSL]
}

func NewService(repo *repository.MongoRepository[SSL]) *Service {
	return &Service{repository: repo}
}

func (s *Service) FindByID(ctx context.Context, id primitive.ObjectID) (*SSL, error) {
	return s.repository.FindById(ctx, id)
}

func (s *Service) Insert(ctx context.Context, ssl SSL) (primitive.ObjectID, error) {
	return s.repository.Insert(ctx, ssl)
}

func (s *Service) FindAll(ctx context.Context) ([]SSL, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) Delete(ctx context.Context, entity *SSL) error {
	return s.repository.Delete(ctx, entity)
}

func (s *Service) Update(ctx context.Context, entity *SSL) error {
	return s.repository.Update(ctx, entity)
}
