package rule

import (
	"context"
	"fmt"
	"log"
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
		ID:        r.ID,
		Name:      r.Name,
		Enabled:   r.Enabled,
		Expr:      &r.Expr,
		Policies:  []rule.PolicyParam{},
		Actions:   []rule.ActionParam{},
		Overrides: []rule.OverrideParam{},
	}

	// Находим для правила Actions, добавляем в RuleDetail имена Actions
	actions, err := a.actionService.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Преобразовываем в map для удобного поиска
	actionsMap := sliceToMap(actions, func(a action.ActionDoc) primitive.ObjectID {
		return a.ID
	})

	// Заносим в правило ID действия + Имя действия
	for _, a := range r.Actions {
		rd.Actions = append(rd.Actions, rule.ActionParam{
			ID:   a,
			Name: actionsMap[a].Name,
		})
	}

	policies, err := a.policyService.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	policiesMap := sliceToMap(policies, func(p policy.Policy) primitive.ObjectID {
		return p.ID
	})

	// Заполняем Политики (id + имя)
	for _, policyId := range r.Policies {
		rd.Policies = append(rd.Policies, rule.PolicyParam{
			ID:   policyId,
			Name: policiesMap[policyId].Name,
		})
	}

	// Находим политики, в которых действия для этого правила переопределены и добавляем их в RuleDetail
	for _, override := range r.Overrides {
		overrideParam := rule.OverrideParam{
			ID:      override.PolicyId,
			Name:    policiesMap[override.PolicyId].Name,
			Actions: []rule.ActionParam{},
		}
		for _, actionId := range override.Actions {
			overrideParam.Actions = append(overrideParam.Actions, rule.ActionParam{
				ID:   actionId,
				Name: actionsMap[actionId].Name,
			})
		}
		rd.Overrides = append(rd.Overrides, overrideParam)
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

func (a *AppRuleService) UpdateRule(ctx context.Context, r *rule.Rule, dto *ruleDto.RuleDto) error {
	r.Name = dto.Name
	r.Enabled = dto.Enabled

	// String -> primitive.ObjectID: Обновляем список действий
	actions, err := objectIDSliceFromHexSlice(dto.Actions)
	if err != nil {
		return err
	}
	r.Actions = actions

	// Обновляем список переопределений
	for _, policyOverride := range dto.PolicyOverrides {
		actionsOverride, err := objectIDSliceFromHexSlice(policyOverride.Actions)
		if err != nil {
			return err
		}
		policyId, err := primitive.ObjectIDFromHex(policyOverride.ID)
		if err != nil {
			return err
		}

		r.Overrides = append(r.Overrides, rule.Override{
			PolicyId: policyId,
			Actions:  actionsOverride,
		})
	}

	// Собираем политики для которых необходима рекомпиляция
	policyToReCompile, err := getPolicyToReCompile(r, dto)
	if err != nil {
		return fmt.Errorf("failed to find policy to recompile: %w", err)
	}
	log.Println("policyToReCompile filled")

	exp := fillExprDoc(dto.Expression)
	r.Expr = exp
	log.Println("Expr filled")

	err = a.ruleService.Update(ctx, r)
	if err != nil {
		return err
	}
	log.Println("Rule updated in db")

	// Рекомпиляция правила
	err = a.ruleEngine.Update(*r)
	if err != nil {
		return err
	}
	log.Println("Rule compiled")

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
	log.Println("Policies compiled")
	return nil
}

func (a *AppRuleService) CreateRule(ctx context.Context, dto *ruleDto.RuleDto) error {
	newRule := rule.Rule{
		Name:      dto.Name,
		Enabled:   dto.Enabled,
		Actions:   make([]primitive.ObjectID, 0, len(dto.Actions)),
		Policies:  make([]primitive.ObjectID, 0, len(dto.Policies)),
		Overrides: make([]rule.Override, 0, len(dto.PolicyOverrides)),
	}
	// Список политик, которые необходимо скомпилировать
	policyToCompileMap := make(map[primitive.ObjectID]struct{})
	// заполняем Actions для самого правила
	actions, err := objectIDSliceFromHexSlice(dto.Actions)
	if err != nil {
		return err
	}
	newRule.Actions = actions
	// заполняем Политики, в которых используется это правило
	policies, err := objectIDSliceFromHexSlice(dto.Policies)
	if err != nil {
		return err
	}
	newRule.Policies = policies
	// Добавляем id политик в список политик, которым требуется компиляция
	for _, policyId := range policies {
		policyToCompileMap[policyId] = struct{}{}
	}
	// заполняем переопределения для правила
	for _, policyOverride := range dto.PolicyOverrides {
		actionsOverride, err := objectIDSliceFromHexSlice(policyOverride.Actions)
		if err != nil {
			return err
		}
		policyId, err := primitive.ObjectIDFromHex(policyOverride.ID)
		if err != nil {
			return err
		}

		newRule.Overrides = append(newRule.Overrides, rule.Override{
			PolicyId: policyId,
			Actions:  actionsOverride,
		})
		// Добавляем id политики в список политик, которым требуется компиляция
		policyToCompileMap[policyId] = struct{}{}
	}
	// Заполняем поле Expr
	exp := fillExprDoc(dto.Expression)
	newRule.Expr = exp

	// TO DO - а если какой-то шаг зафейлится?
	// Добавляем правило в список скомпилированных
	err = a.ruleEngine.Create(newRule)
	if err != nil {
		return err
	}

	// Компилируем политики
	for policyId, _ := range policyToCompileMap {
		policyToCompile, err := a.policyService.FindById(ctx, policyId)
		if err != nil {
			return err
		}
		a.policyEngine.Update(*policyToCompile)
	}

	_, err = a.ruleService.Insert(ctx, newRule)
	if err != nil {
		return err
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

// Находим политики, которым нужна рекомпиляция
// Рекомпиляция нужна тем политикам, у которых поменялся список правил или у которых поменялся список переопределений для правил
func getPolicyToReCompile(r *rule.Rule, dto *ruleDto.RuleDto) ([]primitive.ObjectID, error) {

	// Ищем разность между политиками которые были ранее r.Policies и теми, которые сейчас - dto.Policies
	policyIdMap := make(map[primitive.ObjectID]struct{})
	for _, policyId := range r.Policies {
		policyIdMap[policyId] = struct{}{}
	}

	for _, policyIdStr := range dto.Policies {
		policyId, err := primitive.ObjectIDFromHex(policyIdStr)
		if err != nil {
			return nil, err
		}
		if _, ok := policyIdMap[policyId]; ok {
			delete(policyIdMap, policyId)
		} else {
			policyIdMap[policyId] = struct{}{}
		}
	}

	// Добавляем разность между переопределениями которые были r.Overrides и которые стали
	overrideIdMap := make(map[primitive.ObjectID]struct{})
	for _, override := range r.Overrides {
		overrideIdMap[override.PolicyId] = struct{}{}
	}

	for _, override := range dto.PolicyOverrides {
		overrideId, err := primitive.ObjectIDFromHex(override.ID)
		if err != nil {
			return nil, err
		}

		// Находим симметрическую разность, если элемент содержится и там и там - не вносим его в результируюущую мапу
		if _, ok := overrideIdMap[overrideId]; !ok {
			policyIdMap[overrideId] = struct{}{}
		}
	}

	result := make([]primitive.ObjectID, 0, len(policyIdMap))
	// Преобразовываем map в slice
	for key, _ := range policyIdMap {
		result = append(result, key)
	}
	return result, nil
}

func objectIDSliceFromHexSlice(ids []string) ([]primitive.ObjectID, error) {
	result := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		result = append(result, objectId)
	}
	return result, nil
}
