// Hand-written decode layer for the date domain. It expands domain-private
// const blobs from data.go into the calendar, day-period, and supported-index
// records consumed by accessors.go, behind per-blob sync.Once gates.
//
// Locale handle ownership: the date blobs pack the locale index assigned by the
// cldr/locale kernel. Borrowing that handle keeps generated date data and
// formatter locale resolution on one stable index space while the dependency
// stays one-way (date -> cldr/locale).

package date

import (
	"sync"
	"time"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

// Undefined is the "und" sentinel locale, re-exported for callers that compare a
// resolved handle against it.
const Undefined = cldrlocale.Undefined

// dayPeriodNames carries the wide/abbreviated/narrow names of one flexible day
// period.
type dayPeriodNames struct{ Wide, Abbr, Narrow string }

// DayPeriodRange is one resolved day-period rule: a half-open [From, To)
// duration window (or a point window when From == To) tagged with its period
// type.
type DayPeriodRange struct {
	From time.Duration
	To   time.Duration
	Type string
}

// Gregorian is the resolved gregorian calendar view DateTimeFormat consumes.
type Gregorian struct {
	Eras       struct{ Wide, Abbr, Narrow [2]string }
	Months     struct{ Wide, Abbr, Narrow, StandWide, StandAbbr, StandNarrow [12]string }
	Weekdays   struct{ Wide, Abbr, Narrow, Short, StandWide, StandAbbr, StandNarrow, StandShort [7]string }
	DayPeriods struct {
		AM, PM dayPeriodNames
		Flex   map[string]dayPeriodNames
	}
	DateFormats, TimeFormats, DateTimeFormats, DateTimeAtFormats [4]string
	AvailableFormats                                             map[string]string
	IntervalFormats                                              map[string]map[string]string
	IntervalFallback                                             string
	AppendItems                                                  map[string]string
}

// calendarNameKey indexes a CalendarNames entry by width and context.
type calendarNameKey struct{ width, context string }

// calendarNames holds the era/month/weekday/day-period name lists for one
// width/context combination. Quarters are not stored: no accessor reads them.
type calendarNames struct{ eras, months, weekdays, dayPeriods []string }

// calendarData is the per-locale gregorian intermediate the accessors read.
type calendarData struct {
	names                            map[calendarNameKey]calendarNames
	date, time, dateTime, dateTimeAt map[string]string
	available, appendItems           map[string]string
	intervalFallback                 string
	intervals                        map[string]map[string]string
}

// calendarNameKeyOrder is the fixed serialization order of CalendarNames
// entries written by tools/gen-cldr/codegen/date_encode.go.
var calendarNameKeyOrder = [...]calendarNameKey{
	{width: "wide", context: "format"},
	{width: "abbreviated", context: "format"},
	{width: "narrow", context: "format"},
	{width: "wide", context: "stand-alone"},
	{width: "abbreviated", context: "stand-alone"},
	{width: "narrow", context: "stand-alone"},
}

var (
	gregorianOnce sync.Once
	gregorianData map[Locale]calendarData

	dayPeriodOnce  sync.Once
	dayPeriodRules map[Locale][]DayPeriodRange

	calendarsOnce sync.Once
	calendarIDs   []string
)

func loadGregorian() {
	r := codec.NewReader(_dateGregorianBlob)
	gregorianData = codec.Uint16DeltaMap[Locale, calendarData](&r, decodeCalendar)
}

func decodeCalendar(r *codec.Reader) calendarData {
	names := make(map[calendarNameKey]calendarNames, len(calendarNameKeyOrder))
	for _, key := range calendarNameKeyOrder {
		names[key] = calendarNames{
			eras:       decodeStringSlice(r),
			months:     decodeStringSlice(r),
			weekdays:   decodeStringSlice(r),
			dayPeriods: decodeStringSlice(r),
		}
	}
	data := calendarData{names: names}
	data.date = r.StringRefMap(_data)
	data.time = r.StringRefMap(_data)
	data.dateTime = r.StringRefMap(_data)
	data.dateTimeAt = r.StringRefMap(_data)
	data.available = r.StringRefMap(_data)
	data.intervalFallback = r.StringRef(_data)
	data.intervals = decodeIntervalSkeletons(r)
	data.appendItems = r.StringRefMap(_data)
	return data
}

func decodeStringSlice(r *codec.Reader) []string {
	out := r.StringRefSlice(_data)
	if len(out) == 0 {
		return nil
	}
	return out
}

func decodeIntervalSkeletons(r *codec.Reader) map[string]map[string]string {
	return codec.StringRefKeyMap[map[string]string](r, _data, decodeIntervalFields)
}

func decodeIntervalFields(r *codec.Reader) map[string]string {
	return r.StringRefMap(_data)
}

func loadDayPeriods() {
	r := codec.NewReader(_dateDayPeriodBlob)
	dayPeriodRules = codec.Uint16DeltaMap[Locale, []DayPeriodRange](&r, decodeDayPeriodRules)
}

func decodeDayPeriodRules(r *codec.Reader) []DayPeriodRange {
	rules := codec.CountedSlice[DayPeriodRange](r, decodeDayPeriodRule)
	if len(rules) == 0 {
		return nil
	}
	return rules
}

func decodeDayPeriodRule(r *codec.Reader) DayPeriodRange {
	from := time.Duration(r.Uvarint())
	to := time.Duration(r.Uvarint())
	return DayPeriodRange{From: from, To: to, Type: r.StringRef(_data)}
}

var supported = codec.NewLazyStrings(_dateSupportedBlob, _data)

func loadCalendars() {
	calendarIDs = codec.StringRefSlice(_dateCalendarBlob, _data)
}
