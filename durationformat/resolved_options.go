package durationformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

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
	Locale              locale.Locale `json:"locale"`
	NumberingSystem     string        `json:"numberingSystem"`
	Style               Style         `json:"style"`
	Years               UnitStyle     `json:"years"`
	YearsDisplay        Display       `json:"yearsDisplay"`
	Months              UnitStyle     `json:"months"`
	MonthsDisplay       Display       `json:"monthsDisplay"`
	Weeks               UnitStyle     `json:"weeks"`
	WeeksDisplay        Display       `json:"weeksDisplay"`
	Days                UnitStyle     `json:"days"`
	DaysDisplay         Display       `json:"daysDisplay"`
	Hours               UnitStyle     `json:"hours"`
	HoursDisplay        Display       `json:"hoursDisplay"`
	Minutes             UnitStyle     `json:"minutes"`
	MinutesDisplay      Display       `json:"minutesDisplay"`
	Seconds             UnitStyle     `json:"seconds"`
	SecondsDisplay      Display       `json:"secondsDisplay"`
	Milliseconds        UnitStyle     `json:"milliseconds"`
	MillisecondsDisplay Display       `json:"millisecondsDisplay"`
	Microseconds        UnitStyle     `json:"microseconds"`
	MicrosecondsDisplay Display       `json:"microsecondsDisplay"`
	Nanoseconds         UnitStyle     `json:"nanoseconds"`
	NanosecondsDisplay  Display       `json:"nanosecondsDisplay"`
	// Nil when ECMA-402 omits fractionalDigits because the option was absent.
	FractionalDigits *int `json:"fractionalDigits,omitempty"`
}

func (f *DurationFormat) ResolvedOptions() ResolvedOptions {
	resolved := f.resolved
	resolved.FractionalDigits = ecma402.CloneResolvedScalar(resolved.FractionalDigits)
	return resolved
}

func resolvedOptionsForDurationFormat(loc locale.Locale, numberingSystem string, cfg config, unitOptions [unitCount]resolvedUnitConfig) ResolvedOptions {
	resolved := ResolvedOptions{
		Locale:          loc,
		NumberingSystem: numberingSystem,
		Style:           Style(cfg.style),
	}
	for _, spec := range durationUnitSpecs[:] {
		resolved.setUnitOption(spec.index, unitOptions[spec.index])
	}
	if cfg.hasFractionalDigits {
		resolved.FractionalDigits = ecma402.ResolvedScalar(cfg.fractionalDigits)
	}
	return resolved
}

func (resolved *ResolvedOptions) setUnitOption(index unitIndex, opt resolvedUnitConfig) {
	style := publicUnitStyle(opt.style)
	switch index {
	case yearsIndex:
		resolved.Years = style
		resolved.YearsDisplay = opt.display
	case monthsIndex:
		resolved.Months = style
		resolved.MonthsDisplay = opt.display
	case weeksIndex:
		resolved.Weeks = style
		resolved.WeeksDisplay = opt.display
	case daysIndex:
		resolved.Days = style
		resolved.DaysDisplay = opt.display
	case hoursIndex:
		resolved.Hours = style
		resolved.HoursDisplay = opt.display
	case minutesIndex:
		resolved.Minutes = style
		resolved.MinutesDisplay = opt.display
	case secondsIndex:
		resolved.Seconds = style
		resolved.SecondsDisplay = opt.display
	case millisecondsIndex:
		resolved.Milliseconds = style
		resolved.MillisecondsDisplay = opt.display
	case microsecondsIndex:
		resolved.Microseconds = style
		resolved.MicrosecondsDisplay = opt.display
	case nanosecondsIndex:
		resolved.Nanoseconds = style
		resolved.NanosecondsDisplay = opt.display
	case unitCount:
	}
}

func publicUnitStyle(style UnitStyle) UnitStyle {
	if style == fractionalUnitStyle {
		return NumericUnitStyle
	}
	return style
}
