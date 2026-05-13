package datetimeformat

import (
	"fmt"
	"slices"

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
	Hour12                 Hour12Option
	DateStyle              Style
	TimeStyle              Style
	FractionalSecondDigits FractionalSecondDigitOptions
}

type Hour12Option struct {
	value bool
	set   bool
}

func Hour12(v bool) Hour12Option {
	return Hour12Option{value: v, set: true}
}

type FractionalSecondDigitOptions struct {
	digits int
	set    bool
}

func FractionalSecondDigits(digits int) FractionalSecondDigitOptions {
	return FractionalSecondDigitOptions{digits: digits, set: true}
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
	if opts.Hour12.set {
		cfg.hour12 = opts.Hour12.value
		cfg.hasHour12 = true
	}
	if opts.DateStyle != "" {
		cfg.dateStyle = string(opts.DateStyle)
	}
	if opts.TimeStyle != "" {
		cfg.timeStyle = string(opts.TimeStyle)
	}
	if opts.FractionalSecondDigits.set {
		cfg.fractionalSecondDigits = opts.FractionalSecondDigits.digits
		cfg.hasFractionalSecondDigits = true
	}
}

func (c config) validate(loc locale.Locale) error {
	checks := []struct {
		name   string
		value  string
		values []string
	}{
		{name: "localeMatcher", value: c.localeMatcher, values: []string{"lookup", "best fit"}},
		{name: "formatMatcher", value: c.formatMatcher, values: []string{"basic", "best fit"}},
		{name: "timeZoneName", value: c.timeZoneName, values: []string{"short", "long", "shortOffset", "longOffset", "shortGeneric", "longGeneric"}},
		{name: "weekday", value: c.weekday, values: []string{"narrow", "short", "long"}},
		{name: "era", value: c.era, values: []string{"narrow", "short", "long"}},
		{name: "year", value: c.year, values: []string{"numeric", "2-digit"}},
		{name: "month", value: c.month, values: []string{"numeric", "2-digit", "narrow", "short", "long"}},
		{name: "day", value: c.day, values: []string{"numeric", "2-digit"}},
		{name: "dayPeriod", value: c.dayPeriod, values: []string{"narrow", "short", "long"}},
		{name: "hour", value: c.hour, values: []string{"numeric", "2-digit"}},
		{name: "minute", value: c.minute, values: []string{"numeric", "2-digit"}},
		{name: "second", value: c.second, values: []string{"numeric", "2-digit"}},
		{name: "hourCycle", value: c.hourCycle, values: []string{"h11", "h12", "h23", "h24"}},
		{name: "dateStyle", value: c.dateStyle, values: []string{"full", "long", "medium", "short"}},
		{name: "timeStyle", value: c.timeStyle, values: []string{"full", "long", "medium", "short"}},
	}
	for _, check := range checks {
		if check.value != "" && !slices.Contains(check.values, check.value) {
			return invalidOption(check.name, check.value, loc)
		}
	}
	if c.calendar != "" && !ecma402.IsWellFormedUnicodeType(c.calendar) {
		return invalidOption("calendar", c.calendar, loc)
	}
	if c.numberingSystem != "" && !ecma402.IsWellFormedUnicodeType(c.numberingSystem) {
		return invalidOption("numberingSystem", c.numberingSystem, loc)
	}
	if c.hasFractionalSecondDigits && (c.fractionalSecondDigits < 1 || c.fractionalSecondDigits > 3) {
		return invalidOption("fractionalSecondDigits", fmt.Sprint(c.fractionalSecondDigits), loc)
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
				return fmt.Errorf("datetimeformat: invalid dateStyle/timeStyle with %s for locale %q: %w", field.name, loc.String(), ErrInvalidOption)
			}
		}
		if c.hasFractionalSecondDigits {
			return fmt.Errorf("datetimeformat: invalid dateStyle/timeStyle with fractionalSecondDigits for locale %q: %w", loc.String(), ErrInvalidOption)
		}
	}
	return nil
}

func invalidOption(name, value string, loc locale.Locale) error {
	return fmt.Errorf("datetimeformat: invalid %s %q for locale %q: %w", name, value, loc.String(), ErrInvalidOption)
}
