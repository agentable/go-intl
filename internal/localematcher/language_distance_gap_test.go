package localematcher

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

func TestBestFitRelatedLanguages(t *testing.T) {
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

	for _, tc := range cases {
		t.Run(tc.requested, func(t *testing.T) {
			t.Parallel()
			got := BestFitMatcherWithMaximizer([]string{tc.requested}, supported, "en", cldrlocale.Maximize)
			if got.Locale != tc.want {
				t.Fatalf("best fit %s = %#v, want CLDR neighbor %q", tc.requested, got, tc.want)
			}
		})
	}
}
