package codegen

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderDates(dates extract.Dates, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import (\n\t\"sync\"\n\t\"time\"\n)\n\n")
	b.WriteString("type CalendarNames struct{ Eras, Months, Weekdays, Quarters, DayPeriods []string }\n\n")
	b.WriteString("type dayPeriodNames struct{ Wide, Abbr, Narrow string }\n\n")
	b.WriteString("type DayPeriodRange struct { From time.Duration; To time.Duration; Type string }\n\n")
	b.WriteString("type Gregorian struct { Eras struct{ Wide, Abbr, Narrow [2]string }; Months struct{ Wide, Abbr, Narrow, StandWide, StandAbbr, StandNarrow [12]string }; Weekdays struct{ Wide, Abbr, Narrow, Short, StandWide, StandAbbr, StandNarrow, StandShort [7]string }; DayPeriods struct{ AM, PM dayPeriodNames; Flex map[string]dayPeriodNames }; DateFormats, TimeFormats, DateTimeFormats, DateTimeAtFormats [4]string; AvailableFormats map[string]string; IntervalFormats map[string]map[string]string; IntervalFallback string; AppendItems map[string]string }\n\n")
	b.WriteString("type calendarNameKey struct{ width, context string }\n\n")
	b.WriteString("type calendarData struct { names map[calendarNameKey]CalendarNames; date, time, dateTime, dateTimeAt, available, appendItems map[string]string; intervals IntervalFormats }\n\n")
	b.WriteString("type IntervalFormats struct{ FallbackPattern string; BySkeleton map[string]map[string]string }\n\n")
	b.WriteString(`var (
	datesByLocaleOnce sync.Once
	datesByLocale     map[Locale]map[string]calendarData
	dayPeriodRules    map[Locale][]DayPeriodRange
)

func dateDataByLocale() map[Locale]map[string]calendarData {
	datesByLocaleOnce.Do(loadDateData)
	return datesByLocale
}

func dayPeriodRuleData() map[Locale][]DayPeriodRange {
	datesByLocaleOnce.Do(loadDateData)
	return dayPeriodRules
}

func loadDateData() {
	datesByLocale = map[Locale]map[string]calendarData{
`)
	for _, locale := range sortedLocaleKeys(dates) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: {\n")
		for _, calendar := range slices.Sorted(maps.Keys(dates[locale].Calendars)) {
			data := dates[locale].Calendars[calendar]
			b.WriteString("\t\t")
			b.WriteString(strconv.Quote(calendar))
			b.WriteString(": {names: ")
			b.WriteString(calendarNamesMapLiteral(data.Names, table))
			b.WriteString(", date: ")
			b.WriteString(stringMapLiteral(data.DateFormats, table))
			b.WriteString(", time: ")
			b.WriteString(stringMapLiteral(data.TimeFormats, table))
			b.WriteString(", dateTime: ")
			b.WriteString(stringMapLiteral(data.DateTimeFormats, table))
			b.WriteString(", dateTimeAt: ")
			b.WriteString(stringMapLiteral(data.DateTimeAtFormats, table))
			b.WriteString(", available: ")
			b.WriteString(stringMapLiteral(data.AvailableFormats, table))
			b.WriteString(", intervals: ")
			b.WriteString(intervalFormatsLiteral(data.IntervalFormats, table))
			b.WriteString(", appendItems: ")
			b.WriteString(stringMapLiteral(data.AppendItems, table))
			b.WriteString("},\n")
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\tdayPeriodRules = ")
	b.WriteString(dayPeriodRulesLiteral(anyDayPeriodRules(dates), table))
	b.WriteString("\n")
	b.WriteString("}\n\n")
	b.WriteString(`func (l Locale) CalendarNames(calendar, width, context string) CalendarNames {
	if calendar == "" { calendar = "gregorian" }
	if width == "" { width = "wide" }
	if context == "" { context = "format" }
	return dateDataByLocale()[l][calendar].names[calendarNameKey{width: width, context: context}]
}

func (l Locale) DateFormat(calendar, style string) string {
	if calendar == "" { calendar = "gregorian" }
	return dateDataByLocale()[l][calendar].date[style]
}

func (l Locale) TimeFormat(calendar, style string) string {
	if calendar == "" { calendar = "gregorian" }
	return dateDataByLocale()[l][calendar].time[style]
}

func (l Locale) DateTimeFormat(calendar, style string) string {
	if calendar == "" { calendar = "gregorian" }
	return dateDataByLocale()[l][calendar].dateTime[style]
}

func (l Locale) DateTimeAtFormat(calendar, style string) string {
	if calendar == "" { calendar = "gregorian" }
	return dateDataByLocale()[l][calendar].dateTimeAt[style]
}

func (l Locale) AvailableFormats(calendar string) map[string]string {
	if calendar == "" { calendar = "gregorian" }
	return dateDataByLocale()[l][calendar].available
}

func (l Locale) IntervalFormats(calendar string) IntervalFormats {
	if calendar == "" { calendar = "gregorian" }
	return dateDataByLocale()[l][calendar].intervals
}

func DayPeriodFor(loc Locale, hour, minute int) string {
	value := time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute
	for _, rule := range dayPeriodRuleData()[loc] {
		if rule.From == rule.To && value == rule.From {
			return rule.Type
		}
	}
	for _, rule := range dayPeriodRuleData()[loc] {
		if rule.From == rule.To {
			continue
		}
		if rule.From < rule.To {
			if value >= rule.From && value < rule.To {
				return rule.Type
			}
			continue
		}
		if value >= rule.From || value < rule.To {
			return rule.Type
		}
	}
	return ""
}

func GregorianFor(loc Locale) Gregorian {
	data := dateDataByLocale()[loc]["gregorian"]
	var gregorian Gregorian
	wideFormat := data.names[calendarNameKey{width: "wide", context: "format"}]
	abbrFormat := data.names[calendarNameKey{width: "abbreviated", context: "format"}]
	narrowFormat := data.names[calendarNameKey{width: "narrow", context: "format"}]
	wideStandalone := data.names[calendarNameKey{width: "wide", context: "stand-alone"}]
	abbrStandalone := data.names[calendarNameKey{width: "abbreviated", context: "stand-alone"}]
	narrowStandalone := data.names[calendarNameKey{width: "narrow", context: "stand-alone"}]
	copy(gregorian.Eras.Wide[:], wideFormat.Eras)
	copy(gregorian.Eras.Abbr[:], abbrFormat.Eras)
	copy(gregorian.Eras.Narrow[:], narrowFormat.Eras)
	copy(gregorian.Months.Wide[:], wideFormat.Months)
	copy(gregorian.Months.Abbr[:], abbrFormat.Months)
	copy(gregorian.Months.Narrow[:], narrowFormat.Months)
	copy(gregorian.Months.StandWide[:], wideStandalone.Months)
	copy(gregorian.Months.StandAbbr[:], abbrStandalone.Months)
	copy(gregorian.Months.StandNarrow[:], narrowStandalone.Months)
	copy(gregorian.Weekdays.Wide[:], wideFormat.Weekdays)
	copy(gregorian.Weekdays.Abbr[:], abbrFormat.Weekdays)
	copy(gregorian.Weekdays.Narrow[:], narrowFormat.Weekdays)
	copy(gregorian.Weekdays.StandWide[:], wideStandalone.Weekdays)
	copy(gregorian.Weekdays.StandAbbr[:], abbrStandalone.Weekdays)
	copy(gregorian.Weekdays.StandNarrow[:], narrowStandalone.Weekdays)
	gregorian.DayPeriods.AM = dayPeriodNames{Wide: wideFormat.DayPeriods[1], Abbr: abbrFormat.DayPeriods[1], Narrow: narrowFormat.DayPeriods[1]}
	gregorian.DayPeriods.PM = dayPeriodNames{Wide: wideFormat.DayPeriods[3], Abbr: abbrFormat.DayPeriods[3], Narrow: narrowFormat.DayPeriods[3]}
	gregorian.DayPeriods.Flex = flexibleDayPeriodNames(data.names)
	gregorian.DateFormats = styleArray(data.date)
	gregorian.TimeFormats = styleArray(data.time)
	gregorian.DateTimeFormats = styleArray(data.dateTime)
	gregorian.DateTimeAtFormats = styleArray(data.dateTimeAt)
	gregorian.AvailableFormats = data.available
	gregorian.IntervalFormats = data.intervals.BySkeleton
	gregorian.IntervalFallback = data.intervals.FallbackPattern
	gregorian.AppendItems = data.appendItems
	return gregorian
}

func styleArray(values map[string]string) [4]string { return [4]string{values["full"], values["long"], values["medium"], values["short"]} }

func dayPeriodName(values []string, key string) string {
	index := dayPeriodNameIndex(key)
	if index < len(values) {
		return values[index]
	}
	return ""
}

func dayPeriodNameIndex(key string) int {
	switch key {
	case "midnight":
		return 0
	case "am":
		return 1
	case "noon":
		return 2
	case "pm":
		return 3
	case "morning1":
		return 4
	case "morning2":
		return 5
	case "afternoon1":
		return 6
	case "afternoon2":
		return 7
	case "evening1":
		return 8
	case "evening2":
		return 9
	case "night1":
		return 10
	case "night2":
		return 11
	default:
		return 100
	}
}

func flexibleDayPeriodNames(names map[calendarNameKey]CalendarNames) map[string]dayPeriodNames {
	wide := names[calendarNameKey{width: "wide", context: "format"}].DayPeriods
	abbr := names[calendarNameKey{width: "abbreviated", context: "format"}].DayPeriods
	narrow := names[calendarNameKey{width: "narrow", context: "format"}].DayPeriods
	keys := []string{"midnight", "noon", "morning1", "morning2", "afternoon1", "afternoon2", "evening1", "evening2", "night1", "night2"}
	out := make(map[string]dayPeriodNames)
	for _, key := range keys {
		if name := dayPeriodName(wide, key); name != "" {
			out[key] = dayPeriodNames{Wide: name, Abbr: dayPeriodName(abbr, key), Narrow: dayPeriodName(narrow, key)}
		}
	}
	return out
}
`)
	return FormatFile([]byte(b.String()))
}

func renderPreference(data extract.PreferenceData, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import (\n\t\"sync\"\n\t\"time\"\n)\n\n")
	b.WriteString("type weekPreference struct{ first, weekendStart, weekendEnd time.Weekday; minDays int }\n\n")
	b.WriteString(`var (
	preferenceDataOnce     sync.Once
	hourCyclePreference    map[string][]string
	weekPreferenceByRegion map[string]weekPreference
	calendarPreference     map[string][]string
)

func loadPreferenceData() {
	hourCyclePreference = map[string][]string{
`)
	for _, region := range slices.Sorted(maps.Keys(data.HourCycle)) {
		fmt.Fprintf(&b, "\t%q: {", region)
		for _, value := range data.HourCycle[region] {
			fmt.Fprintf(&b, "%q, ", value)
			table.Add(value)
		}
		b.WriteString("},\n")
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\tweekPreferenceByRegion = map[string]weekPreference{\n")
	for _, region := range slices.Sorted(maps.Keys(data.Week)) {
		week := data.Week[region]
		fmt.Fprintf(&b, "\t%q: {first: %s, weekendStart: %s, weekendEnd: %s, minDays: %d},\n", region, weekdayLiteral(week.FirstDay), weekdayLiteral(week.WeekendStart), weekdayLiteral(week.WeekendEnd), week.MinDays)
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\tcalendarPreference = map[string][]string{\n")
	for _, region := range slices.Sorted(maps.Keys(data.Calendar)) {
		fmt.Fprintf(&b, "\t%q: {", region)
		for _, value := range data.Calendar[region] {
			fmt.Fprintf(&b, "%q, ", value)
			table.Add(value)
		}
		b.WriteString("},\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString(`func HourCyclePreference(region string) []string {
	preferenceDataOnce.Do(loadPreferenceData)
	if data, ok := hourCyclePreference[region]; ok { return data }
	return hourCyclePreference["001"]
}

func HasHourCyclePreference(region string) bool {
	preferenceDataOnce.Do(loadPreferenceData)
	_, ok := hourCyclePreference[region]
	return ok
}

func FirstDayOfWeek(region string) time.Weekday {
	return weekPreferenceRecord(region).first
}

func Weekend(region string) (start, end time.Weekday) {
	week := weekPreferenceRecord(region)
	return week.weekendStart, week.weekendEnd
}

func MinimalDaysInFirstWeek(region string) int {
	return weekPreferenceRecord(region).minDays
}

func CalendarPreference(region string) []string {
	preferenceDataOnce.Do(loadPreferenceData)
	if data, ok := calendarPreference[region]; ok { return data }
	return calendarPreference["001"]
}

func HasCalendarPreference(region string) bool {
	preferenceDataOnce.Do(loadPreferenceData)
	_, ok := calendarPreference[region]
	return ok
}

func HasWeekPreference(region string) bool {
	preferenceDataOnce.Do(loadPreferenceData)
	_, ok := weekPreferenceByRegion[region]
	return ok
}

func weekPreferenceRecord(region string) weekPreference {
	preferenceDataOnce.Do(loadPreferenceData)
	if data, ok := weekPreferenceByRegion[region]; ok { return data }
	return weekPreferenceByRegion["001"]
}
`)
	return FormatFile([]byte(b.String()))
}

func anyDayPeriodRules(dates extract.Dates) map[string][]cldr.DayPeriodRange {
	out := make(map[string][]cldr.DayPeriodRange)
	for _, locale := range sortedLocaleKeys(dates) {
		for key, rules := range dates[locale].DayPeriodRules {
			out[key] = rules
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func dayPeriodRulesLiteral(values map[string][]cldr.DayPeriodRange, table *StringTable) string {
	if len(values) == 0 {
		return "map[Locale][]DayPeriodRange{}"
	}
	var b strings.Builder
	b.WriteString("map[Locale][]DayPeriodRange{\n")
	for _, locale := range sortedLocaleKeys(values) {
		b.WriteString("localeIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(dayPeriodRangeSliceLiteral(values[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("}")
	return b.String()
}

func dayPeriodRangeSliceLiteral(values []cldr.DayPeriodRange, table *StringTable) string {
	if len(values) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("[]DayPeriodRange{")
	for _, value := range values {
		b.WriteString("{From: ")
		b.WriteString(durationLiteral(value.From))
		b.WriteString(", To: ")
		b.WriteString(durationLiteral(value.To))
		b.WriteString(", Type: ")
		b.WriteString(refStringLiteral(value.Type, table))
		b.WriteString("}, ")
	}
	b.WriteString("}")
	return b.String()
}

func durationLiteral(value time.Duration) string {
	return strconv.FormatInt(int64(value), 10)
}

func calendarNamesMapLiteral(values map[cldr.CalendarNameKey]cldr.CalendarNames, table *StringTable) string {
	if len(values) == 0 {
		return "nil"
	}
	keys := slices.SortedFunc(maps.Keys(values), func(a, b cldr.CalendarNameKey) int {
		return strings.Compare(a.Width+"/"+a.Context, b.Width+"/"+b.Context)
	})
	var b strings.Builder
	b.WriteString("map[calendarNameKey]CalendarNames{\n")
	for _, key := range keys {
		b.WriteString("{width: ")
		b.WriteString(strconv.Quote(key.Width))
		b.WriteString(", context: ")
		b.WriteString(strconv.Quote(key.Context))
		b.WriteString("}: ")
		b.WriteString(calendarNamesLiteral(values[key], table))
		b.WriteString(",\n")
	}
	b.WriteString("}")
	return b.String()
}

func calendarNamesLiteral(names cldr.CalendarNames, table *StringTable) string {
	return "CalendarNames{" +
		"Eras: " + stringSliceLiteral(names.Eras, table) + ", " +
		"Months: " + stringSliceLiteral(names.Months, table) + ", " +
		"Weekdays: " + stringSliceLiteral(names.Weekdays, table) + ", " +
		"Quarters: " + stringSliceLiteral(names.Quarters, table) + ", " +
		"DayPeriods: " + stringSliceLiteral(names.DayPeriods, table) + "}"
}

func intervalFormatsLiteral(values cldr.IntervalFormats, table *StringTable) string {
	return "IntervalFormats{FallbackPattern: " + refStringLiteral(values.FallbackPattern, table) + ", BySkeleton: " + nestedStringMapLiteral(values.BySkeleton, table) + "}"
}

func stringSliceLiteral(values []string, table *StringTable) string {
	if len(values) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("[]string{")
	for _, value := range values {
		b.WriteString(refStringLiteral(value, table))
		b.WriteString(", ")
	}
	b.WriteString("}")
	return b.String()
}

func nestedStringMapLiteral(values map[string]map[string]string, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]map[string]string", values, func(value map[string]string) string {
		return stringMapLiteral(value, table)
	})
}

func weekdayLiteral(day string) string {
	switch day {
	case "sun":
		return "time.Sunday"
	case "mon":
		return "time.Monday"
	case "tue":
		return "time.Tuesday"
	case "wed":
		return "time.Wednesday"
	case "thu":
		return "time.Thursday"
	case "fri":
		return "time.Friday"
	case "sat":
		return "time.Saturday"
	default:
		return "time.Sunday"
	}
}
