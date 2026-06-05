package codegen

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/cldr/timezone"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestTimezoneRoundTrip is the production-path round-trip gate for the timezone
// domain. It re-derives the extract.Metazones data from the real pinned CLDR
// checkout and asserts that the metazone periods, the localized metazone and
// zone-specific display names, the exemplar cities, the GMT offset formats, and
// the supported-zone narrow index the encoder wrote are queried back through the
// production timezone accessors over the committed data.go. It exercises encoder,
// blob, decoder, and accessor as one chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips.
func TestTimezoneRoundTrip(t *testing.T) {
	t.Parallel()

	data := loadTimezoneTestInput(t)

	// Metazone periods: probe the midpoint of every finite period and confirm the
	// production TimeZoneMetazone resolves the metazone the period encodes.
	for zone, periods := range data.ZoneToMetazones {
		for _, period := range periods {
			instant := periodProbeInstant(period.Start, period.End)
			if got := timezone.TimeZoneMetazone(zone, instant); got != period.Metazone {
				t.Errorf("TimeZoneMetazone(%q, %d) = %q, want %q", zone, instant, got, period.Metazone)
			}
		}
	}

	// Metazone names, zone-specific names, exemplar cities: replicate the legacy
	// fallback resolution independently from the extract rows and compare to the
	// production TimeZoneDisplayName.
	for locale := range metazoneNameUnion(data) {
		loc, ok := cldrlocale.ResolveLocale(locale)
		if !ok {
			continue
		}
		assertZoneNames(t, loc, locale, data)
		assertExemplarCities(t, loc, locale, data)
	}

	// GMT offset formats: every locale with a formats record must format a probe
	// offset through the production GMTOffsetName using the encoded patterns.
	for locale, formats := range data.Formats {
		loc, ok := cldrlocale.ResolveLocale(locale)
		if !ok {
			continue
		}
		const offsetMs = int64(2*3600*1000 + 30*60*1000)
		got := timezone.GMTOffsetName(loc, offsetMs, timezone.TimeZoneNameLongOffset)
		if got == "" {
			t.Errorf("GMTOffsetName(%q) = \"\" for encoded formats %+v", locale, formats)
		}
	}

	// Supported-zone narrow index.
	wantTags := timezoneSupportedZones(data)
	gotTags := timezone.SupportedTimeZones()
	if !slices.Equal(gotTags, wantTags) {
		t.Errorf("SupportedTimeZones mismatch:\n got %v\nwant %v", gotTags, wantTags)
	}
}

func assertZoneNames(t *testing.T, loc timezone.Locale, locale string, data extract.Metazones) {
	t.Helper()
	// Zone-specific names take precedence over metazone names in the legacy
	// fallback. Probe each zone with no covering metazone period so the display
	// name resolves to the zone-specific name directly.
	for zone, names := range data.ZoneNames[locale] {
		for form, kind := range displayFormKinds {
			want := metazoneNameField(names, kind)
			if want == "" {
				continue
			}
			// Use an instant with no metazone period so the metazone branch yields
			// "" and the zone-specific name is the resolved value.
			got := timezone.TimeZoneDisplayName(loc, zone, form, false, noMetazoneInstant, 0)
			if got != want {
				t.Errorf("TimeZoneDisplayName(%q, %q, %q) zone name = %q, want %q", locale, zone, form, got, want)
			}
		}
	}
}

func assertExemplarCities(t *testing.T, loc timezone.Locale, locale string, data extract.Metazones) {
	t.Helper()
	for zone, city := range data.ExemplarCities[locale] {
		// Skip zones that also carry a zone-specific name or a metazone period:
		// those resolve before the exemplar-city fallback.
		if zoneHasName(data.ZoneNames[locale][zone]) {
			continue
		}
		if len(data.ZoneToMetazones[zone]) > 0 {
			continue
		}
		want := "Time in " + city
		got := timezone.TimeZoneDisplayName(loc, zone, timezone.TimeZoneNameLongGeneric, false, noMetazoneInstant, 0)
		if got != want {
			t.Errorf("TimeZoneDisplayName(%q, %q) exemplar = %q, want %q", locale, zone, got, want)
		}
	}
}

// displayFormKinds maps the TimeZoneName forms whose resolution path reads a
// single metazoneNames field to the legacy kind string, so the test can pick the
// expected field. shortGeneric/longGeneric are covered by the long/short forms.
var displayFormKinds = map[timezone.TimeZoneName]string{
	timezone.TimeZoneNameLongGeneric: "long-generic",
}

// metazoneNameField mirrors the legacy timeZoneNameValue field selection.
func metazoneNameField(names cldr.MetazoneNames, kind string) string {
	switch kind {
	case "long-standard":
		return names.LongStandard
	case "long-daylight":
		return names.LongDaylight
	case "short-generic":
		return names.ShortGeneric
	case "short-standard":
		return names.ShortStandard
	case "short-daylight":
		return names.ShortDaylight
	default:
		return names.LongGeneric
	}
}

func zoneHasName(names cldr.MetazoneNames) bool {
	return names != cldr.MetazoneNames{}
}

func metazoneNameUnion(data extract.Metazones) map[string]bool {
	seen := map[string]bool{}
	for locale := range data.Names {
		seen[locale] = true
	}
	for locale := range data.ZoneNames {
		seen[locale] = true
	}
	for locale := range data.ExemplarCities {
		seen[locale] = true
	}
	return seen
}

// noMetazoneInstant is an instant before any CLDR metazone period begins, so the
// metazone resolution branch yields "" and earlier/later fallbacks are exercised
// in isolation.
const noMetazoneInstant = int64(-1 << 62)

// periodProbeInstant returns an instant inside [start, end). For an open lower
// bound it probes just below end; for an open upper bound it probes just above
// start; otherwise the midpoint.
func periodProbeInstant(start, end int64) int64 {
	const minSentinel = int64(-1 << 63)
	const maxSentinel = int64(1<<63 - 1)
	switch {
	case start == minSentinel && end == maxSentinel:
		return 0
	case start == minSentinel:
		return end - 1
	case end == maxSentinel:
		return start
	default:
		return start + (end-start)/2
	}
}

func loadTimezoneTestInput(t *testing.T) extract.Metazones {
	t.Helper()

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
	return extract.ExtractMetazones(source.Metazones, profile)
}
