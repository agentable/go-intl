package numberformat

import "strings"

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
}

type RangePart struct {
	Type   PartType    `json:"type"`
	Value  string      `json:"value"`
	Source RangeSource `json:"source"`
}

type RangeSource string

const (
	SourceStartRange RangeSource = "startRange"
	SourceShared     RangeSource = "shared"
	SourceEndRange   RangeSource = "endRange"
)

type PartType string

const (
	PartInteger           PartType = "integer"
	PartGroup             PartType = "group"
	PartDecimal           PartType = "decimal"
	PartFraction          PartType = "fraction"
	PartCurrency          PartType = "currency"
	PartPercentSign       PartType = "percentSign"
	PartMinusSign         PartType = "minusSign"
	PartPlusSign          PartType = "plusSign"
	PartNaN               PartType = "nan"
	PartInfinity          PartType = "infinity"
	PartUnit              PartType = "unit"
	PartLiteral           PartType = "literal"
	PartExponentSeparator PartType = "exponentSeparator"
	PartExponentMinusSign PartType = "exponentMinusSign"
	PartExponentInteger   PartType = "exponentInteger"
	PartCompact           PartType = "compact"
	PartApproximatelySign PartType = "approximatelySign"
)

func partsText(parts []Part) string {
	size := 0
	for _, part := range parts {
		size += len(part.Value)
	}
	var b strings.Builder
	b.Grow(size)
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}
