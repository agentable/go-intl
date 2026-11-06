package numberformat

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func intPtr(v int) *int {
	return &v
}

func stringPtr[T ~string](v T) *string {
	value := string(v)
	return &value
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	digits := 2
	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		MinimumFractionDigits: &digits,
		MaximumFractionDigits: &digits,
	})
	if err != nil {
		t.Fatal(err)
	}
	digits = 0

	if got := format.Format(Int(1)); got != "1.00" {
		t.Fatalf("Format(1) = %q, want 1.00", got)
	}
}

func TestNumberFormatOptionErrorsUseCanonicalLocaleName(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en-US-u-nu-latn")}, Options{Style: stringPtr(CurrencyStyle)})
	if err == nil {
		t.Fatal("New() error = nil, want invalid option")
	}
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
	testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "currency", "", "en-US-u-nu-latn")
	testcontract.AssertOptionExpected(t, err, `a currency code when style is "currency"`)
}

func TestNumberFormatStringOptionErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		opts      Options
		wantName  string
		wantValue string
	}{
		{
			name:      "style",
			opts:      Options{Style: stringPtr(Style("bad"))},
			wantName:  "style",
			wantValue: "bad",
		},
		{
			name:      "currency display",
			opts:      Options{CurrencyDisplay: stringPtr(CurrencyDisplay("bad"))},
			wantName:  "currencyDisplay",
			wantValue: "bad",
		},
		{
			name:      "currency sign",
			opts:      Options{CurrencySign: stringPtr(CurrencySign("bad"))},
			wantName:  "currencySign",
			wantValue: "bad",
		},
		{
			name:      "unit display",
			opts:      Options{UnitDisplay: stringPtr(UnitDisplay("bad"))},
			wantName:  "unitDisplay",
			wantValue: "bad",
		},
		{
			name:      "sign display",
			opts:      Options{SignDisplay: stringPtr(SignDisplay("bad"))},
			wantName:  "signDisplay",
			wantValue: "bad",
		},
		{
			name:      "use grouping",
			opts:      Options{UseGrouping: stringPtr(UseGrouping("sometimes"))},
			wantName:  "useGrouping",
			wantValue: "sometimes",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, tc.wantName, tc.wantValue, "en")
		})
	}
}

func TestInactiveFractionOptions(t *testing.T) {
	t.Parallel()
	locales := intltest.LocaleList(t, "en")
	control, err := New(locales, Options{MaximumSignificantDigits: new(2)})
	if err != nil {
		t.Fatal(err)
	}
	for _, digits := range [][2]int{{-1, 2}, {0, 101}, {5, 2}} {
		for _, priority := range []string{"auto", "morePrecision", "lessPrecision"} {
			f, err := New(locales, Options{MaximumSignificantDigits: new(2), MinimumFractionDigits: new(digits[0]), MaximumFractionDigits: new(digits[1]), RoundingPriority: new(priority)})
			if priority != "auto" {
				if !errors.Is(err, intlerr.ErrInvalidOption) {
					t.Fatalf("%s %v: error = %v", priority, digits, err)
				}
				continue
			}
			if err != nil {
				t.Fatalf("inactive %v: %v", digits, err)
			}
			resolved := f.ResolvedOptions()
			if resolved.MinimumFractionDigits != nil || resolved.MaximumFractionDigits != nil {
				t.Fatal("inactive fraction slots present")
			}
			for _, value := range []int64{1, 12, 123} {
				if got, want := f.Format(Int(value)), control.Format(Int(value)); got != want {
					t.Fatalf("%v: got %v, want %v", digits, got, want)
				}
			}
		}
	}
}

func TestNumberFormatOptionErrorOrder(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		opts  Options
		value string
	}{
		{"localeMatcher", Options{LocaleMatcher: new("bad"), Style: new("bad")}, "bad"},
		{"numberingSystem", Options{NumberingSystem: new("!"), Style: new("bad")}, "!"},
		{"style", Options{Style: new("bad"), Currency: new("!")}, "bad"},
		{"currency", Options{Style: new("currency"), Notation: new("bad")}, ""},
		{"currency", Options{Currency: new("!"), CurrencyDisplay: new("bad")}, "!"},
		{"currencyDisplay", Options{CurrencyDisplay: new("bad"), CurrencySign: new("bad")}, "bad"},
		{"currencySign", Options{CurrencySign: new("bad"), Unit: new("bad")}, "bad"},
		{"unit", Options{Style: new("unit"), UnitDisplay: new("bad")}, ""},
		{"unit", Options{Unit: new("bad"), UnitDisplay: new("bad")}, "bad"},
		{"unitDisplay", Options{UnitDisplay: new("bad"), Notation: new("bad")}, "bad"},
		{"notation", Options{Notation: new("bad"), MinimumIntegerDigits: new(0)}, "bad"},
		{"minimumIntegerDigits", Options{MinimumIntegerDigits: new(0), CompactDisplay: new("bad")}, "0"},
		{"compactDisplay", Options{CompactDisplay: new("bad"), UseGrouping: new("bad")}, "bad"},
		{"useGrouping", Options{UseGrouping: new("bad"), SignDisplay: new("bad")}, "bad"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(intltest.LocaleList(t, "en"), tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("error = %v, want ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, tc.name, tc.value, "en")
		})
	}
}
