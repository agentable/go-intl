// Hand-written accessor layer for the number domain. It exposes symbol,
// pattern, numbering-system, and supported-index queries over lazily decoded
// const blobs.
//
// The accessors are methods on number.Locale so formatter call sites read as
// direct locale-domain queries: loc.DecimalPattern(...), loc.NumberSymbols(...),
// and loc.DefaultNumberingSystem().

package number

import (
	"slices"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
	"github.com/agentable/go-intl/internal/numbering"
)

const (
	defaultCurrencyPattern = "\u00a4#,##0.00"
	defaultTimeSeparator   = ":"
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
	return supported.Get()
}

// SupportedNumberingSystems returns the canonical numbering-system identifiers
// backed by generated number data merged with the ECMA-402 simple numbering
// systems, in sorted order. It reads only the narrow numbering-system blob and
// never triggers the main number-data decode.
func SupportedNumberingSystems() []string {
	numberingSystemOnce.Do(loadNumberingSystemExtras)
	return slices.Clone(supportedNumberingSystems)
}

func mergeSupportedNumberingSystems(extras []string) []string {
	out := numbering.SimpleNumberingSystems()
	for _, numberingSystem := range extras {
		if numberingSystem != "" && !slices.Contains(out, numberingSystem) {
			out = append(out, numberingSystem)
		}
	}
	slices.Sort(out)
	return out
}

// NumberSymbols returns explicit or inherited symbols. Missing rows follow the
// validated root alias to this locale's latn symbols; an empty numberingSystem
// selects the locale default.
func (l Locale) NumberSymbols(numberingSystem string) NumberSymbols {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	if symbols, ok := data.symbols[resolvedNumberingSystem]; ok {
		return withNumberSymbolDefaults(symbols)
	}
	return withNumberSymbolDefaults(data.symbols["latn"])
}

func withNumberSymbolDefaults(symbols NumberSymbols) NumberSymbols {
	if symbols.TimeSeparator == "" {
		symbols.TimeSeparator = defaultTimeSeparator
	}
	return symbols
}

// DecimalPattern returns the decimal pattern for the given numbering system,
// defaulting to the locale's default numbering system when numberingSystem is
// empty or has no generated pattern row.
func (l Locale) DecimalPattern(numberingSystem string) string {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	return numberPattern(data.decimal, resolvedNumberingSystem, data.defaultNumberingSystem)
}

// PercentPattern returns the percent pattern for the given numbering system,
// defaulting to the locale's default numbering system when numberingSystem is
// empty or has no generated pattern row.
func (l Locale) PercentPattern(numberingSystem string) string {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	return numberPattern(data.percent, resolvedNumberingSystem, data.defaultNumberingSystem)
}

// CurrencyPattern returns the currency pattern for the (numberingSystem, sign)
// pair, falling back to the locale's default numbering system and then the
// "standard" sign.
func (l Locale) CurrencyPattern(numberingSystem, sign string) string {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	patterns := data.currency[resolvedNumberingSystem]
	if len(patterns) == 0 && resolvedNumberingSystem != data.defaultNumberingSystem {
		patterns = data.currency[data.defaultNumberingSystem]
	}
	if pattern := patterns[sign]; pattern != "" {
		return pattern
	}
	if pattern := patterns["standard"]; pattern != "" {
		return pattern
	}
	return defaultCurrencyPattern
}

// CurrencySpacing returns the validated insertion text, using the locale's
// default numbering-system row when the requested row is absent.
func (l Locale) CurrencySpacing(numberingSystem string) CurrencySpacing {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	if spacing, ok := data.currencySpacing[resolvedNumberingSystem]; ok {
		return spacing
	}
	return data.currencySpacing[data.defaultNumberingSystem]
}

// CurrencyNamePattern returns the currency-name placement pattern for the
// given numbering system and plural category. It falls back to "other" within
// that numbering system, then repeats the lookup in the locale default.
func (l Locale) CurrencyNamePattern(numberingSystem, plural string) string {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	if pattern := currencyNamePattern(data.currencyName[resolvedNumberingSystem], plural); pattern != "" {
		return pattern
	}
	if resolvedNumberingSystem != data.defaultNumberingSystem {
		return currencyNamePattern(data.currencyName[data.defaultNumberingSystem], plural)
	}
	return ""
}

func currencyNamePattern(patterns currencyNamePluralPatterns, plural string) string {
	if pattern := patterns[plural]; pattern != "" {
		return pattern
	}
	return patterns["other"]
}

// ScientificPattern returns the scientific pattern for the given numbering
// system, defaulting to the locale's default numbering system when
// numberingSystem is empty or has no generated pattern row.
func (l Locale) ScientificPattern(numberingSystem string) string {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	return numberPattern(data.scientific, resolvedNumberingSystem, data.defaultNumberingSystem)
}

func numberPattern(patterns numberPatternsByNumberingSystem, numberingSystem, defaultNumberingSystem string) string {
	if pattern := patterns[numberingSystem]; pattern != "" {
		return pattern
	}
	if numberingSystem != defaultNumberingSystem {
		return patterns[defaultNumberingSystem]
	}
	return ""
}

// CompactPattern returns the compact pattern for the (numberingSystem, display,
// exponent, plural) tuple, falling back to the "other" plural form. An empty
// numbering system defaults to the locale's default numbering system.
func (l Locale) CompactPattern(numberingSystem, display string, exponent int, plural string) string {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	patterns := compactPatternRecord(data.compact, resolvedNumberingSystem, display, exponent)
	if pattern := patterns[plural]; pattern != "" {
		return pattern
	}
	return patterns["other"]
}

// CurrencyCompactPattern returns a currency compact row, with an optional
// alphaNextToNumber variant. Missing categories fall back to other within the
// selected variant. Missing tuples, including missing alternate rows, stay empty.
func (l Locale) CurrencyCompactPattern(numberingSystem, display string, exponent int, plural string, alphaNextToNumber bool) string {
	data, resolvedNumberingSystem := l.dataForResolvedNumberingSystem(numberingSystem)
	patterns := compactPatternRecord(data.currencyCompact, resolvedNumberingSystem, display, exponent)
	other := "other"
	if alphaNextToNumber {
		plural += "-alt-alphaNextToNumber"
		other += "-alt-alphaNextToNumber"
	}
	if pattern := patterns[plural]; pattern != "" {
		return pattern
	}
	return patterns[other]
}

func compactPatternRecord(data compactPatternsByNumberingSystem, numberingSystem, display string, exponent int) compactPluralPatterns {
	return data[numberingSystem][display][exponent]
}

// DefaultNumberingSystem returns the locale's default numbering system from the
// generated number data.
func (l Locale) DefaultNumberingSystem() string {
	return numberDataForLocale(l).defaultNumberingSystem
}

func (l Locale) dataForResolvedNumberingSystem(numberingSystem string) (numberData, string) {
	data := numberDataForLocale(l)
	if numberingSystem == "" {
		return data, data.defaultNumberingSystem
	}
	return data, numberingSystem
}

func numberDataForLocale(loc Locale) numberData {
	numbersOnce.Do(loadNumbers)
	record, ok := byLocale[loc]
	if !ok {
		return numberData{}
	}
	record.once.Do(func() {
		r := codec.NewReader(record.blob)
		record.numberData = decodeNumberLocale(&r)
	})
	return record.numberData
}

// NumberLocaleData exposes the relevant extension keys used by
// Intl.NumberFormat ResolveLocale.
type NumberLocaleData struct{}

// For returns the candidate values for a relevant extension key. Only "nu"
// (numbering system) is meaningful for the number domain.
func (NumberLocaleData) For(locale, key string) []string {
	if key != "nu" {
		return nil
	}
	return numberingSystemLocaleData(locale)
}

// numberingSystemLocaleData returns the locale's default numbering system
// followed by the simple numbering systems, deduplicated. The default is read
// from this domain's own number data.
func numberingSystemLocaleData(locale string) []string {
	defaultNumberingSystem := numbering.DefaultNumberingSystem
	if loc, ok := ResolveLocale(locale); ok {
		if ns := loc.DefaultNumberingSystem(); ns != "" {
			defaultNumberingSystem = ns
		}
	}
	return localeid.RelevantExtensionValues(defaultNumberingSystem, numbering.SimpleNumberingSystems()...)
}

// MinimumGroupingDigits returns the locale minimum for the leading group.
func (l Locale) MinimumGroupingDigits() int { return numberDataForLocale(l).minimumGroupingDigits }
