package service

import (
	"fmt"
	config "reverseproxy/config/mongo_config"
	"reverseproxy/model/webApp"

	"go.mongodb.org/mongo-driver/bson"
)

type WebAppService struct {
	deps *config.MongoDeps
}

func NewWebAppService(deps *config.MongoDeps) *WebAppService {
	return &WebAppService{deps: deps}
}

func (wA *WebAppService) FindAllWebApps() (*[]webApp.WebApp, error) {
	return findAll[webApp.WebApp](wA.deps, wA.deps.Config.Database, WEBAPP_COLLECTION)
}

func (wA *WebAppService) GetWebAppForHost(host string) (*webApp.WebApp, error) {
	mongoConfig := wA.deps.Config
	client := wA.deps.Client
	ctx, cancel := wA.deps.Ctx()
	defer cancel()

	db := mongoConfig.Database
	webAppCollection := client.Database(db).Collection(WEBAPP_COLLECTION)

	filter := bson.M{"hosts": host}
	var webApp webApp.WebApp
	err := webAppCollection.FindOne(ctx, filter).Decode(&webApp)
	if err != nil {
		return nil, fmt.Errorf("failed to find webapp for host %s in db %s", host, err)
	}

	return &webApp, nil
}
