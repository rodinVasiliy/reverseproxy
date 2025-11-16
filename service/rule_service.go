package service

import (
	config "reverseproxy/config/mongo_config"
	rule "reverseproxy/model/rule"
)

type RuleService struct {
	deps *config.MongoDeps
	as   *ActionService
}

func NewRuleService(deps *config.MongoDeps, as *ActionService) *RuleService {
	return &RuleService{deps: deps, as: as}
}

func (rs *RuleService) FindAll() (*[]rule.Rule, error) {
	return findAll[rule.Rule](rs.deps, rs.deps.Config.Database, RULE_COLLECTION)
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

// TODO

// выглядит пока стремненько, подумать как сложить в несколько строк, чтобы не раздувать код при появлении новых правил.
// func (rs *RuleService) LoadRules() {
// 	mongoConfig := rs.deps.Config
// 	client := rs.deps.Client

// 	actions, err := FindAllActions(rs.as)
// 	if err != nil {
// 		fmt.Printf("failed to get actions from db %s", err)
// 	}
// 	actionIDs := make([]primitive.ObjectID, len(actions))
// 	for i, action := range actions {
// 		actionIDs[i] = action.ID
// 	}

// 	blockByUARuleExpr := getBlockByUARuleExpr()
// 	blockByUARuleExprDoc, err := rule.ExprToDoc(blockByUARuleExpr)
// 	if err != nil {
// 		fmt.Printf("failed to serialize rule %s", err)
// 	}
// 	getGeoRuleExpr := getGeoRuleExpr()
// 	getGeoRuleExprDoc, err := rule.ExprToDoc(getGeoRuleExpr)
// 	if err != nil {
// 		fmt.Printf("failed to serialize rule %s", err)
// 	}
// 	geoRule := rule.Rule{
// 		Enabled: true,
// 		Name:    "GeoBlock",
// 		Expr:    getGeoRuleExprDoc,
// 		Actions: actionIDs,
// 	}
// 	BlockByUARule := rule.Rule{
// 		Enabled: true,
// 		Name:    "BlockByUA",
// 		Expr:    blockByUARuleExprDoc,
// 		Actions: actionIDs,
// 	}

// 	db := mongoConfig.Database
// 	collection := client.Database(db).Collection("rule")
// 	ctx, cancel := rs.deps.Ctx()
// 	defer cancel()

// 	_, err = collection.InsertOne(ctx, geoRule)
// 	if err != nil {
// 		fmt.Printf("failed to insert rule to DB %s", err)
// 	}
// 	_, err = collection.InsertOne(ctx, BlockByUARule)
// 	if err != nil {
// 		fmt.Printf("failed to insert rule to DB %s", err)
// 	}
// }
