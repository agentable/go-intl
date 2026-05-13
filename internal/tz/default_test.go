package tz

import (
	"testing"
	"time"
)

func TestDefaultLocationUsesIANAName(t *testing.T) {
	t.Parallel()

	local, err := time.LoadLocation("US/Eastern")
	if err != nil {
		t.Fatal(err)
	}
	name, loc := defaultLocation(local)
	if got, want := name, "America/New_York"; got != want {
		t.Fatalf("defaultLocation(US/Eastern) name = %q, want %q", got, want)
	}
	if got := LookupAt(loc, time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)).OffsetMs; got != -5*3600*1000 {
		t.Fatalf("defaultLocation(US/Eastern) offset = %d, want EST", got)
	}
}

func TestDefaultLocationUsesFixedOffsetName(t *testing.T) {
	t.Parallel()

	name, loc := defaultLocation(time.FixedZone("+05:30", 5*3600+30*60))
	if got, want := name, "+05:30"; got != want {
		t.Fatalf("defaultLocation(+05:30) name = %q, want %q", got, want)
	}
	if got := LookupAt(loc, time.Unix(0, 0)).OffsetMs; got != (5*3600+30*60)*1000 {
		t.Fatalf("defaultLocation(+05:30) offset = %d, want +05:30", got)
	}
}
