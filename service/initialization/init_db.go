package initialization

import (
	"fmt"
	service "reverseproxy/service"
)

func InItDB(ps *service.PolicyService, as *service.ActionService, rs *service.RuleService) {
	actions := GetDefaultActions()
	err := LoadActionsToDB(as, actions)
	if err != nil {
		fmt.Printf("failed to load default actions to DB %s", err)
	}
	rules, err := GetDefaultRules(as)
	if err != nil {
		fmt.Printf("failed to get default rules %s", err)
	}
	for _, rule := range *rules {
		_, err := rs.Add(&rule)
		if err != nil {
			fmt.Printf("failed to add rule %s to db %s", rule.Name, err)
		}
	}

	defaultPolicy, err := GetDefaultPolicy(ps, rs)
	if err != nil {
		fmt.Printf("failed to get default policy %s", err)
	}
	_, err = ps.Add(defaultPolicy)
	if err != nil {
		fmt.Printf("failed to add policy to db %s", err)
	}
	fmt.Printf("in it db successfull")
}
