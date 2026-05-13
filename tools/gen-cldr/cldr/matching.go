package cldr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type LanguageMatching struct {
	ParadigmLocales []string
	MatchVariables  map[string][]string
	Matches         []LanguageMatch
}

type LanguageMatch struct {
	Desired   string
	Supported string
	Distance  int
	Oneway    bool
}

type flexibleBool bool

func (b *flexibleBool) UnmarshalJSON(data []byte) error {
	var value bool
	if err := json.Unmarshal(data, &value); err == nil {
		*b = flexibleBool(value)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*b = flexibleBool(s == "true")
	return nil
}

type flexibleInt int

func (i *flexibleInt) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*i = flexibleInt(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*i = flexibleInt(n)
	return nil
}

type stringList []string

func (l *stringList) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*l = strings.Fields(single)
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*l = many
	return nil
}

func loadLanguageMatching(root string) (LanguageMatching, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "languageMatching.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return LanguageMatching{}, nil
	}
	var doc struct {
		Supplemental struct {
			LanguageMatching map[string]struct {
				ParadigmLocales struct {
					Locales stringList `json:"_locales"`
				} `json:"paradigmLocales"`
				MatchVariables map[string]struct {
					Value string `json:"_value"`
				} `json:"matchVariables"`
				MatchVariable []struct {
					ID    string `json:"_id"`
					Value string `json:"_value"`
				} `json:"matchVariable"`
				LanguageMatch []struct {
					Desired   string       `json:"_desired"`
					Supported string       `json:"_supported"`
					Distance  flexibleInt  `json:"_distance"`
					Oneway    flexibleBool `json:"_oneway"`
				} `json:"languageMatch"`
			} `json:"languageMatching"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return LanguageMatching{}, fmt.Errorf("parse %s: %w", path, err)
	}
	written := doc.Supplemental.LanguageMatching["written-new"]
	if len(written.LanguageMatch) == 0 {
		written = doc.Supplemental.LanguageMatching["written_new"]
	}
	out := LanguageMatching{
		ParadigmLocales: []string(written.ParadigmLocales.Locales),
		MatchVariables:  make(map[string][]string),
	}
	for id, v := range written.MatchVariables {
		out.MatchVariables[id] = splitMatchVariable(v.Value)
	}
	for _, v := range written.MatchVariable {
		out.MatchVariables[v.ID] = splitMatchVariable(v.Value)
	}
	for _, m := range written.LanguageMatch {
		out.Matches = append(out.Matches, LanguageMatch{
			Desired:   m.Desired,
			Supported: m.Supported,
			Distance:  int(m.Distance),
			Oneway:    bool(m.Oneway),
		})
	}
	return out, nil
}

func splitMatchVariable(value string) []string {
	return strings.Fields(strings.ReplaceAll(value, "+", " "))
}

func loadRegions(root string) (map[string][]string, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "territoryContainment.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	var doc struct {
		Supplemental struct {
			TerritoryContainment map[string]struct {
				Contains stringList `json:"_contains"`
			} `json:"territoryContainment"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string][]string, len(doc.Supplemental.TerritoryContainment))
	for region, data := range doc.Supplemental.TerritoryContainment {
		out[region] = []string(data.Contains)
	}
	return out, nil
}
