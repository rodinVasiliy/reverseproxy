package action

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActionDoc будет храниться в базе данных, а по названию будут выдергивать уже сами actions с функциями
type ActionDoc struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

func (d *ActionDoc) GetID() primitive.ObjectID {
	return d.ID
}

func ActionIdsByNames(actions []ActionDoc, actionNames map[string]struct{}) []primitive.ObjectID {
	result := make([]primitive.ObjectID, 0, len(actionNames))
	for _, action := range actions {
		if _, ok := actionNames[action.Name]; ok {
			result = append(result, action.ID)
		}
	}
	return result
}
