package conformance

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	errMixedFixtureSources  = errors.New("mixed fixture sources")
	errMissingFixtureField  = errors.New("missing fixture field")
	errInvalidDateTimeInput = errors.New("invalid datetime input")
	errDuplicateFixtureID   = errors.New("duplicate fixture id")
)

type Fixture struct {
	ID                 string          `json:"id"`
	Source             string          `json:"source"`
	Locale             string          `json:"locale"`
	Feature            string          `json:"feature,omitempty"`
	Options            json.RawMessage `json:"options"`
	Input              json.RawMessage `json:"input"`
	Expected           *string         `json:"expected,omitempty"`
	ExpectedParts      []Part          `json:"expectedParts,omitempty"`
	ExpectedRange      *string         `json:"expectedRange,omitempty"`
	ExpectedRangeParts []RangePart     `json:"expectedRangeParts,omitempty"`
	ErrorCode          string          `json:"errorCode,omitempty"`
}

type Part struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type RangePart struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

func LoadFixtures(root string) ([]Fixture, error) {
	var fixtures []Fixture
	base := filepath.Join(root, "testdata", "conformance")
	if _, err := os.Stat(base); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(filepath.Join(base, filepath.Clean(rel)))
		if err != nil {
			return err
		}
		var fileFixtures []Fixture
		if err := json.Unmarshal(data, &fileFixtures); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		fileSource := ""
		for _, fixture := range fileFixtures {
			if err := validateFixtureShape(path, fixture); err != nil {
				return err
			}
			if strings.HasSuffix(filepath.Clean(root), "datetimeformat") {
				if err := validateDateTimeInput(path, fixture); err != nil {
					return err
				}
			}
			if fileSource == "" {
				fileSource = fixture.Source
				continue
			}
			if fixture.Source != fileSource {
				return fmt.Errorf("%s: mixed sources %q and %q: %w", path, fileSource, fixture.Source, errMixedFixtureSources)
			}
		}
		fixtures = append(fixtures, fileFixtures...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fixtures, nil
}

func validateFixtureShape(path string, fixture Fixture) error {
	checks := []struct {
		name  string
		empty bool
	}{
		{name: "id", empty: fixture.ID == ""},
		{name: "source", empty: fixture.Source == ""},
		{name: "locale", empty: fixture.Locale == ""},
		{name: "options", empty: fixture.Options == nil},
		{name: "input", empty: fixture.Input == nil},
	}
	for _, check := range checks {
		if check.empty {
			return fmt.Errorf("%s: missing required field %q: %w", path, check.name, errMissingFixtureField)
		}
	}
	return nil
}

func validateDateTimeInput(path string, fixture Fixture) error {
	var value any
	if err := json.Unmarshal(fixture.Input, &value); err != nil {
		return fmt.Errorf("%s: fixture %q input: %w", path, fixture.ID, err)
	}
	if input, ok := value.(string); ok {
		if _, err := time.Parse(time.RFC3339, input); err == nil {
			return nil
		}
	}
	if input, ok := value.(map[string]any); ok {
		start, hasStart := input["start"].(string)
		end, hasEnd := input["end"].(string)
		if hasStart && hasEnd {
			if _, err := time.Parse(time.RFC3339, start); err == nil {
				if _, err := time.Parse(time.RFC3339, end); err == nil {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("%s: fixture %q datetime input must be an ISO-8601 string: %w", path, fixture.ID, errInvalidDateTimeInput)
}

func ValidateFixtures(root string, now time.Time) error {
	fixtures, err := LoadFixtures(root)
	if err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, fixture := range fixtures {
		if _, ok := seen[fixture.ID]; ok {
			return fmt.Errorf("duplicate fixture id %q: %w", fixture.ID, errDuplicateFixtureID)
		}
		seen[fixture.ID] = struct{}{}
	}
	return ValidateXFailFile(root, now)
}

func ValidateFixtureRoots(roots []string, now time.Time) error {
	seen := map[string]string{}
	for _, root := range roots {
		fixtures, err := LoadFixtures(root)
		if err != nil {
			return err
		}
		for _, fixture := range fixtures {
			if previous := seen[fixture.ID]; previous != "" {
				return fmt.Errorf("duplicate fixture id %q in %s and %s: %w", fixture.ID, previous, root, errDuplicateFixtureID)
			}
			seen[fixture.ID] = root
		}
		if err := ValidateXFailFile(root, now); err != nil {
			return err
		}
	}
	return nil
}
