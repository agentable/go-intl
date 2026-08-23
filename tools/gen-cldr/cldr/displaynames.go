package cldr

import (
	"encoding/json/v2"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type DisplayNames struct {
	Languages      LanguageDisplay
	Territories    StyledNames
	Scripts        StyledNames
	Calendars      StyledNames
	DateTimeFields StyledNames
	LocalePattern  string
}

type LanguageDisplay struct {
	Dialect  StyledNames
	Standard StyledNames
}

type StyledNames struct {
	Long   map[string]string
	Short  map[string]string
	Narrow map[string]string
}

func loadDisplayNames(root string, locales []string) (map[string]DisplayNames, error) {
	loaded := make(map[string]DisplayNames)
	for _, locale := range locales {
		if locale == undefinedLocale {
			continue
		}
		data, ok, err := readDisplayNamesLocale(root, locale)
		if err != nil {
			return nil, err
		}
		if ok {
			loaded[locale] = data
		}
	}
	return inheritedLocaleData(locales, loaded), nil
}

func readDisplayNamesLocale(root, locale string) (DisplayNames, bool, error) {
	languages, hasLanguages, err := readLocaleNamesFile(root, locale, "languages.json", "languages")
	if err != nil {
		return DisplayNames{}, false, err
	}
	territories, _, err := readLocaleNamesFile(root, locale, "territories.json", "territories")
	if err != nil {
		return DisplayNames{}, false, err
	}
	scripts, _, err := readLocaleNamesFile(root, locale, "scripts.json", "scripts")
	if err != nil {
		return DisplayNames{}, false, err
	}
	calendars, pattern, err := readLocaleDisplayNamesFile(root, locale)
	if err != nil {
		return DisplayNames{}, false, err
	}
	dateTimeFields, err := readDateTimeFieldNames(root, locale)
	if err != nil {
		return DisplayNames{}, false, err
	}
	if !hasLanguages {
		return DisplayNames{}, false, nil
	}
	data := DisplayNames{
		Languages: LanguageDisplay{
			Dialect:  splitStyleData(languages),
			Standard: buildStandardLanguageNames(languages, territories, pattern),
		},
		Territories:    splitStyleData(territories),
		Scripts:        splitStyleData(scripts),
		Calendars:      splitStyleData(calendars),
		DateTimeFields: splitDateTimeFieldStyleData(dateTimeFields),
		LocalePattern:  pattern,
	}
	return data, true, nil
}

func readLocaleNamesFile(root, locale, file, field string) (map[string]string, bool, error) {
	path := filepath.Join(root, "cldr-localenames-full", "main", locale, file)
	raw, ok, err := readOptionalFile(path)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	var doc struct {
		Main map[string]struct {
			LocaleDisplayNames map[string]map[string]string `json:"localeDisplayNames"`
		} `json:"main"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, false, fmt.Errorf("parse %s: %w", path, err)
	}
	if doc.Main == nil {
		return nil, false, fmt.Errorf("%s body missing for %s", file, locale)
	}
	body, ok := doc.Main[locale]
	if !ok {
		return nil, false, fmt.Errorf("%s body missing for %s", file, locale)
	}
	if body.LocaleDisplayNames == nil {
		return nil, false, fmt.Errorf("%s localeDisplayNames missing for %s", file, locale)
	}
	values := body.LocaleDisplayNames[field]
	if len(values) == 0 {
		return nil, false, fmt.Errorf("%s %s data missing for %s", file, field, locale)
	}
	return values, true, nil
}

func readLocaleDisplayNamesFile(root, locale string) (calendars map[string]string, pattern string, err error) {
	path := filepath.Join(root, "cldr-localenames-full", "main", locale, "localeDisplayNames.json")
	raw, ok, err := readOptionalFile(path)
	if err != nil {
		return nil, "", err
	}
	if !ok {
		return nil, "", nil
	}
	var doc struct {
		Main map[string]struct {
			LocaleDisplayNames *struct {
				LocaleDisplayPattern *struct {
					LocalePattern string `json:"localePattern"`
				} `json:"localeDisplayPattern"`
				Types *struct {
					Calendar map[string]string `json:"calendar"`
				} `json:"types"`
			} `json:"localeDisplayNames"`
		} `json:"main"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", path, err)
	}
	if doc.Main == nil {
		return nil, "", fmt.Errorf("localeDisplayNames body missing for %s", locale)
	}
	body, ok := doc.Main[locale]
	if !ok {
		return nil, "", fmt.Errorf("localeDisplayNames body missing for %s", locale)
	}
	data := body.LocaleDisplayNames
	if data == nil {
		return nil, "", fmt.Errorf("localeDisplayNames data missing for %s", locale)
	}
	if data.LocaleDisplayPattern == nil || data.LocaleDisplayPattern.LocalePattern == "" {
		return nil, "", fmt.Errorf("locale display pattern missing for %s", locale)
	}
	if data.Types == nil || len(data.Types.Calendar) == 0 {
		return nil, "", fmt.Errorf("calendar display-name data missing for %s", locale)
	}
	return data.Types.Calendar, data.LocaleDisplayPattern.LocalePattern, nil
}

func readDateTimeFieldNames(root, locale string) (map[string]string, error) {
	path := filepath.Join(root, "cldr-dates-full", "main", locale, "dateFields.json")
	raw, ok, err := readOptionalFile(path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	var doc struct {
		Main map[string]struct {
			Dates struct {
				Fields map[string]struct {
					DisplayName string `json:"displayName"`
				} `json:"fields"`
			} `json:"dates"`
		} `json:"main"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if doc.Main == nil {
		return nil, fmt.Errorf("dateFields body missing for %s", locale)
	}
	body, ok := doc.Main[locale]
	if !ok {
		return nil, fmt.Errorf("dateFields body missing for %s", locale)
	}
	if len(body.Dates.Fields) == 0 {
		return nil, fmt.Errorf("dateFields data missing for %s", locale)
	}
	out := make(map[string]string, len(body.Dates.Fields))
	for key, value := range body.Dates.Fields {
		if value.DisplayName != "" {
			out[key] = value.DisplayName
		}
	}
	return out, nil
}

func splitStyleData(raw map[string]string) StyledNames {
	out := newStyledNames()
	for key, value := range raw {
		if value == "" {
			continue
		}
		name, style, _, ok := displayNameStyleKey(key)
		if !ok {
			continue
		}
		putStyledName(out, style, name, value)
	}
	return compactStyledNames(out)
}

func splitDateTimeFieldStyleData(raw map[string]string) StyledNames {
	out := newStyledNames()
	for key, value := range raw {
		if value == "" {
			continue
		}
		base, suffix, hasSuffix := strings.Cut(key, "-")
		field := dateTimeFieldLookupCode(base)
		style := displayNameStyleLong
		if hasSuffix {
			if !isDisplayNameAltStyle(suffix) {
				continue
			}
			style = suffix
		}
		putStyledName(out, style, field, value)
	}
	return compactStyledNames(out)
}

func dateTimeFieldLookupCode(code string) string {
	switch code {
	case "week":
		return "weekOfYear"
	case "zone":
		return "timeZoneName"
	default:
		return code
	}
}

var regionSuffixRe = regexp.MustCompile(`-([A-Za-z]{2}|\d{3})\b`)

func buildStandardLanguageNames(languages, territories map[string]string, pattern string) StyledNames {
	if pattern == "" {
		return splitStyleData(languages)
	}
	out := newStyledNames()
	for key, value := range languages {
		if value == "" {
			continue
		}
		base, style, territorySuffix, ok := displayNameStyleKey(key)
		if !ok {
			continue
		}
		putStyledName(out, style, base, standardLanguageValue(base, value, languages, territories, territorySuffix, pattern))
	}
	return compactStyledNames(out)
}

const (
	displayNameStyleLong   = "long"
	displayNameStyleShort  = "short"
	displayNameStyleNarrow = "narrow"
)

func isDisplayNameAltStyle(style string) bool {
	return style == displayNameStyleShort || style == displayNameStyleNarrow
}

func displayNameStyleKey(key string) (base, style, altSuffix string, ok bool) {
	base = key
	style = displayNameStyleLong
	altBase, alt, isAlt := strings.Cut(key, "-alt-")
	if !isAlt {
		return base, style, "", true
	}
	if !isDisplayNameAltStyle(alt) {
		return "", "", "", false
	}
	return altBase, alt, "-alt-" + alt, true
}

func putStyledName(out StyledNames, style, key, value string) {
	switch style {
	case displayNameStyleLong:
		out.Long[key] = value
	case displayNameStyleShort:
		out.Short[key] = value
	case displayNameStyleNarrow:
		out.Narrow[key] = value
	}
}

func newStyledNames() StyledNames {
	return StyledNames{
		Long:   make(map[string]string),
		Short:  make(map[string]string),
		Narrow: make(map[string]string),
	}
}

func compactStyledNames(out StyledNames) StyledNames {
	if len(out.Long) == 0 {
		out.Long = nil
	}
	if len(out.Short) == 0 {
		out.Short = nil
	}
	if len(out.Narrow) == 0 {
		out.Narrow = nil
	}
	return out
}

func standardLanguageValue(tag, dialectValue string, languages, territories map[string]string, territorySuffix, pattern string) string {
	match := regionSuffixRe.FindStringIndex(tag)
	if match == nil {
		return dialectValue
	}
	region := tag[match[0]+1 : match[1]]
	languageSubtag := tag[:match[0]] + tag[match[1]:]
	languageName := languages[languageSubtag]
	if languageName == "" {
		return dialectValue
	}
	regionName := territories[region+territorySuffix]
	if regionName == "" {
		regionName = territories[region]
	}
	if regionName == "" {
		regionName = region
	}
	result := strings.Replace(pattern, "{0}", languageName, 1)
	result = strings.Replace(result, "{1}", regionName, 1)
	return result
}
