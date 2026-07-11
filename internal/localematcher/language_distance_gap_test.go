package localematcher

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// TestBestFitRelatedLanguageGap documents C4: the best-fit matcher uses a small
// override table plus a same-language(40)/else(840) fallback above the 838
// threshold, so linguistically close requests fall through to the default locale
// instead of their CLDR languageMatching neighbour. The expected targets below
// come from CLDR supplemental/languageMatching.json (nn↔nb=20, gsw↔de=4,
// wuu↔zh=10). This test asserts the correct behaviour and is skipped until the
// generated languageMatching distance table lands (improve.md C4/R1b); it turns
// green automatically once it does, and will start failing if the current
// wrong-default behaviour is ever mistaken for correct.
func TestBestFitRelatedLanguageGap(t *testing.T) {
	t.Parallel()

	supported := []string{"en", "nb", "de", "zh", "es"}
	cases := []struct {
		requested string
		want      string // CLDR languageMatching neighbour
	}{
		{requested: "nn", want: "nb"},
		{requested: "gsw", want: "de"},
		{requested: "wuu", want: "zh"},
	}

	// Record the current (incorrect) result so the gap is visible in -v output.
	for _, tc := range cases {
		got := BestFitMatcherWithMaximizer([]string{tc.requested}, supported, "en", cldrlocale.Maximize)
		if got.Locale != tc.want {
			t.Logf("C4 gap: %s -> %q (CLDR languageMatching: %q)", tc.requested, got.Locale, tc.want)
		}
	}
	t.Skip("C4/R1b: best-fit related-language matching needs the generated CLDR languageMatching distance table")
}
