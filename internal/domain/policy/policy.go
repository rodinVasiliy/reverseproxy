package policy

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Policy struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name"`
	WL            []string           `bson:"wl"`        // Список префиксов(подсетей), которые в белом списке
}


func (p *Policy) GetID() primitive.ObjectID {
	return p.ID
}
