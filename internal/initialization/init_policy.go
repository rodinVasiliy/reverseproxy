package initialization

import (
	"context"
	"fmt"
	policy "reverseproxy/internal/domain/policy"
	rule "reverseproxy/internal/domain/rule"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getDefaultPolicy(rs *rule.Service) (*policy.Policy, error) {
	ctx := context.Background()
	rules, err := rs.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all rules %w", err)
	}

	// Добавляем все правила в политику
	ruleIds := make([]primitive.ObjectID, 0, len(rules))
	for _, rule := range rules {
		ruleIds = append(ruleIds, rule.ID)
	}

	// TODO - а можно ли тогда эту секцию убрать?
	// тут мы не указываем у PolicyRuleRef Actions [] = не переопределяем стандартный набор действий при сработке
	var policyRuleRef []policy.RuleRef
	for _, rule := range rules {
		policyRuleRef = append(policyRuleRef, policy.RuleRef{RuleID: rule.ID})
	}
	wl := "95.67.162.0/24" // для тестов
	defaultPolicy := policy.Policy{WL: []string{wl}, RuleOverrides: policyRuleRef, Name: DEFAULT_POLICY_NAME}
	return &defaultPolicy, nil
}
