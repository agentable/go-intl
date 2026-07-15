package localematcher

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

func TestGeneratedLanguageMatchingProfileContract(t *testing.T) {
	t.Parallel()

	profile := defaultLanguageMatchingProfile()
	if len(profile.paradigmLocales) != 6 || len(profile.matchVariables) != 4 || len(profile.rules) != 378 {
		t.Fatalf("profile shape = paradigms:%d variables:%d rules:%d, want 6/4/378",
			len(profile.paradigmLocales), len(profile.matchVariables), len(profile.rules))
	}
	if last := profile.rules[len(profile.rules)-1]; last.desired != "*-*-*" || last.supported != "*-*-*" {
		t.Fatalf("final rule = %#v, want catch-all", last)
	}
	profile.paradigmLocales[0] = "mutated"
	profile.matchVariables[0].expandedRegions[0] = "XX"
	profile.rules[0].desired = "mutated"
	fresh := defaultLanguageMatchingProfile()
	if fresh.paradigmLocales[0] == "mutated" || fresh.matchVariables[0].expandedRegions[0] == "XX" || fresh.rules[0].desired == "mutated" {
		t.Fatal("defaultLanguageMatchingProfile returned mutable generated storage")
	}
}

func TestLanguageMatchingDistance(t *testing.T) {
	t.Parallel()

	profile := compileLanguageMatchingProfile(defaultLanguageMatchingProfile(), cldrlocale.Maximize)
	tests := []struct {
		name      string
		desired   string
		supported string
		want      int
	}{
		{name: "related language", desired: "gsw", supported: "de", want: 80},
		{name: "one way reverse", desired: "de", supported: "gsw", want: 840},
		{name: "norwegian macro language", desired: "nn", supported: "nb", want: 200},
		{name: "enUS variable and paradigm", desired: "en-CA", supported: "en-US", want: 39},
		{name: "americas variable and paradigm", desired: "es-KY", supported: "es-419", want: 39},
		{name: "cnsar variable", desired: "zh-Hant-HK", supported: "zh-Hant-MO", want: 40},
		{name: "unrelated language and script", desired: "ar-Arab-EG", supported: "zh-Hans-CN", want: 1340},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			maxDesired := cldrlocale.Maximize(tc.desired)
			maxSupported := cldrlocale.Maximize(tc.supported)
			got := profile.distance(maxDesired, maxSupported)
			if got != tc.want {
				t.Fatalf("distance(%q -> %q, %q -> %q) = %d, want %d; paradigms=%v/%v", tc.desired, maxDesired, tc.supported, maxSupported, got, tc.want, profile.paradigms[maxDesired], profile.paradigms[maxSupported])
			}
		})
	}
}

func TestLanguageMatchingParadigmTieBreak(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"en-AU", "en-GB"}, cldrlocale.Maximize)
	got := matcher.Match([]string{"en-IN"}, "en-AU", AlgorithmBestFit)
	if got.Locale != "en-GB" {
		t.Fatalf("best fit en-IN = %#v, want paradigm locale en-GB", got)
	}
}

func TestLanguageMatchingThresholdRejectsDistantLocale(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"zh"}, cldrlocale.Maximize)
	got := matcher.Match([]string{"ar"}, "en", AlgorithmBestFit)
	if got.Locale != "en" {
		t.Fatalf("best fit ar against zh = %#v, want default en beyond threshold", got)
	}
}

func BenchmarkCompileLanguageMatchingProfile(b *testing.B) {
	for b.Loop() {
		_ = compileLanguageMatchingProfile(defaultLanguageMatchingProfile(), cldrlocale.Maximize)
	}
}

func BenchmarkLanguageMatchingCachedDistance(b *testing.B) {
	matcher := NewMatcher([]string{"en", "nb", "de", "zh", "es"}, cldrlocale.Maximize)
	maxDesired := matcher.maximize("gsw")
	maxSupported := matcher.maximize("de")
	_ = matcher.cachedMatchingDistance("gsw", "de", maxDesired, maxSupported)
	b.ResetTimer()
	for b.Loop() {
		_ = matcher.cachedMatchingDistance("gsw", "de", maxDesired, maxSupported)
	}
}
