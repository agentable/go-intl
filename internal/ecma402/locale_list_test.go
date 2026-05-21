package ecma402

import (
	"errors"
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

func TestRequestedLocaleStringsTreatsNilAndEmptyAsOmitted(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		locales locale.List
	}{
		{name: "nil", locales: nil},
		{name: "empty", locales: locale.List{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := RequestedLocaleStrings(tc.locales); got != nil {
				t.Fatalf("RequestedLocaleStrings(%s) = %v, want nil", tc.name, got)
			}
		})
	}
}

func TestRequestedLocaleStringsDedupesCanonicalLocalesAndClones(t *testing.T) {
	t.Parallel()

	locales := locale.List{
		locale.MustParse("en-us"),
		locale.MustParse("fr"),
		locale.MustParse("en-US"),
	}
	got := RequestedLocaleStrings(locales)
	want := []string{"en-US", "fr"}
	if !slices.Equal(got, want) {
		t.Fatalf("RequestedLocaleStrings() = %v, want %v", got, want)
	}

	got[0] = "mutated"
	next := RequestedLocaleStrings(locales)
	if !slices.Equal(next, want) {
		t.Fatalf("RequestedLocaleStrings() after caller mutation = %v, want %v", next, want)
	}
}

func TestValidationLocaleUsesFirstCanonicalRequestOrDefault(t *testing.T) {
	t.Parallel()

	got := ValidationLocale(locale.List{locale.MustParse("fr-ca"), locale.MustParse("en-US")})
	if got.String() != "fr-CA" {
		t.Fatalf("ValidationLocale(requested) = %q, want fr-CA", got.String())
	}

	got = ValidationLocale(nil)
	if got.String() != DefaultLocale() {
		t.Fatalf("ValidationLocale(nil) = %q, want %q", got.String(), DefaultLocale())
	}
}

func TestSupportedLocalesCanonicalizesBeforeFiltering(t *testing.T) {
	t.Parallel()

	requested := locale.List{
		locale.MustParse("fr-ca"),
		locale.MustParse("en-US"),
		locale.MustParse("fr-CA"),
		locale.MustParse("xh"),
	}
	got := SupportedLocales([]string{"en-US", "fr"}, requested, localematcher.AlgorithmLookup, nil)
	want := locale.List{locale.MustParse("fr-CA"), locale.MustParse("en-US")}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocales() length = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].String() != want[i].String() {
			t.Fatalf("SupportedLocales()[%d] = %q, want %q", i, got[i].String(), want[i].String())
		}
	}
}

func TestSupportedLocalesOfAppliesLocaleMatcherOption(t *testing.T) {
	t.Parallel()

	requested := locale.List{
		locale.MustParse("en-US"),
		locale.MustParse("fr-CA"),
	}
	got, err := SupportedLocalesOf("testformat", []string{"en", "fr"}, requested, "lookup", nil, ErrInvalidOption)
	if err != nil {
		t.Fatalf("SupportedLocalesOf(lookup) error = %v", err)
	}
	want := locale.List{locale.MustParse("en-US"), locale.MustParse("fr-CA")}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() length = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].String() != want[i].String() {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i].String())
		}
	}

	got, err = SupportedLocalesOf("testformat", []string{"en"}, requested, "fast", nil, ErrInvalidOption)
	if got != nil {
		t.Fatalf("SupportedLocalesOf(invalid matcher) = %v, want nil", got)
	}
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want ErrInvalidOption", err)
	}
	optErr, ok := errors.AsType[*OptionError](err)
	if !ok {
		t.Fatalf("SupportedLocalesOf(invalid matcher) error = %T, want OptionError", err)
	}
	if optErr.Owner != "testformat" || optErr.Name != "localeMatcher" || optErr.Value != "fast" || optErr.Locale != "en-US" {
		t.Fatalf("OptionError = %+v, want testformat localeMatcher fast en-US", optErr)
	}
}
