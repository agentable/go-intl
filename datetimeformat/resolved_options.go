package datetimeformat

import "github.com/agentable/go-intl/locale"

// LocaleMatcher selects locale negotiation behavior. Mirrors Intl.DateTimeFormat option "localeMatcher".
type LocaleMatcher string

// HourCycle selects the resolved hour cycle. Mirrors Intl.DateTimeFormat option "hourCycle".
type HourCycle string

// FormatMatcher selects date-time component matching. Mirrors Intl.DateTimeFormat option "formatMatcher".
type FormatMatcher string

// FieldStyle selects textual field width. Mirrors Intl.DateTimeFormat options "weekday", "era", and "dayPeriod".
type FieldStyle string

// NumericStyle selects numeric or two-digit field width. Mirrors Intl.DateTimeFormat numeric field options.
type NumericStyle string

// MonthStyle selects month field width. Mirrors Intl.DateTimeFormat option "month".
type MonthStyle string

// TimeZoneName selects time-zone name width. Mirrors Intl.DateTimeFormat option "timeZoneName".
type TimeZoneName string

// Style selects date or time style width. Mirrors Intl.DateTimeFormat options "dateStyle" and "timeStyle".
type Style string

// PartType identifies a format part record. Mirrors Intl.DateTimeFormat part field "type".
type PartType string

// RangeSource identifies a range part source. Mirrors Intl.DateTimeFormat range part field "source".
type RangeSource string

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	BasicFormatMatcher   FormatMatcher = "basic"
	BestFitFormatMatcher FormatMatcher = "best fit"

	H11HourCycle HourCycle = "h11"
	H12HourCycle HourCycle = "h12"
	H23HourCycle HourCycle = "h23"
	H24HourCycle HourCycle = "h24"

	NarrowFieldStyle FieldStyle = "narrow"
	ShortFieldStyle  FieldStyle = "short"
	LongFieldStyle   FieldStyle = "long"

	NumericFieldStyle  NumericStyle = "numeric"
	TwoDigitFieldStyle NumericStyle = "2-digit"

	NumericMonthStyle  MonthStyle = "numeric"
	TwoDigitMonthStyle MonthStyle = "2-digit"
	NarrowMonthStyle   MonthStyle = "narrow"
	ShortMonthStyle    MonthStyle = "short"
	LongMonthStyle     MonthStyle = "long"

	ShortTimeZoneName        TimeZoneName = "short"
	LongTimeZoneName         TimeZoneName = "long"
	ShortOffsetTimeZoneName  TimeZoneName = "shortOffset"
	LongOffsetTimeZoneName   TimeZoneName = "longOffset"
	ShortGenericTimeZoneName TimeZoneName = "shortGeneric"
	LongGenericTimeZoneName  TimeZoneName = "longGeneric"

	FullDateTimeStyle   Style = "full"
	LongDateTimeStyle   Style = "long"
	MediumDateTimeStyle Style = "medium"
	ShortDateTimeStyle  Style = "short"
)

const (
	// ECMA-402 §15.5.1 Table 9 (Type Field of FormatDateTimePattern part records).
	PartYear                   PartType = "year"
	PartMonth                  PartType = "month"
	PartDay                    PartType = "day"
	PartHour                   PartType = "hour"
	PartMinute                 PartType = "minute"
	PartSecond                 PartType = "second"
	PartWeekday                PartType = "weekday"
	PartEra                    PartType = "era"
	PartDayPeriod              PartType = "dayPeriod"
	PartTimeZoneName           PartType = "timeZoneName"
	PartLiteral                PartType = "literal"
	PartFractionalSecondDigits PartType = "fractionalSecondDigits"
	// PartRelatedYear and PartYearName appear when the resolved calendar is a
	// non-Gregorian cyclic calendar (e.g. Chinese, Dangi); the active Gregorian
	// path never emits them, but the constants must exist so consumer switches
	// remain exhaustive when calendar support widens.
	PartRelatedYear PartType = "relatedYear"
	PartYearName    PartType = "yearName"
	// PartUnknown is ECMA-402's explicit fallback for an unrecognized format
	// pattern element.
	PartUnknown PartType = "unknown"
)

const (
	SourceStartRange RangeSource = "startRange"
	SourceShared     RangeSource = "shared"
	SourceEndRange   RangeSource = "endRange"
)

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
}

type RangePart struct {
	Type   PartType    `json:"type"`
	Value  string      `json:"value"`
	Source RangeSource `json:"source"`
}

type ResolvedOptions struct {
	// Locale is the resolved locale. Mirrors Intl.DateTimeFormat resolved option "locale".
	Locale locale.Locale `json:"locale"`
	// Calendar is the resolved calendar. Mirrors Intl.DateTimeFormat resolved option "calendar".
	Calendar string `json:"calendar"`
	// NumberingSystem is the resolved numbering system. Mirrors Intl.DateTimeFormat resolved option "numberingSystem".
	NumberingSystem string `json:"numberingSystem"`
	// TimeZone is the resolved time zone. Mirrors Intl.DateTimeFormat resolved option "timeZone".
	TimeZone string `json:"timeZone"`
	// HourCycle is the resolved hour cycle. Mirrors Intl.DateTimeFormat resolved option "hourCycle".
	// Empty when ECMA-402 omits hour-cycle properties because no hour field is present.
	HourCycle HourCycle `json:"hourCycle,omitempty"`
	// Hour12 is the resolved 12-hour preference. Mirrors Intl.DateTimeFormat resolved option "hour12".
	// Nil when ECMA-402 omits hour-cycle properties because no hour field is present.
	Hour12 *bool `json:"hour12,omitempty"`
	// Weekday is the resolved weekday style. Mirrors Intl.DateTimeFormat resolved option "weekday".
	// Empty when the weekday component is absent.
	Weekday FieldStyle `json:"weekday,omitempty"`
	// Era is the resolved era style. Mirrors Intl.DateTimeFormat resolved option "era".
	// Empty when the era component is absent.
	Era FieldStyle `json:"era,omitempty"`
	// Year is the resolved year style. Mirrors Intl.DateTimeFormat resolved option "year".
	// Empty when the year component is absent.
	Year NumericStyle `json:"year,omitempty"`
	// Month is the resolved month style. Mirrors Intl.DateTimeFormat resolved option "month".
	// Empty when the month component is absent.
	Month MonthStyle `json:"month,omitempty"`
	// Day is the resolved day style. Mirrors Intl.DateTimeFormat resolved option "day".
	// Empty when the day component is absent.
	Day NumericStyle `json:"day,omitempty"`
	// DayPeriod is the resolved day-period style. Mirrors Intl.DateTimeFormat resolved option "dayPeriod".
	// Empty when the day-period component is absent.
	DayPeriod FieldStyle `json:"dayPeriod,omitempty"`
	// Hour is the resolved hour style. Mirrors Intl.DateTimeFormat resolved option "hour".
	// Empty when the hour component is absent.
	Hour NumericStyle `json:"hour,omitempty"`
	// Minute is the resolved minute style. Mirrors Intl.DateTimeFormat resolved option "minute".
	// Empty when the minute component is absent.
	Minute NumericStyle `json:"minute,omitempty"`
	// Second is the resolved second style. Mirrors Intl.DateTimeFormat resolved option "second".
	// Empty when the second component is absent.
	Second NumericStyle `json:"second,omitempty"`
	// FractionalSecondDigits is the resolved fractional second digit count. Mirrors Intl.DateTimeFormat resolved option "fractionalSecondDigits".
	// Zero when the fractional-second component is absent.
	FractionalSecondDigits int `json:"fractionalSecondDigits,omitempty"`
	// TimeZoneName is the resolved time-zone name style. Mirrors Intl.DateTimeFormat resolved option "timeZoneName".
	// Empty when the time-zone name component is absent.
	TimeZoneName TimeZoneName `json:"timeZoneName,omitempty"`
	// DateStyle is the resolved date style. Mirrors Intl.DateTimeFormat resolved option "dateStyle".
	// Empty when dateStyle was not used.
	DateStyle Style `json:"dateStyle,omitempty"`
	// TimeStyle is the resolved time style. Mirrors Intl.DateTimeFormat resolved option "timeStyle".
	// Empty when timeStyle was not used.
	TimeStyle Style `json:"timeStyle,omitempty"`
}

func (f *DateTimeFormat) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
