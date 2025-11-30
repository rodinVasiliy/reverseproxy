package action

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// будет храниться в базе данных, а по названию будут выдергивать уже сами actions с функциями
type ActionDoc struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

func (ad *ActionDoc) GetID() primitive.ObjectID {
	return ad.ID
}
