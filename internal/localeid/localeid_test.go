package localeid

import (
	"testing"

	"golang.org/x/text/language"
)

func TestPartsAndJoin(t *testing.T) {
	t.Parallel()

	lang, script, region := Parts(language.MustParse("en-Latn-US"))
	if lang != "en" || script != "Latn" || region != "US" {
		t.Fatalf("Parts(en-Latn-US) = %q, %q, %q; want en, Latn, US", lang, script, region)
	}
	if got := Join("en", "", "US"); got != "en-US" {
		t.Fatalf("Join(en, empty, US) = %q, want en-US", got)
	}
}

func TestMaximize(t *testing.T) {
	t.Parallel()

	if got := Maximize("bad_locale", nil); got != "bad_locale" {
		t.Fatalf("Maximize(invalid) = %q, want original input", got)
	}
	if got := Maximize("en-us", nil); got != "en-US" {
		t.Fatalf("Maximize(en-us, nil) = %q, want canonical parsed tag", got)
	}
	maximizer := func(lang, script, region string) (string, string, string, bool) {
		if lang == "zh" && script == "" && region == "" {
			return "zh", "Hans", "CN", true
		}
		return "", "", "", false
	}
	if got := Maximize("zh", maximizer); got != "zh-Hans-CN" {
		t.Fatalf("Maximize(zh) = %q, want zh-Hans-CN", got)
	}
	if got := Maximize("fr-CA", maximizer); got != "fr-CA" {
		t.Fatalf("Maximize(fr-CA) = %q, want parsed fallback", got)
	}
}
