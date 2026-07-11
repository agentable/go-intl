// Hand-written decode layer for the list domain. It expands domain-private const
// blobs from data.go into list-pattern records consumed by accessors.go, behind
// per-blob sync.Once gates.
//
// Locale handle ownership: the pattern blob packs the locale index assigned by
// the cldr/locale kernel. Borrowing that handle keeps generated list data and
// formatter locale resolution on one stable index space while the dependency
// stays one-way (list -> cldr/locale).

package list

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

// ListPattern holds the four list-join patterns for one (locale, type, style)
// combination.
type ListPattern struct{ Pair, Start, Middle, End string }

// listPatternRefs holds the already-resolved pattern strings for one
// (type, style) cell. The blob StringRefs are resolved against _data at decode
// time, so each field carries the final string the accessor returns.
type listPatternRefs struct{ pair, start, middle, end string }

var (
	patternOnce      sync.Once
	patternsByLocale map[Locale]map[string]map[string]listPatternRefs
)

func loadPatterns() {
	r := codec.NewReader(_listPatternBlob)
	patternsByLocale = codec.Uint16DeltaMap[Locale, map[string]map[string]listPatternRefs](&r, decodeListPatternTypes)
}

func decodeListPatternTypes(r *codec.Reader) map[string]map[string]listPatternRefs {
	return codec.StringRefKeyMap[map[string]listPatternRefs](r, _data, decodeListPatternStyles)
}

func decodeListPatternStyles(r *codec.Reader) map[string]listPatternRefs {
	return codec.StringRefKeyMap[listPatternRefs](r, _data, decodeListPatternRefs)
}

func decodeListPatternRefs(r *codec.Reader) listPatternRefs {
	return listPatternRefs{
		pair:   r.StringRef(_data),
		start:  r.StringRef(_data),
		middle: r.StringRef(_data),
		end:    r.StringRef(_data),
	}
}

var supported = codec.NewLazyStrings(_listSupportedBlob, _data)
