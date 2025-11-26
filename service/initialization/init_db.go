package initialization

import (
	"fmt"
	service "reverseproxy/service"
)

func InItDB(ps *service.PolicyService, as *service.ActionService, rs *service.RuleService) error {
	actions := GetDefaultActions()
	err := LoadActionsToDB(as, actions)
	if err != nil {
		return fmt.Errorf("failed to load default actions to DB %w", err)
	}
	rules, err := GetDefaultRules(as)
	if err != nil {
		return fmt.Errorf("failed to get default rules %w", err)
	}
	for _, rule := range *rules {
		_, err := rs.Add(&rule)
		if err != nil {
			return fmt.Errorf("failed to add rule %s to db %w", rule.Name, err)
		}
	}

	defaultPolicy, err := GetDefaultPolicy(ps, rs)
	if err != nil {
		return fmt.Errorf("failed to get default policy %w", err)
	}
	_, err = ps.Add(defaultPolicy)
	if err != nil {
		return fmt.Errorf("failed to add policy to db %w", err)
	}
	fmt.Println("in it db successfull")
	return nil
}
