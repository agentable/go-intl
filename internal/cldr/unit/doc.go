// Package unit holds the CLDR unit-pattern domain: simple unit patterns,
// compound unit patterns, and the unit-supported locale list. The payload lives
// in const blobs (data.go); decode.go expands them lazily into the same
// structures the legacy root accessors used, and accessors.go exposes the query
// surface with byte-identical semantics.
package unit
