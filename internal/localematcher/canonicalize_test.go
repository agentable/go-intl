package localematcher

import (
	"errors"
	"slices"
	"testing"

	"github.com/agentable/go-intl/locale"
)

func TestCanonicalizeLocaleList(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("de-de-u-ca-gregorian")
	tests := []struct {
		name string
		in   any
		want []string
	}{
		{name: "nil", in: nil, want: []string{}},
		{name: "string", in: "en-us", want: []string{"en-US"}},
		{name: "locale", in: loc, want: []string{"de-DE-u-ca-gregory"}},
		{name: "string slice dedups", in: []string{"en-us", "en-US", "fr-fr"}, want: []string{"en-US", "fr-FR"}},
		{name: "locale slice", in: []locale.Locale{locale.MustParse("en-us"), locale.MustParse("fr-fr")}, want: []string{"en-US", "fr-FR"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := CanonicalizeLocaleList(tc.in)
			if err != nil {
				t.Fatalf("CanonicalizeLocaleList() error = %v", err)
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("CanonicalizeLocaleList() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestCanonicalizeLocaleListInvalid(t *testing.T) {
	t.Parallel()

	_, err := CanonicalizeLocaleList([]string{"en", "bad_locale"})
	if !errors.Is(err, locale.ErrInvalidLocale) {
		t.Fatalf("CanonicalizeLocaleList() error = %v, want ErrInvalidLocale", err)
	}
}
