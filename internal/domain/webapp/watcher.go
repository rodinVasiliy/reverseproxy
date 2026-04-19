package webapp

import (
	"context"
	repository "reverseproxy/internal/infrastructure/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChangeHandler interface {
	OnCreate(ctx context.Context, app WebApp)
	OnUpdate(ctx context.Context, app WebApp)
	OnDelete(ctx context.Context, app WebApp)
}

type Watcher struct {
	collection *mongo.Collection
	handler    ChangeHandler
}

func NewWatcher(repo *repository.MongoRepository[WebApp], handler ChangeHandler) *Watcher {
	return &Watcher{
		collection: repo.Collection(),
		handler:    handler,
	}
}

func (w *Watcher) Watch(ctx context.Context) error {
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	stream, err := w.collection.Watch(ctx, mongo.Pipeline{}, opts)
	if err != nil {
		return err
	}
	defer stream.Close(ctx)

	for stream.Next(ctx) {
		var event struct {
			OperationType string `bson:"operationType"`
			FullDocument  WebApp `bson:"fullDocument"`
			DocumentKey   struct {
				ID primitive.ObjectID `bson:"_id"`
			} `bson:"documentKey"`
		}

		if err := stream.Decode(&event); err != nil {
			continue
		}

		switch event.OperationType {
		case "insert":
			w.handler.OnCreate(ctx, event.FullDocument)
		case "update":
			w.handler.OnUpdate(ctx, event.FullDocument)
		case "delete":
			w.handler.OnDelete(ctx, event.FullDocument)
		}
	}

	return stream.Err()
}
