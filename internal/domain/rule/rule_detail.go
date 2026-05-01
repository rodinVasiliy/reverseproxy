package rule

import "go.mongodb.org/mongo-driver/bson/primitive"

type RuleDetail struct {
	ID                 primitive.ObjectID
	Name               string
	Enabled            bool
	Expr               *ExprDoc
	Actions            []ActionParam
	PolicyActionParams []PolicyActionParam
}

type ActionParam struct {
	ID   primitive.ObjectID
	Name string
}

type PolicyActionParam struct {
	ID      primitive.ObjectID
	Name    string
	Actions []ActionParam
}
