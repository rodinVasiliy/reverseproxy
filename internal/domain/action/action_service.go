package action

import (
	"context"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository *repository.MongoRepository[ActionDoc]
}

func NewService(repo *repository.MongoRepository[ActionDoc]) *Service {
	return &Service{repository: repo}
}

func (s *Service) Insert(ctx context.Context, actionDoc ActionDoc) (primitive.ObjectID, error) {
	return s.repository.Insert(ctx, actionDoc)
}

func (s *Service) FindAll(ctx context.Context) ([]ActionDoc, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) FindByIds(ctx context.Context, ids []primitive.ObjectID) ([]ActionDoc, error) {
	filter := bson.M{"_id": bson.M{"$in": ids}}
	return s.repository.FindMany(ctx, filter)
}

func (s *Service) FindById(ctx context.Context, id primitive.ObjectID) (*ActionDoc, error) {
	return s.repository.FindById(ctx, id)
}
