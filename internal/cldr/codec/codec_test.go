package codec

import (
	"math"
	"testing"
)

// appendUvarint mirrors the encoder the generator will eventually emit in
// tools/gen-cldr; it exists here only to build byte streams the decoder can read
// back, so the primitives are testable before the production encoder lands.
func appendUvarint(dst []byte, x uint64) []byte {
	for x >= 0x80 {
		dst = append(dst, byte(x)|0x80)
		x >>= 7
	}
	return append(dst, byte(x))
}

// appendDelta encodes a monotonically non-decreasing sequence as gaps.
func appendDelta(dst []byte, values []uint64) []byte {
	var prev uint64
	for _, v := range values {
		dst = appendUvarint(dst, v-prev)
		prev = v
	}
	return dst
}

// appendStringRef encodes an (offset, length) pair into data.
func appendStringRef(dst []byte, off, length uint64) []byte {
	dst = appendUvarint(dst, off)
	dst = appendUvarint(dst, length)
	return dst
}

func TestUvarint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		val  uint64
	}{
		{"zero", 0},
		{"one", 1},
		{"max single byte", 127},
		{"first two byte", 128},
		{"two byte mid", 300},
		{"max uint32", math.MaxUint32},
		{"max uint64", math.MaxUint64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blob := string(appendUvarint(nil, tt.val))
			r := NewReader(blob)
			got := r.Uvarint()
			if got != tt.val {
				t.Errorf("Uvarint() = %d, want %d", got, tt.val)
			}
			if !r.Done() {
				t.Errorf("Done() = false after reading whole stream, pos=%d len=%d", r.Pos(), len(blob))
			}
		})
	}
}

func TestUvarintByteWidth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		val   uint64
		width int
	}{
		{"zero one byte", 0, 1},
		{"127 one byte", 127, 1},
		{"128 two bytes", 128, 2},
		{"max uint64 ten bytes", math.MaxUint64, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blob := appendUvarint(nil, tt.val)
			if len(blob) != tt.width {
				t.Errorf("encoded width = %d, want %d", len(blob), tt.width)
			}
		})
	}
}

func TestUvarintSequence(t *testing.T) {
	t.Parallel()
	// Multiple varints packed back to back are read in order with the cursor
	// advancing across boundaries between single- and multi-byte values.
	values := []uint64{0, 1, 127, 128, 16384, math.MaxUint64, 42}
	var buf []byte
	for _, v := range values {
		buf = appendUvarint(buf, v)
	}
	r := NewReader(string(buf))
	for i, want := range values {
		if got := r.Uvarint(); got != want {
			t.Fatalf("value %d: Uvarint() = %d, want %d", i, got, want)
		}
	}
	if !r.Done() {
		t.Errorf("Done() = false after reading all values, pos=%d", r.Pos())
	}
}

func TestDelta(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		values []uint64
	}{
		{"empty", nil},
		{"single zero", []uint64{0}},
		{"single nonzero", []uint64{500}},
		{"strictly increasing", []uint64{1, 2, 3, 130, 131, 100000}},
		{"with repeats", []uint64{0, 0, 5, 5, 5, 200}},
		{"boundary gaps", []uint64{127, 255, 16383, 16384}},
		{"to max uint64", []uint64{0, math.MaxUint64}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blob := string(appendDelta(nil, tt.values))
			r := NewReader(blob)
			d := Delta(&r)
			for i, want := range tt.values {
				if got := d.Next(); got != want {
					t.Fatalf("value %d: Next() = %d, want %d", i, got, want)
				}
			}
			if !r.Done() {
				t.Errorf("Done() = false after reading delta stream")
			}
		})
	}
}

func TestDeltaNext32(t *testing.T) {
	t.Parallel()
	// Next32 returns the same sequence as Next for 32-bit-bounded keys, narrowed
	// to uint32. The unit domain packs its pattern keys this way.
	values := []uint64{0, 1, 256, 4096, 1 << 20, math.MaxUint32}
	blob := string(appendDelta(nil, values))
	r := NewReader(blob)
	d := Delta(&r)
	for i, want := range values {
		if got := d.Next32(); uint64(got) != want {
			t.Fatalf("value %d: Next32() = %d, want %d", i, got, want)
		}
	}
	if !r.Done() {
		t.Errorf("Done() = false after reading delta stream")
	}
}

// appendZigzag mirrors the generator's signed-varint encoder so the decoder can
// be exercised before the production encoder lands.
func appendZigzag(dst []byte, v int64) []byte {
	return appendUvarint(dst, uint64((v<<1)^(v>>63)))
}

func TestZigzag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		val  int64
	}{
		{"zero", 0},
		{"one", 1},
		{"minus one", -1},
		{"small positive", 1234567},
		{"small negative", -1234567},
		{"max int64", math.MaxInt64},
		{"min int64", math.MinInt64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blob := string(appendZigzag(nil, tt.val))
			r := NewReader(blob)
			if got := r.Zigzag(); got != tt.val {
				t.Errorf("Zigzag() = %d, want %d", got, tt.val)
			}
			if !r.Done() {
				t.Errorf("Done() = false after reading whole stream, pos=%d len=%d", r.Pos(), len(blob))
			}
		})
	}
}

func TestZigzagSequence(t *testing.T) {
	t.Parallel()
	values := []int64{0, -1, 1, math.MinInt64, math.MaxInt64, -42, 42}
	var buf []byte
	for _, v := range values {
		buf = appendZigzag(buf, v)
	}
	r := NewReader(string(buf))
	for i, want := range values {
		if got := r.Zigzag(); got != want {
			t.Fatalf("value %d: Zigzag() = %d, want %d", i, got, want)
		}
	}
	if !r.Done() {
		t.Errorf("Done() = false after reading all values, pos=%d", r.Pos())
	}
}

func TestStringRef(t *testing.T) {
	t.Parallel()
	const data = "undafamarzh-Hant" // und|af|am|ar|zh-Hant at offsets 0|3|5|7|9
	tests := []struct {
		name string
		off  uint64
		ln   uint64
		want string
	}{
		{"leading", 0, 3, "und"},
		{"empty at start", 0, 0, ""},
		{"middle", 3, 2, "af"},
		{"empty mid", 5, 0, ""},
		{"tail", 9, uint64(len("zh-Hant")), "zh-Hant"},
		{"whole table", 0, uint64(len(data)), data},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blob := string(appendStringRef(nil, tt.off, tt.ln))
			r := NewReader(blob)
			got := r.StringRef(data)
			if got != tt.want {
				t.Errorf("StringRef() = %q, want %q", got, tt.want)
			}
			if !r.Done() {
				t.Errorf("Done() = false after reading ref")
			}
		})
	}
}

func TestStringRefSequence(t *testing.T) {
	t.Parallel()
	// A blob of consecutive refs into one shared data table, as a domain
	// decode.go would walk after reading a leading record count.
	const data = "deenfrjaes"
	type ref struct {
		off, ln uint64
		want    string
	}
	refs := []ref{
		{0, 2, "de"},
		{2, 2, "en"},
		{4, 2, "fr"},
		{6, 2, "ja"},
		{8, 2, "es"},
	}
	var buf []byte
	buf = appendUvarint(buf, uint64(len(refs))) // record count, per blob convention
	for _, rf := range refs {
		buf = appendStringRef(buf, rf.off, rf.ln)
	}
	r := NewReader(string(buf))
	n := r.Uvarint()
	if n != uint64(len(refs)) {
		t.Fatalf("record count = %d, want %d", n, len(refs))
	}
	for i := range n {
		got := r.StringRef(data)
		if got != refs[i].want {
			t.Errorf("ref %d: StringRef() = %q, want %q", i, got, refs[i].want)
		}
	}
	if !r.Done() {
		t.Errorf("Done() = false after reading all refs")
	}
}

func TestEmptyBlob(t *testing.T) {
	t.Parallel()
	r := NewReader("")
	if !r.Done() {
		t.Errorf("Done() = false on empty blob")
	}
	if r.Pos() != 0 {
		t.Errorf("Pos() = %d on empty blob, want 0", r.Pos())
	}
}

func TestReaderCopySemantics(t *testing.T) {
	t.Parallel()
	// Copying a Reader snapshots the cursor: advancing the copy must not move
	// the original.
	blob := string(appendUvarint(appendUvarint(nil, 10), 20))
	r := NewReader(blob)
	snapshot := r
	if got := r.Uvarint(); got != 10 {
		t.Fatalf("r.Uvarint() = %d, want 10", got)
	}
	if snapshot.Pos() != 0 {
		t.Errorf("snapshot advanced with original; Pos() = %d, want 0", snapshot.Pos())
	}
	if got := snapshot.Uvarint(); got != 10 {
		t.Errorf("snapshot.Uvarint() = %d, want 10 (independent cursor)", got)
	}
}

// mustPanic asserts that fn trips a runtime panic. The trust contract makes
// Go's bounds check the package's only runtime assertion on malformed blobs;
// these tests pin that a defect panics instead of being silently clamped.
func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected bounds-check panic, got none", name)
		}
	}()
	fn()
}

func TestUvarintTruncatedPanics(t *testing.T) {
	t.Parallel()
	// A blob ending mid-varint (continuation bit set on the final byte) must
	// run the cursor past the end and panic.
	r := NewReader("\xff")
	mustPanic(t, "Uvarint on truncated stream", func() { r.Uvarint() })
}

func TestStringRefOutOfRangePanics(t *testing.T) {
	t.Parallel()
	const data = "abc"
	tests := []struct {
		name        string
		off, length uint64
	}{
		{"offset past end", 4, 1},
		{"length past end", 1, 3},
		{"offset plus length overflows", math.MaxUint64, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewReader(string(appendStringRef(nil, tt.off, tt.length)))
			mustPanic(t, tt.name, func() { r.StringRef(data) })
		})
	}
}
