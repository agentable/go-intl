package ecma402

import (
	"github.com/agentable/go-intl/internal/decimal"
	ecma402types "github.com/agentable/go-intl/internal/ecma402/types"
)

func ToIntlMathematicalValue(value any) (ecma402types.MathematicalValue, error) {
	return decimal.ToIntlMathematicalValue(value)
}
