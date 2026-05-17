package rule

import (
	"context"
	"reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/policy"
	"reverseproxy/internal/domain/rule"
	ruleDto "reverseproxy/internal/dto/rule"
	policyEngine "reverseproxy/internal/waf/policy"
	engine "reverseproxy/internal/waf/rule"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppRuleService struct {
	ruleService   *rule.Service
	actionService *action.Service
	policyService *policy.Service

	ruleEngine   *engine.RuleEngine
	policyEngine *policyEngine.PolicyEngine
}

func NewAppRuleService(rs *rule.Service, as *action.Service, ps *policy.Service, re *engine.RuleEngine, pe *policyEngine.PolicyEngine) *AppRuleService {
	return &AppRuleService{
		ruleService:   rs,
		actionService: as,
		policyService: ps,
		ruleEngine:    re,
		policyEngine:  pe,
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

func (a *AppRuleService) RuleResponse(ctx context.Context, id primitive.ObjectID) (*ruleDto.RuleDetailResponse, error) {
	ruleDetail, err := a.RuleDetailById(ctx, id)
	if err != nil {
		return nil, err
	}

	actions, err := a.actionService.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	policies, err := a.policyService.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	ruleResponse, err := ruleDto.BuildRuleResponse(ruleDetail, actions, policies)
	if err != nil {
		return nil, err
	}
	return ruleResponse, nil
}

func sliceToMap[T any, K comparable](items []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(items))
	for _, item := range items {
		result[keyFn(item)] = item
	}
	return result
}

func (a *AppRuleService) SyncOverrideToPolicy(ctx context.Context, policyId primitive.ObjectID,
	ruleId primitive.ObjectID, actionIDs []primitive.ObjectID) error {
	p, err := a.policyService.FindById(ctx, policyId)
	if err != nil {
		return err
	}

	isExist := false // Есть ли уже для правила с ruleId какой-либо Override
	for i, ruleRef := range p.Rules {
		if ruleRef.RuleID == ruleId {
			p.Rules[i].Actions = actionIDs
			isExist = true
		}
	}
	if !isExist { // Если не было Override до этого
		p.Rules = append(p.Rules, policy.RuleRef{
			RuleID:  ruleId,
			Actions: actionIDs,
		})
	}

	err = a.policyService.Update(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

func (a *AppRuleService) UpdateRule(ctx context.Context, r *rule.Rule, dto *ruleDto.RuleDto) error {
	r.Name = dto.Name
	r.Enabled = dto.Enabled
	r.Actions = make([]primitive.ObjectID, 0, len(dto.Actions))
	for _, action := range dto.Actions {
		id, err := primitive.ObjectIDFromHex(action)
		if err != nil {
			return err
		}
		r.Actions = append(r.Actions, id)
	}

	policyToReCompile := make([]primitive.ObjectID, 0, len(dto.Overrides))

	for i, policyId := range dto.Overrides {
		pId, err := primitive.ObjectIDFromHex(policyId.ID)
		policyToReCompile = append(policyToReCompile, pId)
		if err != nil {
			return err
		}
		actionIDs := make([]primitive.ObjectID, 0, len(dto.Overrides[i].Actions))
		for _, action := range dto.Overrides[i].Actions {
			id, err := primitive.ObjectIDFromHex(action)
			if err != nil {
				return err
			}
			actionIDs = append(actionIDs, id)
		}
		err = a.SyncOverrideToPolicy(ctx, pId, r.ID, actionIDs)
		if err != nil {
			return err
		}
	}

	exp := fillExprDoc(dto.Expression)
	r.Expr = exp

	err := a.ruleService.Update(ctx, r)
	if err != nil {
		return err
	}

	// Рекомпиляция правила
	err = a.ruleEngine.Update(*r)
	if err != nil {
		return err
	}

	// Рекомпиляция политик
	for _, id := range policyToReCompile {
		p, err := a.policyService.FindById(ctx, id)
		if err != nil {
			return err
		}

		err = a.policyEngine.Update(*p)
		if err != nil {
			return err
		}
	}
	return nil
}

func fillExprDoc(exprDto ruleDto.ExprDto) rule.ExprDoc {
	exp := rule.ExprDoc{
		NodeType: exprDto.NodeType,
		IsNot:    exprDto.IsNot,
		Operator: exprDto.Operator,
		Match:    exprDto.Match,
		Field:    exprDto.Field,
		Raw:      exprDto.Value,
		Children: make([]rule.ExprDoc, 0, len(exprDto.Children)),
	}

	for _, child := range exprDto.Children {
		exp.Children = append(exp.Children, fillExprDoc(child))
	}

	return exp
}
