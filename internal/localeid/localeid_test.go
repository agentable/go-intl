package localeid

import (
	"strings"
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

func TestUnicodeExtensionCanonicalization(t *testing.T) {
	t.Parallel()

	ext, err := ParseUnicodeExtension("-u-attr2-attr1-ca-buddhist-ca-gregory-kk-true")
	if err != nil {
		t.Fatalf("ParseUnicodeExtension() error = %v", err)
	}
	if got, want := ext.ValueForKey("ca"), "buddhist"; got != want {
		t.Fatalf("ValueForKey(ca) = %q, want %q", got, want)
	}
	if got, want := strings.Join(ext.Parts(), "-"), "attr1-attr2-ca-buddhist-kk"; got != want {
		t.Fatalf("Parts() = %q, want %q", got, want)
	}
}

func TestRemoveUnicodeExtensionPreservesOtherExtensions(t *testing.T) {
	t.Parallel()

	base, extension, err := RemoveUnicodeExtension("sr-cyrl-rs-t-ja-u-ca-islamic-x-whatever")
	if err != nil {
		t.Fatalf("RemoveUnicodeExtension() error = %v", err)
	}
	if base != "sr-cyrl-rs-t-ja-x-whatever" || extension != "-u-ca-islamic" {
		t.Fatalf("RemoveUnicodeExtension() = %q, %q; want base without u and u extension", base, extension)
	}
}

func TestRemoveUnicodeExtensionIgnoresPrivateUseU(t *testing.T) {
	t.Parallel()

	base, extension, err := RemoveUnicodeExtension("en-x-foo-u-ca-gregory")
	if err != nil {
		t.Fatalf("RemoveUnicodeExtension() error = %v", err)
	}
	if base != "en-x-foo-u-ca-gregory" || extension != "" {
		t.Fatalf("RemoveUnicodeExtension() = %q, %q; want private-use u left untouched", base, extension)
	}
}

func TestInsertUnicodeExtensionBeforePrivateUse(t *testing.T) {
	t.Parallel()

	got := InsertUnicodeExtension("en-x-private", nil, []UnicodeKeyword{{Key: "ca", Value: "gregory"}})
	if want := "en-u-ca-gregory-x-private"; got != want {
		t.Fatalf("InsertUnicodeExtension() = %q, want %q", got, want)
	}
}
