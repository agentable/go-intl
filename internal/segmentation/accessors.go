// Package segmentation exposes the locale tags supported by the embedded segmenter.
//
// Segmenter boundaries are computed by github.com/rivo/uniseg using Unicode
// Standard Annex #29 algorithms. The boundary algorithm itself is
// locale-independent; this list mirrors the locales we actively conformance-test.
package segmentation

import "slices"

// SupportedLocales returns locale tags whose active word/sentence boundaries do
// not require dictionary or locale-specific tailoring beyond UAX #29 defaults.
func SupportedLocales() []string {
	return slices.Clone(supportedLocales)
}

var supportedLocales = []string{
	"ar",
	"de",
	"en",
	"en-GB",
	"en-US",
	"es",
	"fr",
	"hi",
	"it",
	"pt",
	"ru",
}
