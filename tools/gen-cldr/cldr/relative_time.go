package cldr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RelativeTimeFields map[string]map[string]RelativeTimeField

type RelativeTimeField struct {
	Future   map[string]string
	Past     map[string]string
	Relative map[string]string
}

func loadRelativeTimeFields(root string, locales []string) (map[string]RelativeTimeFields, error) {
	loaded := make(map[string]RelativeTimeFields)
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		path := filepath.Join(root, "cldr-dates-full", "main", locale, "dateFields.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc struct {
			Main map[string]struct {
				Dates struct {
					Fields map[string]map[string]json.RawMessage `json:"fields"`
				} `json:"dates"`
			} `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, fmt.Errorf("dateFields body missing for %s", locale)
		}
		fields, err := parseRelativeTimeFields(body.Dates.Fields)
		if err != nil {
			return nil, fmt.Errorf("parse relative time fields for %s: %w", locale, err)
		}
		if len(fields) > 0 {
			loaded[locale] = fields
		}
	}
	out := make(map[string]RelativeTimeFields, len(locales))
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		if fields, ok := loaded[locale]; ok {
			out[locale] = fields
			continue
		}
		if fields, ok := inheritedRelativeTimeFields(locale, loaded); ok {
			out[locale] = fields
		}
	}
	return out, nil
}

func inheritedRelativeTimeFields(locale string, loaded map[string]RelativeTimeFields) (RelativeTimeFields, bool) {
	for parent := parentLocale(locale); parent != ""; parent = parentLocale(parent) {
		if fields, ok := loaded[parent]; ok {
			return fields, true
		}
	}
	return nil, false
}

func parseRelativeTimeFields(raw map[string]map[string]json.RawMessage) (RelativeTimeFields, error) {
	out := make(RelativeTimeFields)
	for _, key := range relativeTimeFieldKeys {
		rawField, ok := raw[key.cldr]
		if !ok {
			continue
		}
		field, err := parseRelativeTimeField(rawField)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key.cldr, err)
		}
		if len(field.Future) == 0 && len(field.Past) == 0 && len(field.Relative) == 0 {
			continue
		}
		if out[key.unit] == nil {
			out[key.unit] = make(map[string]RelativeTimeField)
		}
		out[key.unit][key.style] = field
	}
	return out, nil
}

func parseRelativeTimeField(raw map[string]json.RawMessage) (RelativeTimeField, error) {
	field := RelativeTimeField{
		Future:   relativeTimePatternMap(raw["relativeTime-type-future"]),
		Past:     relativeTimePatternMap(raw["relativeTime-type-past"]),
		Relative: make(map[string]string),
	}
	for key, value := range raw {
		relativeKey, ok := strings.CutPrefix(key, "relative-type-")
		if !ok {
			continue
		}
		var literal string
		if err := json.Unmarshal(value, &literal); err != nil {
			return RelativeTimeField{}, err
		}
		field.Relative[relativeKey] = literal
	}
	if len(field.Relative) == 0 {
		field.Relative = nil
	}
	return field, nil
}

func relativeTimePatternMap(raw json.RawMessage) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	var patterns map[string]string
	if err := json.Unmarshal(raw, &patterns); err != nil {
		return nil
	}
	out := make(map[string]string, len(patterns))
	for key, value := range patterns {
		plural, ok := strings.CutPrefix(key, "relativeTimePattern-count-")
		if ok {
			out[plural] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

var relativeTimeFieldKeys = []struct {
	cldr  string
	unit  string
	style string
}{
	{cldr: "second", unit: "second", style: "long"},
	{cldr: "second-short", unit: "second", style: "short"},
	{cldr: "second-narrow", unit: "second", style: "narrow"},
	{cldr: "minute", unit: "minute", style: "long"},
	{cldr: "minute-short", unit: "minute", style: "short"},
	{cldr: "minute-narrow", unit: "minute", style: "narrow"},
	{cldr: "hour", unit: "hour", style: "long"},
	{cldr: "hour-short", unit: "hour", style: "short"},
	{cldr: "hour-narrow", unit: "hour", style: "narrow"},
	{cldr: "day", unit: "day", style: "long"},
	{cldr: "day-short", unit: "day", style: "short"},
	{cldr: "day-narrow", unit: "day", style: "narrow"},
	{cldr: "week", unit: "week", style: "long"},
	{cldr: "week-short", unit: "week", style: "short"},
	{cldr: "week-narrow", unit: "week", style: "narrow"},
	{cldr: "month", unit: "month", style: "long"},
	{cldr: "month-short", unit: "month", style: "short"},
	{cldr: "month-narrow", unit: "month", style: "narrow"},
	{cldr: "quarter", unit: "quarter", style: "long"},
	{cldr: "quarter-short", unit: "quarter", style: "short"},
	{cldr: "quarter-narrow", unit: "quarter", style: "narrow"},
	{cldr: "year", unit: "year", style: "long"},
	{cldr: "year-short", unit: "year", style: "short"},
	{cldr: "year-narrow", unit: "year", style: "narrow"},
}
