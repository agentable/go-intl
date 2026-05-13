package ecma402nf

import (
	"testing"

	"github.com/agentable/go-intl/internal/decimal"
)

func TestFormatDecimal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts DigitOptions
		in   string
		want string
	}{
		{
			name: "fixed fraction",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 2, MaximumFractionDigits: 2, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "1.234",
			want: "1.23",
		},
		{
			name: "minimum integer",
			opts: DigitOptions{MinimumIntegerDigits: 3, MinimumFractionDigits: 0, MaximumFractionDigits: 0, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "5",
			want: "005",
		},
		{
			name: "zero keeps fixed fraction scale",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 2, MaximumFractionDigits: 2, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "0",
			want: "0.00",
		},
		{
			name: "significant digits",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 3, MinimumSignificantDigits: 2, MaximumSignificantDigits: 4, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "0.0012345",
			want: "0.001235",
		},
		{
			name: "rounding priority",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 2, MaximumFractionDigits: 2, MinimumSignificantDigits: 2, MaximumSignificantDigits: 4, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "lessPrecision", TrailingZeroDisplay: "auto"},
			in:   "12345",
			want: "12350",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d, err := decimal.ParseString(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if got := FormatDecimal(d, tc.opts); got != tc.want {
				t.Fatalf("FormatDecimal(%s) = %q, want %q", tc.in, got, tc.want)
			}
			if got := FormatNumericToString(d, tc.opts).Formatted; got != tc.want {
				t.Fatalf("FormatNumericToString(%s).Formatted = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatNumericToStringReturnsRoundedValue(t *testing.T) {
	t.Parallel()

	d, err := decimal.ParseString("1.5")
	if err != nil {
		t.Fatal(err)
	}
	got := FormatNumericToString(d, DigitOptions{
		MinimumIntegerDigits:  1,
		MaximumFractionDigits: 0,
		RoundingIncrement:     1,
		RoundingMode:          "halfExpand",
		RoundingPriority:      "auto",
		TrailingZeroDisplay:   "auto",
	})
	if got.Formatted != "2" {
		t.Fatalf("Formatted = %q, want 2", got.Formatted)
	}
	if got.Rounded.String() != "2" {
		t.Fatalf("Rounded = %s, want 2", got.Rounded.String())
	}
}
