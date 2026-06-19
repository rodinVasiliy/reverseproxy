package rule

import "go.mongodb.org/mongo-driver/bson/primitive"

type RuleDetail struct {
	ID        primitive.ObjectID
	Name      string
	Enabled   bool
	Expr      *ExprDoc
	Actions   []ActionParam
	Policies  []PolicyParam // Политики, в которых есть это правило
	Overrides []OverrideParam
}

type PolicyParam struct {
	ID   primitive.ObjectID // ID политики
	Name string             // Имя политики
}

type ActionParam struct {
	ID   primitive.ObjectID // ID действия
	Name string             // Имя действия
}

type OverrideParam struct {
	ID      primitive.ObjectID // ID политики
	Name    string             // Имя политики
	Actions []ActionParam      // список действий(id действия + название)
}
