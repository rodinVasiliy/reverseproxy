package webapp

import (
	"context"
	"fmt"
	"log"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/domain/ssl"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	repository    *repository.MongoRepository[WebApp]
	sslService    *ssl.Service
	policyService *policy.Service
}

func NewService(repo *repository.MongoRepository[WebApp], sslService *ssl.Service, policyService *policy.Service) *Service {
	return &Service{repository: repo, sslService: sslService, policyService: policyService}
}

func (s *Service) FindAll(ctx context.Context) ([]WebApp, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	webapps, err := s.repository.FindAll(ctx)
	if err != nil {
		log.Printf("failed to find all webapps: %v", err)
		return nil, err
	}

	responses := make([]Response, 0, len(webapps))
	for _, w := range webapps {
		policy, err := s.policyService.FindById(ctx, w.PolicyId)
		if err != nil {
			return nil, fmt.Errorf("failed to find policy for webapp %v: %s", w.Name, err)
		}
		ssl, err := s.sslService.FindByID(ctx, w.SSLId)
		if err != nil {
			return nil, fmt.Errorf("failed to find ssl config for webapp %v: %s", w.Name, err)
		}
		responses = append(responses, Response{
			ID:         w.ID.Hex(),
			Name:       w.Name,
			PolicyId:   policy.ID.Hex(),
			PolicyName: policy.Name,
			SSLId:      ssl.ID.Hex(),
			SSLName:    ssl.Name,
			Upstream:   w.Upstream,
			Port:       w.Port,
			Hosts:      w.Hosts,
		})
	}
	return responses, nil
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

func (s *Service) WatchChanges() {
	ctx := context.Background()
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	collection := s.repository.Collection()

	stream, err := collection.Watch(ctx, mongo.Pipeline{}, opts)
	if err != nil {
		log.Println(err)
	}

	defer stream.Close(ctx)

	for stream.Next(ctx) {
		var event struct {
			OperationType string `bson:"operationType"`
			FullDocument  WebApp `bson:"fullDocument"`
			DocumentKey   struct {
				ID primitive.ObjectID `bson:"_id"`
			} `bson:"documentKey"`
		}

		if err := stream.Decode(&event); err != nil {
			log.Println(err)
			continue
		}

		switch event.OperationType {

		case "insert":
			s.create(event.FullDocument, ctx)

		case "update":
			s.update(event.FullDocument, ctx)

		case "delete":
			s.remove(event.DocumentKey.ID, ctx)
		}

	}
	if err := stream.Err(); err != nil {
		log.Println("watch stream error:", err)
	}
}

func (s *Service) create(app WebApp, ctx context.Context) {
	ssl, err := s.sslService.FindByID(ctx, app.SSLId)
	if err != nil {
		log.Println(err)
		return
	}
	config := generateNginxConfig(app, ssl.CertFileName, ssl.KeyFileName)
	createNginxFiles(app, config)
}

func (s *Service) remove(id primitive.ObjectID, ctx context.Context) {
	app, err := s.FindById(ctx, id)
	if err != nil {
		log.Println(err)
		return
	}
	deleteNginxFiles(*app)
}

func (s *Service) update(app WebApp, ctx context.Context) {
	ssl, err := s.sslService.FindByID(ctx, app.SSLId)
	if err != nil {
		log.Println(err)
		return
	}
	nginxConfig := generateNginxConfig(app, ssl.CertFileName, ssl.KeyFileName)
	editNginxFiles(app, nginxConfig)
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
	filter := bson.M{"PolicyId": id}
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
