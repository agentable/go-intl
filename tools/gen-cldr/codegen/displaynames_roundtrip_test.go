package codegen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentable/go-intl/internal/cldr/displaynames"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestDisplayNamesRoundTrip is the production-path round-trip gate for the
// displaynames domain. It re-derives the extract.DisplayNames map from the real
// pinned CLDR checkout and asserts that every styled display name the encoder
// wrote is queried back byte-for-byte through the production displaynames.Of
// accessor over the committed data.go, plus the supported-locale narrow index.
// It exercises encoder, blob, decoder, and accessor as one chain — not internal
// structures.
//
// The accessor walks the truncation parent chain and falls back to "en", so a
// row is verified only when its own locale is the first chain entry. Because the
// walk starts at the locale itself and the encoder stores no empty entries, the
// locale's own code always resolves on the first hop, making Of(locale, …) an
// exact identity check for that locale's data.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips.
func TestDisplayNamesRoundTrip(t *testing.T) {
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

	data := extract.ExtractDisplayNames(source.DisplayNames, profile)

	for locale, d := range data {
		// Language names, dialect and standard, exercised over the language blob.
		checkStyled(t, locale, "language", "dialect", d.Languages.Dialect)
		checkStyled(t, locale, "language", "standard", d.Languages.Standard)
		// Region, script, calendar, and date-time-field names, each over its own
		// blob.
		checkStyled(t, locale, "region", "", d.Territories)
		checkStyled(t, locale, "script", "", d.Scripts)
		checkStyled(t, locale, "calendar", "", d.Calendars)
		checkStyled(t, locale, "dateTimeField", "", d.DateTimeFields)
	}

	// Supported-locale narrow index: sorted-tag order, matching the encoder.
	wantTags := sortedDisplayNamesLocales(data)
	gotTags := displaynames.SupportedLocales()
	if len(gotTags) != len(wantTags) {
		t.Fatalf("SupportedLocales len = %d, want %d", len(gotTags), len(wantTags))
	}
	for i := range wantTags {
		if gotTags[i] != wantTags[i] {
			t.Errorf("SupportedLocales[%d] = %q, want %q", i, gotTags[i], wantTags[i])
		}
	}
}

// checkStyled asserts that every long/short/narrow code in s resolves back
// through displaynames.Of for the given locale and kind. The aliasing the
// accessor applies for calendar and date-time-field codes means a stored alias
// key (e.g. CLDR "gregorian") would not round-trip under its ECMA-402 spelling;
// those keys are exercised through the calendarAlias/dateTimeFieldAlias path in
// the package-level smoke tests instead, so here we skip the small set of alias
// source keys to keep the identity check exact.
func checkStyled(t *testing.T, locale, kind, languageDisplay string, s cldr.StyledNames) {
	t.Helper()
	checkStyleMap(t, locale, kind, "long", languageDisplay, s.Long)
	checkStyleMap(t, locale, kind, "short", languageDisplay, s.Short)
	checkStyleMap(t, locale, kind, "narrow", languageDisplay, s.Narrow)
}

func checkStyleMap(t *testing.T, locale, kind, style, languageDisplay string, m map[string]string) {
	t.Helper()
	for code, want := range m {
		if displayNamesAliasMasks(kind, code) {
			continue
		}
		got, ok := displaynames.Of(locale, kind, style, languageDisplay, code, "code")
		if !ok || got != want {
			t.Errorf("Of(%q, %q, %q, %q, %q) = %q (ok=%v), want %q",
				locale, kind, style, languageDisplay, code, got, ok, want)
		}
	}
}

// displayNamesAliasMasks reports whether the accessor would rewrite code before
// lookup, which makes a direct identity check against the stored key invalid.
// The accessor maps a small set of ECMA-402 spellings onto CLDR keys, so the
// reverse (a stored CLDR key queried directly) is the alias target, not the
// source; only the alias target keys are affected.
func displayNamesAliasMasks(kind, code string) bool {
	switch kind {
	case "calendar":
		return code == "gregory" || code == "ethioaa"
	case "dateTimeField":
		return code == "dayPeriod"
	}
	return false
}
