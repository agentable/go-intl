package displaynames_test

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func stringPtr[T ~string](v T) *string {
	value := string(v)
	return &value
}

func TestDisplayNames_Of(t *testing.T) {
	t.Parallel()
	en := intltest.Locale(t, "en")

	tests := []struct {
		name string
		opts displaynames.Options
		code string
		want string
		ok   bool
	}{
		{"language", displaynames.Options{Type: stringPtr(displaynames.Language)}, "fr", "French", true},
		{"language-tag", displaynames.Options{Type: stringPtr(displaynames.Language)}, "zh-Hans", "Simplified Chinese", true},
		{"language-standard", displaynames.Options{Type: stringPtr(displaynames.Language), LanguageDisplay: stringPtr(displaynames.StandardLanguageDisplay)}, "en-GB", "English (United Kingdom)", true},
		{"region", displaynames.Options{Type: stringPtr(displaynames.Region)}, "US", "United States", true},
		{"region-lower-canonicalized", displaynames.Options{Type: stringPtr(displaynames.Region)}, "us", "United States", true},
		{"region-short", displaynames.Options{Type: stringPtr(displaynames.Region), Style: stringPtr(displaynames.ShortStyle)}, "GB", "UK", true},
		{"script", displaynames.Options{Type: stringPtr(displaynames.Script)}, "latn", "Latin", true},
		{"currency", displaynames.Options{Type: stringPtr(displaynames.Currency)}, "usd", "US Dollar", true},
		{"currency-narrow", displaynames.Options{Type: stringPtr(displaynames.Currency), Style: stringPtr(displaynames.NarrowStyle)}, "EUR", "Euro", true},
		{"calendar", displaynames.Options{Type: stringPtr(displaynames.Calendar)}, "gregory", "Gregorian Calendar", true},
		{"datetimefield", displaynames.Options{Type: stringPtr(displaynames.DateTimeField)}, "year", "year", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dn, err := displaynames.New(locale.List{en}, tc.opts)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			got, ok, err := dn.Of(tc.code)
			if err != nil {
				t.Fatalf("Of(%q) error = %v", tc.code, err)
			}
			if ok != tc.ok {
				t.Fatalf("Of(%q) ok = %v, want %v", tc.code, ok, tc.ok)
			}
			if got != tc.want {
				t.Errorf("Of(%q) = %q, want %q", tc.code, got, tc.want)
			}
		})
	}
}

func TestDisplayNames_Fallback(t *testing.T) {
	t.Parallel()
	en := intltest.Locale(t, "en")

	t.Run("code returns canonicalized code", func(t *testing.T) {
		t.Parallel()
		dn, err := displaynames.New(locale.List{en}, displaynames.Options{Type: stringPtr(displaynames.Region)})
		if err != nil {
			t.Fatal(err)
		}
		got, ok, err := dn.Of("qq")
		if err != nil {
			t.Fatal(err)
		}
		if !ok || got != "QQ" {
			t.Errorf("Of(qq) = (%q, %v), want (QQ, true)", got, ok)
		}
	})

	t.Run("none returns empty", func(t *testing.T) {
		t.Parallel()
		dn, err := displaynames.New(locale.List{en}, displaynames.Options{Type: stringPtr(displaynames.Region), Fallback: stringPtr(displaynames.NoneFallback)})
		if err != nil {
			t.Fatal(err)
		}
		got, ok, err := dn.Of("qq")
		if err != nil {
			t.Fatal(err)
		}
		if ok || got != "" {
			t.Errorf("Of(qq) = (%q, %v), want (\"\", false)", got, ok)
		}
	})
}

func TestDisplayNames_RejectsInvalidLanguageCodes(t *testing.T) {
	t.Parallel()

	dn, err := displaynames.New(locale.List{intltest.Locale(t, "en")}, displaynames.Options{Type: stringPtr(displaynames.Language)})
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{
		"x-private",
		"en-x-private",
		"en-u-ca-gregory",
		"i-klingon",
		"zh-cmn",
		"abcd",
	} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			_, _, err := dn.Of(code)
			if !errors.Is(err, intlerr.ErrInvalidCode) {
				t.Fatalf("Of(%q) error = %v, want intlerr.ErrInvalidCode", code, err)
			}
		})
	}
}

func TestDisplayNames_New_Errors(t *testing.T) {
	t.Parallel()
	en := intltest.Locale(t, "en")

	tests := []struct {
		name         string
		opts         displaynames.Options
		wantName     string
		wantValue    string
		wantExpected string
	}{
		{
			name:         "missing type",
			opts:         displaynames.Options{},
			wantName:     "type",
			wantExpected: `one of "language", "region", "script", "currency", "calendar", "dateTimeField"`,
		},
		{
			name:         "empty type",
			opts:         displaynames.Options{Type: stringPtr("")},
			wantName:     "type",
			wantExpected: `one of "language", "region", "script", "currency", "calendar", "dateTimeField"`,
		},
		{
			name:         "invalid style",
			opts:         displaynames.Options{Type: stringPtr(displaynames.Region), Style: stringPtr("bogus")},
			wantName:     "style",
			wantValue:    "bogus",
			wantExpected: `one of "long", "short", "narrow"`,
		},
		{
			name:         "empty style",
			opts:         displaynames.Options{Type: stringPtr(displaynames.Region), Style: stringPtr("")},
			wantName:     "style",
			wantExpected: `one of "long", "short", "narrow"`,
		},
		{
			name:         "invalid locale matcher",
			opts:         displaynames.Options{LocaleMatcher: stringPtr("bogus"), Type: stringPtr(displaynames.Region)},
			wantName:     "localeMatcher",
			wantValue:    "bogus",
			wantExpected: `one of "lookup", "best fit"`,
		},
		{
			name:         "empty locale matcher",
			opts:         displaynames.Options{LocaleMatcher: stringPtr(""), Type: stringPtr(displaynames.Region)},
			wantName:     "localeMatcher",
			wantExpected: `one of "lookup", "best fit"`,
		},
		{
			name:         "invalid fallback",
			opts:         displaynames.Options{Type: stringPtr(displaynames.Region), Fallback: stringPtr("bogus")},
			wantName:     "fallback",
			wantValue:    "bogus",
			wantExpected: `one of "code", "none"`,
		},
		{
			name:         "empty fallback",
			opts:         displaynames.Options{Type: stringPtr(displaynames.Region), Fallback: stringPtr("")},
			wantName:     "fallback",
			wantExpected: `one of "code", "none"`,
		},
		{
			name:         "invalid language display",
			opts:         displaynames.Options{Type: stringPtr(displaynames.Language), LanguageDisplay: stringPtr("bogus")},
			wantName:     "languageDisplay",
			wantValue:    "bogus",
			wantExpected: `one of "dialect", "standard"`,
		},
		{
			name:         "empty language display",
			opts:         displaynames.Options{Type: stringPtr(displaynames.Language), LanguageDisplay: stringPtr("")},
			wantName:     "languageDisplay",
			wantExpected: `one of "dialect", "standard"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := displaynames.New(locale.List{en}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Errorf("err = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "displaynames", intlerr.InvalidOption, tc.wantName, tc.wantValue, en.String())
			testcontract.AssertOptionExpected(t, err, tc.wantExpected)
		})
	}
}

func TestDisplayNames_ResolvedOptions(t *testing.T) {
	t.Parallel()
	en := intltest.Locale(t, "en")
	dn, err := displaynames.New(locale.List{en}, displaynames.Options{Type: stringPtr(displaynames.Region)})
	if err != nil {
		t.Fatal(err)
	}
	got := dn.ResolvedOptions()
	if got.Type != displaynames.Region {
		t.Errorf("Type = %q, want %q", got.Type, displaynames.Region)
	}
	if got.Style != displaynames.LongStyle {
		t.Errorf("Style = %q, want %q", got.Style, displaynames.LongStyle)
	}
	if got.Fallback != displaynames.CodeFallback {
		t.Errorf("Fallback = %q, want %q", got.Fallback, displaynames.CodeFallback)
	}
	// ECMA-402 §12.4.2: LanguageDisplay is only written for Type=language.
	if got.LanguageDisplay != nil {
		t.Errorf("LanguageDisplay = %v, want nil for Type=Region", got.LanguageDisplay)
	}
}

func TestDisplayNames_ResolvedOptionsLanguageDisplayPresentForLanguageType(t *testing.T) {
	t.Parallel()
	en := intltest.Locale(t, "en")
	dn, err := displaynames.New(locale.List{en}, displaynames.Options{Type: stringPtr(displaynames.Language), LanguageDisplay: stringPtr(displaynames.StandardLanguageDisplay)})
	if err != nil {
		t.Fatal(err)
	}
	got := dn.ResolvedOptions()
	if got.LanguageDisplay == nil || *got.LanguageDisplay != displaynames.StandardLanguageDisplay {
		t.Errorf("LanguageDisplay = %v, want %q", got.LanguageDisplay, displaynames.StandardLanguageDisplay)
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()
	requested := locale.List{intltest.Locale(t, "en-US"), intltest.Locale(t, "xh")}
	got, err := displaynames.SupportedLocalesOf(requested, displaynames.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].String() != "en-US" {
		t.Errorf("SupportedLocalesOf = %v, want [en-US]", got)
	}
}

func TestSupportedLocalesOfErrors(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "en-US")}
	for _, matcher := range []string{"bogus", ""} {
		t.Run(matcher, func(t *testing.T) {
			t.Parallel()
			_, err := displaynames.SupportedLocalesOf(requested, displaynames.Options{LocaleMatcher: stringPtr(matcher)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "displaynames", intlerr.InvalidOption, "localeMatcher", matcher, "en-US")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}
}
