package locale

import (
	"github.com/agentable/go-intl/internal/localeid"

	"golang.org/x/text/language"
)

type Locale struct {
	_         [0]func()
	tag       language.Tag
	ext       extensions
	canonical string
}

type extensions struct {
	attributes []string
	keywords   map[string]string
}

func Parse(s string) (Locale, error) {
	lower := localeid.LowercaseUnicodeLocaleID(s)
	if !localeid.HasUnicodeLocaleStructure(lower) {
		return Locale{}, invalidLocaleValue("languageTag", s, nil)
	}
	base, unicodeExtension, err := localeid.SplitUnicodeExtension(lower)
	if err != nil {
		return Locale{}, invalidLocaleValue("languageTag", s, err)
	}
	base = normalizeLanguageAliases(base)
	tag, err := localeid.ParseTag(base)
	if err != nil {
		return Locale{}, invalidLocaleValue("languageTag", s, err)
	}
	loc := Locale{tag: tag}
	loc.readExtensions(unicodeExtension)
	loc.freeze()
	return loc, nil
}

func New(tag string, opts Options) (Locale, error) {
	loc, err := Parse(tag)
	if err != nil {
		return Locale{}, err
	}
	if err := applyLanguageOptions(&loc, opts); err != nil {
		return Locale{}, err
	}
	if err := applyOptions(&loc, opts); err != nil {
		return Locale{}, err
	}
	loc.freeze()
	return loc, nil
}

func FromTag(tag language.Tag, opts Options) (Locale, error) {
	return New(tag.String(), opts)
}

func (l Locale) String() string {
	return l.canonicalString()
}

func (l Locale) canonicalString() string {
	if l.canonical != "" {
		return l.canonical
	}
	return l.formatString()
}

func (l Locale) formatString() string {
	base := l.tag.String()
	if l.ext.empty() {
		return base
	}
	return localeid.InsertUnicodeExtension(base, l.ext.attributes, l.unicodeExtensionKeywords())
}

func (l *Locale) freeze() {
	l.canonical = l.formatString()
}

func (l Locale) BaseName() string {
	return localeid.BaseName(l.tag)
}

func (l Locale) Tag() language.Tag {
	return l.tag
}

func (l Locale) Calendar() string {
	return l.ext.keywords["ca"]
}

func (l Locale) Collation() string {
	return l.ext.keywords["co"]
}

func (l Locale) HourCycle() string {
	return l.ext.keywords["hc"]
}

func (l Locale) CaseFirst() string {
	return l.ext.keywords["kf"]
}

func (l Locale) Numeric() bool {
	value, present := l.ext.keywords["kn"]
	return present && value == ""
}

func (l Locale) NumberingSystem() string {
	return l.ext.keywords["nu"]
}

func (l Locale) FirstDayOfWeek() string {
	return l.ext.keywords["fw"]
}

func (l Locale) Language() string {
	lang, _, _ := localeid.Parts(l.tag)
	return lang
}

func (l Locale) Script() string {
	_, script, _ := localeid.Parts(l.tag)
	return script
}

func (l Locale) Region() string {
	_, _, region := localeid.Parts(l.tag)
	return region
}

func (l Locale) Variants() []string {
	return localeid.VariantSubtags(l.tag)
}

func (l Locale) Equal(other Locale) bool {
	return l.canonicalString() == other.canonicalString()
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
