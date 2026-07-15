package tz

import (
	"errors"
	"testing"
	"time"
)

func TestResolveFixedOffset(t *testing.T) {
	t.Parallel()

	loc, err := Resolve("+05:30")
	if err != nil {
		t.Fatalf("Resolve(+05:30) error = %v", err)
	}
	info := LookupAt(loc, time.Unix(0, 0).UTC())
	if got, want := info.OffsetMs, int64(5*3600*1000+30*60*1000); got != want {
		t.Fatalf("LookupAt fixed offset OffsetMs = %d, want %d", got, want)
	}
	if info.IsDST {
		t.Fatal("LookupAt fixed offset IsDST = true, want false")
	}
}

func TestResolveCanonicalizesEquivalentFixedOffsets(t *testing.T) {
	t.Parallel()

	colon, err := Resolve("+05:30")
	if err != nil {
		t.Fatalf("Resolve(+05:30) error = %v", err)
	}
	compact, err := Resolve("+0530")
	if err != nil {
		t.Fatalf("Resolve(+0530) error = %v", err)
	}
	if colon != compact {
		t.Fatal("Resolve(+05:30) and Resolve(+0530) returned different cached locations")
	}
	if got := compact.String(); got != "+05:30" {
		t.Fatalf("Resolve(+0530).String() = %q, want +05:30", got)
	}
}

func TestResolveIANAZoneUsesDST(t *testing.T) {
	t.Parallel()

	loc, err := Resolve("America/New_York")
	if err != nil {
		t.Fatalf("Resolve(America/New_York) error = %v", err)
	}
	summer := LookupAt(loc, time.Date(2025, time.July, 1, 12, 0, 0, 0, time.UTC))
	winter := LookupAt(loc, time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC))
	if summer.OffsetMs == winter.OffsetMs {
		t.Fatalf("summer offset = winter offset = %d, want DST difference", summer.OffsetMs)
	}
	if !summer.IsDST {
		t.Fatal("summer IsDST = false, want true")
	}
	if winter.IsDST {
		t.Fatal("winter IsDST = true, want false")
	}
	if summer.Abbrv == "" || winter.Abbrv == "" {
		t.Fatalf("abbreviations = %q/%q, want non-empty", summer.Abbrv, winter.Abbrv)
	}
}

func TestResolveSouthernHemisphereZoneUsesDST(t *testing.T) {
	t.Parallel()

	loc, err := Resolve("Australia/Sydney")
	if err != nil {
		t.Fatalf("Resolve(Australia/Sydney) error = %v", err)
	}
	summer := LookupAt(loc, time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC))
	winter := LookupAt(loc, time.Date(2025, time.July, 1, 12, 0, 0, 0, time.UTC))
	if summer.OffsetMs == winter.OffsetMs {
		t.Fatalf("summer offset = winter offset = %d, want DST difference", summer.OffsetMs)
	}
	if !summer.IsDST {
		t.Fatal("southern hemisphere summer IsDST = false, want true")
	}
	if winter.IsDST {
		t.Fatal("southern hemisphere winter IsDST = true, want false")
	}
}

func TestLookupAtUsesTZDataTransitionDSTFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		zone    string
		instant time.Time
		want    bool
	}{
		{name: "New York winter", zone: "America/New_York", instant: time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC), want: false},
		{name: "New York summer", zone: "America/New_York", instant: time.Date(2025, time.July, 1, 12, 0, 0, 0, time.UTC), want: true},
		{name: "Sydney summer", zone: "Australia/Sydney", instant: time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC), want: true},
		{name: "Sydney winter", zone: "Australia/Sydney", instant: time.Date(2025, time.July, 1, 12, 0, 0, 0, time.UTC), want: false},
		{name: "Casablanca before Ramadan", zone: "Africa/Casablanca", instant: time.Date(2024, time.February, 1, 12, 0, 0, 0, time.UTC), want: true},
		{name: "Casablanca during Ramadan", zone: "Africa/Casablanca", instant: time.Date(2024, time.March, 20, 12, 0, 0, 0, time.UTC), want: false},
		{name: "Casablanca after Ramadan", zone: "Africa/Casablanca", instant: time.Date(2024, time.April, 20, 12, 0, 0, 0, time.UTC), want: true},
		{name: "London wartime winter", zone: "Europe/London", instant: time.Date(1941, time.January, 15, 12, 0, 0, 0, time.UTC), want: true},
		{name: "London postwar winter", zone: "Europe/London", instant: time.Date(1946, time.January, 15, 12, 0, 0, 0, time.UTC), want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, err := Resolve(tc.zone)
			if err != nil {
				t.Fatalf("Resolve(%q) error = %v", tc.zone, err)
			}
			if got := LookupAt(loc, tc.instant).IsDST; got != tc.want {
				t.Fatalf("LookupAt(%q, %s).IsDST = %t, want %t", tc.zone, tc.instant.Format(time.RFC3339), got, tc.want)
			}
		})
	}
}

func TestCanonicalLink(t *testing.T) {
	t.Parallel()

	if got, want := CanonicalLink("US/Eastern"), "America/New_York"; got != want {
		t.Fatalf("CanonicalLink(US/Eastern) = %q, want %q", got, want)
	}
	if got, want := CanonicalLink("America/Montreal"), "America/Toronto"; got != want {
		t.Fatalf("CanonicalLink(America/Montreal) = %q, want %q", got, want)
	}
	if got, want := CanonicalLink("America/New_York"), "America/New_York"; got != want {
		t.Fatalf("CanonicalLink(America/New_York) = %q, want %q", got, want)
	}
}

func TestResolveIanaLink(t *testing.T) {
	t.Parallel()

	link, err := Resolve("US/Eastern")
	if err != nil {
		t.Fatalf("Resolve(US/Eastern) error = %v", err)
	}
	canonical, err := Resolve("America/New_York")
	if err != nil {
		t.Fatalf("Resolve(America/New_York) error = %v", err)
	}
	instant := time.Date(2025, time.July, 1, 12, 0, 0, 0, time.UTC)
	if got, want := LookupAt(link, instant).OffsetMs, LookupAt(canonical, instant).OffsetMs; got != want {
		t.Fatalf("Resolve(US/Eastern) offset = %d, want %d", got, want)
	}
}

func TestResolveIanaLinkUsesCanonicalCache(t *testing.T) {
	t.Parallel()

	link, err := Resolve("US/Eastern")
	if err != nil {
		t.Fatalf("Resolve(US/Eastern) error = %v", err)
	}
	canonical, err := Resolve("America/New_York")
	if err != nil {
		t.Fatalf("Resolve(America/New_York) error = %v", err)
	}
	if link != canonical {
		t.Fatal("Resolve(US/Eastern) and Resolve(America/New_York) returned different cached locations")
	}
}

func TestResolveUsesCaseInsensitiveRegistryPrimary(t *testing.T) {
	t.Parallel()

	for input, primary := range map[string]string{
		"us/eastern":         "America/New_York",
		"ATLANTIC/JAN_MAYEN": "Arctic/Longyearbyen",
		"pacific/truk":       "Pacific/Chuuk",
		"europe/kiev":        "Europe/Kyiv",
		"etc/utc":            "UTC",
	} {
		loc, err := Resolve(input)
		if err != nil {
			t.Errorf("Resolve(%q) error = %v", input, err)
			continue
		}
		if loc.String() != primary {
			t.Errorf("Resolve(%q).String() = %q, want %q", input, loc, primary)
		}
		if got := CanonicalLink(input); got != primary {
			t.Errorf("CanonicalLink(%q) = %q, want %q", input, got, primary)
		}
	}
}

func TestResolveCLDRTimeZoneAliasUsesCanonicalCache(t *testing.T) {
	t.Parallel()

	link, err := Resolve("America/Montreal")
	if err != nil {
		t.Fatalf("Resolve(America/Montreal) error = %v", err)
	}
	canonical, err := Resolve("America/Toronto")
	if err != nil {
		t.Fatalf("Resolve(America/Toronto) error = %v", err)
	}
	if link != canonical {
		t.Fatal("Resolve(America/Montreal) and Resolve(America/Toronto) returned different cached locations")
	}
}

func TestResolveUnknownZone(t *testing.T) {
	t.Parallel()

	_, err := Resolve("Mars/Olympus")
	if !errors.Is(err, ErrUnsupportedTimeZone) {
		t.Fatalf("Resolve(Mars/Olympus) error = %v, want ErrUnsupportedTimeZone", err)
	}
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("Resolve(Mars/Olympus) error = %v, want errors.ErrUnsupported", err)
	}
	detail, ok := errors.AsType[unsupportedTimeZoneError](err)
	if !ok {
		t.Fatalf("Resolve(Mars/Olympus) error = %T, want unsupportedTimeZoneError", err)
	}
	if detail.name != "Mars/Olympus" {
		t.Fatalf("unsupportedTimeZoneError.name = %q, want Mars/Olympus", detail.name)
	}
}

func TestResolveRejectsHostLoadableNameAbsentFromRegistry(t *testing.T) {
	t.Parallel()

	if _, err := time.LoadLocation("posixrules"); err != nil {
		t.Skipf("host does not expose posixrules: %v", err)
	}
	_, err := Resolve("posixrules")
	if !errors.Is(err, ErrUnsupportedTimeZone) {
		t.Fatalf("Resolve(posixrules) error = %v, want registry rejection", err)
	}
}

func TestResolveRejectsInvalidFixedOffset(t *testing.T) {
	t.Parallel()

	_, err := Resolve("+24:00")
	if !errors.Is(err, ErrUnsupportedTimeZone) {
		t.Fatalf("Resolve(+24:00) error = %v, want ErrUnsupportedTimeZone", err)
	}
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("Resolve(+24:00) error = %v, want errors.ErrUnsupported", err)
	}
	detail, ok := errors.AsType[unsupportedTimeZoneError](err)
	if !ok {
		t.Fatalf("Resolve(+24:00) error = %T, want unsupportedTimeZoneError", err)
	}
	if detail.reason != "invalid offset" || detail.name != "+24:00" {
		t.Fatalf("unsupportedTimeZoneError = %+v, want invalid offset +24:00", detail)
	}
}

func BenchmarkTZ_Resolve(b *testing.B) {
	for _, name := range []string{"America/New_York", "US/Eastern", "+05:30"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				loc, err := Resolve(name)
				if err != nil {
					b.Fatal(err)
				}
				if loc == nil {
					b.Fatal("Resolve returned nil location")
				}
			}
		})
	}
}
