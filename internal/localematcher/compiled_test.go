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

func TestResolveLocaleBestFitKeepsExtensionFromMatchedRequest(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"zh-Hant", "en"}, func(tag string) string {
		switch tag {
		case "zh-TW", "zh-Hant":
			return "zh-Hant-TW"
		default:
			return tag
		}
	})
	got := ResolveLocale(ResolveOptions{
		Algorithm:             AlgorithmBestFit,
		Matcher:               matcher,
		Requested:             []string{"zh-TW-u-ca-buddhist", "zh-TW-u-ca-gregory"},
		DefaultLocale:         "en",
		RelevantExtensionKeys: []string{"ca"},
		LocaleData: testLocaleData{
			"zh-Hant": {"ca": []string{"gregory", "buddhist"}},
		},
	})
	if got.Locale != "zh-TW-u-ca-buddhist" || got.DataLocale != "zh-Hant" {
		t.Fatalf("ResolveLocale() = %#v, want zh-TW-u-ca-buddhist / zh-Hant", got)
	}
	if got.Extensions["ca"] != "buddhist" {
		t.Fatalf("ResolveLocale().Extensions[ca] = %q, want buddhist", got.Extensions["ca"])
	}
}

func TestMatcherBestFitExactMatchKeepsFirstMatchedRequest(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"en", "fr"}, nil)
	got := matcher.Match([]string{"ban", "en-u-nu-thai", "fr-u-nu-latn"}, "de", AlgorithmBestFit)
	if got.Locale != "en" || got.Extension != "-u-nu-thai" || got.Distance != 40 {
		t.Fatalf(`Matcher.Match(best fit exact) = %#v, want locale "en", extension "-u-nu-thai", distance 40`, got)
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

func TestMatcherMapsDefaultFallbackToDataLocale(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"en-US"}, nil)
	got := matcher.Match([]string{"ban"}, "en", AlgorithmLookup)
	if got.Locale != "en" || got.DataLocale != "en-US" {
		t.Fatalf("Matcher.Match(lookup fallback) = %#v, want locale en data en-US", got)
	}

	got = matcher.Match([]string{"ban"}, "en", AlgorithmBestFit)
	if got.Locale != "en" || got.DataLocale != "en-US" {
		t.Fatalf("Matcher.Match(best fit fallback) = %#v, want locale en data en-US", got)
	}
}

func TestMatcherDerivesDefaultRegionAliasFromMaximizedBase(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"en"}, defaultRegionTestMaximizer)
	for _, algorithm := range []Algorithm{AlgorithmLookup, AlgorithmBestFit} {
		got := matcher.Match([]string{"en-US"}, "en", algorithm)
		if got.Locale != "en-US" || got.DataLocale != "en" {
			t.Errorf("Matcher.Match(en-US, %v) = %#v, want locale en-US data en", algorithm, got)
		}
	}
}

func TestMatcherBackedDefaultRegionWinsOverDerivedAlias(t *testing.T) {
	t.Parallel()

	for _, supported := range [][]string{{"en", "en-US"}, {"en-US", "en"}} {
		matcher := NewMatcher(supported, defaultRegionTestMaximizer)
		for _, algorithm := range []Algorithm{AlgorithmLookup, AlgorithmBestFit} {
			got := matcher.Match([]string{"en-US"}, "en", algorithm)
			if got.Locale != "en-US" || got.DataLocale != "en-US" {
				t.Errorf("NewMatcher(%v).Match(en-US, %v) = %#v, want locale/data en-US", supported, algorithm, got)
			}
		}
	}
}

func TestMatcherDoesNotExposeFullMaximizedTagAsAlias(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"en"}, defaultRegionTestMaximizer)
	for _, algorithm := range []Algorithm{AlgorithmLookup, AlgorithmBestFit} {
		got := matcher.Match([]string{"en-Latn-US"}, "en", algorithm)
		if got.Locale != "en" || got.DataLocale != "en" {
			t.Errorf("Matcher.Match(en-Latn-US, %v) = %#v, want locale/data en", algorithm, got)
		}
	}
}

func defaultRegionTestMaximizer(tag string) string {
	switch tag {
	case "en", "en-US", "en-Latn-US":
		return "en-Latn-US"
	default:
		return tag
	}
}

func TestLanguageRegionAliasUsesLocaleSubtagGrammar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		loc  string
		want string
		ok   bool
	}{
		{name: "alpha region", loc: "zh-Hant-HK", want: "zh-HK", ok: true},
		{name: "numeric region", loc: "es-Latn-419", want: "es-419", ok: true},
		{name: "following subtags", loc: "en-Latn-US-posix", want: "en-US-posix", ok: true},
		{name: "invalid script", loc: "en-123-US"},
		{name: "invalid region", loc: "en-Latn-12x"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := languageRegionAlias(tc.loc)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("languageRegionAlias(%q) = %q, %v; want %q, %v", tc.loc, got, ok, tc.want, tc.ok)
			}
		})
	}
}
