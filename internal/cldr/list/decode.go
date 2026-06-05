// Hand-written decode layer for the list domain. It expands the const blobs in
// data.go into the same in-memory structures the legacy root loadListPatternData
// produced, behind per-blob sync.Once gates.
//
// Locale handle ownership: the pattern blob packs a locale index that must match
// the runtime locale assignment, so this package borrows the Locale handle type
// from the cldr/locale kernel package. That kernel assigns locale indices
// identically to the legacy root package, so keys align; importing the kernel
// (rather than the root) keeps list's compile cost off the root data bomb. The
// dependency is one-way (list -> cldr/locale; the kernel never imports list, so
// there is no cycle), and it is the permanent home every domain points at.

package list

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

// ListPattern mirrors the legacy root type: the four list-join patterns for one
// (locale, type, style) combination.
type ListPattern struct{ Pair, Start, Middle, End string }

// listPatternRefs holds the already-resolved pattern strings for one
// (type, style) cell. The blob StringRefs are resolved against _data at decode
// time, so each field carries the final string the accessor returns — the same
// value the legacy sliceRef.string() yielded.
type listPatternRefs struct{ pair, start, middle, end string }

var (
	patternOnce      sync.Once
	patternsByLocale map[Locale]map[string]map[string]listPatternRefs

	supportedOnce sync.Once
	supportedTags []string
)

func loadPatterns() {
	r := codec.NewReader(_listPatternBlob)
	count := r.Uvarint()
	patternsByLocale = make(map[Locale]map[string]map[string]listPatternRefs, count)
	indices := codec.Delta(&r)
	for i := uint64(0); i < count; i++ {
		loc := Locale(indices.Next())
		patternsByLocale[loc] = decodePatterns(&r)
	}
}

func decodePatterns(r *codec.Reader) map[string]map[string]listPatternRefs {
	typeCount := r.Uvarint()
	if typeCount == 0 {
		return nil
	}
	types := make(map[string]map[string]listPatternRefs, typeCount)
	for t := uint64(0); t < typeCount; t++ {
		typ := r.StringRef(_data)
		styleCount := r.Uvarint()
		styles := make(map[string]listPatternRefs, styleCount)
		for s := uint64(0); s < styleCount; s++ {
			style := r.StringRef(_data)
			styles[style] = listPatternRefs{
				pair:   r.StringRef(_data),
				start:  r.StringRef(_data),
				middle: r.StringRef(_data),
				end:    r.StringRef(_data),
			}
		}
		types[typ] = styles
	}
	return types
}

func loadSupported() {
	r := codec.NewReader(_listSupportedBlob)
	count := r.Uvarint()
	supportedTags = make([]string, count)
	for i := range supportedTags {
		supportedTags[i] = r.StringRef(_data)
	}
}
