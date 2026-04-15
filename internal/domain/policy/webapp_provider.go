package policy

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebappProvider interface {
	FindByPolicyId(id primitive.ObjectID, ctx context.Context) ([]string, error)
}
