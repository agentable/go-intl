package cldr

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

type LanguageMatching struct {
	ParadigmLocales []string
	MatchVariables  []LanguageMatchVariable
	Rules           []LanguageMatchRule
}

type LanguageMatchVariable struct {
	Name            string
	SourceRegions   []string
	ExpandedRegions []string
}

type LanguageMatchRule struct {
	Desired   string
	Supported string
	Distance  int
	OneWay    bool
}

func loadLanguageMatching(root string) (LanguageMatching, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "languageMatching.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return LanguageMatching{}, err
	}
	var doc struct {
		Supplemental struct {
			LanguageMatching map[string]struct {
				ParadigmLocales struct {
					Locales []string `json:"_locales"`
				} `json:"paradigmLocales"`
				MatchVariables map[string]struct {
					Value string `json:"_value"`
				} `json:"matchVariables"`
				LanguageMatch []struct {
					Desired   string `json:"_desired"`
					Supported string `json:"_supported"`
					Distance  *int   `json:"_distance"`
					OneWay    bool   `json:"_oneway"`
				} `json:"languageMatch"`
			} `json:"languageMatching"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return LanguageMatching{}, fmt.Errorf("parse %s: %w", path, err)
	}
	written, ok := doc.Supplemental.LanguageMatching["written-new"]
	if !ok {
		return LanguageMatching{}, fmt.Errorf("%s: missing supplemental.languageMatching.written-new", path)
	}

	containment, err := loadTerritoryContainment(root)
	if err != nil {
		return LanguageMatching{}, err
	}
	out := LanguageMatching{
		ParadigmLocales: slices.Clone(written.ParadigmLocales.Locales),
		MatchVariables:  make([]LanguageMatchVariable, 0, len(written.MatchVariables)),
		Rules:           make([]LanguageMatchRule, len(written.LanguageMatch)),
	}
	seenParadigms := make(map[string]bool, len(out.ParadigmLocales))
	for _, paradigm := range out.ParadigmLocales {
		if !validLocaleIdentifier(paradigm) {
			return LanguageMatching{}, fmt.Errorf("%s: invalid paradigm locale %q", path, paradigm)
		}
		if seenParadigms[paradigm] {
			return LanguageMatching{}, fmt.Errorf("%s: duplicate paradigm locale %q", path, paradigm)
		}
		seenParadigms[paradigm] = true
	}

	variableNames := slices.Sorted(maps.Keys(written.MatchVariables))
	knownVariables := make(map[string]bool, len(variableNames))
	for _, sourceName := range variableNames {
		name, ok := strings.CutPrefix(sourceName, "$")
		if !ok || name == "" {
			return LanguageMatching{}, fmt.Errorf("%s: invalid match variable %q", path, sourceName)
		}
		regions := strings.Split(written.MatchVariables[sourceName].Value, "+")
		if len(regions) == 1 && regions[0] == "" {
			return LanguageMatching{}, fmt.Errorf("%s: match variable %q has empty region list", path, sourceName)
		}
		expanded, err := expandTerritorySet(regions, containment)
		if err != nil {
			return LanguageMatching{}, fmt.Errorf("%s: match variable %q: %w", path, sourceName, err)
		}
		out.MatchVariables = append(out.MatchVariables, LanguageMatchVariable{
			Name:            name,
			SourceRegions:   slices.Clone(regions),
			ExpandedRegions: expanded,
		})
		knownVariables[name] = true
	}

	if len(written.LanguageMatch) == 0 {
		return LanguageMatching{}, fmt.Errorf("%s: languageMatch is empty", path)
	}
	for i, source := range written.LanguageMatch {
		if source.Desired == "" {
			return LanguageMatching{}, fmt.Errorf("%s: languageMatch[%d] missing desired pattern", path, i)
		}
		if source.Supported == "" {
			return LanguageMatching{}, fmt.Errorf("%s: languageMatch[%d] missing supported pattern", path, i)
		}
		if err := validateLanguageMatchPattern(source.Desired, knownVariables); err != nil {
			return LanguageMatching{}, fmt.Errorf("%s: languageMatch[%d] desired: %w", path, i, err)
		}
		if err := validateLanguageMatchPattern(source.Supported, knownVariables); err != nil {
			return LanguageMatching{}, fmt.Errorf("%s: languageMatch[%d] supported: %w", path, i, err)
		}
		if source.Distance == nil || *source.Distance < 0 || *source.Distance > 100 {
			return LanguageMatching{}, fmt.Errorf("%s: languageMatch[%d] invalid distance", path, i)
		}
		out.Rules[i] = LanguageMatchRule{
			Desired:   source.Desired,
			Supported: source.Supported,
			Distance:  *source.Distance,
			OneWay:    source.OneWay,
		}
	}
	last := out.Rules[len(out.Rules)-1]
	if last.Desired != "*-*-*" || last.Supported != "*-*-*" {
		return LanguageMatching{}, fmt.Errorf("%s: languageMatch final catch-all rule is missing", path)
	}
	return out, nil
}

type territoryContainmentRecord struct {
	Contains []string `json:"_contains"`
}

func loadTerritoryContainment(root string) (map[string][]string, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "territoryContainment.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			TerritoryContainment map[string]territoryContainmentRecord `json:"territoryContainment"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(doc.Supplemental.TerritoryContainment) == 0 {
		return nil, fmt.Errorf("%s: territoryContainment is empty", path)
	}
	out := make(map[string][]string, len(doc.Supplemental.TerritoryContainment))
	for region, record := range doc.Supplemental.TerritoryContainment {
		out[region] = slices.Clone(record.Contains)
	}
	for region, children := range maps.Clone(out) {
		base, grouping := strings.CutSuffix(region, "-status-grouping")
		if !grouping {
			continue
		}
		out[base] = append(out[base], children...)
		delete(out, region)
	}
	return out, nil
}

func expandTerritorySet(regions []string, containment map[string][]string) ([]string, error) {
	seen := make(map[string]bool)
	visiting := make(map[string]bool)
	var expand func(string) error
	expand = func(region string) error {
		if visiting[region] {
			return fmt.Errorf("territory containment cycle at %q", region)
		}
		children, group := containment[region]
		seen[region] = true
		if !group || len(children) == 0 {
			if !validRegion(region) {
				return fmt.Errorf("invalid region %q", region)
			}
			return nil
		}
		visiting[region] = true
		for _, child := range children {
			if err := expand(child); err != nil {
				return err
			}
		}
		delete(visiting, region)
		return nil
	}
	for _, region := range regions {
		if err := expand(region); err != nil {
			return nil, err
		}
	}
	return slices.Sorted(maps.Keys(seen)), nil
}

func validateLanguageMatchPattern(pattern string, variables map[string]bool) error {
	parts := strings.Split(pattern, "-")
	if len(parts) < 1 || len(parts) > 3 {
		return fmt.Errorf("invalid pattern %q", pattern)
	}
	if parts[0] != "*" && !validLanguage(parts[0]) {
		return fmt.Errorf("invalid language in pattern %q", pattern)
	}
	if len(parts) >= 2 && parts[1] != "*" && !validScript(parts[1]) {
		return fmt.Errorf("invalid script in pattern %q", pattern)
	}
	if len(parts) == 3 {
		region := parts[2]
		if name, variable := matchVariableName(region); variable {
			if !variables[name] {
				return fmt.Errorf("unknown match variable %q", name)
			}
		} else if region != "*" && !validRegion(region) {
			return fmt.Errorf("invalid region in pattern %q", pattern)
		}
	}
	return nil
}

func matchVariableName(region string) (string, bool) {
	if !strings.HasPrefix(region, "$") {
		return "", false
	}
	name := strings.TrimPrefix(strings.TrimPrefix(region, "$"), "!")
	return name, name != ""
}

func validLocaleIdentifier(locale string) bool {
	parts := strings.Split(locale, "-")
	if len(parts) < 1 || len(parts) > 3 || !validLanguage(parts[0]) {
		return false
	}
	for _, part := range parts[1:] {
		if !validScript(part) && !validRegion(part) {
			return false
		}
	}
	return true
}

func validLanguage(value string) bool {
	return len(value) >= 2 && len(value) <= 8 && asciiLetters(value)
}

func validScript(value string) bool {
	return len(value) == 4 && asciiLetters(value)
}

func validRegion(value string) bool {
	return len(value) == 2 && asciiLetters(value) || len(value) == 3 && asciiDigitsString(value)
}

func asciiLetters(value string) bool {
	for i := range len(value) {
		if value[i] >= 'A' && value[i] <= 'Z' || value[i] >= 'a' && value[i] <= 'z' {
			continue
		}
		return false
	}
	return true
}

func asciiDigitsString(value string) bool {
	for i := range len(value) {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}
