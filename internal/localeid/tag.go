package localeid

import (
	"strings"

	"golang.org/x/text/language"
)

func Parts(tag language.Tag) (lang, script, region string) {
	base, scr, reg := tag.Raw()
	lang = base.String()
	if !scr.IsPrivateUse() {
		if s := scr.String(); s != "Zzzz" {
			script = s
		}
	}
	if !reg.IsPrivateUse() {
		if r := reg.String(); r != "ZZ" {
			region = r
		}
	}
	return lang, script, region
}

func Join(lang, script, region string) string {
	return strings.Join(languageSubtags(lang, script, region), "-")
}

func languageSubtags(lang, script, region string) []string {
	parts := []string{lang}
	if script != "" {
		parts = append(parts, script)
	}
	if region != "" {
		parts = append(parts, region)
	}
	return parts
}

// BaseName returns the language identifier without extensions.
func BaseName(tag language.Tag) string {
	lang, script, region := Parts(tag)
	parts := append(languageSubtags(lang, script, region), VariantSubtags(tag)...)
	return strings.Join(parts, "-")
}

// VariantSubtags returns the tag's canonical variant subtags.
func VariantSubtags(tag language.Tag) []string {
	variants := tag.Variants()
	if len(variants) == 0 {
		return nil
	}
	out := make([]string, len(variants))
	for i, variant := range variants {
		out[i] = variant.String()
	}
	return out
}

// ReplaceLanguageSubtags returns tag with only its language, script, and region
// replaced. Variants and extensions retain their canonical order.
func ReplaceLanguageSubtags(tag language.Tag, lang, script, region string) (language.Tag, error) {
	parts := append(languageSubtags(lang, script, region), VariantSubtags(tag)...)
	for _, extension := range tag.Extensions() {
		parts = append(parts, extension.String())
	}
	return language.Parse(strings.Join(parts, "-"))
}
