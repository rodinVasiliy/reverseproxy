package service

import (
	"fmt"
	config "reverseproxy/config/mongo_config"
	webapp "reverseproxy/model/web_app"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebAppService struct {
	deps       *config.MongoDeps
	sslService *SSLConfigurationService
}

func NewWebAppService(deps *config.MongoDeps, sslService *SSLConfigurationService) *WebAppService {
	return &WebAppService{deps: deps, sslService: sslService}
}

func (wA *WebAppService) FindAllWebApps() (*[]webapp.WebApp, error) {
	return findAll[webapp.WebApp](wA.deps, wA.deps.Config.Database, WEBAPP_COLLECTION)
}

func (wA *WebAppService) Add(webApp *webapp.WebApp) (primitive.ObjectID, error) {
	ssl, err := wA.sslService.FindById(webApp.SSLId)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to find ssl for web app %w", err)
	}
	id, err := add(wA.deps, wA.deps.Config.Database, WEBAPP_COLLECTION, webApp)
	if err != nil {
		return primitive.NewObjectID(), err
	}
	webApp.ID = id
	nginxConfig := GenerateNginxConfig(*webApp, ssl.CertPath, ssl.KeyPath)
	createNginxFiles(*webApp, nginxConfig)
	return id, nil
}

func (wA *WebAppService) Delete(app *webapp.WebApp) error {
	deleteNginxFiles(*app)
	return delete[webapp.WebApp](wA.deps, wA.deps.Config.Database, WEBAPP_COLLECTION, app.ID)
}

func (wA *WebAppService) Edit(app *webapp.WebApp) error {
	ssl, err := wA.sslService.FindById(app.SSLId)
	if err != nil {
		return fmt.Errorf("failed to find ssl for web app %w", err)
	}
	nginxConfig := GenerateNginxConfig(*app, ssl.CertPath, ssl.KeyPath)
	editNginxFiles(*app, nginxConfig)
	return edit(wA.deps, wA.deps.Config.Database, WEBAPP_COLLECTION, app)
}

func (wA *WebAppService) GetWebAppForHost(host string) (*webapp.WebApp, error) {
	mongoConfig := wA.deps.Config
	client := wA.deps.Client
	ctx, cancel := wA.deps.Ctx()
	defer cancel()

	db := mongoConfig.Database
	webAppCollection := client.Database(db).Collection(WEBAPP_COLLECTION)

	filter := bson.M{"hosts": host}
	var webApp webapp.WebApp
	err := webAppCollection.FindOne(ctx, filter).Decode(&webApp)
	if err != nil {
		return nil, fmt.Errorf("failed to find webapp for host %s in db %s", host, err)
	}

	return &webApp, nil
}
