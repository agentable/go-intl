package ecma402nf

import (
	"reflect"
	"testing"
)

func TestSetNumberFormatDigitOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    DigitOptionInput
		mnfd     int
		mxfd     int
		notation string
		want     ResolvedDigitOptions
	}{
		{
			name:     "standard defaults",
			input:    defaultDigitInput(),
			mnfd:     0,
			mxfd:     3,
			notation: "standard",
			want: resolvedDigits(
				DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
				RoundingTypeFractionDigits,
			),
		},
		{
			name: "minimum fraction raises maximum default",
			input: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.MinimumFractionDigits = 4
				in.HasMinimumFractionDigits = true
			}),
			mnfd:     0,
			mxfd:     3,
			notation: "standard",
			want: resolvedDigits(
				DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 4, MaximumFractionDigits: 4, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
				RoundingTypeFractionDigits,
			),
		},
		{
			name: "maximum fraction lowers minimum default",
			input: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.MaximumFractionDigits = 1
				in.HasMaximumFractionDigits = true
			}),
			mnfd:     2,
			mxfd:     3,
			notation: "standard",
			want: resolvedDigits(
				DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 1, MaximumFractionDigits: 1, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
				RoundingTypeFractionDigits,
			),
		},
		{
			name: "minimum significant defaults maximum",
			input: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.MinimumSignificantDigits = 2
				in.HasMinimumSignificantDigits = true
			}),
			mnfd:     0,
			mxfd:     3,
			notation: "standard",
			want: resolvedDigits(
				DigitOptions{MinimumIntegerDigits: 1, MinimumSignificantDigits: 2, MaximumSignificantDigits: 21, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
				RoundingTypeSignificantDigits,
			),
		},
		{
			name:     "compact defaults use precision priority",
			input:    defaultDigitInput(),
			mnfd:     0,
			mxfd:     3,
			notation: "compact",
			want: resolvedDigits(
				DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 0, MinimumSignificantDigits: 1, MaximumSignificantDigits: 2, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "morePrecision", TrailingZeroDisplay: "auto"},
				RoundingTypeMorePrecision,
			),
		},
		{
			name: "more precision resolves both digit families",
			input: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.MaximumFractionDigits = 2
				in.HasMaximumFractionDigits = true
				in.RoundingPriority = "morePrecision"
			}),
			mnfd:     0,
			mxfd:     3,
			notation: "standard",
			want: resolvedDigits(
				DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 2, MinimumSignificantDigits: 1, MaximumSignificantDigits: 21, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "morePrecision", TrailingZeroDisplay: "auto"},
				RoundingTypeMorePrecision,
			),
		},
		{
			name: "rounding increment equalizes fraction defaults",
			input: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.RoundingIncrement = 5
			}),
			mnfd:     0,
			mxfd:     3,
			notation: "standard",
			want: resolvedDigits(
				DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 0, RoundingIncrement: 5, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
				RoundingTypeFractionDigits,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, invalid, ok := SetNumberFormatDigitOptions(tc.input, tc.mnfd, tc.mxfd, tc.notation)
			if ok {
				t.Fatalf("SetNumberFormatDigitOptions() invalid = %+v, want valid", invalid)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("SetNumberFormatDigitOptions() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestSetNumberFormatDigitOptionsRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   DigitOptionInput
		want string
	}{
		{
			name: "fraction maximum below minimum",
			in: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.MinimumFractionDigits = 4
				in.MaximumFractionDigits = 2
				in.HasMinimumFractionDigits = true
				in.HasMaximumFractionDigits = true
			}),
			want: "maximumFractionDigits",
		},
		{
			name: "significant maximum below minimum",
			in: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.MinimumSignificantDigits = 4
				in.MaximumSignificantDigits = 2
				in.HasMinimumSignificantDigits = true
				in.HasMaximumSignificantDigits = true
			}),
			want: "maximumSignificantDigits",
		},
		{
			name: "invalid rounding increment",
			in: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.RoundingIncrement = 3
			}),
			want: "roundingIncrement",
		},
		{
			name: "rounding increment cannot use precision priority",
			in: withDigitInput(defaultDigitInput(), func(in *DigitOptionInput) {
				in.RoundingIncrement = 5
				in.RoundingPriority = "morePrecision"
			}),
			want: "roundingIncrement",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, invalid, ok := SetNumberFormatDigitOptions(tc.in, 0, 3, "standard")
			if !ok {
				t.Fatal("SetNumberFormatDigitOptions() accepted invalid input")
			}
			if invalid.Name != tc.want {
				t.Fatalf("invalid option = %q, want %q", invalid.Name, tc.want)
			}
		})
	}
}

func defaultDigitInput() DigitOptionInput {
	return DigitOptionInput{
		MinimumIntegerDigits: 1,
		RoundingIncrement:    1,
		RoundingMode:         "halfExpand",
		RoundingPriority:     "auto",
		TrailingZeroDisplay:  "auto",
	}
}

func withDigitInput(in DigitOptionInput, apply func(*DigitOptionInput)) DigitOptionInput {
	apply(&in)
	return in
}

func resolvedDigits(options DigitOptions, roundingType RoundingType) ResolvedDigitOptions {
	return ResolvedDigitOptions{DigitOptions: options, RoundingType: roundingType}
}
