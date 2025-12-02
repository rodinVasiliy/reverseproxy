package initialization

import (
	"fmt"
	"reverseproxy/model/policy"
	service "reverseproxy/service"
)

func GetDefaultPolicy(ps *service.PolicyService, rs *service.RuleService) (*policy.Policy, error) {
	rules, err := rs.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to find all rules %w", err)
	}
	// добавляем к правилам дефолтные actions
	var policyRuleRef []policy.PolicyRuleRef
	for _, rule := range *rules {
		policyRuleRef = append(policyRuleRef, policy.PolicyRuleRef{RuleID: rule.ID})
	}
	wl := "95.67.162.0/24"
	defaultPolicy := policy.Policy{WL: []string{wl}, Rules: policyRuleRef, Name: service.DEFAULT_POLICY_NAME}
	return &defaultPolicy, nil
}
