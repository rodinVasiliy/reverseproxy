package rule

import (
	"context"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository *repository.MongoRepository[Rule]
}

func NewService(repo *repository.MongoRepository[Rule]) *Service {
	return &Service{repository: repo}
}

func (s *Service) FindAll(ctx context.Context) ([]Rule, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) Insert(ctx context.Context, rule Rule) (primitive.ObjectID, error) {
	return s.repository.Insert(ctx, rule)
}

func (s *Service) FindById(ctx context.Context, id primitive.ObjectID) (*Rule, error) {
	return s.repository.FindById(ctx, id)
}

func (s *Service) Update(ctx context.Context, rule *Rule) error {
	return s.repository.Update(ctx, rule)
}

func (s *Service) FindByPolicyId(ctx context.Context, id primitive.ObjectID) ([]Rule, error) {
	filter := bson.M{
		"policies": id,
	}
	policies, err := s.repository.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	return policies, nil
}
