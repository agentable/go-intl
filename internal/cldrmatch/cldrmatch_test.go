package cldrmatch

import (
	"testing"

	"golang.org/x/text/language"

	"github.com/agentable/go-intl/internal/cldr"
)

func TestNumberFallsBackToDataLocale(t *testing.T) {
	t.Parallel()

	loc := Number(language.MustParse("en-US"), "best fit")
	if got, want := loc.DefaultNumberingSystem(), "latn"; got != want {
		t.Fatalf("Number(en-US).DefaultNumberingSystem() = %q, want %q", got, want)
	}
}

func TestNumberFallsBackWhenGeneratedLocaleHasNoNumberData(t *testing.T) {
	t.Parallel()

	loc := Number(language.MustParse("fr-FR"), "best fit")
	if got, want := loc.DefaultNumberingSystem(), "latn"; got != want {
		t.Fatalf("Number(fr-FR).DefaultNumberingSystem() = %q, want %q", got, want)
	}
}

func TestDateFallsBackToDateDataLocale(t *testing.T) {
	t.Parallel()

	loc := Date(language.MustParse("zh-Hans-CN"), "best fit")
	if got := cldr.GregorianFor(loc).Months.Wide[0]; got != "一月" {
		t.Fatalf("Date(zh-Hans-CN) month name = %q, want zh-Hans data", got)
	}
}

func TestSupportedNumberLocalesHaveNumberData(t *testing.T) {
	t.Parallel()

	for _, tag := range cldr.NumberSupportedLocales() {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()

			loc := Number(language.MustParse(tag), "lookup")
			if got := loc.DefaultNumberingSystem(); got == "" {
				t.Fatalf("Number(%s).DefaultNumberingSystem() is empty", tag)
			}
		})
	}
}

func TestSupportedDateLocalesHaveGregorianData(t *testing.T) {
	t.Parallel()

	for _, tag := range cldr.DateSupportedLocales() {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()

			loc := Date(language.MustParse(tag), "lookup")
			if got := cldr.GregorianFor(loc).DateFormats[0]; got == "" {
				t.Fatalf("Date(%s).DateFormats[0] is empty", tag)
			}
		})
	}
}

func TestFallbackDefaultsHaveData(t *testing.T) {
	t.Parallel()

	if got := defaultLocale(KindNumber).DefaultNumberingSystem(); got == "" {
		t.Fatal("default number locale has empty numbering system")
	}
	if got := cldr.GregorianFor(defaultLocale(KindDate)).DateFormats[0]; got == "" {
		t.Fatal("default date locale has empty date format")
	}
}
