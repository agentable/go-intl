package localeid

import (
	"errors"
	"slices"
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

	ext, err := ParseUnicodeExtension("-u-attr2-attr1-attr2-ca-buddhist-ca-gregory-kk-true")
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

func TestLowercaseUnicodeLocaleIDIsASCIIOnly(t *testing.T) {
	t.Parallel()

	in := "EN-u-CA-\u212AARAB-x-PRIV"
	want := "en-u-ca-\u212Aarab-x-priv"
	if got := LowercaseUnicodeLocaleID(in); got != want {
		t.Fatalf("LowercaseUnicodeLocaleID(%q) = %q, want %q", in, got, want)
	}
}

func TestUnicodeTypeSyntaxAndCanonicalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key           string
		in            string
		wellFormed    bool
		wantCanonical string
		canonicalOK   bool
	}{
		{key: "nu", in: "arab", wellFormed: true, wantCanonical: "arab", canonicalOK: true},
		{key: "nu", in: "ARAB", wellFormed: true, wantCanonical: "arab", canonicalOK: true},
		{key: "nu", in: "ARAB-LATN", wellFormed: true, wantCanonical: "arab-latn", canonicalOK: true},
		{key: "ca", in: "islamicc", wellFormed: true, wantCanonical: "islamic-civil", canonicalOK: true},
		{key: "ca", in: "islamic-civil", wellFormed: true, wantCanonical: "islamic-civil", canonicalOK: true},
		{key: "co", in: "islamic-civil", wellFormed: true, wantCanonical: "islamic-civil", canonicalOK: true},
		{key: "ca", in: "gregorian", wellFormed: false, canonicalOK: false},
		{in: "\u212Aarab", wellFormed: false, canonicalOK: false},
		{in: "a", wellFormed: false, canonicalOK: false},
		{in: "abcdefghi", wellFormed: false, canonicalOK: false},
		{in: "bad!", wellFormed: false, canonicalOK: false},
		{in: "", wellFormed: false, canonicalOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := IsUnicodeType(tc.in); got != tc.wellFormed {
				t.Fatalf("IsUnicodeType(%q) = %v, want %v", tc.in, got, tc.wellFormed)
			}
			got, ok := CanonicalUnicodeType(tc.key, tc.in)
			if ok != tc.canonicalOK || got != tc.wantCanonical {
				t.Fatalf("CanonicalUnicodeType(%q, %q) = %q, %v; want %q, %v",
					tc.key, tc.in, got, ok, tc.wantCanonical, tc.canonicalOK)
			}
		})
	}
}

func TestUnicodeLanguageIdentifierSubtagSyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		language bool
		script   bool
		region   bool
		variant  bool
	}{
		{name: "language alpha 2", value: "en", language: true, region: true},
		{name: "language alpha 3", value: "fil", language: true},
		{name: "language alpha 5", value: "abcde", language: true, variant: true},
		{name: "language alpha 8", value: "abcdefgh", language: true, variant: true},
		{name: "language alpha 4 rejected", value: "abcd", script: true},
		{name: "language digit rejected", value: "en1"},
		{name: "script", value: "Latn", script: true},
		{name: "script digit rejected", value: "Lat1"},
		{name: "alpha region", value: "US", language: true, region: true},
		{name: "numeric region", value: "419", region: true},
		{name: "mixed region rejected", value: "U1"},
		{name: "variant alpha numeric", value: "emodeng", language: true, variant: true},
		{name: "variant uppercase", value: "VARIANT", language: true, variant: true},
		{name: "variant digit leading", value: "1901", variant: true},
		{name: "variant too short rejected", value: "abc", language: true},
		{name: "empty rejected"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := IsUnicodeLanguageSubtag(tc.value); got != tc.language {
				t.Fatalf("IsUnicodeLanguageSubtag(%q) = %v, want %v", tc.value, got, tc.language)
			}
			if got := IsUnicodeScriptSubtag(tc.value); got != tc.script {
				t.Fatalf("IsUnicodeScriptSubtag(%q) = %v, want %v", tc.value, got, tc.script)
			}
			if got := IsUnicodeRegionSubtag(tc.value); got != tc.region {
				t.Fatalf("IsUnicodeRegionSubtag(%q) = %v, want %v", tc.value, got, tc.region)
			}
			if got := IsUnicodeVariantSubtag(tc.value); got != tc.variant {
				t.Fatalf("IsUnicodeVariantSubtag(%q) = %v, want %v", tc.value, got, tc.variant)
			}
		})
	}
}

func TestUnicodeLanguageIdentifierSubtagCanonicalization(t *testing.T) {
	t.Parallel()

	type canonicalFunc func(string) (string, bool)
	tests := []struct {
		name      string
		canonical canonicalFunc
		in        string
		want      string
		wantOK    bool
	}{
		{
			name:      "language lower",
			canonical: CanonicalUnicodeLanguageSubtag,
			in:        "EN",
			want:      "en",
			wantOK:    true,
		},
		{
			name:      "language rejects digit",
			canonical: CanonicalUnicodeLanguageSubtag,
			in:        "en1",
		},
		{
			name:      "script title",
			canonical: CanonicalUnicodeScriptSubtag,
			in:        "hANS",
			want:      "Hans",
			wantOK:    true,
		},
		{
			name:      "script rejects digit",
			canonical: CanonicalUnicodeScriptSubtag,
			in:        "Lat1",
		},
		{
			name:      "region upper",
			canonical: CanonicalUnicodeRegionSubtag,
			in:        "us",
			want:      "US",
			wantOK:    true,
		},
		{
			name:      "numeric region unchanged",
			canonical: CanonicalUnicodeRegionSubtag,
			in:        "419",
			want:      "419",
			wantOK:    true,
		},
		{
			name:      "region rejects mixed alnum",
			canonical: CanonicalUnicodeRegionSubtag,
			in:        "U1",
		},
		{
			name:      "variant lower",
			canonical: CanonicalUnicodeVariantSubtag,
			in:        "VARIANT",
			want:      "variant",
			wantOK:    true,
		},
		{
			name:      "digit-leading variant unchanged",
			canonical: CanonicalUnicodeVariantSubtag,
			in:        "1901",
			want:      "1901",
			wantOK:    true,
		},
		{
			name:      "variant rejects short alpha",
			canonical: CanonicalUnicodeVariantSubtag,
			in:        "abc",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := tc.canonical(tc.in)
			if ok != tc.wantOK || got != tc.want {
				t.Fatalf("canonical(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestRelevantExtensionValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		defaultValue string
		values       []string
		want         []string
	}{
		{
			name:         "default first and duplicates skipped",
			defaultValue: "latn",
			values:       []string{"latn", "arab", "", "thai", "arab"},
			want:         []string{"latn", "arab", "thai"},
		},
		{
			name:   "empty default",
			values: []string{"h12", "h23", "h12"},
			want:   []string{"h12", "h23"},
		},
		{
			name:         "default absent from supported values",
			defaultValue: "thai",
			values:       []string{"latn", "arab"},
			want:         []string{"thai", "latn", "arab"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := RelevantExtensionValues(tc.defaultValue, tc.values...); !slices.Equal(got, tc.want) {
				t.Fatalf("RelevantExtensionValues(%q, %v) = %v, want %v", tc.defaultValue, tc.values, got, tc.want)
			}
		})
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

func TestRemoveUnicodeExtensionAtEnd(t *testing.T) {
	t.Parallel()

	base, extension, err := RemoveUnicodeExtension("en-latn-u-nu-latn")
	if err != nil {
		t.Fatalf("RemoveUnicodeExtension() error = %v", err)
	}
	if base != "en-latn" || extension != "-u-nu-latn" {
		t.Fatalf("RemoveUnicodeExtension() = %q, %q; want terminal u extension removed", base, extension)
	}
}

func TestRemoveUnicodeExtensionRejectsEmptyBase(t *testing.T) {
	t.Parallel()

	base, extension, err := RemoveUnicodeExtension("u-ca-gregory")
	if !errors.Is(err, ErrInvalidUnicodeExtension) {
		t.Fatalf("RemoveUnicodeExtension() error = %v, want ErrInvalidUnicodeExtension", err)
	}
	if base != "" || extension != "" {
		t.Fatalf("RemoveUnicodeExtension() = %q, %q; want empty result on invalid base", base, extension)
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
