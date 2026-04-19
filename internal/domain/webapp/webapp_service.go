package webapp

import (
	"context"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository *repository.MongoRepository[WebApp]
}

func NewService(repo *repository.MongoRepository[WebApp]) *Service {
	return &Service{repository: repo}
}

func (s *Service) FindAll(ctx context.Context) ([]WebApp, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) Insert(ctx context.Context, app WebApp) (primitive.ObjectID, error) {
	id, err := s.repository.Insert(ctx, app)
	// TODO а если не primitive
	if err != nil {
		return primitive.NilObjectID, err
	}
	app.ID = id
	return id, nil
}

func (s *Service) Delete(ctx context.Context, app *WebApp) error {
	return s.repository.Delete(ctx, app)
}

func (s *Service) Edit(ctx context.Context, app *WebApp) error {
	return s.repository.Update(ctx, app)
}

func (s *Service) GetWebAppForHost(ctx context.Context, host string) (*WebApp, error) {
	return s.repository.FindOne(ctx, bson.M{"hosts": host})
}

func (s *Service) FindById(ctx context.Context, id primitive.ObjectID) (*WebApp, error) {
	return s.repository.FindById(ctx, id)
}

func (s *Service) FindBySSLId(id primitive.ObjectID, ctx context.Context) ([]string, error) {
	filter := bson.M{"SSLId": id}
	webapps, err := s.repository.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(webapps) == 0 {
		return []string{}, nil
	}
	result := make([]string, len(webapps))
	for i, w := range webapps {
		result[i] = w.Name
	}
	return result, nil
}

func (s *Service) FindByPolicyId(id primitive.ObjectID, ctx context.Context) ([]string, error) {
	filter := bson.M{"policyId": id}
	webapps, err := s.repository.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(webapps) == 0 {
		return []string{}, nil
	}
	result := make([]string, len(webapps))
	for i, w := range webapps {
		result[i] = w.Name
	}
	return result, nil
}

func (s *Service) FindByPolicyIDs(ids []primitive.ObjectID, ctx context.Context) (map[primitive.ObjectID][]string, error) {
	filter := bson.M{
		"policyId": bson.M{"$in": ids},
	}
	webapps, err := s.repository.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make(map[primitive.ObjectID][]string)
	for _, w := range webapps {
		result[w.PolicyId] = append(result[w.PolicyId], w.Name)
	}
	return result, nil
}
