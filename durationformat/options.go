package durationformat

import (
	"fmt"
	"slices"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

type Options struct {
	LocaleMatcher       LocaleMatcher
	NumberingSystem     string
	Style               Style
	Years               UnitStyle
	YearsDisplay        Display
	Months              UnitStyle
	MonthsDisplay       Display
	Weeks               UnitStyle
	WeeksDisplay        Display
	Days                UnitStyle
	DaysDisplay         Display
	Hours               UnitStyle
	HoursDisplay        Display
	Minutes             UnitStyle
	MinutesDisplay      Display
	Seconds             UnitStyle
	SecondsDisplay      Display
	Milliseconds        UnitStyle
	MillisecondsDisplay Display
	Microseconds        UnitStyle
	MicrosecondsDisplay Display
	Nanoseconds         UnitStyle
	NanosecondsDisplay  Display
	FractionalDigits    *int
}

type config struct {
	localeMatcher       string
	numberingSystem     string
	style               string
	years               unitConfig
	months              unitConfig
	weeks               unitConfig
	days                unitConfig
	hours               unitConfig
	minutes             unitConfig
	seconds             unitConfig
	milliseconds        unitConfig
	microseconds        unitConfig
	nanoseconds         unitConfig
	fractionalDigits    int
	hasFractionalDigits bool
}

type unitConfig struct {
	style   string
	display string
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
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	if opts.NumberingSystem != "" {
		cfg.numberingSystem = opts.NumberingSystem
	}
	if opts.Style != "" {
		cfg.style = string(opts.Style)
	}
	cfg.years = unitConfig{style: string(opts.Years), display: string(opts.YearsDisplay)}
	cfg.months = unitConfig{style: string(opts.Months), display: string(opts.MonthsDisplay)}
	cfg.weeks = unitConfig{style: string(opts.Weeks), display: string(opts.WeeksDisplay)}
	cfg.days = unitConfig{style: string(opts.Days), display: string(opts.DaysDisplay)}
	cfg.hours = unitConfig{style: string(opts.Hours), display: string(opts.HoursDisplay)}
	cfg.minutes = unitConfig{style: string(opts.Minutes), display: string(opts.MinutesDisplay)}
	cfg.seconds = unitConfig{style: string(opts.Seconds), display: string(opts.SecondsDisplay)}
	cfg.milliseconds = unitConfig{style: string(opts.Milliseconds), display: string(opts.MillisecondsDisplay)}
	cfg.microseconds = unitConfig{style: string(opts.Microseconds), display: string(opts.MicrosecondsDisplay)}
	cfg.nanoseconds = unitConfig{style: string(opts.Nanoseconds), display: string(opts.NanosecondsDisplay)}
	if opts.FractionalDigits != nil {
		cfg.fractionalDigits = *opts.FractionalDigits
		cfg.hasFractionalDigits = true
	}
}

func (cfg config) validate(loc locale.Locale) error {
	if check, ok := ecma402.InvalidStringOption(
		ecma402.LocaleMatcherOption(cfg.localeMatcher),
		ecma402.RequiredStringOption("style", cfg.style, string(LongStyle), string(ShortStyle), string(NarrowStyle), string(DigitalStyle)),
	); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	if cfg.numberingSystem != "" && !ecma402.IsWellFormedUnicodeType(cfg.numberingSystem) {
		return invalidOption("numberingSystem", cfg.numberingSystem, loc)
	}
	if check, ok := ecma402.InvalidIntegerOption(ecma402.IntegerOption{
		Name:  "fractionalDigits",
		Value: cfg.fractionalDigits,
		Min:   0,
		Max:   9,
		Set:   cfg.hasFractionalDigits,
	}); ok {
		return invalidOption(check.Name, fmt.Sprint(check.Value), loc)
	}
	return nil
}

func invalidOption(name, value string, loc locale.Locale) error {
	return ecma402.InvalidOptionError("durationformat", name, value, loc.String(), intlerr.ErrInvalidOption)
}

func resolveUnitOptions(cfg config, loc locale.Locale) ([unitCount]resolvedUnitConfig, error) {
	var out [unitCount]resolvedUnitConfig
	prevStyle := UnitStyle("")
	for _, spec := range durationUnitSpecs {
		unit := getUnitConfig(cfg, spec.index)
		resolved, err := getDurationUnitOptions(spec, unit, Style(cfg.style), prevStyle, loc)
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

func getUnitConfig(cfg config, index unitIndex) unitConfig {
	switch index {
	case yearsIndex:
		return cfg.years
	case monthsIndex:
		return cfg.months
	case weeksIndex:
		return cfg.weeks
	case daysIndex:
		return cfg.days
	case hoursIndex:
		return cfg.hours
	case minutesIndex:
		return cfg.minutes
	case secondsIndex:
		return cfg.seconds
	case millisecondsIndex:
		return cfg.milliseconds
	case microsecondsIndex:
		return cfg.microseconds
	case nanosecondsIndex:
		return cfg.nanoseconds
	case unitCount:
	}
	return unitConfig{}
}

func getDurationUnitOptions(spec durationUnitSpec, unit unitConfig, baseStyle Style, prevStyle UnitStyle, loc locale.Locale) (resolvedUnitConfig, error) {
	style := UnitStyle(unit.style)
	displayDefault := AlwaysDisplay
	if style == "" {
		switch {
		case baseStyle == DigitalStyle:
			style = spec.digitalDefault
			if spec.unit != "hours" && spec.unit != "minutes" && spec.unit != "seconds" {
				displayDefault = AutoDisplay
			}
		case prevStyle == fractionalUnitStyle || prevStyle == NumericUnitStyle || prevStyle == TwoDigitUnitStyle:
			style = NumericUnitStyle
			if spec.unit != "minutes" && spec.unit != "seconds" {
				displayDefault = AutoDisplay
			}
		default:
			style = UnitStyle(baseStyle)
			displayDefault = AutoDisplay
		}
	}
	if !slices.Contains(spec.styles, style) {
		return resolvedUnitConfig{}, invalidOption(spec.unit, string(style), loc)
	}
	if style == NumericUnitStyle && spec.fractional {
		style = fractionalUnitStyle
		displayDefault = AutoDisplay
	}
	display := Display(unit.display)
	if display == "" {
		display = displayDefault
	}
	if display != AutoDisplay && display != AlwaysDisplay {
		return resolvedUnitConfig{}, invalidOption(spec.unit+"Display", string(display), loc)
	}
	if err := validateDurationUnitStyle(spec.unit, style, display, prevStyle, loc); err != nil {
		return resolvedUnitConfig{}, err
	}
	if (spec.unit == "minutes" || spec.unit == "seconds") && (prevStyle == NumericUnitStyle || prevStyle == TwoDigitUnitStyle) {
		style = TwoDigitUnitStyle
	}
	return resolvedUnitConfig{style: style, display: display}, nil
}

func validateDurationUnitStyle(unit string, style UnitStyle, display Display, prevStyle UnitStyle, loc locale.Locale) error {
	if display == AlwaysDisplay && style == fractionalUnitStyle {
		return invalidOption(unit+"Display", string(display), loc)
	}
	if prevStyle == fractionalUnitStyle && style != fractionalUnitStyle {
		return invalidOption(unit, string(style), loc)
	}
	if (prevStyle == NumericUnitStyle || prevStyle == TwoDigitUnitStyle) &&
		style != fractionalUnitStyle && style != NumericUnitStyle && style != TwoDigitUnitStyle {
		return invalidOption(unit, string(style), loc)
	}
	return nil
}
