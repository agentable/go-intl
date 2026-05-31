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
	if l.ext.calendar, err = normalizeUnicodeType(unicodeTypeForKey(ext, "ca")); err != nil {
		return invalidLocaleValue("calendar", unicodeTypeForKey(ext, "ca"), nil)
	}
	if l.ext.collation, err = normalizeUnicodeType(unicodeTypeForKey(ext, "co")); err != nil {
		return invalidLocaleValue("collation", unicodeTypeForKey(ext, "co"), nil)
	}
	if l.ext.numberingSystem, err = normalizeUnicodeType(unicodeTypeForKey(ext, "nu")); err != nil {
		return invalidLocaleValue("numberingSystem", unicodeTypeForKey(ext, "nu"), nil)
	}
	if l.ext.hourCycle, err = normalizeHourCycle(unicodeTypeForKey(ext, "hc")); err != nil {
		return invalidLocaleValue("hourCycle", unicodeTypeForKey(ext, "hc"), nil)
	}
	if l.ext.caseFirst, err = normalizeCaseFirst(unicodeTypeForKey(ext, "kf")); err != nil {
		return invalidLocaleValue("caseFirst", unicodeTypeForKey(ext, "kf"), nil)
	}
	if l.ext.firstDayOfWeek, err = normalizeFirstDayOfWeek(unicodeTypeForKey(ext, "fw")); err != nil {
		return invalidLocaleValue("firstDayOfWeek", unicodeTypeForKey(ext, "fw"), nil)
	}
	numeric, hasNumeric := ext.TypeForKey("kn")
	if hasNumeric {
		l.ext.numericSet = true
		l.ext.numeric, l.ext.numericValue = normalizeNumeric(numeric)
	}
	return nil
}

func unicodeTypeForKey(ext localeid.UnicodeExtension, key string) string {
	value, _ := ext.TypeForKey(key)
	return value
}

func (l Locale) unicodeExtensionKeywords() []localeid.UnicodeKeyword {
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
	out := make([]localeid.UnicodeKeyword, 0, len(keys))
	for _, key := range keys {
		out = append(out, localeid.UnicodeKeyword{Key: key, Value: keywords[key]})
	}
	return out
}

func unknownUnicodeKeywords(keywords []localeid.UnicodeKeyword) map[string]string {
	var unknown map[string]string
	for _, keyword := range keywords {
		switch keyword.Key {
		case "ca", "co", "fw", "hc", "kf", "kn", "nu":
			continue
		}
		if unknown == nil {
			unknown = map[string]string{}
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
