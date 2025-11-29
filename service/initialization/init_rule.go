package initialization

import (
	"fmt"
	rule "reverseproxy/model/rule"
	service "reverseproxy/service"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDefaultRules(as *service.ActionService) (*[]rule.Rule, error) {

	// expression docs
	geoRuleDoc, err := getGeoRuleExprDoc()
	if err != nil {
		return nil, fmt.Errorf("failed to get geo expr %w", err)
	}
	blockByUADoc, err := getBlockByUARuleExprDoc()
	if err != nil {
		return nil, fmt.Errorf("failed to get block by ua expr %w", err)
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
		Expr:    *geoRuleDoc,
		Actions: actionsIds,
	}
	blockByUARule := rule.Rule{
		Enabled: true,
		Name:    "Block by Geo",
		Expr:    *blockByUADoc,
		Actions: actionsIds,
	}
	var rules []rule.Rule
	rules = append(rules, blockByGeoRule, blockByUARule)
	return &rules, nil
}

func getGeoRuleExprDoc() (*rule.ExprDoc, error) {
	cond := rule.Condition{
		IsNot:                false,
		MatchType:            rule.MatchNotEquals,
		RequestParameterType: "countryCode",
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
		RequestParameterType: "ua",
		Raw:                  "^Mozilla\\/5.0.+",
	}
	exprDoc, err := rule.ExprToDoc(&cond)
	if err != nil {
		return nil, fmt.Errorf("failed to get Block By UA Rule ExprDoc %w", err)
	}
	return &exprDoc, nil
}
