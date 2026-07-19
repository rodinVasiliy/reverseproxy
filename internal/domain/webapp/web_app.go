package webapp

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebApp Port - то, что слушает nginx, Upstream - ip + port, куда мы будем отправлять запрос(порты могут быть разные)
type WebApp struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`     //
	PolicyId primitive.ObjectID `bson:"policyId"` //
	Port     int                `bson:"port"`     // Порт, который будет слушать nginx
	SSLId    primitive.ObjectID `bson:"SSLId"`    //
	Upstream string             `bson:"upstream"` // (ip:port) куда будут направляться запросы
	Hosts    []string           `bson:"hosts"`    // Хосты, для которых будет использоваться это приложение(webapp)
}

func (app *WebApp) GetID() primitive.ObjectID {
	return app.ID
}
