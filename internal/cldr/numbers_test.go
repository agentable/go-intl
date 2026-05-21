package cldr

import (
	"testing"
)

func TestNumberAccessors(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("en-US")
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	if got, want := loc.DefaultNumberingSystem(), "latn"; got != want {
		t.Fatalf("DefaultNumberingSystem = %q, want %q", got, want)
	}
	symbols := loc.NumberSymbols("latn")
	if symbols.Decimal != "." || symbols.Group != "," || symbols.Percent != "%" {
		t.Fatalf("NumberSymbols = %+v", symbols)
	}
	if got, want := loc.DecimalPattern("latn"), "#,##0.###"; got != want {
		t.Fatalf("DecimalPattern = %q, want %q", got, want)
	}
	if got, want := loc.PercentPattern("latn"), "#,##0%"; got != want {
		t.Fatalf("PercentPattern = %q, want %q", got, want)
	}
	if got, want := loc.CurrencyPattern("latn", "standard"), "¤#,##0.00"; got != want {
		t.Fatalf("CurrencyPattern = %q, want %q", got, want)
	}
	if got, want := loc.CurrencyPattern("latn", "accounting"), "¤#,##0.00;(¤#,##0.00)"; got != want {
		t.Fatalf("CurrencyPattern(accounting) = %q, want %q", got, want)
	}
	if got, want := loc.CompactPattern("latn", "short", 3, "one"), "0K"; got != want {
		t.Fatalf("CompactPattern(one) = %q, want %q", got, want)
	}
	if got, want := loc.CompactPattern("latn", "short", 3, "few"), "0K"; got != want {
		t.Fatalf("CompactPattern(few) = %q, want fallback %q", got, want)
	}
	if got, want := loc.CompactPattern("latn", "short", 6, "other"), "0M"; got != want {
		t.Fatalf("CompactPattern(6) = %q, want %q", got, want)
	}
	if got, want := loc.CompactPattern("latn", "long", 3, "other"), "0 thousand"; got != want {
		t.Fatalf("CompactPattern(long, 3) = %q, want %q", got, want)
	}
}
