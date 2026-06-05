package codegen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestLocaleKernelRoundTrip is the production-path round-trip gate for the locale
// kernel domain. It re-derives the locale registry, likely-subtags, numbering,
// and preference inputs from the real pinned CLDR checkout and asserts that every
// row the encoder wrote is queried back byte-for-byte through the production
// cldrlocale accessors over the committed data.go. It exercises encoder, blob,
// decoder, and accessor as one chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips.
func TestLocaleKernelRoundTrip(t *testing.T) {
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

	locales := extract.ExtractLocales(source.Available, profile)
	likely := extract.ExtractLikelySubtags(source.LikelySubtags)
	numbers := extract.ExtractNumbers(source.Numbers, profile)
	preferences := extract.ExtractPreference(source.Preference)

	// Locale registry: index assignment and available tags.
	for i, tag := range locales.Tags {
		loc, ok := cldrlocale.ResolveLocale(tag)
		if !ok {
			t.Errorf("ResolveLocale(%q) = false, want true", tag)
			continue
		}
		if uint16(loc) != uint16(i) {
			t.Errorf("ResolveLocale(%q) = %d, want index %d", tag, loc, i)
		}
	}
	gotTags := cldrlocale.AvailableLocales()
	if len(gotTags) != len(locales.Tags) {
		t.Fatalf("AvailableLocales len = %d, want %d", len(gotTags), len(locales.Tags))
	}
	for i := range locales.Tags {
		if gotTags[i] != locales.Tags[i] {
			t.Errorf("AvailableLocales[%d] = %q, want %q", i, gotTags[i], locales.Tags[i])
		}
	}

	// Likely-subtags maximize rows.
	for key, triple := range likely.Maximize {
		language, script, region := splitTriple(key)
		lang, scr, reg, ok := cldrlocale.MaximizeSubtags(language, script, region)
		if !ok {
			t.Errorf("MaximizeSubtags(%q) = false, want true", key)
			continue
		}
		if lang != triple.Lang || scr != triple.Script || reg != triple.Region {
			t.Errorf("MaximizeSubtags(%q) = %q, %q, %q; want %q, %q, %q",
				key, lang, scr, reg, triple.Lang, triple.Script, triple.Region)
		}
	}

	// Likely-subtags minimize rows.
	for triple, minimized := range likely.Minimize {
		min, _, _, ok := cldrlocale.MinimizeSubtags(triple.Lang, triple.Script, triple.Region)
		if !ok {
			t.Errorf("MinimizeSubtags(%q,%q,%q) = false, want true", triple.Lang, triple.Script, triple.Region)
			continue
		}
		if min != minimized {
			t.Errorf("MinimizeSubtags(%q,%q,%q) = %q, want %q", triple.Lang, triple.Script, triple.Region, min, minimized)
		}
	}

	// Numbering overrides: every non-latn locale resolves to its system; every
	// other locale defaults to latn.
	for _, tag := range locales.Tags {
		loc, ok := cldrlocale.ResolveLocale(tag)
		if !ok {
			continue
		}
		want := numbers[tag].DefaultNumberingSystem
		if want == "" {
			want = "latn"
		}
		if got := loc.DefaultNumberingSystem(); got != want {
			t.Errorf("%q DefaultNumberingSystem = %q, want %q", tag, got, want)
		}
	}

	// Preference rows.
	for region, values := range preferences.HourCycle {
		got := cldrlocale.HourCyclePreference(region)
		if !equalStrings(got, values) {
			t.Errorf("HourCyclePreference(%q) = %v, want %v", region, got, values)
		}
		if !cldrlocale.HasHourCyclePreference(region) {
			t.Errorf("HasHourCyclePreference(%q) = false, want true", region)
		}
	}
	for region, values := range preferences.Calendar {
		got := cldrlocale.CalendarPreference(region)
		if !equalStrings(got, values) {
			t.Errorf("CalendarPreference(%q) = %v, want %v", region, got, values)
		}
		if !cldrlocale.HasCalendarPreference(region) {
			t.Errorf("HasCalendarPreference(%q) = false, want true", region)
		}
	}
	for region, week := range preferences.Week {
		if got := uint64(cldrlocale.FirstDayOfWeek(region)); got != weekdayOrdinal(week.FirstDay) {
			t.Errorf("FirstDayOfWeek(%q) = %d, want %d", region, got, weekdayOrdinal(week.FirstDay))
		}
		start, end := cldrlocale.Weekend(region)
		if uint64(start) != weekdayOrdinal(week.WeekendStart) || uint64(end) != weekdayOrdinal(week.WeekendEnd) {
			t.Errorf("Weekend(%q) = %d,%d; want %d,%d", region, start, end,
				weekdayOrdinal(week.WeekendStart), weekdayOrdinal(week.WeekendEnd))
		}
		if got := cldrlocale.MinimalDaysInFirstWeek(region); got != week.MinDays {
			t.Errorf("MinimalDaysInFirstWeek(%q) = %d, want %d", region, got, week.MinDays)
		}
		if !cldrlocale.HasWeekPreference(region) {
			t.Errorf("HasWeekPreference(%q) = false, want true", region)
		}
	}
}

func splitTriple(key string) (language, script, region string) {
	parts := splitDash(key)
	if len(parts) > 0 {
		language = parts[0]
	}
	if len(parts) > 1 {
		script = parts[1]
	}
	if len(parts) > 2 {
		region = parts[2]
	}
	return language, script, region
}

func splitDash(s string) []string {
	var out []string
	start := 0
	for i := range len(s) {
		if s[i] == '-' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
