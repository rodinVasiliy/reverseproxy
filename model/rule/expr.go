package rule

import (
	"fmt"
	"regexp"

	"strings"
)

type Expr interface {
	Match(requestMap map[string]string) bool
}

type MatchType int

const (
	MatchUnknown MatchType = -1
	MatchEquals  MatchType = iota
	MatchNotEquals
	MatchIn
	MatchRegex
)

func (m MatchType) String() string {
	switch m {
	case MatchEquals:
		return "equals"
	case MatchNotEquals:
		return "not_equals"
	case MatchIn:
		return "in"
	case MatchRegex:
		return "regex"
	default:
		return "unknown"
	}
}

func MatchTypeFromString(nodeType string) MatchType {
	switch nodeType {
	case "equals":
		return MatchEquals
	case "not_equals":
		return MatchNotEquals
	case "in":
		return MatchIn
	case "regex":
		return MatchRegex
	default:
		fmt.Printf("unknown Match type %s", nodeType)
		return -1
	}
}

type Condition struct { // реализовывыает Expr
	IsNot                bool
	MatchType            MatchType
	RequestParameterType string         // ip или UA, или host и т.д.
	Raw                  string         // исходное значение из конфигурации
	inVals               []string       // для MatchIn
	regex                *regexp.Regexp // для MatchRegex
}

func (c *Condition) Init() error {
	switch c.MatchType {
	case MatchIn:
		c.inVals = strings.Split(c.Raw, ",")
	case MatchRegex:
		re, err := regexp.Compile(c.Raw)
		if err != nil {
			return err
		}
		c.regex = re
	}
	return nil
}

// Проверка
func (c *Condition) Match(requstMap map[string]string) bool {
	value := requstMap[c.RequestParameterType]
	switch c.MatchType {
	case MatchEquals:
		return value == c.Raw
	case MatchNotEquals:
		return value != c.Raw
	case MatchIn:
		for _, v := range c.inVals {
			if value == strings.TrimSpace(v) {
				return true
			}
		}
		return false
	case MatchRegex:
		return c.regex.MatchString(value)
	}
	return false
}

type AlwaysTrueCondition struct {
}

func (aTrCond *AlwaysTrueCondition) Match(requestMap map[string]string) bool {
	return true
}

type Group struct {
	IsNot    bool
	Operator OpenandType
	Children []Expr // здесь могут быть как Condition, так и другие Group
}

func (g *Group) Match(requestMap map[string]string) bool {
	var result bool
	switch g.Operator {
	case OpenandAnd:
		result = true
		for _, child := range g.Children {
			if !child.Match(requestMap) {
				result = false
				break
			}
		}
	case OperandOr:
		result = false
		for _, child := range g.Children {
			if child.Match(requestMap) {
				result = true
				break
			}
		}
	case OpenandNot:
		result = true
		for _, child := range g.Children {
			if !child.Match(requestMap) {
				result = false
				break
			}
		}
		result = !result
	default:
		panic(fmt.Sprintf("unsupported operator: %d", g.Operator))
	}

	if g.IsNot {
		return !result
	}
	return result
}
