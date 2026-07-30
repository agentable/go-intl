package numberformat

import (
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestStylePluralOperandMatchesNative(t *testing.T) {
	t.Parallel()

	// Node v26.0.0 / ICU 78.3 preserves visible fraction digits and notation
	// exponents when selecting currency names, currency placement, and units.
	tests := []struct {
		name   string
		locale string
		opts   Options
		want   string
		parts  []Part
	}{
		{
			name: "standard currency preserves visible fraction", locale: "en",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName), MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2)},
			want:  "1.00 US dollars",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollars"}},
		},
		{
			name: "standard unit preserves visible fraction", locale: "en",
			opts:  Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter"), UnitDisplay: stringPtr(LongUnitDisplay), MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2)},
			want:  "1.00 meters",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "meters"}},
		},
		{
			name: "scientific currency uses exponent", locale: "en",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName), Notation: stringPtr(ScientificNotation)},
			want:  "1E3 US dollars",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartExponentSeparator, Value: "E"}, {Type: PartExponentInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollars"}},
		},
		{
			name: "engineering currency uses exponent", locale: "en",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName), Notation: stringPtr(EngineeringNotation)},
			want:  "1E3 US dollars",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartExponentSeparator, Value: "E"}, {Type: PartExponentInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollars"}},
		},
		{
			name: "compact currency uses exponent", locale: "en",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName), Notation: stringPtr(CompactNotation)},
			want:  "1K US dollars",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartCompact, Value: "K"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollars"}},
		},
		{
			name: "scientific unit uses exponent", locale: "en",
			opts:  Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter"), UnitDisplay: stringPtr(LongUnitDisplay), Notation: stringPtr(ScientificNotation)},
			want:  "1E3 meters",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartExponentSeparator, Value: "E"}, {Type: PartExponentInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "meters"}},
		},
		{
			name: "engineering unit uses exponent", locale: "en",
			opts:  Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter"), UnitDisplay: stringPtr(LongUnitDisplay), Notation: stringPtr(EngineeringNotation)},
			want:  "1E3 meters",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartExponentSeparator, Value: "E"}, {Type: PartExponentInteger, Value: "3"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "meters"}},
		},
		{
			name: "compact unit uses exponent", locale: "en",
			opts:  Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter"), UnitDisplay: stringPtr(LongUnitDisplay), Notation: stringPtr(CompactNotation)},
			want:  "1K meters",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartCompact, Value: "K"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "meters"}},
		},
		{
			name: "sw scientific exponent selects plural placement", locale: "sw",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName), Notation: stringPtr(ScientificNotation)},
			want:  "dola za Marekani 1E3",
			parts: []Part{{Type: PartCurrency, Value: "dola za Marekani"}, {Type: PartLiteral, Value: " "}, {Type: PartInteger, Value: "1"}, {Type: PartExponentSeparator, Value: "E"}, {Type: PartExponentInteger, Value: "3"}},
		},
		{
			name: "sw compact exponent selects plural placement", locale: "sw",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName), Notation: stringPtr(CompactNotation)},
			want:  "dola za Marekani elfu\u00a01",
			parts: []Part{{Type: PartCurrency, Value: "dola za Marekani"}, {Type: PartLiteral, Value: " "}, {Type: PartCompact, Value: "elfu"}, {Type: PartLiteral, Value: "\u00a0"}, {Type: PartInteger, Value: "1"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, tc.locale)}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			value := Int(1000)
			if tc.opts.Notation == nil {
				value = Int(1)
			}
			if got := format.Format(value); got != tc.want {
				t.Errorf("Format() = %q, want %q", got, tc.want)
			}
			if got := format.FormatToParts(value); !reflect.DeepEqual(got, tc.parts) {
				t.Errorf("FormatToParts() = %#v, want %#v", got, tc.parts)
			}
		})
	}
}

func TestStylePluralOperandNegativeExponentMatchesNative(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:       stringPtr(UnitStyle),
		Unit:        stringPtr("meter"),
		UnitDisplay: stringPtr(LongUnitDisplay),
		Notation:    stringPtr(ScientificNotation),
	})
	if err != nil {
		t.Fatal(err)
	}
	value := Float(0.001)
	want := "1E-3 meters"
	if got := format.Format(value); got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
	wantParts := []Part{
		{Type: PartInteger, Value: "1"},
		{Type: PartExponentSeparator, Value: "E"},
		{Type: PartExponentMinusSign, Value: "-"},
		{Type: PartExponentInteger, Value: "3"},
		{Type: PartLiteral, Value: " "},
		{Type: PartUnit, Value: "meters"},
	}
	if got := format.FormatToParts(value); !reflect.DeepEqual(got, wantParts) {
		t.Errorf("FormatToParts() = %#v, want %#v", got, wantParts)
	}
}
