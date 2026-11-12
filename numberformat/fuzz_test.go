package numberformat

import (
	"errors"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/pluralrules"
)

// This crosses the public numeric bridges, shared digit resolution, and both
// partition sinks. Inputs are bounded without excluding Number special values.
func FuzzNumberFormatPartition(f *testing.F) {
	for _, value := range []float64{0, math.Copysign(0, -1), 1.005, -1.005, 999.995, 9007199254740992, math.NaN(), math.Inf(1), math.Inf(-1), 1e100} {
		for scenario := uint8(0); scenario < 6; scenario++ {
			f.Add(math.Float64bits(value), math.Float64bits(-value), scenario, uint8(0), scenario, uint8(2), uint8(6))
		}
	}
	for bridge := uint8(1); bridge < 5; bridge++ {
		f.Add(uint64(9007199254740993), uint64(17), uint8(0), bridge, uint8(0), uint8(0), uint8(8))
	}
	f.Fuzz(func(t *testing.T, a, b uint64, scenario, bridges, language, precision, rounding uint8) {
		tag := []string{"en", "ru", "ar", "zh", "fr"}[int(language)%5]
		locales := intltest.LocaleList(t, tag)
		digits := int(precision % 7)
		modes := []string{"ceil", "floor", "expand", "trunc", "halfCeil", "halfFloor", "halfExpand", "halfTrunc", "halfEven", "invalid"}
		mode := modes[int(rounding)%len(modes)]
		opts := Options{MaximumFractionDigits: new(digits), RoundingMode: new(mode)}
		switch scenario % 6 {
		case 1:
			opts.Style, opts.Currency = new("currency"), new("USD")
		case 2:
			opts.Style, opts.Unit, opts.UnitDisplay = new("unit"), new("meter-per-second"), new("long")
		case 3:
			opts.Notation = new("compact")
		case 4:
			opts.Notation = new("scientific")
		case 5:
			opts.NumberingSystem = new("arab")
		}
		if precision&8 != 0 {
			opts.MinimumFractionDigits = new(digits)
			opts.RoundingIncrement = new([]int{1, 5, 25}[int(precision>>4)%3])
		} else if precision&16 != 0 {
			opts.MaximumSignificantDigits = new(3)
			opts.RoundingPriority = new([]string{"auto", "morePrecision", "lessPrecision"}[int(precision>>5)%3])
		}
		format, err := New(locales, opts)
		if mode == "invalid" {
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("invalid mode error = %v", err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if string(format.ResolvedOptions().RoundingMode) != mode {
			t.Fatal("rounding mode did not reach resolved options")
		}
		input, nanA := fuzzNumberValue(t, a, bridges%5, precision)
		endpoint, nanB := fuzzNumberValue(t, b, (bridges/5)%5, precision)
		var joined strings.Builder
		for _, part := range format.FormatToParts(input) {
			joined.WriteString(part.Value)
		}
		if got := format.Format(input); got != joined.String() {
			t.Fatalf("text %q != parts %q", got, joined.String())
		}
		// Exact integer bridges must agree without passing through binary64.
		if bridges%5 == 1 {
			exact, err := ParseValue(strconv.FormatInt(int64(a), 10))
			if err != nil {
				t.Fatal(err)
			}
			if format.Format(input) != format.Format(BigInt(big.NewInt(int64(a)))) || format.Format(input) != format.Format(exact) {
				t.Fatal("exact integer bridges disagree")
			}
		}
		text, err := format.FormatRange(input, endpoint)
		parts, partsErr := format.FormatRangeToParts(input, endpoint)
		if nanA || nanB {
			if !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(partsErr, intlerr.ErrInvalidValue) {
				t.Fatalf("NaN range errors = %v, %v", err, partsErr)
			}
		} else {
			if err != nil || partsErr != nil {
				t.Fatalf("valid range errors = %v, %v", err, partsErr)
			}
			joined.Reset()
			for _, part := range parts {
				if part.Source != SourceStartRange && part.Source != SourceEndRange && part.Source != SourceShared {
					t.Fatalf("invalid source %q", part.Source)
				}
				joined.WriteString(part.Value)
			}
			if text != joined.String() {
				t.Fatalf("range %q != parts %q", text, joined.String())
			}
		}
		rules, err := pluralrules.New(locales, pluralrules.Options{MaximumFractionDigits: new(digits), RoundingMode: new(mode)})
		if err != nil {
			t.Fatal(err)
		}
		category := rules.Select(pluralrules.Float(math.Float64frombits(a)))
		if !slices.Contains(rules.ResolvedOptions().PluralCategories, category) {
			t.Fatalf("category %v outside resolved set", category)
		}
	})
}

func fuzzNumberValue(t *testing.T, bits uint64, bridge, scale uint8) (Value, bool) {
	t.Helper()
	switch bridge {
	case 1:
		return Int(int64(bits)), false
	case 2:
		return Uint(bits), false
	case 3:
		return BigInt(new(big.Int).Lsh(new(big.Int).SetUint64(bits), uint(scale%128))), false
	case 4:
		v, err := ParseValue(strconv.FormatInt(int64(bits), 10) + "e-" + strconv.Itoa(int(scale%10)))
		if err != nil {
			t.Fatal(err)
		}
		return v, false
	default:
		value := math.Float64frombits(bits)
		if !math.IsNaN(value) && !math.IsInf(value, 0) && math.Abs(value) > 1e100 {
			value = math.Copysign(1e100, value)
		}
		return Float(value), math.IsNaN(value)
	}
}

// The independent rational oracle catches shared rendering/rounding mistakes
// that a string-versus-parts comparison alone cannot detect.
func FuzzNumberFormatRationalRounding(f *testing.F) {
	for _, n := range []int64{0, 5, -5, 15, -15, 1005, -1005, 999995, 9007199254740993} {
		for mode := uint8(0); mode < 3; mode++ {
			f.Add(n, uint8(3), uint8(2), uint8(1), mode)
		}
	}
	f.Fuzz(func(t *testing.T, n int64, scale, digits, step, mode uint8) {
		power, fraction := int(scale%10), int(digits%4)
		increment := []int64{1, 5, 25}[step%3]
		rounding := []string{"halfExpand", "halfEven", "trunc"}[mode%3]
		denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(power)), nil)
		rat := new(big.Rat).SetFrac(big.NewInt(n), denominator)
		input, err := ParseValue(strconv.FormatInt(n, 10) + "e-" + strconv.Itoa(power))
		if err != nil {
			t.Fatal(err)
		}
		format, err := New(intltest.LocaleList(t, "en"), Options{
			UseGrouping: new("false"), MinimumFractionDigits: new(fraction), MaximumFractionDigits: new(fraction),
			RoundingIncrement: new(int(increment)), RoundingMode: new(rounding),
		})
		if err != nil {
			t.Fatal(err)
		}
		unit := new(big.Rat).SetFrac(big.NewInt(increment), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(fraction)), nil))
		units := new(big.Rat).Quo(new(big.Rat).Abs(rat), unit)
		quotient, remainder := new(big.Int), new(big.Int)
		quotient.QuoRem(units.Num(), units.Denom(), remainder)
		cmp := new(big.Int).Lsh(remainder, 1).Cmp(units.Denom())
		if rounding != "trunc" && (cmp > 0 || cmp == 0 && (rounding == "halfExpand" || quotient.Bit(0) == 1)) {
			quotient.Add(quotient, big.NewInt(1))
		}
		want := new(big.Rat).Mul(new(big.Rat).SetInt(quotient), unit).FloatString(fraction)
		if n < 0 {
			want = "-" + want
		}
		if got := format.Format(input); got != want {
			t.Fatalf("Format(%s, %s, increment %d, digits %d)=%q, want %q", rat, rounding, increment, fraction, got, want)
		}
	})
}
