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
}

func (pe *PolicyEngine) Load(policies []policy.Policy, ruleEngine *rule.RuleEngine,
	actionEngine *action.ActionEngine) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	newPolicies := make(map[primitive.ObjectID]*CompiledPolicy)

	for _, p := range policies {
		cp, err := CompilePolicy(p, ruleEngine, actionEngine)
		if err != nil {
			return err
		}
		newPolicies[p.ID] = cp
	}

	pe.policies = newPolicies
	return nil
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
