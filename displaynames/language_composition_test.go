package displaynames_test

import (
	"strings"
	"testing"

	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/option"
)

func TestLanguageCompositionRetainsScriptAndRegion(t *testing.T) {
	t.Parallel()
	for _, tag := range []string{"en", "zh"} {
		for _, display := range []string{"standard", "dialect"} {
			t.Run(tag+"/"+display, func(t *testing.T) {
				t.Parallel()
				d, err := displaynames.New(locale.List{intltest.Locale(t, tag)}, displaynames.Options{Type: option.String("language"), LanguageDisplay: option.String(display)})
				if err != nil {
					t.Fatal(err)
				}
				for _, tc := range []struct {
					code      string
					fragments []string
				}{
					{"en-Cyrl", []string{"Cyrillic"}},
					{"en-Latn-US", []string{"Latin"}},
				} {
					name, ok, err := d.Of(tc.code)
					if err != nil || !ok {
						t.Fatalf("%s: %q %v %v", tc.code, name, ok, err)
					}
					if tag == "en" {
						for _, part := range tc.fragments {
							if !strings.Contains(name, part) {
								t.Errorf("%s => %q missing %s", tc.code, name, part)
							}
						}
					}
					if name == "English" || name == "英语" {
						t.Errorf("%s lost component in %q", tc.code, name)
					}
				}
			})
		}
	}
}

func TestLanguageVariantsAndFallback(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		code      string
		fragments []string
	}{
		{"sl-rozaj", []string{"Slovenian", "Resian"}},
		{"de-1901", []string{"German", "Traditional German orthography"}},
		{"sl-Latn-SI-rozaj", []string{"Resian", "Latin"}},
		{"sl-rozaj-1994", []string{"Resian", "Standardized Resian orthography"}},
	} {
		d, err := displaynames.New(locale.List{intltest.Locale(t, "en")}, displaynames.Options{Type: option.String("language")})
		if err != nil {
			t.Fatal(err)
		}
		name, ok, err := d.Of(tc.code)
		if err != nil || !ok {
			t.Fatalf("Of %s=%q,%v,%v", tc.code, name, ok, err)
		}
		for _, fragment := range tc.fragments {
			if !strings.Contains(name, fragment) {
				t.Errorf("%s=%q missing %q", tc.code, name, fragment)
			}
		}
	}
	for _, fallback := range []string{"code", "none"} {
		d, err := displaynames.New(locale.List{intltest.Locale(t, "en")}, displaynames.Options{Type: option.String("language"), Fallback: option.String(fallback)})
		if err != nil {
			t.Fatal(err)
		}
		name, ok, err := d.Of("en-Qaaa")
		if err != nil {
			t.Fatal(err)
		}
		if fallback == "none" && ok {
			t.Errorf("none returned %q", name)
		}
		if fallback == "code" && (!ok || !strings.Contains(name, "Qaaa")) {
			t.Errorf("code returned %q,%v", name, ok)
		}
	}
}

func TestUnknownLanguageComponentFallback(t *testing.T) {
	t.Parallel()
	for _, display := range []string{"standard", "dialect"} {
		for _, fallback := range []string{"code", "none"} {
			names, err := displaynames.New(intltest.LocaleList(t, "en"), displaynames.Options{Type: new("language"), LanguageDisplay: new(display), Fallback: new(fallback)})
			if err != nil {
				t.Fatal(err)
			}
			for _, tc := range []struct{ code, want string }{{"zz-US", "zz (United States)"}, {"qaa-Latn-US", "qaa (Latin, United States)"}, {"zz-rozaj", "zz (Resian)"}, {"zz", "zz"}} {
				want, wantOK := tc.want, true
				if fallback == "none" {
					want, wantOK = "", false
				}
				got, ok, err := names.Of(tc.code)
				if err != nil || ok != wantOK || got != want {
					t.Errorf("%s/%s %s = %q,%v,%v; want %q,%v", display, fallback, tc.code, got, ok, err, want, wantOK)
				}
			}
		}
	}
}
