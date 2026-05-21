package ecma402pr

import pluralop "github.com/agentable/go-intl/internal/plural"

type OperandValue = pluralop.OperandValue
type OperandsRecord = pluralop.OperandsRecord

func NewOperandValue(formatted string) OperandValue {
	return pluralop.NewOperandValue(formatted)
}

func NewIntegerOperand(n int64) OperandValue {
	return pluralop.NewIntegerOperand(n)
}

func NewUnsignedIntegerOperand(n uint64) OperandValue {
	return pluralop.NewUnsignedIntegerOperand(n)
}

func GetOperands(formatted string, exponent int) OperandsRecord {
	return pluralop.GetOperands(formatted, exponent)
}

func GetIntegerOperands(n int64) OperandsRecord {
	return pluralop.GetIntegerOperands(n)
}

func GetUnsignedIntegerOperands(n uint64) OperandsRecord {
	return pluralop.GetUnsignedIntegerOperands(n)
}
