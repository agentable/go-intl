package collator_test

import (
	"errors"
	"sync"
	"testing"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func TestCollator_Compare_Basic(t *testing.T) {
	t.Parallel()
	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Compare("a", "b"); got >= 0 {
		t.Errorf("Compare(a,b) = %d, want < 0", got)
	}
	if got := c.Compare("b", "a"); got <= 0 {
		t.Errorf("Compare(b,a) = %d, want > 0", got)
	}
	if got := c.Compare("a", "a"); got != 0 {
		t.Errorf("Compare(a,a) = %d, want 0", got)
	}
}

func TestCollator_Sensitivity_Base(t *testing.T) {
	t.Parallel()
	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{Sensitivity: gointl.String(collator.BaseSensitivity)})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Compare("a", "A"); got != 0 {
		t.Errorf("base Compare(a,A) = %d, want 0", got)
	}
	if got := c.Compare("a", "á"); got != 0 {
		t.Errorf("base Compare(a,á) = %d, want 0", got)
	}
}

func TestCollator_Sensitivity_Variant(t *testing.T) {
	t.Parallel()
	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{Sensitivity: gointl.String(collator.VariantSensitivity)})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Compare("a", "A"); got == 0 {
		t.Errorf("variant Compare(a,A) = 0, want non-zero")
	}
}

func TestCollator_Numeric(t *testing.T) {
	t.Parallel()
	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{Numeric: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Compare("2", "10"); got >= 0 {
		t.Errorf("numeric Compare(2,10) = %d, want < 0", got)
	}
}

func TestCollator_Compare_Concurrent(t *testing.T) {
	t.Parallel()

	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{
		Sensitivity: gointl.String(collator.BaseSensitivity),
		Numeric:     gointl.Bool(true),
	})
	if err != nil {
		t.Fatal(err)
	}

	const workers = 16
	errs := make(chan string, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				if got := c.Compare("2", "10"); got >= 0 {
					errs <- "numeric Compare(2,10) returned non-negative"
					return
				}
				if got := c.Compare("a", "A"); got != 0 {
					errs <- "base Compare(a,A) returned non-zero"
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestCollator_CopiesRemainSafeAfterUse(t *testing.T) {
	t.Parallel()

	original, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{
		Sensitivity: gointl.String(collator.BaseSensitivity),
		Numeric:     gointl.Bool(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	copyBeforeUse := *original
	if got := original.Compare("2", "10"); got >= 0 {
		t.Fatalf("initial Compare(2,10) = %d, want < 0", got)
	}

	copyAfterUse := *original
	copyOfCopy := copyAfterUse
	values := []*collator.Collator{original, &copyBeforeUse, &copyAfterUse, &copyOfCopy}

	var wg sync.WaitGroup
	for _, value := range values {
		for range 8 {
			wg.Go(func() {
				for range 100 {
					if got := value.Compare("2", "10"); got >= 0 {
						t.Errorf("numeric Compare(2,10) = %d, want < 0", got)
						return
					}
					if got := value.Compare("a", "A"); got != 0 {
						t.Errorf("base Compare(a,A) = %d, want 0", got)
						return
					}
				}
			})
		}
	}
	wg.Wait()

	for _, value := range values {
		if got := value.ResolvedOptions(); !got.Numeric || got.Sensitivity != collator.BaseSensitivity {
			t.Fatalf("ResolvedOptions() = %#v, want numeric base sensitivity", got)
		}
	}
}

func TestCollator_NumericLocaleExtensionPrecedence(t *testing.T) {
	t.Parallel()

	fromLocale, err := collator.New(locale.List{intltest.Locale(t, "en-u-kn")}, collator.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !fromLocale.ResolvedOptions().Numeric {
		t.Fatal("locale kn ResolvedOptions().Numeric = false, want true")
	}
	if got := fromLocale.Compare("2", "10"); got >= 0 {
		t.Errorf("locale kn Compare(2,10) = %d, want < 0", got)
	}

	overridden, err := collator.New(locale.List{intltest.Locale(t, "en-u-kn")}, collator.Options{Numeric: boolPtr(false)})
	if err != nil {
		t.Fatal(err)
	}
	if overridden.ResolvedOptions().Numeric {
		t.Fatal("explicit false ResolvedOptions().Numeric = true, want false")
	}
	if got := overridden.Compare("2", "10"); got <= 0 {
		t.Errorf("explicit false Compare(2,10) = %d, want > 0", got)
	}
}

func TestCollator_LocaleExtensionsComeFromMatchedLocale(t *testing.T) {
	t.Parallel()

	unsupportedFirst, err := collator.New(locale.List{intltest.Locale(t, "xh-u-kn"), intltest.Locale(t, "en")}, collator.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if unsupportedFirst.ResolvedOptions().Numeric {
		t.Fatal("unsupported first locale Numeric = true, want false from matched en locale")
	}
	if got := unsupportedFirst.Compare("2", "10"); got <= 0 {
		t.Fatalf("unsupported first locale Compare(2,10) = %d, want lexical ordering > 0", got)
	}

	unsupportedCaseFirst, err := collator.New(locale.List{intltest.Locale(t, "xh-u-kf-upper"), intltest.Locale(t, "en")}, collator.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := unsupportedCaseFirst.ResolvedOptions().CaseFirst; got != collator.FalseCaseFirst {
		t.Fatalf("unsupported first locale CaseFirst = %q, want false from matched en locale", got)
	}
}

func TestCollator_ExplicitFalseCaseFirstOverridesUnsupportedLocaleExtension(t *testing.T) {
	t.Parallel()

	c, err := collator.New(locale.List{intltest.Locale(t, "en-u-kf-upper")}, collator.Options{CaseFirst: gointl.String(collator.FalseCaseFirst)})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.ResolvedOptions().CaseFirst; got != collator.FalseCaseFirst {
		t.Fatalf("ResolvedOptions().CaseFirst = %q, want false", got)
	}
	if got := c.ResolvedOptions().Locale.String(); got != "en" {
		t.Fatalf("ResolvedOptions().Locale = %q, want en after explicit caseFirst override", got)
	}
}

func TestCollator_CollationRequestsFollowBackendCapability(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		locales    locale.List
		options    collator.Options
		want       string
		wantLocale string
	}{
		{
			name:    "explicit default",
			locales: locale.List{intltest.Locale(t, "en")},
			options: collator.Options{Collation: gointl.String("default")},
			want:    "default",
		},
		{
			name:    "locale co default",
			locales: locale.List{intltest.Locale(t, "en-u-co-default")},
			options: collator.Options{},
			want:    "default",
		},
		{
			name:    "locale co standard",
			locales: locale.List{intltest.Locale(t, "en-u-co-standard")},
			options: collator.Options{},
			want:    "default",
		},
		{
			name:    "locale co search",
			locales: locale.List{intltest.Locale(t, "en-u-co-search")},
			options: collator.Options{},
			want:    "default",
		},
		{
			name:       "explicit backend-supported",
			locales:    locale.List{intltest.Locale(t, "de")},
			options:    collator.Options{Collation: gointl.String("phonebk")},
			want:       "phonebk",
			wantLocale: "de",
		},
		{
			name:       "explicit backend-supported canonicalized",
			locales:    locale.List{intltest.Locale(t, "de")},
			options:    collator.Options{Collation: gointl.String("PHONEBK")},
			want:       "phonebk",
			wantLocale: "de",
		},
		{
			name:       "locale backend-supported",
			locales:    locale.List{intltest.Locale(t, "de-u-co-phonebk")},
			options:    collator.Options{},
			want:       "phonebk",
			wantLocale: "de-u-co-phonebk",
		},
		{
			name:    "explicit default overrides unsupported locale co",
			locales: locale.List{intltest.Locale(t, "en-u-co-phonebk")},
			options: collator.Options{Collation: gointl.String("default")},
			want:    "default",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := collator.New(tc.locales, tc.options)
			if err != nil {
				t.Fatal(err)
			}
			if got := c.ResolvedOptions().Collation; got != tc.want {
				t.Fatalf("ResolvedOptions().Collation = %q, want %q", got, tc.want)
			}
			if tc.wantLocale != "" {
				if got := c.ResolvedOptions().Locale.String(); got != tc.wantLocale {
					t.Fatalf("ResolvedOptions().Locale = %q, want %q", got, tc.wantLocale)
				}
			}
		})
	}
}

func TestCollator_UnsupportedLocaleCaseFirstFallsBackToDefault(t *testing.T) {
	t.Parallel()

	for _, tag := range []string{"en-u-kf-upper", "en-u-kf-lower"} {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()

			c, err := collator.New(locale.List{intltest.Locale(t, tag)}, collator.Options{})
			if err != nil {
				t.Fatalf("New(%s) error = %v", tag, err)
			}
			if got := c.ResolvedOptions().CaseFirst; got != collator.FalseCaseFirst {
				t.Fatalf("ResolvedOptions().CaseFirst = %q, want false fallback", got)
			}
		})
	}
}

func TestCollator_IgnorePunctuation(t *testing.T) {
	t.Parallel()

	_, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{
		IgnorePunctuation: boolPtr(true),
	})
	if !errors.Is(err, gointl.ErrUnsupportedOption) {
		t.Fatalf("New(ignorePunctuation=true) error = %v, want ErrUnsupportedOption", err)
	}
	detail, ok := errors.AsType[*gointl.Error](err)
	if !ok {
		t.Fatalf("New(ignorePunctuation=true) error = %T, want *gointl.Error", err)
	}
	if detail.Owner != "collator" || detail.Name != "ignorePunctuation" || detail.Value != "true" || detail.Locale != "en" {
		t.Fatalf("New(ignorePunctuation=true) detail = %#v", detail)
	}

	withoutPunctuation, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := withoutPunctuation.Compare("a-b", "ab"); got == 0 {
		t.Errorf("default Compare(a-b,ab) = 0, want non-zero")
	}

	explicitFalse, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{
		IgnorePunctuation: boolPtr(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	if explicitFalse.ResolvedOptions().IgnorePunctuation {
		t.Error("explicit false ResolvedOptions().IgnorePunctuation = true, want false")
	}
	if got := explicitFalse.Compare("a-b", "ab"); got == 0 {
		t.Errorf("explicit false Compare(a-b,ab) = 0, want non-zero")
	}
}

func TestCollator_ResolvedOptions(t *testing.T) {
	t.Parallel()
	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{Sensitivity: gointl.String(collator.AccentSensitivity)})
	if err != nil {
		t.Fatal(err)
	}
	got := c.ResolvedOptions()
	if got.Usage != collator.SortUsage {
		t.Errorf("Usage = %q, want %q", got.Usage, collator.SortUsage)
	}
	if got.Sensitivity != collator.AccentSensitivity {
		t.Errorf("Sensitivity = %q, want %q", got.Sensitivity, collator.AccentSensitivity)
	}
	if got.CaseFirst != collator.FalseCaseFirst {
		t.Errorf("CaseFirst = %q, want %q", got.CaseFirst, collator.FalseCaseFirst)
	}
	if got.Collation != "default" {
		t.Errorf("Collation = %q, want default", got.Collation)
	}
}

func TestCollator_CaseFirstFalseIsAccepted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		locales locale.List
		options collator.Options
	}{
		{
			name:    "explicit option",
			locales: locale.List{intltest.Locale(t, "en")},
			options: collator.Options{CaseFirst: gointl.String(collator.FalseCaseFirst)},
		},
		{
			name:    "locale extension",
			locales: locale.List{intltest.Locale(t, "en-u-kf-false")},
			options: collator.Options{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := collator.New(tc.locales, tc.options)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if got := c.ResolvedOptions().CaseFirst; got != collator.FalseCaseFirst {
				t.Fatalf("ResolvedOptions().CaseFirst = %q, want %q", got, collator.FalseCaseFirst)
			}
		})
	}
}

func TestCollator_New_Errors(t *testing.T) {
	t.Parallel()
	en := intltest.Locale(t, "en")

	t.Run("invalid usage", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Usage: gointl.String("bogus")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "usage", "bogus", en.String())
		testcontract.AssertOptionExpected(t, err, `one of "sort", "search"`)
	})

	t.Run("empty usage", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Usage: gointl.String("")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "usage", "", en.String())
	})

	t.Run("invalid sensitivity", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Sensitivity: gointl.String("bogus")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "sensitivity", "bogus", en.String())
	})

	t.Run("empty sensitivity", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Sensitivity: gointl.String("")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "sensitivity", "", en.String())
	})

	t.Run("invalid locale matcher", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{LocaleMatcher: gointl.String("bogus")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "localeMatcher", "bogus", en.String())
	})

	t.Run("empty locale matcher", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{LocaleMatcher: gointl.String("")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "localeMatcher", "", en.String())
	})

	t.Run("invalid case first", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{CaseFirst: gointl.String("bogus")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "caseFirst", "bogus", en.String())
	})

	t.Run("empty case first", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{CaseFirst: gointl.String("")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "caseFirst", "", en.String())
	})

	t.Run("unsupported search usage", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Usage: gointl.String(collator.SearchUsage)})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.UnsupportedOption, "usage", string(collator.SearchUsage), en.String())
		testcontract.AssertOptionExpected(t, err, `"sort"`)
	})

	t.Run("unsupported case first", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{CaseFirst: gointl.String(collator.UpperCaseFirst)})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.UnsupportedOption, "caseFirst", string(collator.UpperCaseFirst), en.String())
		testcontract.AssertOptionExpected(t, err, `"false"`)
	})

	t.Run("unsupported lower case first", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{CaseFirst: gointl.String(collator.LowerCaseFirst)})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.UnsupportedOption, "caseFirst", string(collator.LowerCaseFirst), en.String())
		testcontract.AssertOptionExpected(t, err, `"false"`)
	})

	t.Run("invalid collation syntax", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Collation: gointl.String("a")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "collation", "a", en.String())
	})

	t.Run("empty collation syntax", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Collation: gointl.String("")})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "collation", "", en.String())
	})

	t.Run("unicode folded collation syntax", func(t *testing.T) {
		t.Parallel()
		const foldedASCII = "\u212Aphonebk"
		_, err := collator.New(locale.List{en}, collator.Options{Collation: gointl.String(foldedASCII)})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "collation", foldedASCII, en.String())
	})
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()
	requested := locale.List{intltest.Locale(t, "en-US"), intltest.Locale(t, "xh")}
	got, err := collator.SupportedLocalesOf(requested, collator.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected at least one supported locale, got none")
	}
	if got[0].String() != "en-US" {
		t.Errorf("first supported = %q, want en-US", got[0].String())
	}
}

func TestSupportedLocalesOfPreservesUnicodeExtensions(t *testing.T) {
	t.Parallel()

	requested := locale.List{
		intltest.Locale(t, "en-u-co-default"),
		intltest.Locale(t, "en-u-co-phonebk"),
		intltest.Locale(t, "en-u-co-search"),
		intltest.Locale(t, "en-u-kf-false"),
		intltest.Locale(t, "en-u-kf-upper"),
	}
	got, err := collator.SupportedLocalesOf(requested, collator.Options{})
	if err != nil {
		t.Fatal(err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf", got, []string{"en-u-co-default", "en-u-co-phonebk", "en-u-co-search", "en-u-kf-false", "en-u-kf-upper"})
}

func TestSupportedLocalesOfErrors(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "en-US")}
	for _, matcher := range []string{"bogus", ""} {
		t.Run(matcher, func(t *testing.T) {
			t.Parallel()
			_, err := collator.SupportedLocalesOf(requested, collator.Options{LocaleMatcher: gointl.String(matcher)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "collator", intlerr.InvalidOption, "localeMatcher", matcher, "en-US")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}
}
