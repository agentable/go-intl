package durationformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
)

const durationFormatOwner = "durationformat"

var (
	durationStyleValues = [...]string{
		string(LongStyle),
		string(ShortStyle),
		string(NarrowStyle),
		string(DigitalStyle),
	}
	durationDisplayValues = [...]string{
		string(AutoDisplay),
		string(AlwaysDisplay),
	}
	durationDateUnitStyleValues = [...]string{
		string(LongUnitStyle),
		string(ShortUnitStyle),
		string(NarrowUnitStyle),
	}
	durationTimeUnitStyleValues = [...]string{
		string(LongUnitStyle),
		string(ShortUnitStyle),
		string(NarrowUnitStyle),
		string(NumericUnitStyle),
		string(TwoDigitUnitStyle),
	}
	durationSubsecondUnitStyleValues = [...]string{
		string(LongUnitStyle),
		string(ShortUnitStyle),
		string(NarrowUnitStyle),
		string(NumericUnitStyle),
	}
)

type Options struct {
	LocaleMatcher       *string
	NumberingSystem     *string
	Style               *string
	Years               *string
	YearsDisplay        *string
	Months              *string
	MonthsDisplay       *string
	Weeks               *string
	WeeksDisplay        *string
	Days                *string
	DaysDisplay         *string
	Hours               *string
	HoursDisplay        *string
	Minutes             *string
	MinutesDisplay      *string
	Seconds             *string
	SecondsDisplay      *string
	Milliseconds        *string
	MillisecondsDisplay *string
	Microseconds        *string
	MicrosecondsDisplay *string
	Nanoseconds         *string
	NanosecondsDisplay  *string
	FractionalDigits    *int
}

type config struct {
	localeMatcher       string
	hasLocaleMatcher    bool
	numberingSystem     string
	hasNumberingSystem  bool
	style               string
	units               [unitCount]unitConfig
	fractionalDigits    int
	hasFractionalDigits bool
}

type unitConfig struct {
	style      string
	hasStyle   bool
	display    string
	hasDisplay bool
}

type resolvedUnitConfig struct {
	style   UnitStyle
	display Display
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		style:         string(ShortStyle),
	}
}

func applyOptions(cfg *config, opts Options) {
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.hasLocaleMatcher, opts.LocaleMatcher)
	ecma402.ApplyOptionInput(&cfg.numberingSystem, &cfg.hasNumberingSystem, opts.NumberingSystem)
	ecma402.ApplyOption(&cfg.style, opts.Style)
	cfg.units = optionUnitConfigs(opts)
	ecma402.ApplyOptionInput(&cfg.fractionalDigits, &cfg.hasFractionalDigits, opts.FractionalDigits)
}

func optionUnitConfigs(opts Options) [unitCount]unitConfig {
	return [unitCount]unitConfig{
		yearsIndex:        optionUnitConfig(opts.Years, opts.YearsDisplay),
		monthsIndex:       optionUnitConfig(opts.Months, opts.MonthsDisplay),
		weeksIndex:        optionUnitConfig(opts.Weeks, opts.WeeksDisplay),
		daysIndex:         optionUnitConfig(opts.Days, opts.DaysDisplay),
		hoursIndex:        optionUnitConfig(opts.Hours, opts.HoursDisplay),
		minutesIndex:      optionUnitConfig(opts.Minutes, opts.MinutesDisplay),
		secondsIndex:      optionUnitConfig(opts.Seconds, opts.SecondsDisplay),
		millisecondsIndex: optionUnitConfig(opts.Milliseconds, opts.MillisecondsDisplay),
		microsecondsIndex: optionUnitConfig(opts.Microseconds, opts.MicrosecondsDisplay),
		nanosecondsIndex:  optionUnitConfig(opts.Nanoseconds, opts.NanosecondsDisplay),
	}
}

func optionUnitConfig(style, display *string) unitConfig {
	var cfg unitConfig
	ecma402.ApplyOptionInput(&cfg.style, &cfg.hasStyle, style)
	ecma402.ApplyOptionInput(&cfg.display, &cfg.hasDisplay, display)
	return cfg
}

func durationStyleOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("style", value, durationStyleValues[:]...)
}

func durationDisplayOption(name, value string) ecma402.StringOption {
	return ecma402.RequiredStringOption(name, value, durationDisplayValues[:]...)
}

func durationUnitStyleOption(name, value string, set durationUnitStyleSet) ecma402.StringOption {
	switch set {
	case durationDateUnitStyleSet:
		return ecma402.RequiredStringOption(name, value, durationDateUnitStyleValues[:]...)
	case durationTimeUnitStyleSet:
		return ecma402.RequiredStringOption(name, value, durationTimeUnitStyleValues[:]...)
	case durationSubsecondUnitStyleSet:
		return ecma402.RequiredStringOption(name, value, durationSubsecondUnitStyleValues[:]...)
	default:
		return ecma402.RequiredStringOption(name, value)
	}
}

func (cfg config) validate(locName string) error {
	if err := ecma402.ValidateStringOptions(
		durationFormatOwner,
		locName,
		ecma402.LocaleMatcherOptionInput(cfg.localeMatcher, cfg.hasLocaleMatcher),
		durationStyleOption(cfg.style),
	); err != nil {
		return err
	}
	if err := ecma402.ValidateUnicodeTypeOptionInput(durationFormatOwner, "numberingSystem", cfg.numberingSystem, locName, cfg.hasNumberingSystem); err != nil {
		return err
	}
	return ecma402.ValidateIntegerOptions(durationFormatOwner, locName, ecma402.IntegerOption{
		Name:  "fractionalDigits",
		Value: cfg.fractionalDigits,
		Min:   0,
		Max:   9,
		Set:   cfg.hasFractionalDigits,
	})
}

func resolveUnitOptions(cfg config, locName string) ([unitCount]resolvedUnitConfig, error) {
	var out [unitCount]resolvedUnitConfig
	prevStyle := UnitStyle("")
	for _, spec := range durationUnitSpecs[:] {
		resolved, err := getDurationUnitOptions(spec, cfg.units[spec.index], Style(cfg.style), prevStyle, locName)
		if err != nil {
			return out, err
		}
		out[spec.index] = resolved
		if spec.trackPrevious {
			prevStyle = resolved.style
		}
	}
	return out, nil
}

func getDurationUnitOptions(spec durationUnitSpec, unit unitConfig, baseStyle Style, prevStyle UnitStyle, loc string) (resolvedUnitConfig, error) {
	style := UnitStyle(unit.style)
	displayDefault := AlwaysDisplay
	if !unit.hasStyle {
		switch {
		case baseStyle == DigitalStyle:
			style = spec.digitalDefault
			if !spec.clockUnit {
				displayDefault = AutoDisplay
			}
		case continuesDurationStyleChain(prevStyle):
			style = NumericUnitStyle
			if !spec.clockContinuation {
				displayDefault = AutoDisplay
			}
		default:
			style = UnitStyle(baseStyle)
			displayDefault = AutoDisplay
		}
	}
	if err := ecma402.ValidateStringOptions(durationFormatOwner, loc, durationUnitStyleOption(spec.unit, string(style), spec.styleSet)); err != nil {
		return resolvedUnitConfig{}, err
	}
	if style == NumericUnitStyle && spec.fractional {
		style = fractionalUnitStyle
		displayDefault = AutoDisplay
	}
	display := Display(unit.display)
	if !unit.hasDisplay {
		display = displayDefault
	}
	if err := ecma402.ValidateStringOptions(
		durationFormatOwner,
		loc,
		durationDisplayOption(spec.unit+"Display", string(display)),
	); err != nil {
		return resolvedUnitConfig{}, err
	}
	if err := validateDurationUnitStyle(spec.unit, style, display, prevStyle, loc); err != nil {
		return resolvedUnitConfig{}, err
	}
	if spec.clockContinuation && continuesClockStyleChain(prevStyle) {
		style = TwoDigitUnitStyle
	}
	return resolvedUnitConfig{style: style, display: display}, nil
}

func validateDurationUnitStyle(unit string, style UnitStyle, display Display, prevStyle UnitStyle, loc string) error {
	if display == AlwaysDisplay && style == fractionalUnitStyle {
		return invalidDurationUnitOption(
			unit+"Display",
			string(display),
			"auto display when formatting subsecond units as a fractional part",
			loc,
		)
	}
	if prevStyle == fractionalUnitStyle && style != fractionalUnitStyle {
		return invalidDurationUnitOption(unit, string(style), "fractional style while continuing a subsecond fractional chain", loc)
	}
	if continuesClockStyleChain(prevStyle) &&
		style != fractionalUnitStyle && style != NumericUnitStyle && style != TwoDigitUnitStyle {
		return invalidDurationUnitOption(unit, string(style), "numeric, 2-digit, or fractional style while continuing a digital time chain", loc)
	}
	return nil
}

func continuesDurationStyleChain(style UnitStyle) bool {
	return style == fractionalUnitStyle || continuesClockStyleChain(style)
}

func continuesClockStyleChain(style UnitStyle) bool {
	return style == NumericUnitStyle || style == TwoDigitUnitStyle
}

func invalidDurationUnitOption(name, value, expected, loc string) error {
	return ecma402.InvalidOptionErrorExpected(durationFormatOwner, name, value, loc, expected, nil)
}
