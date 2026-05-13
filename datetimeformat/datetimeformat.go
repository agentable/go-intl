package datetimeformat

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/internal/tz"
	"github.com/agentable/go-intl/locale"
)

type DateTimeFormat struct {
	resolved      ResolvedOptions
	cldrLoc       cldr.Locale
	gregorian     cldr.Gregorian
	location      *time.Location
	formatMatcher FormatMatcher
	pattern       selectedPattern
}

func New(loc locale.Locale, opts ...Options) (*DateTimeFormat, error) {
	cfg := defaultConfig()
	if len(opts) > 1 {
		return nil, fmt.Errorf("datetimeformat: multiple options for locale %q: %w", loc.Tag().String(), ErrInvalidOption)
	}
	if len(opts) > 0 {
		applyOptions(&cfg, opts[0])
	}
	if err := cfg.validate(loc); err != nil {
		return nil, err
	}
	cfg = withDefaultDateFields(cfg)

	resolution := resolveLocale(loc, cfg)
	calendar := resolution.calendar
	timeZone, location, err := resolveTimeZone(loc, cfg)
	if err != nil {
		return nil, err
	}
	hourCycle, hour12 := resolveHourCycle(cfg, resolution.hourCycle)
	cldrLoc, gregorian := resolveDateData(resolution.cldrLoc, cfg)
	numberingSystem := resolution.numberingSystem

	format := &DateTimeFormat{cldrLoc: cldrLoc, gregorian: gregorian, formatMatcher: FormatMatcher(cfg.formatMatcher), resolved: ResolvedOptions{
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
	cldrLoc         cldr.Locale
	calendar        string
	numberingSystem string
	hourCycle       string
}

func resolveLocale(loc locale.Locale, cfg config) localeResolution {
	options := map[string]string{"ca": cfg.calendar, "nu": cfg.numberingSystem}
	if cfg.hasHour12 {
		if cfg.hour12 {
			options["hc"] = string(H12HourCycle)
		} else {
			options["hc"] = string(H23HourCycle)
		}
	} else {
		options["hc"] = cfg.hourCycle
	}
	result := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:             localeMatcherAlgorithm(cfg.localeMatcher),
		Requested:             []string{loc.String()},
		Supported:             cldr.DateSupportedLocales(),
		DefaultLocale:         "en",
		RelevantExtensionKeys: []string{"ca", "hc", "nu"},
		Options:               options,
		LocaleData:            cldr.DateLocaleData{},
	})
	cldrLoc, ok := cldr.ResolveLocale(language.Make(result.DataLocale))
	if !ok {
		cldrLoc, _ = cldr.ResolveLocale(language.English)
	}
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = loc
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
		return "", nil, fmt.Errorf("datetimeformat: unsupported time zone %q for locale %q: %w", cfg.timeZone, loc.String(), ErrUnsupportedTimeZone)
	}
	timeZone := cfg.timeZone
	if strings.HasPrefix(timeZone, "+") || strings.HasPrefix(timeZone, "-") {
		timeZone, err = tz.CanonicalOffsetString(timeZone)
		if err != nil {
			return "", nil, fmt.Errorf("datetimeformat: unsupported time zone %q for locale %q: %w", cfg.timeZone, loc.String(), ErrUnsupportedTimeZone)
		}
	} else {
		timeZone = tz.CanonicalLink(timeZone)
	}
	return timeZone, location, nil
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

func resolveDateData(cldrLoc cldr.Locale, cfg config) (cldr.Locale, cldr.Gregorian) {
	if !needsDateData(cfg) {
		return cldr.Undefined, cldr.Gregorian{}
	}
	return cldrLoc, cldr.GregorianFor(cldrLoc)
}

func localeMatcherAlgorithm(matcher string) localematcher.Algorithm {
	if matcher == string(LookupLocaleMatcher) {
		return localematcher.AlgorithmLookup
	}
	return localematcher.AlgorithmBestFit
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func (f *DateTimeFormat) Format(t time.Time) string {
	return joinParts(f.FormatToParts(t))
}

func joinParts(parts []Part) string {
	size := 0
	for _, part := range parts {
		size += len(part.Value)
	}
	var b strings.Builder
	b.Grow(size)
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
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
	switch c.month {
	case string(ShortMonthStyle), string(LongMonthStyle), string(NarrowMonthStyle):
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
