package localematcher_test

import (
	"errors"
	"testing"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/x/localematcher"
)

func locales(t testing.TB, tags ...string) locale.List {
	t.Helper()
	values, err := locale.ParseList(tags...)
	if err != nil {
		t.Fatal(err)
	}
	return values
}

func TestMatcherLookupRequiresAvailableResource(t *testing.T) {
	t.Parallel()
	matcher, err := localematcher.New(locales(t, "de-DE"), localematcher.Options{Algorithm: gointl.String("lookup")})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := matcher.Match(locales(t, "de-AT")); ok {
		t.Fatalf("lookup returned unavailable resource %q", got.String())
	}
	if got, ok := matcher.Match(locales(t, "de-DE")); !ok || got.String() != "de-DE" {
		t.Fatalf("exact lookup = (%s, %v), want (de-DE, true)", got.String(), ok)
	}
}

func TestMatcherLookupContinuesAfterUnavailableParent(t *testing.T) {
	t.Parallel()
	matcher, err := localematcher.New(locales(t, "de-DE", "en"), localematcher.Options{Algorithm: gointl.String("lookup")})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := matcher.Match(locales(t, "de-AT", "en"))
	if !ok || got.String() != "en" {
		t.Fatalf("lookup = (%s, %v), want (en, true)", got.String(), ok)
	}
}

func TestNewValidation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		tags      []string
		algorithm *string
		want      error
		field     string
	}{
		{"unknown algorithm", []string{"en"}, gointl.String("invalid"), gointl.ErrInvalidOption, "algorithm"},
		{"explicit empty algorithm", []string{"en"}, gointl.String(""), gointl.ErrInvalidOption, "algorithm"},
		{"algorithm casing", []string{"en"}, gointl.String("LOOKUP"), gointl.ErrInvalidOption, "algorithm"},
		{"available keyword", []string{"en-u-ca-gregory"}, nil, gointl.ErrInvalidValue, "availableLocales"},
		{"available attribute", []string{"en-u-foobar"}, nil, gointl.ErrInvalidValue, "availableLocales"},
		{"available wins evaluation order", []string{"en-u-nu-latn"}, gointl.String("invalid"), gointl.ErrInvalidValue, "availableLocales"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := localematcher.New(locales(t, tc.tags...), localematcher.Options{Algorithm: tc.algorithm})
			if got != nil || !errors.Is(err, tc.want) {
				t.Fatalf("New = (%v, %v), want nil, %v", got, err, tc.want)
			}
			detail, ok := errors.AsType[*gointl.Error](err)
			if !ok || detail.Owner != "localematcher" || detail.Name != tc.field {
				t.Fatalf("error context = %#v", detail)
			}
		})
	}
}

func TestMatcherMatch(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                 string
		available, requested []string
		algorithm            *string
		want                 string
	}{
		{"proposal example", []string{"fr", "en"}, []string{"fr-XX", "en"}, nil, "fr"},
		{"traditional Chinese", []string{"en", "zh-Hans", "zh-Hant", "ja"}, []string{"zh-HK"}, nil, "zh-Hant"},
		{"simplified Chinese", []string{"en", "zh-Hans", "zh-Hant", "ja"}, []string{"zh-SG"}, nil, "zh-Hans"},
		{"best fit returns actual resource", []string{"zh-Hant"}, []string{"zh-TW"}, nil, "zh-Hant"},
		{"lookup does not add alias", []string{"zh-Hant-HK"}, []string{"zh-HK"}, gointl.String("lookup"), ""},
		{"lookup does not add parent", []string{"zh-Hans"}, []string{"zh-SG"}, gointl.String("lookup"), ""},
		{"lookup actual parent", []string{"de"}, []string{"de-AT"}, gointl.String("lookup"), "de"},
		{"lookup strips unicode extension", []string{"en-US"}, []string{"en-US-u-ca-buddhist"}, gointl.String("lookup"), "en-US"},
		{"best fit strips unicode extension", []string{"en-US"}, []string{"en-US-u-nu-thai"}, nil, "en-US"},
		{"canonical case", []string{"EN-us"}, []string{"en-us"}, gointl.String("lookup"), "en-US"},
		{"canonical aliases", []string{"he"}, []string{"iw"}, gointl.String("lookup"), "he"},
		{"duplicate available", []string{"en-US", "EN-us"}, []string{"en-US"}, nil, "en-US"},
		{"empty available", nil, []string{"en"}, nil, ""},
		{"empty requested", []string{"en"}, nil, nil, ""},
		{"distant language", []string{"en", "zh-Hant", "ja"}, []string{"ko"}, nil, ""},
		{"well formed unknown locale", []string{"qaa-Qaaa-ZZ"}, []string{"qaa-Qaaa-ZZ"}, nil, "qaa-Qaaa-ZZ"},
		{"private use containing u", []string{"en-x-u-ca-gregory"}, []string{"en-x-u-ca-gregory"}, gointl.String("lookup"), "en-x-u-ca-gregory"},
		{"und differs from no match", []string{"und"}, []string{"und"}, gointl.String("lookup"), "und"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			matcher, err := localematcher.New(locales(t, tc.available...), localematcher.Options{Algorithm: tc.algorithm})
			if err != nil {
				t.Fatal(err)
			}
			got, ok := matcher.Match(locales(t, tc.requested...))
			if ok != (tc.want != "") || ok && got.String() != tc.want {
				t.Fatalf("Match = (%s, %v), want (%s, %v)", got.String(), ok, tc.want, tc.want != "")
			}
		})
	}
}

func TestMatcherSnapshotsInputs(t *testing.T) {
	t.Parallel()
	available := locales(t, "de-DE")
	algorithm := "lookup"
	matcher, err := localematcher.New(available, localematcher.Options{Algorithm: &algorithm})
	if err != nil {
		t.Fatal(err)
	}
	available[0] = locales(t, "en")[0]
	algorithm = "best fit"
	if got, ok := matcher.Match(locales(t, "de-AT")); ok {
		t.Fatalf("changed algorithm: %s", got.String())
	}
	if got, ok := matcher.Match(locales(t, "de-DE")); !ok || got.String() != "de-DE" {
		t.Fatalf("changed available: (%s,%v)", got.String(), ok)
	}
}

func TestMatcherPerKeyCatalog(t *testing.T) {
	t.Parallel()
	requested := locales(t, "zh-SG")
	for _, tc := range []struct {
		name         string
		translations map[string]string
		want         string
	}{
		{"title", map[string]string{"en": "Title", "zh-Hans": "标题", "zh-Hant": "標題"}, "标题"},
		{"summary", map[string]string{"en": "Summary", "zh-Hant": "摘要"}, "摘要"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			available := make(locale.List, 0, len(tc.translations))
			for tag := range tc.translations {
				available = append(available, locales(t, tag)[0])
			}
			matcher, err := localematcher.New(available, localematcher.Options{})
			if err != nil {
				t.Fatal(err)
			}
			matched, ok := matcher.Match(requested)
			if !ok {
				t.Fatal("no match")
			}
			if got := tc.translations[matched.String()]; got != tc.want {
				t.Fatalf("translation = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMatcherConcurrentRequests(t *testing.T) {
	t.Parallel()
	matcher, err := localematcher.New(locales(t, "en", "zh-Hans", "zh-Hant"), localematcher.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for range 16 {
		for _, tc := range []struct{ requested, want string }{{"zh-HK", "zh-Hant"}, {"zh-SG", "zh-Hans"}, {"ko", ""}} {
			t.Run(tc.requested, func(t *testing.T) {
				t.Parallel()
				for range 20 {
					got, ok := matcher.Match(locales(t, tc.requested))
					if ok != (tc.want != "") || ok && got.String() != tc.want {
						t.Fatalf("Match = (%s,%v), want %s", got.String(), ok, tc.want)
					}
				}
			})
		}
	}
}

func TestZeroMatcher(t *testing.T) {
	t.Parallel()
	var matcher localematcher.Matcher
	if got, ok := matcher.Match(locales(t, "en")); ok {
		t.Fatalf("zero matcher = %s", got.String())
	}
}
