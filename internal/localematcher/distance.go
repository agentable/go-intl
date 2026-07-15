package localematcher

import (
	"math"
	"strings"
)

type languageScriptRegion struct {
	language string
	script   string
	region   string
}

type languageMatchPattern struct {
	language       string
	script         string
	region         string
	regionVariable string
	excludeRegion  bool
}

type compiledLanguageMatchRule struct {
	desired   languageMatchPattern
	supported languageMatchPattern
	distance  int
	oneWay    bool
}

type compiledLanguageMatchingProfile struct {
	paradigms map[string]bool
	variables map[string]map[string]bool
	rules     []compiledLanguageMatchRule
}

func compileLanguageMatchingProfile(profile languageMatchingProfile, maximizer Maximizer) compiledLanguageMatchingProfile {
	compiled := compiledLanguageMatchingProfile{
		paradigms: make(map[string]bool, len(profile.paradigmLocales)*2),
		variables: make(map[string]map[string]bool, len(profile.matchVariables)),
		rules:     make([]compiledLanguageMatchRule, len(profile.rules)),
	}
	for _, locale := range profile.paradigmLocales {
		compiled.paradigms[locale] = true
		compiled.paradigms[maximizer(locale)] = true
	}
	for _, variable := range profile.matchVariables {
		regions := make(map[string]bool, len(variable.expandedRegions))
		for _, region := range variable.expandedRegions {
			regions[region] = true
		}
		compiled.variables[variable.name] = regions
	}
	for i, rule := range profile.rules {
		compiled.rules[i] = compiledLanguageMatchRule{
			desired:   parseLanguageMatchPattern(rule.desired),
			supported: parseLanguageMatchPattern(rule.supported),
			distance:  rule.distance * 10,
			oneWay:    rule.oneWay,
		}
	}
	return compiled
}

func (p compiledLanguageMatchingProfile) distance(maximizedDesired, maximizedSupported string) int {
	desired := parseLanguageScriptRegion(maximizedDesired)
	supported := parseLanguageScriptRegion(maximizedSupported)
	if desired == supported {
		return 0
	}
	distance := 0
	if desired.language != supported.language {
		distance += p.componentDistance(
			languageScriptRegion{language: desired.language},
			languageScriptRegion{language: supported.language},
		)
	}
	if desired.script != supported.script {
		distance += p.componentDistance(
			languageScriptRegion{language: desired.language, script: desired.script},
			languageScriptRegion{language: supported.language, script: supported.script},
		)
	}
	if desired.region != supported.region {
		distance += p.componentDistance(desired, supported)
	}
	return distance
}

func (p compiledLanguageMatchingProfile) componentDistance(desired, supported languageScriptRegion) int {
	for _, rule := range p.rules {
		matched := p.matches(desired, rule.desired) && p.matches(supported, rule.supported)
		if !rule.oneWay && !matched {
			matched = p.matches(desired, rule.supported) && p.matches(supported, rule.desired)
		}
		if !matched {
			continue
		}
		distance := rule.distance
		if p.paradigms[desired.String()] != p.paradigms[supported.String()] {
			distance--
		}
		return distance
	}
	return math.MaxInt / 4
}

func (p compiledLanguageMatchingProfile) matches(locale languageScriptRegion, pattern languageMatchPattern) bool {
	if locale.language != "" && pattern.language != "*" && pattern.language != locale.language {
		return false
	}
	if locale.script != "" && pattern.script != "*" && pattern.script != locale.script {
		return false
	}
	if locale.region == "" {
		return true
	}
	if pattern.regionVariable != "" {
		contains := p.variables[pattern.regionVariable][locale.region]
		if pattern.excludeRegion {
			return !contains
		}
		return contains
	}
	return pattern.region == "*" || pattern.region == locale.region
}

func parseLanguageMatchPattern(pattern string) languageMatchPattern {
	parts := strings.Split(pattern, "-")
	out := languageMatchPattern{language: parts[0]}
	if len(parts) >= 2 {
		out.script = parts[1]
	}
	if len(parts) == 3 {
		out.region = parts[2]
		if variable, ok := strings.CutPrefix(out.region, "$"); ok {
			out.excludeRegion = strings.HasPrefix(variable, "!")
			out.regionVariable = strings.TrimPrefix(variable, "!")
		}
	}
	return out
}

func parseLanguageScriptRegion(locale string) languageScriptRegion {
	parts := strings.Split(locale, "-")
	if len(parts) == 0 {
		return languageScriptRegion{}
	}
	out := languageScriptRegion{language: parts[0]}
	for _, part := range parts[1:] {
		switch {
		case len(part) == 4:
			out.script = part
		case len(part) == 2 || len(part) == 3:
			out.region = part
		}
	}
	return out
}

func (l languageScriptRegion) String() string {
	parts := [3]string{l.language, l.script, l.region}
	var out strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		if out.Len() > 0 {
			out.WriteByte('-')
		}
		out.WriteString(part)
	}
	return out.String()
}
