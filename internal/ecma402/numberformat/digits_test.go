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
		{
			name: "more precision chooses closer fixed candidate",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 2, MinimumSignificantDigits: 1, MaximumSignificantDigits: 2, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "morePrecision", TrailingZeroDisplay: "auto"},
			in:   "1.2345",
			want: "1.23",
		},
		{
			name: "more precision chooses closer significant candidate",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 1, MinimumSignificantDigits: 1, MaximumSignificantDigits: 4, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "morePrecision", TrailingZeroDisplay: "auto"},
			in:   "1.2345",
			want: "1.235",
		},
		{
			name: "less precision chooses coarser fixed candidate",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 0, MinimumSignificantDigits: 1, MaximumSignificantDigits: 4, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "lessPrecision", TrailingZeroDisplay: "auto"},
			in:   "1.2345",
			want: "1",
		},
		{
			name: "minimum significant digits pads fraction",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 3, MinimumSignificantDigits: 3, MaximumSignificantDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "1",
			want: "1.00",
		},
		{
			name: "significant strip if integer",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 3, MinimumSignificantDigits: 2, MaximumSignificantDigits: 2, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "stripIfInteger"},
			in:   "1.0",
			want: "1",
		},
		{
			name: "significant strip leaves non-zero fraction",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 3, MinimumSignificantDigits: 3, MaximumSignificantDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "stripIfInteger"},
			in:   "1.20",
			want: "1.20",
		},
		{
			name: "fixed fraction trims to minimum fraction digits",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 1, MaximumFractionDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "1",
			want: "1.0",
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

func TestFormatNumericToStringReturnsNonFiniteValue(t *testing.T) {
	t.Parallel()

	for _, d := range []decimal.Decimal{decimal.NaNValue, decimal.PosInfinity, decimal.NegInfinity} {
		t.Run(d.String(), func(t *testing.T) {
			t.Parallel()

			got := FormatNumericToString(d, DigitOptions{
				MinimumIntegerDigits:  1,
				MaximumFractionDigits: 2,
				RoundingIncrement:     1,
				RoundingMode:          "halfExpand",
				RoundingPriority:      "auto",
				TrailingZeroDisplay:   "auto",
			})
			if got.Formatted != d.String() {
				t.Fatalf("Formatted = %q, want %q", got.Formatted, d.String())
			}
			if got.Rounded.String() != d.String() {
				t.Fatalf("Rounded = %s, want %s", got.Rounded.String(), d.String())
			}
		})
	}
}

func TestFormatNumericToStringPreservesNegativeRoundedValue(t *testing.T) {
	t.Parallel()

	d, err := decimal.ParseString("-1.5")
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
	if got.Formatted != "-2" {
		t.Fatalf("Formatted = %q, want -2", got.Formatted)
	}
	if got.Rounded.String() != "-2" {
		t.Fatalf("Rounded = %s, want -2", got.Rounded.String())
	}
}

func TestFormatNumericToStringFormatsNegativeZeroWithoutNegativeRoundedValue(t *testing.T) {
	t.Parallel()

	d, err := decimal.ParseString("-0.1")
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
	if got.Formatted != "-0" {
		t.Fatalf("Formatted = %q, want -0", got.Formatted)
	}
	if got.Rounded.String() != "0" {
		t.Fatalf("Rounded = %s, want 0", got.Rounded.String())
	}
}

func TestFormatDecimalSignificantDigitEdges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts DigitOptions
		in   string
		want string
	}{
		{
			name: "zero pads to minimum significant digits",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumSignificantDigits: 3, MaximumSignificantDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "0",
			want: "0.00",
		},
		{
			name: "fraction trims surplus significant trailing zeros",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumSignificantDigits: 3, MaximumSignificantDigits: 5, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "0.0012300",
			want: "0.00123",
		},
		{
			name: "negative significant rounding preserves sign",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumSignificantDigits: 1, MaximumSignificantDigits: 2, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "-0.001234",
			want: "-0.0012",
		},
		{
			name: "invalid rounding mode falls back to half expand",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 0, RoundingIncrement: 1, RoundingMode: "not-a-mode", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			in:   "1.5",
			want: "2",
		},
		{
			name: "fixed strip removes integer fraction",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 2, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "stripIfInteger"},
			in:   "1.00",
			want: "1",
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
		})
	}
}
