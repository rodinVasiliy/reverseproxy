package ssl

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SSLConfiguration struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Name         string             `bson:"name"`
	CertFileName string             `bson:"cert"`
	KeyFileName  string             `bson:"key"`
}

func (sslConfig *SSLConfiguration) GetID() primitive.ObjectID {
	return sslConfig.ID
}
