// Package segmenter implements the ECMA-402 Intl.Segmenter constructor.
//
//	locales, _ := locale.ParseList("en-US")
//	seg, _ := segmenter.New(locales, segmenter.Options{Granularity: segmenter.WordGranularity})
//	segments := seg.Segment("Hello, world!")
//	_ = segments
//
// See README.md for usage examples and SPECS/46-segmenter.md for the contract.
package segmenter
