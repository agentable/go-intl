package ecma402

import "sync"

var defaultLocale = struct {
	sync.RWMutex
	value string
}{
	value: "en",
}

// DefaultLocale returns the implementation default locale used when Intl
// locale negotiation has no supported requested locale.
func DefaultLocale() string {
	defaultLocale.RLock()
	defer defaultLocale.RUnlock()
	return defaultLocale.value
}

// OverrideDefaultLocaleForTest replaces the implementation default locale and
// returns a restore function. It is an internal test hook, not public API.
func OverrideDefaultLocaleForTest(locale string) func() {
	defaultLocale.Lock()
	previous := defaultLocale.value
	defaultLocale.value = locale
	defaultLocale.Unlock()
	return func() {
		defaultLocale.Lock()
		defaultLocale.value = previous
		defaultLocale.Unlock()
	}
}
