package rule

import (
	"fmt"
)

type ExprDoc struct {
	NodeType string    `bson:"nodeType"` // "condition" | "group"
	IsNot    bool      `bson:"isNot"`
	Operator string    `bson:"operator,omitempty"` // если группа
	Children []ExprDoc `bson:"children,omitempty"` // если группа
	Match    string    `bson:"match,omitempty"`    // equals/in/regex
	Field    string    `bson:"field,omitempty"`    // например "ua"
	Raw      string    `bson:"value"`              // значений на которое будет матчится параметр запроса
}

func BuildExpr(doc ExprDoc) Expr {
	switch doc.NodeType {
	case "condition":
		return &Condition{
			IsNot:                doc.IsNot,
			MatchType:            MatchTypeFromString(doc.Match),
			RequestParameterType: doc.Field,
			Raw:                  doc.Raw,
		}
	case "group":
		children := make([]Expr, 0, len(doc.Children))
		for _, child := range doc.Children {
			children = append(children, BuildExpr(child))
		}
		return &Group{
			Operator: OperatorFromString(doc.Operator),
			Children: children,
		}
	default:
		fmt.Printf("unknown node type %s", doc.NodeType)
		return nil
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
		op := e.Operator.String()
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
