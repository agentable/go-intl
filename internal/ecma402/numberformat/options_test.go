package ecma402nf

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestSetNumberFormatDigitOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    DigitOptionConfig
		mnfd     int
		mxfd     int
		notation string
		want     ResolvedDigitOptions
	}{
		{
			name:     "standard defaults",
			input:    defaultDigitConfig(),
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
			input: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
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
			input: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
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
			input: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
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
			input:    defaultDigitConfig(),
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
			input: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
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
			input: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
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
		name         string
		in           DigitOptionConfig
		wantName     string
		wantExpected string
	}{
		{
			name: "fraction maximum below minimum",
			in: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
				in.MinimumFractionDigits = 4
				in.MaximumFractionDigits = 2
				in.HasMinimumFractionDigits = true
				in.HasMaximumFractionDigits = true
			}),
			wantName:     "maximumFractionDigits",
			wantExpected: "greater than or equal to minimumFractionDigits",
		},
		{
			name: "significant maximum below minimum",
			in: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
				in.MinimumSignificantDigits = 4
				in.MaximumSignificantDigits = 2
				in.HasMinimumSignificantDigits = true
				in.HasMaximumSignificantDigits = true
			}),
			wantName:     "maximumSignificantDigits",
			wantExpected: "greater than or equal to minimumSignificantDigits",
		},
		{
			name: "minimum integer digits below range",
			in: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
				in.MinimumIntegerDigits = 0
			}),
			wantName:     "minimumIntegerDigits",
			wantExpected: "an integer from 1 through 21",
		},
		{
			name: "invalid rounding increment",
			in: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
				in.RoundingIncrement = 3
			}),
			wantName:     "roundingIncrement",
			wantExpected: "one of 1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000, 2000, 2500, 5000",
		},
		{
			name: "rounding increment cannot use precision priority",
			in: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
				in.RoundingIncrement = 5
				in.RoundingPriority = "morePrecision"
			}),
			wantName:     "roundingIncrement",
			wantExpected: "roundingIncrement 1 unless fraction digit rounding uses equal minimumFractionDigits and maximumFractionDigits",
		},
		{
			name: "invalid rounding mode",
			in: withDigitConfig(defaultDigitConfig(), func(in *DigitOptionConfig) {
				in.RoundingMode = "bankers"
			}),
			wantName:     "roundingMode",
			wantExpected: `one of "ceil", "floor", "expand", "trunc", "halfCeil", "halfFloor", "halfExpand", "halfTrunc", "halfEven"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, invalid, ok := SetNumberFormatDigitOptions(tc.in, 0, 3, "standard")
			if !ok {
				t.Fatal("SetNumberFormatDigitOptions() accepted invalid input")
			}
			if invalid.Name != tc.wantName {
				t.Fatalf("invalid option = %q, want %q", invalid.Name, tc.wantName)
			}
			if invalid.Expected != tc.wantExpected {
				t.Fatalf("invalid expected = %q, want %q", invalid.Expected, tc.wantExpected)
			}
		})
	}
}

func TestInvalidDigitOptionErrorCarriesExpectedGuidance(t *testing.T) {
	t.Parallel()

	err := InvalidDigitOptionError("numberformat", InvalidDigitOption{
		Name:     "roundingIncrement",
		Value:    "3",
		Expected: "one of 1, 2, 5",
	}, "en-US")
	if !errors.Is(err, ecma402.ErrInvalidOption) {
		t.Fatalf("InvalidDigitOptionError() error = %v, want ErrInvalidOption", err)
	}
	detail, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("InvalidDigitOptionError() error = %T, want OptionError", err)
	}
	if detail.Owner != "numberformat" || detail.Name != "roundingIncrement" || detail.Value != "3" || detail.Locale != "en-US" {
		t.Fatalf("OptionError = %+v, want numberformat roundingIncrement 3 en-US", detail)
	}
	if detail.Expected != "one of 1, 2, 5" {
		t.Fatalf("OptionError.Expected = %q, want rounding increment guidance", detail.Expected)
	}
}

func TestSharedStringOptions(t *testing.T) {
	t.Parallel()

	if check, ok := ecma402.InvalidStringOption(NotationOption("compact"), CompactDisplayOption("long")); ok {
		t.Fatalf("InvalidStringOption(shared number options) = %+v, true; want valid", check)
	}

	notation, ok := ecma402.InvalidStringOption(NotationOption("spellout"))
	if !ok {
		t.Fatal("InvalidStringOption(NotationOption(spellout)) ok = false, want true")
	}
	if notation.Name != "notation" || notation.Value != "spellout" || notation.Expected() != `one of "standard", "scientific", "engineering", "compact"` {
		t.Fatalf("NotationOption invalid check = %+v, want notation spellout with shared expected values", notation)
	}

	compactDisplay, ok := ecma402.InvalidStringOption(CompactDisplayOption("medium"))
	if !ok {
		t.Fatal("InvalidStringOption(CompactDisplayOption(medium)) ok = false, want true")
	}
	if compactDisplay.Name != "compactDisplay" || compactDisplay.Value != "medium" || compactDisplay.Expected() != `one of "short", "long"` {
		t.Fatalf("CompactDisplayOption invalid check = %+v, want compactDisplay medium with shared expected values", compactDisplay)
	}
}

func TestDigitOptionConfigApplyOverrides(t *testing.T) {
	t.Parallel()

	minInt, minFrac, maxFrac := 2, 3, 4
	minSig, maxSig, increment := 5, 6, 25
	cfg := DefaultDigitOptionConfig()
	cfg.ApplyOverrides(DigitOptionOverrides{
		MinimumIntegerDigits:     &minInt,
		MinimumFractionDigits:    &minFrac,
		MaximumFractionDigits:    &maxFrac,
		MinimumSignificantDigits: &minSig,
		MaximumSignificantDigits: &maxSig,
		RoundingIncrement:        &increment,
		RoundingMode:             stringPtr("ceil"),
		RoundingPriority:         stringPtr("lessPrecision"),
		TrailingZeroDisplay:      stringPtr("stripIfInteger"),
	})

	if cfg.MinimumIntegerDigits != 2 ||
		cfg.MinimumFractionDigits != 3 ||
		cfg.MaximumFractionDigits != 4 ||
		cfg.MinimumSignificantDigits != 5 ||
		cfg.MaximumSignificantDigits != 6 ||
		cfg.RoundingIncrement != 25 ||
		cfg.RoundingMode != "ceil" ||
		cfg.RoundingPriority != "lessPrecision" ||
		cfg.TrailingZeroDisplay != "stripIfInteger" {
		t.Fatalf("DigitOptionConfig after ApplyOverrides = %+v, want caller values", cfg)
	}
	if !cfg.HasMinimumFractionDigits ||
		!cfg.HasMaximumFractionDigits ||
		!cfg.HasMinimumSignificantDigits ||
		!cfg.HasMaximumSignificantDigits {
		t.Fatalf("DigitOptionConfig after ApplyOverrides = %+v, want presence flags", cfg)
	}
}

func TestDigitOptionConfigApplyOverridesKeepsExplicitEmptyStrings(t *testing.T) {
	t.Parallel()

	cfg := DefaultDigitOptionConfig()
	cfg.ApplyOverrides(DigitOptionOverrides{
		RoundingMode:        stringPtr(""),
		RoundingPriority:    stringPtr(""),
		TrailingZeroDisplay: stringPtr(""),
	})

	if cfg.RoundingMode != "" ||
		cfg.RoundingPriority != "" ||
		cfg.TrailingZeroDisplay != "" {
		t.Fatalf("DigitOptionConfig after explicit empty string overrides = %+v, want empty values", cfg)
	}
}

func TestDigitOptionConfigApplyOverridesKeepsOmittedDefaults(t *testing.T) {
	t.Parallel()

	cfg := DefaultDigitOptionConfig()
	cfg.ApplyOverrides(DigitOptionOverrides{})

	want := DefaultDigitOptionConfig()
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("DigitOptionConfig after empty ApplyOverrides = %+v, want %+v", cfg, want)
	}
}

func stringPtr[T ~string](value T) *string {
	out := string(value)
	return &out
}

func TestResolvedDigitOptionsResolvedProperties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  RoundingType
		want ResolvedDigitProperties
	}{
		{
			name: "fraction digits",
			typ:  RoundingTypeFractionDigits,
			want: ResolvedDigitProperties{
				MinimumFractionDigits: ecma402.ResolvedScalar(2),
				MaximumFractionDigits: ecma402.ResolvedScalar(4),
			},
		},
		{
			name: "significant digits",
			typ:  RoundingTypeSignificantDigits,
			want: ResolvedDigitProperties{
				MinimumSignificantDigits: ecma402.ResolvedScalar(3),
				MaximumSignificantDigits: ecma402.ResolvedScalar(5),
			},
		},
		{
			name: "precision priority",
			typ:  RoundingTypeMorePrecision,
			want: ResolvedDigitProperties{
				MinimumFractionDigits:    ecma402.ResolvedScalar(2),
				MaximumFractionDigits:    ecma402.ResolvedScalar(4),
				MinimumSignificantDigits: ecma402.ResolvedScalar(3),
				MaximumSignificantDigits: ecma402.ResolvedScalar(5),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			digits := resolvedDigits(DigitOptions{
				MinimumFractionDigits:    2,
				MaximumFractionDigits:    4,
				MinimumSignificantDigits: 3,
				MaximumSignificantDigits: 5,
			}, tc.typ)
			if got := digits.ResolvedProperties(); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ResolvedDigitOptions.ResolvedProperties() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestResolvedDigitOptionsCanUseIntegerOperands(t *testing.T) {
	t.Parallel()

	base := resolvedDigits(DigitOptions{
		MinimumIntegerDigits:  1,
		MinimumFractionDigits: 0,
		RoundingIncrement:     1,
		RoundingPriority:      "auto",
	}, RoundingTypeFractionDigits)
	tests := []struct {
		name     string
		digits   ResolvedDigitOptions
		notation string
		want     bool
	}{
		{name: "standard fraction digits with zero fraction minimum", digits: base, notation: "standard", want: true},
		{name: "compact notation", digits: base, notation: "compact"},
		{name: "minimum fraction digits visible", digits: resolvedDigits(DigitOptions{MinimumFractionDigits: 1, RoundingIncrement: 1, RoundingPriority: "auto"}, RoundingTypeFractionDigits), notation: "standard"},
		{name: "rounding increment can alter integers", digits: resolvedDigits(DigitOptions{MinimumFractionDigits: 0, RoundingIncrement: 5, RoundingPriority: "auto"}, RoundingTypeFractionDigits), notation: "standard"},
		{name: "significant digit rounding", digits: resolvedDigits(DigitOptions{MinimumFractionDigits: 0, RoundingIncrement: 1, RoundingPriority: "auto"}, RoundingTypeSignificantDigits), notation: "standard"},
		{name: "precision priority", digits: resolvedDigits(DigitOptions{MinimumFractionDigits: 0, RoundingIncrement: 1, RoundingPriority: "morePrecision"}, RoundingTypeMorePrecision), notation: "standard"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.digits.CanUseIntegerOperands(tc.notation); got != tc.want {
				t.Fatalf("ResolvedDigitOptions.CanUseIntegerOperands(%q) = %v, want %v", tc.notation, got, tc.want)
			}
		})
	}
}

func defaultDigitConfig() DigitOptionConfig {
	return DigitOptionConfig{
		MinimumIntegerDigits: 1,
		RoundingIncrement:    1,
		RoundingMode:         "halfExpand",
		RoundingPriority:     "auto",
		TrailingZeroDisplay:  "auto",
	}
}

func withDigitConfig(in DigitOptionConfig, apply func(*DigitOptionConfig)) DigitOptionConfig {
	apply(&in)
	return in
}

func resolvedDigits(options DigitOptions, roundingType RoundingType) ResolvedDigitOptions {
	return ResolvedDigitOptions{DigitOptions: options, RoundingType: roundingType}
}
