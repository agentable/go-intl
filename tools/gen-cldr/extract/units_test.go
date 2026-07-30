package extract

import (
	"maps"
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/unitid"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestExtractUnitsKeepsWellFormedUnitData(t *testing.T) {
	t.Parallel()

	rawUnits := cldr.Units{
		"hertz":              testUnitData("hertz"),
		"kilometer-per-hour": testUnitData("kilometer-per-hour"),
		"milligram":          testUnitData("milligram"),
		"newton":             testUnitData("newton"),
	}
	for _, unit := range unitid.SanctionedSimpleUnitIdentifiers() {
		rawUnits[unit] = testUnitData(unit)
	}

	got := ExtractUnits(map[string]cldr.Units{"en": rawUnits}, []string{"en"})
	gotUnits, ok := got["en"]
	if !ok {
		t.Fatal(`ExtractUnits(...) missing requested locale "en"`)
	}

	gotNames := slices.Sorted(maps.Keys(gotUnits))
	wantNames := append(unitid.SanctionedSimpleUnitIdentifiers(), "kilometer-per-hour")
	slices.Sort(wantNames)
	if !slices.Equal(gotNames, wantNames) {
		t.Fatalf("ExtractUnits(...) units = %v, want %v", gotNames, wantNames)
	}
	for _, unit := range []string{"hertz", "milligram", "newton"} {
		if _, ok := gotUnits[unit]; ok {
			t.Fatalf("ExtractUnits(...) kept unsanctioned unit %q", unit)
		}
	}
	if got := gotUnits["meter"].PerUnit["long"]; got != "{0} per meter" {
		t.Fatalf("ExtractUnits(...) meter per-unit pattern = %q, want %q", got, "{0} per meter")
	}
}

func TestExtractUnitsFiltersLocales(t *testing.T) {
	t.Parallel()

	got := ExtractUnits(map[string]cldr.Units{
		"en": {"meter": testUnitData("meter")},
		"fr": {"hertz": testUnitData("hertz")},
		"de": {"meter": testUnitData("meter")},
	}, []string{"en", "fr", "es"})

	if _, ok := got["de"]; ok {
		t.Fatal(`ExtractUnits(...) kept unrequested locale "de"`)
	}
	if _, ok := got["fr"]; ok {
		t.Fatal(`ExtractUnits(...) kept requested locale "fr" without sanctioned unit data`)
	}
	gotUnits, ok := got["en"]
	if !ok {
		t.Fatal(`ExtractUnits(...) missing requested locale "en"`)
	}
	gotNames := slices.Sorted(maps.Keys(gotUnits))
	if !slices.Equal(gotNames, []string{"meter"}) {
		t.Fatalf("ExtractUnits(...) en units = %v, want [meter]", gotNames)
	}
}

func testUnitData(unit string) cldr.UnitData {
	return cldr.UnitData{
		Patterns: map[string]map[string]map[string]string{
			"long": {
				unit: {"other": "{0} " + unit},
			},
		},
		PerUnit:  map[string]string{"long": "{0} per " + unit},
		Compound: map[string]string{"long": "{0} per {1}"},
	}
}
