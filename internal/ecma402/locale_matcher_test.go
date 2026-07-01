package ecma402

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

func TestLocaleMatcherAlgorithm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		want  localematcher.Algorithm
		ok    bool
	}{
		{value: "", want: localematcher.AlgorithmBestFit, ok: true},
		{value: "best fit", want: localematcher.AlgorithmBestFit, ok: true},
		{value: "lookup", want: localematcher.AlgorithmLookup, ok: true},
		{value: "bogus", ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.value, func(t *testing.T) {
			t.Parallel()

			got, ok := LocaleMatcherAlgorithm(tc.value)
			if ok != tc.ok {
				t.Fatalf("LocaleMatcherAlgorithm(%q) ok = %t, want %t", tc.value, ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("LocaleMatcherAlgorithm(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestLocaleMatcherOptionInput(t *testing.T) {
	t.Parallel()

	check := LocaleMatcherOptionInput("fast", true)
	if check.Name != "localeMatcher" {
		t.Fatalf("LocaleMatcherOptionInput().Name = %q, want localeMatcher", check.Name)
	}
	if invalid, ok := InvalidStringOption(check); !ok || invalid.Name != "localeMatcher" || invalid.Value != "fast" {
		t.Fatalf("InvalidStringOption(LocaleMatcherOptionInput) = %+v, %t; want invalid localeMatcher fast", invalid, ok)
	}
	if _, ok := InvalidStringOption(LocaleMatcherOptionInput("", false)); ok {
		t.Fatal("InvalidStringOption(LocaleMatcherOptionInput(empty omitted)) ok = true, want false")
	}
	if invalid, ok := InvalidStringOption(LocaleMatcherOptionInput("", true)); !ok || invalid.Name != "localeMatcher" || invalid.Value != "" {
		t.Fatalf("InvalidStringOption(LocaleMatcherOptionInput(empty present)) = %+v, %t; want invalid localeMatcher empty", invalid, ok)
	}
	if _, ok := InvalidStringOption(LocaleMatcherOptionInput("lookup", true)); ok {
		t.Fatal("InvalidStringOption(LocaleMatcherOptionInput(lookup present)) ok = true, want false")
	}
}

type constructorLocaleData map[string]map[string][]string

func (d constructorLocaleData) For(loc, key string) []string {
	if keys, ok := d[loc]; ok {
		return keys[key]
	}
	return nil
}

func TestResolveConstructorLocaleAppliesMatcherAndRelevantExtensions(t *testing.T) {
	t.Parallel()

	fallback := intltest.Locale(t, "en")
	requested := locale.List{intltest.Locale(t, "th-u-nu-thai"), intltest.Locale(t, "en")}
	matcher := localematcher.NewMatcher([]string{"th", "en"}, nil)

	got := ResolveConstructorLocale(ConstructorLocaleOptions{
		Locales:               requested,
		Fallback:              fallback,
		LocaleMatcher:         "lookup",
		Matcher:               matcher,
		RelevantExtensionKeys: []UnicodeExtensionKey{UnicodeExtensionKeyNumberingSystem},
		OptionValues:          []UnicodeExtensionOption{{Key: UnicodeExtensionKeyNumberingSystem, Value: "latn"}},
		LocaleData: constructorLocaleData{
			"th": {string(UnicodeExtensionKeyNumberingSystem): []string{"thai", "latn"}},
		},
	})

	if got.Locale.String() != "th" {
		t.Fatalf("ResolveConstructorLocale().Locale = %q, want th", got.Locale.String())
	}
	if got.DataLocale != "th" {
		t.Fatalf("ResolveConstructorLocale().DataLocale = %q, want th", got.DataLocale)
	}
	if got.Extensions[UnicodeExtensionKeyNumberingSystem] != "latn" {
		t.Fatalf("ResolveConstructorLocale().Extensions[nu] = %q, want latn", got.Extensions[UnicodeExtensionKeyNumberingSystem])
	}
}

func TestResolveConstructorLocaleDispatchesLocaleMatcher(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "zh-TW")}
	fallback := intltest.Locale(t, "en")
	matcher := localematcher.NewMatcher([]string{"zh", "zh-Hant", "en"}, nil)

	lookup := ResolveConstructorLocale(ConstructorLocaleOptions{
		Locales:       requested,
		Fallback:      fallback,
		LocaleMatcher: "lookup",
		Matcher:       matcher,
	})
	if lookup.Locale.String() != "zh" {
		t.Fatalf("ResolveConstructorLocale(lookup).Locale = %q, want zh", lookup.Locale.String())
	}

	bestFit := ResolveConstructorLocale(ConstructorLocaleOptions{
		Locales:       requested,
		Fallback:      fallback,
		LocaleMatcher: "best fit",
		Matcher:       matcher,
	})
	if bestFit.Locale.String() != "zh-Hant" {
		t.Fatalf("ResolveConstructorLocale(best fit).Locale = %q, want zh-Hant", bestFit.Locale.String())
	}
}

func TestResolveConstructorLocaleFallsBackWhenResolvedLocaleCannotParse(t *testing.T) {
	restore := OverrideDefaultLocaleForTest("bad_locale")
	t.Cleanup(restore)

	fallback := intltest.Locale(t, "en")
	got := ResolveConstructorLocale(ConstructorLocaleOptions{
		Fallback: fallback,
	})

	if got.Locale.String() != "en" {
		t.Fatalf("ResolveConstructorLocale().Locale = %q, want fallback en", got.Locale.String())
	}
	if got.DataLocale != "bad_locale" {
		t.Fatalf("ResolveConstructorLocale().DataLocale = %q, want bad_locale", got.DataLocale)
	}
}

func TestResolveDataLocale(t *testing.T) {
	restore := OverrideDefaultLocaleForTest("en")
	t.Cleanup(restore)

	resolve := func(tag string) (string, bool) {
		switch tag {
		case "fr":
			return "data-fr", true
		case "en":
			return "data-en", true
		default:
			return "", false
		}
	}

	if got := ResolveDataLocale(ConstructorLocaleResolution{DataLocale: "fr"}, resolve); got != "data-fr" {
		t.Fatalf("ResolveDataLocale(fr) = %q, want data-fr", got)
	}
	if got := ResolveDataLocale(ConstructorLocaleResolution{DataLocale: "missing"}, resolve); got != "data-en" {
		t.Fatalf("ResolveDataLocale(missing) = %q, want data-en", got)
	}
}

func TestResolveDataLocaleTag(t *testing.T) {
	restore := OverrideDefaultLocaleForTest("en")
	t.Cleanup(restore)

	if got := ResolveDataLocaleTag(ConstructorLocaleResolution{DataLocale: "fr"}); got != "fr" {
		t.Fatalf("ResolveDataLocaleTag(fr) = %q, want fr", got)
	}
	if got := ResolveDataLocaleTag(ConstructorLocaleResolution{}); got != "en" {
		t.Fatalf("ResolveDataLocaleTag(empty) = %q, want en", got)
	}
}
