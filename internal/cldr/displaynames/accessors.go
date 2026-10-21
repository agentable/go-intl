// Hand-written accessor layer for the displaynames domain. The query semantics
// mirror the ECMA-402 DisplayNames lookup surface. Currency display names reuse
// the internal/cldr currency name accessors, not NumberFormat symbols.

package displaynames

import (
	"strings"

	"github.com/agentable/go-intl/internal/cldr/currency"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
	"github.com/agentable/go-intl/internal/pattern"
)

const (
	// ICU LocaleDisplayNamesImpl::initialize supplies this pattern and the
	// separator below when localeDisplayPattern data is absent.
	defaultLocalePattern   = "{0} ({1})"
	defaultLocaleSeparator = "{0}, {1}"
)

// Of returns the localized display name for a code and whether one exists.
//
// kind: "language" | "region" | "script" | "currency" | "calendar" | "dateTimeField".
// style: "long" | "short" | "narrow".
// languageDisplay: "dialect" | "standard" (only meaningful when kind == "language").
// fallbackCode controls whether language-with-region composition can use the
// region code when no localized region name exists.
//
// When data for the requested locale is unavailable, the lookup walks the
// truncation parent chain (e.g. en-US -> en).
func Of(dataLocale, kind, style, languageDisplay, code string, fallbackCode bool) (string, bool) {
	if kind == "calendar" {
		code = calendarLookupCode(code)
	}
	if kind == "dateTimeField" {
		code = dateTimeFieldLookupCode(code)
	}
	for tag := dataLocale; tag != ""; tag = parentTag(tag) {
		if value, ok := lookupInLocale(tag, kind, style, languageDisplay, code, fallbackCode); ok {
			return value, true
		}
	}
	return "", false
}

// calendarLookupCode maps ECMA-402 calendar identifiers (Unicode BCP 47 `u-ca`
// keys) to CLDR's localeDisplayNames.types.calendar keys.
func calendarLookupCode(code string) string {
	switch code {
	case "gregory":
		return "gregorian"
	case "ethioaa":
		return "ethiopic-amete-alem"
	default:
		return code
	}
}

func dateTimeFieldLookupCode(code string) string {
	if code == "dayPeriod" {
		return "dayperiod"
	}
	return code
}

// SupportedLocales returns the locale tags with display-name data. It reads only
// the narrow supported blob and never triggers any names blob decode.
func SupportedLocales() []string {
	return supported.Get()
}

func parentTag(tag string) string {
	idx := strings.LastIndex(tag, "-")
	if idx < 0 {
		return ""
	}
	return tag[:idx]
}

func lookupInLocale(tag, kind, style, languageDisplay, code string, fallbackCode bool) (string, bool) {
	switch kind {
	case "language":
		rec, ok := languageData()[tag]
		if !ok {
			return "", false
		}
		display := rec.display.dialect
		if languageDisplay == "standard" {
			display = rec.display.standard
		}
		if value, ok := resolveStyled(display, style, code); ok && (languageDisplay != "standard" || !strings.Contains(code, "-")) {
			return value, true
		}
		return resolveLanguage(tag, rec.localePattern, rec.localeSeparator, display, style, code, fallbackCode, languageDisplay != "standard")
	case "region":
		return resolveStyledForTag(territoryData(), tag, style, code)
	case "script":
		return resolveStyledForTag(scriptData(), tag, style, code)
	case "currency":
		return currencyDisplay(tag, code)
	case "calendar":
		return resolveStyledForTag(calendarData(), tag, style, code)
	case "dateTimeField":
		return resolveStyledForTag(fieldData(), tag, style, code)
	}
	return "", false
}

func resolveStyledForTag(byLocale map[string]styledNames, tag, style, code string) (string, bool) {
	s, ok := byLocale[tag]
	if !ok {
		return "", false
	}
	return resolveStyled(s, style, code)
}

func resolveLanguage(tag, localePattern, localeSeparator string, display styledNames, style, code string, fallbackCode, dialect bool) (string, bool) {
	parts := strings.Split(code, "-")
	if len(parts) == 1 {
		return "", false
	}
	base, ok := resolveStyled(display, style, parts[0])
	if !ok {
		if !fallbackCode {
			return "", false
		}
		base = parts[0]
	}
	index := 1
	script := ""
	if index < len(parts) && localeid.IsUnicodeScriptSubtag(parts[index]) {
		script = parts[index]
		index++
	}
	region := ""
	if index < len(parts) && localeid.IsUnicodeRegionSubtag(parts[index]) {
		region = parts[index]
		index++
	}
	// A dialect row may consume the script and/or region. A standard name
	// deliberately starts from the bare language and composes every component.
	if dialect && script != "" && region != "" {
		if value, found := resolveStyled(display, style, parts[0]+"-"+script+"-"+region); found {
			base = value
			script = ""
			region = ""
		}
	}
	if dialect && script != "" {
		if value, found := resolveStyled(display, style, parts[0]+"-"+script); found {
			base = value
			script = ""
		}
	}
	if dialect && region != "" {
		if value, found := resolveStyled(display, style, parts[0]+"-"+region); found {
			base = value
			region = ""
		}
	}
	components := make([]string, 0, 2+len(parts)-index)
	appendName := func(value string, found bool, code string) bool {
		if !found {
			if !fallbackCode {
				return false
			}
			value = code
		}
		components = append(components, value)
		return true
	}
	if script != "" {
		value, found := resolveStyledForTag(scriptData(), tag, style, script)
		if !appendName(value, found, script) {
			return "", false
		}
	}
	if region != "" {
		value, found := resolveStyledForTag(territoryData(), tag, style, region)
		if !appendName(value, found, region) {
			return "", false
		}
	}
	for _, variant := range parts[index:] {
		value, found := resolveStyledForTag(variantData(), tag, style, strings.ToUpper(variant))
		if !appendName(value, found, variant) {
			return "", false
		}
	}
	if len(components) == 0 {
		return base, true
	}
	joined := components[0]
	if localeSeparator == "" {
		localeSeparator = defaultLocaleSeparator
	}
	for _, part := range components[1:] {
		joined = pattern.FormatIndexed(localeSeparator, joined, part)
	}
	return applyLocalePattern(localePattern, base, joined), true
}

func applyLocalePattern(text, language, region string) string {
	if text == "" {
		text = defaultLocalePattern
	}
	if strings.ContainsRune(text, '（') {
		region = strings.ReplaceAll(strings.ReplaceAll(region, "（", "［"), "）", "］")
	} else {
		region = strings.ReplaceAll(strings.ReplaceAll(region, "(", "["), ")", "]")
	}
	return pattern.FormatIndexed(text, language, region)
}

func resolveStyled(s styledNames, style, code string) (string, bool) {
	switch style {
	case "narrow":
		if v, ok := s.narrow[code]; ok {
			return v, true
		}
		if v, ok := s.short[code]; ok {
			return v, true
		}
	case "short":
		if v, ok := s.short[code]; ok {
			return v, true
		}
	}
	if v, ok := s.long[code]; ok {
		return v, true
	}
	return "", false
}

func currencyDisplay(dataLocale, code string) (string, bool) {
	loc, ok := cldrlocale.ResolveLocale(dataLocale)
	if !ok {
		return "", false
	}
	if name := currency.CanonicalName(loc, code); name != "" {
		return name, true
	}
	return "", false
}
