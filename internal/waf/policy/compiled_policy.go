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

func CompilePolicy(p policy.Policy, ruleEngine *rule.RuleEngine,
	actionEngine *action.ActionEngine) (*CompiledPolicy, error) {
	compiledPolicy := &CompiledPolicy{
		ID:   p.ID,
		Name: p.Name,
		WL:   p.WL,
	}
	for _, ref := range p.Rules {
		baseRule, ok := ruleEngine.Get(ref.RuleID)
		if !ok {
			return nil, fmt.Errorf("rule not found: %s", ref.RuleID.Hex())
		}

		var actions []action.Action

		if len(ref.Actions) == 0 {
			// используем дефолтные
			actions = baseRule.Actions
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
		}

		compiledPolicy.Rules = append(compiledPolicy.Rules, PolicyRule{
			Rule:    baseRule,
			Actions: actions,
		})
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
