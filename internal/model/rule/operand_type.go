package rule

import "fmt"

type OpenandType int

const (
	OpenandUnknown OpenandType = -1
	OpenandAnd     OpenandType = iota
	OperandOr
	OpenandNot
)

func (op OpenandType) String() string {
	switch op {
	case OpenandAnd:
		return "and"
	case OpenandNot:
		return "not"
	case OperandOr:
		return "or"
	default:
		return "unknown"
	}
}

func OperatorFromString(operator string) OpenandType {
	switch operator {
	case "and":
		return OpenandAnd
	case "or":
		return OperandOr
	case "not":
		return OpenandNot
	default:
		fmt.Printf("unknown operand type %s", operator)
		return -1
	}
}
