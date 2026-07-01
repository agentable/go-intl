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
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
	"github.com/agentable/go-intl/segmenter"
)

func TestGetCanonicalLocales(t *testing.T) {
	t.Parallel()

	enUS := intltest.Locale(t, "en-us")
	enUSAgain := intltest.Locale(t, "en-US")
	zh := intltest.Locale(t, "zh-hans-cn-u-nu-latn")

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
		testcontract.AssertStringSliceContainsAll(t, tt.name+" supported values", got, tt.want)
		testcontract.AssertStringSliceSortedUnique(t, tt.name, got)
	}
}

func TestSupportedNumberingSystemsECMA402SimpleDigits(t *testing.T) {
	t.Parallel()

	got := SupportedNumberingSystems()
	testcontract.AssertStringSliceContainsAll(t, "SupportedNumberingSystems()", got,
		"adlm", "arab", "arabext", "beng", "deva", "fullwide", "hanidec", "latn", "thai",
	)
}

func TestSupportedCalendarsMatchActiveDateTimeFormat(t *testing.T) {
	t.Parallel()

	got := SupportedCalendars()
	want := []string{"gregory", "iso8601"}
	if !slices.Equal(got, want) {
		t.Fatalf("SupportedCalendars() = %v, want %v", got, want)
	}

	for _, calendar := range got {
		format, err := datetimeformat.New(locale.List{intltest.Locale(t, "en-US")}, datetimeformat.Options{Calendar: String(calendar)})
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
			locales: locale.List{intltest.Locale(t, "en-US")},
			options: datetimeformat.Options{Calendar: String("buddhist")},
		},
		{
			name:    "buddhist locale",
			locales: locale.List{intltest.Locale(t, "en-US-u-ca-buddhist")},
			options: datetimeformat.Options{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := datetimeformat.New(tc.locales, tc.options)
			if err != nil {
				t.Fatalf("datetimeformat.New() error = %v", err)
			}
			if got := format.ResolvedOptions().Calendar; got != "gregory" {
				t.Fatalf("datetimeformat.New() resolved calendar = %q, want gregory fallback", got)
			}
		})
	}

	supported, err := datetimeformat.SupportedLocalesOf(intltest.LocaleList(t, "en-US-u-ca-buddhist", "en-US-u-ca-iso8601", "en-US-u-ca-gregory", "en-US"), datetimeformat.Options{})
	if err != nil {
		t.Fatalf("datetimeformat.SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "datetimeformat.SupportedLocalesOf()", supported, []string{"en-US-u-ca-buddhist", "en-US-u-ca-iso8601", "en-US-u-ca-gregory", "en-US"})
}

func TestRootErrorSentinelsClassifyFormatterErrors(t *testing.T) {
	t.Parallel()

	if _, err := numberformat.New(intltest.LocaleList(t, "en-US"), numberformat.Options{Style: String("bad")}); !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("numberformat.New(invalid style) error = %v, want root ErrInvalidOption", err)
	}

	if _, err := datetimeformat.New(intltest.LocaleList(t, "en-US"), datetimeformat.Options{TimeZone: String("Mars/Olympus")}); !errors.Is(err, ErrUnsupportedOption) {
		t.Fatalf("datetimeformat.New(unsupported timeZone) error = %v, want root ErrUnsupportedOption", err)
	} else if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("datetimeformat.New(unsupported timeZone) error = %v, want errors.ErrUnsupported", err)
	}

	dn, err := displaynames.New(intltest.LocaleList(t, "en-US"), displaynames.Options{Type: String(displaynames.Language)})
	if err != nil {
		t.Fatalf("displaynames.New() error = %v", err)
	}
	if _, _, err := dn.Of("bad_code"); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("displaynames.Of(invalid code) error = %v, want root ErrInvalidCode", err)
	}

	if _, err := numberformat.Decimal("not a number"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("numberformat.Decimal(invalid) error = %v, want root ErrInvalidValue", err)
	} else {
		testcontract.AssertIntlError(t, err, InvalidValue, "numberformat", "decimal", "not a number", "")
		testcontract.AssertErrorExpected(t, err, "a well-formed decimal string, NaN, Infinity, or -Infinity")
	}
}

func TestRootErrorTextTeachesWithoutAbstractOperationNames(t *testing.T) {
	t.Parallel()

	displayNames, err := displaynames.New(intltest.LocaleList(t, "en-US"), displaynames.Options{Type: String(displaynames.Language)})
	if err != nil {
		t.Fatal(err)
	}
	_, _, displayNameErr := displayNames.Of("bad_code")

	relative, err := relativetimeformat.New(intltest.LocaleList(t, "en-US"), relativetimeformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	_, relativeErr := relative.Format(relativetimeformat.Int(1), relativetimeformat.Unit("bad"))

	duration, err := durationformat.New(intltest.LocaleList(t, "en-US"), durationformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	_, durationErr := duration.Format(durationformat.Duration{Hours: 1, Minutes: -1})

	tests := []struct {
		name string
		err  error
	}{
		{name: "numberformat", err: errorFrom(func() (*numberformat.NumberFormat, error) {
			return numberformat.New(intltest.LocaleList(t, "en-US"), numberformat.Options{Style: String("bad")})
		})},
		{name: "locale parse", err: errorFrom(func() (locale.Locale, error) {
			return locale.Parse("bad_locale")
		})},
		{name: "locale options", err: errorFrom(func() (locale.Locale, error) {
			return locale.New("en", locale.Options{HourCycle: String("bad")})
		})},
		{name: "datetimeformat", err: errorFrom(func() (*datetimeformat.DateTimeFormat, error) {
			return datetimeformat.New(intltest.LocaleList(t, "en-US"), datetimeformat.Options{Calendar: String("bad!")})
		})},
		{name: "pluralrules", err: errorFrom(func() (*pluralrules.PluralRules, error) {
			return pluralrules.New(intltest.LocaleList(t, "en-US"), pluralrules.Options{Type: String("bad")})
		})},
		{name: "listformat", err: errorFrom(func() (*listformat.ListFormat, error) {
			return listformat.New(intltest.LocaleList(t, "en-US"), listformat.Options{Type: String("bad")})
		})},
		{name: "collator", err: errorFrom(func() (*collator.Collator, error) {
			return collator.New(intltest.LocaleList(t, "en-US"), collator.Options{Usage: String(collator.SearchUsage)})
		})},
		{name: "segmenter", err: errorFrom(func() (*segmenter.Segmenter, error) {
			return segmenter.New(intltest.LocaleList(t, "en-US"), segmenter.Options{Granularity: String("bad")})
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

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedUnits", SupportedUnits)
	testcontract.AssertStringSliceSortedUnique(t, "SupportedUnits", SupportedUnits())
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
	testcontract.AssertStringSliceContainsAll(t, "SupportedCollations()", got, "phonebk")
	testcontract.AssertStringSliceSortedUnique(t, "SupportedCollations()", got)
	for _, forbidden := range []string{"default", "search", "standard"} {
		if slices.Contains(got, forbidden) {
			t.Fatalf("SupportedCollations() contains reserved value %q: %v", forbidden, got)
		}
	}

	for _, tc := range []struct {
		name    string
		locales locale.List
		options collator.Options
	}{
		{
			name:    "search usage",
			locales: locale.List{intltest.Locale(t, "en")},
			options: collator.Options{Usage: String(collator.SearchUsage)},
		},
		{
			name:    "case first upper option",
			locales: locale.List{intltest.Locale(t, "en")},
			options: collator.Options{CaseFirst: String(collator.UpperCaseFirst)},
		},
		{
			name:    "case first lower option",
			locales: locale.List{intltest.Locale(t, "en")},
			options: collator.Options{CaseFirst: String(collator.LowerCaseFirst)},
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

	for _, tc := range []struct {
		name          string
		locales       locale.List
		options       collator.Options
		wantCollation string
	}{
		{
			name:          "locale case first upper",
			locales:       locale.List{intltest.Locale(t, "en-u-kf-upper")},
			options:       collator.Options{},
			wantCollation: "default",
		},
		{
			name:          "locale case first lower",
			locales:       locale.List{intltest.Locale(t, "en-u-kf-lower")},
			options:       collator.Options{},
			wantCollation: "default",
		},
		{
			name:          "phonebook option",
			locales:       locale.List{intltest.Locale(t, "de")},
			options:       collator.Options{Collation: String("phonebk")},
			wantCollation: "phonebk",
		},
		{
			name:          "phonebook locale",
			locales:       locale.List{intltest.Locale(t, "de-u-co-phonebk")},
			options:       collator.Options{},
			wantCollation: "phonebk",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := collator.New(tc.locales, tc.options)
			if err != nil {
				t.Fatalf("collator.New() error = %v", err)
			}
			resolved := c.ResolvedOptions()
			if resolved.CaseFirst != collator.FalseCaseFirst || resolved.Collation != tc.wantCollation {
				t.Fatalf("collator.New() resolved caseFirst/collation = %q/%q, want false/%s", resolved.CaseFirst, resolved.Collation, tc.wantCollation)
			}
		})
	}

	supported, err := collator.SupportedLocalesOf(intltest.LocaleList(t, "de-u-co-phonebk", "en-u-kf-upper", "en-u-kf-lower", "de", "en-u-kf-false"), collator.Options{})
	if err != nil {
		t.Fatalf("collator.SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "collator.SupportedLocalesOf()", supported, []string{"de-u-co-phonebk", "en-u-kf-upper", "en-u-kf-lower", "de", "en-u-kf-false"})
}

func TestRootConstructorAliases(t *testing.T) {
	t.Parallel()

	locales := intltest.LocaleList(t, "en-US")
	requireRootLocaleAlias(locales[0])
	requireRootNumberFormatAlias((*numberformat.NumberFormat)(nil))
	requireRootDateTimeFormatAlias((*datetimeformat.DateTimeFormat)(nil))
	requireRootPluralRulesAlias((*pluralrules.PluralRules)(nil))
	requireRootDisplayNamesAlias((*displaynames.DisplayNames)(nil))
	requireRootCollatorAlias((*collator.Collator)(nil))
	requireRootSegmenterAlias((*segmenter.Segmenter)(nil))

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
	got, err := relative.Format(relativetimeformat.Int(-1), relativetimeformat.Second)
	if err != nil {
		t.Fatalf("root RelativeTimeFormat alias Format() error = %v", err)
	}
	if got != "1 second ago" {
		t.Fatalf("root RelativeTimeFormat alias Format() = %q, want %q", got, "1 second ago")
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

func requireRootLocaleAlias(Locale) {}

func requireRootNumberFormatAlias(*NumberFormat) {}

func requireRootDateTimeFormatAlias(*DateTimeFormat) {}

func requireRootPluralRulesAlias(*PluralRules) {}

func requireRootListFormatAlias(*ListFormat) {}

func requireRootRelativeTimeFormatAlias(*RelativeTimeFormat) {}

func requireRootDurationFormatAlias(*DurationFormat) {}

func requireRootDisplayNamesAlias(*DisplayNames) {}

func requireRootCollatorAlias(*Collator) {}

func requireRootSegmenterAlias(*Segmenter) {}
