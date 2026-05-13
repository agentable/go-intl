package cldr

import (
	"testing"

	"golang.org/x/text/language"
)

func TestMetazoneAccessors(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale(language.MustParse("en-US"))
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	if got, want := ZoneToMetazone("America/Los_Angeles"), "America_Pacific"; got != want {
		t.Fatalf("ZoneToMetazone(America/Los_Angeles) = %q, want %q", got, want)
	}
	if got, want := loc.MetazoneName("America_Pacific", "long-generic"), "Pacific Time"; got != want {
		t.Fatalf("MetazoneName(long-generic) = %q, want %q", got, want)
	}
	if got, want := loc.MetazoneName("America_Pacific", "short-standard"), "PST"; got != want {
		t.Fatalf("MetazoneName(short-standard) = %q, want %q", got, want)
	}
	if got, want := loc.ExemplarCity("Europe/Tirane"), "Tirana"; got != want {
		t.Fatalf("ExemplarCity(Europe/Tirane) = %q, want %q", got, want)
	}
}

func BenchmarkCLDR_TimeZoneDisplayName(b *testing.B) {
	loc, ok := ResolveLocale(language.MustParse("en-US"))
	if !ok {
		b.Fatal("ResolveLocale(en-US) ok=false")
	}
	const instant = int64(1735689600000)
	for b.Loop() {
		if got := TimeZoneDisplayName(loc, "America/New_York", TimeZoneNameLongGeneric, false, instant, -5*3600*1000); got == "" {
			b.Fatal("TimeZoneDisplayName returned empty string")
		}
	}
}
