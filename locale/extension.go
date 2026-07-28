package locale

import (
	"maps"
	"slices"

	"github.com/agentable/go-intl/internal/localeid"
)

func (l *Locale) readExtensions(ext localeid.UnicodeExtension) error {
	var err error
	l.ext.attributes = ext.Attributes()
	keywords := ext.Keywords()
	l.ext.keywords = unknownUnicodeKeywords(keywords)
	if l.ext.calendar, err = normalizeUnicodeExtension(ext, "ca", "calendar", normalizeUnicodeTypeForKey("ca")); err != nil {
		return err
	}
	if l.ext.collation, err = normalizeUnicodeExtension(ext, "co", "collation", normalizeUnicodeTypeForKey("co")); err != nil {
		return err
	}
	if l.ext.numberingSystem, err = normalizeUnicodeExtension(ext, "nu", "numberingSystem", normalizeUnicodeTypeForKey("nu")); err != nil {
		return err
	}
	if l.ext.hourCycle, err = normalizeUnicodeExtension(ext, "hc", "hourCycle", normalizeHourCycle); err != nil {
		return err
	}
	if l.ext.caseFirst, err = normalizeUnicodeExtension(ext, "kf", "caseFirst", normalizeCaseFirst); err != nil {
		return err
	}
	if l.ext.firstDayOfWeek, err = normalizeUnicodeExtension(ext, "fw", "firstDayOfWeek", normalizeFirstDayOfWeek); err != nil {
		return err
	}
	numeric, hasNumeric := ext.TypeForKey("kn")
	if hasNumeric {
		l.ext.hasNumeric = true
		l.ext.numeric, l.ext.numericValue = normalizeNumeric(numeric)
	}
	return nil
}

func normalizeUnicodeExtension(ext localeid.UnicodeExtension, key, name string, normalize func(string) (string, error)) (string, error) {
	value, _ := ext.TypeForKey(key)
	normalized, err := normalize(value)
	if err != nil {
		return "", invalidLocaleValue(name, value, nil)
	}
	return normalized, nil
}

func (e extensions) empty() bool {
	return e.calendar == "" &&
		e.collation == "" &&
		e.hourCycle == "" &&
		e.caseFirst == "" &&
		!e.hasNumeric &&
		e.numberingSystem == "" &&
		e.firstDayOfWeek == "" &&
		len(e.attributes) == 0 &&
		len(e.keywords) == 0
}

func (l Locale) unicodeExtensionKeywords() []localeid.UnicodeKeyword {
	keywords := make(map[string]string, len(l.ext.keywords)+7)
	maps.Copy(keywords, l.ext.keywords)
	setUnicodeKeyword(keywords, "ca", l.ext.calendar)
	setUnicodeKeyword(keywords, "co", l.ext.collation)
	setUnicodeKeyword(keywords, "fw", l.ext.firstDayOfWeek)
	setUnicodeKeyword(keywords, "hc", l.ext.hourCycle)
	setUnicodeKeyword(keywords, "kf", l.ext.caseFirst)
	if l.ext.hasNumeric {
		keywords["kn"] = l.ext.numericKeyword()
	}
	setUnicodeKeyword(keywords, "nu", l.ext.numberingSystem)
	keys := slices.Sorted(maps.Keys(keywords))
	out := make([]localeid.UnicodeKeyword, len(keys))
	for i, key := range keys {
		out[i] = localeid.UnicodeKeyword{Key: key, Value: keywords[key]}
	}
	return out
}

func setUnicodeKeyword(keywords map[string]string, key, value string) {
	if value != "" {
		keywords[key] = value
	}
}

func (e extensions) numericKeyword() string {
	if e.numeric {
		return ""
	}
	if e.numericValue != "" {
		return e.numericValue
	}
	return "false"
}

func (e *extensions) setNumericOption(value bool) {
	e.hasNumeric = true
	e.numeric = value
	if value {
		e.numericValue = ""
		return
	}
	e.numericValue = "false"
}

func unknownUnicodeKeywords(keywords []localeid.UnicodeKeyword) map[string]string {
	var unknown map[string]string
	for _, keyword := range keywords {
		switch keyword.Key {
		case "ca", "co", "fw", "hc", "kf", "kn", "nu":
			continue
		}
		if unknown == nil {
			unknown = make(map[string]string, len(keywords))
		}
		value := keyword.Value
		if value == "true" {
			value = ""
		}
		unknown[keyword.Key] = value
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
