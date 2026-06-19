package initialization

import (
	"context"
	"fmt"
	action "reverseproxy/internal/domain/action"
	policy "reverseproxy/internal/domain/policy"
	rule "reverseproxy/internal/domain/rule"
)

func InItDB(ps *policy.Service, as *action.Service, rs *rule.Service) error {
	actions := getDefaultActions()
	err := loadActionsToDB(as, actions)
	if err != nil {
		return fmt.Errorf("failed to load default actions to DB: %w", err)
	}
	rules, err := getDefaultRules(as)
	if err != nil {
		return fmt.Errorf("failed to get default rules: %w", err)
	}
	err = loadRulesToDB(rs, rules)
	if err != nil {
		return fmt.Errorf("failed to load rules to db: %w", err)
	}

	defaultPolicy, err := getDefaultPolicy()
	if err != nil {
		return fmt.Errorf("failed to get default policy: %w", err)
	}
	_, err = ps.Insert(context.Background(), *defaultPolicy)
	if err != nil {
		return fmt.Errorf("failed to insert policy to db: %w", err)
	}
	fmt.Println("in it db successfull")
	return nil
}
