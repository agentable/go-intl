package tz

import (
	"errors"
	"strings"
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

func TestCanonicalLink(t *testing.T) {
	t.Parallel()

	if got, want := CanonicalLink("US/Eastern"), "America/New_York"; got != want {
		t.Fatalf("CanonicalLink(US/Eastern) = %q, want %q", got, want)
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

func TestResolveUnknownZone(t *testing.T) {
	t.Parallel()

	_, err := Resolve("Mars/Olympus")
	if !errors.Is(err, ErrUnsupportedTimeZone) {
		t.Fatalf("Resolve(Mars/Olympus) error = %v, want ErrUnsupportedTimeZone", err)
	}
	if !strings.Contains(err.Error(), "Mars/Olympus") {
		t.Fatalf("Resolve(Mars/Olympus) error = %v, want input in message", err)
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
