package codegen

import (
	"fmt"
	"maps"
	"slices"
)

// blobEncoder accumulates a CLDR domain payload as raw bytes using the same wire
// contract read by internal/cldr/codec. The encoder and decoder ship together;
// the round-trip gate in tools/gen-cldr proves they stay in sync.
//
// Each primitive here has a decode counterpart:
//
//   - appendUvarint  <-> codec.Reader.Uvarint
//   - appendDelta    <-> codec.DeltaReader.Next
//   - appendStringRef <-> codec.Reader.StringRef
//   - appendStringRefSlice <-> codec.Reader.StringRefSlice
//   - appendStringRefMap   <-> codec.Reader.StringRefMap
//   - appendStringRefKeyMap <-> codec.StringRefKeyMap
//   - appendCountedSlice <-> codec.CountedSlice
//   - appendLocaleDeltaRecords <-> codec.Uint16DeltaMap
//   - appendUint32DeltaSlice <-> codec.Uint32DeltaSlice
//   - appendZigzag   <-> codec.Reader.Zigzag
//
// A new primitive is added only when a domain blob requires it, never
// speculatively.
type blobEncoder struct {
	buf  []byte
	prev uint64 // running previous value for the delta stream
}

// appendUvarint writes one unsigned LEB128 varint, the same wire form
// codec.Reader.Uvarint reads.
func (e *blobEncoder) appendUvarint(x uint64) {
	for x >= 0x80 {
		e.buf = append(e.buf, byte(x)|0x80)
		x >>= 7
	}
	e.buf = append(e.buf, byte(x))
}

// appendDelta writes the gap from the previous delta value (the first gap is
// measured from zero), reconstructed by codec.DeltaReader.Next. Values must be
// supplied in non-decreasing order; a regressing value would wrap the unsigned
// gap into garbage that decodes self-consistently but mis-keys every later
// record (and silently breaks binary-searched blobs), so fail loudly at
// generation time instead.
func (e *blobEncoder) appendDelta(v uint64) error {
	if v < e.prev {
		return fmt.Errorf("delta stream regressed: key %d follows %d", v, e.prev)
	}
	e.appendUvarint(v - e.prev)
	e.prev = v
	return nil
}

// appendZigzag writes one zigzag-encoded signed LEB128 varint, the wire form
// codec.Reader.Zigzag reads. It maps a signed value to unsigned as
// (v<<1)^(v>>63) so small-magnitude negatives stay compact.
func (e *blobEncoder) appendZigzag(v int64) {
	e.appendUvarint(uint64((v << 1) ^ (v >> 63)))
}

// resetDelta clears the running delta baseline so an encoder can carry several
// independent delta streams in sequence.
func (e *blobEncoder) resetDelta() {
	e.prev = 0
}

// appendStringRef writes a (offset, length) pair into the domain _data table,
// resolved by codec.Reader.StringRef against the same table.
func (e *blobEncoder) appendStringRef(ref StringRef) {
	e.appendUvarint(uint64(ref.start))
	e.appendUvarint(uint64(ref.length))
}

// appendStringRefSlice writes a count-prefixed sequence of string references.
func (e *blobEncoder) appendStringRefSlice(values []string, table *StringTable) {
	e.appendUvarint(uint64(len(values)))
	for _, value := range values {
		e.appendStringRef(table.Add(value))
	}
}

// appendStringRefMap writes a count-prefixed, key-sorted string-ref map.
func (e *blobEncoder) appendStringRefMap(values map[string]string, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(value string) {
		e.appendStringRef(table.Add(value))
	})
}

// appendStringRefKeyMap writes a count-prefixed, key-sorted map whose keys are
// string references and whose values are written by encode.
func appendStringRefKeyMap[V any](e *blobEncoder, values map[string]V, table *StringTable, encode func(V)) {
	keys := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(keys)))
	for _, key := range keys {
		e.appendStringRef(table.Add(key))
		encode(values[key])
	}
}

// appendCountedSlice writes a count-prefixed sequence whose values are written by
// encode.
func appendCountedSlice[V any](e *blobEncoder, values []V, encode func(V)) {
	e.appendUvarint(uint64(len(values)))
	for _, value := range values {
		encode(value)
	}
}

type localeDeltaRecord struct {
	locale string
	index  uint64
}

type uintDeltaRecord[V any] struct {
	value V
	key   uint64
}

// appendLocaleDeltaRecords writes a count-prefixed locale-index delta stream.
// Callers provide the domain-owned locale set; this primitive owns kernel-index
// lookup and appendUint64DeltaSlice owns ordering.
func (e *blobEncoder) appendLocaleDeltaRecords(locales []string, localeIndex map[string]uint64, encode func(locale string)) error {
	records := make([]localeDeltaRecord, len(locales))
	for i, locale := range locales {
		idx, err := localeIndexValue(localeIndex, locale)
		if err != nil {
			return err
		}
		records[i] = localeDeltaRecord{locale: locale, index: idx}
	}
	return appendUint64DeltaSlice(e, records, func(record localeDeltaRecord) (uint64, error) {
		return record.index, nil
	}, func(record localeDeltaRecord) {
		encode(record.locale)
	})
}

// appendUint32DeltaSlice writes a count-prefixed sequence whose uint32 keys are
// sorted, delta-coded in ascending order, and whose values are written by encode.
func appendUint32DeltaSlice[V any](e *blobEncoder, values []V, key func(V) (uint32, error), encode func(V)) error {
	return appendUint64DeltaSlice(e, values, func(value V) (uint64, error) {
		k, err := key(value)
		return uint64(k), err
	}, encode)
}

// appendUint64DeltaSlice writes a count-prefixed sequence whose uint64 keys are
// sorted, delta-coded in ascending order, and whose values are written by encode.
func appendUint64DeltaSlice[V any](e *blobEncoder, values []V, key func(V) (uint64, error), encode func(V)) error {
	records := make([]uintDeltaRecord[V], len(values))
	for i, value := range values {
		k, err := key(value)
		if err != nil {
			return err
		}
		records[i] = uintDeltaRecord[V]{value: value, key: k}
	}
	slices.SortStableFunc(records, func(a, b uintDeltaRecord[V]) int {
		return cmpUint64(a.key, b.key)
	})

	e.resetDelta()
	e.appendUvarint(uint64(len(records)))
	for _, record := range records {
		if err := e.appendDelta(record.key); err != nil {
			return err
		}
		encode(record.value)
	}
	return nil
}

// bytes returns the accumulated blob.
func (e *blobEncoder) bytes() []byte {
	return e.buf
}

func localeIndexValue(idx map[string]uint64, locale string) (uint64, error) {
	i, ok := idx[locale]
	if !ok {
		return 0, fmt.Errorf("locale %q missing from kernel locale registry", locale)
	}
	return i, nil
}

func cmpUint64(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
