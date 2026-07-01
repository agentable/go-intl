package collator

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		if fixture.IsSupportedLocalesOf() {
			runSupportedLocalesFixture(t, fixture)
			return
		}

		format, err := New(locale.List{intltest.Locale(t, fixture.Locale)}, conformanceCollatorOptions(t, fixture))
		if testcontract.AssertErrorCode(t, "New()", err, fixture.ErrorCode, func(code string) error {
			return conformanceCollatorError(t, code)
		}) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(fixture.ExpectedResolved) != 0 {
			assertCollatorResolvedOptions(t, fixture, format.ResolvedOptions())
		}
		var input struct {
			Left  string `json:"left"`
			Right string `json:"right"`
		}
		if err := json.Unmarshal(fixture.Input, &input); err != nil {
			t.Fatal(err)
		}
		if fixture.ExpectedComparison == nil {
			t.Fatal("fixture expectedComparison is required")
		}
		if got := compareSign(format.Compare(input.Left, input.Right)); got != *fixture.ExpectedComparison {
			t.Fatalf("Compare(%q, %q) sign = %d, want %d", input.Left, input.Right, got, *fixture.ExpectedComparison)
		}
	})
}

func TestConformanceCollatorOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceCollatorOptions(t, conformance.Fixture{
		Options: json.RawMessage(`{"usage":""}`),
	}))
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want %v", err, intlerr.ErrInvalidOption)
	}
	testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "usage", "", "en")
	testcontract.AssertOptionExpected(t, err, `one of "sort", "search"`)
}

func runSupportedLocalesFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	testcontract.AssertSupportedLocalesOfFixture(t, fixture, intltest.LocaleListJSON, func(locales locale.List) (locale.List, error) {
		return SupportedLocalesOf(locales, conformanceCollatorOptions(t, fixture))
	}, func(code string) error {
		return conformanceCollatorError(t, code)
	})
}

func assertCollatorResolvedOptions(t *testing.T, fixture conformance.Fixture, got ResolvedOptions) {
	t.Helper()

	want := testcontract.ExpectedResolvedOptions(t, fixture)
	testcontract.AssertResolvedString(t, want, "locale", got.Locale.String())
	testcontract.AssertResolvedString(t, want, "usage", string(got.Usage))
	testcontract.AssertResolvedString(t, want, "sensitivity", string(got.Sensitivity))
	testcontract.AssertResolvedBool(t, want, "ignorePunctuation", got.IgnorePunctuation)
	testcontract.AssertResolvedString(t, want, "collation", got.Collation)
	testcontract.AssertResolvedBool(t, want, "numeric", got.Numeric)
	testcontract.AssertResolvedString(t, want, "caseFirst", string(got.CaseFirst))
}

func conformanceCollatorOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher     *string `json:"localeMatcher"`
		Usage             *string `json:"usage"`
		Sensitivity       *string `json:"sensitivity"`
		CaseFirst         *string `json:"caseFirst"`
		Numeric           *bool   `json:"numeric"`
		IgnorePunctuation *bool   `json:"ignorePunctuation"`
		Collation         *string `json:"collation"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	return Options{
		LocaleMatcher:     options.LocaleMatcher,
		Usage:             options.Usage,
		Sensitivity:       options.Sensitivity,
		CaseFirst:         options.CaseFirst,
		Numeric:           options.Numeric,
		IgnorePunctuation: options.IgnorePunctuation,
		Collation:         options.Collation,
	}
}

func conformanceCollatorError(t *testing.T, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "collator", code, "invalidOption", "unsupportedOption")
}

func compareSign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
