package localematcher

import "testing"

func TestMatcherMatchesLegacyFunctions(t *testing.T) {
	t.Parallel()

	maximizer := func(tag string) string {
		switch tag {
		case "zh-TW", "zh-Hant":
			return "zh-Hant-TW"
		default:
			return tag
		}
	}
	supported := []string{"en", "en-US", "fr", "zh-Hant"}
	matcher := NewMatcher(supported, maximizer)
	tests := []struct {
		name      string
		alg       Algorithm
		requested []string
	}{
		{name: "lookup", alg: AlgorithmLookup, requested: []string{"fr-CA", "en"}},
		{name: "best fit", alg: AlgorithmBestFit, requested: []string{"zh-TW", "en"}},
		{name: "extension", alg: AlgorithmLookup, requested: []string{"en-US-u-nu-thai"}},
		{name: "default", alg: AlgorithmBestFit, requested: []string{"ban"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := matcher.Match(tc.requested, "en", tc.alg)
			want := MatchWithMaximizer(tc.requested, supported, "en", tc.alg, maximizer)
			if got != want {
				t.Fatalf("Matcher.Match() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestResolveLocaleUsesCompiledMatcher(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"th", "en"}, nil)
	got := ResolveLocale(ResolveOptions{
		Algorithm:             AlgorithmLookup,
		Matcher:               matcher,
		Requested:             []string{"th-u-ca-buddhist-nu-thai"},
		DefaultLocale:         "en",
		RelevantExtensionKeys: []string{"ca", "nu"},
		OptionValues:          []Option{{Key: "ca", Value: "gregory"}},
		LocaleData: testLocaleData{
			"th": {
				"ca": []string{"buddhist", "gregory"},
				"nu": []string{"latn", "thai"},
			},
		},
	})
	if got.Locale != "th-u-nu-thai" || got.DataLocale != "th" {
		t.Fatalf("ResolveLocale() = %#v, want th-u-nu-thai / th", got)
	}
	if got.Extensions["ca"] != "gregory" || got.Extensions["nu"] != "thai" {
		t.Fatalf("ResolveLocale().Extensions = %#v, want ca=gregory nu=thai", got.Extensions)
	}
}

func TestMatcherMapsDerivedAvailableLocaleToDataLocale(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"zh-Hant-HK"}, func(tag string) string {
		switch tag {
		case "zh-HK", "zh-Hant-HK":
			return "zh-Hant-HK"
		default:
			return tag
		}
	})
	got := matcher.Match([]string{"zh-HK"}, "en", AlgorithmLookup)
	if got.Locale != "zh-HK" || got.DataLocale != "zh-Hant-HK" {
		t.Fatalf("Matcher.Match(lookup) = %#v, want locale zh-HK data zh-Hant-HK", got)
	}

	got = matcher.Match([]string{"zh-HK"}, "en", AlgorithmBestFit)
	if got.Locale != "zh-HK" || got.DataLocale != "zh-Hant-HK" {
		t.Fatalf("Matcher.Match(best fit) = %#v, want locale zh-HK data zh-Hant-HK", got)
	}
}
