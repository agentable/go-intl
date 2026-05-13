package locale

import (
	"fmt"
	"regexp"
	"strings"
)

var unicodeTypePattern = regexp.MustCompile(`^[a-z0-9]{3,8}(-[a-z0-9]{3,8})*$`)

func (l *Locale) validate() error {
	var err error
	if l.ext.calendar, err = normalizeUnicodeType(l.ext.calendar); err != nil {
		return fmt.Errorf("locale: invalid calendar %q: %w", l.ext.calendar, ErrInvalidLocale)
	}
	if l.ext.collation, err = normalizeUnicodeType(l.ext.collation); err != nil {
		return fmt.Errorf("locale: invalid collation %q: %w", l.ext.collation, ErrInvalidLocale)
	}
	if l.ext.numberingSystem, err = normalizeUnicodeType(l.ext.numberingSystem); err != nil {
		return fmt.Errorf("locale: invalid numberingSystem %q: %w", l.ext.numberingSystem, ErrInvalidLocale)
	}
	if l.ext.hourCycle, err = normalizeHourCycle(l.ext.hourCycle); err != nil {
		return fmt.Errorf("locale: invalid hourCycle %q: %w", l.ext.hourCycle, ErrInvalidLocale)
	}
	if l.ext.caseFirst, err = normalizeCaseFirst(l.ext.caseFirst); err != nil {
		return fmt.Errorf("locale: invalid caseFirst %q: %w", l.ext.caseFirst, ErrInvalidLocale)
	}
	if l.ext.firstDayOfWeek, err = normalizeFirstDayOfWeek(l.ext.firstDayOfWeek); err != nil {
		return fmt.Errorf("locale: invalid firstDayOfWeek %q: %w", l.ext.firstDayOfWeek, ErrInvalidLocale)
	}
	return nil
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
		return "", ErrInvalidLocale
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
	return "", ErrInvalidLocale
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
	return "", ErrInvalidLocale
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
	return "", ErrInvalidLocale
}
