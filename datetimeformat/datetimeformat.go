package datetimeformat

import (
	"strconv"
	"sync"
	"time"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/internal/tz"
	"github.com/agentable/go-intl/locale"
)

type DateTimeFormat struct {
	resolved             ResolvedOptions
	cldrLoc              cldrdate.Locale
	gregorian            cldrdate.Gregorian
	location             *time.Location
	uses24Hour           bool
	pattern              selectedPattern
	fallbackRangePattern ecma402.Pattern
}

const timeZoneExpected = "a supported IANA time-zone name or UTC offset identifier"

var dateLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrdate.SupportedLocales(), cldrlocale.Maximize)
})

func New(locales locale.List, opts Options) (*DateTimeFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	validationLocaleName := validationLocale.String()
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocaleName); err != nil {
		return nil, err
	}
	cfg = withDefaultDateFields(cfg)

	resolution := resolveLocale(locales, validationLocale, cfg)
	calendar := resolution.calendar
	timeZone, location, err := resolveTimeZone(validationLocaleName, cfg)
	if err != nil {
		return nil, err
	}
	hourCycle, hour12 := resolveHourCycle(cfg, resolution.hourCycle)
	cldrLoc, gregorian := resolveDateData(resolution.cldrLoc, cfg)
	patterns := patternDataFor(cldrLoc, gregorian)
	numberingSystem := resolution.numberingSystem

	resolved := ResolvedOptions{
		Locale:          resolution.locale,
		Calendar:        calendar,
		NumberingSystem: numberingSystem,
		TimeZone:        timeZone,
		HourCycle:       resolvedStringOption[HourCycle](string(hourCycle)),
		Hour12:          hour12,
		Weekday:         resolvedStringOption[FieldStyle](cfg.weekday),
		Era:             resolvedStringOption[FieldStyle](cfg.era),
		Year:            resolvedStringOption[NumericStyle](cfg.year),
		Month:           resolvedStringOption[MonthStyle](cfg.month),
		Day:             resolvedStringOption[NumericStyle](cfg.day),
		DayPeriod:       resolvedStringOption[FieldStyle](cfg.dayPeriod),
		Hour:            resolvedStringOption[NumericStyle](cfg.hour),
		Minute:          resolvedStringOption[NumericStyle](cfg.minute),
		Second:          resolvedStringOption[NumericStyle](cfg.second),
		TimeZoneName:    resolvedStringOption[TimeZoneName](cfg.timeZoneName),
		DateStyle:       resolvedStringOption[Style](cfg.dateStyle),
		TimeStyle:       resolvedStringOption[Style](cfg.timeStyle),
	}
	if cfg.hasFractionalSecondDigits {
		resolved.FractionalSecondDigits = ecma402.ResolvedScalar(cfg.fractionalSecondDigits)
	}

	uses24Hour := resolvedUses24HourTime(resolved)
	pattern := selectPattern(patterns, FormatMatcher(cfg.formatMatcher), resolved, uses24Hour, gregorian)
	return &DateTimeFormat{
		resolved:   resolved,
		cldrLoc:    cldrLoc,
		gregorian:  gregorian,
		location:   location,
		uses24Hour: uses24Hour,
		pattern:    pattern,
		fallbackRangePattern: partitionRangeFallbackPattern(
			gregorian.IntervalFallback,
		),
	}, nil
}

type localeResolution struct {
	locale          locale.Locale
	cldrLoc         cldrdate.Locale
	calendar        string
	numberingSystem string
	hourCycle       string
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) localeResolution {
	options := []ecma402.UnicodeExtensionOption{
		{Key: ecma402.UnicodeExtensionKeyCalendar, Value: cfg.calendar},
		{Key: ecma402.UnicodeExtensionKeyHourCycle, Value: cfg.hourCycle},
		{Key: ecma402.UnicodeExtensionKeyNumberingSystem, Value: cfg.numberingSystem},
	}
	if cfg.hasHour12 {
		if cfg.hour12 {
			options[1].Value = string(H12HourCycle)
		} else {
			options[1].Value = string(H23HourCycle)
		}
	}
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:               locales,
		Fallback:              fallback,
		LocaleMatcher:         cfg.localeMatcher,
		Matcher:               dateLocaleMatcher(),
		RelevantExtensionKeys: []ecma402.UnicodeExtensionKey{ecma402.UnicodeExtensionKeyCalendar, ecma402.UnicodeExtensionKeyHourCycle, ecma402.UnicodeExtensionKeyNumberingSystem},
		OptionValues:          options,
		LocaleData:            dateLocaleData{},
	})
	cldrLoc := ecma402.ResolveDataLocale(resolution, cldrdate.ResolveLocale)
	calendar := resolution.Extensions[ecma402.UnicodeExtensionKeyCalendar]
	if calendar == "" {
		calendar = "gregory"
	}
	numberingSystem := resolution.Extensions[ecma402.UnicodeExtensionKeyNumberingSystem]
	if numberingSystem == "" {
		numberingSystem = cldrLoc.DefaultNumberingSystem()
	}
	return localeResolution{
		locale:          resolution.Locale,
		cldrLoc:         cldrLoc,
		calendar:        calendar,
		numberingSystem: numberingSystem,
		hourCycle:       resolution.Extensions[ecma402.UnicodeExtensionKeyHourCycle],
	}
}

func resolveTimeZone(locName string, cfg config) (string, *time.Location, error) {
	if !cfg.timeZoneSet {
		timeZone, location := tz.Default()
		return timeZone, location, nil
	}
	if cfg.timeZone == "" {
		return "", nil, unsupportedTimeZone(cfg.timeZone, locName, nil)
	}
	location, err := tz.Resolve(cfg.timeZone)
	if err != nil {
		return "", nil, unsupportedTimeZone(cfg.timeZone, locName, err)
	}
	timeZone := location.String()
	return timeZone, location, nil
}

func unsupportedTimeZone(value, locName string, err error) error {
	return ecma402.UnsupportedOptionErrorExpected(dateTimeFormatOwner, "timeZone", value, locName, timeZoneExpected, err)
}

func resolveHourCycle(cfg config, resolvedHourCycle string) (HourCycle, *bool) {
	if cfg.hour == "" && cfg.timeStyle == "" {
		return "", nil
	}
	if resolvedHourCycle == "" {
		resolvedHourCycle = string(H23HourCycle)
	}
	hourCycle := HourCycle(resolvedHourCycle)
	switch hourCycle {
	case H11HourCycle, H12HourCycle:
		return hourCycle, ecma402.ResolvedScalar(true)
	case H23HourCycle, H24HourCycle:
		return hourCycle, ecma402.ResolvedScalar(false)
	default:
		return hourCycle, nil
	}
}

func resolveDateData(cldrLoc cldrdate.Locale, cfg config) (cldrdate.Locale, cldrdate.Gregorian) {
	if !needsDateData(cfg) {
		return cldrdate.Undefined, cldrdate.Gregorian{}
	}
	return cldrLoc, gregorianDataFor(cldrLoc)
}

func (f *DateTimeFormat) Format(t time.Time) string {
	_, local := gregoryTimeInLocation(t.Round(0), f.location)
	return string(f.pattern.appendTo(f, nil, local))
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
