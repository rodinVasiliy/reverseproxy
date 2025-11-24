package service

import (
	config "reverseproxy/config/mongo_config"
	action "reverseproxy/model/action"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (as *ActionService) Add(actionDoc *action.ActionDoc) (primitive.ObjectID, error) {
	return add(as.deps, as.deps.Config.Database, ACTIONS_COLLECTION, actionDoc)
}

func (as *ActionService) FindById(id primitive.ObjectID) (*action.Action, error) {
	return findById[action.Action](as.deps, as.deps.Config.Database, ACTIONS_COLLECTION, id)
}
