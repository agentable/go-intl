package codegen

import (
	"testing"

	"github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestRelativeTimeRoundTrip is the production-path round-trip gate for the
// relativetime domain. It re-derives the extract.RelativeTimeFields row stream
// from the real pinned CLDR checkout and asserts that every (unit, style,
// future/past/relative) pattern and every supported-locale tag the encoder wrote
// is queried back byte-for-byte through the production relativetime accessors
// over the committed data.go. It exercises encoder, blob, decoder, and accessor
// as one chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips, since there is no
// data to round-trip against.
func TestRelativeTimeRoundTrip(t *testing.T) {
	t.Parallel()

	input := loadRoundTripSource(t)
	fields := extract.ExtractRelativeTimeFields(input.source.RelativeTime, input.profile)
	locales := sortedLocaleKeys(fields)

	for _, localeTag := range locales {
		loc := resolveKernelLocale(t, localeTag)
		got := relativetime.FieldsFor(loc)
		want := fields[localeTag]

		if len(got) != len(want) {
			t.Errorf("FieldsFor(%q) has %d units, want %d", localeTag, len(got), len(want))
		}
		for unit, styles := range want {
			gotStyles := got[unit]
			for style, field := range styles {
				gotField := gotStyles[style]
				assertPatternMap(t, localeTag, unit, style, "future", field.Future, gotField.Future)
				assertPatternMap(t, localeTag, unit, style, "past", field.Past, gotField.Past)
				assertPatternMap(t, localeTag, unit, style, "relative", field.Relative, gotField.Relative)
			}
		}
	}

	wantTags := locales
	gotTags := relativetime.SupportedLocales()
	assertStringSliceEqual(t, "SupportedLocales", gotTags, wantTags)
}

func assertPatternMap(t *testing.T, locale, unit, style, section string, want, got map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s/%s/%s/%s has %d entries, want %d", locale, unit, style, section, len(got), len(want))
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s/%s/%s/%s[%q] = %q, want %q", locale, unit, style, section, key, got[key], value)
		}
	}
}
