package policy

import (
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/waf/action"
	"reverseproxy/internal/waf/rule"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PolicyEngine struct {
	mu       sync.RWMutex
	policies map[primitive.ObjectID]*CompiledPolicy

	ruleEngine   *rule.RuleEngine
	actionEngine *action.ActionEngine
}

func (pe *PolicyEngine) SetRuleEngine(re *rule.RuleEngine) {
	pe.ruleEngine = re
}

func (pe *PolicyEngine) SetActionEngine(ae *action.ActionEngine) {
	pe.actionEngine = ae
}

func (pe *PolicyEngine) Load(policies []policy.Policy) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	newPolicies := make(map[primitive.ObjectID]*CompiledPolicy)

	for _, p := range policies {
		cp, err := CompilePolicy(p, pe.ruleEngine, pe.actionEngine)
		if err != nil {
			return err
		}
		newPolicies[p.ID] = cp
	}

	pe.policies = newPolicies
	return nil
}

func (pe *PolicyEngine) Delete(p policy.Policy) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	delete(pe.policies, p.ID)
}

func (pe *PolicyEngine) Get(id primitive.ObjectID) (*CompiledPolicy, bool) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	cp, ok := pe.policies[id]
	if !ok {
		return nil, false
	}
	return cp, true
}

func (pe *PolicyEngine) Update(p policy.Policy) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	cp, err := CompilePolicy(p, pe.ruleEngine, pe.actionEngine)
	if err != nil {
		return err
	}
	pe.policies[p.ID] = cp
	return nil
}
