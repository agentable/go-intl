package codegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/cldr/unit"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestUnitRoundTrip is the production-path round-trip gate for the unit domain.
// It re-derives the extract.Units row stream from the real pinned CLDR checkout
// and asserts that every simple pattern, compound pattern, and supported-locale
// tag the encoder wrote is queried back byte-for-byte through the production
// unit accessors over the committed data.go. It exercises encoder, blob,
// decoder, and accessor as one chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips, since there is no
// data to round-trip against.
func TestUnitRoundTrip(t *testing.T) {
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

	units := extract.ExtractUnits(source.Units, profile)

	// Simple unit patterns.
	for _, row := range unitPatternRows(units) {
		loc := resolveUnitLocale(t, row.locale)
		got := unit.UnitPattern(loc, row.unit, row.width, row.plural)
		if got != row.pattern {
			t.Errorf("UnitPattern(%q, %q, %q, %q) = %q, want %q",
				row.locale, row.unit, row.width, row.plural, got, row.pattern)
		}
	}

	// Compound unit patterns.
	for _, row := range compoundUnitPatternRows(units) {
		loc := resolveUnitLocale(t, row.locale)
		got := unit.CompoundUnitPattern(loc, row.width)
		if got != row.pattern {
			t.Errorf("CompoundUnitPattern(%q, %q) = %q, want %q",
				row.locale, row.width, got, row.pattern)
		}
	}

	// Supported-locale narrow index.
	wantTags := unitSupportedLocaleTags(units)
	gotTags := unit.SupportedLocales()
	if len(gotTags) != len(wantTags) {
		t.Fatalf("SupportedLocales len = %d, want %d", len(gotTags), len(wantTags))
	}
	for i := range wantTags {
		if gotTags[i] != wantTags[i] {
			t.Errorf("SupportedLocales[%d] = %q, want %q", i, gotTags[i], wantTags[i])
		}
	}
}

// resolveUnitLocale resolves a tag to the kernel handle the unit accessors take.
// The packed unit keys depend on the kernel locale index, so a tag that fails to
// resolve would silently mis-key every pattern lookup.
func resolveUnitLocale(t *testing.T, tag string) cldrlocale.Locale {
	t.Helper()
	loc, ok := cldrlocale.ResolveLocale(tag)
	if !ok {
		t.Fatalf("kernel locale %q not resolvable", tag)
	}
	return loc
}

func readUnitTestProfile(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read locale profile: %v", err)
	}
	var profile struct {
		Locales []string `json:"locales"`
	}
	if err := json.Unmarshal(raw, &profile); err != nil {
		t.Fatalf("parse locale profile: %v", err)
	}
	return profile.Locales
}
