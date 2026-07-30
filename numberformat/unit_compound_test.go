package numberformat

import (
	"reflect"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestCompoundUnitFormattingMatchesNative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, locale, unit string
		display            UnitDisplay
		value              int64
		want               string
		parts              []Part
	}{
		{name: "en long", locale: "en", unit: "meter-per-second", display: LongUnitDisplay, value: 3, want: "3 meters per second", parts: []Part{{Type: PartInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "meters per second"}}},
		{name: "en short", locale: "en", unit: "meter-per-second", display: ShortUnitDisplay, value: 3, want: "3 m/s", parts: []Part{{Type: PartInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "m/s"}}},
		{name: "en narrow", locale: "en", unit: "meter-per-second", display: NarrowUnitDisplay, value: 3, want: "3m/s", parts: []Part{{Type: PartInteger, Value: "3"}, {Type: PartUnit, Value: "m/s"}}},
		{name: "ja direct short", locale: "ja", unit: "kilometer-per-hour", display: ShortUnitDisplay, value: 3, want: "3 km/h", parts: []Part{{Type: PartInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "km/h"}}},
		{name: "ja direct long", locale: "ja", unit: "kilometer-per-hour", display: LongUnitDisplay, value: 3, want: "時速 3 キロメートル", parts: []Part{{Type: PartUnit, Value: "時速"}, {Type: PartLiteral, Value: " "}, {Type: PartInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "キロメートル"}}},
		{name: "fr direct pinned CLDR", locale: "fr", unit: "liter-per-kilometer", display: LongUnitDisplay, value: 3, want: "3 litres au kilomètre", parts: []Part{{Type: PartInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "litres au kilomètre"}}},
		{name: "en generic", locale: "en", unit: "meter-per-megabyte", display: LongUnitDisplay, value: 3, want: "3 meters per megabyte", parts: []Part{{Type: PartInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "meters per megabyte"}}},
		{name: "ru one", locale: "ru", unit: "meter-per-second", display: LongUnitDisplay, value: 1, want: "1 метр в секунду", parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "метр в секунду"}}},
		{name: "ru few", locale: "ru", unit: "meter-per-second", display: LongUnitDisplay, value: 2, want: "2 метра в секунду", parts: []Part{{Type: PartInteger, Value: "2"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "метра в секунду"}}},
		{name: "ru many", locale: "ru", unit: "meter-per-second", display: LongUnitDisplay, value: 5, want: "5 метров в секунду", parts: []Part{{Type: PartInteger, Value: "5"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "метров в секунду"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := New(locale.List{intltest.Locale(t, tc.locale)}, Options{
				Style:       stringPtr(UnitStyle),
				Unit:        stringPtr(tc.unit),
				UnitDisplay: stringPtr(tc.display),
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			value := Int(tc.value)
			gotParts := f.FormatToParts(value)
			if !reflect.DeepEqual(gotParts, tc.parts) {
				t.Errorf("FormatToParts() = %#v, want %#v", gotParts, tc.parts)
			}
			if got := f.Format(value); got != tc.want {
				t.Errorf("Format() = %q, want %q", got, tc.want)
			}
			var joined strings.Builder
			for _, part := range gotParts {
				joined.WriteString(part.Value)
			}
			if got := f.Format(value); got != joined.String() {
				t.Errorf("Format() = %q, parts join = %q", got, joined.String())
			}
			if strings.Contains(f.Format(value), "-per-") {
				t.Errorf("Format() leaked raw compound identifier: %q", f.Format(value))
			}
		})
	}
}

func TestUnitPatternsMayOmitNumericPartition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, unit, want string
		value            int64
		parts            []Part
	}{
		{name: "simple one", unit: "degree", value: 1, want: "درجة", parts: []Part{{Type: PartUnit, Value: "درجة"}}},
		{name: "simple two", unit: "degree", value: 2, want: "درجتان", parts: []Part{{Type: PartUnit, Value: "درجتان"}}},
		{name: "specialized compound one", unit: "degree-per-second", value: 1, want: "درجة في الثانية", parts: []Part{{Type: PartUnit, Value: "درجة في الثانية"}}},
		{name: "generic compound two", unit: "degree-per-megabyte", value: 2, want: "درجتان لكل ميغابايت", parts: []Part{{Type: PartUnit, Value: "درجتان لكل ميغابايت"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := New(locale.List{intltest.Locale(t, "ar")}, Options{
				Style:       stringPtr(UnitStyle),
				Unit:        stringPtr(tc.unit),
				UnitDisplay: stringPtr(LongUnitDisplay),
			})
			if err != nil {
				t.Fatal(err)
			}
			value := Int(tc.value)
			if got := f.Format(value); got != tc.want {
				t.Errorf("Format() = %q, want %q", got, tc.want)
			}
			if got := f.FormatToParts(value); !reflect.DeepEqual(got, tc.parts) {
				t.Errorf("FormatToParts() = %#v, want %#v", got, tc.parts)
			}
		})
	}
}

func TestSimpleUnitKeepsDisplayPhraseInOnePart(t *testing.T) {
	t.Parallel()

	f, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:       stringPtr(UnitStyle),
		Unit:        stringPtr("fluid-ounce"),
		UnitDisplay: stringPtr(LongUnitDisplay),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []Part{
		{Type: PartInteger, Value: "3"},
		{Type: PartLiteral, Value: " "},
		{Type: PartUnit, Value: "fluid ounces"},
	}
	if got := f.FormatToParts(Int(3)); !reflect.DeepEqual(got, want) {
		t.Errorf("FormatToParts() = %#v, want %#v", got, want)
	}
}
