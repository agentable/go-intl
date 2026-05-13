package cldr

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/text/language"
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

func TestTimeZoneMetazoneUsesInstantPeriods(t *testing.T) {
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

func TestTimeZoneDisplayNameMetazone(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale(language.MustParse("en-US"))
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	instant := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	if got, want := TimeZoneMetazone("America/New_York", instant), "America_Eastern"; got != want {
		t.Fatalf("TimeZoneMetazone(America/New_York) = %q, want %q", got, want)
	}
	for _, tc := range []struct {
		name string
		zone string
		form TimeZoneName
		dst  bool
		want string
	}{
		{name: "long generic", zone: "America/New_York", form: TimeZoneNameLongGeneric, want: "Eastern Time"},
		{name: "long standard", zone: "America/New_York", form: TimeZoneNameLong, want: "Eastern Standard Time"},
		{name: "long daylight", zone: "America/New_York", form: TimeZoneNameLong, dst: true, want: "Eastern Daylight Time"},
		{name: "short standard", zone: "America/Los_Angeles", form: TimeZoneNameShort, want: "PST"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := TimeZoneDisplayName(loc, tc.zone, tc.form, tc.dst, instant, -5*3600*1000); got != tc.want {
				t.Fatalf("TimeZoneDisplayName(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestTimeZoneDisplayNameFallback(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale(language.MustParse("en-US"))
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	instant := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	if got, want := TimeZoneDisplayName(loc, "Mars/Olympus", TimeZoneNameShortGeneric, false, instant, -5*3600*1000), "GMT-5"; got != want {
		t.Fatalf("TimeZoneDisplayName(fallback) = %q, want %q", got, want)
	}
}

func TestGMTOffsetNameUsesLocaleFormats(t *testing.T) {
	t.Parallel()

	en, ok := ResolveLocale(language.MustParse("en-US"))
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	zh, ok := ResolveLocale(language.MustParse("zh-Hans-CN"))
	if !ok {
		t.Fatal("ResolveLocale(zh-Hans-CN) ok=false")
	}

	tests := []struct {
		name     string
		loc      Locale
		form     TimeZoneName
		offsetMs int64
		want     string
	}{
		{name: "short whole hour", loc: en, form: TimeZoneNameShortOffset, offsetMs: 2 * 3600 * 1000, want: "GMT+2"},
		{name: "long whole hour", loc: en, form: TimeZoneNameLongOffset, offsetMs: 2 * 3600 * 1000, want: "GMT+02:00"},
		{name: "short zero", loc: en, form: TimeZoneNameShortOffset, offsetMs: 0, want: "GMT"},
		{name: "long zero", loc: en, form: TimeZoneNameLongOffset, offsetMs: 0, want: "GMT+00:00"},
		{name: "short half hour", loc: en, form: TimeZoneNameShortOffset, offsetMs: 5*3600*1000 + 30*60*1000, want: "GMT+5:30"},
		{name: "short quarter hour", loc: en, form: TimeZoneNameShortOffset, offsetMs: 5*3600*1000 + 45*60*1000, want: "GMT+5:45"},
		{name: "zh active locale", loc: zh, form: TimeZoneNameShortOffset, offsetMs: -3*3600*1000 - 30*60*1000, want: "GMT-3:30"},
	}
	for _, tc := range tests {
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
