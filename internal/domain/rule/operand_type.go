package rule

import "fmt"

type OperandType int

const (
	OperandUnknown OperandType = -1
	OperandAnd     OperandType = iota
	OperandOr
)

func (op OperandType) String() (string, error) {
	switch op {
	case OperandAnd:
		return "and", nil
	case OperandOr:
		return "or", nil
	default:
		return "", fmt.Errorf("unknown operator: %d", op)
	}
}

func OperatorFromString(operator string) (OperandType, error) {
	switch operator {
	case "and":
		return OperandAnd, nil
	case "or":
		return OperandOr, nil
	default:
		return OperandUnknown, fmt.Errorf("unknown operand type %s", operator)
	}
}
