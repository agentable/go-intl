package localematcher_test

import (
	"fmt"

	"github.com/agentable/go-intl/x/localematcher"
)

// Match mirrors the Stage 1 Intl.LocaleMatcher.match proposal: pick the best
// available locale for one request list.
func ExampleMatch() {
	result := localematcher.Match(
		[]string{"fr-XX", "en"},
		[]string{"fr", "en"},
		"en",
		localematcher.AlgorithmBestFit,
	)
	fmt.Println(result.Locale)
	// Output: fr
}

// New compiles a fixed available-locale set once, so repeated per-key
// matching — for example an i18n catalog whose keys exist in different
// locale subsets — stays on the cached path.
func ExampleNew() {
	matcher := localematcher.New([]string{"en", "zh-Hans", "zh-Hant", "ja"})

	for _, requested := range [][]string{{"zh-HK"}, {"zh-SG"}, {"ko"}} {
		result := matcher.Match(requested, "en", localematcher.AlgorithmBestFit)
		fmt.Println(result.Locale)
	}
	// Output:
	// zh-Hant
	// zh-Hans
	// en
}
