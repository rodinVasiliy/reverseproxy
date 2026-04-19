package policy

import "go.mongodb.org/mongo-driver/bson/primitive"

type ListItem struct {
	ID      primitive.ObjectID `json:"id"`
	Name    string             `json:"name"`
	WL      []string           `json:"wl"`
	Webapps []string           `json:"webapps"`
}
