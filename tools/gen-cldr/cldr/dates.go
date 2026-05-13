package cldr

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
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
		if locale == "und" {
			continue
		}
		path := filepath.Join(root, "cldr-dates-full", "main", locale, "ca-gregorian.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc struct {
			Main map[string]struct {
				Dates struct {
					Calendars map[string]json.RawMessage `json:"calendars"`
				} `json:"dates"`
			} `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, fmt.Errorf("dates body missing for %s", locale)
		}
		gregorian, ok := body.Dates.Calendars["gregorian"]
		if !ok {
			continue
		}
		calendar, err := parseCalendar(gregorian)
		if err != nil {
			return nil, fmt.Errorf("parse gregorian calendar for %s: %w", locale, err)
		}
		out[locale] = Dates{Calendars: map[string]Calendar{"gregorian": calendar}, DayPeriodRules: rules}
	}
	return out, nil
}

func parseCalendar(raw json.RawMessage) (Calendar, error) {
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
		DateFormats       map[string]json.RawMessage              `json:"dateFormats"`
		TimeFormats       map[string]json.RawMessage              `json:"timeFormats"`
		DateTimeFormats   map[string]json.RawMessage              `json:"dateTimeFormats"`
		DateTimeAtFormats struct {
			Standard map[string]json.RawMessage `json:"standard"`
		} `json:"dateTimeFormats-atTime"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return Calendar{}, err
	}
	names := make(map[CalendarNameKey]CalendarNames)
	for _, context := range []string{"format", "stand-alone"} {
		for _, width := range []string{"wide", "abbreviated", "narrow"} {
			names[CalendarNameKey{Width: width, Context: context}] = CalendarNames{
				Eras:       erasForWidth(doc.Eras.Names, doc.Eras.Abbr, doc.Eras.Narrow, width),
				Months:     orderedNumberValues(contextValues(doc.Months, context, width), 12),
				Weekdays:   orderedWeekdayValues(contextValues(doc.Days, context, width)),
				Quarters:   orderedNumberValues(contextValues(doc.Quarters, context, width), 4),
				DayPeriods: orderedDayPeriodValues(contextValues(doc.DayPeriods, context, width)),
			}
		}
	}
	dateTimeFormats := styleFormats(doc.DateTimeFormats)
	dateTimeAtFormats := styleFormats(doc.DateTimeAtFormats.Standard)
	if len(dateTimeAtFormats) == 0 {
		dateTimeAtFormats = dateTimeFormats
	}
	return Calendar{
		Names:             names,
		DateFormats:       styleFormats(doc.DateFormats),
		TimeFormats:       styleFormats(doc.TimeFormats),
		DateTimeFormats:   dateTimeFormats,
		DateTimeAtFormats: dateTimeAtFormats,
		AvailableFormats:  availableFormats(doc.DateTimeFormats["availableFormats"]),
		IntervalFormats:   intervalFormats(doc.DateTimeFormats["intervalFormats"]),
		AppendItems:       stringMap(doc.DateTimeFormats["appendItems"]),
	}, nil
}

func erasForWidth(names, abbr, narrow map[string]string, width string) []string {
	switch width {
	case "wide":
		return orderedEraValues(names)
	case "narrow":
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
	if context != "format" {
		return contextValues(values, "format", width)
	}
	return nil
}

func orderedEraValues(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	keys := make([]int, 0, len(values))
	for key := range values {
		idx, err := strconv.Atoi(key)
		if err == nil {
			keys = append(keys, idx)
		}
	}
	slices.Sort(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, values[strconv.Itoa(key)])
	}
	return out
}

func orderedNumberValues(values map[string]string, count int) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		if value := values[strconv.Itoa(i)]; value != "" {
			out = append(out, value)
		}
	}
	return out
}

func orderedWeekdayValues(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, 7)
	for _, day := range []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"} {
		if value := values[day]; value != "" {
			out = append(out, value)
		}
	}
	return out
}

func orderedDayPeriodValues(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	keys := []string{"midnight", "am", "noon", "pm", "morning1", "morning2", "afternoon1", "afternoon2", "evening1", "evening2", "night1", "night2"}
	out := make([]string, len(keys))
	for i, key := range keys {
		out[i] = values[key]
	}
	return out
}

func styleFormats(values map[string]json.RawMessage) map[string]string {
	out := make(map[string]string, 4)
	for _, style := range []string{"full", "long", "medium", "short"} {
		if value := rawString(values[style]); value != "" {
			out[style] = value
		}
	}
	return out
}

func stringMap(raw json.RawMessage) map[string]string {
	var values map[string]string
	if len(raw) == 0 || json.Unmarshal(raw, &values) != nil {
		return nil
	}
	return values
}

func availableFormats(raw json.RawMessage) map[string]string {
	var fields map[string]json.RawMessage
	if len(raw) == 0 || json.Unmarshal(raw, &fields) != nil {
		return nil
	}
	out := make(map[string]string, len(fields))
	for key, rawValue := range fields {
		if strings.HasSuffix(key, "alt-variant") {
			continue
		}
		if value := rawString(rawValue); value != "" {
			out[key] = value
		}
	}
	return out
}

func intervalFormats(raw json.RawMessage) IntervalFormats {
	var fields map[string]json.RawMessage
	if len(raw) == 0 || json.Unmarshal(raw, &fields) != nil {
		return IntervalFormats{}
	}
	out := IntervalFormats{BySkeleton: make(map[string]map[string]string)}
	for key, rawValue := range fields {
		if strings.HasSuffix(key, "alt-variant") {
			continue
		}
		if key == "intervalFormatFallback" {
			out.FallbackPattern = rawString(rawValue)
			continue
		}
		var byField map[string]json.RawMessage
		if json.Unmarshal(rawValue, &byField) != nil {
			continue
		}
		patterns := make(map[string]string, len(byField))
		for field, rawPattern := range byField {
			if strings.HasSuffix(field, "alt-variant") {
				continue
			}
			if value := rawString(rawPattern); value != "" {
				patterns[field] = value
			}
		}
		if len(patterns) > 0 {
			out.BySkeleton[key] = patterns
		}
	}
	if len(out.BySkeleton) == 0 {
		out.BySkeleton = nil
	}
	return out
}

func loadDayPeriodRules(root string) (map[string][]DayPeriodRange, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "dayPeriods.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read dayPeriods.json: %w", err)
	}
	var doc struct {
		Supplemental struct {
			DayPeriodRuleSet map[string]json.RawMessage `json:"dayPeriodRuleSet"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
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

func parseDayPeriodRuleSet(raw json.RawMessage) ([]DayPeriodRange, error) {
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
	out := make([]DayPeriodRange, 0, len(rules))
	for _, typ := range slices.Sorted(maps.Keys(rules)) {
		rule := rules[typ]
		if rule.At != "" {
			at, err := parseDayPeriodClock(rule.At)
			if err != nil {
				return nil, fmt.Errorf("parse day period at %q: %w", rule.At, err)
			}
			out = append(out, DayPeriodRange{From: at, To: at, Type: typ})
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
		out = append(out, DayPeriodRange{From: from, To: before, Type: typ})
	}
	return out, nil
}

func parseDayPeriodClock(value string) (time.Duration, error) {
	if value == "24:00" {
		return 24 * time.Hour, nil
	}
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid clock %q", value)
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute, nil
}

func rawString(raw json.RawMessage) string {
	var value string
	if len(raw) == 0 {
		return ""
	}
	if err := json.Unmarshal(raw, &value); err == nil {
		return value
	}
	var doc struct {
		Value string `json:"_value"`
	}
	if err := json.Unmarshal(raw, &doc); err == nil {
		return doc.Value
	}
	return ""
}
