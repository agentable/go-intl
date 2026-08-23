package cldr

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"path/filepath"
)

const undefinedLocale = "und"

type Source struct {
	Root               string
	Available          []string
	LikelySubtags      map[string]string
	ScriptDirections   map[string]bool
	UnicodeTypeAliases []UnicodeTypeAlias
	LanguageMatching   LanguageMatching
	Numbers            map[string]Numbers
	Currencies         map[string]Currencies
	CurrencyFractions  map[string]CurrencyFraction
	Dates              map[string]Dates
	Preference         PreferenceData
	Metazones          Metazones
	Units              map[string]Units
	ListPatterns       map[string]ListPatterns
	RelativeTime       map[string]RelativeTimeFields
	DisplayNames       map[string]DisplayNames
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
	scriptDirections, err := loadScriptDirections(resolved)
	if err != nil {
		return nil, err
	}
	unicodeTypeAliases, err := loadUnicodeTypeAliases(resolved)
	if err != nil {
		return nil, err
	}
	languageMatching, err := loadLanguageMatching(resolved)
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
	return &Source{
		Root:               resolved,
		Available:          available,
		LikelySubtags:      likely,
		ScriptDirections:   scriptDirections,
		UnicodeTypeAliases: unicodeTypeAliases,
		LanguageMatching:   languageMatching,
		Numbers:            numbers,
		Currencies:         currencies,
		CurrencyFractions:  fractions,
		Dates:              dates,
		Preference:         preference,
		Metazones:          metazones,
		Units:              units,
		ListPatterns:       listPatterns,
		RelativeTime:       relativeTime,
		DisplayNames:       displayNames,
	}, nil
}

type availableLocaleList []string

func (l *availableLocaleList) UnmarshalJSON(data []byte) error {
	var flat []string
	if err := json.Unmarshal(data, &flat); err == nil && flat != nil {
		*l = flat
		return nil
	}
	var nested struct {
		Modern []string `json:"modern"`
	}
	if err := json.Unmarshal(data, &nested); err != nil {
		return err
	}
	if nested.Modern == nil {
		return fmt.Errorf("expected locale array or nested modern list")
	}
	*l = nested.Modern
	return nil
}

func filterAvailableLocales(available, allowlist []string) []string {
	filter := len(allowlist) > 0
	var allowed map[string]bool
	if filter {
		allowed = make(map[string]bool, len(allowlist))
		for _, locale := range allowlist {
			if locale == "" || locale == undefinedLocale {
				continue
			}
			allowed[locale] = true
		}
	}
	seen := map[string]bool{undefinedLocale: true}
	out := []string{undefinedLocale}
	for _, locale := range available {
		if locale == "" || locale == undefinedLocale || seen[locale] {
			continue
		}
		if filter && !allowed[locale] {
			continue
		}
		out = append(out, locale)
		seen[locale] = true
	}
	return out
}

func loadAvailableLocales(root string) ([]string, error) {
	path := filepath.Join(root, "cldr-core", "availableLocales.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
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
	if len(doc.AvailableLocales.Full) > 0 {
		return []string(doc.AvailableLocales.Full), nil
	}
	return nil, fmt.Errorf("expected availableLocales modern or full locale list")
}

func loadLikelySubtags(root string) (map[string]string, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "likelySubtags.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			LikelySubtags map[string]string `json:"likelySubtags"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse likelySubtags.json: %w", err)
	}
	if len(doc.Supplemental.LikelySubtags) == 0 {
		return nil, fmt.Errorf("expected supplemental likelySubtags map")
	}
	return doc.Supplemental.LikelySubtags, nil
}
