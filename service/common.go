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

func findByName[T any](deps *config.MongoDeps, dbName, collectionName, entityName string) (*T, error) {
	ctx, cancel := deps.Ctx()
	defer cancel()

	collection := deps.Client.Database(dbName).Collection(collectionName)
	var result T
	err := collection.FindOne(ctx, bson.M{"name": entityName}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("document with name %s not found", entityName)
		}
		return nil, fmt.Errorf("failed to find document by name %s: %w", entityName, err)
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

func add[T any](deps *config.MongoDeps, dbName, collectionName string, entity T) (primitive.ObjectID, error) {
	ctx, cancel := deps.Ctx()
	defer cancel()

	collection := deps.Client.Database(dbName).Collection(collectionName)
	res, err := collection.InsertOne(ctx, entity)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to insert entity %w, ", err)
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("inserted ID is not ObjectID (got %T)", res.InsertedID)
	}

	return oid, nil
}
