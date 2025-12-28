package webapp

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebApp struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`     // название вебапа
	PolicyId primitive.ObjectID `bson:"policyId"` // id политики которая будет использоваться в вебапе
	Port     int                `bson:"port"`     // порт который будет слушать nginx
	SSLId    primitive.ObjectID `bson:"SSLId"`    // id ssl конфига
	Upstream string             `bson:"upstream"` // upstream(ip) куда будут направляться запросы
	Hosts    []string           `bson:"hosts"`    // хосты для которых будет использоваться этот вебап
}

func (app *WebApp) GetID() primitive.ObjectID {
	return app.ID
}
