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

func TestLocaleMatcherOption(t *testing.T) {
	t.Parallel()

	check := LocaleMatcherOption("fast")
	if check.Name != "localeMatcher" {
		t.Fatalf("LocaleMatcherOption().Name = %q, want localeMatcher", check.Name)
	}
	if invalid, ok := InvalidStringOption(check); !ok || invalid.Name != "localeMatcher" || invalid.Value != "fast" {
		t.Fatalf("InvalidStringOption(LocaleMatcherOption) = %+v, %t; want invalid localeMatcher fast", invalid, ok)
	}
	if _, ok := InvalidStringOption(LocaleMatcherOption("lookup")); ok {
		t.Fatal("InvalidStringOption(LocaleMatcherOption(lookup)) ok = true, want false")
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
		RelevantExtensionKeys: []string{"nu"},
		OptionValues:          []localematcher.Option{{Key: "nu", Value: "latn"}},
		LocaleData: constructorLocaleData{
			"th": {"nu": []string{"thai", "latn"}},
		},
	})

	if got.Locale.String() != "th" {
		t.Fatalf("ResolveConstructorLocale().Locale = %q, want th", got.Locale.String())
	}
	if got.DataLocale != "th" {
		t.Fatalf("ResolveConstructorLocale().DataLocale = %q, want th", got.DataLocale)
	}
	if got.Extension != "-u-nu-thai" {
		t.Fatalf("ResolveConstructorLocale().Extension = %q, want -u-nu-thai", got.Extension)
	}
	if got.Extensions["nu"] != "latn" {
		t.Fatalf("ResolveConstructorLocale().Extensions[nu] = %q, want latn", got.Extensions["nu"])
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
