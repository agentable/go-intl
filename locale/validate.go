package locale

import (
	"errors"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/localeid"
)

var errInvalidLocaleOptionValue = errors.New("invalid locale option value")

const (
	localeUnicodeTypeExpected        = "a well-formed Unicode locale type"
	localeLanguageTagExpected        = "a well-formed BCP 47 language tag"
	localeLanguageIdentifierExpected = "a well-formed BCP 47 language identifier"
	localeLanguageExpected           = "a well-formed BCP 47 language subtag"
	localeScriptExpected             = "a well-formed BCP 47 script subtag"
	localeRegionExpected             = "a well-formed BCP 47 region subtag"
	localeHourCycleExpected          = `one of "h11", "h12", "h23", "h24"`
	localeCaseFirstExpected          = `one of "upper", "lower", "false"`
	localeFirstDayExpected           = "a weekday name or number from 0 through 7"
)

func (l *Locale) validate() error {
	if err := normalizeOption(&l.ext.calendar, "calendar", localeUnicodeTypeExpected, normalizeUnicodeType); err != nil {
		return err
	}
	if err := normalizeOption(&l.ext.collation, "collation", localeUnicodeTypeExpected, normalizeUnicodeType); err != nil {
		return err
	}
	if err := normalizeOption(&l.ext.numberingSystem, "numberingSystem", localeUnicodeTypeExpected, normalizeUnicodeType); err != nil {
		return err
	}
	if err := normalizeOption(&l.ext.hourCycle, "hourCycle", localeHourCycleExpected, normalizeHourCycle); err != nil {
		return err
	}
	if err := normalizeOption(&l.ext.caseFirst, "caseFirst", localeCaseFirstExpected, normalizeCaseFirst); err != nil {
		return err
	}
	return normalizeOption(&l.ext.firstDayOfWeek, "firstDayOfWeek", localeFirstDayExpected, normalizeFirstDayOfWeek)
}

func normalizeOption(dst *string, name, expected string, normalize func(string) (string, error)) error {
	normalized, err := normalize(*dst)
	if err != nil {
		return invalidLocaleOptionExpected(name, *dst, expected, err)
	}
	*dst = normalized
	return nil
}

func invalidLocaleOptionExpected(name, value, expected string, err error) error {
	return intlerr.NewInvalidOptionExpected("locale", name, value, "", expected, err)
}

func invalidLocaleValue(name, value string, err error) error {
	return intlerr.NewInvalidValueExpected("locale", name, value, "", expectedLocaleValue(name), err)
}

func expectedLocaleValue(name string) string {
	switch name {
	case "languageTag":
		return localeLanguageTagExpected
	case "calendar", "collation", "numberingSystem":
		return localeUnicodeTypeExpected
	case "hourCycle":
		return localeHourCycleExpected
	case "caseFirst":
		return localeCaseFirstExpected
	case "firstDayOfWeek":
		return localeFirstDayExpected
	default:
		return "a well-formed locale value"
	}
}

func normalizeLocaleAliases(tag string) string {
	base, privateUse, hasPrivateUse := strings.Cut(tag, "-x-")
	if strings.Contains(base, "-u-") {
		base = normalizeCalendarAliases(base)
		base = normalizeFirstDayAliases(base)
	}
	if !hasPrivateUse {
		return base
	}
	return base + "-x-" + privateUse
}

func normalizeCalendarAliases(tag string) string {
	for _, alias := range [...]string{"gregorian", "islamic-civil"} {
		canonical, ok := localeid.CanonicalUnicodeType(alias)
		if ok && canonical != alias {
			tag = strings.ReplaceAll(tag, "-ca-"+alias, "-ca-"+canonical)
		}
	}
	return tag
}

func normalizeFirstDayAliases(tag string) string {
	for day := range 8 {
		value := strconv.Itoa(day)
		canonical, _ := canonicalFirstDay(value)
		tag = strings.ReplaceAll(tag, "-fw-"+value, "-fw-"+canonical)
	}
	return tag
}

func normalizeLanguageAliases(tag string) string {
	parts := strings.Split(tag, "-")
	if len(parts) == 0 {
		return tag
	}
	if parts[0] == "twi" {
		parts[0] = "ak"
	}
	if parts[0] == "und" && len(parts) >= 3 && parts[1] == "armn" && parts[2] == "su" {
		parts[2] = "am"
	}
	return strings.Join(parts, "-")
}

func normalizeUnicodeType(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	canonical, ok := localeid.CanonicalUnicodeType(value)
	if !ok {
		return "", errInvalidLocaleOptionValue
	}
	return canonical, nil
}

func normalizeHourCycle(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	value, err := normalizeUnicodeType(value)
	if err != nil {
		return "", err
	}
	switch value {
	case "h11", "h12", "h23", "h24":
		return value, nil
	}
	return "", errInvalidLocaleOptionValue
}

func normalizeCaseFirst(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	value, err := normalizeUnicodeType(value)
	if err != nil {
		return "", err
	}
	switch value {
	case "upper", "lower", "false":
		return value, nil
	}
	return "", errInvalidLocaleOptionValue
}

func normalizeFirstDayOfWeek(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if canonical, ok := canonicalFirstDay(value); ok {
		return canonical, nil
	}
	value, err := normalizeUnicodeType(value)
	if err != nil {
		return "", err
	}
	if canonical, ok := canonicalFirstDay(value); ok {
		return canonical, nil
	}
	return "", errInvalidLocaleOptionValue
}

func canonicalFirstDay(value string) (string, bool) {
	switch value {
	case "sun", "0", "7":
		return "sun", true
	case "mon", "1":
		return "mon", true
	case "tue", "2":
		return "tue", true
	case "wed", "3":
		return "wed", true
	case "thu", "4":
		return "thu", true
	case "fri", "5":
		return "fri", true
	case "sat", "6":
		return "sat", true
	default:
		return "", false
	}
}
