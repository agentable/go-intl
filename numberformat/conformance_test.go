package numberformat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentable/go-intl/locale"
)

type formatFixture struct {
	Name                  string `json:"name"`
	Locale                string `json:"locale"`
	Style                 string `json:"style,omitempty"`
	Currency              string `json:"currency,omitempty"`
	Notation              string `json:"notation,omitempty"`
	MaximumFractionDigits *int   `json:"maximumFractionDigits,omitempty"`
	Input                 any    `json:"input"`
	Want                  string `json:"want"`
}

func TestNumberFormatConformanceFixtures(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("testdata", "format.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []formatFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()

			var opts Options
			if fixture.Style != "" {
				opts.Style = Style(fixture.Style)
			}
			if fixture.Currency != "" {
				opts.Currency = CurrencyCode(fixture.Currency)
			}
			if fixture.Notation != "" {
				opts.Notation = Notation(fixture.Notation)
			}
			if fixture.MaximumFractionDigits != nil {
				opts.FractionDigits = MaximumFractionDigits(*fixture.MaximumFractionDigits)
			}
			format, err := New(locale.MustParse(fixture.Locale), opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(fixture.Input); got != fixture.Want {
				t.Fatalf("Format(%v) = %q, want %q", fixture.Input, got, fixture.Want)
			}
		})
	}
}
