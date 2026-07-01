// Hand-written decode layer for the currency domain. It expands domain-private
// const blobs from data.go into the maps consumed by accessors.go, behind
// per-blob sync.Once gates.
//
// Locale handle ownership: the names blob packs the locale index assigned by the
// cldr/locale kernel. Borrowing that handle keeps currency data, supported
// indexes, and formatter locale resolution on one stable index space while the
// dependency stays one-way (currency -> cldr/locale).
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

// Data holds the fraction-digit metadata for one currency code.
type Data struct{ DefaultDigits, CashDigits, Rounding int }

// currencyNames holds the per-locale display map keyed by plural category plus
// the canonical, symbol, and narrow scalars.
type currencyNames struct {
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
	byLocale  map[Locale]map[string]currencyNames

	supportedOnce sync.Once
	supportedTags []string
)

func loadFractions() {
	r := codec.NewReader(_currencyFractionBlob)
	fractions = codec.StringRefKeyMap[Data](&r, _data, decodeCurrencyFractionData)
}

func decodeCurrencyFractionData(r *codec.Reader) Data {
	return Data{
		DefaultDigits: int(r.Uvarint()),
		CashDigits:    int(r.Uvarint()),
		Rounding:      int(r.Uvarint()),
	}
}

func loadNames() {
	r := codec.NewReader(_currencyNamesBlob)
	byLocale = codec.Uint16DeltaMap[Locale, map[string]currencyNames](&r, decodeLocaleNames)
}

func decodeLocaleNames(r *codec.Reader) map[string]currencyNames {
	return codec.StringRefKeyMap[currencyNames](r, _data, decodeCurrencyNames)
}

func decodeCurrencyNames(r *codec.Reader) currencyNames {
	return currencyNames{
		display:   r.StringRefMap(_data),
		canonical: r.StringRef(_data),
		symbol:    r.StringRef(_data),
		narrow:    r.StringRef(_data),
	}
}

func loadSupported() {
	supportedTags = codec.StringRefSlice(_currencySupportedBlob, _data)
}
