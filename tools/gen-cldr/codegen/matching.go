package codegen

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderLocaleMatching(matching extract.LocaleMatching, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import (\n\t\"slices\"\n\t\"strings\"\n)\n\n")
	b.WriteString("type languageMatch struct{ desired, supported string; distance int; oneway bool }\n\n")
	b.WriteString("var paradigmLocales = []string{\n")
	for _, locale := range matching.Language.ParadigmLocales {
		fmt.Fprintf(&b, "\t%q,\n", locale)
		table.Add(locale)
	}
	b.WriteString("}\n\n")
	b.WriteString("var matchVariables = map[string][]string{\n")
	for _, key := range slices.Sorted(maps.Keys(matching.Language.MatchVariables)) {
		fmt.Fprintf(&b, "\t%q: {", key)
		for _, value := range matching.Language.MatchVariables[key] {
			fmt.Fprintf(&b, "%q, ", value)
			table.Add(value)
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n\n")
	b.WriteString("var languageMatches = []languageMatch{\n")
	for _, match := range matching.Language.Matches {
		fmt.Fprintf(&b, "\t{desired: %q, supported: %q, distance: %d, oneway: %t},\n", match.Desired, match.Supported, match.Distance, match.Oneway)
		table.Add(match.Desired)
		table.Add(match.Supported)
	}
	b.WriteString("}\n\n")
	b.WriteString(`func MatchingDistance(desired, supported string) int {
	if desired == supported {
		return 0
	}
	bestDistance := 100
	bestSpecificity := -1
	for _, match := range languageMatches {
		if specificity, ok := languageMatchSpecificity(match.desired, desired, match.supported, supported); ok && specificity > bestSpecificity {
			bestDistance = match.distance
			bestSpecificity = specificity
		}
		if match.oneway {
			continue
		}
		if specificity, ok := languageMatchSpecificity(match.desired, supported, match.supported, desired); ok && specificity > bestSpecificity {
			bestDistance = match.distance
			bestSpecificity = specificity
		}
	}
	return bestDistance
}

func languageMatchSpecificity(desiredPattern, desiredTag, supportedPattern, supportedTag string) (int, bool) {
	desiredSpecificity, ok := languageMatchApplies(desiredPattern, desiredTag)
	if !ok {
		return 0, false
	}
	supportedSpecificity, ok := languageMatchApplies(supportedPattern, supportedTag)
	if !ok {
		return 0, false
	}
	return desiredSpecificity + supportedSpecificity, true
}

func languageMatchApplies(pattern, tag string) (int, bool) {
	patternParts := strings.Split(pattern, "-")
	tagParts := languageMatchFields(tag)
	if len(patternParts) > len(tagParts) {
		return 0, false
	}
	specificity := 0
	for i, part := range patternParts {
		if part == "*" {
			continue
		}
		if variable, ok := strings.CutPrefix(part, "$!"); ok {
			if slices.Contains(matchVariables["$"+variable], tagParts[i]) {
				return 0, false
			}
			specificity++
			continue
		}
		if _, ok := strings.CutPrefix(part, "$"); ok {
			if !slices.Contains(matchVariables[part], tagParts[i]) {
				return 0, false
			}
			specificity++
			continue
		}
		if part != tagParts[i] {
			return 0, false
		}
		specificity += 2
	}
	return specificity, true
}

func languageMatchFields(tag string) []string {
	fields := []string{"", "", ""}
	for i, part := range strings.Split(tag, "-") {
		if i == 0 {
			fields[0] = part
			continue
		}
		if len(part) == 4 && fields[1] == "" {
			fields[1] = part
			continue
		}
		if (len(part) == 2 || len(part) == 3) && fields[2] == "" {
			fields[2] = part
		}
	}
	return fields
}

func ParadigmLocales() []string { return paradigmLocales }

func MatchVariables() map[string][]string { return matchVariables }
`)
	return FormatFile([]byte(b.String()))
}

func renderRegions(matching extract.LocaleMatching, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("var regionContainment = map[string][]string{\n")
	for _, key := range slices.Sorted(maps.Keys(matching.Regions)) {
		b.WriteString("\t")
		b.WriteString(strconv.Quote(key))
		b.WriteString(": {")
		for _, value := range matching.Regions[key] {
			b.WriteString(strconv.Quote(value))
			b.WriteString(", ")
			table.Add(value)
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n")
	return FormatFile([]byte(b.String()))
}
