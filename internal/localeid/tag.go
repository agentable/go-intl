package localeid

import (
	"strings"

	"golang.org/x/text/language"
)

// ParseTag uses x/text's Unicode aliases while retaining explicit scripts.
func ParseTag(s string) (language.Tag, error) {
	tag, err := language.Parse(s)
	if err != nil {
		return tag, err
	}
	// Apply macro aliases separately: Legacy|CLDR would conflate no and nb.
	return (language.Macro | language.CLDR).Canonicalize(tag)
}

// Parts projects explicit subtags from the canonical tag. Raw conflates absent
// script/region fields with the explicit unknown identifiers Zzzz and ZZ.
func Parts(tag language.Tag) (lang, script, region string) {
	lang, rest, _ := strings.Cut(tag.String(), "-")
	subtag, rest, _ := strings.Cut(rest, "-")
	if IsUnicodeScriptSubtag(subtag) {
		script = subtag
		subtag, _, _ = strings.Cut(rest, "-")
	}
	if IsUnicodeRegionSubtag(subtag) {
		region = subtag
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
	return ParseTag(strings.Join(parts, "-"))
}
