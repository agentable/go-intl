package locale

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

type Locale struct {
	_   [0]func()
	tag language.Tag
	ext extensions
}

type extensions struct {
	calendar        string
	collation       string
	hourCycle       string
	caseFirst       string
	numeric         bool
	numericSet      bool
	numericValue    string
	numberingSystem string
	firstDayOfWeek  string
	attributes      []string
	keywords        map[string]string
}

func Parse(s string) (Locale, error) {
	lower := strings.ToLower(s)
	if s == "" || strings.Contains(s, "_") || strings.HasPrefix(lower, "x-") {
		return Locale{}, fmt.Errorf("locale: %q: %w", s, ErrInvalidLocale)
	}
	lower = normalizeCalendarAliases(lower)
	lower = normalizeFirstDayAliases(lower)
	base, unicodeExtension, err := splitUnicodeExtension(lower)
	if err != nil {
		return Locale{}, fmt.Errorf("locale: %q: %w", s, ErrInvalidLocale)
	}
	base = normalizeLanguageAliases(base)
	tag, err := language.Parse(base)
	if err != nil {
		return Locale{}, fmt.Errorf("locale: %q: %w", s, ErrInvalidLocale)
	}
	loc := Locale{tag: tag}
	if err := loc.readExtensions(unicodeExtension); err != nil {
		return Locale{}, err
	}
	return loc, nil
}

func MustParse(s string) Locale {
	loc, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return loc
}

func New(tag language.Tag, opts ...Options) (Locale, error) {
	loc, err := Parse(tag.String())
	if err != nil {
		return Locale{}, err
	}
	if len(opts) > 1 {
		return Locale{}, fmt.Errorf("locale: multiple options: %w", ErrInvalidOption)
	}
	if len(opts) > 0 {
		if err := applyLanguageOptions(&loc, opts[0]); err != nil {
			return Locale{}, fmt.Errorf("locale: invalid language identifier option: %w", ErrInvalidOption)
		}
		applyOptions(&loc, opts[0])
	}
	if err := loc.validate(); err != nil {
		return Locale{}, fmt.Errorf("%s: %w", err.Error(), ErrInvalidOption)
	}
	return loc, nil
}

func (l Locale) String() string {
	unicode := l.unicodeExtensionParts()
	parts := make([]string, 1, 2+len(unicode))
	parts[0] = l.BaseName()
	if len(unicode) == 0 {
		return parts[0]
	}
	parts = append(parts, "u")
	parts = append(parts, unicode...)
	return strings.Join(parts, "-")
}

func (l Locale) BaseName() string {
	return l.tag.String()
}

func (l Locale) Tag() language.Tag {
	return l.tag
}

func (l Locale) Calendar() string {
	return l.ext.calendar
}

func (l Locale) Collation() string {
	return l.ext.collation
}

func (l Locale) HourCycle() string {
	return l.ext.hourCycle
}

func (l Locale) CaseFirst() string {
	return l.ext.caseFirst
}

func (l Locale) Numeric() bool {
	return l.ext.numeric
}

func (l Locale) NumberingSystem() string {
	return l.ext.numberingSystem
}

func (l Locale) FirstDayOfWeek() string {
	return l.ext.firstDayOfWeek
}

func (l Locale) Language() string {
	lang, _, _ := tagParts(l.tag)
	return lang
}

func (l Locale) Script() string {
	_, script, _ := tagParts(l.tag)
	return script
}

func (l Locale) Region() string {
	_, _, region := tagParts(l.tag)
	return region
}

func (l Locale) Variants() []string {
	parts := strings.Split(l.BaseName(), "-")
	if len(parts) <= 1 {
		return nil
	}
	_, script, region := tagParts(l.tag)
	idx := 1
	if script != "" && idx < len(parts) && strings.EqualFold(parts[idx], script) {
		idx++
	}
	if region != "" && idx < len(parts) && strings.EqualFold(parts[idx], region) {
		idx++
	}
	return append([]string(nil), parts[idx:]...)
}

func (l Locale) Equal(other Locale) bool {
	return l.String() == other.String()
}

func (l Locale) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *Locale) UnmarshalText(text []byte) error {
	loc, err := Parse(string(text))
	if err != nil {
		return err
	}
	*l = loc
	return nil
}
