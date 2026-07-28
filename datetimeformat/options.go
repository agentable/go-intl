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
	hasCalendar               bool
	numberingSystem           string
	hasNumberingSystem        bool
	localeMatcher             string
	hasLocaleMatcher          bool
	formatMatcher             string
	hasFormatMatcher          bool
	timeZone                  string
	hasTimeZone               bool
	timeZoneName              string
	hasTimeZoneName           bool
	weekday                   string
	hasWeekday                bool
	era                       string
	hasEra                    bool
	year                      string
	hasYear                   bool
	month                     string
	hasMonth                  bool
	day                       string
	hasDay                    bool
	dayPeriod                 string
	hasDayPeriod              bool
	hour                      string
	hasHour                   bool
	minute                    string
	hasMinute                 bool
	second                    string
	hasSecond                 bool
	hourCycle                 string
	hasHourCycle              bool
	hour12                    bool
	hasHour12                 bool
	dateStyle                 string
	hasDateStyle              bool
	timeStyle                 string
	hasTimeStyle              bool
	fractionalSecondDigits    int
	hasFractionalSecondDigits bool
}

func defaultConfig() config {
	return config{localeMatcher: string(BestFitLocaleMatcher), formatMatcher: string(BestFitFormatMatcher)}
}

func applyOptions(cfg *config, opts Options) {
	ecma402.ApplyUnicodeTypeOptionInput(&cfg.calendar, &cfg.hasCalendar, "ca", opts.Calendar)
	ecma402.ApplyUnicodeTypeOptionInput(&cfg.numberingSystem, &cfg.hasNumberingSystem, "nu", opts.NumberingSystem)
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.hasLocaleMatcher, opts.LocaleMatcher)
	ecma402.ApplyOptionInput(&cfg.formatMatcher, &cfg.hasFormatMatcher, opts.FormatMatcher)
	ecma402.ApplyOptionInput(&cfg.timeZone, &cfg.hasTimeZone, opts.TimeZone)
	ecma402.ApplyOptionInput(&cfg.timeZoneName, &cfg.hasTimeZoneName, opts.TimeZoneName)
	ecma402.ApplyOptionInput(&cfg.weekday, &cfg.hasWeekday, opts.Weekday)
	ecma402.ApplyOptionInput(&cfg.era, &cfg.hasEra, opts.Era)
	ecma402.ApplyOptionInput(&cfg.year, &cfg.hasYear, opts.Year)
	ecma402.ApplyOptionInput(&cfg.month, &cfg.hasMonth, opts.Month)
	ecma402.ApplyOptionInput(&cfg.day, &cfg.hasDay, opts.Day)
	ecma402.ApplyOptionInput(&cfg.dayPeriod, &cfg.hasDayPeriod, opts.DayPeriod)
	ecma402.ApplyOptionInput(&cfg.hour, &cfg.hasHour, opts.Hour)
	ecma402.ApplyOptionInput(&cfg.minute, &cfg.hasMinute, opts.Minute)
	ecma402.ApplyOptionInput(&cfg.second, &cfg.hasSecond, opts.Second)
	ecma402.ApplyOptionInput(&cfg.hourCycle, &cfg.hasHourCycle, opts.HourCycle)
	ecma402.ApplyOptionInput(&cfg.hour12, &cfg.hasHour12, opts.Hour12)
	ecma402.ApplyOptionInput(&cfg.dateStyle, &cfg.hasDateStyle, opts.DateStyle)
	ecma402.ApplyOptionInput(&cfg.timeStyle, &cfg.hasTimeStyle, opts.TimeStyle)
	ecma402.ApplyOptionInput(&cfg.fractionalSecondDigits, &cfg.hasFractionalSecondDigits, opts.FractionalSecondDigits)
}

func (c config) validate(locName string) error {
	if err := ecma402.ValidateStringOptions(
		dateTimeFormatOwner,
		locName,
		dateTimeFormatMatcherOptionInput(c.formatMatcher, c.hasFormatMatcher),
		dateTimeZoneNameOptionInput(c.timeZoneName, c.hasTimeZoneName),
		dateTimeFieldStyleOptionInput("weekday", c.weekday, c.hasWeekday),
		dateTimeFieldStyleOptionInput("era", c.era, c.hasEra),
		dateTimeNumericStyleOptionInput("year", c.year, c.hasYear),
		dateTimeMonthStyleOptionInput(c.month, c.hasMonth),
		dateTimeNumericStyleOptionInput("day", c.day, c.hasDay),
		dateTimeFieldStyleOptionInput("dayPeriod", c.dayPeriod, c.hasDayPeriod),
		dateTimeNumericStyleOptionInput("hour", c.hour, c.hasHour),
		dateTimeNumericStyleOptionInput("minute", c.minute, c.hasMinute),
		dateTimeNumericStyleOptionInput("second", c.second, c.hasSecond),
		dateTimeHourCycleOptionInput(c.hourCycle, c.hasHourCycle),
		dateTimeStyleOptionInput("dateStyle", c.dateStyle, c.hasDateStyle),
		dateTimeStyleOptionInput("timeStyle", c.timeStyle, c.hasTimeStyle),
		ecma402.LocaleMatcherOptionInput(c.localeMatcher, c.hasLocaleMatcher),
	); err != nil {
		return err
	}
	if err := ecma402.ValidateUnicodeTypeOptionInput(dateTimeFormatOwner, "calendar", c.calendar, locName, c.hasCalendar); err != nil {
		return err
	}
	if err := ecma402.ValidateUnicodeTypeOptionInput(dateTimeFormatOwner, "numberingSystem", c.numberingSystem, locName, c.hasNumberingSystem); err != nil {
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
