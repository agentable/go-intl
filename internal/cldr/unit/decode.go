// Hand-written decode layer for the unit domain. It expands domain-private const
// blobs from data.go into unit-pattern, per-unit-pattern, compound-pattern, and
// supported-index records consumed by accessors.go, behind per-blob sync.Once
// gates.
//
// Locale handle ownership: packed unit keys use the locale index assigned by the
// cldr/locale kernel. Borrowing that handle keeps generated unit data and
// formatter locale resolution on one stable index space while the dependency
// stays one-way (unit -> cldr/locale).

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
// so the record carries the final string the accessor returns.
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

	perUnitPatternOnce sync.Once
	perUnitPatternRows []unitPatternRecord
)

func loadUnitPatterns() {
	nameReader := codec.NewReader(_unitNameBlob)
	nameCount := nameReader.Uvarint()
	unitNameIDs = make(map[string]uint32, nameCount)
	for i := uint64(0); i < nameCount; i++ {
		unitNameIDs[nameReader.StringRef(_data)] = uint32(i) + 1
	}

	r := codec.NewReader(_unitPatternBlob)
	unitPatterns = codec.Uint32DeltaSlice[unitPatternRecord](&r, decodeUnitPatternRecord)
}

func decodeUnitPatternRecord(key uint32, r *codec.Reader) unitPatternRecord {
	return unitPatternRecord{key: key, pattern: r.StringRef(_data)}
}

func loadCompoundUnits() {
	r := codec.NewReader(_compoundUnitBlob)
	compoundUnitRows = codec.Uint32DeltaSlice[compoundUnitPatternRecord](&r, decodeCompoundUnitPatternRecord)
}

func decodeCompoundUnitPatternRecord(key uint32, r *codec.Reader) compoundUnitPatternRecord {
	return compoundUnitPatternRecord{key: key, pattern: r.StringRef(_data)}
}

func loadPerUnitPatterns() {
	r := codec.NewReader(_perUnitPatternBlob)
	perUnitPatternRows = codec.Uint32DeltaSlice[unitPatternRecord](&r, decodeUnitPatternRecord)
}

var supported = codec.NewLazyStrings(_unitSupportedBlob, _data)
