package locale

import (
	"errors"
	"regexp"
	"strings"

	"github.com/agentable/go-intl/internal/intlerr"
)

var unicodeTypePattern = regexp.MustCompile(`^[a-z0-9]{3,8}(-[a-z0-9]{3,8})*$`)

func (l *Locale) validate() error {
	var err error
	if l.ext.calendar, err = normalizeOption("calendar", l.ext.calendar, normalizeUnicodeType); err != nil {
		return err
	}
	if l.ext.collation, err = normalizeOption("collation", l.ext.collation, normalizeUnicodeType); err != nil {
		return err
	}
	if l.ext.numberingSystem, err = normalizeOption("numberingSystem", l.ext.numberingSystem, normalizeUnicodeType); err != nil {
		return err
	}
	if l.ext.hourCycle, err = normalizeOption("hourCycle", l.ext.hourCycle, normalizeHourCycle); err != nil {
		return err
	}
	if l.ext.caseFirst, err = normalizeOption("caseFirst", l.ext.caseFirst, normalizeCaseFirst); err != nil {
		return err
	}
	if l.ext.firstDayOfWeek, err = normalizeOption("firstDayOfWeek", l.ext.firstDayOfWeek, normalizeFirstDayOfWeek); err != nil {
		return err
	}
	return nil
}

func normalizeOption(name, value string, normalize func(string) (string, error)) (string, error) {
	normalized, err := normalize(value)
	if err != nil {
		return "", invalidLocaleOption(name, value, nil)
	}
	return normalized, nil
}

func invalidLocaleOption(name, value string, err error) error {
	return intlerr.New(intlerr.InvalidOption, "locale", name, value, "", localeErrorCause(intlerr.ErrInvalidOption, err))
}

func invalidLocaleValue(name, value string, err error) error {
	return intlerr.New(intlerr.InvalidValue, "locale", name, value, "", localeErrorCause(intlerr.ErrInvalidValue, err))
}

func localeErrorCause(sentinel, err error) error {
	if err == nil || errors.Is(err, sentinel) {
		return sentinel
	}
	return errors.Join(sentinel, err)
}

func normalizeCalendarAliases(tag string) string {
	tag = strings.ReplaceAll(tag, "-ca-gregorian", "-ca-gregory")
	tag = strings.ReplaceAll(tag, "-ca-islamic-civil", "-ca-islamicc")
	return tag
}

func normalizeFirstDayAliases(tag string) string {
	replacements := []struct {
		old string
		new string
	}{
		{"-fw-0", "-fw-sun"},
		{"-fw-1", "-fw-mon"},
		{"-fw-2", "-fw-tue"},
		{"-fw-3", "-fw-wed"},
		{"-fw-4", "-fw-thu"},
		{"-fw-5", "-fw-fri"},
		{"-fw-6", "-fw-sat"},
		{"-fw-7", "-fw-sun"},
	}
	for _, repl := range replacements {
		tag = strings.ReplaceAll(tag, repl.old, repl.new)
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
	value = strings.ToLower(value)
	switch value {
	case "gregorian":
		value = "gregory"
	case "islamic-civil":
		value = "islamicc"
	}
	if !unicodeTypePattern.MatchString(value) {
		return "", intlerr.ErrInvalidValue
	}
	return value, nil
}

func normalizeHourCycle(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	value = strings.ToLower(value)
	switch value {
	case "h11", "h12", "h23", "h24":
		return value, nil
	}
	return "", intlerr.ErrInvalidValue
}

func normalizeCaseFirst(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	value = strings.ToLower(value)
	switch value {
	case "upper", "lower", "false":
		return value, nil
	}
	return "", intlerr.ErrInvalidValue
}

func normalizeFirstDayOfWeek(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	value = strings.ToLower(value)
	switch value {
	case "sun", "0", "7":
		return "sun", nil
	case "mon", "1":
		return "mon", nil
	case "tue", "2":
		return "tue", nil
	case "wed", "3":
		return "wed", nil
	case "thu", "4":
		return "thu", nil
	case "fri", "5":
		return "fri", nil
	case "sat", "6":
		return "sat", nil
	}
	return "", intlerr.ErrInvalidValue
}
