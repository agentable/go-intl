// Hand-written decode layer for the currency domain. It expands the const blobs
// in data.go into the same in-memory structures the legacy root loadCurrencyData
// produced, behind per-blob sync.Once gates.
//
// Locale handle ownership: the names blob packs a locale index that must match
// the runtime locale assignment, so this package borrows the Locale handle type
// from the cldr/locale kernel package. That kernel assigns locale indices
// identically to the legacy root package, so keys align; importing the kernel
// (rather than the root) keeps currency's compile cost off the root data bomb.
// The dependency is one-way (currency -> cldr/locale; the kernel never imports
// currency, so there is no cycle), and it is the permanent home every domain
// points at.
//
// currency is the one sanctioned shared owner: both numberformat and
// displaynames read these accessors. That sharing is import-graph data, not a
// cycle — neither numberformat nor displaynames is imported here.

package currency

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

// Data mirrors the legacy root CurrencyData type: the fraction-digit metadata
// for one currency code.
type Data struct{ DefaultDigits, CashDigits, Rounding int }

// names mirrors the legacy root currencyNames type: the per-locale display map
// keyed by plural category plus the canonical, symbol, and narrow scalars.
type names struct {
	display                   map[string]string
	canonical, symbol, narrow string
}

var (
	// The fraction table and the per-locale names map decode independently: a
	// CurrencyDigits call must not force the much larger names decode, and a
	// name/symbol lookup must not force the fraction decode.
	fractionOnce sync.Once
	fractions    map[string]Data

	namesOnce sync.Once
	byLocale  map[Locale]map[string]names

	supportedOnce sync.Once
	supportedTags []string
)

func loadFractions() {
	r := codec.NewReader(_currencyFractionBlob)
	count := r.Uvarint()
	fractions = make(map[string]Data, count)
	for i := uint64(0); i < count; i++ {
		code := r.StringRef(_data)
		fractions[code] = Data{
			DefaultDigits: int(r.Uvarint()),
			CashDigits:    int(r.Uvarint()),
			Rounding:      int(r.Uvarint()),
		}
	}
}

func loadNames() {
	r := codec.NewReader(_currencyNamesBlob)
	count := r.Uvarint()
	byLocale = make(map[Locale]map[string]names, count)
	indices := codec.Delta(&r)
	for i := uint64(0); i < count; i++ {
		loc := Locale(indices.Next())
		byLocale[loc] = decodeLocaleNames(&r)
	}
}

func decodeLocaleNames(r *codec.Reader) map[string]names {
	codeCount := r.Uvarint()
	out := make(map[string]names, codeCount)
	for c := uint64(0); c < codeCount; c++ {
		code := r.StringRef(_data)
		out[code] = names{
			display:   decodeStringMap(r),
			canonical: r.StringRef(_data),
			symbol:    r.StringRef(_data),
			narrow:    r.StringRef(_data),
		}
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
	r := codec.NewReader(_currencySupportedBlob)
	count := r.Uvarint()
	supportedTags = make([]string, count)
	for i := range supportedTags {
		supportedTags[i] = r.StringRef(_data)
	}
}
