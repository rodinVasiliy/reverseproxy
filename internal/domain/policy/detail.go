package policy

import "go.mongodb.org/mongo-driver/bson/primitive"

type Detail struct {
	ID    primitive.ObjectID
	Name  string
	WL    []string
	Rules []RuleDetail // Правила, которые используются для этой политики
}

type RuleDetail struct {
	ID      primitive.ObjectID
	Name    string
	Enabled bool
}
