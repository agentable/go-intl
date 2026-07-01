package collator

import (
	"slices"
	"testing"

	cldrcoll "github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
)

func TestCollatorLocaleDataReturnsFreshSlices(t *testing.T) {
	t.Parallel()

	adapter := collatorLocaleData{}
	tests := []struct {
		name   string
		locale string
		key    ecma402.UnicodeExtensionKey
		want   func() []string
	}{
		{
			name:   "collation",
			locale: "de",
			key:    ecma402.UnicodeExtensionKeyCollation,
			want: func() []string {
				return cldrcoll.SupportedCollationsForLocale("de")
			},
		},
		{
			name:   "numeric",
			locale: "en",
			key:    ecma402.UnicodeExtensionKeyNumeric,
			want: func() []string {
				return []string{"false", "true"}
			},
		},
		{
			name:   "case-first",
			locale: "en",
			key:    ecma402.UnicodeExtensionKeyCaseFirst,
			want: func() []string {
				return []string{string(FalseCaseFirst)}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first := adapter.For(tc.locale, string(tc.key))
			want := tc.want()
			if !slices.Equal(first, want) {
				t.Fatalf("collatorLocaleData.For(%q, %q) = %v, want %v", tc.locale, tc.key, first, want)
			}
			if len(first) == 0 {
				t.Fatalf("collatorLocaleData.For(%q, %q) returned no values", tc.locale, tc.key)
			}

			first[0] = "mutated"
			got := adapter.For(tc.locale, string(tc.key))
			if !slices.Equal(got, want) {
				t.Fatalf("collatorLocaleData.For(%q, %q) after caller mutation = %v, want %v", tc.locale, tc.key, got, want)
			}
		})
	}
}
