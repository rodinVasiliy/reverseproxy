package rule

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rule struct {
	ID      primitive.ObjectID   `bson:"_id,omitempty"`
	Enabled bool                 `bson:"enabled"`
	Name    string               `bson:"name"`
	Expr    ExprDoc              `bson:"expr"`    // тут нужен кастомный сериализатор
	Actions []primitive.ObjectID `bson:"actions"` // список ID из коллекции actions
}

func (rule *Rule) Match(requestMap map[string]string) bool {
	Expr := BuildExpr(rule.Expr)
	return Expr.Match(requestMap)
}
