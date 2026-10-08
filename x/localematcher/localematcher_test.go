package localematcher_test

import (
	"testing"

	"github.com/agentable/go-intl/x/localematcher"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		requested     []string
		available     []string
		defaultLocale string
		algorithm     localematcher.Algorithm
		wantLocale    string
		wantData      string
		wantExtension string
		wantDistance  int
	}{
		{
			name:          "proposal example: fr-XX falls back to fr",
			requested:     []string{"fr-XX", "en"},
			available:     []string{"fr", "en"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmBestFit,
			wantLocale:    "fr",
			wantData:      "fr",
		},
		{
			// zh-Hant is data-backed, so its maximize alias zh-TW is
			// available as a derived fallback; the exact hit keeps the
			// requested locale and points DataLocale at the payload.
			name:          "best fit serves zh-TW from the zh-Hant payload",
			requested:     []string{"zh-TW"},
			available:     []string{"zh-Hans", "zh-Hant", "en"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmBestFit,
			wantLocale:    "zh-TW",
			wantData:      "zh-Hant",
			wantDistance:  0,
		},
		{
			name:          "best fit maximizes zh-HK onto zh-Hant",
			requested:     []string{"zh-HK"},
			available:     []string{"en", "zh-Hans", "zh-Hant", "ja"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmBestFit,
			wantLocale:    "zh-Hant",
			wantData:      "zh-Hant",
		},
		{
			name:          "best fit maximizes zh-SG onto zh-Hans",
			requested:     []string{"zh-SG"},
			available:     []string{"en", "zh-Hans", "zh-Hant", "ja"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmBestFit,
			wantLocale:    "zh-Hans",
			wantData:      "zh-Hans",
		},
		{
			// zh-Hant-HK is data-backed, so the zh-HK language-region
			// alias is available and keeps the concrete data locale.
			name:          "derived alias keeps the concrete data locale",
			requested:     []string{"zh-HK"},
			available:     []string{"zh-Hant-HK"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmBestFit,
			wantLocale:    "zh-HK",
			wantData:      "zh-Hant-HK",
		},
		{
			// Lookup truncates zh-SG to zh; best fit maximizes it to
			// zh-Hans-SG and lands on the zh-Hans payload instead.
			name:          "lookup truncates without maximizing",
			requested:     []string{"zh-SG"},
			available:     []string{"zh-Hans"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmLookup,
			wantLocale:    "zh",
			wantData:      "zh-Hans",
		},
		{
			name:          "best fit maximizes where lookup truncates",
			requested:     []string{"zh-SG"},
			available:     []string{"zh-Hans"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmBestFit,
			wantLocale:    "zh-Hans",
			wantData:      "zh-Hans",
		},
		{
			name:          "lookup matches and keeps the requested extension",
			requested:     []string{"en-US-u-ca-buddhist"},
			available:     []string{"en-US"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmLookup,
			wantLocale:    "en-US",
			wantData:      "en-US",
			wantExtension: "-u-ca-buddhist",
		},
		{
			name:          "lookup misses an unrelated language",
			requested:     []string{"fr-FR"},
			available:     []string{"zh-Hant"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmLookup,
			wantLocale:    "en",
			wantData:      "en",
		},
		{
			name:          "best fit rejects distant languages",
			requested:     []string{"ko"},
			available:     []string{"en", "zh-Hans", "zh-Hant", "ja"},
			defaultLocale: "en",
			algorithm:     localematcher.AlgorithmBestFit,
			wantLocale:    "en",
			wantData:      "en",
		},
		{
			name:          "empty request falls back to the default locale",
			requested:     nil,
			available:     []string{"de"},
			defaultLocale: "de",
			algorithm:     localematcher.AlgorithmBestFit,
			wantLocale:    "de",
			wantData:      "de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := localematcher.Match(tt.requested, tt.available, tt.defaultLocale, tt.algorithm)
			if got.Locale != tt.wantLocale {
				t.Errorf("Locale = %q, want %q", got.Locale, tt.wantLocale)
			}
			if got.DataLocale != tt.wantData {
				t.Errorf("DataLocale = %q, want %q", got.DataLocale, tt.wantData)
			}
			if got.Extension != tt.wantExtension {
				t.Errorf("Extension = %q, want %q", got.Extension, tt.wantExtension)
			}
			if tt.wantDistance != 0 && got.Distance != tt.wantDistance {
				t.Errorf("Distance = %d, want %d", got.Distance, tt.wantDistance)
			}
		})
	}
}

// Per-key best-fit: an i18n catalog whose keys exist in uneven locale subsets
// matches each key against the locales in which that key actually exists, and
// reads the payload from Result.DataLocale.
func TestMatchPerKeyCatalog(t *testing.T) {
	t.Parallel()

	catalog := map[string][]string{
		"title":   {"en", "zh-Hans", "zh-Hant", "ja"},
		"summary": {"en", "zh-Hant"},
	}

	title := localematcher.Match([]string{"zh-SG"}, catalog["title"], "en", localematcher.AlgorithmBestFit)
	if title.Locale != "zh-Hans" || title.DataLocale != "zh-Hans" {
		t.Errorf("title = (%q, %q), want (zh-Hans, zh-Hans)", title.Locale, title.DataLocale)
	}

	summary := localematcher.Match([]string{"zh-SG"}, catalog["summary"], "en", localematcher.AlgorithmBestFit)
	if summary.Locale != "zh" || summary.DataLocale != "zh-Hant" {
		t.Errorf("summary = (%q, %q), want (zh, zh-Hant)", summary.Locale, summary.DataLocale)
	}
}

func TestMatcherReuse(t *testing.T) {
	t.Parallel()

	matcher := localematcher.New([]string{"en", "zh-Hans", "zh-Hant", "ja"})

	if got := matcher.Match([]string{"zh-HK"}, "en", localematcher.AlgorithmBestFit); got.Locale != "zh-Hant" {
		t.Errorf("zh-HK: Locale = %q, want zh-Hant", got.Locale)
	}
	if got := matcher.Match([]string{"zh-SG"}, "en", localematcher.AlgorithmBestFit); got.Locale != "zh-Hans" {
		t.Errorf("zh-SG: Locale = %q, want zh-Hans", got.Locale)
	}
	if got := matcher.Match([]string{"ko"}, "en", localematcher.AlgorithmBestFit); got.Locale != "en" {
		t.Errorf("ko: Locale = %q, want en", got.Locale)
	}
}

func TestDefaultMatchingThreshold(t *testing.T) {
	t.Parallel()

	if localematcher.DefaultMatchingThreshold != 838 {
		t.Errorf("DefaultMatchingThreshold = %d, want 838", localematcher.DefaultMatchingThreshold)
	}
}
