package collator

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		format, err := New(locale.List{locale.MustParse(fixture.Locale)}, conformanceCollatorOptions(t, fixture))
		if fixture.ErrorCode != "" {
			if !errors.Is(err, conformanceCollatorError(t, fixture.ErrorCode)) {
				t.Fatalf("New() error = %v, want %q", err, fixture.ErrorCode)
			}
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

func assertCollatorResolvedOptions(t *testing.T, fixture conformance.Fixture, got ResolvedOptions) {
	t.Helper()

	var want struct {
		Locale            *string      `json:"locale"`
		Usage             *Usage       `json:"usage"`
		Sensitivity       *Sensitivity `json:"sensitivity"`
		IgnorePunctuation *bool        `json:"ignorePunctuation"`
		Collation         *string      `json:"collation"`
		Numeric           *bool        `json:"numeric"`
		CaseFirst         *CaseFirst   `json:"caseFirst"`
	}
	if err := json.Unmarshal(fixture.ExpectedResolved, &want); err != nil {
		t.Fatal(err)
	}
	if want.Locale != nil && got.Locale.String() != *want.Locale {
		t.Fatalf("ResolvedOptions().Locale = %q, want %q", got.Locale.String(), *want.Locale)
	}
	if want.Usage != nil && got.Usage != *want.Usage {
		t.Fatalf("ResolvedOptions().Usage = %q, want %q", got.Usage, *want.Usage)
	}
	if want.Sensitivity != nil && got.Sensitivity != *want.Sensitivity {
		t.Fatalf("ResolvedOptions().Sensitivity = %q, want %q", got.Sensitivity, *want.Sensitivity)
	}
	if want.IgnorePunctuation != nil && got.IgnorePunctuation != *want.IgnorePunctuation {
		t.Fatalf("ResolvedOptions().IgnorePunctuation = %t, want %t", got.IgnorePunctuation, *want.IgnorePunctuation)
	}
	if want.Collation != nil && got.Collation != *want.Collation {
		t.Fatalf("ResolvedOptions().Collation = %q, want %q", got.Collation, *want.Collation)
	}
	if want.Numeric != nil && got.Numeric != *want.Numeric {
		t.Fatalf("ResolvedOptions().Numeric = %t, want %t", got.Numeric, *want.Numeric)
	}
	if want.CaseFirst != nil && got.CaseFirst != *want.CaseFirst {
		t.Fatalf("ResolvedOptions().CaseFirst = %q, want %q", got.CaseFirst, *want.CaseFirst)
	}
}

func conformanceCollatorOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher     string `json:"localeMatcher"`
		Usage             string `json:"usage"`
		Sensitivity       string `json:"sensitivity"`
		CaseFirst         string `json:"caseFirst"`
		Numeric           *bool  `json:"numeric"`
		IgnorePunctuation *bool  `json:"ignorePunctuation"`
		Collation         string `json:"collation"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	var opts Options
	if options.LocaleMatcher != "" {
		opts.LocaleMatcher = LocaleMatcher(options.LocaleMatcher)
	}
	if options.Usage != "" {
		opts.Usage = Usage(options.Usage)
	}
	if options.Sensitivity != "" {
		opts.Sensitivity = Sensitivity(options.Sensitivity)
	}
	if options.CaseFirst != "" {
		opts.CaseFirst = CaseFirst(options.CaseFirst)
	}
	if options.Numeric != nil {
		opts.Numeric = options.Numeric
	}
	if options.IgnorePunctuation != nil {
		opts.IgnorePunctuation = options.IgnorePunctuation
	}
	if options.Collation != "" {
		opts.Collation = options.Collation
	}
	return opts
}

func conformanceCollatorError(t *testing.T, code string) error {
	t.Helper()

	switch code {
	case "invalidOption":
		return intlerr.ErrInvalidOption
	case "unsupportedOption":
		return intlerr.ErrUnsupportedOption
	default:
		t.Fatalf("unsupported collator error code %q", code)
		return nil
	}
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
