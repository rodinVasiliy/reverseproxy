package service

import (
	config "reverseproxy/config/mongo_config"
	action "reverseproxy/model/action"
)

type ActionService struct {
	deps *config.MongoDeps
}

func NewActionService(deps *config.MongoDeps) *ActionService {
	return &ActionService{deps: deps}
}

func (as *ActionService) FindAllActions() (*[]action.ActionDoc, error) {
	return findAll[action.ActionDoc](as.deps, as.deps.Config.Database, ACTIONS_COLLECTION)
}

func (as *ActionService) Add(actionDoc *action.ActionDoc) (interface{}, error) {
	return add(as.deps, as.deps.Config.Database, ACTIONS_COLLECTION, actionDoc)
}
