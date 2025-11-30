package sslconfig

import "go.mongodb.org/mongo-driver/bson/primitive"

type SSLConfiguration struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string
	CertPath string
	KeyPath  string
}

func (sslConfig *SSLConfiguration) GetID() primitive.ObjectID {
	return sslConfig.ID
}
