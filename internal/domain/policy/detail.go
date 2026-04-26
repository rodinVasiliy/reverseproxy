package policy

import "go.mongodb.org/mongo-driver/bson/primitive"

type Detail struct {
	ID    primitive.ObjectID
	Name  string
	WL    []string
	Rules []PolicyRuleDetail
}

type PolicyRuleDetail struct {
	ID      primitive.ObjectID
	Name    string
	Enabled bool
	Actions []ActionDetail
}

type ActionDetail struct {
	ID   primitive.ObjectID
	Name string
}
