package cldrmatch

import (
	"testing"

	"golang.org/x/text/language"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
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
	if got := cldrdate.GregorianFor(loc).Months.Wide[0]; got != "一月" {
		t.Fatalf("Date(zh-Hans-CN) month name = %q, want zh-Hans data", got)
	}
}

func TestSupportedNumberLocalesHaveNumberData(t *testing.T) {
	t.Parallel()

	for _, tag := range cldrnumber.SupportedLocales() {
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

	for _, tag := range cldrdate.SupportedLocales() {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()

			loc := Date(language.MustParse(tag), "lookup")
			if got := cldrdate.GregorianFor(loc).DateFormats[0]; got == "" {
				t.Fatalf("Date(%s).DateFormats[0] is empty", tag)
			}
		})
	}
}

func TestFallbackDefaultsHaveData(t *testing.T) {
	t.Parallel()

	defaultNumber, _ := cldrnumber.ResolveLocale(defaultLocale(KindNumber))
	if got := defaultNumber.DefaultNumberingSystem(); got == "" {
		t.Fatal("default number locale has empty numbering system")
	}
	defaultDate, _ := cldrdate.ResolveLocale(defaultLocale(KindDate))
	if got := cldrdate.GregorianFor(defaultDate).DateFormats[0]; got == "" {
		t.Fatal("default date locale has empty date format")
	}
}

func TestFallbackDefaultsCanBeUndefinedWhenEnglishIsMissing(t *testing.T) {
	t.Parallel()

	if got := newLocaleSet([]string{"fr"}).defaultLocale(); got != "" {
		t.Fatalf("localeSet without en default = %q, want empty", got)
	}
}

func TestResolveFallsBackToDefaultForUnsupportedLocale(t *testing.T) {
	t.Parallel()

	loc := Date(language.MustParse("ban"), "lookup")
	if got := cldrdate.GregorianFor(loc).DateFormats[0]; got == "" {
		t.Fatalf("resolve(ban, date) = %v with empty Gregorian data", loc)
	}
}

func TestSupportedLocalesAndDirectResolution(t *testing.T) {
	t.Parallel()

	if len(supportedLocales(KindNumber)) == 0 || len(supportedLocales(KindDate)) == 0 {
		t.Fatal("supportedLocales returned an empty generated data set")
	}
	if ok := directString("ban", KindNumber); ok {
		t.Fatal("directString(ban, number) ok = true, want unsupported")
	}
	if ok := directString("en", KindNumber); !ok {
		t.Fatalf("directString(en, number) = %v; want number data", ok)
	}
}
