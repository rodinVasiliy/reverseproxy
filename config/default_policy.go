package config

import (
	"context"
	"fmt"
	config "reverseproxy/config/mongo_config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getDefaultPolicy(mongoConfig config.MongoConfig, client *mongo.Client) (Policy, error) {
	db := mongoConfig.Database
	ruleCollection := client.Database(db).Collection("rule")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Находим все документы
	cursor, err := ruleCollection.Find(ctx, bson.M{}) // bson.M{} — пустой фильтр = "всё"
	if err != nil {
		fmt.Errorf("failed to find actions %s", err)
	}
	defer cursor.Close(ctx)
	var rules []Rule
	if err := cursor.All(ctx, &rules); err != nil {
		fmt.Errorf("failed to decode actions %s", err)
	}
	var policyRuleRef []PolicyRuleRef
	for _, rule := range rules {
		policyRuleRef = append(policyRuleRef, PolicyRuleRef{RuleID: rule.ID})
	}
	defaultPolicy := Policy{WL: nil, Rules: policyRuleRef}
	return defaultPolicy, nil
}

func LoadDefaultPolicyToDB(mongoConfig config.MongoConfig, client *mongo.Client) {
	db := mongoConfig.Database
	collection := client.Database(db).Collection("policy")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	policy, err := getDefaultPolicy(mongoConfig, client)
	if err != nil {
		fmt.Printf("failed to get default policy %s", err)
		return
	}
	collection.InsertOne(ctx, policy)
}
