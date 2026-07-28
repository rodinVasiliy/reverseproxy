package rule

import (
	"fmt"
	"regexp"

	"strings"
)

type Expr interface {
	Match(requestMap map[string]string) (bool, error)
}

// MatchType - ==, !-, In, Regex
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

type Condition struct { // Реализовывает Expr
	IsNot                bool
	MatchType            MatchType
	RequestParameterType string         // IP или UA, или host и т.д.
	Raw                  string         // Исходное значение из конфигурации, Если MatchEquals - берем для проверки именно это значение
	inVals               []string       // Для MatchIn, не будет сериализоваться, нужна инициализация, чтобы поле появилось
	regex                *regexp.Regexp // Для MatchRegex, не будет сериализоваться, нужна инициализация, чтобы поле появилось
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
	case MatchEquals:
		return nil
	default:
		return fmt.Errorf("unknown condition type: %s", c.MatchType)
	}
	return nil
}

// Match Проверяет запрос на совпадение правилу
func (c *Condition) Match(requestMap map[string]string) (bool, error) {

	// TODO а что, если нет значения?
	for key, value := range requestMap {
		fmt.Printf("%v:%v ||", key, value)
	}
	value, ok := requestMap[c.RequestParameterType]
	if !ok {
		fmt.Printf("value of %v not found\n", c.RequestParameterType)
		return false, nil
	}
	fmt.Printf("value: %v ||| param: %v\n", value, c.Raw)
	result := false
	switch c.MatchType {
	case MatchEquals:
		if value == c.Raw {
			result = true
		}
	case MatchNotEquals:
		if value != c.Raw {
			result = true
		}
	case MatchIn:
		for _, v := range c.inVals {
			if value == strings.TrimSpace(v) {
				result = true
				break
			}
		}
	case MatchRegex:
		result = c.regex.MatchString(value)
	default:
		return false, fmt.Errorf("failed to match request, unknown condition type: %s", c.MatchType)
	}
	if c.IsNot {
		return !result, nil
	}
	return result, nil
}

type AlwaysTrueCondition struct {
}

func (aTrCond *AlwaysTrueCondition) Match(requestMap map[string]string) (bool, error) {
	return true, nil
}

type Group struct {
	IsNot    bool
	Operator OperandType
	Children []Expr // Может быть как Condition, так и другие Group
}

func (g *Group) Match(requestMap map[string]string) (bool, error) {
	var result bool
	switch g.Operator {
	case OperandAnd:
		result = true
		for _, child := range g.Children {
			ok, err := child.Match(requestMap)
			if err != nil {
				return false, err
			}
			if !ok {
				result = false
				break
			}
		}
	case OperandOr:
		result = false
		for _, child := range g.Children {
			ok, err := child.Match(requestMap)
			if err != nil {
				return false, err
			}
			if ok {
				result = true
				break
			}
		}
	default:
		panic(fmt.Sprintf("unsupported operator: %d", g.Operator))
	}

	if g.IsNot {
		return !result, nil
	}
	return result, nil
}
