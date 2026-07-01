package displaynames_test

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func TestDisplayNames_OfRejectsInvalidShape(t *testing.T) {
	t.Parallel()
	en := intltest.Locale(t, "en")
	tests := []struct {
		name string
		opts displaynames.Options
		code string
		want string
	}{
		{"region-too-long", displaynames.Options{Type: stringPtr(displaynames.Region)}, "USA", ecma402.DisplayNamesCodeExpected("region")},
		{"region-with-digit", displaynames.Options{Type: stringPtr(displaynames.Region)}, "U1", ecma402.DisplayNamesCodeExpected("region")},
		{"region-empty", displaynames.Options{Type: stringPtr(displaynames.Region)}, "", ecma402.DisplayNamesCodeExpected("region")},
		{"script-three-letter", displaynames.Options{Type: stringPtr(displaynames.Script)}, "lat", ecma402.DisplayNamesCodeExpected("script")},
		{"script-five-letter", displaynames.Options{Type: stringPtr(displaynames.Script)}, "latnn", ecma402.DisplayNamesCodeExpected("script")},
		{"script-with-digit", displaynames.Options{Type: stringPtr(displaynames.Script)}, "lat1", ecma402.DisplayNamesCodeExpected("script")},
		{"currency-two-letter", displaynames.Options{Type: stringPtr(displaynames.Currency)}, "US", ecma402.DisplayNamesCodeExpected("currency")},
		{"currency-with-digit", displaynames.Options{Type: stringPtr(displaynames.Currency)}, "US1", ecma402.DisplayNamesCodeExpected("currency")},
		{"calendar-too-short", displaynames.Options{Type: stringPtr(displaynames.Calendar)}, "gr", ecma402.DisplayNamesCodeExpected("calendar")},
		{"calendar-underscore", displaynames.Options{Type: stringPtr(displaynames.Calendar)}, "invalid_calendar", ecma402.DisplayNamesCodeExpected("calendar")},
		{"dateTimeField-unknown", displaynames.Options{Type: stringPtr(displaynames.DateTimeField)}, "century", ecma402.DisplayNamesCodeExpected("dateTimeField")},
		{"language-empty", displaynames.Options{Type: stringPtr(displaynames.Language)}, "", ecma402.DisplayNamesCodeExpected("language")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dn, err := displaynames.New(locale.List{en}, tc.opts)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			got, ok, err := dn.Of(tc.code)
			if !errors.Is(err, intlerr.ErrInvalidCode) {
				t.Fatalf("Of(%q) error = %v, want intlerr.ErrInvalidCode", tc.code, err)
			}
			if ok || got != "" {
				t.Fatalf("Of(%q) = (%q, %v), want (\"\", false)", tc.code, got, ok)
			}
			testcontract.AssertIntlError(t, err, intlerr.InvalidCode, "displaynames", string(*tc.opts.Type), tc.code, en.String())
			testcontract.AssertErrorExpected(t, err, tc.want)
		})
	}
}

func TestDisplayNames_OfInvalidCodeDoesNotMatchInvalidOption(t *testing.T) {
	t.Parallel()

	en := intltest.Locale(t, "en")
	dn, err := displaynames.New(locale.List{en}, displaynames.Options{Type: stringPtr(displaynames.Language)})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = dn.Of("bad_code")
	if !errors.Is(err, intlerr.ErrInvalidCode) {
		t.Fatalf("Of(invalid code) error = %v, want intlerr.ErrInvalidCode", err)
	}
	if errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("Of(invalid code) error = %v, must not match intlerr.ErrInvalidOption", err)
	}
	testcontract.AssertIntlError(t, err, intlerr.InvalidCode, "displaynames", string(displaynames.Language), "bad_code", en.String())
}

func TestDisplayNames_LanguageDisplayValidatedForNonLanguageTypes(t *testing.T) {
	t.Parallel()

	en := intltest.Locale(t, "en")
	for _, typ := range []displaynames.Type{
		displaynames.Region,
		displaynames.Script,
		displaynames.Currency,
		displaynames.Calendar,
		displaynames.DateTimeField,
	} {
		t.Run(string(typ), func(t *testing.T) {
			t.Parallel()

			_, err := displaynames.New(locale.List{en}, displaynames.Options{
				Type:            stringPtr(typ),
				LanguageDisplay: stringPtr("bogus"),
			})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(Type=%q, LanguageDisplay=bogus) error = %v, want intlerr.ErrInvalidOption", typ, err)
			}
			testcontract.AssertOptionError(t, err, "displaynames", intlerr.InvalidOption, "languageDisplay", "bogus", en.String())
		})
	}
}
