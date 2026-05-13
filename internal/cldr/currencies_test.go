package cldr

import (
	"testing"

	"golang.org/x/text/language"
)

func TestCurrencyAccessors(t *testing.T) {
	t.Parallel()

	usd := CurrencyDigits("USD")
	if usd != (CurrencyData{DefaultDigits: 2, CashDigits: 2, Rounding: 0}) {
		t.Fatalf("CurrencyDigits(USD) = %+v", usd)
	}
	jpy := CurrencyDigits("JPY")
	if jpy != (CurrencyData{DefaultDigits: 0, CashDigits: 0, Rounding: 0}) {
		t.Fatalf("CurrencyDigits(JPY) = %+v", jpy)
	}
	unknown := CurrencyDigits("XXX")
	if unknown.DefaultDigits != 2 || unknown.CashDigits != 2 {
		t.Fatalf("CurrencyDigits(XXX) = %+v, want DEFAULT 2 digits", unknown)
	}

	loc, ok := ResolveLocale(language.MustParse("en-US"))
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	if got, want := loc.CurrencyDisplayName("USD", "one"), "US dollar"; got != want {
		t.Fatalf("CurrencyDisplayName(USD, one) = %q, want %q", got, want)
	}
	if got, want := loc.CurrencySymbol("JPY", "narrow"), "¥"; got != want {
		t.Fatalf("CurrencySymbol(JPY, narrow) = %q, want %q", got, want)
	}
}
