package conformance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	ExpectedResolved   json.RawMessage `json:"expectedResolvedOptions,omitempty"`
	ErrorCode          string          `json:"errorCode,omitempty"`
}

// FeatureSupportedLocalesOf is the conformance fixture feature for
// Intl.<Constructor>.supportedLocalesOf behavior.
const FeatureSupportedLocalesOf = "supportedLocalesOf"

// IsSupportedLocalesOf reports whether the fixture exercises supportedLocalesOf.
func (f Fixture) IsSupportedLocalesOf() bool {
	return f.Feature == FeatureSupportedLocalesOf
}

// RequiredExpected returns the fixture's expected string output.
func (f Fixture) RequiredExpected(t testing.TB) string {
	t.Helper()

	if f.Expected == nil {
		t.Fatal("fixture expected is required")
	}
	return *f.Expected
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

type fixtureField string

const (
	fixtureFieldID      fixtureField = "id"
	fixtureFieldSource  fixtureField = "source"
	fixtureFieldLocale  fixtureField = "locale"
	fixtureFieldOptions fixtureField = "options"
	fixtureFieldInput   fixtureField = "input"
)

type fixtureSourceKind string

const (
	fixtureSourceUnknown  fixtureSourceKind = ""
	fixtureSourceManual   fixtureSourceKind = "manual"
	fixtureSourceFormatJS fixtureSourceKind = "formatjs"
	fixtureSourceNode     fixtureSourceKind = "node"
)

const (
	nodeFixtureDirPrefix    = "node-v"
	nodeFixtureSourcePrefix = "node:v"
)

func LoadFixtures(root string) ([]Fixture, error) {
	var fixtures []Fixture
	base := conformanceFixturesPath(root)
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

func fixturesByID(fixtures []Fixture) map[string]Fixture {
	byID := make(map[string]Fixture, len(fixtures))
	for _, fixture := range fixtures {
		byID[fixture.ID] = fixture
	}
	return byID
}

func validateFixtureFile(path, rel string, fixtures []Fixture, validateDateTimeInputs bool) error {
	fileSource := ""
	inErrorsFile := filepath.Base(path) == "errors.json"
	for _, fixture := range fixtures {
		if err := validateFixture(path, rel, fixture, inErrorsFile, validateDateTimeInputs); err != nil {
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

func validateFixture(path, rel string, fixture Fixture, inErrorsFile bool, validateDateTimeInputs bool) error {
	if err := validateFixtureShape(path, fixture); err != nil {
		return err
	}
	if err := validateFixtureSourceDirectory(path, rel, fixture); err != nil {
		return err
	}
	if err := validateErrorFixtureFile(path, fixture, inErrorsFile); err != nil {
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

func missingFixtureField(fixture Fixture) fixtureField {
	switch {
	case fixture.ID == "":
		return fixtureFieldID
	case fixture.Source == "":
		return fixtureFieldSource
	case fixture.Locale == "":
		return fixtureFieldLocale
	case fixture.Options == nil:
		return fixtureFieldOptions
	case fixture.Input == nil:
		return fixtureFieldInput
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
	kind := fixtureSourceKindOf(fixture.Source)
	switch {
	case sourceDir == string(fixtureSourceManual):
		return kind == fixtureSourceManual
	case sourceDir == string(fixtureSourceFormatJS):
		return kind == fixtureSourceFormatJS
	case strings.HasPrefix(sourceDir, nodeFixtureDirPrefix):
		return kind == fixtureSourceNode && nodeFixtureMatchesDirectory(sourceDir, fixture)
	case sourceDir == string(fixtureSourceNode):
		return kind == fixtureSourceNode
	default:
		return true
	}
}

func nodeFixtureMatchesDirectory(sourceDir string, fixture Fixture) bool {
	version := strings.TrimPrefix(sourceDir, nodeFixtureDirPrefix)
	sourcePrefix := nodeFixtureSourcePrefix + version
	return strings.Contains(fixture.ID, "-"+sourceDir+"-") &&
		(strings.HasPrefix(fixture.Source, sourcePrefix+".") || strings.HasPrefix(fixture.Source, sourcePrefix+":"))
}

func fixtureSourceKindOf(source string) fixtureSourceKind {
	switch {
	case source == string(fixtureSourceManual) || strings.HasPrefix(source, string(fixtureSourceManual)+":"):
		return fixtureSourceManual
	case strings.HasPrefix(source, string(fixtureSourceFormatJS)+":"):
		return fixtureSourceFormatJS
	case strings.HasPrefix(source, string(fixtureSourceNode)+":"):
		return fixtureSourceNode
	default:
		return fixtureSourceUnknown
	}
}

func fixtureHasNativeExpectation(fixture Fixture) bool {
	return fixture.Expected != nil ||
		fixture.ExpectedOK != nil ||
		len(fixture.ExpectedLocales) > 0 ||
		len(fixture.ExpectedParts) > 0 ||
		fixture.ExpectedRange != nil ||
		len(fixture.ExpectedRangeParts) > 0 ||
		len(fixture.ExpectedResolved) > 0 ||
		fixture.ErrorCode != ""
}

type nativeWitnessStatus string

const (
	nativeWitnessValid         nativeWitnessStatus = ""
	nativeWitnessUnknown       nativeWitnessStatus = "unknown"
	nativeWitnessNotNode       nativeWitnessStatus = "not-node"
	nativeWitnessNoExpectation nativeWitnessStatus = "no-expectation"
)

func classifyNativeWitness(fixturesByID map[string]Fixture, id string) (Fixture, nativeWitnessStatus) {
	witness, ok := fixturesByID[id]
	if !ok {
		return Fixture{}, nativeWitnessUnknown
	}
	if fixtureSourceKindOf(witness.Source) != fixtureSourceNode {
		return witness, nativeWitnessNotNode
	}
	if !fixtureHasNativeExpectation(witness) {
		return witness, nativeWitnessNoExpectation
	}
	return witness, nativeWitnessValid
}

func validateErrorFixtureFile(path string, fixture Fixture, inErrorsFile bool) error {
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
		suite, err := loadRunSuite(root, now)
		if err != nil {
			return err
		}
		for _, fixture := range suite.fixtures {
			if previous := seen[fixture.ID]; previous != "" {
				return fmt.Errorf("duplicate fixture id %q in %s and %s: %w", fixture.ID, previous, root, errDuplicateFixtureID)
			}
			seen[fixture.ID] = root
		}
	}
	return nil
}
