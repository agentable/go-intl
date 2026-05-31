package durationformat

import (
	"sync"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

type DurationFormat struct {
	resolved    ResolvedOptions
	unitOptions [unitCount]resolvedUnitConfig
	separator   string
	formatters  durationFormatters
}

var durationLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(supportedLocales(), cldrlocale.Maximize)
})

func New(locales locale.List, opts Options) (*DurationFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}
	resolvedLocale, cldrLoc, numberingSystem := resolveLocale(locales, validationLocale, cfg)
	unitOptions, err := resolveUnitOptions(cfg, validationLocale)
	if err != nil {
		return nil, err
	}
	resolved := ResolvedOptions{
		Locale:              resolvedLocale,
		NumberingSystem:     numberingSystem,
		Style:               Style(cfg.style),
		Years:               publicUnitStyle(unitOptions[yearsIndex].style),
		YearsDisplay:        unitOptions[yearsIndex].display,
		Months:              publicUnitStyle(unitOptions[monthsIndex].style),
		MonthsDisplay:       unitOptions[monthsIndex].display,
		Weeks:               publicUnitStyle(unitOptions[weeksIndex].style),
		WeeksDisplay:        unitOptions[weeksIndex].display,
		Days:                publicUnitStyle(unitOptions[daysIndex].style),
		DaysDisplay:         unitOptions[daysIndex].display,
		Hours:               publicUnitStyle(unitOptions[hoursIndex].style),
		HoursDisplay:        unitOptions[hoursIndex].display,
		Minutes:             publicUnitStyle(unitOptions[minutesIndex].style),
		MinutesDisplay:      unitOptions[minutesIndex].display,
		Seconds:             publicUnitStyle(unitOptions[secondsIndex].style),
		SecondsDisplay:      unitOptions[secondsIndex].display,
		Milliseconds:        publicUnitStyle(unitOptions[millisecondsIndex].style),
		MillisecondsDisplay: unitOptions[millisecondsIndex].display,
		Microseconds:        publicUnitStyle(unitOptions[microsecondsIndex].style),
		MicrosecondsDisplay: unitOptions[microsecondsIndex].display,
		Nanoseconds:         publicUnitStyle(unitOptions[nanosecondsIndex].style),
		NanosecondsDisplay:  unitOptions[nanosecondsIndex].display,
	}
	if cfg.hasFractionalDigits {
		resolved.FractionalDigits = &cfg.fractionalDigits
	}
	symbols := cldrLoc.NumberSymbols(numberingSystem)
	separator := symbols.TimeSeparator
	if separator == "" {
		separator = ":"
	}
	formatters, err := buildDurationFormatters(resolved, unitOptions)
	if err != nil {
		return nil, err
	}
	return &DurationFormat{resolved: resolved, unitOptions: unitOptions, separator: separator, formatters: formatters}, nil
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, cldrnumber.Locale, string) {
	defaultLocale := ecma402.DefaultLocale()
	matcher, _ := ecma402.LocaleMatcherAlgorithm(cfg.localeMatcher)
	result := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:             matcher,
		Matcher:               durationLocaleMatcher(),
		Requested:             ecma402.RequestedLocaleStrings(locales),
		DefaultLocale:         defaultLocale,
		RelevantExtensionKeys: []string{"nu"},
		OptionValues:          []localematcher.Option{{Key: "nu", Value: cfg.numberingSystem}},
		LocaleData:            cldrnumber.NumberLocaleData{},
	})
	cldrLoc, ok := cldrnumber.ResolveLocale(result.DataLocale)
	if !ok {
		cldrLoc, _ = cldrnumber.ResolveLocale(defaultLocale)
	}
	numberingSystem := result.Extensions["nu"]
	if numberingSystem == "" {
		numberingSystem = cldrLoc.DefaultNumberingSystem()
	}
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = fallback
	}
	return resolvedLocale, cldrLoc, numberingSystem
}

func publicUnitStyle(style UnitStyle) UnitStyle {
	if style == fractionalUnitStyle {
		return NumericUnitStyle
	}
	return style
}
