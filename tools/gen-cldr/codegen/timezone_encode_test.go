package codegen

import (
	"fmt"
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func TestAppendMetazoneNames(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	appendMetazoneNames(&e, cldr.MetazoneNames{
		LongGeneric:   "a",
		LongStandard:  "bb",
		LongDaylight:  "ccc",
		ShortGeneric:  "dddd",
		ShortStandard: "eeeee",
		ShortDaylight: "ffffff",
	}, table)

	assertBytesEqual(t, "appendMetazoneNames() bytes", e.bytes(), []byte{0, 1, 1, 2, 3, 3, 6, 4, 10, 5, 15, 6})
}

func TestAppendTimeZoneFormats(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	appendTimeZoneFormats(&e, cldr.TimeZoneFormats{
		GMTFormat:     "a",
		GMTZeroFormat: "bb",
		HourFormat:    "ccc",
	}, table)

	assertBytesEqual(t, "appendTimeZoneFormats() bytes", e.bytes(), []byte{0, 1, 1, 2, 3, 3})
}

func TestMetazoneNameLocalesOwnsNameLocaleUnion(t *testing.T) {
	t.Parallel()

	data := extract.Metazones{
		Names: map[string]map[string]cldr.MetazoneNames{
			"fr": nil,
			"en": nil,
		},
		ZoneNames: map[string]map[string]cldr.MetazoneNames{
			"de": nil,
			"en": nil,
		},
		ExemplarCities: map[string]map[string]string{
			"ja": nil,
			"fr": nil,
		},
	}

	got := metazoneNameLocales(data)
	want := []string{"de", "en", "fr", "ja"}
	assertStringSliceEqual(t, "metazoneNameLocales()", got, want)
}

func TestCanonicalTimeZoneLinksDriveEncoderAndRuntimeRender(t *testing.T) {
	t.Parallel()

	src, err := renderTimezones()
	if err != nil {
		t.Fatalf("renderTimezones() error = %v", err)
	}
	rendered := string(src)
	assertSourceContains(t, "renderTimezones() output", rendered, "type regionTimeZonesRecord struct")
	assertSourceContains(t, "renderTimezones() output", rendered, "var timeZonesByRegion = [...]regionTimeZonesRecord")
	assertSourceContains(t, "renderTimezones() output", rendered, "switch name")
	assertSourceContains(t, "renderTimezones() output", rendered, "for _, record := range timeZonesByRegion")

	for _, link := range canonicalTimeZoneLinks[:] {
		if got := canonicalTimeZoneLink(link.alias); got != link.canonical {
			t.Errorf("canonicalTimeZoneLink(%q) = %q, want %q", link.alias, got, link.canonical)
		}
		assertSourceContains(t, "renderTimezones() output", rendered, fmt.Sprintf("case %q:", link.alias))
		assertSourceContains(t, "renderTimezones() output", rendered, fmt.Sprintf("return %q", link.canonical))
	}

	const unknown = "Europe/Paris"
	if got := canonicalTimeZoneLink(unknown); got != unknown {
		t.Errorf("canonicalTimeZoneLink(%q) = %q, want identity", unknown, got)
	}
}
