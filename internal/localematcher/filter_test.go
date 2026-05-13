package localematcher

import (
	"testing"

	"github.com/agentable/go-intl/locale"
)

func TestFilterLocales(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		supported []string
		requested []locale.Locale
		matcher   Algorithm
		want      []string
	}{
		{
			name:      "lookup preserves canonical requested locales",
			supported: []string{"en", "fr", "zh-Hant"},
			requested: []locale.Locale{
				locale.MustParse("en-US-u-ca-gregory"),
				locale.MustParse("en-US-u-ca-gregory"),
				locale.MustParse("fr-FR"),
				locale.MustParse("zh-Hant-TW"),
				locale.MustParse("de"),
			},
			matcher: AlgorithmLookup,
			want:    []string{"en-US-u-ca-gregory", "fr-FR", "zh-Hant-TW"},
		},
		{
			name:      "best fit preserves requested locale",
			supported: []string{"zh", "zh-Hant"},
			requested: []locale.Locale{locale.MustParse("zh-TW")},
			matcher:   AlgorithmBestFit,
			want:      []string{"zh-TW"},
		},
		{
			name:      "filters unsupported locale",
			supported: []string{"en", "fr"},
			requested: []locale.Locale{locale.MustParse("de")},
			matcher:   AlgorithmLookup,
			want:      nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FilterLocales(tc.supported, tc.requested, tc.matcher)
			if len(got) != len(tc.want) {
				t.Fatalf("FilterLocales() = %v, want %v", localesToStrings(got), tc.want)
			}
			for i := range tc.want {
				if got[i].String() != tc.want[i] {
					t.Fatalf("FilterLocales() = %v, want %v", localesToStrings(got), tc.want)
				}
			}
		})
	}
}

func localesToStrings(locales []locale.Locale) []string {
	out := make([]string, 0, len(locales))
	for _, loc := range locales {
		out = append(out, loc.String())
	}
	return out
}
