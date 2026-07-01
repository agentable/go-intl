package cldr

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	cldrpattern "github.com/agentable/go-intl/internal/pattern"
)

type ListPatterns map[string]map[string]ListPattern

type ListPattern struct {
	Pair   string `json:"2"`
	Start  string `json:"start"`
	Middle string `json:"middle"`
	End    string `json:"end"`
}

func loadListPatterns(root string, locales []string) (map[string]ListPatterns, error) {
	loaded := make(map[string]ListPatterns)
	for _, locale := range locales {
		if locale == undefinedLocale {
			continue
		}
		path := filepath.Join(root, "cldr-misc-full", "main", locale, "listPatterns.json")
		raw, ok, err := readOptionalFile(path)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		var doc struct {
			Main map[string]struct {
				ListPatterns map[string]ListPattern `json:"listPatterns"`
			} `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, fmt.Errorf("listPatterns body missing for %s", locale)
		}
		if body.ListPatterns == nil {
			return nil, fmt.Errorf("listPatterns data missing for %s", locale)
		}
		patterns := make(ListPatterns)
		for _, key := range listPatternKeys {
			pattern, ok := body.ListPatterns[key.cldr]
			if !ok {
				return nil, fmt.Errorf("listPatterns %s missing for %s", key.cldr, locale)
			}
			if err := validateListPattern(key.cldr, pattern); err != nil {
				return nil, err
			}
			if patterns[key.typ] == nil {
				patterns[key.typ] = make(map[string]ListPattern)
			}
			patterns[key.typ][key.style] = pattern
		}
		if len(patterns) > 0 {
			loaded[locale] = patterns
		}
	}
	return inheritedLocaleData(locales, loaded), nil
}

func validateListPattern(cldrKey string, pattern ListPattern) error {
	for _, field := range [...]struct {
		name  string
		value string
	}{
		{name: "2", value: pattern.Pair},
		{name: "start", value: pattern.Start},
		{name: "middle", value: pattern.Middle},
		{name: "end", value: pattern.End},
	} {
		if strings.Count(field.value, "{0}") != 1 || strings.Count(field.value, "{1}") != 1 {
			return fmt.Errorf("%s.%s invalid: expected one {0} and one {1}, got %q", cldrKey, field.name, field.value)
		}
		if _, err := cldrpattern.Partition(field.value); err != nil {
			return fmt.Errorf("%s.%s invalid pattern: %w", cldrKey, field.name, err)
		}
	}
	return nil
}

type listPatternKey struct {
	cldr  string
	typ   string
	style string
}

var listPatternKeys = [...]listPatternKey{
	{cldr: "listPattern-type-standard", typ: "conjunction", style: "long"},
	{cldr: "listPattern-type-standard-short", typ: "conjunction", style: "short"},
	{cldr: "listPattern-type-standard-narrow", typ: "conjunction", style: "narrow"},
	{cldr: "listPattern-type-or", typ: "disjunction", style: "long"},
	{cldr: "listPattern-type-or-short", typ: "disjunction", style: "short"},
	{cldr: "listPattern-type-or-narrow", typ: "disjunction", style: "narrow"},
	{cldr: "listPattern-type-unit", typ: "unit", style: "long"},
	{cldr: "listPattern-type-unit-short", typ: "unit", style: "short"},
	{cldr: "listPattern-type-unit-narrow", typ: "unit", style: "narrow"},
}
