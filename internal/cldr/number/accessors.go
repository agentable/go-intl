// Hand-written accessor layer for the number domain. The query semantics mirror
// the legacy root cldr number accessors exactly, so NumberFormat,
// DurationFormat, and RelativeTimeFormat output is byte-for-byte unchanged.
//
// The accessors are methods on number.Locale so the formatter call sites
// (loc.DecimalPattern(…), loc.NumberSymbols(…), loc.DefaultNumberingSystem())
// stay identical to the legacy root cldr.Locale methods they replaced.

package number

import (
	"slices"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/numbering"
)

// ResolveLocale resolves a tag to its number-domain locale handle, forwarding to
// the shared locale kernel so number handles index identically to every other
// domain.
func ResolveLocale(tag string) (Locale, bool) {
	loc, ok := cldrlocale.ResolveLocale(tag)
	return Locale(loc), ok
}

// SupportedLocales returns the locales with generated number data in
// sorted-locale order. It reads only the narrow supported blob and never
// triggers the main number-data decode.
func SupportedLocales() []string {
	supportedOnce.Do(loadSupported)
	tags := make([]string, len(supportedTags))
	copy(tags, supportedTags)
	return tags
}

// SupportedNumberingSystems returns the canonical numbering-system identifiers
// backed by generated number data merged with the ECMA-402 simple numbering
// systems, in sorted order. It reads only the narrow numbering-system blob and
// never triggers the main number-data decode. The result mirrors the legacy
// root supported path exactly.
func SupportedNumberingSystems() []string {
	numberingSystemOnce.Do(loadNumberingSystemExtras)
	out := slices.Clone(numbering.SimpleNumberingSystems)
	for _, numberingSystem := range numberingSystemExtras {
		if numberingSystem != "" && !slices.Contains(out, numberingSystem) {
			out = append(out, numberingSystem)
		}
	}
	slices.Sort(out)
	return out
}

// NumberSymbols returns the symbols for the given numbering system, defaulting
// to the locale's default numbering system when numberingSystem is empty.
func (l Locale) NumberSymbols(numberingSystem string) NumberSymbols {
	data := numberDataByLocale()[l]
	if numberingSystem == "" {
		numberingSystem = data.defaultNS
	}
	return data.symbols[numberingSystem]
}

// DecimalPattern returns the decimal pattern for the given numbering system,
// defaulting to the locale's default numbering system when numberingSystem is
// empty.
func (l Locale) DecimalPattern(numberingSystem string) string {
	data := numberDataByLocale()[l]
	if numberingSystem == "" {
		numberingSystem = data.defaultNS
	}
	return data.decimal[numberingSystem]
}

// PercentPattern returns the percent pattern for the given numbering system,
// defaulting to the locale's default numbering system when numberingSystem is
// empty.
func (l Locale) PercentPattern(numberingSystem string) string {
	data := numberDataByLocale()[l]
	if numberingSystem == "" {
		numberingSystem = data.defaultNS
	}
	return data.percent[numberingSystem]
}

// CurrencyPattern returns the currency pattern for the (numberingSystem, sign)
// pair, falling back to the "standard" sign. An empty numbering system defaults
// to the locale's default numbering system.
func (l Locale) CurrencyPattern(numberingSystem, sign string) string {
	data := numberDataByLocale()[l]
	if numberingSystem == "" {
		numberingSystem = data.defaultNS
	}
	patterns := data.currency[numberingSystem]
	if pattern := patterns[sign]; pattern != "" {
		return pattern
	}
	return patterns["standard"]
}

// ScientificPattern returns the scientific pattern for the given numbering
// system, defaulting to the locale's default numbering system when
// numberingSystem is empty.
func (l Locale) ScientificPattern(numberingSystem string) string {
	data := numberDataByLocale()[l]
	if numberingSystem == "" {
		numberingSystem = data.defaultNS
	}
	return data.scientific[numberingSystem]
}

// CompactPattern returns the compact pattern for the (numberingSystem, display,
// exp, plural) tuple, falling back to the "other" plural form. An empty
// numbering system defaults to the locale's default numbering system.
func (l Locale) CompactPattern(numberingSystem, display string, exp int, plural string) string {
	data := numberDataByLocale()[l]
	if numberingSystem == "" {
		numberingSystem = data.defaultNS
	}
	patterns := data.compact[numberingSystem][display][exp]
	if pattern := patterns[plural]; pattern != "" {
		return pattern
	}
	return patterns["other"]
}

// DefaultNumberingSystem returns the locale's default numbering system from the
// generated number data, mirroring the legacy root cldr.Locale method.
func (l Locale) DefaultNumberingSystem() string {
	return numberDataByLocale()[l].defaultNS
}

// NumberLocaleData exposes the relevant extension keys used by
// Intl.NumberFormat ResolveLocale. It mirrors the legacy root cldr type.
type NumberLocaleData struct{}

// For returns the candidate values for a relevant extension key. Only "nu"
// (numbering system) is meaningful for the number domain.
func (NumberLocaleData) For(locale, key string) []string {
	if key != "nu" {
		return nil
	}
	return numberingSystemLocaleData(locale)
}

// numberingSystemLocaleData mirrors the legacy root helper: it returns the
// locale's default numbering system followed by the simple numbering systems,
// deduplicated. The default is read from this domain's own number data, exactly
// as the legacy root path did.
func numberingSystemLocaleData(locale string) []string {
	defaultNS := "latn"
	if loc, ok := ResolveLocale(locale); ok {
		if ns := loc.DefaultNumberingSystem(); ns != "" {
			defaultNS = ns
		}
	}
	return keywordLocaleData(defaultNS, numbering.SimpleNumberingSystems)
}

func keywordLocaleData(defaultValue string, values []string) []string {
	out := make([]string, 0, len(values)+1)
	if defaultValue != "" {
		out = append(out, defaultValue)
	}
	for _, value := range values {
		if value != "" && !slices.Contains(out, value) {
			out = append(out, value)
		}
	}
	return out
}
