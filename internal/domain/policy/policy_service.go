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

func (s *Service) Delete(ctx context.Context, entity *Policy) error {
	return s.repository.Delete(ctx, entity)
}

func (s *Service) Update(ctx context.Context, entity *Policy) error {
	return s.repository.Update(ctx, entity)
}

func (s *Service) FindOverrideForRule(ctx context.Context, ruleID primitive.ObjectID) ([]Policy, error) {
	filter := bson.M{
		"rules": bson.M{
			"$elemMatch": bson.M{
				"ruleId":    ruleID,
				"actions.0": bson.M{"$exists": true},
			},
		},
	}

	projection := bson.M{
		"name": 1,
		"rules": bson.M{
			"$elemMatch": bson.M{
				"ruleId": ruleID,
			},
		},
	}

	policies, err := s.repository.FindManyWithProjection(ctx, filter, projection)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (s *Service) FindByRuleID(ctx context.Context, ruleID primitive.ObjectID) ([]Policy, error) {
	filter := bson.M{
		"rules": bson.M{"$in": []primitive.ObjectID{ruleID}},
	}
	policies, err := s.repository.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	return policies, nil
}
