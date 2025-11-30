package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type HasID interface {
	GetID() primitive.ObjectID
}
