package pluralrules

import (
	"math"
	"math/big"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
)

func TestSelectRangeBridgeEquality(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		options      Options
		starts, ends []Value
		want         Category
	}{
		{"opposite signs", Options{}, []Value{Int(-1), Float(-1), BigInt(big.NewInt(-1)), decimalValue("-1")}, []Value{Int(1), Uint(1), Float(1), BigInt(big.NewInt(1)), decimalValue("1.00")}, One},
		{"visible zeros", Options{MinimumFractionDigits: intPtr(2)}, []Value{Int(-1), Float(-1), decimalValue("-1")}, []Value{Int(1), Uint(1), Float(1), decimalValue("1")}, Other},
		{"rounded equality", Options{MaximumFractionDigits: intPtr(0)}, []Value{Float(-1.2), decimalValue("-1.2")}, []Value{Float(1.3), decimalValue("1.3")}, One},
		{"zero", Options{}, []Value{Int(0), Float(math.Copysign(0, -1)), decimalValue("-0")}, []Value{Int(0), Uint(0), Float(0), decimalValue("0")}, Other},
		{"minimum signed integer", Options{}, []Value{Int(math.MinInt64), Float(math.MinInt64), BigInt(new(big.Int).SetInt64(math.MinInt64)), decimalValue("-9223372036854775808")}, []Value{Uint(1 << 63), Float(1 << 63), BigInt(new(big.Int).SetUint64(1 << 63)), decimalValue("9223372036854775808")}, Other},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rules, err := New(intltest.LocaleList(t, "en"), tc.options)
			if err != nil {
				t.Fatal(err)
			}
			for i, start := range tc.starts {
				for j, end := range tc.ends {
					got, err := rules.SelectRange(start, end)
					if err != nil || got != tc.want {
						t.Errorf("SelectRange(bridge %d, bridge %d) = %s, %v; want %s, nil", i, j, got, err, tc.want)
					}
				}
			}
		})
	}
}

func TestSelectRangeNotationIdentity(t *testing.T) {
	t.Parallel()
	for _, notation := range []Notation{ScientificNotation, EngineeringNotation} {
		for _, tc := range []struct {
			name, loc    string
			starts, ends []Value
			maxFraction  *int
			want         Category
		}{
			{"different magnitudes", "en", []Value{Int(1), Float(1), decimalValue("1")}, []Value{Int(1000), Float(1000), decimalValue("1000")}, nil, Other},
			{"different magnitudes two", "ar", []Value{Int(2), Float(2), decimalValue("2")}, []Value{Int(2000), Float(2000), decimalValue("2000")}, nil, Other},
			{"other control", "en", []Value{Int(2), Float(2), decimalValue("2")}, []Value{Int(2000), Float(2000), decimalValue("2000")}, nil, Other},
			{"rounded equality", "en", []Value{Float(1.1), decimalValue("1.1")}, []Value{Float(1.2), decimalValue("1.2")}, intPtr(0), One},
			{"rounded equality two", "ar", []Value{Float(2.1), decimalValue("2.1")}, []Value{Float(2.2), decimalValue("2.2")}, intPtr(0), Two},
		} {
			t.Run(string(notation)+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				rules, err := New(intltest.LocaleList(t, tc.loc), Options{
					Notation:              stringPtr(notation),
					MaximumFractionDigits: tc.maxFraction,
				})
				if err != nil {
					t.Fatal(err)
				}
				for i, start := range tc.starts {
					for j, end := range tc.ends {
						if got, err := rules.SelectRange(start, end); err != nil || got != tc.want {
							t.Errorf("SelectRange(bridges %d/%d) = %s, %v; want %s, nil", i, j, got, err, tc.want)
						}
					}
				}
			})
		}
	}
}

func TestOrdinalRangeCategoryPairs(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, loc  string
		start, end int64
		want       Category
	}{
		{"english sparse", "en", 1, 2, Other},
		{"welsh explicit", "cy", 1, 2, Two},
		{"welsh sparse", "cy", 3, 2, Other},
		{"english equal", "en", 2, 2, Two},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rules, err := New(intltest.LocaleList(t, tc.loc), Options{Type: stringPtr(Ordinal)})
			if err != nil {
				t.Fatal(err)
			}
			got, err := rules.SelectRange(Int(tc.start), Int(tc.end))
			if err != nil || got != tc.want {
				t.Fatalf("SelectRange(%d,%d) = %s, %v; want %s, nil", tc.start, tc.end, got, err, tc.want)
			}
		})
	}
}

func TestStripIfIntegerBridgeValues(t *testing.T) {
	t.Parallel()
	rules, err := New(intltest.LocaleList(t, "en"), Options{TrailingZeroDisplay: new("stripIfInteger")})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		n    int64
		text string
	}{
		{10, "10"}, {100, "100"}, {1000, "1000"}, {-100, "-100"},
	} {
		values := []Value{Int(tc.n), Float(float64(tc.n)), BigInt(big.NewInt(tc.n)), decimalValue(tc.text)}
		if tc.n >= 0 {
			values = append(values, Uint(uint64(tc.n)))
		}
		for i, value := range values {
			if got := rules.Select(value); got != Other {
				t.Errorf("Select(%s, bridge %d) = %s; want other", tc.text, i, got)
			}
			for j, end := range values {
				if got, err := rules.SelectRange(value, end); err != nil || got != Other {
					t.Errorf("SelectRange(%s, bridges %d/%d) = %s, %v; want other, nil", tc.text, i, j, got, err)
				}
			}
		}
	}
}
