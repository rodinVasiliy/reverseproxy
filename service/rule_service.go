package service

import (
	config "reverseproxy/config/mongo_config"
	rule "reverseproxy/model/rule"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RuleService struct {
	deps *config.MongoDeps
	as   *ActionService
}

func NewRuleService(deps *config.MongoDeps, as *ActionService) *RuleService {
	return &RuleService{deps: deps, as: as}
}

func (rs *RuleService) FindAll() (*[]rule.Rule, error) {
	return findAll[rule.Rule](rs.deps, rs.deps.Config.Database, RULE_COLLECTION)
}

func (rs *RuleService) Add(rule *rule.Rule) (primitive.ObjectID, error) {
	return add(rs.deps, rs.deps.Config.Database, RULE_COLLECTION, rule)
}

func (rs *RuleService) FindById(id primitive.ObjectID) (*rule.Rule, error) {
	return findById[rule.Rule](rs.deps, rs.deps.Config.Database, RULE_COLLECTION, id)
}
