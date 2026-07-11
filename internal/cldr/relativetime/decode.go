// Hand-written decode layer for the relativetime domain. It expands
// domain-private const blobs from data.go into relative-field records consumed by
// accessors.go, behind per-blob sync.Once gates.
//
// Locale handle ownership: the field blob packs the locale index assigned by the
// cldr/locale kernel. Borrowing that handle keeps generated relative-time data
// and formatter locale resolution on one stable index space while the dependency
// stays one-way (relativetime -> cldr/locale).

package relativetime

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

// RelativeTimeField holds the future, past, and relative pattern maps for one
// (unit, style) combination. Future/Past are keyed by plural category; Relative
// is keyed by the numeric offset string.
type RelativeTimeField struct{ Future, Past, Relative map[string]string }

// RelativeTimeFields maps unit -> style -> RelativeTimeField.
type RelativeTimeFields map[string]map[string]RelativeTimeField

var (
	fieldOnce      sync.Once
	fieldsByLocale map[Locale]RelativeTimeFields
)

func loadFields() {
	r := codec.NewReader(_relativeTimeFieldBlob)
	fieldsByLocale = codec.Uint16DeltaMap[Locale, RelativeTimeFields](&r, decodeFields)
}

func decodeFields(r *codec.Reader) RelativeTimeFields {
	return codec.StringRefKeyMap[map[string]RelativeTimeField](r, _data, decodeFieldStyles)
}

func decodeFieldStyles(r *codec.Reader) map[string]RelativeTimeField {
	return codec.StringRefKeyMap[RelativeTimeField](r, _data, decodeRelativeTimeField)
}

func decodeRelativeTimeField(r *codec.Reader) RelativeTimeField {
	return RelativeTimeField{
		Future:   r.StringRefMap(_data),
		Past:     r.StringRefMap(_data),
		Relative: r.StringRefMap(_data),
	}
}

var supported = codec.NewLazyStrings(_relativeTimeSupportedBlob, _data)
