package main

import (
	config "reverseproxy/config"
	mconfig "reverseproxy/config/mongo_config"

	"go.mongodb.org/mongo-driver/mongo"
)

func InItDB(mongoConfig *mconfig.MongoConfig, mongoClient *mongo.Client) {
	config.LoadActionsToDB(*mongoConfig, mongoClient)
	config.LoadRules(*mongoConfig, mongoClient)
	// TODO добавить в загрузку базы дефолтные правила и дефолтную политику.
}
