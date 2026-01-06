package initialization

import (
	"context"
	"fmt"
	actionDoc "reverseproxy/internal/domain/action"
	rule "reverseproxy/internal/domain/rule"
	action "reverseproxy/internal/waf/action"
	parsedrequest "reverseproxy/internal/waf/parsed_request"
)

var (
	GEO_RULE_NAME    = "Block by Geo IP"
	UA_RULE_NAME     = "Block By UA"
	BITRIX_RULE_NAME = "Block Bitrix Access"
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
		action.LOG_TO_DB_ACTION_NAME:     {},
		action.BLOCK_REQUEST_ACTION_NAME: {},
	}
	logAndBlockActionIds := actionDoc.ActionIdsByNames(actions, logAndBlockActionNames)
	logAndBlockAndSendToBlActionNames := map[string]struct{}{
		action.LOG_TO_DB_ACTION_NAME:     {},
		action.SEND_TO_BL_ACTION_NAME:    {},
		action.BLOCK_REQUEST_ACTION_NAME: {},
	}
	logAndBlockAndSendToBlActionIds := actionDoc.ActionIdsByNames(actions, logAndBlockAndSendToBlActionNames)

	blockByGeoRule := rule.Rule{
		Enabled: true,
		Name:    GEO_RULE_NAME,
		Expr:    *geoRuleDoc,
		Actions: logAndBlockActionIds,
	}
	blockByUARule := rule.Rule{
		Enabled: true,
		Name:    UA_RULE_NAME,
		Expr:    *blockByUADoc,
		Actions: logAndBlockActionIds,
	}
	blockBitrixRule := rule.Rule{
		Enabled: true,
		Name:    BITRIX_RULE_NAME,
		Expr:    *blockBitrixDoc,
		Actions: logAndBlockAndSendToBlActionIds,
	}
	var rules []rule.Rule
	rules = append(rules, blockByGeoRule, blockByUARule, blockBitrixRule)
	return rules, nil
}

func loadRulesToDB(rs *rule.Service, rules []rule.Rule) error {
	for _, rule := range rules {
		_, err := rs.Insert(context.Background(), rule)
		if err != nil {
			return fmt.Errorf("failed to insert rule %s :%w", rule.Name, err)
		}
	}
	return nil
}

// дефолтные правила(именно их логическая часть)
func getGeoRuleExprDoc() (*rule.ExprDoc, error) {
	cond := rule.Condition{
		IsNot:                false,
		MatchType:            rule.MatchNotEquals,
		RequestParameterType: parsedrequest.COUNTRY_CODE,
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
