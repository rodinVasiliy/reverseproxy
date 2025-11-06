package service

import (
	config "reverseproxy/config/mongo_config"
	"reverseproxy/model/webapp"
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
