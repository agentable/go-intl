package localematcher

import "testing"

func TestMatchDispatchesAlgorithm(t *testing.T) {
	t.Parallel()

	lookup := Match([]string{"zh-TW"}, []string{"zh", "zh-Hant"}, "en", AlgorithmLookup)
	if lookup.Locale != "zh" {
		t.Fatalf("Match(lookup) = %#v, want zh lookup truncation", lookup)
	}

	bestFit := Match([]string{"zh-TW"}, []string{"zh", "zh-Hant"}, "en", AlgorithmBestFit)
	if bestFit.Locale != "zh-Hant" {
		t.Fatalf("Match(best fit) = %#v, want zh-Hant distance match", bestFit)
	}
}
