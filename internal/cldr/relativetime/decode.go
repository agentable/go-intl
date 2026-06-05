// Hand-written decode layer for the relativetime domain. It expands the const
// blobs in data.go into the same in-memory structures the legacy root
// loadRelativeTimeFieldData produced, behind per-blob sync.Once gates.
//
// Locale handle ownership: the field blob packs a locale index that must match
// the runtime locale assignment, so this package borrows the Locale handle type
// from the cldr/locale kernel package. That kernel assigns locale indices
// identically to the legacy root package, so keys align; importing the kernel
// (rather than the root) keeps relativetime's compile cost off the root data
// bomb. The dependency is one-way (relativetime -> cldr/locale; the kernel never
// imports relativetime, so there is no cycle), and it is the permanent home
// every domain points at.

package relativetime

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

// RelativeTimeField mirrors the legacy root type: the future, past, and relative
// pattern maps for one (unit, style) combination. Future/Past are keyed by
// plural category; Relative is keyed by the numeric offset string.
type RelativeTimeField struct{ Future, Past, Relative map[string]string }

// RelativeTimeFields mirrors the legacy root type: unit -> style ->
// RelativeTimeField.
type RelativeTimeFields map[string]map[string]RelativeTimeField

var (
	fieldOnce      sync.Once
	fieldsByLocale map[Locale]RelativeTimeFields

	supportedOnce sync.Once
	supportedTags []string
)

func loadFields() {
	r := codec.NewReader(_relativeTimeFieldBlob)
	count := r.Uvarint()
	fieldsByLocale = make(map[Locale]RelativeTimeFields, count)
	indices := codec.Delta(&r)
	for i := uint64(0); i < count; i++ {
		loc := Locale(indices.Next())
		fieldsByLocale[loc] = decodeFields(&r)
	}
}

func decodeFields(r *codec.Reader) RelativeTimeFields {
	unitCount := r.Uvarint()
	if unitCount == 0 {
		return nil
	}
	fields := make(RelativeTimeFields, unitCount)
	for u := uint64(0); u < unitCount; u++ {
		unit := r.StringRef(_data)
		styleCount := r.Uvarint()
		styles := make(map[string]RelativeTimeField, styleCount)
		for s := uint64(0); s < styleCount; s++ {
			style := r.StringRef(_data)
			styles[style] = RelativeTimeField{
				Future:   decodeStringMap(r),
				Past:     decodeStringMap(r),
				Relative: decodeStringMap(r),
			}
		}
		fields[unit] = styles
	}
	return fields
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
	r := codec.NewReader(_relativeTimeSupportedBlob)
	count := r.Uvarint()
	supportedTags = make([]string, count)
	for i := range supportedTags {
		supportedTags[i] = r.StringRef(_data)
	}
}
