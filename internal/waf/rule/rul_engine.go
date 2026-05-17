package rule

import (
	"reverseproxy/internal/domain/rule"
	"reverseproxy/internal/waf/action"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RuleEngine struct {
	mu    sync.RWMutex
	rules map[primitive.ObjectID]*CompiledRule

	actionEngine *action.ActionEngine
}

func (re *RuleEngine) SetActionEngine(ae *action.ActionEngine) {
	re.actionEngine = ae
}

func (re *RuleEngine) Load(rules []rule.Rule) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	newRules := make(map[primitive.ObjectID]*CompiledRule)
	for _, r := range rules {
		cr, err := CompileRule(r, re.actionEngine)
		if err != nil {
			return err
		}
		newRules[cr.ID] = cr
	}

	re.rules = newRules
	return nil
}

func (re *RuleEngine) Get(id primitive.ObjectID) (*CompiledRule, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	r, ok := re.rules[id]
	return r, ok
}

// TODO добавить create + delete

func (re *RuleEngine) Update(rule rule.Rule) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	cr, err := CompileRule(rule, re.actionEngine)
	if err != nil {
		return err
	}
	re.rules[cr.ID] = cr
	return nil
}
