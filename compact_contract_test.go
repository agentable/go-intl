package gointl_test

import (
	"testing"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

func TestCompactNumberFormatAndPluralRulesUseDistinctPluralInputs(t *testing.T) {
	t.Parallel()

	locales := mustLocaleList("ru")
	options := struct {
		number numberformat.Options
		rules  pluralrules.Options
	}{
		number: numberformat.Options{
			Notation:       gointl.String(numberformat.CompactNotation),
			CompactDisplay: gointl.String(numberformat.LongCompactDisplay),
		},
		rules: pluralrules.Options{
			Notation:       gointl.String(pluralrules.CompactNotation),
			CompactDisplay: gointl.String(pluralrules.LongCompactDisplay),
		},
	}

	format, err := numberformat.New(locales, options.number)
	if err != nil {
		t.Fatal(err)
	}
	rules, err := pluralrules.New(locales, options.rules)
	if err != nil {
		t.Fatal(err)
	}

	if got := format.Format(numberformat.Int(2000)); got != "2 тысячи" {
		t.Fatalf("NumberFormat compact Format(2000) = %q, want %q", got, "2 тысячи")
	}
	got := rules.Select(pluralrules.Int(2000))
	if got != pluralrules.Many {
		t.Fatalf("PluralRules compact Select(2000) = %s, want %s", got, pluralrules.Many)
	}
}
