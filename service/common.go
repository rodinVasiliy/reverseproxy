package service

import (
	"errors"
	"fmt"
	config "reverseproxy/config/mongo_config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func findById[T any](deps *config.MongoDeps, dbName string, collectionName string, id primitive.ObjectID) (*T, error) {
	ctx, cancel := deps.Ctx()
	defer cancel()

	collection := deps.Client.Database(dbName).Collection(collectionName)

	var result T
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("document with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to find document by ID %s: %w", id, err)
	}

	return &result, nil
}

func findAll[T any](deps *config.MongoDeps, dbName, collectionName string) (*[]T, error) {
	ctx, cancel := deps.Ctx()
	defer cancel()

	collection := deps.Client.Database(dbName).Collection(collectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find all documents in %s, %s", dbName, collectionName)
	}
	defer cursor.Close(ctx)
	var documents []T
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode actions %s", err)
	}

	return &documents, nil
}

func add[T any](deps *config.MongoDeps, dbName, collectionName string, entity T) (interface{}, error) {
	ctx, cancel := deps.Ctx()
	defer cancel()

	collection := deps.Client.Database(dbName).Collection(collectionName)
	id, err := collection.InsertOne(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to insert entity %w, ", err)
	}
	return id, nil
}

// TODO добавить add
