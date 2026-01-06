package initialization

import (
	"context"
	"fmt"
	policy "reverseproxy/internal/domain/policy"
	rule "reverseproxy/internal/domain/rule"
)

func getDefaultPolicy(rs *rule.Service) (*policy.Policy, error) {
	ctx := context.Background()
	rules, err := rs.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all rules %w", err)
	}
	// добавляем к правилам дефолтные actions
	// тут мы не указываем у PolicyRuleRef Actions [] = не переопределяем стандартный набор действий при сработке
	var policyRuleRef []policy.PolicyRuleRef
	for _, rule := range rules {
		policyRuleRef = append(policyRuleRef, policy.PolicyRuleRef{RuleID: rule.ID})
	}
	wl := "95.67.162.0/24" // для тестов
	defaultPolicy := policy.Policy{WL: []string{wl}, Rules: policyRuleRef, Name: DEFAULT_POLICY_NAME}
	return &defaultPolicy, nil
}
