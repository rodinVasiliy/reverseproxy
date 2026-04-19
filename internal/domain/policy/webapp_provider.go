package policy

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebappProvider interface {
	FindByPolicyIDs(ids []primitive.ObjectID, ctx context.Context) (map[primitive.ObjectID][]string, error)
	FindByPolicyId(id primitive.ObjectID, ctx context.Context) ([]string, error)
}
