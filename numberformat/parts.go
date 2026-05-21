package numberformat

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
