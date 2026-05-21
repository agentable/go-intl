package locale

import (
	"maps"
	"slices"
	"strings"

	"github.com/agentable/go-intl/internal/intlerr"
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
		return invalidLocaleValue("calendar", keywords["ca"], nil)
	}
	if l.ext.collation, err = normalizeUnicodeType(keywords["co"]); err != nil {
		return invalidLocaleValue("collation", keywords["co"], nil)
	}
	if l.ext.numberingSystem, err = normalizeUnicodeType(keywords["nu"]); err != nil {
		return invalidLocaleValue("numberingSystem", keywords["nu"], nil)
	}
	if l.ext.hourCycle, err = normalizeHourCycle(keywords["hc"]); err != nil {
		return invalidLocaleValue("hourCycle", keywords["hc"], nil)
	}
	if l.ext.caseFirst, err = normalizeCaseFirst(keywords["kf"]); err != nil {
		return invalidLocaleValue("caseFirst", keywords["kf"], nil)
	}
	if l.ext.firstDayOfWeek, err = normalizeFirstDayOfWeek(keywords["fw"]); err != nil {
		return invalidLocaleValue("firstDayOfWeek", keywords["fw"], nil)
	}
	numeric, hasNumeric := keywords["kn"]
	if hasNumeric {
		l.ext.numericSet = true
		l.ext.numeric, l.ext.numericValue = normalizeNumeric(numeric)
	}
	return nil
}

func (l Locale) unicodeExtensionParts() []string {
	parts := slices.Clone(l.ext.attributes)
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
	maps.Copy(keywords, l.ext.keywords)
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
		baseParts := slices.Clone(parts[:i])
		baseParts = append(baseParts, parts[end:]...)
		if len(baseParts) == 0 {
			return "", unicodeExtension{}, intlerr.ErrInvalidValue
		}
		return strings.Join(baseParts, "-"), ext, nil
	}
	return tag, unicodeExtension{}, nil
}

func parseUnicodeExtension(parts []string) (unicodeExtension, error) {
	if len(parts) == 0 {
		return unicodeExtension{}, intlerr.ErrInvalidValue
	}
	var ext unicodeExtension
	i := 0
	for i < len(parts) && len(parts[i]) >= 3 {
		if !isUnicodeExtensionSubtag(parts[i]) {
			return unicodeExtension{}, intlerr.ErrInvalidValue
		}
		ext.attributes = append(ext.attributes, parts[i])
		i++
	}
	slices.Sort(ext.attributes)
	for i < len(parts) {
		key := parts[i]
		if len(key) != 2 || !asciiAlnum(key) {
			return unicodeExtension{}, intlerr.ErrInvalidValue
		}
		i++
		start := i
		for i < len(parts) && len(parts[i]) >= 3 {
			if !isUnicodeExtensionSubtag(parts[i]) {
				return unicodeExtension{}, intlerr.ErrInvalidValue
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
