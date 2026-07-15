package localematcher

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

func TestMatchDispatchesAlgorithm(t *testing.T) {
	t.Parallel()

	lookup := MatchWithMaximizer([]string{"zh-TW"}, []string{"zh", "zh-Hant"}, "en", AlgorithmLookup, cldrlocale.Maximize)
	if lookup.Locale != "zh" {
		t.Fatalf("Match(lookup) = %#v, want zh lookup truncation", lookup)
	}

	bestFit := MatchWithMaximizer([]string{"zh-TW"}, []string{"zh", "zh-Hant"}, "en", AlgorithmBestFit, cldrlocale.Maximize)
	if bestFit.Locale != "zh-Hant" {
		t.Fatalf("Match(best fit) = %#v, want zh-Hant distance match", bestFit)
	}
}
