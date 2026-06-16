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

	// Находим для правила Actions, добавляем в RuleDetail имена Actions
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

	// Находим для правила политики, в которых оно используется и добавляем их ID в RuleDetail
	policies, err := a.policyService.FindByRuleID(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	rd.Policies = make([]primitive.ObjectID, 0, len(policies))
	for _, policy := range policies {
		rd.Policies = append(rd.Policies, policy.ID)
	}

	// Находим политики, в которых действия для этого правила переопределены и добавляем их в RuleDetail
	overrides, err := a.policyService.FindOverrideForRule(ctx, r.ID)
	if err != nil {
		return nil, err
	}
	if len(overrides) != 0 {
		for _, p := range overrides {
			pap := rule.PolicyActionParam{
				ID:      p.ID,
				Name:    p.Name,
				Actions: make([]rule.ActionParam, 0, len(p.RuleOverrides[0].Actions)),
			}

			for _, act := range p.RuleOverrides[0].Actions {
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

func (a *AppRuleService) SyncOverrideToPolicy(ctx context.Context, policyId primitive.ObjectID, ruleId primitive.ObjectID,
	actionIDs []primitive.ObjectID) error {
	// Работает следующим образом: в политике ищет, есть ли для правила с ruleId переопределение, если есть - меняет в нем список действий на actionIDs
	// Если нет - в случае, если actionIDs не пустой - создает для политики переопределение, добавляя туда actionIDs
	p, err := a.policyService.FindById(ctx, policyId)
	if err != nil {
		return err
	}

	found := false // Есть ли уже для правила с ruleId какой-либо Override
	for i, ruleRef := range p.RuleOverrides {
		if ruleRef.RuleID == ruleId {
			p.RuleOverrides[i].Actions = actionIDs
			found = true
		}
	}
	if !found { // Если не было Override до этого
		if len(actionIDs) != 0 {
			p.RuleOverrides = append(p.RuleOverrides, policy.RuleRef{
				RuleID:  ruleId,
				Actions: actionIDs,
			})
		}
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

	// String -> primitive.ObjectID
	r.Actions = make([]primitive.ObjectID, 0, len(dto.Actions))
	for _, action := range dto.Actions {
		id, err := primitive.ObjectIDFromHex(action)
		if err != nil {
			return err
		}
		r.Actions = append(r.Actions, id)
	}

	// Собираем политики для которых необходима рекомпиляция
	policyToReCompile := make([]primitive.ObjectID, 0, len(dto.PolicyOverrides))

	for i, policyId := range dto.PolicyOverrides {
		pId, err := primitive.ObjectIDFromHex(policyId.ID)
		policyToReCompile = append(policyToReCompile, pId)
		if err != nil {
			return err
		}

		actionIDs := make([]primitive.ObjectID, 0, len(dto.PolicyOverrides[i].Actions))
		for _, action := range dto.PolicyOverrides[i].Actions {
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

func (a *AppRuleService) CreateRule(ctx context.Context, dto *ruleDto.RuleDto) error {
	var r rule.Rule
	r.Name = dto.Name
	r.Enabled = dto.Enabled
	r.Actions = make([]primitive.ObjectID, 0, len(dto.Actions))
	for _, actionId := range dto.Actions {
		id, err := primitive.ObjectIDFromHex(actionId)
		if err != nil {
			return err
		}
		r.Actions = append(r.Actions, id)
	}

	// Определяем список политик, для которых будет повторная компиляция
	policyToReCompileMap := make(map[string][]string)
	for _, policyId := range dto.Policies {
		policyToReCompileMap[policyId] = []string{} // Назначаем пустой срез, т.к. для этой политики пока нет Override(но он может быть далее)
	}

	for _, override := range dto.PolicyOverrides {
		policyToReCompileMap[override.ID] = override.Actions
	}

	for key, value := range policyToReCompileMap {
		policyId, err := primitive.ObjectIDFromHex(key)
		if err != nil {
			return err
		}

		actionIds := make([]primitive.ObjectID, 0, len(value))
		for _, actionId := range value {
			id, err := primitive.ObjectIDFromHex(actionId)
			if err != nil {
				return err
			}
			actionIds = append(actionIds, id)
		}

		err = a.SyncOverrideToPolicy(ctx, policyId, r.ID, actionIds)
		if err != nil {
			return err
		}
	}

	_, err := a.ruleService.Insert(ctx, r)
	if err != nil {
		return err
	}

	// TODO переделать метод SyncOverrideToPolicy

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
