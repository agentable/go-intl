package gointl

import (
	"testing"
	"time"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
)

func TestMessageformatIntegrationContract_PublicConsumerPattern(t *testing.T) {
	t.Parallel()

	locales, err := locale.ParseList("en-US")
	if err != nil {
		t.Fatal(err)
	}
	nf, err := numberformat.New(locales, numberformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := nf.Format(numberformat.Float(1234.5)); got == "" {
		t.Fatal("numberformat produced empty output")
	}
	dtf, err := datetimeformat.New(locales, datetimeformat.Options{
		Year:  datetimeformat.NumericFieldStyle,
		Month: datetimeformat.ShortMonthStyle,
		Day:   datetimeformat.NumericFieldStyle,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := dtf.Format(time.Date(2026, time.May, 8, 12, 0, 0, 0, time.UTC)); got == "" {
		t.Fatal("datetimeformat produced empty output")
	}
	pr, err := pluralrules.New(locales, pluralrules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := pr.Select(pluralrules.Int(2))
	if err != nil {
		t.Fatal(err)
	}
	if got != pluralrules.Other {
		t.Fatalf("pluralrules.SelectInt(2) = %s, want %s", got, pluralrules.Other)
	}
	lf, err := listformat.New(locales, listformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := lf.Format([]string{"A", "B"}); got == "" {
		t.Fatal("listformat produced empty output")
	}
	rtf, err := relativetimeformat.New(locales, relativetimeformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := rtf.FormatInt(-1, relativetimeformat.Second); err != nil || got == "" {
		t.Fatalf("relativetimeformat.FormatInt(-1, second) = %q, %v; want non-empty output", got, err)
	}
	df, err := durationformat.New(locales, durationformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := df.Format(durationformat.Duration{Hours: 1}); err != nil || got == "" {
		t.Fatalf("durationformat.Format({Hours: 1}) = %q, %v; want non-empty output", got, err)
	}
}
