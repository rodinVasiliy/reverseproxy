package policy

import (
	"fmt"
	"log"
	"net"
	"reverseproxy/internal/domain/policy"
	bl "reverseproxy/internal/infrastructure/config/bl"
	"reverseproxy/internal/waf/action"
	parsedRequest "reverseproxy/internal/waf/parsedrequest"
	"reverseproxy/internal/waf/rule"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CompiledPolicy struct {
	ID    primitive.ObjectID
	Name  string
	WL    []string // TODO поменять на WL []*net.IPNet
	Rules []PolicyRule
}

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
			Rule: &compiledRule,
		}

		// Если для политики есть переопределение, используем именно Actions из переопределения
		var actions []action.Action
		if v, ok := compiledRule.Overrides[p.ID]; ok {
			actions = v
		} else {
			actions = compiledRule.Actions
		}
		policyRule.Actions = actions
		ruleMap[compiledRule.ID] = policyRule
	}

	// Добавляем правила в политику, переопределения учитываются
	compiledPolicy.Rules = make([]PolicyRule, 0, len(ruleMap))
	for _, value := range ruleMap {
		compiledPolicy.Rules = append(compiledPolicy.Rules, value)
	}

	return compiledPolicy, nil
}

func checkInList(ip net.IP, list []string) (bool, error) {
	for i := range list {
		_, ipNet, err := net.ParseCIDR(list[i])
		if err != nil {
			return false, fmt.Errorf("failed to parse cidr %s %w", list[i], err)
		}
		if ipNet.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

// Проверяем запрос по всем правилам из политики
func (cp *CompiledPolicy) Evaluate(req *parsedRequest.ParsedRequest, logger *log.Logger, bl *bl.RedisBL) (bool, error) {
	ok, err := checkInList(req.IP, cp.WL)
	if err != nil {
		return false, fmt.Errorf("failed to check ip in list %w", err)
	}
	if ok {
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
