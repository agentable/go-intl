package gointl

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/durationformat"
	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrdisplaynames "github.com/agentable/go-intl/internal/cldr/displaynames"
	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrplural "github.com/agentable/go-intl/internal/cldr/plural"
	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	cldrtimezone "github.com/agentable/go-intl/internal/cldr/timezone"
	cldrunit "github.com/agentable/go-intl/internal/cldr/unit"
	internalcollation "github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intltest"
	internalsegmentation "github.com/agentable/go-intl/internal/segmentation"
	"github.com/agentable/go-intl/internal/tz"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
	"github.com/agentable/go-intl/segmenter"
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
				return numberformat.SupportedLocalesOf(requested, numberformat.Options{LocaleMatcher: numberformat.LookupLocaleMatcher})
			},
		},
		{
			name:      "DateTimeFormat",
			supported: cldrdate.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return datetimeformat.SupportedLocalesOf(requested, datetimeformat.Options{LocaleMatcher: datetimeformat.LookupLocaleMatcher})
			},
		},
		{
			name:      "PluralRules",
			supported: cldrplural.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return pluralrules.SupportedLocalesOf(requested, pluralrules.Options{LocaleMatcher: pluralrules.LookupLocaleMatcher})
			},
		},
		{
			name:      "ListFormat",
			supported: cldrlist.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return listformat.SupportedLocalesOf(requested, listformat.Options{LocaleMatcher: listformat.LookupLocaleMatcher})
			},
		},
		{
			name:      "RelativeTimeFormat",
			supported: relativeTimeSupportedLocaleContract(),
			filter: func(requested locale.List) (locale.List, error) {
				return relativetimeformat.SupportedLocalesOf(requested, relativetimeformat.Options{LocaleMatcher: relativetimeformat.LookupLocaleMatcher})
			},
		},
		{
			name:      "DurationFormat",
			supported: durationSupportedLocaleContract(),
			filter: func(requested locale.List) (locale.List, error) {
				return durationformat.SupportedLocalesOf(requested, durationformat.Options{LocaleMatcher: durationformat.LookupLocaleMatcher})
			},
		},
		{
			name:      "DisplayNames",
			supported: cldrdisplaynames.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return displaynames.SupportedLocalesOf(requested, displaynames.Options{LocaleMatcher: displaynames.LookupLocaleMatcher})
			},
		},
		{
			name:      "Collator",
			supported: internalcollation.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return collator.SupportedLocalesOf(requested, collator.Options{LocaleMatcher: collator.LookupLocaleMatcher})
			},
		},
		{
			name:      "Segmenter",
			supported: internalsegmentation.SupportedLocales(),
			filter: func(requested locale.List) (locale.List, error) {
				return segmenter.SupportedLocalesOf(requested, segmenter.Options{LocaleMatcher: segmenter.LookupLocaleMatcher})
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
			if !slices.Equal(got.Strings(), want) {
				t.Fatalf("SupportedLocalesOf(all supported locales) = %v, want %v", got.Strings(), want)
			}
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
				_, err := displaynames.New(locales, displaynames.Options{Type: displaynames.Language})
				return err
			},
		},
		{
			name:      "Collator",
			supported: internalcollation.SupportedLocales(),
			construct: func(locales locale.List) error {
				_, err := collator.New(locales, collator.Options{})
				return err
			},
		},
		{
			name:      "Segmenter",
			supported: internalsegmentation.SupportedLocales(),
			construct: func(locales locale.List) error {
				_, err := segmenter.New(locales, segmenter.Options{})
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

			got := tt.values()
			if len(got) == 0 {
				return
			}
			original := tt.want()
			got[0] = "mutated"
			again := tt.values()
			if !slices.Equal(again, original) {
				t.Fatalf("%s supported values share mutable backing storage: got %v, want %v", tt.name, again, original)
			}
		})
	}
}

func TestSupportedValuesAreSortedUnique(t *testing.T) {
	t.Parallel()

	for _, tt := range supportedValueTests() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			values := tt.values()
			if !slices.IsSorted(values) {
				t.Fatalf("%s supported values = %v, want sorted", tt.name, values)
			}
			for i := 1; i < len(values); i++ {
				if values[i] == values[i-1] {
					t.Fatalf("%s supported values contain duplicate %q: %v", tt.name, values[i], values)
				}
			}
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

			_, err := datetimeformat.New(en, datetimeformat.Options{Calendar: calendar})
			if err != nil {
				t.Fatalf("DateTimeFormat calendar %q error = %v", calendar, err)
			}
		})
	}

	for _, collation := range SupportedCollations() {
		t.Run("collation/"+collation, func(t *testing.T) {
			t.Parallel()

			_, err := collator.New(en, collator.Options{Collation: collation})
			if err != nil {
				t.Fatalf("Collator collation %q error = %v", collation, err)
			}
		})
	}

	for _, currency := range SupportedCurrencies() {
		t.Run("currency/"+currency, func(t *testing.T) {
			t.Parallel()

			_, err := numberformat.New(en, numberformat.Options{
				Style:    numberformat.CurrencyStyle,
				Currency: numberformat.Currency(currency),
			})
			if err != nil {
				t.Fatalf("NumberFormat currency %q error = %v", currency, err)
			}
		})
	}

	for _, numberingSystem := range SupportedNumberingSystems() {
		t.Run("numberingSystem/"+numberingSystem, func(t *testing.T) {
			t.Parallel()

			_, err := numberformat.New(en, numberformat.Options{NumberingSystem: numberingSystem})
			if err != nil {
				t.Fatalf("NumberFormat numberingSystem %q error = %v", numberingSystem, err)
			}
		})
	}

	for _, timeZone := range SupportedTimeZones() {
		t.Run("timeZone/"+strings.ReplaceAll(timeZone, "/", "_"), func(t *testing.T) {
			t.Parallel()

			_, err := datetimeformat.New(en, datetimeformat.Options{TimeZone: timeZone})
			if err != nil {
				t.Fatalf("DateTimeFormat timeZone %q error = %v", timeZone, err)
			}
		})
	}

	for _, unit := range SupportedUnits() {
		t.Run("unit/"+unit, func(t *testing.T) {
			t.Parallel()

			_, err := numberformat.New(en, numberformat.Options{
				Style: numberformat.UnitStyle,
				Unit:  numberformat.Unit(unit),
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
		{name: "collation", values: SupportedCollations, want: internalcollation.SupportedCollations},
		{name: "currency", values: SupportedCurrencies, want: cldrcurrency.SupportedCodes},
		{name: "numberingSystem", values: SupportedNumberingSystems, want: cldrnumber.SupportedNumberingSystems},
		{name: "timeZone", values: SupportedTimeZones, want: cldrtimezone.SupportedTimeZones},
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
	if !slices.Contains(witness.Values["collation"], "emoji") {
		t.Fatal("native supported-values witness must record Node's broader collation support")
	}
	if slices.Contains(SupportedCollations(), "emoji") {
		t.Fatal("go-intl must not mirror Node collation values until Collator can apply them truthfully")
	}
}

func TestConstructorSupportedLocalesAvoidHandWrittenAllowLists(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"numberformat/supported.go",
		"datetimeformat/supported.go",
		"pluralrules/supported.go",
		"listformat/supported.go",
		"relativetimeformat/supported.go",
		"durationformat/supported.go",
		"displaynames/supported.go",
		"collator/supported.go",
		"segmenter/supported.go",
	} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			hasLiteral, err := hasStringSliceLiteral(path)
			if err != nil {
				t.Fatal(err)
			}
			if hasLiteral {
				t.Fatalf("%s must use generated or engine capability accessors, not a hand-written []string allow-list", path)
			}
		})
	}
}

func TestConstructorSupportedLocalesUseSharedECMA402Operation(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"numberformat/supported.go",
		"datetimeformat/supported.go",
		"pluralrules/supported.go",
		"listformat/supported.go",
		"relativetimeformat/supported.go",
		"durationformat/supported.go",
		"displaynames/supported.go",
		"collator/supported.go",
		"segmenter/supported.go",
	} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			calls, err := fileCallsImportedSelector(path, "github.com/agentable/go-intl/internal/ecma402", "SupportedLocalesOf")
			if err != nil {
				t.Fatal(err)
			}
			if !calls {
				t.Fatalf("%s must call internal/ecma402.SupportedLocalesOf instead of duplicating locale filtering", path)
			}
		})
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
	numbers := cldrnumber.SupportedLocales()
	plurals := cldrplural.SupportedLocales()

	out := slices.Clone(cldrrelativetime.SupportedLocales())
	return slices.DeleteFunc(out, func(loc string) bool {
		return !slices.Contains(numbers, loc) || !slices.Contains(plurals, loc)
	})
}

func durationSupportedLocaleContract() []string {
	lists := cldrlist.SupportedLocales()
	plurals := cldrplural.SupportedLocales()
	units := cldrunit.SupportedLocales()

	out := slices.Clone(cldrnumber.SupportedLocales())
	return slices.DeleteFunc(out, func(loc string) bool {
		return !slices.Contains(lists, loc) ||
			!slices.Contains(plurals, loc) ||
			!slices.Contains(units, loc)
	})
}

func hasStringSliceLiteral(path string) (bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return false, err
	}

	hasLiteral := false
	ast.Inspect(file, func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		array, ok := lit.Type.(*ast.ArrayType)
		if !ok {
			return true
		}
		elt, ok := array.Elt.(*ast.Ident)
		if ok && elt.Name == "string" {
			hasLiteral = true
			return false
		}
		return true
	})
	return hasLiteral, nil
}

func fileCallsImportedSelector(path, importPath, selectorName string) (bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return false, err
	}
	imports := map[string]string{}
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		name := path[strings.LastIndex(path, "/")+1:]
		if spec.Name != nil {
			name = spec.Name.Name
		}
		imports[name] = path
	}

	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != selectorName {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if ok && imports[ident.Name] == importPath {
			found = true
			return false
		}
		return true
	})
	return found, nil
}
