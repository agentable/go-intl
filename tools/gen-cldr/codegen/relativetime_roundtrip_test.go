package codegen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
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

	repoRoot := filepath.Clean("../../..")
	cldrDir := filepath.Join(repoRoot, "tools", "gen-cldr", ".cldr-json", "node_modules")
	if _, err := os.Stat(cldrDir); err != nil {
		t.Skipf("pinned cldr-json checkout absent (%v); run task data:fetch", err)
	}

	versions, err := cldr.ReadVersionFile(filepath.Join(repoRoot, "internal", "cldr", "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	profile := readUnitTestProfile(t, filepath.Join(repoRoot, "tools", "locale-profile.json"))

	source, err := cldr.LoadAll(context.Background(), cldrDir, versions, profile)
	if err != nil {
		t.Fatalf("load cldr-json: %v", err)
	}

	fields := extract.ExtractRelativeTimeFields(source.RelativeTime, profile)

	for _, localeTag := range relativeTimeLocales(fields) {
		loc := resolveRelativeTimeLocale(t, localeTag)
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

	wantTags := relativeTimeSupportedLocaleTags(fields)
	gotTags := relativetime.SupportedLocales()
	if len(gotTags) != len(wantTags) {
		t.Fatalf("SupportedLocales len = %d, want %d", len(gotTags), len(wantTags))
	}
	for i := range wantTags {
		if gotTags[i] != wantTags[i] {
			t.Errorf("SupportedLocales[%d] = %q, want %q", i, gotTags[i], wantTags[i])
		}
	}
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

// resolveRelativeTimeLocale resolves a tag to the kernel handle the relativetime
// accessors take. The field map is keyed by the kernel locale index, so a tag
// that fails to resolve would silently mis-key every field lookup.
func resolveRelativeTimeLocale(t *testing.T, tag string) cldrlocale.Locale {
	t.Helper()
	loc, ok := cldrlocale.ResolveLocale(tag)
	if !ok {
		t.Fatalf("kernel locale %q not resolvable", tag)
	}
	return loc
}
