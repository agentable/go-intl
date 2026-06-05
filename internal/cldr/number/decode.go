// Hand-written decode layer for the number domain. It expands the const blobs in
// data.go into the same in-memory structures the legacy root loadNumbersByLocale
// produced, behind per-blob sync.Once gates.
//
// Locale handle ownership: the main blob packs a locale index that must match
// the runtime locale assignment, so this package borrows the locale index space
// from the cldr/locale kernel package. number.Locale is a distinct named type
// (not an alias) so it can carry the number accessor methods the formatters call
// as loc.DecimalPattern(…); it converts to the kernel handle only to share the
// kernel's locale-resolution and index assignment, which is identical to the
// legacy root package so keys align. Importing the kernel (rather than the root)
// keeps number's compile cost off the root data bomb. The dependency is one-way
// (number -> cldr/locale; the kernel never imports number, so there is no
// cycle), and it is the permanent home every domain points at.

package number

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
)

// Locale is the number-domain locale handle. It shares the kernel locale index
// space (see file header) and carries the number accessor methods in
// accessors.go.
type Locale uint16

// NumberSymbols holds the locale's numbering-system symbols. Its field set and
// order match the legacy root cldr.NumberSymbols so formatter output is
// byte-for-byte unchanged.
type NumberSymbols struct {
	Decimal, Group, Percent, Plus, Minus, NaN, Infinity, ApproxSign, PerMille, Exponential, SuperscriptingExponent, TimeSeparator string
}

// numberData mirrors the legacy root numberData: one locale's complete
// number-formatting payload, keyed internally by numbering system.
type numberData struct {
	defaultNS                    string
	symbols                      map[string]NumberSymbols
	decimal, percent, scientific map[string]string
	currency                     map[string]map[string]string
	compact                      map[string]map[string]map[int]map[string]string
}

var (
	numbersOnce sync.Once
	byLocale    map[Locale]numberData

	supportedOnce sync.Once
	supportedTags []string

	numberingSystemOnce   sync.Once
	numberingSystemExtras []string
)

func numberDataByLocale() map[Locale]numberData {
	numbersOnce.Do(loadNumbers)
	return byLocale
}

func loadNumbers() {
	r := codec.NewReader(_numberBlob)
	count := r.Uvarint()
	byLocale = make(map[Locale]numberData, count)
	indices := codec.Delta(&r)
	for i := uint64(0); i < count; i++ {
		loc := Locale(indices.Next())
		byLocale[loc] = decodeNumberLocale(&r)
	}
}

func decodeNumberLocale(r *codec.Reader) numberData {
	return numberData{
		defaultNS:  r.StringRef(_data),
		symbols:    decodeNumberSymbols(r),
		decimal:    decodeStringMap(r),
		percent:    decodeStringMap(r),
		scientific: decodeStringMap(r),
		currency:   decodeCurrencyPatterns(r),
		compact:    decodeCompactPatterns(r),
	}
}

func decodeNumberSymbols(r *codec.Reader) map[string]NumberSymbols {
	n := r.Uvarint()
	if n == 0 {
		return nil
	}
	out := make(map[string]NumberSymbols, n)
	for i := uint64(0); i < n; i++ {
		ns := r.StringRef(_data)
		out[ns] = NumberSymbols{
			Decimal:                r.StringRef(_data),
			Group:                  r.StringRef(_data),
			Percent:                r.StringRef(_data),
			Plus:                   r.StringRef(_data),
			Minus:                  r.StringRef(_data),
			NaN:                    r.StringRef(_data),
			Infinity:               r.StringRef(_data),
			ApproxSign:             r.StringRef(_data),
			PerMille:               r.StringRef(_data),
			Exponential:            r.StringRef(_data),
			SuperscriptingExponent: r.StringRef(_data),
			TimeSeparator:          r.StringRef(_data),
		}
	}
	return out
}

func decodeCurrencyPatterns(r *codec.Reader) map[string]map[string]string {
	n := r.Uvarint()
	if n == 0 {
		return nil
	}
	out := make(map[string]map[string]string, n)
	for i := uint64(0); i < n; i++ {
		ns := r.StringRef(_data)
		out[ns] = decodeStringMap(r)
	}
	return out
}

func decodeCompactPatterns(r *codec.Reader) map[string]map[string]map[int]map[string]string {
	nsCount := r.Uvarint()
	if nsCount == 0 {
		return nil
	}
	out := make(map[string]map[string]map[int]map[string]string, nsCount)
	for i := uint64(0); i < nsCount; i++ {
		ns := r.StringRef(_data)
		displayCount := r.Uvarint()
		displays := make(map[string]map[int]map[string]string, displayCount)
		for d := uint64(0); d < displayCount; d++ {
			display := r.StringRef(_data)
			expCount := r.Uvarint()
			exps := make(map[int]map[string]string, expCount)
			for e := uint64(0); e < expCount; e++ {
				exp := int(r.Uvarint())
				exps[exp] = decodeStringMap(r)
			}
			displays[display] = exps
		}
		out[ns] = displays
	}
	return out
}

func decodeStringMap(r *codec.Reader) map[string]string {
	n := r.Uvarint()
	if n == 0 {
		return nil
	}
	out := make(map[string]string, n)
	for i := uint64(0); i < n; i++ {
		key := r.StringRef(_data)
		out[key] = r.StringRef(_data)
	}
	return out
}

func loadSupported() {
	r := codec.NewReader(_numberSupportedBlob)
	count := r.Uvarint()
	supportedTags = make([]string, count)
	for i := range supportedTags {
		supportedTags[i] = r.StringRef(_data)
	}
}

func loadNumberingSystemExtras() {
	r := codec.NewReader(_numberingSystemBlob)
	count := r.Uvarint()
	numberingSystemExtras = make([]string, count)
	for i := range numberingSystemExtras {
		numberingSystemExtras[i] = r.StringRef(_data)
	}
}
