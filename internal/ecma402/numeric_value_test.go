package ecma402

import (
	"math"
	"math/big"
	"testing"

	"github.com/agentable/go-intl/internal/decimal"
)

func TestNumericValueConstructorsPreserveBridgeKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       NumericValue
		wantKind    NumericValueKind
		wantDecimal string
		wantInt64   int64
		wantUint64  uint64
	}{
		{
			name:        "decimal",
			value:       DecimalNumericValue(decimal.FromInt64(12)),
			wantKind:    NumericValueDecimal,
			wantDecimal: "12",
		},
		{
			name:        "int64",
			value:       Int64NumericValue(-42),
			wantKind:    NumericValueInt64,
			wantDecimal: "-42",
			wantInt64:   -42,
		},
		{
			name:        "uint64",
			value:       Uint64NumericValue(1<<63 + 5),
			wantKind:    NumericValueUint64,
			wantDecimal: "9223372036854775813",
			wantUint64:  1<<63 + 5,
		},
		{
			name:        "float64",
			value:       Float64NumericValue(1.25),
			wantKind:    NumericValueDecimal,
			wantDecimal: "1.25",
		},
		{
			name:        "big int",
			value:       BigIntNumericValue(big.NewInt(-9001)),
			wantKind:    NumericValueDecimal,
			wantDecimal: "-9001",
		},
		{
			name:        "nil big int",
			value:       BigIntNumericValue(nil),
			wantKind:    NumericValueDecimal,
			wantDecimal: "0",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.value.Kind != tc.wantKind {
				t.Fatalf("Kind = %v, want %v", tc.value.Kind, tc.wantKind)
			}
			if got := tc.value.Decimal.String(); got != tc.wantDecimal {
				t.Fatalf("Decimal = %q, want %q", got, tc.wantDecimal)
			}
			if tc.value.Int64 != tc.wantInt64 {
				t.Fatalf("Int64 = %d, want %d", tc.value.Int64, tc.wantInt64)
			}
			if tc.value.Uint64 != tc.wantUint64 {
				t.Fatalf("Uint64 = %d, want %d", tc.value.Uint64, tc.wantUint64)
			}
		})
	}
}

func TestInt64Magnitude(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value int64
		want  uint64
	}{
		{name: "zero", value: 0, want: 0},
		{name: "positive", value: 42, want: 42},
		{name: "negative", value: -42, want: 42},
		{name: "min", value: math.MinInt64, want: 1 << 63},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := Int64Magnitude(tc.value); got != tc.want {
				t.Fatalf("Int64Magnitude(%d) = %d, want %d", tc.value, got, tc.want)
			}
		})
	}
}
