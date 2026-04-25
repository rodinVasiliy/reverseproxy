package webapp

import (
	"context"
	"fmt"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/domain/ssl"
	"reverseproxy/internal/domain/webapp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppWebappService struct {
	webappService *webapp.Service
	policyService *policy.Service
	sslService    *ssl.Service
}

func NewService(ws *webapp.Service, ps *policy.Service, ss *ssl.Service) *AppWebappService {
	return &AppWebappService{
		webappService: ws,
		policyService: ps,
		sslService:    ss,
	}
}

func (ws *AppWebappService) List(ctx context.Context) ([]webapp.Response, error) {
	webapps, err := ws.webappService.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]webapp.Response, 0, len(webapps))
	for _, w := range webapps {
		p, err := ws.policyService.FindById(ctx, w.PolicyId)
		if err != nil {
			return nil, fmt.Errorf("failed to find policy for webapp %v: %s", w.Name, err)
		}
		sslConfiguration, err := ws.sslService.FindByID(ctx, w.SSLId)
		if err != nil {
			return nil, fmt.Errorf("failed to find ssl config for webapp %v: %s", w.Name, err)
		}
		responses = append(responses, webapp.Response{
			ID:         w.ID.Hex(),
			Name:       w.Name,
			PolicyId:   p.ID.Hex(),
			PolicyName: p.Name,
			SSLId:      sslConfiguration.ID.Hex(),
			SSLName:    sslConfiguration.Name,
			Upstream:   w.Upstream,
			Port:       w.Port,
			Hosts:      w.Hosts,
		})
	}
	return responses, nil
}

func (ws *AppWebappService) CreateNginxFiles(app webapp.WebApp, ctx context.Context) error {
	sslConfiguration, err := ws.sslService.FindByID(ctx, app.SSLId)
	if err != nil {
		return err
	}
	config := generateNginxConfig(app, sslConfiguration.CertFileName, sslConfiguration.KeyFileName)
	createNginxFiles(app, config)
	return nil
}

func (ws *AppWebappService) UpdateNginxFiles(app webapp.WebApp, ctx context.Context) error {
	sslConfiguration, err := ws.sslService.FindByID(ctx, app.SSLId)
	if err != nil {
		return err
	}
	nginxConfig := generateNginxConfig(app, sslConfiguration.CertFileName, sslConfiguration.KeyFileName)
	editNginxFiles(app, nginxConfig)
	return nil
}

func (ws *AppWebappService) RemoveNginxFiles(id primitive.ObjectID, ctx context.Context) error {
	app, err := ws.webappService.FindById(ctx, id)
	if err != nil {
		return err
	}
	deleteNginxFiles(*app)
	return nil
}
