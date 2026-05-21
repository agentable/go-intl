// Package segmenter implements the ECMA-402 Intl.Segmenter constructor.
//
//	seg, _ := segmenter.New(locale.MustParseList("en-US"), segmenter.Options{Granularity: segmenter.WordGranularity})
//	segments := seg.Segment("Hello, world!")
//	_ = segments
//
// See README.md for usage examples and SPECS/46-segmenter.md for the contract.
package segmenter
