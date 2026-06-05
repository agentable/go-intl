package timezone

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no main-blob decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_TIMEZONE_NARROW_INDEX_SUBPROCESS"

// TestSupportedTimeZonesDoesNotDecodeMainBlobs asserts the narrow-index rule:
// SupportedTimeZones reads only the supported blob and must never trigger the
// metazone-period, names, or formats decode.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// metazone, name, or format lookup populates those package-level maps, after
// which a same-binary assertion would see them already decoded. Rather than
// depend on test order, this test re-executes the test binary as a subprocess
// that runs only the inner assertion via -test.run, so the assertion always
// observes a process that has decoded nothing.
func TestSupportedTimeZonesDoesNotDecodeMainBlobs(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertSupportedTimeZonesDoesNotDecodeMainBlobs(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSupportedTimeZonesDoesNotDecodeMainBlobs$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertSupportedTimeZonesDoesNotDecodeMainBlobs(t *testing.T) {
	t.Helper()

	tags := SupportedTimeZones()
	if len(tags) == 0 {
		t.Fatal("SupportedTimeZones returned no tags")
	}
	if zoneToMetazones != nil {
		t.Error("SupportedTimeZones decoded the metazone-period blob; narrow index must not")
	}
	if metazoneNamesByLocale != nil || timeZoneNamesByLocale != nil || exemplarCitiesByLocale != nil {
		t.Error("SupportedTimeZones decoded the names blob; narrow index must not")
	}
	if timeZoneFormatsByLocale != nil {
		t.Error("SupportedTimeZones decoded the formats blob; narrow index must not")
	}
}

func TestSupportedTimeZonesReturnsCopy(t *testing.T) {
	t.Parallel()

	a := SupportedTimeZones()
	if len(a) == 0 {
		t.Fatal("SupportedTimeZones returned no tags")
	}
	a[0] = "mutated"
	b := SupportedTimeZones()
	if b[0] == "mutated" {
		t.Error("SupportedTimeZones returned a shared slice; callers can corrupt the cache")
	}
}

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
	if got, want := ZoneToMetazone("America/Los_Angeles"), "America_Pacific"; got != want {
		t.Fatalf("ZoneToMetazone(America/Los_Angeles) = %q, want %q", got, want)
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
		{"gmt fallback", enUS, "Mars/Olympus", TimeZoneNameShortGeneric, false, -5 * 3600 * 1000, "GMT-5"},
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

func TestCanonicalTimeZoneLink(t *testing.T) {
	t.Parallel()

	if got, want := CanonicalTimeZoneLink("US/Eastern"), "America/New_York"; got != want {
		t.Fatalf("CanonicalTimeZoneLink(US/Eastern) = %q, want %q", got, want)
	}
	if got, want := CanonicalTimeZoneLink("America/New_York"), "America/New_York"; got != want {
		t.Fatalf("CanonicalTimeZoneLink(America/New_York) = %q, want %q", got, want)
	}
}
