package webapp

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebApp struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	PolicyId primitive.ObjectID `bson:"policyId"`
	SSLId    primitive.ObjectID `bson:"SSLId"`
	Upstream string             `bson:"upstream"`
	Hosts    []string           `bson:"hosts"`
}
