package codegen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
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

	patterns := extract.ExtractListPatterns(source.ListPatterns, profile)

	for _, localeTag := range listLocales(patterns) {
		loc := resolveListLocale(t, localeTag)
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

	wantTags := listSupportedLocaleTags(patterns)
	gotTags := cldrlist.SupportedLocales()
	if len(gotTags) != len(wantTags) {
		t.Fatalf("SupportedLocales len = %d, want %d", len(gotTags), len(wantTags))
	}
	for i := range wantTags {
		if gotTags[i] != wantTags[i] {
			t.Errorf("SupportedLocales[%d] = %q, want %q", i, gotTags[i], wantTags[i])
		}
	}
}

// resolveListLocale resolves a tag to the kernel handle the list accessors take.
// The pattern map is keyed by the kernel locale index, so a tag that fails to
// resolve would silently mis-key every pattern lookup.
func resolveListLocale(t *testing.T, tag string) cldrlocale.Locale {
	t.Helper()
	loc, ok := cldrlocale.ResolveLocale(tag)
	if !ok {
		t.Fatalf("kernel locale %q not resolvable", tag)
	}
	return loc
}
