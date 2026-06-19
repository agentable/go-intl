package conformance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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
	errInvalidErrorFixture  = errors.New("invalid error fixture file")
	errFixtureSourceDir     = errors.New("fixture source directory mismatch")
)

type Fixture struct {
	ID                 string          `json:"id"`
	Source             string          `json:"source"`
	Locale             string          `json:"locale"`
	Feature            string          `json:"feature,omitempty"`
	Options            json.RawMessage `json:"options"`
	Input              json.RawMessage `json:"input"`
	Expected           *string         `json:"expected,omitempty"`
	ExpectedOK         *bool           `json:"expectedOk,omitempty"`
	ExpectedLocales    []string        `json:"expectedLocales,omitempty"`
	ExpectedParts      []Part          `json:"expectedParts,omitempty"`
	ExpectedRange      *string         `json:"expectedRange,omitempty"`
	ExpectedRangeParts []RangePart     `json:"expectedRangeParts,omitempty"`
	ExpectedComparison *int            `json:"expectedComparison,omitempty"`
	ExpectedResolved   json.RawMessage `json:"expectedResolvedOptions,omitempty"`
	ExpectedSegments   []SegmentRecord `json:"expectedSegments,omitempty"`
	ErrorCode          string          `json:"errorCode,omitempty"`
}

type Part struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

type RangePart struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

type SegmentRecord struct {
	Segment       string `json:"segment"`
	CodeUnitIndex int    `json:"codeUnitIndex"`
	ByteIndex     *int   `json:"byteIndex,omitempty"`
	IsWordLike    *bool  `json:"isWordLike,omitempty"`
}

func LoadFixtures(root string) ([]Fixture, error) {
	var fixtures []Fixture
	base := filepath.Join(root, "testdata", "conformance")
	validateDateTimeInputs := strings.HasSuffix(filepath.Clean(root), "datetimeformat")
	if _, err := os.Stat(base); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	fixtureFS := os.DirFS(base)
	err := fs.WalkDir(fixtureFS, ".", func(rel string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(rel) != ".json" {
			return nil
		}
		path := filepath.Join(base, rel)
		data, err := fs.ReadFile(fixtureFS, rel)
		if err != nil {
			return err
		}
		var fileFixtures []Fixture
		if err := json.Unmarshal(data, &fileFixtures); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := validateFixtureFile(path, rel, fileFixtures, validateDateTimeInputs); err != nil {
			return err
		}
		fixtures = append(fixtures, fileFixtures...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fixtures, nil
}

func validateFixtureFile(path, rel string, fixtures []Fixture, validateDateTimeInputs bool) error {
	fileSource := ""
	for _, fixture := range fixtures {
		if err := validateFixture(path, rel, fixture, validateDateTimeInputs); err != nil {
			return err
		}
		if fileSource == "" {
			fileSource = fixture.Source
			continue
		}
		if fixture.Source != fileSource {
			return fmt.Errorf("%s: mixed sources %q and %q: %w", path, fileSource, fixture.Source, errMixedFixtureSources)
		}
	}
	return nil
}

func validateFixture(path, rel string, fixture Fixture, validateDateTimeInputs bool) error {
	if err := validateFixtureShape(path, fixture); err != nil {
		return err
	}
	if err := validateFixtureSourceDirectory(path, rel, fixture); err != nil {
		return err
	}
	if err := validateErrorFixtureFile(path, fixture); err != nil {
		return err
	}
	if !validateDateTimeInputs {
		return nil
	}
	return validateDateTimeInput(path, fixture)
}

func validateFixtureShape(path string, fixture Fixture) error {
	if field := missingFixtureField(fixture); field != "" {
		return fmt.Errorf("%s: missing required field %q: %w", path, field, errMissingFixtureField)
	}
	return nil
}

func missingFixtureField(fixture Fixture) string {
	switch {
	case fixture.ID == "":
		return "id"
	case fixture.Source == "":
		return "source"
	case fixture.Locale == "":
		return "locale"
	case fixture.Options == nil:
		return "options"
	case fixture.Input == nil:
		return "input"
	default:
		return ""
	}
}

func validateFixtureSourceDirectory(path, rel string, fixture Fixture) error {
	sourceDir, _, _ := strings.Cut(filepath.ToSlash(rel), "/")
	if fixtureSourceMatchesDirectory(sourceDir, fixture) {
		return nil
	}
	return fmt.Errorf("%s: fixture %q source %q does not match conformance source directory %q: %w", path, fixture.ID, fixture.Source, sourceDir, errFixtureSourceDir)
}

func fixtureSourceMatchesDirectory(sourceDir string, fixture Fixture) bool {
	kind := fixtureSourceKind(fixture.Source)
	switch {
	case sourceDir == "manual":
		return kind == "manual"
	case sourceDir == "formatjs":
		return kind == "formatjs"
	case strings.HasPrefix(sourceDir, "node-v"):
		return kind == "node" && nodeFixtureMatchesDirectory(sourceDir, fixture)
	case sourceDir == "node":
		return kind == "node"
	default:
		return true
	}
}

func nodeFixtureMatchesDirectory(sourceDir string, fixture Fixture) bool {
	version := strings.TrimPrefix(sourceDir, "node-v")
	sourcePrefix := "node:v" + version
	return strings.Contains(fixture.ID, "-"+sourceDir+"-") &&
		(strings.HasPrefix(fixture.Source, sourcePrefix+".") || strings.HasPrefix(fixture.Source, sourcePrefix+":"))
}

func fixtureSourceKind(source string) string {
	switch {
	case source == "manual" || strings.HasPrefix(source, "manual:"):
		return "manual"
	case strings.HasPrefix(source, "formatjs:"):
		return "formatjs"
	case strings.HasPrefix(source, "node:"):
		return "node"
	default:
		return ""
	}
}

func validateErrorFixtureFile(path string, fixture Fixture) error {
	inErrorsFile := filepath.Base(path) == "errors.json"
	if fixture.ErrorCode != "" && !inErrorsFile {
		return fmt.Errorf("%s: fixture %q has errorCode outside errors.json: %w", path, fixture.ID, errInvalidErrorFixture)
	}
	if inErrorsFile && fixture.ErrorCode == "" {
		return fmt.Errorf("%s: fixture %q in errors.json missing errorCode: %w", path, fixture.ID, errInvalidErrorFixture)
	}
	return nil
}

func validateDateTimeInput(path string, fixture Fixture) error {
	var value any
	if err := json.Unmarshal(fixture.Input, &value); err != nil {
		return fmt.Errorf("%s: fixture %q input: %w", path, fixture.ID, err)
	}
	switch input := value.(type) {
	case string:
		if isRFC3339String(input) {
			return nil
		}
	case map[string]any:
		if isRFC3339Range(input) {
			return nil
		}
	}
	return fmt.Errorf("%s: fixture %q datetime input must be an ISO-8601 string: %w", path, fixture.ID, errInvalidDateTimeInput)
}

func isRFC3339String(value string) bool {
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

func isRFC3339Range(input map[string]any) bool {
	start, hasStart := input["start"].(string)
	end, hasEnd := input["end"].(string)
	return hasStart && hasEnd && isRFC3339String(start) && isRFC3339String(end)
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
		if err := validateXFailFile(root, fixtures, now); err != nil {
			return err
		}
	}
	return nil
}
