package tz

import (
	"slices"
	"strings"
	"testing"
)

func TestIdentifierRegistryContract(t *testing.T) {
	t.Parallel()

	if timeZoneDataVersion != "2025b" || len(timeZoneDataSHA256) != 64 {
		t.Fatalf("time-zone data pin = %q/%q, want 2025b and SHA-256", timeZoneDataVersion, timeZoneDataSHA256)
	}
	previous := ""
	seen := make(map[string]struct{}, len(timeZoneIdentifierRecords))
	for _, record := range timeZoneIdentifierRecords {
		if record.Identifier <= previous {
			t.Fatalf("identifier records not strictly sorted at %q after %q", record.Identifier, previous)
		}
		previous = record.Identifier
		lower := strings.ToLower(record.Identifier)
		if _, ok := seen[lower]; ok {
			t.Fatalf("duplicate case-folded identifier %q", record.Identifier)
		}
		seen[lower] = struct{}{}
		primary, ok := LookupIdentifier(record.Primary)
		if !ok || primary.Identifier != record.Primary || primary.Primary != record.Primary {
			t.Fatalf("record %+v has non-idempotent primary %+v, found %t", record, primary, ok)
		}
		lookup, ok := LookupIdentifier(strings.ToLower(record.Identifier))
		if !ok || lookup != record {
			t.Fatalf("case-insensitive lookup for %q = %+v, %t", record.Identifier, lookup, ok)
		}
	}
}

func TestIdentifierRegistryWitnesses(t *testing.T) {
	t.Parallel()

	for input, want := range map[string]string{
		"utc":                "UTC",
		"EtC/uTc":            "UTC",
		"us/eastern":         "America/New_York",
		"atlantic/jan_mayen": "Arctic/Longyearbyen",
		"pacific/truk":       "Pacific/Chuuk",
		"europe/kiev":        "Europe/Kyiv",
		"europe/kyiv":        "Europe/Kyiv",
		"asia/calcutta":      "Asia/Kolkata",
	} {
		record, ok := LookupIdentifier(input)
		if !ok || record.Primary != want {
			t.Errorf("LookupIdentifier(%q) = %+v, %t; want primary %q", input, record, ok, want)
		}
	}
	if _, ok := LookupIdentifier("Mars/Olympus"); ok {
		t.Fatal("LookupIdentifier(Mars/Olympus) succeeded")
	}
	if _, ok := LookupIdentifier("Asia/Kolkata"); ok {
		t.Fatal("LookupIdentifier accepted a non-ASCII case-fold lookalike")
	}
}

func TestSupportedTimeZonesAreExactPrimaryProjection(t *testing.T) {
	t.Parallel()

	got := SupportedTimeZones()
	if !slices.IsSorted(got) {
		t.Fatal("SupportedTimeZones() is not sorted")
	}
	for _, identifier := range got {
		record, ok := LookupIdentifier(identifier)
		if !ok || record.Identifier != record.Primary {
			t.Fatalf("SupportedTimeZones() contains non-primary %q: %+v", identifier, record)
		}
	}
	if len(got) == 0 {
		t.Fatal("SupportedTimeZones() is empty")
	}
	got[0] = "mutated"
	if SupportedTimeZones()[0] == "mutated" {
		t.Fatal("SupportedTimeZones() returned shared storage")
	}
}

func TestTimeZonesForRegionContract(t *testing.T) {
	t.Parallel()

	for _, region := range timeZoneRegionRecords {
		got := TimeZonesForRegion(region.region)
		if !slices.Equal(got, region.zones) || !slices.IsSorted(got) {
			t.Fatalf("TimeZonesForRegion(%q) = %#v, want sorted %#v", region.region, got, region.zones)
		}
		for _, identifier := range got {
			record, ok := LookupIdentifier(identifier)
			if !ok || record.Identifier != record.Primary {
				t.Fatalf("region %q contains non-primary %q: %+v", region.region, identifier, record)
			}
		}
		if len(got) > 0 {
			got[0] = "mutated"
			if TimeZonesForRegion(region.region)[0] == "mutated" {
				t.Fatalf("TimeZonesForRegion(%q) returned shared storage", region.region)
			}
		}
	}
	if got := TimeZonesForRegion("ZZ"); got != nil {
		t.Fatalf("TimeZonesForRegion(ZZ) = %#v, want nil", got)
	}
	if got := TimeZonesForRegion("IN"); !slices.Equal(got, []string{"Asia/Kolkata"}) {
		t.Fatalf("TimeZonesForRegion(IN) = %#v, want Asia/Kolkata", got)
	}
	if got := TimeZonesForRegion("CA"); len(got) < 20 {
		t.Fatalf("TimeZonesForRegion(CA) = %d zones, want multi-zone projection", len(got))
	}
}
