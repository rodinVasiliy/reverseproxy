package config

import (
	"context"
	"fmt"
	config "reverseproxy/config/mongo_config"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
)

func getGeoRuleExpr() Expr {
	cond := Condition{
		MatchType:            MatchNotEquals,
		RequestParameterType: "countryCode",
		Raw:                  "RU",
	}
	return &cond
}

func getBlockByUARuleExpr() Expr {
	cond := Condition{
		MatchType:            MatchRegex,
		RequestParameterType: "ua",
		Raw:                  "^Mozilla\\/5.0.+",
	}
	var expr []Expr
	expr = append(expr, &cond)
	group := Group{
		Operator: OpenandNot,
		Children: expr,
	}
	return &group
}

// выглядит пока стремненько, подумать как сложить в несколько строк, чтобы не раздувать код при появлении новых правил.
func LoadRules(mongoConfig config.MongoConfig, client *mongo.Client) {
	actions, err := FindAllActions(mongoConfig, client)
	if err != nil {
		fmt.Printf("failed to get actions from db %s", err)
	}
	actionIDs := make([]primitive.ObjectID, len(actions))
	for i, action := range actions {
		actionIDs[i] = action.ID
	}

	blockByUARuleExpr := getBlockByUARuleExpr()
	blockByUARuleExprDoc, err := ExprToDoc(blockByUARuleExpr)
	if err != nil {
		fmt.Printf("failed to serialize rule %s", err)
	}
	getGeoRuleExpr := getGeoRuleExpr()
	getGeoRuleExprDoc, err := ExprToDoc(getGeoRuleExpr)
	geoRule := Rule{
		Enabled: true,
		Name:    "GeoBlock",
		Expr:    getGeoRuleExprDoc,
		Actions: actionIDs,
	}
	BlockByUARule := Rule{
		Enabled: true,
		Name:    "BlockByUA",
		Expr:    blockByUARuleExprDoc,
		Actions: actionIDs,
	}
	db := mongoConfig.Database
	collection := client.Database(db).Collection("rule")
	_, err = collection.InsertOne(context.TODO(), geoRule)
	if err != nil {
		fmt.Printf("failed to insert rule to DB %s", err)
	}
	_, err = collection.InsertOne(context.TODO(), BlockByUARule)
	if err != nil {
		fmt.Printf("failed to insert rule to DB %s", err)
	}
}
