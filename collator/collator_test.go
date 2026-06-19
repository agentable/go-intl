package collator_test

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestCollatorUnsupportedOptionsUseSingleErrorHelper(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			imports := map[string]string{}
			for _, spec := range parsed.Imports {
				path := strings.Trim(spec.Path.Value, `"`)
				name := filepath.Base(path)
				if spec.Name != nil {
					name = spec.Name.Name
				}
				imports[name] = path
			}

			var currentFunc string
			ast.Inspect(parsed, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.FuncDecl:
					currentFunc = node.Name.Name
					return true
				case *ast.CallExpr:
					if currentFunc == "unsupportedOption" {
						return true
					}
					if selectorUsesImport(node.Fun, imports, "github.com/agentable/go-intl/internal/ecma402", "UnsupportedOptionError") {
						t.Fatalf("%s calls ecma402.UnsupportedOptionError outside unsupportedOption", file)
					}
				case *ast.SelectorExpr:
					if currentFunc == "unsupportedOption" {
						return true
					}
					if selectorUsesImport(node, imports, "github.com/agentable/go-intl/internal/intlerr", "ErrUnsupportedOption") {
						t.Fatalf("%s references intlerr.ErrUnsupportedOption outside unsupportedOption", file)
					}
				}
				return true
			})
		})
	}
}

func selectorUsesImport(expr ast.Expr, imports map[string]string, importPath, selectorName string) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != selectorName {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && imports[ident.Name] == importPath
}

func assertOptionError(t *testing.T, err error, kind, name, value, loc string) {
	t.Helper()

	wantKind := kind
	switch kind {
	case "invalid":
		wantKind = "invalidOption"
	case "unsupported":
		wantKind = "unsupportedOption"
	}
	optErr, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("error = %T, want OptionError", err)
	}
	if optErr.Owner != "collator" || string(optErr.Kind) != wantKind || optErr.Name != name || optErr.Value != value || optErr.Locale != loc {
		t.Fatalf("OptionError = %+v, want owner=collator kind=%q name=%q value=%q locale=%q", optErr, kind, name, value, loc)
	}
}

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
	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{Sensitivity: collator.BaseSensitivity})
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
	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{Sensitivity: collator.VariantSensitivity})
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

	c, err := collator.New(locale.List{intltest.Locale(t, "en-u-kf-upper")}, collator.Options{CaseFirst: collator.FalseCaseFirst})
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

func TestCollator_DefaultCollationRequestsResolveToDefault(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		locales locale.List
		options collator.Options
	}{
		{
			name:    "explicit default",
			locales: locale.List{intltest.Locale(t, "en")},
			options: collator.Options{Collation: "default"},
		},
		{
			name:    "locale co default",
			locales: locale.List{intltest.Locale(t, "en-u-co-default")},
			options: collator.Options{},
		},
		{
			name:    "locale co standard",
			locales: locale.List{intltest.Locale(t, "en-u-co-standard")},
			options: collator.Options{},
		},
		{
			name:    "explicit default overrides unsupported locale co",
			locales: locale.List{intltest.Locale(t, "en-u-co-phonebk")},
			options: collator.Options{Collation: "default"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := collator.New(tc.locales, tc.options)
			if err != nil {
				t.Fatal(err)
			}
			if got := c.ResolvedOptions().Collation; got != "default" {
				t.Fatalf("ResolvedOptions().Collation = %q, want default", got)
			}
		})
	}
}

func TestCollator_IgnorePunctuation(t *testing.T) {
	t.Parallel()

	withPunctuation, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{
		IgnorePunctuation: boolPtr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := withPunctuation.Compare("a-b", "ab"); got != 0 {
		t.Errorf("ignore punctuation Compare(a-b,ab) = %d, want 0", got)
	}
	if !withPunctuation.ResolvedOptions().IgnorePunctuation {
		t.Error("ResolvedOptions().IgnorePunctuation = false, want true")
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
	c, err := collator.New(locale.List{intltest.Locale(t, "en")}, collator.Options{Sensitivity: collator.AccentSensitivity})
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
			options: collator.Options{CaseFirst: collator.FalseCaseFirst},
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
		_, err := collator.New(locale.List{en}, collator.Options{Usage: "bogus"})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		assertOptionError(t, err, "invalid", "usage", "bogus", en.String())
	})

	t.Run("invalid sensitivity", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Sensitivity: "bogus"})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		assertOptionError(t, err, "invalid", "sensitivity", "bogus", en.String())
	})

	t.Run("invalid locale matcher", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{LocaleMatcher: "bogus"})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		assertOptionError(t, err, "invalid", "localeMatcher", "bogus", en.String())
	})

	t.Run("invalid case first", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{CaseFirst: "bogus"})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		assertOptionError(t, err, "invalid", "caseFirst", "bogus", en.String())
	})

	t.Run("unsupported search usage", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Usage: collator.SearchUsage})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "usage", string(collator.SearchUsage), en.String())
	})

	t.Run("unsupported case first", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{CaseFirst: collator.UpperCaseFirst})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "caseFirst", string(collator.UpperCaseFirst), en.String())
	})

	t.Run("unsupported lower case first", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{CaseFirst: collator.LowerCaseFirst})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "caseFirst", string(collator.LowerCaseFirst), en.String())
	})

	t.Run("unsupported locale case first", func(t *testing.T) {
		t.Parallel()
		loc := intltest.Locale(t, "en-u-kf-upper")
		_, err := collator.New(locale.List{loc}, collator.Options{})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "caseFirst", string(collator.UpperCaseFirst), loc.String())
	})

	t.Run("unsupported lower locale case first", func(t *testing.T) {
		t.Parallel()
		loc := intltest.Locale(t, "en-u-kf-lower")
		_, err := collator.New(locale.List{loc}, collator.Options{})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "caseFirst", string(collator.LowerCaseFirst), loc.String())
	})

	t.Run("unsupported collation", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Collation: "phonebk"})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "collation", "phonebk", en.String())
	})

	t.Run("unsupported locale collation", func(t *testing.T) {
		t.Parallel()
		loc := intltest.Locale(t, "en-u-co-phonebk")
		_, err := collator.New(locale.List{loc}, collator.Options{})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "collation", "phonebk", loc.String())
	})

	t.Run("unsupported derived locale collation reports matched locale", func(t *testing.T) {
		t.Parallel()
		loc := intltest.Locale(t, "zh-HK-u-co-phonebk")
		_, err := collator.New(locale.List{loc}, collator.Options{})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "collation", "phonebk", "zh-u-co-phonebk")
	})

	t.Run("invalid collation syntax", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Collation: "a"})
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
		}
		assertOptionError(t, err, "invalid", "collation", "a", en.String())
	})

	t.Run("unsupported canonicalized collation", func(t *testing.T) {
		t.Parallel()
		_, err := collator.New(locale.List{en}, collator.Options{Collation: "PHONEBK"})
		if !errors.Is(err, intlerr.ErrUnsupportedOption) {
			t.Errorf("err = %v, want intlerr.ErrUnsupportedOption", err)
		}
		assertOptionError(t, err, "unsupported", "collation", "phonebk", en.String())
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

func TestSupportedLocalesOfErrors(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "en-US")}
	if _, err := collator.SupportedLocalesOf(requested, collator.Options{LocaleMatcher: "bogus"}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
	}
}
