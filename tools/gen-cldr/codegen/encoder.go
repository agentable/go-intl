package codegen

// blobEncoder accumulates a CLDR domain payload as raw bytes, mirroring the
// decode primitives in internal/cldr/codec. The encoder and that decoder always
// ship in the same commit, so they are byte-for-byte mirrors by construction;
// the round-trip gate in tools/gen-cldr proves they stay in sync.
//
// Each primitive here has a decode counterpart:
//
//   - appendUvarint  <-> codec.Reader.Uvarint
//   - appendDelta    <-> codec.DeltaReader.Next
//   - appendStringRef <-> codec.Reader.StringRef
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
func (e *blobEncoder) appendDelta(v uint64) {
	if v < e.prev {
		panic("delta stream regressed: keys must be encoded in non-decreasing order")
	}
	e.appendUvarint(v - e.prev)
	e.prev = v
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

// bytes returns the accumulated blob.
func (e *blobEncoder) bytes() []byte {
	return e.buf
}
