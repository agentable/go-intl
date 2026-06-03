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
	"github.com/agentable/go-intl/internal/cldr"
	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrdisplaynames "github.com/agentable/go-intl/internal/cldr/displaynames"
	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrplural "github.com/agentable/go-intl/internal/cldr/plural"
	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	internalcollation "github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
	internalsegmentation "github.com/agentable/go-intl/internal/segmentation"
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

type supportedValueTest struct {
	name   string
	values func() []string
	want   func() []string
}

func supportedValueTests() []supportedValueTest {
	return []supportedValueTest{
		{name: "calendar", values: SupportedCalendars, want: cldr.SupportedCalendars},
		{name: "collation", values: SupportedCollations, want: internalcollation.SupportedCollations},
		{name: "currency", values: SupportedCurrencies, want: cldr.SupportedCurrencies},
		{name: "numberingSystem", values: SupportedNumberingSystems, want: cldr.SupportedNumberingSystems},
		{name: "timeZone", values: SupportedTimeZones, want: cldr.SupportedTimeZones},
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
	units := cldr.UnitSupportedLocales()

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
