package initialization

import (
	"fmt"
	rule "reverseproxy/model/rule"
	service "reverseproxy/service"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDefaultRules(as *service.ActionService) (*[]rule.Rule, error) {
	// expressions
	geoRuleExp := getGeoRuleExpr()
	blockByUARuleExp := getBlockByUARuleExpr()

	// expression docs
	geoRuleDoc, err := rule.ExprToDoc(geoRuleExp)
	if err != nil {
		fmt.Printf("failed to convert Exp %s to ExpDoc %s", "geo", err)
	}
	blockByUADoc, err := rule.ExprToDoc(blockByUARuleExp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Exp %s to ExpDoc %w", "blockByUADoc", err)
	}
	actions, err := as.FindAllActions()
	if err != nil {
		return nil, fmt.Errorf("failed to find actions %w", err)
	}
	actionsIds := make([]primitive.ObjectID, len(*actions))
	for i, action := range *actions {
		actionsIds[i] = action.ID
	}
	blockByGeoRule := rule.Rule{
		Enabled: true,
		Name:    "Block by Geo",
		Expr:    geoRuleDoc,
		Actions: actionsIds,
	}
	blockByUARule := rule.Rule{
		Enabled: true,
		Name:    "Block by Geo",
		Expr:    blockByUADoc,
		Actions: actionsIds,
	}
	var rules []rule.Rule
	rules = append(rules, blockByGeoRule, blockByUARule)
	return &rules, nil
}

func getGeoRuleExpr() rule.Expr {
	cond := rule.Condition{
		MatchType:            rule.MatchNotEquals,
		RequestParameterType: "countryCode",
		Raw:                  "RU",
	}
	return &cond
}

func getBlockByUARuleExpr() rule.Expr {
	cond := rule.Condition{
		MatchType:            rule.MatchRegex,
		RequestParameterType: "ua",
		Raw:                  "^Mozilla\\/5.0.+",
	}
	var expr []rule.Expr
	expr = append(expr, &cond)
	group := rule.Group{
		Operator: rule.OpenandNot,
		Children: expr,
	}
	return &group
}
