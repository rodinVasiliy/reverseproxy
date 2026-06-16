package policy

import (
	"fmt"
	"log"
	"net"
	"reverseproxy/internal/domain/policy"
	bl "reverseproxy/internal/infrastructure/config/bl"
	"reverseproxy/internal/waf/action"
	parsedRequest "reverseproxy/internal/waf/parsed_request"
	"reverseproxy/internal/waf/rule"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CompiledPolicy struct {
	ID    primitive.ObjectID
	Name  string
	WL    []string // TODO поменять на WL []*net.IPNet
	Rules []PolicyRule
}

// TODO переделать, чтобы для Rule, где нет переопределения вызывался стандартный набор функций
func CompilePolicy(p policy.Policy, ruleEngine *rule.RuleEngine,
	actionEngine *action.ActionEngine) (*CompiledPolicy, error) {
	compiledPolicy := &CompiledPolicy{
		ID:   p.ID,
		Name: p.Name,
		WL:   p.WL,
	}

	// Мапа нужна нам, чтобы не проверять одно правило дважды, например, если правило содержится и в списке p.Rules и в списке p.RuleOverrides
	ruleMap := make(map[primitive.ObjectID]PolicyRule)

	// Находим скомпилированные правила для нашей политики + заносим их в мапу
	rules := ruleEngine.GetRulesByPolicyID(p.ID)
	for _, compiledRule := range rules {
		policyRule := PolicyRule{
			Rule:    &compiledRule,
			Actions: compiledRule.Actions,
		}
		ruleMap[compiledRule.ID] = policyRule
	}

	// Проходим все переопределения для правил, если переопределение не пустое - меняем список actions для правила
	for _, ref := range p.RuleOverrides {
		// Если правило не используется(его нет в списке p.Rules) - пропускаем, т.к. его не нужно добавлять в скомпилированную политику
		if _, ok := ruleMap[ref.RuleID]; !ok {
			continue
		}

		baseRule, ok := ruleEngine.Get(ref.RuleID)
		if !ok {
			return nil, fmt.Errorf("rule not found: %s", ref.RuleID.Hex())
		}

		var actions []action.Action

		if len(ref.Actions) == 0 {
			// Если переопределений нет - значит уже нет смысла продолжать, так как всё уже добавлено
			continue
		} else {
			// override
			actions = make([]action.Action, 0, len(ref.Actions))
			for _, id := range ref.Actions {
				act, ok := actionEngine.Get(id)
				if !ok {
					return nil, fmt.Errorf("action not found: %s", id.Hex())
				}
				actions = append(actions, act)
			}
			policyRule := PolicyRule{
				Rule:    baseRule,
				Actions: actions,
			}
			ruleMap[ref.RuleID] = policyRule
		}
	}

	// Добавляем правила в политику, переопределения учитываются
	compiledPolicy.Rules = make([]PolicyRule, 0, len(ruleMap))
	for _, value := range ruleMap {
		compiledPolicy.Rules = append(compiledPolicy.Rules, value)
	}

	return compiledPolicy, nil
}

func checkInList(ip net.IP, list []string) bool {
	for i := range list {
		_, ipNet, err := net.ParseCIDR(list[i])
		if err != nil {
			log.Printf("failed to parse cidr %s %s", list[i], err)
		}
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// Проверяем запрос по всем правилам из политики
func (cp *CompiledPolicy) Evaluate(req *parsedRequest.ParsedRequest, logger *log.Logger, bl *bl.RedisBL) (bool, error) {
	if checkInList(req.IP, cp.WL) {
		return false, nil
	}

	reqMap := req.ToMap()
	isBlock := false
	for _, r := range cp.Rules {
		ctx := action.Context{
			Request: req,
			RuleId:  r.Rule.ID,
		}
		ok, err := r.Rule.Expr.Match(reqMap)
		if err != nil {
			return false, err
		}
		if ok {
			effects := action.ExecuteActions(r.Actions, &ctx)
			err = action.ApplyEffects(effects, logger, bl)
			if err != nil {
				return false, err
			}
			if !isBlock {
				if effects.Block {
					isBlock = true
				}
			}
		}
	}
	return isBlock, nil
}
