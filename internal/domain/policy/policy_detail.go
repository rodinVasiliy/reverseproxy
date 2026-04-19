package policy

import (
	"reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/rule"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Detail struct {
	ID   primitive.ObjectID `json:"id"`
	Name string             `json:"name"`

	WL []string `json:"wl"`

	Rules []RuleRefView `json:"rules"`
}

type RuleRefView struct {
	ID      primitive.ObjectID `json:"id"`
	Name    string             `json:"name"`
	Enabled bool               `json:"enabled"`
	Actions []ActionView       `json:"actions"`
}

type ActionView struct {
	ID   primitive.ObjectID `json:"id"`
	Name string             `json:"name"`
}

func BuildResponse(policy Policy, rulesMap map[primitive.ObjectID]rule.Rule, actionsMap map[primitive.ObjectID]action.ActionDoc) Detail {
	ruleRefViews := make([]RuleRefView, len(policy.Rules))

	// TODO при возможности оптимизировать
	for i, ruleRef := range policy.Rules {
		r := rulesMap[ruleRef.RuleID]

		var actionIds []primitive.ObjectID
		if len(ruleRef.Actions) > 0 {
			actionIds = ruleRef.Actions
		} else {
			actionIds = r.Actions
		}

		var actionViews []ActionView
		for _, actionId := range actionIds {
			actionViews = append(actionViews, ActionView{
				ID:   actionId,
				Name: actionsMap[actionId].Name,
			})
		}

		ruleRefViews[i] = RuleRefView{
			ID:      ruleRef.RuleID,
			Name:    r.Name,
			Enabled: r.Enabled,
			Actions: actionViews,
		}

	}
	response := Detail{
		ID:   policy.ID,
		Name: policy.Name,

		WL: policy.WL,

		Rules: ruleRefViews,
	}
	return response
}
