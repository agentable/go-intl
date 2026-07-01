package localematcher

import (
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
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
			name:      "lookup preserves matched requested locales",
			supported: []string{"en", "fr", "zh-Hant"},
			requested: []locale.Locale{intltest.Locale(t, "en-US-u-ca-gregory"), intltest.Locale(t, "en-US-u-ca-gregory"), intltest.Locale(t, "fr-FR"), intltest.Locale(t, "zh-Hant-TW"), intltest.Locale(t, "de")},
			matcher:   AlgorithmLookup,
			want:      []string{"en-US-u-ca-gregory", "en-US-u-ca-gregory", "fr-FR", "zh-Hant-TW"},
		},
		{
			name:      "best fit preserves requested locale",
			supported: []string{"zh", "zh-Hant"},
			requested: []locale.Locale{intltest.Locale(t, "zh-TW")},
			matcher:   AlgorithmBestFit,
			want:      []string{"zh-TW"},
		},
		{
			name:      "lookup preserves requested locale matched by derived available locale",
			supported: []string{"zh-Hant-HK"},
			requested: []locale.Locale{intltest.Locale(t, "zh-HK-u-nu-hanidec")},
			matcher:   AlgorithmLookup,
			want:      []string{"zh-HK-u-nu-hanidec"},
		},
		{
			name:      "filters unsupported locale",
			supported: []string{"en", "fr"},
			requested: []locale.Locale{intltest.Locale(t, "de")},
			matcher:   AlgorithmLookup,
			want:      nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FilterLocales(tc.supported, tc.requested, tc.matcher)
			if got := localesToStrings(got); !slices.Equal(got, tc.want) {
				t.Fatalf("FilterLocales() = %v, want %v", got, tc.want)
			}
		})
	}
}

func localesToStrings(locales []locale.Locale) []string {
	out := make([]string, len(locales))
	for i, loc := range locales {
		out[i] = loc.String()
	}
	return out
}
