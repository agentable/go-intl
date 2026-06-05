// Package codec provides the lowest-level hand-written decode primitives that
// every cldr domain package builds on. It is a mover, not a parser: it owns the
// byte-level reading convention for the generated CLDR blobs and nothing more.
//
// # Trust contract
//
// codec trusts its input absolutely. Each blob is a const string emitted by the
// generator in tools/gen-cldr; the encoder and the decoder always live in the
// same commit, so the bytes a Reader walks are, by construction, well-formed.
// There is no schema version, no length sentinel, and no runtime validation:
// validating a contract that cannot desync is fitting a statue with a seatbelt.
// Correctness is proven at test time by the round-trip and conformance gates,
// never re-checked at runtime.
//
// # Why no error returns
//
// The primitives do not return errors. A malformed blob is a generator defect,
// not a recoverable user error, so there is no caller who could meaningfully
// handle one. An out-of-range string index is the only failure mode worth
// guarding, and Go's own bounds check shoots it down with a panic at the exact
// offending access; that panic is the single runtime assertion this design
// keeps, and surfacing it as a returned error would only let a corrupt build
// limp forward silently. Decode errors are program bugs, so they crash; they
// are not values to thread through accessor signatures.
//
// # Scope
//
// codec grows on demand. Today it carries only the primitives the first vertical
// slice actually needs:
//
//   - Uvarint  unsigned LEB128 varint, the width-driving record count and every
//     plain unsigned field.
//   - Delta    a monotonically non-decreasing stream layered on Uvarint, for
//     pre-sorted keys whose successive gaps the generator wrote. Its Next32
//     variant serves the unit domain, whose pattern keys the generator packs
//     into 32 bits.
//   - StringRef a (uvarint offset, uvarint length) pair resolved against a _data
//     string the caller passes in, yielding a zero-copy substring.
//   - Zigzag   a zigzag-encoded signed LEB128 varint. The timezone domain added
//     it: metazone period boundaries are int64 instants whose open ends are the
//     int64 min/max sentinels, the first signed stream in the payload.
//
// Each future primitive must name the domain and the specific data that requires
// it to exist.
//
// # Reader value semantics
//
// Reader is a small value type (a string plus an int cursor) and is passed and
// returned by value. The advancing methods take a pointer receiver so the cursor
// mutates in place during a decode loop, but a Reader is cheap to copy and holds
// no resource: the underlying _data is an immutable const string and a copied
// Reader simply shares it with no aliasing hazard. Value semantics keep the type
// allocation-free and make it trivial to snapshot a cursor position by copying.
package codec
