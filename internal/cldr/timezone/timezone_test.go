package timezone

import (
	"testing"
	"time"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

const missingLocale Locale = 65535

// TestSmokeMetazoneAccessors is a checkout-independent smoke test mirroring the
// deleted root metazone assertions, scoped to the timezone domain over the
// kernel handle. These values are hard-coded so a silent encoder/decoder
// regression fails here even when the FormatJS fixtures are unavailable.
func TestSmokeMetazoneAccessors(t *testing.T) {
	t.Parallel()

	loc, ok := cldrlocale.ResolveLocale("en-US")
	if !ok {
		t.Fatal(`ResolveLocale("en-US") = false, want true`)
	}
	if got, want := TimeZoneMetazone("America/Los_Angeles", 0), "America_Pacific"; got != want {
		t.Fatalf("TimeZoneMetazone(America/Los_Angeles, 0) = %q, want %q", got, want)
	}
	if got, want := metazoneName(loc, "America_Pacific", "long-generic"), "Pacific Time"; got != want {
		t.Fatalf("metazoneName(long-generic) = %q, want %q", got, want)
	}
	if got, want := metazoneName(loc, "America_Pacific", "short-standard"), "PST"; got != want {
		t.Fatalf("metazoneName(short-standard) = %q, want %q", got, want)
	}
	if got, want := exemplarCity(loc, "Europe/Tirane"), "Tirana"; got != want {
		t.Fatalf("exemplarCity(Europe/Tirane) = %q, want %q", got, want)
	}
	assertMissingLocalizedTimeZoneNames(t, missingLocale, "America_Pacific", "America/Los_Angeles")
	assertMissingLocalizedTimeZoneNames(t, loc, "Missing_Metazone", "Mars/Olympus")
}

func assertMissingLocalizedTimeZoneNames(t *testing.T, loc Locale, metazone, zone string) {
	t.Helper()

	if got := metazoneName(loc, metazone, zoneNameLongGeneric); got != "" {
		t.Fatalf("metazoneName(%d, %q) = %q, want empty", loc, metazone, got)
	}
	if got := zoneSpecificName(loc, zone, zoneNameLongGeneric); got != "" {
		t.Fatalf("zoneSpecificName(%d, %q) = %q, want empty", loc, zone, got)
	}
	if got := exemplarCity(loc, zone); got != "" {
		t.Fatalf("exemplarCity(%d, %q) = %q, want empty", loc, zone, got)
	}
}

// TestSmokeTimeZoneMetazoneInstants mirrors the deleted root instant-period test.
func TestSmokeTimeZoneMetazoneInstants(t *testing.T) {
	t.Parallel()

	winter2010 := time.Date(2010, time.January, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	summer2012 := time.Date(2012, time.July, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	if got, want := TimeZoneMetazone("Europe/Moscow", winter2010), "Moscow"; got != want {
		t.Fatalf("TimeZoneMetazone(Europe/Moscow, 2010) = %q, want %q", got, want)
	}
	if got, want := TimeZoneMetazone("Europe/Moscow", summer2012), "Moscow"; got != want {
		t.Fatalf("TimeZoneMetazone(Europe/Moscow, 2012) = %q, want %q", got, want)
	}
}

func TestMissingTimeZoneMetazoneReturnsEmpty(t *testing.T) {
	t.Parallel()

	if got := TimeZoneMetazone("Mars/Olympus", 0); got != "" {
		t.Fatalf("TimeZoneMetazone(Mars/Olympus, 0) = %q, want empty", got)
	}
}

// TestSmokeTimeZoneDisplayName mirrors the deleted root display-name resolution
// assertions across the metazone, zone-specific, and GMT fallback branches.
func TestSmokeTimeZoneDisplayName(t *testing.T) {
	t.Parallel()

	enUS, ok := cldrlocale.ResolveLocale("en-US")
	if !ok {
		t.Fatal(`ResolveLocale("en-US") = false, want true`)
	}
	enGB, ok := cldrlocale.ResolveLocale("en-GB")
	if !ok {
		t.Fatal(`ResolveLocale("en-GB") = false, want true`)
	}
	instant := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	if got, want := TimeZoneMetazone("America/New_York", instant), "America_Eastern"; got != want {
		t.Fatalf("TimeZoneMetazone(America/New_York) = %q, want %q", got, want)
	}

	for _, tc := range []struct {
		name string
		loc  Locale
		zone string
		form TimeZoneName
		dst  bool
		off  int64
		want string
	}{
		{"long generic", enUS, "America/New_York", TimeZoneNameLongGeneric, false, -5 * 3600 * 1000, "Eastern Time"},
		{"long standard", enUS, "America/New_York", TimeZoneNameLong, false, -5 * 3600 * 1000, "Eastern Standard Time"},
		{"long daylight", enUS, "America/New_York", TimeZoneNameLong, true, -5 * 3600 * 1000, "Eastern Daylight Time"},
		{"short standard", enUS, "America/Los_Angeles", TimeZoneNameShort, false, -5 * 3600 * 1000, "PST"},
		{"zone specific daylight", enGB, "Europe/London", TimeZoneNameLong, true, 3600 * 1000, "British Summer Time"},
		{"short offset", enUS, "America/New_York", TimeZoneNameShortOffset, false, -5 * 3600 * 1000, "GMT-5"},
		{"long offset", enUS, "America/New_York", TimeZoneNameLongOffset, false, -5 * 3600 * 1000, "GMT-05:00"},
		{"gmt fallback", enUS, "Mars/Olympus", TimeZoneNameShortGeneric, false, -5 * 3600 * 1000, "GMT-5"},
		{"missing locale fallback", missingLocale, "America/New_York", TimeZoneNameShortGeneric, false, -5 * 3600 * 1000, "GMT-5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gbInstant := instant
			if tc.name == "zone specific daylight" {
				gbInstant = time.Date(2021, time.June, 10, 12, 0, 0, 0, time.UTC).UnixMilli()
			}
			if got := TimeZoneDisplayName(tc.loc, tc.zone, tc.form, tc.dst, gbInstant, tc.off); got != tc.want {
				t.Fatalf("TimeZoneDisplayName(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

// TestSmokeGMTOffsetName mirrors the deleted root GMT-offset assertions.
func TestSmokeGMTOffsetName(t *testing.T) {
	t.Parallel()

	en, ok := cldrlocale.ResolveLocale("en-US")
	if !ok {
		t.Fatal(`ResolveLocale("en-US") = false, want true`)
	}
	zh, ok := cldrlocale.ResolveLocale("zh-Hans-CN")
	if !ok {
		t.Fatal(`ResolveLocale("zh-Hans-CN") = false, want true`)
	}

	for _, tc := range []struct {
		name     string
		loc      Locale
		form     TimeZoneName
		offsetMs int64
		want     string
	}{
		{"short whole hour", en, TimeZoneNameShortOffset, 2 * 3600 * 1000, "GMT+2"},
		{"long whole hour", en, TimeZoneNameLongOffset, 2 * 3600 * 1000, "GMT+02:00"},
		{"short zero", en, TimeZoneNameShortOffset, 0, "GMT"},
		{"long zero", en, TimeZoneNameLongOffset, 0, "GMT+00:00"},
		{"short half hour", en, TimeZoneNameShortOffset, 5*3600*1000 + 30*60*1000, "GMT+5:30"},
		{"short quarter hour", en, TimeZoneNameShortOffset, 5*3600*1000 + 45*60*1000, "GMT+5:45"},
		{"zh active locale", zh, TimeZoneNameShortOffset, -3*3600*1000 - 30*60*1000, "GMT-3:30"},
		{"missing locale short", missingLocale, TimeZoneNameShortOffset, -5 * 3600 * 1000, "GMT-5"},
		{"missing locale long", missingLocale, TimeZoneNameLongOffset, -5 * 3600 * 1000, "GMT-05:00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := GMTOffsetName(tc.loc, tc.offsetMs, tc.form); got != tc.want {
				t.Fatalf("GMTOffsetName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestOffsetPatternKeepsLocalizedMinus(t *testing.T) {
	t.Parallel()

	if got, want := offsetPattern("+HH:mm;−HH:mm", -3*3600*1000-30*60*1000, false), "−3:30"; got != want {
		t.Fatalf("offsetPattern(localized minus) = %q, want %q", got, want)
	}
}
