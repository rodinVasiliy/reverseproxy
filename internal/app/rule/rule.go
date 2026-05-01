package rule

import (
	"context"
	"reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/domain/rule"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppRuleService struct {
	ruleService   *rule.Service
	actionService *action.Service
	policyService *policy.Service
}

func NewAppRuleService(rs *rule.Service, as *action.Service, ps *policy.Service) *AppRuleService {
	return &AppRuleService{
		ruleService:   rs,
		actionService: as,
		policyService: ps,
	}
}

func (a *AppRuleService) RuleDetailById(ctx context.Context, id primitive.ObjectID) (*rule.RuleDetail, error) {
	r, err := a.ruleService.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	rd := &rule.RuleDetail{
		ID:                 r.ID,
		Name:               r.Name,
		Enabled:            r.Enabled,
		Expr:               &r.Expr,
		Actions:            []rule.ActionParam{},
		PolicyActionParams: []rule.PolicyActionParam{},
	}

	actions, err := a.actionService.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	actionsMap := sliceToMap(actions, func(a action.ActionDoc) primitive.ObjectID {
		return a.ID
	})

	for _, a := range r.Actions {
		rd.Actions = append(rd.Actions, rule.ActionParam{
			ID:   a,
			Name: actionsMap[a].Name,
		})
	}

	overrides, err := a.policyService.FindOverrideForRule(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	if len(overrides) != 0 {
		for _, p := range overrides {
			pap := rule.PolicyActionParam{
				ID:      p.ID,
				Name:    p.Name,
				Actions: make([]rule.ActionParam, 0, len(p.Rules[0].Actions)),
			}

			for _, act := range p.Rules[0].Actions {
				pap.Actions = append(pap.Actions, rule.ActionParam{
					ID:   act,
					Name: actionsMap[act].Name,
				})
			}
			rd.PolicyActionParams = append(rd.PolicyActionParams, pap)
		}
	}

	return rd, nil
}

func sliceToMap[T any, K comparable](items []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(items))
	for _, item := range items {
		result[keyFn(item)] = item
	}
	return result
}
