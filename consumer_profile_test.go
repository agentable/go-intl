package gointl_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/segmenter"
)

type consumerProfile struct {
	SupportedLocalesOfExcludes map[string][]string `json:"supportedLocalesOfExcludes"`
	SupportedValuesOfIncludes  map[string][]string `json:"supportedValuesOfIncludes"`
	SupportedValuesOfExcludes  map[string][]string `json:"supportedValuesOfExcludes"`
	ReversedRanges             struct {
		NumberFormat struct {
			Locale string `json:"locale"`
			Start  string `json:"start"`
			End    string `json:"end"`
			Want   string `json:"want"`
		} `json:"numberFormat"`
		DateTimeFormat struct {
			Locale    string `json:"locale"`
			TimeZone  string `json:"timeZone"`
			Start     string `json:"start"`
			End       string `json:"end"`
			Want      string `json:"want"`
			WantStart string `json:"wantStart"`
			WantEnd   string `json:"wantEnd"`
		} `json:"dateTimeFormat"`
		PluralRules struct {
			Locale string `json:"locale"`
			Start  string `json:"start"`
			End    string `json:"end"`
			Want   string `json:"want"`
		} `json:"pluralRules"`
	} `json:"reversedRanges"`
}

func TestGoTypeScriptConsumerProfile(t *testing.T) {
	t.Parallel()

	profile := loadConsumerProfile(t)
	assertSupportedLocalesOfExcludes(t, profile)
	assertSupportedValuesOf(t, profile)
	assertConsumerReversedRanges(t, profile)
}

func loadConsumerProfile(t *testing.T) consumerProfile {
	t.Helper()

	var profile consumerProfile
	intltest.ReadFixture(t, "testdata/consumer/go-typescript/intl-profile.json", &profile)
	return profile
}

func assertSupportedLocalesOfExcludes(t *testing.T, profile consumerProfile) {
	t.Helper()

	for surface, tags := range profile.SupportedLocalesOfExcludes {
		t.Run("supportedLocalesOf/"+surface, func(t *testing.T) {
			t.Parallel()

			for _, tag := range tags {
				requested := intltest.LocaleList(t, tag)
				got, err := supportedLocalesOf(surface, requested)
				if err != nil {
					t.Fatalf("%s.SupportedLocalesOf(%q) error = %v", surface, tag, err)
				}
				if len(got) != 0 {
					t.Fatalf("%s.SupportedLocalesOf(%q) = %v, want unsupported", surface, tag, got.Strings())
				}
			}
		})
	}
}

func supportedLocalesOf(surface string, requested locale.List) (locale.List, error) {
	switch surface {
	case "collator":
		return collator.SupportedLocalesOf(requested, collator.Options{})
	case "datetimeformat":
		return datetimeformat.SupportedLocalesOf(requested, datetimeformat.Options{})
	case "segmenter":
		return segmenter.SupportedLocalesOf(requested, segmenter.Options{})
	default:
		panic("unsupported surface " + surface)
	}
}

func assertSupportedValuesOf(t *testing.T, profile consumerProfile) {
	t.Helper()

	for key, values := range profile.SupportedValuesOfIncludes {
		t.Run("supportedValuesOf/include/"+key, func(t *testing.T) {
			t.Parallel()

			got, ok := rootSupportedValues(key)
			if !ok {
				t.Fatalf("unsupported root supported-values profile key %q", key)
			}
			for _, value := range values {
				if !slices.Contains(got, value) {
					t.Fatalf("%s supported values missing %q in %v", key, value, got)
				}
			}
		})
	}
	for key, values := range profile.SupportedValuesOfExcludes {
		t.Run("supportedValuesOf/exclude/"+key, func(t *testing.T) {
			t.Parallel()

			got, ok := rootSupportedValues(key)
			if !ok {
				t.Fatalf("unsupported root supported-values profile key %q", key)
			}
			for _, value := range values {
				if slices.Contains(got, value) {
					t.Fatalf("%s supported values contains unsupported %q in %v", key, value, got)
				}
			}
		})
	}
}

func rootSupportedValues(key string) ([]string, bool) {
	switch key {
	case "calendar":
		return gointl.SupportedCalendars(), true
	case "collation":
		return gointl.SupportedCollations(), true
	case "currency":
		return gointl.SupportedCurrencies(), true
	case "numberingSystem":
		return gointl.SupportedNumberingSystems(), true
	case "timeZone":
		return gointl.SupportedTimeZones(), true
	case "unit":
		return gointl.SupportedUnits(), true
	default:
		return nil, false
	}
}

func assertConsumerReversedRanges(t *testing.T, profile consumerProfile) {
	t.Helper()

	t.Run("numberformat", func(t *testing.T) {
		t.Parallel()

		fixture := profile.ReversedRanges.NumberFormat
		format, err := numberformat.New(intltest.LocaleList(t, fixture.Locale), numberformat.Options{})
		if err != nil {
			t.Fatal(err)
		}
		start, err := numberformat.Decimal(fixture.Start)
		if err != nil {
			t.Fatal(err)
		}
		end, err := numberformat.Decimal(fixture.End)
		if err != nil {
			t.Fatal(err)
		}
		got, err := format.FormatRange(start, end)
		if err != nil {
			t.Fatalf("FormatRange(%q, %q) error = %v", fixture.Start, fixture.End, err)
		}
		if got != fixture.Want {
			t.Fatalf("FormatRange(%q, %q) = %q, want %q", fixture.Start, fixture.End, got, fixture.Want)
		}
	})

	t.Run("datetimeformat", func(t *testing.T) {
		t.Parallel()

		fixture := profile.ReversedRanges.DateTimeFormat
		format, err := datetimeformat.New(intltest.LocaleList(t, fixture.Locale), datetimeformat.Options{
			TimeZone: gointl.String(fixture.TimeZone),
			Year:     gointl.String(string(datetimeformat.NumericFieldStyle)),
			Month:    gointl.String(string(datetimeformat.NumericMonthStyle)),
			Day:      gointl.String(string(datetimeformat.NumericFieldStyle)),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := intltest.MustParseTime(t, time.RFC3339, fixture.Start)
		end := intltest.MustParseTime(t, time.RFC3339, fixture.End)
		got, err := format.FormatRange(start, end)
		if err != nil {
			t.Fatalf("FormatRange(%s, %s) error = %v", fixture.Start, fixture.End, err)
		}
		if got != fixture.Want {
			t.Fatalf("FormatRange(%s, %s) = %q, want %q", fixture.Start, fixture.End, got, fixture.Want)
		}
		parts, err := format.FormatRangeToParts(start, end)
		if err != nil {
			t.Fatalf("FormatRangeToParts(%s, %s) error = %v", fixture.Start, fixture.End, err)
		}
		var joined, startValues, endValues strings.Builder
		for _, part := range parts {
			joined.WriteString(part.Value)
			switch part.Source {
			case datetimeformat.SourceStartRange:
				startValues.WriteString(part.Value)
			case datetimeformat.SourceEndRange:
				endValues.WriteString(part.Value)
			case datetimeformat.SourceShared:
			}
		}
		if joined.String() != got {
			t.Fatalf("joined FormatRangeToParts(%s, %s) = %q, want %q", fixture.Start, fixture.End, joined.String(), got)
		}
		if got := startValues.String(); got != fixture.WantStart {
			t.Fatalf("startRange values = %q, want first argument %q", got, fixture.WantStart)
		}
		if got := endValues.String(); got != fixture.WantEnd {
			t.Fatalf("endRange values = %q, want second argument %q", got, fixture.WantEnd)
		}
	})

	t.Run("pluralrules", func(t *testing.T) {
		t.Parallel()

		fixture := profile.ReversedRanges.PluralRules
		rules, err := pluralrules.New(intltest.LocaleList(t, fixture.Locale), pluralrules.Options{})
		if err != nil {
			t.Fatal(err)
		}
		start, err := pluralrules.Decimal(fixture.Start)
		if err != nil {
			t.Fatal(err)
		}
		end, err := pluralrules.Decimal(fixture.End)
		if err != nil {
			t.Fatal(err)
		}
		got, err := rules.SelectRange(start, end)
		if err != nil {
			t.Fatal(err)
		}
		if got.String() != fixture.Want {
			t.Fatalf("SelectRange(%q, %q) = %s, want %s", fixture.Start, fixture.End, got, fixture.Want)
		}
	})
}
