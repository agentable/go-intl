package datetimeformat

import (
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/cldr"
	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/internal/tz"
	"github.com/agentable/go-intl/locale"
)

type DateTimeFormat struct {
	resolved      ResolvedOptions
	cldrLoc       cldrdate.Locale
	gregorian     cldrdate.Gregorian
	patterns      *patternData
	location      *time.Location
	formatMatcher FormatMatcher
	pattern       selectedPattern
}

var dateLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrdate.SupportedLocales(), cldrlocale.Maximize)
})

func New(locales locale.List, opts Options) (*DateTimeFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}
	if err := validateRequestedCalendars(locales); err != nil {
		return nil, err
	}
	cfg = withDefaultDateFields(cfg)

	resolution := resolveLocale(locales, validationLocale, cfg)
	calendar := resolution.calendar
	timeZone, location, err := resolveTimeZone(validationLocale, cfg)
	if err != nil {
		return nil, err
	}
	hourCycle, hour12 := resolveHourCycle(cfg, resolution.hourCycle)
	cldrLoc, gregorian := resolveDateData(resolution.cldrLoc, cfg)
	patterns := patternDataFor(cldrLoc, gregorian)
	numberingSystem := resolution.numberingSystem

	format := &DateTimeFormat{cldrLoc: cldrLoc, gregorian: gregorian, patterns: patterns, formatMatcher: FormatMatcher(cfg.formatMatcher), resolved: ResolvedOptions{
		Locale:                 resolution.locale,
		Calendar:               calendar,
		NumberingSystem:        numberingSystem,
		TimeZone:               timeZone,
		HourCycle:              hourCycle,
		Hour12:                 hour12,
		Weekday:                FieldStyle(cfg.weekday),
		Era:                    FieldStyle(cfg.era),
		Year:                   NumericStyle(cfg.year),
		Month:                  MonthStyle(cfg.month),
		Day:                    NumericStyle(cfg.day),
		DayPeriod:              FieldStyle(cfg.dayPeriod),
		Hour:                   NumericStyle(cfg.hour),
		Minute:                 NumericStyle(cfg.minute),
		Second:                 NumericStyle(cfg.second),
		FractionalSecondDigits: cfg.fractionalSecondDigits,
		TimeZoneName:           TimeZoneName(cfg.timeZoneName),
		DateStyle:              Style(cfg.dateStyle),
		TimeStyle:              Style(cfg.timeStyle),
	}, location: location}
	format.pattern = format.selectPattern()
	return format, nil
}

type localeResolution struct {
	locale          locale.Locale
	cldrLoc         cldrdate.Locale
	calendar        string
	numberingSystem string
	hourCycle       string
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) localeResolution {
	defaultLocale := ecma402.DefaultLocale()
	matcher, _ := ecma402.LocaleMatcherAlgorithm(cfg.localeMatcher)
	options := []localematcher.Option{
		{Key: "ca", Value: cfg.calendar},
		{Key: "hc", Value: cfg.hourCycle},
		{Key: "nu", Value: cfg.numberingSystem},
	}
	if cfg.hasHour12 {
		if cfg.hour12 {
			options[1].Value = string(H12HourCycle)
		} else {
			options[1].Value = string(H23HourCycle)
		}
	}
	result := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:             matcher,
		Matcher:               dateLocaleMatcher(),
		Requested:             ecma402.RequestedLocaleStrings(locales),
		DefaultLocale:         defaultLocale,
		RelevantExtensionKeys: []string{"ca", "hc", "nu"},
		OptionValues:          options,
		LocaleData:            dateLocaleData{},
	})
	cldrLoc, ok := cldrdate.ResolveLocale(result.DataLocale)
	if !ok {
		cldrLoc, _ = cldrdate.ResolveLocale(defaultLocale)
	}
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = fallback
	}
	return localeResolution{
		locale:          resolvedLocale,
		cldrLoc:         cldrLoc,
		calendar:        defaultString(result.Extensions["ca"], "gregory"),
		numberingSystem: defaultString(result.Extensions["nu"], cldrLoc.DefaultNumberingSystem()),
		hourCycle:       result.Extensions["hc"],
	}
}

func resolveTimeZone(loc locale.Locale, cfg config) (string, *time.Location, error) {
	if cfg.timeZone == "" {
		timeZone, location := tz.Default()
		return timeZone, location, nil
	}
	location, err := tz.Resolve(cfg.timeZone)
	if err != nil {
		return "", nil, unsupportedTimeZone(cfg.timeZone, loc)
	}
	timeZone := cfg.timeZone
	if strings.HasPrefix(timeZone, "+") || strings.HasPrefix(timeZone, "-") {
		timeZone, err = tz.CanonicalOffsetString(timeZone)
		if err != nil {
			return "", nil, unsupportedTimeZone(cfg.timeZone, loc)
		}
		return timeZone, location, nil
	}
	return tz.CanonicalLink(timeZone), location, nil
}

func unsupportedTimeZone(value string, loc locale.Locale) error {
	return ecma402.UnsupportedOptionError("datetimeformat", "timeZone", value, loc.String(), intlerr.ErrUnsupportedOption)
}

func resolveHourCycle(cfg config, resolvedHourCycle string) (HourCycle, *bool) {
	if cfg.hour == "" && cfg.timeStyle == "" {
		return "", nil
	}
	hourCycle := HourCycle(defaultString(resolvedHourCycle, string(H23HourCycle)))
	hour12 := cfg.hour12
	if !cfg.hasHour12 {
		return hourCycle, nil
	}
	return hourCycle, &hour12
}

func resolveDateData(cldrLoc cldrdate.Locale, cfg config) (cldrdate.Locale, cldrdate.Gregorian) {
	if !needsDateData(cfg) {
		return cldrdate.Undefined, cldrdate.Gregorian{}
	}
	return cldrLoc, gregorianDataFor(cldrLoc)
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func validateRequestedCalendars(locales locale.List) error {
	for _, loc := range locale.CanonicalizeList(locales) {
		if calendar := loc.Calendar(); calendar != "" && !isSupportedCalendar(calendar) {
			return unsupportedCalendar(calendar, loc)
		}
	}
	return nil
}

func isSupportedCalendar(calendar string) bool {
	return slices.Contains(cldr.SupportedCalendars(), calendar)
}

func (f *DateTimeFormat) Format(t time.Time) string {
	t = t.Round(0)
	if f.location != nil {
		t = t.In(f.location)
	}
	return string(f.appendDateTime(nil, t))
}

func withDefaultDateFields(c config) config {
	if c.dateStyle != "" || c.timeStyle != "" || c.weekday != "" || c.era != "" || c.year != "" || c.month != "" || c.day != "" || c.dayPeriod != "" || c.hour != "" || c.minute != "" || c.second != "" || c.hasFractionalSecondDigits {
		return c
	}
	c.year = string(NumericFieldStyle)
	c.month = string(NumericMonthStyle)
	c.day = string(NumericFieldStyle)
	return c
}

func needsDateData(c config) bool {
	if c.dayPeriod != "" || c.timeZoneName != "" || c.dateStyle != "" || c.timeStyle != "" || c.weekday != "" {
		return true
	}
	if c.hour != "" || c.minute != "" || c.second != "" || c.hasFractionalSecondDigits {
		return true
	}
	if c.era != "" || c.year != "" || c.month != "" || c.day != "" {
		return true
	}
	return false
}

func twoDigit(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
