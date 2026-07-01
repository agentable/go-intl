// Package unit holds the CLDR unit-pattern domain: simple unit patterns,
// compound unit patterns, and the unit-supported locale list. The payload lives
// in const blobs (data.go); decode.go expands them lazily, and accessors.go
// exposes the domain query surface.
package unit
