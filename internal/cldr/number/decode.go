// Hand-written decode layer for the number domain. It expands domain-private
// const blobs from data.go into numbering-system, pattern, symbol, and
// supported-index records consumed by accessors.go, behind per-blob sync.Once
// gates.
//
// Locale handle ownership: the main blob packs the locale index assigned by the
// cldr/locale kernel. number.Locale is a distinct named type so it can carry the
// number accessor methods formatters call as loc.DecimalPattern(...); it converts
// to the kernel handle only for shared locale resolution and index assignment.

package number

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
)

// Locale is the number-domain locale handle. It shares the kernel locale index
// space (see file header) and carries the number accessor methods in
// accessors.go.
type Locale uint16

// NumberSymbols holds the locale's numbering-system symbols in the wire order
// shared with the generator.
type NumberSymbols struct {
	Decimal, Group, Percent, Plus, Minus, NaN, Infinity, ApproxSign, RangeSign, PerMille, Exponential, SuperscriptingExponent, TimeSeparator string
}

// numberData holds one locale's number-formatting payload, keyed internally by
// numbering system.
type numberData struct {
	defaultNumberingSystem       string
	symbols                      numberSymbolsByNumberingSystem
	decimal, percent, scientific numberPatternsByNumberingSystem
	currency                     currencyPatternsByNumberingSystem
	compact                      compactPatternsByNumberingSystem
}

type numberSymbolsByNumberingSystem map[string]NumberSymbols
type numberPatternsByNumberingSystem map[string]string

type currencyPatternsByNumberingSystem map[string]currencySignPatterns
type currencySignPatterns map[string]string

type compactPatternsByNumberingSystem map[string]compactDisplayPatterns
type compactDisplayPatterns map[string]compactExponentPatterns
type compactExponentPatterns map[int]compactPluralPatterns
type compactPluralPatterns map[string]string

var (
	numbersOnce sync.Once
	byLocale    map[Locale]numberData

	numberingSystemOnce       sync.Once
	supportedNumberingSystems []string
)

func numberDataByLocale() map[Locale]numberData {
	numbersOnce.Do(loadNumbers)
	return byLocale
}

func loadNumbers() {
	r := codec.NewReader(_numberBlob)
	byLocale = codec.Uint16DeltaMap[Locale, numberData](&r, decodeNumberLocale)
}

func decodeNumberLocale(r *codec.Reader) numberData {
	return numberData{
		defaultNumberingSystem: r.StringRef(_data),
		symbols:                decodeNumberSymbols(r),
		decimal:                decodeNumberPatterns(r),
		percent:                decodeNumberPatterns(r),
		scientific:             decodeNumberPatterns(r),
		currency:               decodeCurrencyPatterns(r),
		compact:                decodeCompactPatterns(r),
	}
}

func decodeNumberSymbols(r *codec.Reader) numberSymbolsByNumberingSystem {
	return codec.StringRefKeyMap[NumberSymbols](r, _data, decodeNumberSymbolRow)
}

func decodeNumberSymbolRow(r *codec.Reader) NumberSymbols {
	return NumberSymbols{
		Decimal:                r.StringRef(_data),
		Group:                  r.StringRef(_data),
		Percent:                r.StringRef(_data),
		Plus:                   r.StringRef(_data),
		Minus:                  r.StringRef(_data),
		NaN:                    r.StringRef(_data),
		Infinity:               r.StringRef(_data),
		ApproxSign:             r.StringRef(_data),
		RangeSign:              r.StringRef(_data),
		PerMille:               r.StringRef(_data),
		Exponential:            r.StringRef(_data),
		SuperscriptingExponent: r.StringRef(_data),
		TimeSeparator:          r.StringRef(_data),
	}
}

func decodeNumberPatterns(r *codec.Reader) numberPatternsByNumberingSystem {
	return numberPatternsByNumberingSystem(r.StringRefMap(_data))
}

func decodeCurrencyPatterns(r *codec.Reader) currencyPatternsByNumberingSystem {
	return codec.StringRefKeyMap[currencySignPatterns](r, _data, decodeCurrencyPatternSet)
}

func decodeCurrencyPatternSet(r *codec.Reader) currencySignPatterns {
	return currencySignPatterns(r.StringRefMap(_data))
}

func decodeCompactPatterns(r *codec.Reader) compactPatternsByNumberingSystem {
	return codec.StringRefKeyMap[compactDisplayPatterns](r, _data, decodeCompactDisplays)
}

func decodeCompactDisplays(r *codec.Reader) compactDisplayPatterns {
	displayCount := r.Uvarint()
	displays := make(compactDisplayPatterns, displayCount)
	for range displayCount {
		display := r.StringRef(_data)
		displays[display] = decodeCompactExponents(r)
	}
	return displays
}

func decodeCompactExponents(r *codec.Reader) compactExponentPatterns {
	exponentCount := r.Uvarint()
	exponents := make(compactExponentPatterns, exponentCount)
	for range exponentCount {
		exponent := int(r.Uvarint())
		exponents[exponent] = compactPluralPatterns(r.StringRefMap(_data))
	}
	return exponents
}

var supported = codec.NewLazyStrings(_numberSupportedBlob, _data)

func loadNumberingSystemExtras() {
	supportedNumberingSystems = mergeSupportedNumberingSystems(decodeNumberingSystemExtras())
}

func decodeNumberingSystemExtras() []string {
	return codec.StringRefSlice(_numberingSystemBlob, _data)
}
