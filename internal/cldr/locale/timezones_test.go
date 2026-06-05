package cldrlocale

import (
	"reflect"
	"testing"
)

func TestCanonicalTimeZoneLink(t *testing.T) {
	t.Parallel()

	if got, want := CanonicalTimeZoneLink("US/Eastern"), "America/New_York"; got != want {
		t.Fatalf("CanonicalTimeZoneLink(US/Eastern) = %q, want %q", got, want)
	}
	if got, want := CanonicalTimeZoneLink("America/New_York"), "America/New_York"; got != want {
		t.Fatalf("CanonicalTimeZoneLink(America/New_York) = %q, want %q", got, want)
	}
}

func TestTimeZonesForRegion(t *testing.T) {
	t.Parallel()

	if got, want := TimeZonesForRegion("GB"), []string{"Europe/London"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("TimeZonesForRegion(GB) = %#v, want %#v", got, want)
	}
	if got, want := TimeZonesForRegion("CN"), []string{"Asia/Shanghai", "Asia/Urumqi"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("TimeZonesForRegion(CN) = %#v, want %#v", got, want)
	}
	got := TimeZonesForRegion("GB")
	got[0] = "mutated"
	if again := TimeZonesForRegion("GB"); !reflect.DeepEqual(again, []string{"Europe/London"}) {
		t.Fatalf("TimeZonesForRegion returned shared storage: %#v", again)
	}
	if got := TimeZonesForRegion("ZZ"); got != nil {
		t.Fatalf("TimeZonesForRegion(ZZ) = %#v, want nil", got)
	}
}
