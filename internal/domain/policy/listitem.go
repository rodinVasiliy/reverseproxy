package policy

import "go.mongodb.org/mongo-driver/bson/primitive"

// Модель для вывода списка правил
type PolicyListItem struct {
	ID      primitive.ObjectID
	Name    string
	WL      []string
	Webapps []string
}
