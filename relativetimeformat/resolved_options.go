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

	Second   Unit = "second"
	Seconds  Unit = "seconds"
	Minute   Unit = "minute"
	Minutes  Unit = "minutes"
	Hour     Unit = "hour"
	Hours    Unit = "hours"
	Day      Unit = "day"
	Days     Unit = "days"
	Week     Unit = "week"
	Weeks    Unit = "weeks"
	Month    Unit = "month"
	Months   Unit = "months"
	Quarter  Unit = "quarter"
	Quarters Unit = "quarters"
	Year     Unit = "year"
	Years    Unit = "years"

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
