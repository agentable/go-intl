package datetimeformat

import "github.com/agentable/go-intl/locale"

type LocaleMatcher string
type HourCycle string
type FormatMatcher string
type FieldStyle string
type NumericStyle string
type MonthStyle string
type TimeZoneName string
type Style string
type PartType string
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
)

const (
	SourceStartRange RangeSource = "startRange"
	SourceShared     RangeSource = "shared"
	SourceEndRange   RangeSource = "endRange"
)

type Part struct {
	Type  PartType
	Value string
}

type RangePart struct {
	Type   PartType
	Value  string
	Source RangeSource
}

type ResolvedOptions struct {
	Locale                 locale.Locale
	Calendar               string
	NumberingSystem        string
	TimeZone               string
	HourCycle              HourCycle
	Hour12                 *bool
	Weekday                FieldStyle
	Era                    FieldStyle
	Year                   NumericStyle
	Month                  MonthStyle
	Day                    NumericStyle
	DayPeriod              FieldStyle
	Hour                   NumericStyle
	Minute                 NumericStyle
	Second                 NumericStyle
	FractionalSecondDigits int
	TimeZoneName           TimeZoneName
	DateStyle              Style
	TimeStyle              Style
}

func (f *DateTimeFormat) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
