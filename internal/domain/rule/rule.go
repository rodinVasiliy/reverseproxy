package rule

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rule struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty"`
	Enabled   bool                 `bson:"enabled"`   // Включено ли правило
	Name      string               `bson:"name"`      // Название правила
	Expr      ExprDoc              `bson:"expr"`      // Набор условий правила(точнее его версия, которая может храниться в базе)
	Actions   []primitive.ObjectID `bson:"actions"`   // Список actions(их id)
	Policies  []primitive.ObjectID `bson:"policies"`  // Список политик, в которых это правило есть
	Overrides []Override           `bson:"overrides"` // Список переопределений действий для правила
}

// Override : PolicyId - id политики, для которых список действий переопределен
type Override struct {
	PolicyId primitive.ObjectID   `bson:"policyId"`
	Actions  []primitive.ObjectID `bson:"actions,omitempty"`
}

// Match Возвращает true, если запрос попал под правило, false - иначе
func (rule *Rule) Match(requestMap map[string]string) (bool, error) {
	Expr, err := BuildExpr(rule.Expr)
	if err != nil {
		return false, err
	}
	return Expr.Match(requestMap)
}

func (rule *Rule) GetID() primitive.ObjectID {
	return rule.ID
}
