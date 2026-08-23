package cldr

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
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
		if locale == undefinedLocale {
			continue
		}
		path := filepath.Join(root, "cldr-dates-full", "main", locale, "dateFields.json")
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
					Fields map[string]map[string]jsontext.Value `json:"fields"`
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
		if body.Dates.Fields == nil {
			return nil, fmt.Errorf("dateFields data missing for %s", locale)
		}
		fields, err := parseRelativeTimeFields(body.Dates.Fields)
		if err != nil {
			return nil, fmt.Errorf("parse relative time fields for %s: %w", locale, err)
		}
		if len(fields) > 0 {
			loaded[locale] = fields
		}
	}
	return inheritedLocaleData(locales, loaded), nil
}

func parseRelativeTimeFields(raw map[string]map[string]jsontext.Value) (RelativeTimeFields, error) {
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

func parseRelativeTimeField(raw map[string]jsontext.Value) (RelativeTimeField, error) {
	future, err := relativeTimePatternMap(raw["relativeTime-type-future"])
	if err != nil {
		return RelativeTimeField{}, fmt.Errorf("future: %w", err)
	}
	past, err := relativeTimePatternMap(raw["relativeTime-type-past"])
	if err != nil {
		return RelativeTimeField{}, fmt.Errorf("past: %w", err)
	}
	field := RelativeTimeField{
		Future:   future,
		Past:     past,
		Relative: make(map[string]string),
	}
	for key, value := range raw {
		relativeKey, ok, err := relativeLiteralKeyFromField(key)
		if err != nil {
			return RelativeTimeField{}, fmt.Errorf("parse %s: %w", key, err)
		}
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

func relativeLiteralKeyFromField(field string) (string, bool, error) {
	key, ok := strings.CutPrefix(field, "relative-type-")
	if !ok {
		return "", false, nil
	}
	value, err := strconv.ParseFloat(key, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return "", true, fmt.Errorf("invalid relative literal key %q: expected finite numeric key", key)
	}
	return key, true, nil
}

func relativeTimePatternMap(raw jsontext.Value) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var patterns map[string]string
	if err := json.Unmarshal(raw, &patterns); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(patterns))
	for key, value := range patterns {
		plural, ok, err := pluralCategoryFromField(key, "relativeTimePattern-count-")
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", key, err)
		}
		if ok {
			if value == "" {
				return nil, fmt.Errorf("%s empty", key)
			}
			out[plural] = value
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

type relativeTimeFieldKey struct {
	cldr  string
	unit  string
	style string
}

var relativeTimeFieldKeys = [...]relativeTimeFieldKey{
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
