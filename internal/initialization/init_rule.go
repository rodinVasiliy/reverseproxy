package initialization

import (
	"context"
	"fmt"
	actionDoc "reverseproxy/internal/domain/action"
	"reverseproxy/internal/domain/rule"
	"reverseproxy/internal/waf/action"
	parsedrequest "reverseproxy/internal/waf/parsed_request"
)

var (
	GeoRuleName    = "Block by Geo IP"
	UaRuleName     = "Block By UA"
	BitrixRuleName = "Block Bitrix Access"
)

func getDefaultRules(as *actionDoc.Service) ([]rule.Rule, error) {
	ctx := context.Background()
	// expression docs
	geoRuleDoc, err := getGeoRuleExprDoc()
	if err != nil {
		return nil, fmt.Errorf("failed to get geo expr %w", err)
	}
	blockByUADoc, err := getBlockByUARuleExprDoc()
	if err != nil {
		return nil, fmt.Errorf("failed to get block by ua expr %w", err)
	}
	blockBitrixDoc, err := getBlockBitrixRule()
	if err != nil {
		return nil, fmt.Errorf("failed to get block bitrix expr %w", err)
	}
	actions, err := as.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find actions %w", err)
	}
	logAndBlockActionNames := map[string]struct{}{
		action.LogToDbActionName:      {},
		action.BlockRequestActionName: {},
	}
	logAndBlockActionIds := actionDoc.ActionIdsByNames(actions, logAndBlockActionNames)
	logAndBlockAndSendToBlActionNames := map[string]struct{}{
		action.LogToDbActionName:      {},
		action.SendToBlActionName:     {},
		action.BlockRequestActionName: {},
	}
	logAndBlockAndSendToBlActionIds := actionDoc.ActionIdsByNames(actions, logAndBlockAndSendToBlActionNames)

	blockByGeoRule := rule.Rule{
		Enabled: true,
		Name:    GeoRuleName,
		Expr:    *geoRuleDoc,
		Actions: logAndBlockActionIds,
	}
	blockByUARule := rule.Rule{
		Enabled: true,
		Name:    UaRuleName,
		Expr:    *blockByUADoc,
		Actions: logAndBlockActionIds,
	}
	blockBitrixRule := rule.Rule{
		Enabled: true,
		Name:    BitrixRuleName,
		Expr:    *blockBitrixDoc,
		Actions: logAndBlockAndSendToBlActionIds,
	}
	var rules []rule.Rule
	rules = append(rules, blockByGeoRule, blockByUARule, blockBitrixRule)
	return rules, nil
}

func loadRulesToDB(rs *rule.Service, rules []rule.Rule) error {
	for _, r := range rules {
		_, err := rs.Insert(context.Background(), r)
		if err != nil {
			return fmt.Errorf("failed to insert rule %s :%w", r.Name, err)
		}
	}
	return nil
}

// Дефолтные правила(именно их логическая часть)
func getGeoRuleExprDoc() (*rule.ExprDoc, error) {
	cond := rule.Condition{
		IsNot:                false,
		MatchType:            rule.MatchNotEquals,
		RequestParameterType: parsedrequest.CountryCode,
		Raw:                  "RU",
	}
	exprDoc, err := rule.ExprToDoc(&cond)
	if err != nil {
		return nil, fmt.Errorf("failed to get Geo Rule ExprDoc %w", err)
	}
	return &exprDoc, nil
}

func getBlockByUARuleExprDoc() (*rule.ExprDoc, error) {
	cond := rule.Condition{
		IsNot:                true,
		MatchType:            rule.MatchRegex,
		RequestParameterType: parsedrequest.UA,
		Raw:                  `^(Mozilla\/5\.0|Opera\/|Chrome\/|Safari\/|Firefox\/)`,
	}
	exprDoc, err := rule.ExprToDoc(&cond)
	if err != nil {
		return nil, fmt.Errorf("failed to get Block By UA Rule ExprDoc %w", err)
	}
	return &exprDoc, nil
}

func getBlockBitrixRule() (*rule.ExprDoc, error) {
	cond := rule.Condition{
		IsNot:                false,
		MatchType:            rule.MatchEquals,
		RequestParameterType: parsedrequest.PATH,
		Raw:                  "/bitrix",
	}
	exprDoc, err := rule.ExprToDoc(&cond)
	if err != nil {
		return nil, fmt.Errorf("failed to get Block Bitrix ExprDoc %w", err)
	}
	return &exprDoc, nil
}
