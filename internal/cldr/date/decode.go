// Hand-written decode layer for the date domain. It expands the const blobs in
// data.go into the same in-memory structures the legacy root loadDateData
// produced, behind per-blob sync.Once gates.
//
// Locale handle ownership: the date blobs pack a locale index that must match
// the runtime locale assignment, so this package borrows the Locale handle type
// from the cldr/locale kernel package. That kernel assigns locale indices
// identically to the legacy root package, so keys align; importing the kernel
// (rather than the root) keeps date's compile cost off the root data bomb. The
// dependency is one-way (date -> cldr/locale; the kernel never imports date, so
// there is no cycle), and it is the permanent home every domain points at.

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
// period. It mirrors the legacy root type.
type dayPeriodNames struct{ Wide, Abbr, Narrow string }

// DayPeriodRange is one resolved day-period rule: a half-open [From, To)
// duration window (or a point window when From == To) tagged with its period
// type. It mirrors the legacy root type.
type DayPeriodRange struct {
	From time.Duration
	To   time.Duration
	Type string
}

// Gregorian is the resolved gregorian calendar view DateTimeFormat consumes. It
// mirrors the legacy root Gregorian field-for-field so formatter output is
// byte-identical.
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

// calendarNameKey indexes a CalendarNames entry by width and context, mirroring
// the legacy root key.
type calendarNameKey struct{ width, context string }

// calendarNames holds the era/month/weekday/day-period name lists for one
// width/context combination. Quarters are not stored: no accessor reads them.
type calendarNames struct{ eras, months, weekdays, dayPeriods []string }

// calendarData is the per-locale gregorian intermediate the accessors read,
// mirroring the legacy root calendarData (the gregorian calendar only).
type calendarData struct {
	names                            map[calendarNameKey]calendarNames
	date, time, dateTime, dateTimeAt map[string]string
	available, appendItems           map[string]string
	intervalFallback                 string
	intervals                        map[string]map[string]string
}

var (
	gregorianOnce sync.Once
	gregorianData map[Locale]calendarData

	dayPeriodOnce  sync.Once
	dayPeriodRules map[Locale][]DayPeriodRange

	supportedOnce sync.Once
	supportedTags []string

	calendarsOnce sync.Once
	calendarIDs   []string
)

// calendarNameKeyOrder is the fixed serialization order of the six CalendarNames
// entries, identical to the encoder side.
var calendarNameKeyOrder = [...]calendarNameKey{
	{width: "wide", context: "format"},
	{width: "abbreviated", context: "format"},
	{width: "narrow", context: "format"},
	{width: "wide", context: "stand-alone"},
	{width: "abbreviated", context: "stand-alone"},
	{width: "narrow", context: "stand-alone"},
}

func loadGregorian() {
	r := codec.NewReader(_dateGregorianBlob)
	count := r.Uvarint()
	gregorianData = make(map[Locale]calendarData, count)
	indices := codec.Delta(&r)
	for i := uint64(0); i < count; i++ {
		loc := Locale(indices.Next())
		gregorianData[loc] = decodeCalendar(&r)
	}
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
	data.date = decodeStringMap(r)
	data.time = decodeStringMap(r)
	data.dateTime = decodeStringMap(r)
	data.dateTimeAt = decodeStringMap(r)
	data.available = decodeStringMap(r)
	data.intervalFallback = r.StringRef(_data)
	data.intervals = decodeIntervalSkeletons(r)
	data.appendItems = decodeStringMap(r)
	return data
}

func decodeStringSlice(r *codec.Reader) []string {
	n := r.Uvarint()
	if n == 0 {
		return nil
	}
	out := make([]string, n)
	for i := range out {
		out[i] = r.StringRef(_data)
	}
	return out
}

func decodeStringMap(r *codec.Reader) map[string]string {
	n := r.Uvarint()
	if n == 0 {
		return nil
	}
	out := make(map[string]string, n)
	for i := uint64(0); i < n; i++ {
		key := r.StringRef(_data)
		out[key] = r.StringRef(_data)
	}
	return out
}

func decodeIntervalSkeletons(r *codec.Reader) map[string]map[string]string {
	n := r.Uvarint()
	if n == 0 {
		return nil
	}
	out := make(map[string]map[string]string, n)
	for i := uint64(0); i < n; i++ {
		key := r.StringRef(_data)
		out[key] = decodeStringMap(r)
	}
	return out
}

func loadDayPeriods() {
	r := codec.NewReader(_dateDayPeriodBlob)
	count := r.Uvarint()
	dayPeriodRules = make(map[Locale][]DayPeriodRange, count)
	indices := codec.Delta(&r)
	for i := uint64(0); i < count; i++ {
		loc := Locale(indices.Next())
		nRules := r.Uvarint()
		if nRules == 0 {
			dayPeriodRules[loc] = nil
			continue
		}
		rules := make([]DayPeriodRange, nRules)
		for j := range rules {
			from := time.Duration(r.Uvarint())
			to := time.Duration(r.Uvarint())
			rules[j] = DayPeriodRange{From: from, To: to, Type: r.StringRef(_data)}
		}
		dayPeriodRules[loc] = rules
	}
}

func loadSupported() {
	r := codec.NewReader(_dateSupportedBlob)
	count := r.Uvarint()
	supportedTags = make([]string, count)
	for i := range supportedTags {
		supportedTags[i] = r.StringRef(_data)
	}
}

func loadCalendars() {
	r := codec.NewReader(_dateCalendarBlob)
	count := r.Uvarint()
	calendarIDs = make([]string, count)
	for i := range calendarIDs {
		calendarIDs[i] = r.StringRef(_data)
	}
}
