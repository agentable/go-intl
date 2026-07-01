package datetimeformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
)

const (
	dateTimeFormatOwner           = "datetimeformat"
	dateTimeStyleConflictExpected = "no explicit component options when dateStyle or timeStyle is set"
)

var dateTimeStyleValues = [...]string{
	string(FullDateTimeStyle),
	string(LongDateTimeStyle),
	string(MediumDateTimeStyle),
	string(ShortDateTimeStyle),
}

type Options struct {
	Calendar               *string
	NumberingSystem        *string
	LocaleMatcher          *string
	FormatMatcher          *string
	TimeZone               *string
	TimeZoneName           *string
	Weekday                *string
	Era                    *string
	Year                   *string
	Month                  *string
	Day                    *string
	DayPeriod              *string
	Hour                   *string
	Minute                 *string
	Second                 *string
	HourCycle              *string
	Hour12                 *bool
	DateStyle              *string
	TimeStyle              *string
	FractionalSecondDigits *int
}

type config struct {
	calendar                  string
	calendarSet               bool
	numberingSystem           string
	numberingSystemSet        bool
	localeMatcher             string
	localeMatcherSet          bool
	formatMatcher             string
	formatMatcherSet          bool
	timeZone                  string
	timeZoneSet               bool
	timeZoneName              string
	timeZoneNameSet           bool
	weekday                   string
	weekdaySet                bool
	era                       string
	eraSet                    bool
	year                      string
	yearSet                   bool
	month                     string
	monthSet                  bool
	day                       string
	daySet                    bool
	dayPeriod                 string
	dayPeriodSet              bool
	hour                      string
	hourSet                   bool
	minute                    string
	minuteSet                 bool
	second                    string
	secondSet                 bool
	hourCycle                 string
	hourCycleSet              bool
	hour12                    bool
	hasHour12                 bool
	dateStyle                 string
	dateStyleSet              bool
	timeStyle                 string
	timeStyleSet              bool
	fractionalSecondDigits    int
	hasFractionalSecondDigits bool
}

func defaultConfig() config {
	return config{localeMatcher: string(BestFitLocaleMatcher), formatMatcher: string(BestFitFormatMatcher)}
}

func applyOptions(cfg *config, opts Options) {
	ecma402.ApplyOptionInput(&cfg.calendar, &cfg.calendarSet, opts.Calendar)
	ecma402.ApplyOptionInput(&cfg.numberingSystem, &cfg.numberingSystemSet, opts.NumberingSystem)
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.localeMatcherSet, opts.LocaleMatcher)
	ecma402.ApplyOptionInput(&cfg.formatMatcher, &cfg.formatMatcherSet, opts.FormatMatcher)
	ecma402.ApplyOptionInput(&cfg.timeZone, &cfg.timeZoneSet, opts.TimeZone)
	ecma402.ApplyOptionInput(&cfg.timeZoneName, &cfg.timeZoneNameSet, opts.TimeZoneName)
	ecma402.ApplyOptionInput(&cfg.weekday, &cfg.weekdaySet, opts.Weekday)
	ecma402.ApplyOptionInput(&cfg.era, &cfg.eraSet, opts.Era)
	ecma402.ApplyOptionInput(&cfg.year, &cfg.yearSet, opts.Year)
	ecma402.ApplyOptionInput(&cfg.month, &cfg.monthSet, opts.Month)
	ecma402.ApplyOptionInput(&cfg.day, &cfg.daySet, opts.Day)
	ecma402.ApplyOptionInput(&cfg.dayPeriod, &cfg.dayPeriodSet, opts.DayPeriod)
	ecma402.ApplyOptionInput(&cfg.hour, &cfg.hourSet, opts.Hour)
	ecma402.ApplyOptionInput(&cfg.minute, &cfg.minuteSet, opts.Minute)
	ecma402.ApplyOptionInput(&cfg.second, &cfg.secondSet, opts.Second)
	ecma402.ApplyOptionInput(&cfg.hourCycle, &cfg.hourCycleSet, opts.HourCycle)
	ecma402.ApplyOptionInput(&cfg.hour12, &cfg.hasHour12, opts.Hour12)
	ecma402.ApplyOptionInput(&cfg.dateStyle, &cfg.dateStyleSet, opts.DateStyle)
	ecma402.ApplyOptionInput(&cfg.timeStyle, &cfg.timeStyleSet, opts.TimeStyle)
	ecma402.ApplyOptionInput(&cfg.fractionalSecondDigits, &cfg.hasFractionalSecondDigits, opts.FractionalSecondDigits)
}

func (c config) validate(locName string) error {
	if err := ecma402.ValidateStringOptions(
		dateTimeFormatOwner,
		locName,
		dateTimeFormatMatcherOptionInput(c.formatMatcher, c.formatMatcherSet),
		dateTimeZoneNameOptionInput(c.timeZoneName, c.timeZoneNameSet),
		dateTimeFieldStyleOptionInput("weekday", c.weekday, c.weekdaySet),
		dateTimeFieldStyleOptionInput("era", c.era, c.eraSet),
		dateTimeNumericStyleOptionInput("year", c.year, c.yearSet),
		dateTimeMonthStyleOptionInput(c.month, c.monthSet),
		dateTimeNumericStyleOptionInput("day", c.day, c.daySet),
		dateTimeFieldStyleOptionInput("dayPeriod", c.dayPeriod, c.dayPeriodSet),
		dateTimeNumericStyleOptionInput("hour", c.hour, c.hourSet),
		dateTimeNumericStyleOptionInput("minute", c.minute, c.minuteSet),
		dateTimeNumericStyleOptionInput("second", c.second, c.secondSet),
		dateTimeHourCycleOptionInput(c.hourCycle, c.hourCycleSet),
		dateTimeStyleOptionInput("dateStyle", c.dateStyle, c.dateStyleSet),
		dateTimeStyleOptionInput("timeStyle", c.timeStyle, c.timeStyleSet),
		ecma402.LocaleMatcherOptionInput(c.localeMatcher, c.localeMatcherSet),
	); err != nil {
		return err
	}
	if err := ecma402.ValidateUnicodeTypeOptionInput(dateTimeFormatOwner, "calendar", c.calendar, locName, c.calendarSet); err != nil {
		return err
	}
	if err := ecma402.ValidateUnicodeTypeOptionInput(dateTimeFormatOwner, "numberingSystem", c.numberingSystem, locName, c.numberingSystemSet); err != nil {
		return err
	}
	if err := ecma402.ValidateIntegerOptions(dateTimeFormatOwner, locName, ecma402.IntegerOption{
		Name:  "fractionalSecondDigits",
		Value: c.fractionalSecondDigits,
		Min:   1,
		Max:   3,
		Set:   c.hasFractionalSecondDigits,
	}); err != nil {
		return err
	}
	return c.validateDateTimeStyleConflicts(locName)
}

func (c config) validateDateTimeStyleConflicts(loc string) error {
	if c.dateStyle == "" && c.timeStyle == "" {
		return nil
	}
	if field, ok := c.firstDateTimeStyleConflict(); ok {
		return invalidDateTimeStyleConflict(field, loc)
	}
	return nil
}

func (c config) firstDateTimeStyleConflict() (string, bool) {
	switch {
	case c.weekday != "":
		return "weekday", true
	case c.era != "":
		return "era", true
	case c.year != "":
		return "year", true
	case c.month != "":
		return "month", true
	case c.day != "":
		return "day", true
	case c.dayPeriod != "":
		return "dayPeriod", true
	case c.hour != "":
		return "hour", true
	case c.minute != "":
		return "minute", true
	case c.second != "":
		return "second", true
	case c.timeZoneName != "":
		return "timeZoneName", true
	case c.hasFractionalSecondDigits:
		return "fractionalSecondDigits", true
	default:
		return "", false
	}
}

func dateTimeFormatMatcherOptionInput(value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput("formatMatcher", value, present,
		string(BasicFormatMatcher),
		string(BestFitFormatMatcher),
	)
}

func dateTimeZoneNameOptionInput(value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput("timeZoneName", value, present,
		string(ShortTimeZoneName),
		string(LongTimeZoneName),
		string(ShortOffsetTimeZoneName),
		string(LongOffsetTimeZoneName),
		string(ShortGenericTimeZoneName),
		string(LongGenericTimeZoneName),
	)
}

func dateTimeFieldStyleOptionInput(name, value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput(name, value, present,
		string(NarrowFieldStyle),
		string(ShortFieldStyle),
		string(LongFieldStyle),
	)
}

func dateTimeNumericStyleOptionInput(name, value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput(name, value, present,
		string(NumericFieldStyle),
		string(TwoDigitFieldStyle),
	)
}

func dateTimeMonthStyleOptionInput(value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput("month", value, present,
		string(NumericMonthStyle),
		string(TwoDigitMonthStyle),
		string(NarrowMonthStyle),
		string(ShortMonthStyle),
		string(LongMonthStyle),
	)
}

func dateTimeHourCycleOptionInput(value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput("hourCycle", value, present,
		string(H11HourCycle),
		string(H12HourCycle),
		string(H23HourCycle),
		string(H24HourCycle),
	)
}

func dateTimeStyleOptionInput(name, value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput(name, value, present, dateTimeStyleValues[:]...)
}

func invalidDateTimeStyleConflict(value, loc string) error {
	return ecma402.InvalidOptionErrorExpected(
		dateTimeFormatOwner,
		"dateStyle/timeStyle",
		value,
		loc,
		dateTimeStyleConflictExpected,
		nil,
	)
}
