package cldr

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Dates struct {
	Calendars      map[string]Calendar
	DayPeriodRules map[string][]DayPeriodRange
}

type Calendar struct {
	Names             map[CalendarNameKey]CalendarNames
	DateFormats       map[string]string
	TimeFormats       map[string]string
	DateTimeFormats   map[string]string
	DateTimeAtFormats map[string]string
	AvailableFormats  map[string]string
	IntervalFormats   IntervalFormats
	AppendItems       map[string]string
}

type CalendarNameKey struct {
	Width   string
	Context string
}

type CalendarNames struct {
	Eras, Months, Weekdays, Quarters, DayPeriods []string
}

type IntervalFormats struct {
	FallbackPattern string
	BySkeleton      map[string]map[string]string
}

type DayPeriodRange struct {
	From time.Duration
	To   time.Duration
	Type string
}

func loadDates(root string, locales []string) (map[string]Dates, error) {
	rules, err := loadDayPeriodRules(root)
	if err != nil {
		return nil, err
	}
	out := make(map[string]Dates)
	for _, locale := range locales {
		if locale == undefinedLocale {
			continue
		}
		path := filepath.Join(root, "cldr-dates-full", "main", locale, gregorianCalendarFile)
		raw, ok, err := readOptionalFile(path)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		var doc struct {
			Main map[string]struct {
				Dates struct {
					Calendars map[string]jsontext.Value `json:"calendars"`
				} `json:"dates"`
			} `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if doc.Main == nil {
			return nil, fmt.Errorf("dates body missing for %s", locale)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, fmt.Errorf("dates body missing for %s", locale)
		}
		if body.Dates.Calendars == nil {
			return nil, fmt.Errorf("calendar data missing for %s", locale)
		}
		gregorian, ok := body.Dates.Calendars[gregorianCalendarType]
		if !ok {
			return nil, fmt.Errorf("%s calendar missing for %s", gregorianCalendarType, locale)
		}
		calendar, err := parseCalendar(gregorian)
		if err != nil {
			return nil, fmt.Errorf("parse %s calendar for %s: %w", gregorianCalendarType, locale, err)
		}
		out[locale] = Dates{Calendars: map[string]Calendar{gregorianCalendarType: calendar}, DayPeriodRules: rules}
	}
	return out, nil
}

func parseCalendar(raw jsontext.Value) (Calendar, error) {
	var fields map[string]jsontext.Value
	if err := json.Unmarshal(raw, &fields); err != nil {
		return Calendar{}, err
	}
	if fields == nil {
		return Calendar{}, fmt.Errorf("expected calendar object")
	}
	var doc struct {
		Eras struct {
			Names  map[string]string `json:"eraNames"`
			Abbr   map[string]string `json:"eraAbbr"`
			Narrow map[string]string `json:"eraNarrow"`
		} `json:"eras"`
		Months            map[string]map[string]map[string]string `json:"months"`
		Days              map[string]map[string]map[string]string `json:"days"`
		Quarters          map[string]map[string]map[string]string `json:"quarters"`
		DayPeriods        map[string]map[string]map[string]string `json:"dayPeriods"`
		DateFormats       map[string]jsontext.Value               `json:"dateFormats"`
		TimeFormats       map[string]jsontext.Value               `json:"timeFormats"`
		DateTimeFormats   map[string]jsontext.Value               `json:"dateTimeFormats"`
		DateTimeAtFormats struct {
			Standard map[string]jsontext.Value `json:"standard"`
		} `json:"dateTimeFormats-atTime"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return Calendar{}, err
	}
	names := make(map[CalendarNameKey]CalendarNames)
	for _, context := range calendarNameContexts {
		for _, width := range calendarNameWidths {
			names[CalendarNameKey{Width: width, Context: context}] = CalendarNames{
				Eras:       erasForWidth(doc.Eras.Names, doc.Eras.Abbr, doc.Eras.Narrow, width),
				Months:     orderedNumberValues(contextValues(doc.Months, context, width), 12),
				Weekdays:   orderedWeekdayValues(contextValues(doc.Days, context, width)),
				Quarters:   orderedNumberValues(contextValues(doc.Quarters, context, width), 4),
				DayPeriods: orderedDayPeriodValues(contextValues(doc.DayPeriods, context, width)),
			}
		}
	}
	dateFormats, err := requiredStyleFormats(doc.DateFormats)
	if err != nil {
		return Calendar{}, fmt.Errorf("parse dateFormats: %w", err)
	}
	if err := validateDatePatternMap(dateFormats); err != nil {
		return Calendar{}, fmt.Errorf("validate dateFormats: %w", err)
	}
	timeFormats, err := requiredStyleFormats(doc.TimeFormats)
	if err != nil {
		return Calendar{}, fmt.Errorf("parse timeFormats: %w", err)
	}
	if err := validateDatePatternMap(timeFormats); err != nil {
		return Calendar{}, fmt.Errorf("validate timeFormats: %w", err)
	}
	dateTimeFormats, err := requiredStyleFormats(doc.DateTimeFormats)
	if err != nil {
		return Calendar{}, fmt.Errorf("parse dateTimeFormats: %w", err)
	}
	if err := validateTemplateMap(dateTimeFormats, false); err != nil {
		return Calendar{}, fmt.Errorf("validate dateTimeFormats: %w", err)
	}
	dateTimeAtFormats, err := optionalStyleFormats(doc.DateTimeAtFormats.Standard)
	if err != nil {
		return Calendar{}, fmt.Errorf("parse dateTimeFormats-atTime: %w", err)
	}
	if len(dateTimeAtFormats) == 0 {
		dateTimeAtFormats = dateTimeFormats
	} else if err := validateTemplateMap(dateTimeAtFormats, false); err != nil {
		return Calendar{}, fmt.Errorf("validate dateTimeFormats-atTime: %w", err)
	}
	available, err := availableFormats(doc.DateTimeFormats["availableFormats"])
	if err != nil {
		return Calendar{}, fmt.Errorf("parse availableFormats: %w", err)
	}
	interval, err := intervalFormats(doc.DateTimeFormats["intervalFormats"])
	if err != nil {
		return Calendar{}, fmt.Errorf("parse intervalFormats: %w", err)
	}
	appendItems, err := appendItems(doc.DateTimeFormats["appendItems"])
	if err != nil {
		return Calendar{}, fmt.Errorf("parse appendItems: %w", err)
	}
	return Calendar{
		Names:             names,
		DateFormats:       dateFormats,
		TimeFormats:       timeFormats,
		DateTimeFormats:   dateTimeFormats,
		DateTimeAtFormats: dateTimeAtFormats,
		AvailableFormats:  available,
		IntervalFormats:   interval,
		AppendItems:       appendItems,
	}, nil
}

var (
	calendarNameContexts = [...]string{calendarNameContextFormat, calendarNameContextStandalone}
	calendarNameWidths   = [...]string{calendarNameWidthWide, calendarNameWidthAbbreviated, calendarNameWidthNarrow}
	calendarStyleOrder   = [...]string{"full", "long", "medium", "short"}
	weekdayNameOrder     = [...]string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
	dayPeriodNameOrder   = [...]string{"midnight", "am", "noon", "pm", "morning1", "morning2", "afternoon1", "afternoon2", "evening1", "evening2", "night1", "night2"}
)

const (
	calendarNameContextFormat     = "format"
	calendarNameContextStandalone = "stand-alone"

	calendarNameWidthWide        = "wide"
	calendarNameWidthAbbreviated = "abbreviated"
	calendarNameWidthNarrow      = "narrow"

	calendarAltVariantSuffix = "alt-variant"

	gregorianCalendarType = "gregorian"
	gregorianCalendarFile = "ca-gregorian.json"
)

func erasForWidth(names, abbr, narrow map[string]string, width string) []string {
	switch width {
	case calendarNameWidthWide:
		return orderedEraValues(names)
	case calendarNameWidthNarrow:
		return orderedEraValues(narrow)
	default:
		return orderedEraValues(abbr)
	}
}

func contextValues(values map[string]map[string]map[string]string, context, width string) map[string]string {
	if byContext, ok := values[context]; ok {
		if byWidth, ok := byContext[width]; ok {
			return byWidth
		}
	}
	if context != calendarNameContextFormat {
		return contextValues(values, calendarNameContextFormat, width)
	}
	return nil
}

func orderedEraValues(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	maxIndex := -1
	for key := range values {
		idx, err := strconv.Atoi(key)
		if err == nil && idx >= 0 && idx > maxIndex {
			maxIndex = idx
		}
	}
	if maxIndex < 0 {
		return nil
	}
	out := make([]string, maxIndex+1)
	for key, value := range values {
		idx, err := strconv.Atoi(key)
		if err == nil && idx >= 0 {
			out[idx] = value
		}
	}
	return out
}

func orderedNumberValues(values map[string]string, count int) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, count)
	for i := 1; i <= count; i++ {
		out[i-1] = values[strconv.Itoa(i)]
	}
	return out
}

func orderedWeekdayValues(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(weekdayNameOrder))
	for i, day := range weekdayNameOrder {
		out[i] = values[day]
	}
	return out
}

func orderedDayPeriodValues(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(dayPeriodNameOrder))
	for i, key := range dayPeriodNameOrder {
		out[i] = values[key]
	}
	return out
}

func requiredStyleFormats(values map[string]jsontext.Value) (map[string]string, error) {
	if values == nil {
		return nil, fmt.Errorf("expected style format map")
	}
	return styleFormats(values, true)
}

func optionalStyleFormats(values map[string]jsontext.Value) (map[string]string, error) {
	if values == nil {
		return nil, nil
	}
	return styleFormats(values, false)
}

func styleFormats(values map[string]jsontext.Value, required bool) (map[string]string, error) {
	out := make(map[string]string, 4)
	for _, style := range calendarStyleOrder {
		raw, ok := values[style]
		if !ok {
			if required {
				return nil, fmt.Errorf("missing %s", style)
			}
			continue
		}
		value, err := rawString(raw)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", style, err)
		}
		if value == "" {
			return nil, fmt.Errorf("empty pattern for %s", style)
		}
		out[style] = value
	}
	return out, nil
}

func stringMap(raw jsontext.Value) (map[string]string, error) {
	var values map[string]string
	if len(raw) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func availableFormats(raw jsontext.Value) (map[string]string, error) {
	var fields map[string]jsontext.Value
	if len(raw) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(fields))
	for key, rawValue := range fields {
		if isCalendarAltVariant(key) {
			continue
		}
		value, err := rawString(rawValue)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", key, err)
		}
		if value == "" {
			return nil, fmt.Errorf("empty pattern for %s", key)
		}
		if isExecutableDateSkeleton(key) {
			if err := validateExecutableDatePattern(value); err != nil {
				return nil, fmt.Errorf("validate %s: %w", key, err)
			}
		}
		out[key] = value
	}
	return out, nil
}

func intervalFormats(raw jsontext.Value) (IntervalFormats, error) {
	var fields map[string]jsontext.Value
	if len(raw) == 0 {
		return IntervalFormats{}, nil
	}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return IntervalFormats{}, err
	}
	out := IntervalFormats{BySkeleton: make(map[string]map[string]string)}
	hasFallback := false
	for key, rawValue := range fields {
		if isCalendarAltVariant(key) {
			continue
		}
		if key == "intervalFormatFallback" {
			value, err := rawString(rawValue)
			if err != nil {
				return IntervalFormats{}, fmt.Errorf("parse intervalFormatFallback: %w", err)
			}
			if err := validateIndexedTemplate(value, false); err != nil {
				return IntervalFormats{}, fmt.Errorf("invalid intervalFormatFallback: %w", err)
			}
			out.FallbackPattern = value
			hasFallback = true
			continue
		}
		var byField map[string]jsontext.Value
		if err := json.Unmarshal(rawValue, &byField); err != nil {
			return IntervalFormats{}, fmt.Errorf("parse skeleton %s: %w", key, err)
		}
		patterns := make(map[string]string, len(byField))
		for field, rawPattern := range byField {
			if isCalendarAltVariant(field) {
				continue
			}
			value, err := rawString(rawPattern)
			if err != nil {
				return IntervalFormats{}, fmt.Errorf("parse skeleton %s field %s: %w", key, field, err)
			}
			if value == "" {
				return IntervalFormats{}, fmt.Errorf("empty pattern for skeleton %s field %s", key, field)
			}
			if len(field) != 1 {
				return IntervalFormats{}, fmt.Errorf("invalid interval field key %q for skeleton %s", field, key)
			}
			if _, ok := executableDatePatternFields[field[0]]; !ok {
				return IntervalFormats{}, fmt.Errorf("invalid interval field key %q for skeleton %s", field, key)
			}
			if err := validateExecutableDatePattern(value); err != nil {
				return IntervalFormats{}, fmt.Errorf("validate skeleton %s field %s: %w", key, field, err)
			}
			patterns[field] = value
		}
		if len(patterns) > 0 {
			out.BySkeleton[key] = patterns
		}
	}
	if !hasFallback {
		return IntervalFormats{}, fmt.Errorf("missing intervalFormatFallback")
	}
	if len(out.BySkeleton) == 0 {
		out.BySkeleton = nil
	}
	return out, nil
}

func appendItems(raw jsontext.Value) (map[string]string, error) {
	values, err := stringMap(raw)
	if err != nil {
		return nil, err
	}
	for key, value := range values {
		allowThird := key != "Timezone"
		if err := validateIndexedTemplate(value, allowThird); err != nil {
			return nil, fmt.Errorf("invalid appendItems %s: %w", key, err)
		}
	}
	return values, nil
}

func validateDatePatternMap(values map[string]string) error {
	for key, value := range values {
		if err := validateExecutableDatePattern(value); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	return nil
}

func validateTemplateMap(values map[string]string, allowThird bool) error {
	for key, value := range values {
		if err := validateIndexedTemplate(value, allowThird); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	return nil
}

func isCalendarAltVariant(key string) bool {
	return strings.HasSuffix(key, calendarAltVariantSuffix)
}

func loadDayPeriodRules(root string) (map[string][]DayPeriodRange, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "dayPeriods.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			DayPeriodRuleSet map[string]jsontext.Value `json:"dayPeriodRuleSet"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if doc.Supplemental.DayPeriodRuleSet == nil {
		return nil, fmt.Errorf("expected supplemental dayPeriodRuleSet map")
	}
	out := make(map[string][]DayPeriodRange)
	for _, locale := range slices.Sorted(maps.Keys(doc.Supplemental.DayPeriodRuleSet)) {
		rawRules := doc.Supplemental.DayPeriodRuleSet[locale]
		rules, err := parseDayPeriodRuleSet(rawRules)
		if err != nil {
			return nil, fmt.Errorf("parse day period rules for %s: %w", locale, err)
		}
		out[locale] = rules
	}
	return out, nil
}

func parseDayPeriodRuleSet(raw jsontext.Value) ([]DayPeriodRange, error) {
	var wrapped struct {
		DayPeriodRules map[string]struct {
			At     string `json:"_at"`
			From   string `json:"_from"`
			Before string `json:"_before"`
		} `json:"dayPeriodRules"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}
	rules := wrapped.DayPeriodRules
	if len(rules) == 0 {
		if err := json.Unmarshal(raw, &rules); err != nil {
			return nil, err
		}
	}
	keys := slices.Sorted(maps.Keys(rules))
	out := make([]DayPeriodRange, len(keys))
	for i, typ := range keys {
		rule := rules[typ]
		if rule.At != "" {
			at, err := parseDayPeriodClock(rule.At)
			if err != nil {
				return nil, fmt.Errorf("parse day period at %q: %w", rule.At, err)
			}
			out[i] = DayPeriodRange{From: at, To: at, Type: typ}
			continue
		}
		from, err := parseDayPeriodClock(rule.From)
		if err != nil {
			return nil, fmt.Errorf("parse day period from %q: %w", rule.From, err)
		}
		before, err := parseDayPeriodClock(rule.Before)
		if err != nil {
			return nil, fmt.Errorf("parse day period before %q: %w", rule.Before, err)
		}
		out[i] = DayPeriodRange{From: from, To: before, Type: typ}
	}
	return out, nil
}

func parseDayPeriodClock(value string) (time.Duration, error) {
	if value == "24:00" {
		return 24 * time.Hour, nil
	}
	rawHour, rawMinute, ok := strings.Cut(value, ":")
	if !ok {
		return 0, fmt.Errorf("invalid clock %q", value)
	}
	hour, err := strconv.Atoi(rawHour)
	if err != nil {
		return 0, err
	}
	minute, err := strconv.Atoi(rawMinute)
	if err != nil {
		return 0, err
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("invalid clock %q", value)
	}
	return time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute, nil
}

func rawString(raw jsontext.Value) (string, error) {
	var value string
	if len(raw) == 0 {
		return "", nil
	}
	if err := json.Unmarshal(raw, &value); err == nil {
		return value, nil
	}
	var doc struct {
		Value string `json:"_value"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", err
	}
	return doc.Value, nil
}
