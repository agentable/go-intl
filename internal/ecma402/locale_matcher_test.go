package ecma402

import (
	"testing"

	"github.com/agentable/go-intl/internal/localematcher"
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
