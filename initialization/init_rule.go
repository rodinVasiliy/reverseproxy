package initialization

import (
	"context"
	"fmt"
	action "reverseproxy/internal/model/action"
	parsedrequest "reverseproxy/internal/model/parsed_request"
	rule "reverseproxy/internal/model/rule"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getDefaultRules(as *action.Service) ([]rule.Rule, error) {
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
	actions, err := as.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find actions %w", err)
	}
	actionsIds := make([]primitive.ObjectID, len(actions))
	for i, action := range actions {
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

// / дефолтные правила
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
		Raw:                  "^Mozilla\\/5.0.+",
	}
	exprDoc, err := rule.ExprToDoc(&cond)
	if err != nil {
		return nil, fmt.Errorf("failed to get Block By UA Rule ExprDoc %w", err)
	}
	return &exprDoc, nil
}
