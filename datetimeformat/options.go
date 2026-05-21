package datetimeformat

import (
	"fmt"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

type Options struct {
	Calendar               string
	NumberingSystem        string
	LocaleMatcher          LocaleMatcher
	FormatMatcher          FormatMatcher
	TimeZone               string
	TimeZoneName           TimeZoneName
	Weekday                FieldStyle
	Era                    FieldStyle
	Year                   NumericStyle
	Month                  MonthStyle
	Day                    NumericStyle
	DayPeriod              FieldStyle
	Hour                   NumericStyle
	Minute                 NumericStyle
	Second                 NumericStyle
	HourCycle              HourCycle
	Hour12                 *bool
	DateStyle              Style
	TimeStyle              Style
	FractionalSecondDigits *int
}

type config struct {
	calendar                  string
	numberingSystem           string
	localeMatcher             string
	formatMatcher             string
	timeZone                  string
	timeZoneName              string
	weekday                   string
	era                       string
	year                      string
	month                     string
	day                       string
	dayPeriod                 string
	hour                      string
	minute                    string
	second                    string
	hourCycle                 string
	hour12                    bool
	hasHour12                 bool
	dateStyle                 string
	timeStyle                 string
	fractionalSecondDigits    int
	hasFractionalSecondDigits bool
}

func defaultConfig() config {
	return config{localeMatcher: string(BestFitLocaleMatcher), formatMatcher: string(BestFitFormatMatcher)}
}

func applyOptions(cfg *config, opts Options) {
	if opts.Calendar != "" {
		cfg.calendar = opts.Calendar
	}
	if opts.NumberingSystem != "" {
		cfg.numberingSystem = opts.NumberingSystem
	}
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	if opts.FormatMatcher != "" {
		cfg.formatMatcher = string(opts.FormatMatcher)
	}
	if opts.TimeZone != "" {
		cfg.timeZone = opts.TimeZone
	}
	if opts.TimeZoneName != "" {
		cfg.timeZoneName = string(opts.TimeZoneName)
	}
	if opts.Weekday != "" {
		cfg.weekday = string(opts.Weekday)
	}
	if opts.Era != "" {
		cfg.era = string(opts.Era)
	}
	if opts.Year != "" {
		cfg.year = string(opts.Year)
	}
	if opts.Month != "" {
		cfg.month = string(opts.Month)
	}
	if opts.Day != "" {
		cfg.day = string(opts.Day)
	}
	if opts.DayPeriod != "" {
		cfg.dayPeriod = string(opts.DayPeriod)
	}
	if opts.Hour != "" {
		cfg.hour = string(opts.Hour)
	}
	if opts.Minute != "" {
		cfg.minute = string(opts.Minute)
	}
	if opts.Second != "" {
		cfg.second = string(opts.Second)
	}
	if opts.HourCycle != "" {
		cfg.hourCycle = string(opts.HourCycle)
	}
	if opts.Hour12 != nil {
		cfg.hour12 = *opts.Hour12
		cfg.hasHour12 = true
	}
	if opts.DateStyle != "" {
		cfg.dateStyle = string(opts.DateStyle)
	}
	if opts.TimeStyle != "" {
		cfg.timeStyle = string(opts.TimeStyle)
	}
	if opts.FractionalSecondDigits != nil {
		cfg.fractionalSecondDigits = *opts.FractionalSecondDigits
		cfg.hasFractionalSecondDigits = true
	}
}

func (c config) validate(loc locale.Locale) error {
	checks := []ecma402.StringOption{
		ecma402.RequiredStringOption("formatMatcher", c.formatMatcher, "basic", "best fit"),
		ecma402.OptionalStringOption("timeZoneName", c.timeZoneName, "short", "long", "shortOffset", "longOffset", "shortGeneric", "longGeneric"),
		ecma402.OptionalStringOption("weekday", c.weekday, "narrow", "short", "long"),
		ecma402.OptionalStringOption("era", c.era, "narrow", "short", "long"),
		ecma402.OptionalStringOption("year", c.year, "numeric", "2-digit"),
		ecma402.OptionalStringOption("month", c.month, "numeric", "2-digit", "narrow", "short", "long"),
		ecma402.OptionalStringOption("day", c.day, "numeric", "2-digit"),
		ecma402.OptionalStringOption("dayPeriod", c.dayPeriod, "narrow", "short", "long"),
		ecma402.OptionalStringOption("hour", c.hour, "numeric", "2-digit"),
		ecma402.OptionalStringOption("minute", c.minute, "numeric", "2-digit"),
		ecma402.OptionalStringOption("second", c.second, "numeric", "2-digit"),
		ecma402.OptionalStringOption("hourCycle", c.hourCycle, "h11", "h12", "h23", "h24"),
		ecma402.OptionalStringOption("dateStyle", c.dateStyle, "full", "long", "medium", "short"),
		ecma402.OptionalStringOption("timeStyle", c.timeStyle, "full", "long", "medium", "short"),
		ecma402.LocaleMatcherOption(c.localeMatcher),
	}
	if check, ok := ecma402.InvalidStringOption(checks...); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	if c.calendar != "" && !ecma402.IsWellFormedUnicodeType(c.calendar) {
		return invalidOption("calendar", c.calendar, loc)
	}
	if c.calendar != "" && !isSupportedCalendar(c.calendar) {
		return unsupportedCalendar(c.calendar, loc)
	}
	if c.numberingSystem != "" && !ecma402.IsWellFormedUnicodeType(c.numberingSystem) {
		return invalidOption("numberingSystem", c.numberingSystem, loc)
	}
	if check, ok := ecma402.InvalidIntegerOption(ecma402.IntegerOption{
		Name:  "fractionalSecondDigits",
		Value: c.fractionalSecondDigits,
		Min:   1,
		Max:   3,
		Set:   c.hasFractionalSecondDigits,
	}); ok {
		return invalidOption(check.Name, fmt.Sprint(check.Value), loc)
	}
	if c.dateStyle != "" || c.timeStyle != "" {
		for _, field := range []struct {
			name  string
			value string
		}{
			{name: "weekday", value: c.weekday},
			{name: "era", value: c.era},
			{name: "year", value: c.year},
			{name: "month", value: c.month},
			{name: "day", value: c.day},
			{name: "dayPeriod", value: c.dayPeriod},
			{name: "hour", value: c.hour},
			{name: "minute", value: c.minute},
			{name: "second", value: c.second},
			{name: "timeZoneName", value: c.timeZoneName},
		} {
			if field.value != "" {
				return invalidOption("dateStyle/timeStyle", field.name, loc)
			}
		}
		if c.hasFractionalSecondDigits {
			return invalidOption("dateStyle/timeStyle", "fractionalSecondDigits", loc)
		}
	}
	return nil
}

func invalidOption(name, value string, loc locale.Locale) error {
	return ecma402.InvalidOptionError("datetimeformat", name, value, loc.String(), intlerr.ErrInvalidOption)
}

func unsupportedCalendar(value string, loc locale.Locale) error {
	return ecma402.UnsupportedOptionError("datetimeformat", "calendar", value, loc.String(), intlerr.ErrUnsupportedOption)
}
