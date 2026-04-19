package rule

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rule struct {
	ID      primitive.ObjectID   `bson:"_id,omitempty"`
	Enabled bool                 `bson:"enabled"` // Включено ли правило
	Name    string               `bson:"name"`    // Название правила
	Expr    ExprDoc              `bson:"expr"`    // Набор условий правила(точнее его версия, которая может храниться в базе)
	Actions []primitive.ObjectID `bson:"actions"` // Список actions(их id)
}

// Match Возвращает true, если запрос попал под правило, false - иначе
func (rule *Rule) Match(requestMap map[string]string) bool {
	Expr := BuildExpr(rule.Expr)
	return Expr.Match(requestMap)
}

func (rule *Rule) GetID() primitive.ObjectID {
	return rule.ID
}
