package gointl

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
	"github.com/agentable/go-intl/segmenter"
)

func TestGetCanonicalLocales(t *testing.T) {
	t.Parallel()

	enUS := locale.MustParse("en-us")
	enUSAgain := locale.MustParse("en-US")
	zh := locale.MustParse("zh-hans-cn-u-nu-latn")

	got := GetCanonicalLocales(locale.List{enUS, enUSAgain, zh})
	want := locale.List{enUS, zh}
	if len(got) != len(want) {
		t.Fatalf("GetCanonicalLocales() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].String() != want[i].String() {
			t.Fatalf("GetCanonicalLocales()[%d] = %q, want %q", i, got[i].String(), want[i].String())
		}
	}
}

func TestSupportedValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values func() []string
		want   string
	}{
		{"calendar", SupportedCalendars, "iso8601"},
		{"currency", SupportedCurrencies, "USD"},
		{"numberingSystem", SupportedNumberingSystems, "arab"},
		{"timeZone", SupportedTimeZones, "America/New_York"},
		{"unit", SupportedUnits, "meter"},
	}
	for _, tt := range tests {
		got := tt.values()
		if !slices.Contains(got, tt.want) {
			t.Fatalf("%s supported values missing %q in %v", tt.name, tt.want, got)
		}
		if !slices.IsSorted(got) {
			t.Fatalf("%s supported values = %v, want sorted", tt.name, got)
		}
	}
}

func TestSupportedNumberingSystemsECMA402SimpleDigits(t *testing.T) {
	t.Parallel()

	got := SupportedNumberingSystems()
	for _, want := range []string{"adlm", "arab", "arabext", "beng", "deva", "fullwide", "hanidec", "latn", "thai"} {
		if !slices.Contains(got, want) {
			t.Fatalf("SupportedNumberingSystems() missing %q in %v", want, got)
		}
	}
}

func TestSupportedCalendarsMatchActiveDateTimeFormat(t *testing.T) {
	t.Parallel()

	got := SupportedCalendars()
	want := []string{"gregory", "iso8601"}
	if !slices.Equal(got, want) {
		t.Fatalf("SupportedCalendars() = %v, want %v", got, want)
	}

	for _, calendar := range got {
		format, err := datetimeformat.New(locale.List{locale.MustParse("en-US")}, datetimeformat.Options{Calendar: calendar})
		if err != nil {
			t.Fatalf("datetimeformat.New(calendar=%q) error = %v, want advertised calendar accepted", calendar, err)
		}
		if resolved := format.ResolvedOptions().Calendar; resolved != calendar {
			t.Fatalf("datetimeformat.New(calendar=%q) resolved calendar = %q, want %q", calendar, resolved, calendar)
		}
	}

	for _, tc := range []struct {
		name    string
		locales locale.List
		options datetimeformat.Options
	}{
		{
			name:    "buddhist option",
			locales: locale.List{locale.MustParse("en-US")},
			options: datetimeformat.Options{Calendar: "buddhist"},
		},
		{
			name:    "buddhist locale",
			locales: locale.List{locale.MustParse("en-US-u-ca-buddhist")},
			options: datetimeformat.Options{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := datetimeformat.New(tc.locales, tc.options)
			if !errors.Is(err, ErrUnsupportedOption) {
				t.Fatalf("datetimeformat.New() error = %v, want ErrUnsupportedOption", err)
			}
		})
	}

	supported, err := datetimeformat.SupportedLocalesOf(
		locale.MustParseList("en-US-u-ca-buddhist", "en-US-u-ca-iso8601", "en-US-u-ca-gregory", "en-US"),
		datetimeformat.Options{},
	)
	if err != nil {
		t.Fatalf("datetimeformat.SupportedLocalesOf() error = %v", err)
	}
	wantLocales := []string{"en-US-u-ca-iso8601", "en-US-u-ca-gregory", "en-US"}
	if !slices.Equal(supported.Strings(), wantLocales) {
		t.Fatalf("datetimeformat.SupportedLocalesOf() = %v, want %v", supported.Strings(), wantLocales)
	}
}

func TestRootErrorSentinelsClassifyFormatterErrors(t *testing.T) {
	t.Parallel()

	if _, err := numberformat.New(locale.MustParseList("en-US"), numberformat.Options{Style: "bad"}); !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("numberformat.New(invalid style) error = %v, want root ErrInvalidOption", err)
	}

	if _, err := datetimeformat.New(locale.MustParseList("en-US"), datetimeformat.Options{Calendar: "buddhist"}); !errors.Is(err, ErrUnsupportedOption) {
		t.Fatalf("datetimeformat.New(unsupported calendar) error = %v, want root ErrUnsupportedOption", err)
	}

	dn, err := displaynames.New(locale.MustParseList("en-US"), displaynames.Options{Type: displaynames.Language})
	if err != nil {
		t.Fatalf("displaynames.New() error = %v", err)
	}
	if _, _, err := dn.Of("bad_code"); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("displaynames.Of(invalid code) error = %v, want root ErrInvalidCode", err)
	}
}

func TestRootErrorTextTeachesWithoutAbstractOperationNames(t *testing.T) {
	t.Parallel()

	displayNames, err := displaynames.New(locale.MustParseList("en-US"), displaynames.Options{Type: displaynames.Language})
	if err != nil {
		t.Fatal(err)
	}
	_, _, displayNameErr := displayNames.Of("bad_code")

	relative, err := relativetimeformat.New(locale.MustParseList("en-US"), relativetimeformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	_, relativeErr := relative.FormatInt(1, relativetimeformat.Unit("bad"))

	duration, err := durationformat.New(locale.MustParseList("en-US"), durationformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	_, durationErr := duration.Format(durationformat.Duration{Hours: 1, Minutes: -1})

	tests := []struct {
		name string
		err  error
	}{
		{name: "numberformat", err: errorFrom(func() (*numberformat.NumberFormat, error) {
			return numberformat.New(locale.MustParseList("en-US"), numberformat.Options{Style: "bad"})
		})},
		{name: "locale parse", err: errorFrom(func() (locale.Locale, error) {
			return locale.Parse("bad_locale")
		})},
		{name: "locale options", err: errorFrom(func() (locale.Locale, error) {
			return locale.New("en", locale.Options{HourCycle: "bad"})
		})},
		{name: "datetimeformat", err: errorFrom(func() (*datetimeformat.DateTimeFormat, error) {
			return datetimeformat.New(locale.MustParseList("en-US"), datetimeformat.Options{Calendar: "bad!"})
		})},
		{name: "pluralrules", err: errorFrom(func() (*pluralrules.PluralRules, error) {
			return pluralrules.New(locale.MustParseList("en-US"), pluralrules.Options{Type: pluralrules.Type(99)})
		})},
		{name: "listformat", err: errorFrom(func() (*listformat.ListFormat, error) {
			return listformat.New(locale.MustParseList("en-US"), listformat.Options{Type: "bad"})
		})},
		{name: "collator", err: errorFrom(func() (*collator.Collator, error) {
			return collator.New(locale.MustParseList("en-US"), collator.Options{Usage: collator.SearchUsage})
		})},
		{name: "segmenter", err: errorFrom(func() (*segmenter.Segmenter, error) {
			return segmenter.New(locale.MustParseList("en-US"), segmenter.Options{Granularity: "bad"})
		})},
		{name: "displaynames", err: displayNameErr},
		{name: "relativetimeformat", err: relativeErr},
		{name: "durationformat", err: durationErr},
	}
	banned := []string{"GetOption", "PartitionPattern", "ResolveLocale", "FormatNumericToString", "ToIntlMathematicalValue", "ecma402", "decimal:"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.err == nil {
				t.Fatal("error is nil")
			}
			text := tt.err.Error()
			if !strings.Contains(text, "expected ") || !strings.Contains(text, "; got ") {
				t.Fatalf("error text = %q, want expected/got guidance", text)
			}
			for _, token := range banned {
				if strings.Contains(text, token) {
					t.Fatalf("error text = %q, must not expose abstract operation %q", text, token)
				}
			}
		})
	}
}

func errorFrom[T any](fn func() (T, error)) error {
	_, err := fn()
	return err
}

func TestSupportedValuesReturnIndependentSlices(t *testing.T) {
	t.Parallel()

	got := SupportedUnits()
	if len(got) == 0 {
		t.Fatal("SupportedUnits() returned empty slice")
	}
	got[0] = "mutated"

	next := SupportedUnits()
	if slices.Contains(next, "mutated") {
		t.Fatalf("SupportedUnits() returned shared mutable storage: %v", next)
	}
	if !slices.IsSorted(next) {
		t.Fatalf("SupportedUnits() after mutation = %v, want sorted", next)
	}
}

func TestSupportedUnitsMatchECMA402(t *testing.T) {
	t.Parallel()

	got := SupportedUnits()
	want := []string{
		"acre", "bit", "byte", "celsius", "centimeter", "day", "degree",
		"fahrenheit", "fluid-ounce", "foot", "gallon", "gigabit", "gigabyte",
		"gram", "hectare", "hour", "inch", "kilobit", "kilobyte", "kilogram",
		"kilometer", "liter", "megabit", "megabyte", "meter", "microsecond",
		"mile", "mile-scandinavian", "milliliter", "millimeter", "millisecond",
		"minute", "month", "nanosecond", "ounce", "percent", "petabyte",
		"pound", "second", "stone", "terabit", "terabyte", "week", "yard",
		"year",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("SupportedUnits() = %v, want %v", got, want)
	}
}

func TestSupportedCollationsMatchActiveCollator(t *testing.T) {
	t.Parallel()

	got := SupportedCollations()
	if len(got) != 0 {
		t.Fatalf("SupportedCollations() = %v, want no advertised collation tailoring", got)
	}

	for _, tc := range []struct {
		name    string
		locales locale.List
		options collator.Options
	}{
		{
			name:    "search usage",
			locales: locale.List{locale.MustParse("en")},
			options: collator.Options{Usage: collator.SearchUsage},
		},
		{
			name:    "case first upper option",
			locales: locale.List{locale.MustParse("en")},
			options: collator.Options{CaseFirst: collator.UpperCaseFirst},
		},
		{
			name:    "case first lower option",
			locales: locale.List{locale.MustParse("en")},
			options: collator.Options{CaseFirst: collator.LowerCaseFirst},
		},
		{
			name:    "locale case first upper",
			locales: locale.List{locale.MustParse("en-u-kf-upper")},
			options: collator.Options{},
		},
		{
			name:    "locale case first lower",
			locales: locale.List{locale.MustParse("en-u-kf-lower")},
			options: collator.Options{},
		},
		{
			name:    "phonebook option",
			locales: locale.List{locale.MustParse("de")},
			options: collator.Options{Collation: "phonebk"},
		},
		{
			name:    "phonebook locale",
			locales: locale.List{locale.MustParse("de-u-co-phonebk")},
			options: collator.Options{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := collator.New(tc.locales, tc.options)
			if !errors.Is(err, ErrUnsupportedOption) {
				t.Fatalf("collator.New() error = %v, want ErrUnsupportedOption", err)
			}
		})
	}

	supported, err := collator.SupportedLocalesOf(
		locale.MustParseList("de-u-co-phonebk", "en-u-kf-upper", "en-u-kf-lower", "de", "en-u-kf-false"),
		collator.Options{},
	)
	if err != nil {
		t.Fatalf("collator.SupportedLocalesOf() error = %v", err)
	}
	want := []string{"de", "en-u-kf-false"}
	if !slices.Equal(supported.Strings(), want) {
		t.Fatalf("collator.SupportedLocalesOf() = %v, want %v", supported.Strings(), want)
	}
}

func TestRootConstructorAliases(t *testing.T) {
	t.Parallel()

	locales := locale.MustParseList("en-US")
	list, err := listformat.New(locales, listformat.Options{})
	if err != nil {
		t.Fatalf("listformat.New(en-US) error = %v", err)
	}
	requireRootListFormatAlias(list)
	if got := list.Format([]string{"A", "B"}); got != "A and B" {
		t.Fatalf("root ListFormat alias Format() = %q, want %q", got, "A and B")
	}

	relative, err := relativetimeformat.New(locales, relativetimeformat.Options{})
	if err != nil {
		t.Fatalf("relativetimeformat.New(en-US) error = %v", err)
	}
	requireRootRelativeTimeFormatAlias(relative)
	got, err := relative.FormatInt(-1, relativetimeformat.Second)
	if err != nil {
		t.Fatalf("root RelativeTimeFormat alias FormatInt() error = %v", err)
	}
	if got != "1 second ago" {
		t.Fatalf("root RelativeTimeFormat alias FormatInt() = %q, want %q", got, "1 second ago")
	}

	duration, err := durationformat.New(locales, durationformat.Options{})
	if err != nil {
		t.Fatalf("durationformat.New(en-US) error = %v", err)
	}
	requireRootDurationFormatAlias(duration)
	durationText, err := duration.Format(durationformat.Duration{Hours: 1, Minutes: 2})
	if err != nil {
		t.Fatalf("root DurationFormat alias Format() error = %v", err)
	}
	if durationText != "1 hr, 2 min" {
		t.Fatalf("root DurationFormat alias Format() = %q, want %q", durationText, "1 hr, 2 min")
	}
}

func requireRootListFormatAlias(*ListFormat) {}

func requireRootRelativeTimeFormatAlias(*RelativeTimeFormat) {}

func requireRootDurationFormatAlias(*DurationFormat) {}
