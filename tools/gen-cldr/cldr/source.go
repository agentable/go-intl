package cldr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Source struct {
	Root              string
	Versions          Versions
	Available         []string
	LikelySubtags     map[string]string
	Numbers           map[string]Numbers
	Currencies        map[string]Currencies
	CurrencyFractions map[string]CurrencyFraction
	Collations        []string
	LanguageMatching  LanguageMatching
	Regions           map[string][]string
	Dates             map[string]Dates
	Preference        PreferenceData
	Metazones         Metazones
	Units             map[string]Units
	ListPatterns      map[string]ListPatterns
	RelativeTime      map[string]RelativeTimeFields
	DisplayNames      map[string]DisplayNames
}

func LoadAll(ctx context.Context, root string, versions Versions, localeAllowlist []string) (*Source, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	resolved, err := ResolveLocalDir(root)
	if err != nil {
		return nil, fmt.Errorf("resolve cldr-json dir: %w", err)
	}
	if err := CrossCheck(resolved, versions); err != nil {
		return nil, err
	}
	available, err := loadAvailableLocales(resolved)
	if err != nil {
		return nil, err
	}
	available = filterAvailableLocales(available, localeAllowlist)
	likely, err := loadLikelySubtags(resolved)
	if err != nil {
		return nil, err
	}
	numbers, err := loadNumbers(resolved, available)
	if err != nil {
		return nil, err
	}
	currencies, err := loadCurrencies(resolved, available)
	if err != nil {
		return nil, err
	}
	fractions, err := loadCurrencyFractions(resolved)
	if err != nil {
		return nil, err
	}
	collations, err := loadCollations(resolved)
	if err != nil {
		return nil, err
	}
	matching, err := loadLanguageMatching(resolved)
	if err != nil {
		return nil, err
	}
	regions, err := loadRegions(resolved)
	if err != nil {
		return nil, err
	}
	dates, err := loadDates(resolved, available)
	if err != nil {
		return nil, err
	}
	preference, err := loadPreferenceData(resolved)
	if err != nil {
		return nil, err
	}
	metazones, err := loadMetazones(resolved, available)
	if err != nil {
		return nil, err
	}
	units, err := loadUnits(resolved, available)
	if err != nil {
		return nil, err
	}
	listPatterns, err := loadListPatterns(resolved, available)
	if err != nil {
		return nil, err
	}
	relativeTime, err := loadRelativeTimeFields(resolved, available)
	if err != nil {
		return nil, err
	}
	displayNames, err := loadDisplayNames(resolved, available)
	if err != nil {
		return nil, err
	}
	return &Source{Root: resolved, Versions: versions, Available: available, LikelySubtags: likely, Numbers: numbers, Currencies: currencies, CurrencyFractions: fractions, Collations: collations, LanguageMatching: matching, Regions: regions, Dates: dates, Preference: preference, Metazones: metazones, Units: units, ListPatterns: listPatterns, RelativeTime: relativeTime, DisplayNames: displayNames}, nil
}

type availableLocaleList []string

func (l *availableLocaleList) UnmarshalJSON(data []byte) error {
	var flat []string
	if err := json.Unmarshal(data, &flat); err == nil {
		*l = flat
		return nil
	}
	var nested struct {
		Modern []string `json:"modern"`
	}
	if err := json.Unmarshal(data, &nested); err != nil {
		return err
	}
	*l = nested.Modern
	return nil
}

func filterAvailableLocales(available, allowlist []string) []string {
	allowed := make(map[string]bool, len(allowlist))
	for _, locale := range allowlist {
		allowed[locale] = true
	}
	out := []string{"und"}
	for _, locale := range available {
		if allowed[locale] {
			out = append(out, locale)
		}
	}
	return out
}

func loadAvailableLocales(root string) ([]string, error) {
	raw, err := os.ReadFile(filepath.Join(root, "cldr-core", "availableLocales.json"))
	if err != nil {
		return nil, fmt.Errorf("read availableLocales.json: %w", err)
	}
	var doc struct {
		AvailableLocales struct {
			Modern availableLocaleList `json:"modern"`
			Full   availableLocaleList `json:"full"`
		} `json:"availableLocales"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse availableLocales.json: %w", err)
	}
	if len(doc.AvailableLocales.Modern) > 0 {
		return []string(doc.AvailableLocales.Modern), nil
	}
	return []string(doc.AvailableLocales.Full), nil
}

func loadLikelySubtags(root string) (map[string]string, error) {
	raw, err := os.ReadFile(filepath.Join(root, "cldr-core", "supplemental", "likelySubtags.json"))
	if err != nil {
		return nil, fmt.Errorf("read likelySubtags.json: %w", err)
	}
	var doc struct {
		Supplemental struct {
			LikelySubtags map[string]string `json:"likelySubtags"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse likelySubtags.json: %w", err)
	}
	return doc.Supplemental.LikelySubtags, nil
}
