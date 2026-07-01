package codegen

import (
	"testing"

	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestListRoundTrip is the production-path round-trip gate for the list domain.
// It re-derives the extract.ListPatterns row stream from the real pinned CLDR
// checkout and asserts that every (type, style, {pair,start,middle,end}) pattern
// and every supported-locale tag the encoder wrote is queried back byte-for-byte
// through the production list accessors over the committed data.go. It exercises
// encoder, blob, decoder, and accessor as one chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips, since there is no
// data to round-trip against.
func TestListRoundTrip(t *testing.T) {
	t.Parallel()

	input := loadRoundTripSource(t)
	patterns := extract.ExtractListPatterns(input.source.ListPatterns, input.profile)
	locales := sortedLocaleKeys(patterns)

	for _, localeTag := range locales {
		loc := resolveKernelLocale(t, localeTag)
		for typ, styles := range patterns[localeTag] {
			for style, want := range styles {
				got := cldrlist.Pattern(loc, typ, style)
				if got.Pair != want.Pair {
					t.Errorf("Pattern(%q,%q,%q).Pair = %q, want %q", localeTag, typ, style, got.Pair, want.Pair)
				}
				if got.Start != want.Start {
					t.Errorf("Pattern(%q,%q,%q).Start = %q, want %q", localeTag, typ, style, got.Start, want.Start)
				}
				if got.Middle != want.Middle {
					t.Errorf("Pattern(%q,%q,%q).Middle = %q, want %q", localeTag, typ, style, got.Middle, want.Middle)
				}
				if got.End != want.End {
					t.Errorf("Pattern(%q,%q,%q).End = %q, want %q", localeTag, typ, style, got.End, want.End)
				}
			}
		}
	}

	wantTags := locales
	gotTags := cldrlist.SupportedLocales()
	assertStringSliceEqual(t, "SupportedLocales", gotTags, wantTags)
}
