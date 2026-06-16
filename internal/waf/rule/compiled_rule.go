package rule

import (
	"fmt"
	"reverseproxy/internal/domain/rule"
	"reverseproxy/internal/waf/action"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CompiledRule struct {
	ID       primitive.ObjectID
	Name     string
	Enabled  bool
	Policies []primitive.ObjectID // Список политик, для которых правило включено
	Expr     rule.Expr            // уже собранный
	Actions  []action.Action      // уже резолвленные
}

func CompileRule(r rule.Rule, ae *action.ActionEngine) (*CompiledRule, error) {
	expr, err := rule.BuildExpr(r.Expr)
	if err != nil {
		return nil, err
	}

	actions := make([]action.Action, 0, len(r.Actions))
	for _, id := range r.Actions {
		act, ok := ae.Get(id)
		if !ok {
			return nil, fmt.Errorf("action not found: %s", id.Hex())
		}
		actions = append(actions, act)
	}

	return &CompiledRule{
		ID:       r.ID,
		Name:     r.Name,
		Enabled:  r.Enabled,
		Expr:     expr,
		Actions:  actions,
		Policies: r.Policies,
	}, nil
}
