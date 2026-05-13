package locale

import (
	"fmt"
	"slices"
	"strings"
)

type unicodeExtension struct {
	attributes []string
	keywords   map[string]string
}

func (l *Locale) readExtensions(ext unicodeExtension) error {
	var err error
	l.ext.attributes = ext.attributes
	l.ext.keywords = unknownUnicodeKeywords(ext.keywords)
	keywords := ext.keywords
	if l.ext.calendar, err = normalizeUnicodeType(keywords["ca"]); err != nil {
		return fmt.Errorf("locale: invalid calendar %q: %w", keywords["ca"], ErrInvalidLocale)
	}
	if l.ext.collation, err = normalizeUnicodeType(keywords["co"]); err != nil {
		return fmt.Errorf("locale: invalid collation %q: %w", keywords["co"], ErrInvalidLocale)
	}
	if l.ext.numberingSystem, err = normalizeUnicodeType(keywords["nu"]); err != nil {
		return fmt.Errorf("locale: invalid numberingSystem %q: %w", keywords["nu"], ErrInvalidLocale)
	}
	if l.ext.hourCycle, err = normalizeHourCycle(keywords["hc"]); err != nil {
		return fmt.Errorf("locale: invalid hourCycle %q: %w", keywords["hc"], ErrInvalidLocale)
	}
	if l.ext.caseFirst, err = normalizeCaseFirst(keywords["kf"]); err != nil {
		return fmt.Errorf("locale: invalid caseFirst %q: %w", keywords["kf"], ErrInvalidLocale)
	}
	if l.ext.firstDayOfWeek, err = normalizeFirstDayOfWeek(keywords["fw"]); err != nil {
		return fmt.Errorf("locale: invalid firstDayOfWeek %q: %w", keywords["fw"], ErrInvalidLocale)
	}
	numeric, hasNumeric := keywords["kn"]
	if hasNumeric {
		l.ext.numericSet = true
		l.ext.numeric, l.ext.numericValue = normalizeNumeric(numeric)
	}
	return nil
}

func (l Locale) unicodeExtensionParts() []string {
	parts := append([]string(nil), l.ext.attributes...)
	keywords := make(map[string]string, len(l.ext.keywords)+7)
	appendKeyValue := func(key, value string) {
		if value != "" {
			keywords[key] = value
		}
	}
	appendKeyValue("ca", l.ext.calendar)
	appendKeyValue("co", l.ext.collation)
	appendKeyValue("fw", l.ext.firstDayOfWeek)
	appendKeyValue("hc", l.ext.hourCycle)
	appendKeyValue("kf", l.ext.caseFirst)
	if l.ext.numericSet {
		switch {
		case l.ext.numeric:
			keywords["kn"] = ""
		case l.ext.numericValue != "":
			keywords["kn"] = l.ext.numericValue
		default:
			keywords["kn"] = "false"
		}
	}
	appendKeyValue("nu", l.ext.numberingSystem)
	for key, value := range l.ext.keywords {
		keywords[key] = value
	}
	keys := make([]string, 0, len(keywords))
	for key := range keywords {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		parts = append(parts, key)
		if value := keywords[key]; value != "" {
			parts = append(parts, strings.Split(value, "-")...)
		}
	}
	return parts
}

func splitUnicodeExtension(tag string) (string, unicodeExtension, error) {
	parts := strings.Split(tag, "-")
	for i := range parts {
		if parts[i] != "u" {
			continue
		}
		end := i + 1
		for end < len(parts) && len(parts[end]) != 1 {
			end++
		}
		ext, err := parseUnicodeExtension(parts[i+1 : end])
		if err != nil {
			return "", unicodeExtension{}, err
		}
		baseParts := append([]string(nil), parts[:i]...)
		baseParts = append(baseParts, parts[end:]...)
		if len(baseParts) == 0 {
			return "", unicodeExtension{}, ErrInvalidLocale
		}
		return strings.Join(baseParts, "-"), ext, nil
	}
	return tag, unicodeExtension{}, nil
}

func parseUnicodeExtension(parts []string) (unicodeExtension, error) {
	if len(parts) == 0 {
		return unicodeExtension{}, ErrInvalidLocale
	}
	var ext unicodeExtension
	i := 0
	for i < len(parts) && len(parts[i]) >= 3 {
		if !isUnicodeExtensionSubtag(parts[i]) {
			return unicodeExtension{}, ErrInvalidLocale
		}
		ext.attributes = append(ext.attributes, parts[i])
		i++
	}
	slices.Sort(ext.attributes)
	for i < len(parts) {
		key := parts[i]
		if len(key) != 2 || !asciiAlnum(key) {
			return unicodeExtension{}, ErrInvalidLocale
		}
		i++
		start := i
		for i < len(parts) && len(parts[i]) >= 3 {
			if !isUnicodeExtensionSubtag(parts[i]) {
				return unicodeExtension{}, ErrInvalidLocale
			}
			i++
		}
		if ext.keywords == nil {
			ext.keywords = map[string]string{}
		}
		ext.keywords[key] = strings.Join(parts[start:i], "-")
	}
	return ext, nil
}

func unknownUnicodeKeywords(keywords map[string]string) map[string]string {
	var unknown map[string]string
	for key, value := range keywords {
		switch key {
		case "ca", "co", "fw", "hc", "kf", "kn", "nu":
			continue
		}
		if unknown == nil {
			unknown = map[string]string{}
		}
		if value == "true" {
			value = ""
		}
		unknown[key] = value
	}
	return unknown
}

func normalizeNumeric(value string) (bool, string) {
	switch value {
	case "", "true":
		return true, ""
	case "false":
		return false, "false"
	default:
		return false, value
	}
}

func isUnicodeExtensionSubtag(s string) bool {
	return len(s) >= 3 && len(s) <= 8 && asciiAlnum(s)
}
