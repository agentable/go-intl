package ecma402

import (
	"errors"
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
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

	locales := locale.List{intltest.Locale(t, "en-us"), intltest.Locale(t, "fr"), intltest.Locale(t, "en-US")}
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

func TestCanonicalLocaleListDedupesCanonicalLocales(t *testing.T) {
	t.Parallel()

	locales := locale.List{intltest.Locale(t, "en-us"), intltest.Locale(t, "fr"), intltest.Locale(t, "en-US")}
	got := CanonicalLocaleList(locales)
	want := []string{"en-US", "fr"}
	if !slices.Equal(got.Strings(), want) {
		t.Fatalf("CanonicalLocaleList() = %v, want %v", got.Strings(), want)
	}
}

func TestValidationLocaleUsesFirstCanonicalRequestOrDefault(t *testing.T) {
	t.Parallel()

	got := ValidationLocale(locale.List{intltest.Locale(t, "fr-ca"), intltest.Locale(t, "en-US")})
	if got.String() != "fr-CA" {
		t.Fatalf("ValidationLocale(requested) = %q, want fr-CA", got.String())
	}

	got = ValidationLocale(nil)
	if got.String() != DefaultLocale() {
		t.Fatalf("ValidationLocale(nil) = %q, want %q", got.String(), DefaultLocale())
	}
}

func TestValidationLocaleUsesDefaultLocaleOverride(t *testing.T) {
	restore := OverrideDefaultLocaleForTest("fr")
	t.Cleanup(restore)

	got := ValidationLocale(nil)
	if got.String() != "fr" {
		t.Fatalf("ValidationLocale(nil) = %q, want fr", got.String())
	}
}

func TestSupportedLocalesCanonicalizesBeforeFiltering(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "fr-ca"), intltest.Locale(t, "en-US"), intltest.Locale(t, "fr-CA"), intltest.Locale(t, "xh")}
	got := SupportedLocales([]string{"en-US", "fr"}, requested, localematcher.AlgorithmLookup, nil)
	want := locale.List{intltest.Locale(t, "fr-CA"), intltest.Locale(t, "en-US")}
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

	requested := locale.List{intltest.Locale(t, "en-US"), intltest.Locale(t, "fr-CA")}
	matcher := "lookup"
	got, err := SupportedLocalesOf(SupportedLocalesOptions{
		Owner:         "testformat",
		Supported:     []string{"en", "fr"},
		Requested:     requested,
		LocaleMatcher: &matcher,
	})
	if err != nil {
		t.Fatalf("SupportedLocalesOf(lookup) error = %v", err)
	}
	want := locale.List{intltest.Locale(t, "en-US"), intltest.Locale(t, "fr-CA")}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() length = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].String() != want[i].String() {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i].String())
		}
	}

	got, err = SupportedLocalesOf(SupportedLocalesOptions{
		Owner:     "testformat",
		Supported: []string{"en", "fr"},
		Requested: requested,
	})
	if err != nil {
		t.Fatalf("SupportedLocalesOf(omitted matcher) error = %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf(omitted matcher) length = %d, want %d: %v", len(got), len(want), got)
	}

	matcher = "fast"
	got, err = SupportedLocalesOf(SupportedLocalesOptions{
		Owner:         "testformat",
		Supported:     []string{"en"},
		Requested:     requested,
		LocaleMatcher: &matcher,
	})
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
	if optErr.Owner != "testformat" || optErr.Name != "localeMatcher" || optErr.Value != "fast" || optErr.Locale != "en-US" || optErr.Expected != `one of "lookup", "best fit"` {
		t.Fatalf(`OptionError = %+v, want testformat localeMatcher fast en-US expected one of "lookup", "best fit"`, optErr)
	}
	matcher = ""
	got, err = SupportedLocalesOf(SupportedLocalesOptions{
		Owner:         "testformat",
		Supported:     []string{"en"},
		Requested:     requested,
		LocaleMatcher: &matcher,
	})
	if got != nil {
		t.Fatalf("SupportedLocalesOf(empty matcher) = %v, want nil", got)
	}
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("SupportedLocalesOf(empty matcher) error = %v, want ErrInvalidOption", err)
	}
}
