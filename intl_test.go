package gointl

import (
	"errors"
	"slices"
	"testing"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

func TestGetCanonicalLocales(t *testing.T) {
	t.Parallel()

	enUS := locale.MustParse("en-us")
	enUSAgain := locale.MustParse("en-US")
	zh := locale.MustParse("zh-hans-cn-u-nu-latn")

	got := GetCanonicalLocales(enUS, enUSAgain, zh)
	want := []locale.Locale{enUS, zh}
	if len(got) != len(want) {
		t.Fatalf("GetCanonicalLocales() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].String() != want[i].String() {
			t.Fatalf("GetCanonicalLocales()[%d] = %q, want %q", i, got[i].String(), want[i].String())
		}
	}
}

func TestSupportedValuesOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  SupportedValueKey
		want string
	}{
		{SupportedValueCalendar, "iso8601"},
		{SupportedValueCurrency, "USD"},
		{SupportedValueNumberingSystem, "arab"},
		{SupportedValueTimeZone, "America/New_York"},
		{SupportedValueUnit, "meter"},
	}
	for _, tt := range tests {
		got, err := SupportedValuesOf(tt.key)
		if err != nil {
			t.Fatalf("SupportedValuesOf(%q) error = %v", tt.key, err)
		}
		if !slices.Contains(got, tt.want) {
			t.Fatalf("SupportedValuesOf(%q) missing %q in %v", tt.key, tt.want, got)
		}
		if !slices.IsSorted(got) {
			t.Fatalf("SupportedValuesOf(%q) = %v, want sorted", tt.key, got)
		}
	}
}

func TestSupportedValuesOfECMA402NumberingSystems(t *testing.T) {
	t.Parallel()

	got, err := SupportedValuesOf(SupportedValueNumberingSystem)
	if err != nil {
		t.Fatalf("SupportedValuesOf(numberingSystem) error = %v", err)
	}
	for _, want := range []string{"adlm", "arab", "arabext", "beng", "deva", "fullwide", "hanidec", "latn", "thai"} {
		if !slices.Contains(got, want) {
			t.Fatalf("SupportedValuesOf(numberingSystem) missing %q in %v", want, got)
		}
	}
}

func TestSupportedValuesOfInvalidKey(t *testing.T) {
	t.Parallel()

	_, err := SupportedValuesOf(SupportedValueKey("bad"))
	if !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("SupportedValuesOf(bad) error = %v, want ErrInvalidKey", err)
	}
}

func TestSupportedValuesOfUnitsMatchECMA402(t *testing.T) {
	t.Parallel()

	got, err := SupportedValuesOf(SupportedValueUnit)
	if err != nil {
		t.Fatalf("SupportedValuesOf(unit) error = %v", err)
	}
	want := []string{
		"acre", "bit", "byte", "celsius", "centimeter", "day", "degree",
		"fahrenheit", "fluid-ounce", "foot", "gallon", "gigabit", "gigabyte",
		"gram", "hectare", "hour", "inch", "kilobit", "kilobyte", "kilogram",
		"kilometer", "liter", "megabit", "megabyte", "meter", "microsecond",
		"mile", "mile-scandinavian", "milliliter", "millimeter", "millisecond",
		"minute", "month", "nanosecond", "ounce", "percent", "petabyte",
		"pound", "second", "stone", "terabit", "terabyte", "week", "yard",
		"year",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("SupportedValuesOf(unit) = %v, want %v", got, want)
	}
}

func TestSupportedValuesOfCollationsMatchGeneratedCLDR(t *testing.T) {
	t.Parallel()

	got, err := SupportedValuesOf(SupportedValueCollation)
	if err != nil {
		t.Fatalf("SupportedValuesOf(collation) error = %v", err)
	}
	want := []string{"compat", "dict", "emoji", "eor", "phonebk", "phonetic", "pinyin", "searchjl", "stroke", "trad", "unihan", "zhuyin"}
	if !slices.Equal(got, want) {
		t.Fatalf("SupportedValuesOf(collation) = %v, want %v", got, want)
	}
}

func TestRootErrorSentinelsAliasFormatterErrors(t *testing.T) {
	t.Parallel()

	if !errors.Is(ErrInvalidOption, numberformat.ErrInvalidOption) {
		t.Fatalf("numberformat.ErrInvalidOption does not match root ErrInvalidOption")
	}
	if !errors.Is(ErrInvalidOption, pluralrules.ErrInvalidOption) {
		t.Fatalf("pluralrules.ErrInvalidOption does not match root ErrInvalidOption")
	}
	if !errors.Is(ErrUnsupportedTimeZone, datetimeformat.ErrUnsupportedTimeZone) {
		t.Fatalf("datetimeformat.ErrUnsupportedTimeZone does not match root ErrUnsupportedTimeZone")
	}
}
