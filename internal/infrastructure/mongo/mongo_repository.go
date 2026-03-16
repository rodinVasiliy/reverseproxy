package mongorepository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrNotFound = errors.New("document not found")

type MongoRepository[T any] struct {
	client         *mongo.Client
	dbName         string
	collectionName string
}

func NewMongoRepositoy[T any](client *mongo.Client, dbName, collectionName string) *MongoRepository[T] {
	return &MongoRepository[T]{
		client:         client,
		dbName:         dbName,
		collectionName: collectionName,
	}
}

func (r *MongoRepository[T]) Collection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(r.collectionName)
}

// FindOne возвращает ссылку на документ по фильтру filter
//
// Возвращает :
//   - ErrNotFound если документа не существует
//   - любую другую ошибку при ошибки с базой данных
func (r *MongoRepository[T]) FindOne(ctx context.Context, filter any) (*T, error) {
	var result T
	err := r.Collection().FindOne(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// FindByID возвращает документ по id
//
// Возвращает:
//   - ErrNotFound если документа не существует
//   - любую другую ошибку при ошибки с базой данных
func (r *MongoRepository[T]) FindById(ctx context.Context, id primitive.ObjectID) (*T, error) {
	return r.FindOne(ctx, bson.M{"_id": id})
}

func (r *MongoRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	cur, err := r.Collection().Find(ctx, bson.M{}) // cursor
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var result []T
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// возвращает primitive.ObjectID либо ошибку, если не получилось вставить или id не того типа
func (r *MongoRepository[T]) Insert(ctx context.Context, entity T) (primitive.ObjectID, error) {
	res, err := r.Collection().InsertOne(ctx, entity)
	if err != nil {
		return primitive.NilObjectID, err
	}
	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("result is not primitive.ObjectID")
	}
	return id, nil
}

type WithID interface {
	GetID() primitive.ObjectID
}

func (r *MongoRepository[T]) Delete(ctx context.Context, entity WithID) error {
	id := entity.GetID()
	if id.IsZero() {
		return fmt.Errorf("entity ID is empty")
	}
	res, err := r.Collection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MongoRepository[T]) Update(ctx context.Context, entity WithID) error {
	id := entity.GetID()
	if id.IsZero() {
		return fmt.Errorf("entity ID is empty")
	}
	_, err := r.Collection().ReplaceOne(ctx, bson.M{"_id": id}, entity)
	return err
}

func (r *MongoRepository[T]) FindMany(ctx context.Context, filter any) ([]T, error) {
	cur, err := r.Collection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var result []T
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}
