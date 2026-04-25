package rule

import (
	"fmt"
	"log"
)

type ExprDoc struct {
	NodeType string    `bson:"nodeType"`           // "condition" | "group"
	IsNot    bool      `bson:"isNot"`              // Если выставлено - значит при true вернет false и наоборот
	Operator string    `bson:"operator,omitempty"` //
	Children []ExprDoc `bson:"children,omitempty"` // Если группа
	Match    string    `bson:"match,omitempty"`    // equals/in/regex
	Field    string    `bson:"field,omitempty"`    // Поле которое будет проверяться, например "ua"
	Raw      string    `bson:"value"`              // Значений на которое будет матчится параметр запроса
}

// BuildExpr Преобразовывает expressionDoc в Expression
func BuildExpr(doc ExprDoc) (Expr, error) {
	switch doc.NodeType {
	case "condition":
		return &Condition{
			IsNot:                doc.IsNot,
			MatchType:            MatchTypeFromString(doc.Match),
			RequestParameterType: doc.Field,
			Raw:                  doc.Raw,
		}, nil
	case "group":
		children := make([]Expr, 0, len(doc.Children))
		for _, child := range doc.Children {
			child, err := BuildExpr(child)
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		}
		op, err := OperatorFromString(doc.Operator)
		if err != nil {
			return nil, err
		}
		return &Group{
			Operator: op,
			Children: children,
		}, nil
	default:
		log.Printf("unknown node type in buildExpr function: %s", doc.NodeType)
		return nil, fmt.Errorf("unknown node type")
	}
}

func ExprToDoc(expr Expr) (ExprDoc, error) {
	switch e := expr.(type) {
	case *Condition:
		match := e.MatchType.String()
		if match == "unknown" {
			return ExprDoc{}, fmt.Errorf("unknown MatchType: %v", e.MatchType)
		}
		return ExprDoc{
			NodeType: "condition",
			IsNot:    e.IsNot,
			Match:    match,
			Field:    e.RequestParameterType,
			Raw:      e.Raw,
		}, nil

	case *Group:
		op, err := e.Operator.String()
		if err != nil {
			return ExprDoc{}, err
		}
		if op == "unknown" {
			return ExprDoc{}, fmt.Errorf("unknown Operator: %v", e.Operator)
		}
		children := make([]ExprDoc, 0, len(e.Children))
		for _, child := range e.Children {
			childDoc, err := ExprToDoc(child)
			if err != nil {
				return ExprDoc{}, err
			}
			children = append(children, childDoc)
		}

		return ExprDoc{
			NodeType: "group",
			IsNot:    e.IsNot,
			Operator: op,
			Children: children,
		}, nil

	default:
		return ExprDoc{}, fmt.Errorf("unsupported Expr type: %T", expr)
	}
}
