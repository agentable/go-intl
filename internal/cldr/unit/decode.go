// Hand-written decode layer for the unit domain. It expands the const blobs in
// data.go into the same in-memory structures the legacy root loadUnitData
// produced, behind per-blob sync.Once gates.
//
// Locale handle ownership is transitional. The unit key packs a locale index
// that must match the runtime localeIndex assignment, so this package borrows
// the Locale handle type from the cldr/locale kernel package. That kernel
// assigns locale indices identically to the legacy root package, so keys align;
// importing the kernel (rather than the root) keeps unit's compile cost off the
// root data bomb. The dependency is one-way (unit -> cldr/locale; the kernel
// never imports unit, so there is no cycle), and it is already the permanent
// home the later migration phase points every domain at.

package unit

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

type unitWidth uint8

type unitPlural uint8

const (
	unitWidthLong unitWidth = iota + 1
	unitWidthNarrow
	unitWidthShort
)

const (
	unitPluralFew unitPlural = iota + 1
	unitPluralMany
	unitPluralOne
	unitPluralOther
	unitPluralTwo
	unitPluralZero
)

// unitPatternRecord and compoundUnitPatternRecord hold an already-resolved
// pattern string. The blob StringRef is resolved against _data at decode time,
// so the record carries the final string the accessor returns — the same value
// the legacy sliceRef.string() yielded.
type unitPatternRecord struct {
	key     uint32
	pattern string
}

type compoundUnitPatternRecord struct {
	key     uint32
	pattern string
}

var (
	// Pattern blob and the unit-name -> id map decode together: the name map is
	// the precondition for packing a pattern key, so both ride one Once.
	unitPatternOnce sync.Once
	unitPatterns    []unitPatternRecord
	unitNameIDs     map[string]uint32

	compoundUnitOnce sync.Once
	compoundUnitRows []compoundUnitPatternRecord

	unitSupportedOnce sync.Once
	unitSupportedTags []string
)

func loadUnitPatterns() {
	nameReader := codec.NewReader(_unitNameBlob)
	nameCount := nameReader.Uvarint()
	unitNameIDs = make(map[string]uint32, nameCount)
	for i := uint64(0); i < nameCount; i++ {
		unitNameIDs[nameReader.StringRef(_data)] = uint32(i) + 1
	}

	r := codec.NewReader(_unitPatternBlob)
	count := r.Uvarint()
	unitPatterns = make([]unitPatternRecord, count)
	keys := codec.Delta(&r)
	for i := range unitPatterns {
		key := keys.Next32()
		unitPatterns[i] = unitPatternRecord{key: key, pattern: r.StringRef(_data)}
	}
}

func loadCompoundUnits() {
	r := codec.NewReader(_compoundUnitBlob)
	count := r.Uvarint()
	compoundUnitRows = make([]compoundUnitPatternRecord, count)
	keys := codec.Delta(&r)
	for i := range compoundUnitRows {
		key := keys.Next32()
		compoundUnitRows[i] = compoundUnitPatternRecord{key: key, pattern: r.StringRef(_data)}
	}
}

func loadUnitSupported() {
	r := codec.NewReader(_unitSupportedBlob)
	count := r.Uvarint()
	unitSupportedTags = make([]string, count)
	for i := range unitSupportedTags {
		unitSupportedTags[i] = r.StringRef(_data)
	}
}
