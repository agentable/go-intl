package relativetimeformat

import "github.com/agentable/go-intl/locale"

type LocaleMatcher string
type Style string
type Numeric string
type Unit string
type PartType string

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	LongStyle   Style = "long"
	ShortStyle  Style = "short"
	NarrowStyle Style = "narrow"

	NumericAlways Numeric = "always"
	NumericAuto   Numeric = "auto"

	Second  Unit = "second"
	Minute  Unit = "minute"
	Hour    Unit = "hour"
	Day     Unit = "day"
	Week    Unit = "week"
	Month   Unit = "month"
	Quarter Unit = "quarter"
	Year    Unit = "year"

	// PartType values emitted by Format/FormatToParts.
	// ECMA-402 §17.5.6 PartitionRelativeTimePattern produces literal parts
	// for the pattern, and embeds NumberFormat partition records for the value.
	// RelativeTimeFormat's embedded NumberFormat uses Style="decimal" with no
	// Notation, so only the number-style-neutral part types can appear.
	PartLiteral   PartType = "literal"
	PartInteger   PartType = "integer"
	PartGroup     PartType = "group"
	PartDecimal   PartType = "decimal"
	PartFraction  PartType = "fraction"
	PartPlusSign  PartType = "plusSign"
	PartMinusSign PartType = "minusSign"
	PartInfinity  PartType = "infinity"
	PartNaN       PartType = "nan"
)

type ResolvedOptions struct {
	Locale          locale.Locale `json:"locale"`
	Style           Style         `json:"style"`
	Numeric         Numeric       `json:"numeric"`
	NumberingSystem string        `json:"numberingSystem"`
}

func (f *RelativeTimeFormat) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
