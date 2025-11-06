package config

import (
	"context"
	"fmt"
	"path/filepath"

	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDeps struct {
	Config  MongoConfig
	Client  *mongo.Client
	Timeout time.Duration
}

func (m *MongoDeps) Ctx() (context.Context, context.CancelFunc) {
	timeout := m.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}

func NewMongoDeps() (*MongoDeps, error) {
	mongoConfigFilePath := filepath.Join("config", "mongo_config", "config.yaml")
	mongoConfig, err := LoadConfig(mongoConfigFilePath)
	if err != nil {
		fmt.Errorf("failed to load mongo config %s", err)
	}

	timeout := 5 * time.Second

	mongoClient, err := getMongoClient(timeout, mongoConfig.URI)
	if err != nil {
		fmt.Errorf("failed to connect to mongo %s", err)
	}

	return &MongoDeps{Config: *mongoConfig, Timeout: timeout, Client: mongoClient}, nil
}

func getMongoClient(timeout time.Duration, uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return client, nil
}
