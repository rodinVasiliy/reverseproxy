package compiler

import (
	"fmt"
	"reverseproxy/internal/waf/action"
	"reverseproxy/internal/waf/policy"
	"reverseproxy/internal/waf/rule"

	actionDoc "reverseproxy/internal/domain/action"

	ruleDoc "reverseproxy/internal/domain/rule"

	policyDoc "reverseproxy/internal/domain/policy"
)

type Compiler struct {
	ActionCompiler *action.ActionEngine
	PolicyCompiler *policy.PolicyEngine
	RuleCompiler   *rule.RuleEngine
}

func Compile(actionDocs []actionDoc.ActionDoc, actionRegistry *action.Registry, rules []ruleDoc.Rule, policies []policyDoc.Policy) (*Compiler, error) {
	// Компилируем actions
	actionEngine := &action.ActionEngine{}
	err := actionEngine.Load(actionDocs, actionRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to compile actions: %w", err)
	}

	// Компилируем правила
	ruleEngine := &rule.RuleEngine{}
	ruleEngine.SetActionEngine(actionEngine)
	err = ruleEngine.Load(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to compile rules: %w", err)
	}

	// Компилируем политики
	policyEngine := &policy.PolicyEngine{}
	policyEngine.SetActionEngine(actionEngine)
	policyEngine.SetRuleEngine(ruleEngine)
	err = policyEngine.Load(policies)
	if err != nil {
		return nil, fmt.Errorf("failed to compile policies: %w", err)
	}

	return &Compiler{
		ActionCompiler: actionEngine,
		PolicyCompiler: policyEngine,
		RuleCompiler:   ruleEngine,
	}, nil
}
