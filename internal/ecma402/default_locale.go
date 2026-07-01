package ecma402

import (
	"sync"

	"github.com/agentable/go-intl/locale"
)

const defaultLocaleTag = "en"

var defaultLocale = struct {
	sync.RWMutex
	value  string
	parsed locale.Locale
}{
	value:  defaultLocaleTag,
	parsed: parseDefaultLocale(defaultLocaleTag),
}

// DefaultLocale returns the implementation default locale used when Intl
// locale negotiation has no supported requested locale.
func DefaultLocale() string {
	defaultLocale.RLock()
	defer defaultLocale.RUnlock()
	return defaultLocale.value
}

func defaultLocaleValue() locale.Locale {
	defaultLocale.RLock()
	defer defaultLocale.RUnlock()
	return defaultLocale.parsed
}

// OverrideDefaultLocaleForTest replaces the implementation default locale and
// returns a restore function. It is an internal test hook, not public API.
func OverrideDefaultLocaleForTest(locale string) func() {
	defaultLocale.Lock()
	previousValue := defaultLocale.value
	previousParsed := defaultLocale.parsed
	defaultLocale.value = locale
	defaultLocale.parsed = parseDefaultLocale(locale)
	defaultLocale.Unlock()
	return func() {
		defaultLocale.Lock()
		defaultLocale.value = previousValue
		defaultLocale.parsed = previousParsed
		defaultLocale.Unlock()
	}
}

func parseDefaultLocale(tag string) locale.Locale {
	loc, err := locale.Parse(tag)
	if err != nil {
		return locale.Locale{}
	}
	return loc
}
