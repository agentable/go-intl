package cldr

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
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
	PerUnit  map[string]string
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
				Units map[string]jsontext.Value `json:"units"`
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
			compoundPattern, err := unitCompoundPattern(widthUnits)
			if err != nil {
				return nil, fmt.Errorf("parse %s %s units: %w", path, width, err)
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
				if perUnit, ok := fields["perUnitPattern"]; ok {
					if err := validatePatternPlaceholders("perUnitPattern", perUnit, "{0}"); err != nil {
						return nil, fmt.Errorf("parse %s %s: %w", width, key, err)
					}
					if data.PerUnit == nil {
						data.PerUnit = make(map[string]string)
					}
					data.PerUnit[width] = perUnit
				}
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

func parseUnitWidth(raw jsontext.Value) (map[string]map[string]string, error) {
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

func unitCompoundPattern(widthUnits map[string]map[string]string) (string, error) {
	fields := widthUnits["per"]
	pattern := fields["compoundUnitPattern"]
	if pattern == "" {
		return "", fmt.Errorf("compoundUnitPattern missing")
	}
	if err := validatePatternPlaceholders("compoundUnitPattern", pattern, "{0}", "{1}"); err != nil {
		return "", err
	}
	return pattern, nil
}

func validatePatternPlaceholders(name, pattern string, placeholders ...string) error {
	for _, placeholder := range placeholders {
		if count := strings.Count(pattern, placeholder); count != 1 {
			return fmt.Errorf("%s expected exactly one %s placeholder, got %d in %q", name, placeholder, count, pattern)
		}
	}
	return nil
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
	_, unit, ok := strings.Cut(key, "-")
	if !ok {
		return "", false
	}
	if unitLoaderSanctionedKeys[key] {
		return unit, true
	}
	if strings.Contains(unit, "-per-") && unitid.IsWellFormedUnitIdentifier(unit) {
		return unit, true
	}
	return "", false
}
