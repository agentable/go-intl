// Hand-written decode foundation; see doc.go for the trust contract.

package codec

// Reader walks a generated blob string by byte index.
//
// It is a value type: copy it freely to snapshot a cursor. The advancing methods
// use a pointer receiver so the cursor moves in place inside a decode loop. The
// underlying data is an immutable const string, so copies share it safely with
// no aliasing hazard.
type Reader struct {
	data string
	pos  int
}

// NewReader returns a Reader positioned at the start of blob.
func NewReader(blob string) Reader {
	return Reader{data: blob}
}

// Done reports whether the cursor has reached the end of the blob. A decode loop
// driven by a leading record count does not need this, but it is the natural
// terminator for streams whose length is implied by the blob itself.
func (r *Reader) Done() bool {
	return r.pos >= len(r.data)
}

// Pos returns the current byte offset of the cursor.
func (r *Reader) Pos() int {
	return r.pos
}

// Uvarint reads one unsigned LEB128 varint and advances the cursor past it.
//
// Encoding: 7 bits per byte, little-endian, with the high bit set on every byte
// except the last. This is the same wire form as encoding/binary's Uvarint, but
// inlined over a string cursor to stay zero-allocation and to read straight from
// the const _data without a []byte view.
func (r *Reader) Uvarint() uint64 {
	var x uint64
	var shift uint
	for {
		b := r.data[r.pos]
		r.pos++
		x |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return x
		}
		shift += 7
	}
}

// Zigzag reads one zigzag-encoded signed LEB128 varint and advances the cursor
// past it. Zigzag maps signed values to unsigned so that small-magnitude
// negatives stay small on the wire: n is encoded as (n<<1)^(n>>63), and decoded
// here as (u>>1)^-(u&1). The timezone domain requires it: metazone period
// boundaries are int64 instants whose open ends are the int64 min/max sentinels,
// the first signed stream in the generated payload.
func (r *Reader) Zigzag() int64 {
	u := r.Uvarint()
	return int64(u>>1) ^ -int64(u&1)
}

// DeltaReader decodes a monotonically non-decreasing stream of uint64 values
// layered on Uvarint. The generator writes each value as its gap from the
// previous one (the first gap is measured from zero), so the on-wire numbers
// stay small for sorted keys. DeltaReader holds the running previous value; the
// caller holds nothing, which keeps the call site minimal.
type DeltaReader struct {
	r    *Reader
	prev uint64
}

// Delta returns a DeltaReader that reads its gaps from r.
func Delta(r *Reader) DeltaReader {
	return DeltaReader{r: r}
}

// Next reads the next gap and returns the reconstructed absolute value.
func (d *DeltaReader) Next() uint64 {
	d.prev += d.r.Uvarint()
	return d.prev
}

// Next32 is Next for streams whose keys the generator packed into 32 bits. It
// returns the low 32 bits of the reconstructed value. Per the package trust
// contract the high bits are always zero for such a stream, so the mask is an
// exact, lossless narrowing that keeps the value in uint32 without a conversion
// the toolchain must treat as a possible overflow.
func (d *DeltaReader) Next32() uint32 {
	return uint32(d.Next() & 0xFFFFFFFF)
}

// StringRef reads a (uvarint offset, uvarint length) pair from the cursor and
// resolves it against data, returning a zero-copy substring. data is the caller's
// own domain-private _data table; codec never holds it.
//
// An offset or length past the end of data is a generator defect and trips Go's
// bounds check, which is the intended and only runtime assertion.
func (r *Reader) StringRef(data string) string {
	off := r.Uvarint()
	length := r.Uvarint()
	return data[off : off+length]
}
