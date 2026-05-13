package locale

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestParseSimpleLocale(t *testing.T) {
	t.Parallel()

	loc, err := Parse("en-US")
	if err != nil {
		t.Fatalf("Parse err = %v", err)
	}
	if got := loc.Tag().String(); got != "en-US" {
		t.Fatalf("Tag.String() = %q, want en-US", got)
	}
	if got := loc.BaseName(); got != "en-US" {
		t.Fatalf("BaseName() = %q, want en-US", got)
	}
	if got := loc.Language(); got != "en" {
		t.Fatalf("Language() = %q, want en", got)
	}
	if got := loc.Region(); got != "US" {
		t.Fatalf("Region() = %q, want US", got)
	}
	if got := loc.String(); got != "en-US" {
		t.Fatalf("String() = %q, want en-US", got)
	}
	if loc.Calendar() != "" || loc.Collation() != "" || loc.HourCycle() != "" || loc.CaseFirst() != "" || loc.NumberingSystem() != "" || loc.FirstDayOfWeek() != "" {
		t.Fatalf("extensions = %#v, want empty", loc)
	}
	if loc.Numeric() {
		t.Fatal("Numeric = true, want false")
	}
}

func TestParseCanonicalLanguageAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "twi", want: "ak"},
		{in: "und-Armn-SU", want: "und-Armn-AM"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := MustParse(tc.in).String(); got != tc.want {
				t.Fatalf("Parse(%q).String() = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseVariantGetters(t *testing.T) {
	t.Parallel()

	loc := MustParse("en-Latn-US-emodeng")
	if got := loc.Script(); got != "Latn" {
		t.Fatalf("Script() = %q, want Latn", got)
	}
	if got := loc.Variants(); len(got) != 1 || got[0] != "emodeng" {
		t.Fatalf("Variants() = %#v, want emodeng", got)
	}
}

func TestParseInvalidLocale(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "xx-INVALID", "x-private"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(in)
			if !errors.Is(err, ErrInvalidLocale) {
				t.Fatalf("Parse(%q) err = %v, want errors.Is(ErrInvalidLocale)", in, err)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Parallel()

	if got := MustParse("en").String(); got != "en" {
		t.Fatalf("MustParse(en).String() = %q, want en", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustParse(bad) did not panic")
		}
	}()
	_ = MustParse("")
}

func TestErrorReexports(t *testing.T) {
	t.Parallel()

	if !errors.Is(ErrInvalidOption, ecma402.ErrInvalidOption) {
		t.Fatal("ErrInvalidOption does not match ecma402.ErrInvalidOption")
	}
}
