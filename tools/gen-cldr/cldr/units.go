package cldr

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/agentable/go-intl/internal/unitid"
)

type Units map[string]UnitData

type UnitData struct {
	Patterns map[string]map[string]map[string]string
	Compound map[string]string
}

var (
	unitLoaderWidths         = [...]string{"long", "short", "narrow"}
	unitLoaderSanctionedKeys = sanctionedUnitKeySet()
)

func loadUnits(root string, locales []string) (map[string]Units, error) {
	out := make(map[string]Units)
	for _, locale := range locales {
		if locale == undefinedLocale {
			continue
		}
		path := filepath.Join(root, "cldr-units-full", "main", locale, "units.json")
		raw, ok, err := readOptionalFile(path)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		var doc struct {
			Main map[string]struct {
				Units map[string]json.RawMessage `json:"units"`
			} `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if doc.Main == nil {
			return nil, fmt.Errorf("units body missing for %s", locale)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, fmt.Errorf("units body missing for %s", locale)
		}
		if len(body.Units) == 0 {
			return nil, fmt.Errorf("units data missing for %s", locale)
		}
		units := make(Units)
		compound := make(map[string]string)
		for _, width := range unitLoaderWidths {
			widthUnits, err := parseUnitWidth(body.Units[width])
			if err != nil {
				return nil, fmt.Errorf("parse %s %s units: %w", path, width, err)
			}
			compoundPattern, ok := unitCompoundPattern(widthUnits)
			if !ok {
				return nil, fmt.Errorf("parse %s %s units: compoundUnitPattern missing", path, width)
			}
			compound[width] = compoundPattern
			for _, key := range slices.Sorted(maps.Keys(widthUnits)) {
				fields := widthUnits[key]
				if key == "per" {
					continue
				}
				unit, ok := unitIdentifierFromCLDRKey(key)
				if !ok {
					continue
				}
				patterns, err := unitPluralPatterns(fields)
				if err != nil {
					return nil, fmt.Errorf("parse %s %s: %w", width, key, err)
				}
				data := units[unit]
				if data.Patterns == nil {
					data.Patterns = make(map[string]map[string]map[string]string)
				}
				if data.Patterns[width] == nil {
					data.Patterns[width] = make(map[string]map[string]string)
				}
				data.Patterns[width][unit] = patterns
				units[unit] = data
			}
		}
		for _, unit := range slices.Sorted(maps.Keys(units)) {
			data := units[unit]
			data.Compound = compound
			units[unit] = data
		}
		out[locale] = units
	}
	return out, nil
}

func parseUnitWidth(raw json.RawMessage) (map[string]map[string]string, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("missing unit width data")
	}
	var widthUnits map[string]map[string]string
	if err := json.Unmarshal(raw, &widthUnits); err != nil {
		return nil, err
	}
	if len(widthUnits) == 0 {
		return nil, fmt.Errorf("empty unit width data")
	}
	return widthUnits, nil
}

func unitCompoundPattern(widthUnits map[string]map[string]string) (string, bool) {
	fields := widthUnits["per"]
	pattern := fields["compoundUnitPattern"]
	return pattern, pattern != ""
}

func unitPluralPatterns(fields map[string]string) (map[string]string, error) {
	patterns := make(map[string]string)
	for _, field := range slices.Sorted(maps.Keys(fields)) {
		value := fields[field]
		plural, ok, err := pluralCategoryFromField(field, "unitPattern-count-")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field, err)
		}
		if !ok {
			continue
		}
		if value == "" {
			return nil, fmt.Errorf("%s empty", field)
		}
		patterns[plural] = value
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("unitPattern-count data missing")
	}
	return patterns, nil
}

func sanctionedUnitKeySet() map[string]bool {
	keys := unitid.SanctionedUnitIdentifiers()
	out := make(map[string]bool, len(keys))
	for _, key := range keys {
		out[key] = true
	}
	return out
}

func unitIdentifierFromCLDRKey(key string) (string, bool) {
	if !unitLoaderSanctionedKeys[key] {
		return "", false
	}
	_, unit, ok := strings.Cut(key, "-")
	if !ok {
		return "", false
	}
	return unit, true
}
