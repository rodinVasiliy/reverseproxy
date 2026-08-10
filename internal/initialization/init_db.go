package initialization

import (
	"context"
	"fmt"
	"reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/domain/rule"
)

func InItDB(actionService *action.Service, ruleService *rule.Service, policyService *policy.Service) error {
	actions := getDefaultActions()
	err := loadActionsToDB(actionService, actions)
	if err != nil {
		return fmt.Errorf("failed to load default actions to DB: %w", err)
	}
	rules, err := getDefaultRules(actionService)
	if err != nil {
		return fmt.Errorf("failed to get default rules: %w", err)
	}
	err = loadRulesToDB(ruleService, rules)
	if err != nil {
		return fmt.Errorf("failed to load rules to db: %w", err)
	}

	defaultPolicy, err := getDefaultPolicy()
	if err != nil {
		return fmt.Errorf("failed to get default policy: %w", err)
	}
	_, err = policyService.Insert(context.Background(), *defaultPolicy)
	if err != nil {
		return fmt.Errorf("failed to insert policy to db: %w", err)
	}
	fmt.Println("in it db successfull")
	return nil
}
