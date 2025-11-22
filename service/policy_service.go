package service

import (
	config "reverseproxy/config/mongo_config"
	policy "reverseproxy/model/policy"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PolicyService struct {
	deps *config.MongoDeps
}

func NewPolicyService(deps *config.MongoDeps, ruleService *RuleService) *PolicyService {
	return &PolicyService{deps: deps}
}

func (ps *PolicyService) Add(policy *policy.Policy) (interface{}, error) {
	return add(ps.deps, ps.deps.Config.Database, POLICY_COLLECTION, policy)
}

func (ps *PolicyService) FindById(id primitive.ObjectID) (*policy.Policy, error) {
	return findById[policy.Policy](ps.deps, ps.deps.Config.Database, POLICY_COLLECTION, id)
}
