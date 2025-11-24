package service

import (
	"fmt"
	config "reverseproxy/config/mongo_config"
	webapp "reverseproxy/model/web_app"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebAppService struct {
	deps *config.MongoDeps
}

func NewWebAppService(deps *config.MongoDeps) *WebAppService {
	return &WebAppService{deps: deps}
}

func (wA *WebAppService) FindAllWebApps() (*[]webapp.WebApp, error) {
	return findAll[webapp.WebApp](wA.deps, wA.deps.Config.Database, WEBAPP_COLLECTION)
}

func (wA *WebAppService) Add(webApp *webapp.WebApp) (primitive.ObjectID, error) {
	return add(wA.deps, wA.deps.Config.Database, WEBAPP_COLLECTION, webApp)
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
