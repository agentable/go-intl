package tzdb

import (
	"os"
	"slices"
	"testing"
)

func TestPinnedRegistryWitnesses(t *testing.T) {
	t.Parallel()

	pin, err := ReadPin("../tzdata.json")
	if err != nil {
		t.Fatal(err)
	}
	archive := "../.tzdata/tzdata" + pin.Version + ".tar.gz"
	if _, err := os.Stat(archive); os.IsNotExist(err) {
		t.Skip("pinned tzdb archive not fetched")
	}
	aliases, err := LoadCLDRPrimaryAliases("../.cldr-json/node_modules/cldr-bcp47/bcp47/timezone.json")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := LoadArchive(archive, pin, aliases)
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]string, len(registry.Records))
	for _, record := range registry.Records {
		byID[record.Identifier] = record.Primary
	}
	for identifier, want := range map[string]string{
		"UTC":                "UTC",
		"Etc/UTC":            "UTC",
		"GMT":                "UTC",
		"US/Eastern":         "America/New_York",
		"Atlantic/Jan_Mayen": "Arctic/Longyearbyen",
		"Pacific/Truk":       "Pacific/Chuuk",
		"Europe/Kiev":        "Europe/Kyiv",
		"Europe/Kyiv":        "Europe/Kyiv",
		"Asia/Calcutta":      "Asia/Kolkata",
		"Asia/Kolkata":       "Asia/Kolkata",
	} {
		if got := byID[identifier]; got != want {
			t.Errorf("primary(%q) = %q, want %q", identifier, got, want)
		}
	}
	regions := make(map[string][]string, len(registry.Regions))
	for _, region := range registry.Regions {
		regions[region.Code] = region.Zones
	}
	if len(regions["CA"]) < 20 {
		t.Fatalf("CA zones = %d, want multi-zone country projection", len(regions["CA"]))
	}
	if !slices.Equal(regions["IN"], []string{"Asia/Kolkata"}) {
		t.Fatalf("IN zones = %#v, want Asia/Kolkata", regions["IN"])
	}
}

func TestBuildRegistryRejectsBrokenIdentityGraph(t *testing.T) {
	t.Parallel()

	base := func(backward string) map[string]string {
		return map[string]string{
			"africa":       "Zone Africa/One 0 - GMT\n",
			"antarctica":   "",
			"asia":         "",
			"australasia":  "",
			"europe":       "",
			"northamerica": "",
			"southamerica": "",
			"etcetera":     "Zone Etc/UTC 0 - UTC\nLink Etc/UTC UTC\n",
			"factory":      "",
			"backward":     backward,
			"zone.tab":     "AA\t+0000+00000\tAfrica/One\n",
			"backzone":     "# no overrides\n",
		}
	}
	for _, tc := range []struct {
		name     string
		backward string
	}{
		{name: "unknown target", backward: "Link Missing/Zone Old/Name\n"},
		{name: "cycle", backward: "Link Old/Two Old/One\nLink Old/One Old/Two\n"},
		{name: "case collision", backward: "Link Africa/One old/name\nLink Africa/One OLD/NAME\n"},
		{name: "non-ASCII identifier", backward: "Link Africa/One Old/\u212Aame\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := buildRegistry(base(tc.backward), nil); err == nil {
				t.Fatal("buildRegistry() succeeded, want error")
			}
		})
	}
}
