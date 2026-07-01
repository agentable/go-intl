package codegen

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/agentable/go-intl/internal/cldr/unit"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestUnitRoundTrip is the production-path round-trip gate for the unit domain.
// It re-derives the extract.Units row stream from the real pinned CLDR checkout
// and asserts that every simple pattern, compound pattern, and supported-locale
// tag the encoder wrote is queried back byte-for-byte through the production
// unit accessors over the committed data.go. It exercises encoder, blob,
// decoder, and accessor as one chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips, since there is no
// data to round-trip against.
func TestUnitRoundTrip(t *testing.T) {
	t.Parallel()

	input := loadRoundTripSource(t)
	units := extract.ExtractUnits(input.source.Units, input.profile)

	// Simple unit patterns.
	for _, row := range unitPatternRows(units) {
		loc := resolveKernelLocale(t, row.locale)
		got := unit.UnitPattern(loc, row.unit, row.width, row.plural)
		if got != row.pattern {
			t.Errorf("UnitPattern(%q, %q, %q, %q) = %q, want %q",
				row.locale, row.unit, row.width, row.plural, got, row.pattern)
		}
	}

	// Compound unit patterns.
	for _, row := range compoundUnitPatternRows(units) {
		loc := resolveKernelLocale(t, row.locale)
		got := unit.CompoundUnitPattern(loc, row.width)
		if got != row.pattern {
			t.Errorf("CompoundUnitPattern(%q, %q) = %q, want %q",
				row.locale, row.width, got, row.pattern)
		}
	}

	// Supported-locale narrow index.
	wantTags := sortedLocaleKeys(units)
	gotTags := unit.SupportedLocales()
	assertStringSliceEqual(t, "SupportedLocales", gotTags, wantTags)
}

func TestEncodeUnitsReturnsErrorsForInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input RuntimeInput
		want  string
	}{
		{
			name:  "too many unit IDs",
			input: unitRuntimeInput(unitTestData("en", unitTestNames(256)...)),
			want:  "unit name ID table has 256 entries",
		},
		{
			name:  "too many locale tags",
			input: unitRuntimeInputWithLocales(unitTestData("en", "meter"), unitTestLocaleTags(unitPatternLocaleTagMax+1)),
			want:  "locale table has 65537 tags",
		},
		{
			name:  "unit locale missing from kernel",
			input: unitRuntimeInput(unitTestData("fr", "meter")),
			want:  `locale "fr" missing from kernel locale registry`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := encodeUnits(tt.input, NewStringTable())
			assertErrorContains(t, "encodeUnits()", err, tt.want)
		})
	}
}

func TestUnitPatternOrdinals(t *testing.T) {
	t.Parallel()

	assertStringSliceEqual(t, "unitWidthOrder", unitWidthOrder[:], []string{"long", "narrow", "short"})
	assertStringSliceEqual(t, "unitPluralOrder", unitPluralOrder[:], []string{"few", "many", "one", "other", "two", "zero"})
	for i, width := range unitWidthOrder[:] {
		want := uint64(i + 1)
		got, ok := unitWidthOrdinal(width)
		if !ok || got != want {
			t.Fatalf("unitWidthOrdinal(%q) = %d, %t, want %d, true", width, got, ok, want)
		}
	}
	for i, plural := range unitPluralOrder[:] {
		want := uint64(i + 1)
		got, ok := unitPluralOrdinal(plural)
		if !ok || got != want {
			t.Fatalf("unitPluralOrdinal(%q) = %d, %t, want %d, true", plural, got, ok, want)
		}
	}
	if got, ok := unitWidthOrdinal("bad"); got != 0 || ok {
		t.Fatalf("unitWidthOrdinal(%q) = %d, %t, want 0, false", "bad", got, ok)
	}
	if got, ok := unitPluralOrdinal("bad"); got != 0 || ok {
		t.Fatalf("unitPluralOrdinal(%q) = %d, %t, want 0, false", "bad", got, ok)
	}
}

func TestUnitPatternKeyLayout(t *testing.T) {
	t.Parallel()

	layout := []struct {
		name string
		got  int
		want int
	}{
		{name: "locale shift", got: unitPatternLocaleShift, want: 16},
		{name: "unit shift", got: unitPatternUnitShift, want: 8},
		{name: "width shift", got: unitPatternWidthShift, want: 4},
		{name: "unit id max", got: unitPatternUnitIDMax, want: 255},
		{name: "locale tag max", got: unitPatternLocaleTagMax, want: 65536},
		{name: "locale id max", got: unitPatternLocaleIDMax, want: 65535},
	}
	for _, tc := range layout {
		if tc.got != tc.want {
			t.Fatalf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}

	got, err := unitPatternKeyValue(
		map[string]uint64{"en": 2},
		map[string]int{"meter": 7},
		unitPatternRow{locale: "en", unit: "meter", width: "narrow", plural: "other"},
	)
	if err != nil {
		t.Fatalf("unitPatternKeyValue() error = %v", err)
	}
	if want := uint32(0x00020724); got != want {
		t.Fatalf("unitPatternKeyValue() = %#08x, want %#08x", got, want)
	}

	compound, err := compoundUnitKeyValue(
		map[string]uint64{"en": 2},
		compoundUnitPatternRow{locale: "en", width: "narrow"},
	)
	if err != nil {
		t.Fatalf("compoundUnitKeyValue() error = %v", err)
	}
	if want := uint32(0x00000022); compound != want {
		t.Fatalf("compoundUnitKeyValue() = %#08x, want %#08x", compound, want)
	}
}

func TestReadRoundTripTestProfileUsesSharedProfileContract(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "locale-profile.json")
	if err := os.WriteFile(path, []byte(`{"locales":["fr","en","und","en",""]}`), 0o666); err != nil {
		t.Fatalf("write locale profile: %v", err)
	}
	got := readRoundTripTestProfile(t, path)
	want := []string{"en", "fr"}
	assertStringSliceEqual(t, "readRoundTripTestProfile()", got, want)
}

func unitRuntimeInput(units extract.Units) RuntimeInput {
	return unitRuntimeInputWithLocales(units, []string{"und", "en"})
}

func unitRuntimeInputWithLocales(units extract.Units, tags []string) RuntimeInput {
	input := minimalRuntimeInput()
	input.Locales.Tags = tags
	input.Units = units
	return input
}

func unitTestData(locale string, names ...string) extract.Units {
	units := cldr.Units{}
	for _, name := range names {
		units[name] = cldr.UnitData{
			Patterns: map[string]map[string]map[string]string{
				"long": {
					name: {"other": "{0} " + name},
				},
			},
			Compound: map[string]string{"long": "{0} per {1}"},
		}
	}
	return extract.Units{locale: units}
}

func unitTestNames(n int) []string {
	names := make([]string, n)
	for i := range names {
		names[i] = "unit" + strconv.Itoa(i)
	}
	return names
}

func unitTestLocaleTags(n int) []string {
	tags := make([]string, n)
	tags[0] = "und"
	for i := 1; i < n; i++ {
		tags[i] = "x-" + strconv.Itoa(i)
	}
	return tags
}
