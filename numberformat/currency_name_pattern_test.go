package numberformat

import (
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestCurrencyNamePatternMatchesNative(t *testing.T) {
	t.Parallel()

	// Node 26 / ICU 78.3 uses opposite Swahili placement patterns for one and
	// other when the formatted value has zero fraction digits.
	tests := []struct {
		name     string
		locale   string
		currency string
		opts     Options
		value    Value
		want     string
		parts    []Part
	}{
		{
			name: "sw positive name before number", locale: "sw", value: Int(123),
			want:  "dola za Marekani 123.00",
			parts: []Part{{Type: PartCurrency, Value: "dola za Marekani"}, {Type: PartLiteral, Value: " "}, {Type: PartInteger, Value: "123"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}},
		},
		{
			name: "sw negative name before signed number", locale: "sw", value: Int(-123),
			want:  "dola za Marekani -123.00",
			parts: []Part{{Type: PartCurrency, Value: "dola za Marekani"}, {Type: PartLiteral, Value: " "}, {Type: PartMinusSign, Value: "-"}, {Type: PartInteger, Value: "123"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}},
		},
		{
			name: "sw singular name and number-first pattern", locale: "sw", value: Int(1),
			opts:  Options{MinimumFractionDigits: intPtr(0), MaximumFractionDigits: intPtr(0)},
			want:  "1 dola ya Marekani",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "dola ya Marekani"}},
		},
		{
			name: "sw plural name and currency-first pattern", locale: "sw", value: Int(2),
			opts:  Options{MinimumFractionDigits: intPtr(0), MaximumFractionDigits: intPtr(0)},
			want:  "dola za Marekani 2",
			parts: []Part{{Type: PartCurrency, Value: "dola za Marekani"}, {Type: PartLiteral, Value: " "}, {Type: PartInteger, Value: "2"}},
		},
		{
			name: "en singular", locale: "en", value: Int(1),
			opts:  Options{MinimumFractionDigits: intPtr(0), MaximumFractionDigits: intPtr(0)},
			want:  "1 US dollar",
			parts: []Part{{Type: PartInteger, Value: "1"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollar"}},
		},
		{
			name: "en plural", locale: "en", value: Int(2),
			opts:  Options{MinimumFractionDigits: intPtr(0), MaximumFractionDigits: intPtr(0)},
			want:  "2 US dollars",
			parts: []Part{{Type: PartInteger, Value: "2"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollars"}},
		},
		{
			name: "ar preserves bidi literal and sign", locale: "ar", value: Int(-123),
			want:  "\u200e-123.00 دولار أمريكي",
			parts: []Part{{Type: PartLiteral, Value: "\u200e"}, {Type: PartMinusSign, Value: "-"}, {Type: PartInteger, Value: "123"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "دولار أمريكي"}},
		},
		{
			name: "accounting keeps numeric minus", locale: "en", value: Int(-123),
			opts:  Options{CurrencySign: stringPtr(AccountingCurrencySign)},
			want:  "-123.00 US dollars",
			parts: []Part{{Type: PartMinusSign, Value: "-"}, {Type: PartInteger, Value: "123"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollars"}},
		},
		{
			name: "missing numbering-system row uses locale default placement", locale: "en", value: Int(12),
			opts:  Options{NumberingSystem: stringPtr("mathbold")},
			want:  "𝟏𝟐.𝟎𝟎 US dollars",
			parts: []Part{{Type: PartInteger, Value: "𝟏𝟐"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "𝟎𝟎"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollars"}},
		},
		{
			name: "unknown code is not replaced with an unrelated name", locale: "en", currency: "ZZZ", value: Int(12),
			want:  "12.00 ZZZ",
			parts: []Part{{Type: PartInteger, Value: "12"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "ZZZ"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opts := tc.opts
			opts.Style = stringPtr(CurrencyStyle)
			currency := tc.currency
			if currency == "" {
				currency = "USD"
			}
			opts.Currency = stringPtr(currency)
			opts.CurrencyDisplay = stringPtr(CurrencyDisplayName)
			f, err := New(locale.List{intltest.Locale(t, tc.locale)}, opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := f.Format(tc.value); got != tc.want {
				t.Errorf("Format() = %q, want %q", got, tc.want)
			}
			gotParts := f.FormatToParts(tc.value)
			if !reflect.DeepEqual(gotParts, tc.parts) {
				t.Errorf("FormatToParts() = %#v, want %#v", gotParts, tc.parts)
			}
			if got := joinNumberParts(gotParts); got != tc.want {
				t.Errorf("parts join = %q, want %q", got, tc.want)
			}
		})
	}
}
