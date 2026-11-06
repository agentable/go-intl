package gointl

import (
	"errors"
	"math/big"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

func TestDecimalBridgesRejectBackendSpecialSpellings(t *testing.T) {
	t.Parallel()

	parsers := []struct {
		owner string
		parse func(string) error
	}{
		{"numberformat", func(s string) error { _, err := numberformat.ParseValue(s); return err }},
		{"pluralrules", func(s string) error { _, err := pluralrules.ParseValue(s); return err }},
	}
	for _, parser := range parsers {
		t.Run(parser.owner, func(t *testing.T) {
			t.Parallel()
			for _, input := range []string{"inf", "-inf", "+inf", "nan", "sNaN", "NAN", "infinity", "NaN42", "-NaN"} {
				t.Run(input, func(t *testing.T) {
					t.Parallel()
					err := parser.parse(input)
					if !errors.Is(err, ErrInvalidValue) {
						t.Fatalf("ParseValue(%q) error = %v, want ErrInvalidValue", input, err)
					}
					detail, ok := errors.AsType[*Error](err)
					if !ok || detail.Owner != parser.owner || detail.Name != "decimal" || detail.Value != input {
						t.Fatalf("ParseValue(%q) error details = %#v", input, detail)
					}
				})
			}
		})
	}
}

func TestDecimalBridgesStringNumericGrammar(t *testing.T) {
	t.Parallel()
	nf, err := numberformat.New(intltest.LocaleList(t, "en"), numberformat.Options{UseGrouping: String("false"), MaximumFractionDigits: Int(2)})
	if err != nil {
		t.Fatal(err)
	}
	pr, err := pluralrules.New(intltest.LocaleList(t, "en"), pluralrules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ input, text, category string }{
		{"", "0", "other"}, {" \t\n\r\v\f\u00a0\u1680\u2000\u2028\u2029\u202f\u205f\u3000\ufeff", "0", "other"},
		{" 1 ", "1", "one"}, {"0b1", "1", "one"}, {"0B10", "2", "other"},
		{"0o10", "8", "other"}, {"0O1", "1", "one"}, {"0x10", "16", "other"}, {"0Xf", "15", "other"},
		{"+.5", "0.5", "other"}, {"1.e0", "1", "one"}, {"-0.0", "-0", "other"},
		{"9007199254740993.1", "9007199254740993.1", "other"}, {"0x20000000000001", "9007199254740993", "other"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			n, err := numberformat.ParseValue(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if got := nf.Format(n); got != tc.text {
				t.Errorf("Format = %q, want %q", got, tc.text)
			}
			p, err := pluralrules.ParseValue(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if got := pr.Select(p).String(); got != tc.category {
				t.Errorf("Select = %q, want %q", got, tc.category)
			}
		})
	}
	for _, input := range []string{"\u00851", "\u180e1", "\u200b1", "1_000", "0x", "0b2", "0o8", "+0x1", "-0b1", "0x+1", "1e", ".", "1 2", "1e1.2"} {
		if _, err := numberformat.ParseValue(input); !errors.Is(err, ErrInvalidValue) {
			t.Errorf("NumberFormat ParseValue(%q): %v", input, err)
		}
		if _, err := pluralrules.ParseValue(input); !errors.Is(err, ErrInvalidValue) {
			t.Errorf("PluralRules ParseValue(%q): %v", input, err)
		}
	}
}

func TestDecimalBridgesNormalizeStringExtremes(t *testing.T) {
	t.Parallel()
	nf, err := numberformat.New(intltest.LocaleList(t, "en"), numberformat.Options{UseGrouping: String("false")})
	if err != nil {
		t.Fatal(err)
	}
	pr, err := pluralrules.New(intltest.LocaleList(t, "en"), pluralrules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ input, text string }{{"1e309", "∞"}, {"-1e309", "-∞"}, {"1e-400", "0"}, {"-1e-400", "-0"}} {
		v, err := numberformat.ParseValue(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if got := nf.Format(v); got != tc.text {
			t.Errorf("Format(%s) = %q", tc.input, got)
		}
		var text string
		for _, part := range nf.FormatToParts(v) {
			text += part.Value
		}
		if text != tc.text {
			t.Errorf("parts(%s) = %q", tc.input, text)
		}
		ranged, err := nf.FormatRange(v, v)
		if err != nil || ranged != "~"+tc.text {
			t.Errorf("range(%s) = %q, %v", tc.input, ranged, err)
		}
		parts, err := nf.FormatRangeToParts(v, v)
		if err != nil {
			t.Fatal(err)
		}
		text = ""
		for _, part := range parts {
			text += part.Value
			if part.Source != numberformat.SourceShared {
				t.Errorf("range source = %s", part.Source)
			}
		}
		if text != ranged {
			t.Errorf("range parts = %q, want %q", text, ranged)
		}
		p, err := pluralrules.ParseValue(tc.input)
		if err != nil {
			t.Fatal(err)
		}
		if got := pr.Select(p); got != pluralrules.Other {
			t.Errorf("Select(%s) = %s", tc.input, got)
		}
		if got, err := pr.SelectRange(p, p); err != nil || got != pluralrules.Other {
			t.Errorf("SelectRange(%s) = %s, %v", tc.input, got, err)
		}
	}
	integer := new(big.Int).Exp(big.NewInt(10), big.NewInt(309), nil)
	if got := nf.Format(numberformat.BigInt(integer)); got != integer.String() {
		t.Errorf("BigInt was range-normalized: %s", got)
	}
}
