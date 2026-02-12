package config

import (
	"context"
	"fmt"
	"path/filepath"

	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
	mongoConfigFilePath := filepath.Join("internal", "infrastructure", "config", "mongo_config", "config.yaml")
	mongoConfig, err := LoadConfig(mongoConfigFilePath)

	if err != nil {
		return nil, fmt.Errorf("failed to load mongo config %w", err)
	}

	timeout := 5 * time.Second

	mongoClient, err := getMongoClient(timeout, mongoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo %w", err)
	}

	return &MongoDeps{Config: *mongoConfig, Timeout: timeout, Client: mongoClient}, nil
}

func getMongoClient(timeout time.Duration, mongoConfig *MongoConfig) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI(mongoConfig.URI).SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed connect to mongo %w", err)
	}

	// ping чтобы убедиться, что PRIMARY найден
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("mongo ping failed: %w", err)
	}

	return client, nil
}
