package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func RenderLanguageMatchingProfile(path string, profile cldr.LanguageMatching) error {
	src, err := renderLanguageMatchingProfile(profile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, src, 0o666); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func renderLanguageMatchingProfile(profile cldr.LanguageMatching) ([]byte, error) {
	var body bytes.Buffer
	if _, err := body.WriteString("package localematcher\n\nvar generatedLanguageMatchingProfile = languageMatchingProfile{\n"); err != nil {
		return nil, err
	}
	if _, err := body.WriteString("\tparadigmLocales: []string{\n"); err != nil {
		return nil, err
	}
	for _, locale := range profile.ParadigmLocales {
		if _, err := fmt.Fprintf(&body, "\t\t%q,\n", locale); err != nil {
			return nil, err
		}
	}
	if _, err := body.WriteString("\t},\n\tmatchVariables: []languageMatchVariable{\n"); err != nil {
		return nil, err
	}
	for _, variable := range profile.MatchVariables {
		if _, err := fmt.Fprintf(&body, "\t\t{name: %q, sourceRegions: %#v, expandedRegions: %#v},\n", variable.Name, variable.SourceRegions, variable.ExpandedRegions); err != nil {
			return nil, err
		}
	}
	if _, err := body.WriteString("\t},\n\trules: []languageMatchRule{\n"); err != nil {
		return nil, err
	}
	for _, rule := range profile.Rules {
		if _, err := fmt.Fprintf(&body, "\t\t{desired: %q, supported: %q, distance: %d, oneWay: %t},\n", rule.Desired, rule.Supported, rule.Distance, rule.OneWay); err != nil {
			return nil, err
		}
	}
	if _, err := body.WriteString("\t},\n}\n"); err != nil {
		return nil, err
	}
	return FormatFile(body.Bytes())
}
