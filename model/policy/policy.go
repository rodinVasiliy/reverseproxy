package policy

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Policy struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	WL    []string           `bson:"wl"` // список префиксов(подсетей), которые в белом списке
	Rules []PolicyRuleRef    `bson:"rules"`
}

// если Actions будет пустой - значит берем Actions из самого правила, если не пустой - значит мы переопределили их для политики.
type PolicyRuleRef struct {
	RuleID  primitive.ObjectID   `bson:"ruleId"`
	Actions []primitive.ObjectID `bson:"actions,omitempty"`
}
