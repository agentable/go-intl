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
