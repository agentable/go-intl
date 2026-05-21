package cldr

import (
	"encoding/json"
	"fmt"
	"os"
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
		if locale == "und" {
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
	out := make(map[string]DisplayNames, len(locales))
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		if data, ok := loaded[locale]; ok {
			out[locale] = data
			continue
		}
		if data, ok := inheritedDisplayNames(locale, loaded); ok {
			out[locale] = data
		}
	}
	return out, nil
}

func inheritedDisplayNames(locale string, loaded map[string]DisplayNames) (DisplayNames, bool) {
	for parent := parentLocale(locale); parent != ""; parent = parentLocale(parent) {
		if data, ok := loaded[parent]; ok {
			return data, true
		}
	}
	return DisplayNames{}, false
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
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		Main map[string]struct {
			LocaleDisplayNames map[string]map[string]string `json:"localeDisplayNames"`
		} `json:"main"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, false, fmt.Errorf("parse %s: %w", path, err)
	}
	body, ok := doc.Main[locale]
	if !ok {
		return nil, false, fmt.Errorf("%s body missing for %s", file, locale)
	}
	values := body.LocaleDisplayNames[field]
	if len(values) == 0 {
		return nil, false, nil
	}
	return values, true, nil
}

func readLocaleDisplayNamesFile(root, locale string) (calendars map[string]string, pattern string, err error) {
	path := filepath.Join(root, "cldr-localenames-full", "main", locale, "localeDisplayNames.json")
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("read %s: %w", path, readErr)
	}
	var doc struct {
		Main map[string]struct {
			LocaleDisplayNames struct {
				LocaleDisplayPattern struct {
					LocalePattern string `json:"localePattern"`
				} `json:"localeDisplayPattern"`
				Types struct {
					Calendar map[string]string `json:"calendar"`
				} `json:"types"`
			} `json:"localeDisplayNames"`
		} `json:"main"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", path, err)
	}
	body, ok := doc.Main[locale]
	if !ok {
		return nil, "", fmt.Errorf("localeDisplayNames body missing for %s", locale)
	}
	return body.LocaleDisplayNames.Types.Calendar, body.LocaleDisplayNames.LocaleDisplayPattern.LocalePattern, nil
}

func readDateTimeFieldNames(root, locale string) (map[string]string, error) {
	path := filepath.Join(root, "cldr-dates-full", "main", locale, "dateFields.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
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
	body, ok := doc.Main[locale]
	if !ok {
		return nil, nil
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
	out := StyledNames{
		Long:   make(map[string]string),
		Short:  make(map[string]string),
		Narrow: make(map[string]string),
	}
	for key, value := range raw {
		if value == "" {
			continue
		}
		base, alt, isAlt := strings.Cut(key, "-alt-")
		if !isAlt {
			out.Long[key] = value
			continue
		}
		switch alt {
		case "narrow":
			out.Narrow[base] = value
		case "short":
			out.Short[base] = value
		}
	}
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

var cldrToTC39DateTimeField = map[string]string{
	"week": "weekOfYear",
	"zone": "timeZoneName",
}

func splitDateTimeFieldStyleData(raw map[string]string) StyledNames {
	out := StyledNames{
		Long:   make(map[string]string),
		Short:  make(map[string]string),
		Narrow: make(map[string]string),
	}
	for key, value := range raw {
		if value == "" {
			continue
		}
		base, suffix, hasSuffix := strings.Cut(key, "-")
		field := cldrToTC39DateTimeField[base]
		if field == "" {
			field = base
		}
		switch {
		case !hasSuffix:
			out.Long[field] = value
		case suffix == "narrow":
			out.Narrow[field] = value
		case suffix == "short":
			out.Short[field] = value
		}
	}
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

var regionSuffixRe = regexp.MustCompile(`-([A-Za-z]{2}|\d{3})\b`)

func buildStandardLanguageNames(languages, territories map[string]string, pattern string) StyledNames {
	if pattern == "" {
		return splitStyleData(languages)
	}
	long := make(map[string]string)
	short := make(map[string]string)
	narrow := make(map[string]string)
	for key, value := range languages {
		base, alt, isAlt := strings.Cut(key, "-alt-")
		var target map[string]string
		var territorySuffix string
		switch {
		case !isAlt:
			target = long
		case alt == "short":
			target = short
			territorySuffix = "-alt-short"
		case alt == "narrow":
			target = narrow
			territorySuffix = "-alt-narrow"
		default:
			continue
		}
		target[base] = standardLanguageValue(base, value, languages, territories, territorySuffix, pattern)
	}
	if len(long) == 0 {
		long = nil
	}
	if len(short) == 0 {
		short = nil
	}
	if len(narrow) == 0 {
		narrow = nil
	}
	return StyledNames{Long: long, Short: short, Narrow: narrow}
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
