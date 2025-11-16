package service

import (
	config "reverseproxy/config/mongo_config"
	policy "reverseproxy/model/policy"
)

type PolicyService struct {
	deps        *config.MongoDeps
	ruleService *RuleService
}

func NewPolicyService(deps *config.MongoDeps, ruleService *RuleService) *PolicyService {
	return &PolicyService{deps: deps, ruleService: ruleService}
}

func (ps *PolicyService) RuleService() *RuleService {
	return ps.ruleService
}

func (ps *PolicyService) LoadtPolicyToDB(policy *policy.Policy) {
	mongoConfig := ps.deps.Config
	client := ps.deps.Client

	db := mongoConfig.Database
	collection := client.Database(db).Collection(POLICY_COLLECTION)
	ctx, cancel := ps.deps.Ctx()
	defer cancel()

	collection.InsertOne(ctx, *policy)
}
