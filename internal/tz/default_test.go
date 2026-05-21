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

func TestDefaultLocationFallsBackToUTCForNilLocal(t *testing.T) {
	t.Parallel()

	name, loc := defaultLocation(nil)
	if got, want := name, "UTC"; got != want {
		t.Fatalf("defaultLocation(nil) name = %q, want %q", got, want)
	}
	if loc != time.UTC {
		t.Fatalf("defaultLocation(nil) location = %v, want UTC", loc)
	}
}

func TestDefaultLocationReturnsUsableLocationForLocal(t *testing.T) {
	t.Parallel()

	name, loc := defaultLocation(time.Local)
	if name == "" {
		t.Fatal("defaultLocation(Local) name is empty")
	}
	if loc == nil {
		t.Fatal("defaultLocation(Local) location is nil")
	}
	if info := LookupAt(loc, time.Now().UTC()); info.Name == "" {
		t.Fatal("LookupAt(default local) name is empty")
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

func TestDefaultLocationPreservesFallbackForUnknownNamedLocation(t *testing.T) {
	t.Parallel()

	fallback := time.FixedZone("Mars/Phobos", 1234)
	name, loc := defaultLocation(fallback)
	if got, want := name, "Mars/Phobos"; got != want {
		t.Fatalf("defaultLocation(Mars/Phobos) name = %q, want %q", got, want)
	}
	if loc != fallback {
		t.Fatalf("defaultLocation(Mars/Phobos) location = %v, want original fallback", loc)
	}
}

func TestDefaultLocationPreservesFallbackForInvalidOffsetName(t *testing.T) {
	t.Parallel()

	fallback := time.FixedZone("+24:00", 0)
	name, loc := defaultLocation(fallback)
	if got, want := name, "+24:00"; got != want {
		t.Fatalf("defaultLocation(+24:00) name = %q, want %q", got, want)
	}
	if loc != fallback {
		t.Fatalf("defaultLocation(+24:00) location = %v, want original fallback", loc)
	}
}

func TestDefaultOverrideForTest(t *testing.T) {
	restore := OverrideDefaultForTest("Asia/Shanghai")
	t.Cleanup(restore)

	name, loc := Default()
	if got, want := name, "Asia/Shanghai"; got != want {
		t.Fatalf("Default() name = %q, want %q", got, want)
	}
	if got := LookupAt(loc, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)).OffsetMs; got != 8*3600*1000 {
		t.Fatalf("Default() offset = %d, want CST", got)
	}
}

func TestDefaultOverrideCanonicalizesLink(t *testing.T) {
	restore := OverrideDefaultForTest("US/Eastern")
	t.Cleanup(restore)

	name, _ := Default()
	if got, want := name, "America/New_York"; got != want {
		t.Fatalf("Default() name = %q, want %q", got, want)
	}
}
