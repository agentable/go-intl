package durationformat

import "github.com/agentable/go-intl/locale"

type LocaleMatcher string
type Style string
type UnitStyle string
type Display string
type Unit string
type PartType string

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	LongStyle    Style = "long"
	ShortStyle   Style = "short"
	NarrowStyle  Style = "narrow"
	DigitalStyle Style = "digital"

	LongUnitStyle     UnitStyle = "long"
	ShortUnitStyle    UnitStyle = "short"
	NarrowUnitStyle   UnitStyle = "narrow"
	NumericUnitStyle  UnitStyle = "numeric"
	TwoDigitUnitStyle UnitStyle = "2-digit"

	AlwaysDisplay Display = "always"
	AutoDisplay   Display = "auto"

	Year        Unit = "year"
	Month       Unit = "month"
	Week        Unit = "week"
	Day         Unit = "day"
	Hour        Unit = "hour"
	Minute      Unit = "minute"
	Second      Unit = "second"
	Millisecond Unit = "millisecond"
	Microsecond Unit = "microsecond"
	Nanosecond  Unit = "nanosecond"

	// PartType values emitted by Format/FormatToParts.
	// ECMA-402 §18.5.5 PartitionDurationFormatPattern combines list-pattern
	// literals, digital separators, and embedded NumberFormat partition records.
	// DurationFormat's embedded NumberFormat uses Style="unit" (long/short/narrow)
	// or Style="decimal" (2-digit/numeric); other notations are not exposed,
	// so only the constants below can appear.
	PartLiteral   PartType = "literal"
	PartInteger   PartType = "integer"
	PartGroup     PartType = "group"
	PartDecimal   PartType = "decimal"
	PartFraction  PartType = "fraction"
	PartPlusSign  PartType = "plusSign"
	PartMinusSign PartType = "minusSign"
	PartInfinity  PartType = "infinity"
	PartNaN       PartType = "nan"
	PartUnit      PartType = "unit"
)

const fractionalUnitStyle UnitStyle = "fractional"

type ResolvedOptions struct {
	// Locale is the resolved locale. Mirrors Intl.DurationFormat resolved option "locale".
	Locale locale.Locale `json:"locale"`
	// NumberingSystem is the resolved numbering system. Mirrors Intl.DurationFormat resolved option "numberingSystem".
	NumberingSystem string `json:"numberingSystem"`
	// Style is the resolved overall style. Mirrors Intl.DurationFormat resolved option "style".
	Style Style `json:"style"`
	// Years is the resolved years unit style. Mirrors Intl.DurationFormat resolved option "years".
	Years UnitStyle `json:"years"`
	// YearsDisplay is the resolved years display. Mirrors Intl.DurationFormat resolved option "yearsDisplay".
	YearsDisplay Display `json:"yearsDisplay"`
	// Months is the resolved months unit style. Mirrors Intl.DurationFormat resolved option "months".
	Months UnitStyle `json:"months"`
	// MonthsDisplay is the resolved months display. Mirrors Intl.DurationFormat resolved option "monthsDisplay".
	MonthsDisplay Display `json:"monthsDisplay"`
	// Weeks is the resolved weeks unit style. Mirrors Intl.DurationFormat resolved option "weeks".
	Weeks UnitStyle `json:"weeks"`
	// WeeksDisplay is the resolved weeks display. Mirrors Intl.DurationFormat resolved option "weeksDisplay".
	WeeksDisplay Display `json:"weeksDisplay"`
	// Days is the resolved days unit style. Mirrors Intl.DurationFormat resolved option "days".
	Days UnitStyle `json:"days"`
	// DaysDisplay is the resolved days display. Mirrors Intl.DurationFormat resolved option "daysDisplay".
	DaysDisplay Display `json:"daysDisplay"`
	// Hours is the resolved hours unit style. Mirrors Intl.DurationFormat resolved option "hours".
	Hours UnitStyle `json:"hours"`
	// HoursDisplay is the resolved hours display. Mirrors Intl.DurationFormat resolved option "hoursDisplay".
	HoursDisplay Display `json:"hoursDisplay"`
	// Minutes is the resolved minutes unit style. Mirrors Intl.DurationFormat resolved option "minutes".
	Minutes UnitStyle `json:"minutes"`
	// MinutesDisplay is the resolved minutes display. Mirrors Intl.DurationFormat resolved option "minutesDisplay".
	MinutesDisplay Display `json:"minutesDisplay"`
	// Seconds is the resolved seconds unit style. Mirrors Intl.DurationFormat resolved option "seconds".
	Seconds UnitStyle `json:"seconds"`
	// SecondsDisplay is the resolved seconds display. Mirrors Intl.DurationFormat resolved option "secondsDisplay".
	SecondsDisplay Display `json:"secondsDisplay"`
	// Milliseconds is the resolved milliseconds unit style. Mirrors Intl.DurationFormat resolved option "milliseconds".
	Milliseconds UnitStyle `json:"milliseconds"`
	// MillisecondsDisplay is the resolved milliseconds display. Mirrors Intl.DurationFormat resolved option "millisecondsDisplay".
	MillisecondsDisplay Display `json:"millisecondsDisplay"`
	// Microseconds is the resolved microseconds unit style. Mirrors Intl.DurationFormat resolved option "microseconds".
	Microseconds UnitStyle `json:"microseconds"`
	// MicrosecondsDisplay is the resolved microseconds display. Mirrors Intl.DurationFormat resolved option "microsecondsDisplay".
	MicrosecondsDisplay Display `json:"microsecondsDisplay"`
	// Nanoseconds is the resolved nanoseconds unit style. Mirrors Intl.DurationFormat resolved option "nanoseconds".
	Nanoseconds UnitStyle `json:"nanoseconds"`
	// NanosecondsDisplay is the resolved nanoseconds display. Mirrors Intl.DurationFormat resolved option "nanosecondsDisplay".
	NanosecondsDisplay Display `json:"nanosecondsDisplay"`
	// FractionalDigits is the resolved fractional digit count. Mirrors Intl.DurationFormat resolved option "fractionalDigits".
	// Nil when ECMA-402 omits fractionalDigits because the option was absent.
	FractionalDigits *int `json:"fractionalDigits,omitempty"`
}

func (f *DurationFormat) ResolvedOptions() ResolvedOptions {
	resolved := f.resolved
	if resolved.FractionalDigits != nil {
		fractionalDigits := *resolved.FractionalDigits
		resolved.FractionalDigits = &fractionalDigits
	}
	return resolved
}
