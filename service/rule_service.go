package service

import (
	"context"
	"fmt"
	config "reverseproxy/config/mongo_config"
	rule "reverseproxy/model/rule"

	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RuleService struct {
	deps *config.MongoDeps
	as   *ActionService
}

func NewRuleService(deps *config.MongoDeps, as *ActionService) *RuleService {
	return &RuleService{deps: deps, as: as}
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

// выглядит пока стремненько, подумать как сложить в несколько строк, чтобы не раздувать код при появлении новых правил.
func (rs *RuleService) LoadRules() {
	mongoConfig := rs.deps.Config
	client := rs.deps.Client

	actions, err := FindAllActions(rs.as)
	if err != nil {
		fmt.Printf("failed to get actions from db %s", err)
	}
	actionIDs := make([]primitive.ObjectID, len(actions))
	for i, action := range actions {
		actionIDs[i] = action.ID
	}

	blockByUARuleExpr := getBlockByUARuleExpr()
	blockByUARuleExprDoc, err := rule.ExprToDoc(blockByUARuleExpr)
	if err != nil {
		fmt.Printf("failed to serialize rule %s", err)
	}
	getGeoRuleExpr := getGeoRuleExpr()
	getGeoRuleExprDoc, err := rule.ExprToDoc(getGeoRuleExpr)
	geoRule := rule.Rule{
		Enabled: true,
		Name:    "GeoBlock",
		Expr:    getGeoRuleExprDoc,
		Actions: actionIDs,
	}
	BlockByUARule := rule.Rule{
		Enabled: true,
		Name:    "BlockByUA",
		Expr:    blockByUARuleExprDoc,
		Actions: actionIDs,
	}

	db := mongoConfig.Database
	collection := client.Database(db).Collection("rule")
	ctx, cancel := rs.deps.Ctx()
	defer cancel()

	_, err = collection.InsertOne(ctx, geoRule)
	if err != nil {
		fmt.Printf("failed to insert rule to DB %s", err)
	}
	_, err = collection.InsertOne(ctx, BlockByUARule)
	if err != nil {
		fmt.Printf("failed to insert rule to DB %s", err)
	}
}

func (rs *RuleService) FindAllRules() []rule.Rule {
	mongoConfig := rs.deps.Config
	client := rs.deps.Client

	db := mongoConfig.Database
	ruleCollection := client.Database(db).Collection(RULE_COLLECTION)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Находим все документы
	cursor, err := ruleCollection.Find(ctx, bson.M{}) // bson.M{} — пустой фильтр = "всё"
	if err != nil {
		fmt.Errorf("failed to find rules %s", err)
	}
	defer cursor.Close(ctx)
	var rules []rule.Rule
	if err := cursor.All(ctx, &rules); err != nil {
		fmt.Errorf("failed to decode rules %s", err)
	}
	return rules
}
