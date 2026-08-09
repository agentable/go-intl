package gointl

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/durationformat"
	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrdisplaynames "github.com/agentable/go-intl/internal/cldr/displaynames"
	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrplural "github.com/agentable/go-intl/internal/cldr/plural"
	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	cldrunit "github.com/agentable/go-intl/internal/cldr/unit"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/tz"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
)

func TestSupportedLocalesOfMatchesCapabilityAccessors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		supported []string
		filter    func(locale.List) (locale.List, error)
	}{
		{
			name:      "NumberFormat",
			supported: cldrnumber.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return numberformat.SupportedLocalesOf(requested, numberformat.Options{LocaleMatcher: String(numberformat.LookupLocaleMatcher)})
			},
		},
		{
			name:      "DateTimeFormat",
			supported: cldrdate.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return datetimeformat.SupportedLocalesOf(requested, datetimeformat.Options{LocaleMatcher: String(datetimeformat.LookupLocaleMatcher)})
			},
		},
		{
			name:      "PluralRules",
			supported: cldrplural.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return pluralrules.SupportedLocalesOf(requested, pluralrules.Options{LocaleMatcher: String(pluralrules.LookupLocaleMatcher)})
			},
		},
		{
			name:      "ListFormat",
			supported: cldrlist.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return listformat.SupportedLocalesOf(requested, listformat.Options{LocaleMatcher: String(listformat.LookupLocaleMatcher)})
			},
		},
		{
			name:      "RelativeTimeFormat",
			supported: relativeTimeSupportedLocaleContract(),
			filter: func(requested locale.List) (locale.List, error) {
				return relativetimeformat.SupportedLocalesOf(requested, relativetimeformat.Options{LocaleMatcher: String(relativetimeformat.LookupLocaleMatcher)})
			},
		},
		{
			name:      "DurationFormat",
			supported: durationSupportedLocaleContract(),
			filter: func(requested locale.List) (locale.List, error) {
				return durationformat.SupportedLocalesOf(requested, durationformat.Options{LocaleMatcher: String(durationformat.LookupLocaleMatcher)})
			},
		},
		{
			name:      "DisplayNames",
			supported: cldrdisplaynames.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return displaynames.SupportedLocalesOf(requested, displaynames.Options{LocaleMatcher: String(displaynames.LookupLocaleMatcher)})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requested := localeListFromSupportedTags(t, tt.supported)
			got, err := tt.filter(requested)
			if err != nil {
				t.Fatalf("SupportedLocalesOf(all supported locales) error = %v", err)
			}
			want := requested.Strings()
			testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf(all supported locales)", got, want)
		})
	}
}

func TestSupportedLocalesConstructFormatters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		supported []string
		construct func(locale.List) error
	}{
		{
			name:      "NumberFormat",
			supported: cldrnumber.SupportedLocales(),
			construct: func(locales locale.List) error {
				_, err := numberformat.New(locales, numberformat.Options{})
				return err
			},
		},
		{
			name:      "DateTimeFormat",
			supported: cldrdate.SupportedLocales(),
			construct: func(locales locale.List) error {
				_, err := datetimeformat.New(locales, datetimeformat.Options{})
				return err
			},
		},
		{
			name:      "PluralRules",
			supported: cldrplural.SupportedLocales(),
			construct: func(locales locale.List) error {
				_, err := pluralrules.New(locales, pluralrules.Options{})
				return err
			},
		},
		{
			name:      "ListFormat",
			supported: cldrlist.SupportedLocales(),
			construct: func(locales locale.List) error {
				_, err := listformat.New(locales, listformat.Options{})
				return err
			},
		},
		{
			name:      "RelativeTimeFormat",
			supported: relativeTimeSupportedLocaleContract(),
			construct: func(locales locale.List) error {
				_, err := relativetimeformat.New(locales, relativetimeformat.Options{})
				return err
			},
		},
		{
			name:      "DurationFormat",
			supported: durationSupportedLocaleContract(),
			construct: func(locales locale.List) error {
				_, err := durationformat.New(locales, durationformat.Options{})
				return err
			},
		},
		{
			name:      "DisplayNames",
			supported: cldrdisplaynames.SupportedLocales(),
			construct: func(locales locale.List) error {
				_, err := displaynames.New(locales, displaynames.Options{Type: String(displaynames.Language)})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, tag := range tt.supported {
				t.Run(tag, func(t *testing.T) {
					t.Parallel()

					locales := locale.List{intltest.Locale(t, tag)}
					if err := tt.construct(locales); err != nil {
						t.Fatalf("%s.New(%q) error = %v", tt.name, tag, err)
					}
				})
			}
		})
	}
}

func TestSupportedValuesMatchFacadeAccessors(t *testing.T) {
	t.Parallel()

	tests := supportedValueTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.values()
			if !slices.Equal(got, tt.want()) {
				t.Fatalf("%s supported values = %v, want %v", tt.name, got, tt.want())
			}
		})
	}
}

func TestSupportedValuesReturnFreshSlices(t *testing.T) {
	t.Parallel()

	for _, tt := range supportedValueTests() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testcontract.AssertOptionalStringSliceReturnsCopy(t, tt.name, tt.values)
		})
	}
}

func TestSupportedValuesAreSortedUnique(t *testing.T) {
	t.Parallel()

	for _, tt := range supportedValueTests() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			values := tt.values()
			testcontract.AssertStringSliceSortedUnique(t, tt.name, values)
		})
	}
}

func TestSupportedTimeZonesHaveCanonicalLoadableLocations(t *testing.T) {
	t.Parallel()

	for _, name := range SupportedTimeZones() {
		t.Run(strings.ReplaceAll(name, "/", "_"), func(t *testing.T) {
			t.Parallel()

			if canonical := tz.CanonicalLink(name); canonical != name {
				t.Fatalf("SupportedTimeZones contains non-canonical zone %q; canonical link is %q", name, canonical)
			}
			if _, err := tz.Resolve(name); err != nil {
				t.Fatalf("SupportedTimeZones contains unloadable zone %q: %v", name, err)
			}
		})
	}
}

func TestSupportedValuesAreBackedByConstructors(t *testing.T) {
	t.Parallel()

	en := locale.List{intltest.Locale(t, "en-US")}

	for _, calendar := range SupportedCalendars() {
		t.Run("calendar/"+calendar, func(t *testing.T) {
			t.Parallel()

			_, err := datetimeformat.New(en, datetimeformat.Options{Calendar: String(calendar)})
			if err != nil {
				t.Fatalf("DateTimeFormat calendar %q error = %v", calendar, err)
			}
		})
	}

	for _, currency := range SupportedCurrencies() {
		t.Run("currency/"+currency, func(t *testing.T) {
			t.Parallel()

			_, err := numberformat.New(en, numberformat.Options{
				Style:    String(numberformat.CurrencyStyle),
				Currency: String(currency),
			})
			if err != nil {
				t.Fatalf("NumberFormat currency %q error = %v", currency, err)
			}
		})
	}

	for _, numberingSystem := range SupportedNumberingSystems() {
		t.Run("numberingSystem/"+numberingSystem, func(t *testing.T) {
			t.Parallel()

			_, err := numberformat.New(en, numberformat.Options{NumberingSystem: String(numberingSystem)})
			if err != nil {
				t.Fatalf("NumberFormat numberingSystem %q error = %v", numberingSystem, err)
			}
		})
	}

	for _, timeZone := range SupportedTimeZones() {
		t.Run("timeZone/"+strings.ReplaceAll(timeZone, "/", "_"), func(t *testing.T) {
			t.Parallel()

			_, err := datetimeformat.New(en, datetimeformat.Options{TimeZone: String(timeZone)})
			if err != nil {
				t.Fatalf("DateTimeFormat timeZone %q error = %v", timeZone, err)
			}
		})
	}

	for _, unit := range SupportedUnits() {
		t.Run("unit/"+unit, func(t *testing.T) {
			t.Parallel()

			_, err := numberformat.New(en, numberformat.Options{
				Style: String(numberformat.UnitStyle),
				Unit:  String(unit),
			})
			if err != nil {
				t.Fatalf("NumberFormat unit %q error = %v", unit, err)
			}
		})
	}
}

type supportedValueTest struct {
	name   string
	values func() []string
	want   func() []string
}

func supportedValueTests() []supportedValueTest {
	return []supportedValueTest{
		{name: "calendar", values: SupportedCalendars, want: cldrdate.SupportedCalendars},
		{name: "currency", values: SupportedCurrencies, want: cldrcurrency.SupportedCodes},
		{name: "numberingSystem", values: SupportedNumberingSystems, want: cldrnumber.SupportedNumberingSystems},
		{name: "timeZone", values: SupportedTimeZones, want: tz.SupportedTimeZones},
		{name: "unit", values: SupportedUnits, want: ecma402.SanctionedSimpleUnitIdentifiers},
	}
}

func TestNodeSupportedValuesWitnessIsReferenceOnly(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/native/node-v26/supported-values.json")
	if err != nil {
		t.Fatal(err)
	}
	var witness struct {
		Source   string              `json:"source"`
		Versions map[string]string   `json:"versions"`
		Values   map[string][]string `json:"values"`
	}
	if err := json.Unmarshal(data, &witness); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(witness.Source, "node:v26.") {
		t.Fatalf("native supported-values source = %q, want node v26 witness", witness.Source)
	}
	for _, key := range []string{"node", "v8", "icu", "cldr", "tz"} {
		if witness.Versions[key] == "" {
			t.Fatalf("native supported-values witness missing %q version", key)
		}
	}
	if !slices.Contains(witness.Values["calendar"], "buddhist") {
		t.Fatal("native supported-values witness must record Node's broader calendar support")
	}
	if slices.Contains(SupportedCalendars(), "buddhist") {
		t.Fatal("go-intl must not mirror Node calendar values until DateTimeFormat can format them truthfully")
	}
}

func localeListFromSupportedTags(t *testing.T, tags []string) locale.List {
	t.Helper()

	locales, err := locale.ParseList(tags...)
	if err != nil {
		t.Fatalf("parse supported locale tags: %v", err)
	}
	return locales
}

func relativeTimeSupportedLocaleContract() []string {
	return cldrlocale.IntersectSupportedLocales(
		cldrrelativetime.SupportedLocales(),
		cldrnumber.SupportedLocales(),
		cldrplural.SupportedLocales(),
	)
}

func durationSupportedLocaleContract() []string {
	return cldrlocale.IntersectSupportedLocales(
		cldrnumber.SupportedLocales(),
		cldrlist.SupportedLocales(),
		cldrplural.SupportedLocales(),
		cldrunit.SupportedLocales(),
	)
}
