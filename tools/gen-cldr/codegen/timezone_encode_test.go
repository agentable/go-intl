package codegen

import (
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
		RegionFormat:  "dddd",
	}, table)

	assertBytesEqual(t, "appendTimeZoneFormats() bytes", e.bytes(), []byte{0, 1, 1, 2, 3, 3, 6, 4})
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
