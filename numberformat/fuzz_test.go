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
	for _, value := range []float64{0, math.Copysign(0, -1), 1.005, 9007199254740992, math.NaN(), math.Inf(1), math.Inf(-1), 1e100} {
		for scenario := uint8(0); scenario < 6; scenario++ {
			f.Add(math.Float64bits(value), scenario)
		}
	}
	f.Fuzz(func(t *testing.T, bits uint64, scenario uint8) {
		value := math.Float64frombits(bits)
		if !math.IsNaN(value) && !math.IsInf(value, 0) && math.Abs(value) > 1e100 {
			value = math.Copysign(1e100, value)
		}
		tag := []string{"en", "ru", "ar", "zh", "fr"}[int(scenario)%5]
		locales := intltest.LocaleList(t, tag)
		digits := int(scenario % 7)
		opts := Options{MaximumFractionDigits: new(digits)}
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
		format, err := New(locales, opts)
		if err != nil {
			t.Fatal(err)
		}
		input := Float(value)
		// Integer bridges keep their exact integer instead of passing through float64.
		if scenario&8 != 0 {
			input = Uint(bits)
		}
		var joined strings.Builder
		for _, part := range format.FormatToParts(input) {
			joined.WriteString(part.Value)
		}
		if got := format.Format(input); got != joined.String() {
			t.Fatalf("text %q != parts %q", got, joined.String())
		}
		text, err := format.FormatRange(input, Int(0))
		parts, partsErr := format.FormatRangeToParts(input, Int(0))
		if scenario&8 == 0 && math.IsNaN(value) {
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
		rules, err := pluralrules.New(locales, pluralrules.Options{MaximumFractionDigits: new(digits)})
		if err != nil {
			t.Fatal(err)
		}
		category := rules.Select(pluralrules.Float(value))
		if !slices.Contains(rules.ResolvedOptions().PluralCategories, category) {
			t.Fatalf("category %v outside resolved set", category)
		}
	})
}

// The independent rational oracle catches shared rendering/rounding mistakes
// that a string-versus-parts comparison alone cannot detect.
func FuzzNumberFormatRationalRounding(f *testing.F) {
	for _, n := range []int64{0, 5, -5, 15, -15, 1005, -1005, 9007199254740993} {
		f.Add(n, uint8(2))
	}
	f.Fuzz(func(t *testing.T, n int64, scale uint8) {
		power := int(scale % 10)
		denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(power)), nil)
		rat := new(big.Rat).SetFrac(big.NewInt(n), denominator)
		input, err := ParseValue(strconv.FormatInt(n, 10) + "e-" + strconv.Itoa(power))
		if err != nil {
			t.Fatal(err)
		}
		format, err := New(intltest.LocaleList(t, "en"), Options{UseGrouping: new("false"), MaximumFractionDigits: new(0)})
		if err != nil {
			t.Fatal(err)
		}
		magnitude := new(big.Int).Abs(rat.Num())
		quotient, remainder := new(big.Int), new(big.Int)
		quotient.QuoRem(magnitude, rat.Denom(), remainder)
		if new(big.Int).Lsh(remainder, 1).Cmp(rat.Denom()) >= 0 {
			quotient.Add(quotient, big.NewInt(1))
		}
		want := quotient.String()
		if n < 0 {
			want = "-" + want
		}
		if got := format.Format(input); got != want {
			t.Fatalf("Format(%s)=%q, want %q", rat, got, want)
		}
	})
}
