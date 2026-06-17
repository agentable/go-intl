package conformance

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	errUnknownDivergence           = errors.New("unknown divergence id")
	errMissingPackageRoot          = errors.New("missing package root")
	errDuplicateDivergenceID       = errors.New("duplicate divergence id")
	errMissingDivergenceField      = errors.New("missing divergence field")
	errInvalidDivergenceStatus     = errors.New("invalid divergence status")
	errInvalidDivergenceReviewDate = errors.New("invalid divergence review date")
	errDivergenceSourceMismatch    = errors.New("divergence source mismatch")
	errMissingDivergenceWitness    = errors.New("missing divergence native witness")
	errUnknownDivergenceWitness    = errors.New("unknown divergence native witness")
	errInvalidDivergenceWitness    = errors.New("invalid divergence native witness")
)

func loadDivergenceIDs(path string) (map[string]struct{}, error) {
	entries, err := loadActiveDivergences(path)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]struct{}, len(entries))
	for id := range entries {
		ids[id] = struct{}{}
	}
	return ids, nil
}

func loadActiveDivergences(path string) (map[string]divergenceEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]divergenceEntry{}, nil
		}
		return nil, err
	}
	entries := map[string]divergenceEntry{}
	seen := map[string]struct{}{}
	current := divergenceEntry{}
	flush := func() error {
		if current.ID == "" {
			current = divergenceEntry{}
			return nil
		}
		if err := addActiveDivergence(path, entries, seen, current); err != nil {
			return err
		}
		current = divergenceEntry{}
		return nil
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch key {
		case "id":
			if err := flush(); err != nil {
				return nil, err
			}
			current.ID = value
		case "status":
			current.Status = value
		case "source":
			current.Source = value
		case "owner":
			current.Owner = value
		case "reason":
			current.Reason = value
		case "native_witness":
			current.NativeWitness = value
		case "review_after":
			current.ReviewAfter = value
		case "removal_path":
			current.RemovalPath = value
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return entries, nil
}

func addActiveDivergence(path string, entries map[string]divergenceEntry, seen map[string]struct{}, entry divergenceEntry) error {
	if _, ok := seen[entry.ID]; ok {
		return fmt.Errorf("%s: duplicate divergence id %q: %w", path, entry.ID, errDuplicateDivergenceID)
	}
	seen[entry.ID] = struct{}{}
	if entry.Status != "" && entry.Status != "accepted" && entry.Status != "resolved" {
		return fmt.Errorf("%s: divergence id %q status %q: %w", path, entry.ID, entry.Status, errInvalidDivergenceStatus)
	}
	if entry.Status == "resolved" {
		return nil
	}
	if field := missingDivergenceField(entry); field != "" {
		return fmt.Errorf("%s: divergence id %q missing %s: %w", path, entry.ID, field, errMissingDivergenceField)
	}
	if _, err := time.Parse(time.DateOnly, entry.ReviewAfter); err != nil {
		return fmt.Errorf("%s: divergence id %q review_after %q: %w", path, entry.ID, entry.ReviewAfter, errInvalidDivergenceReviewDate)
	}
	entries[entry.ID] = entry
	return nil
}

func missingDivergenceField(entry divergenceEntry) string {
	switch {
	case entry.Source == "":
		return "source"
	case entry.Owner == "":
		return "owner"
	case entry.Reason == "":
		return "reason"
	case entry.ReviewAfter == "":
		return "review_after"
	case entry.RemovalPath == "":
		return "removal_path"
	default:
		return ""
	}
}

type divergenceEntry struct {
	ID            string
	Source        string
	Owner         string
	Status        string
	Reason        string
	NativeWitness string
	ReviewAfter   string
	RemovalPath   string
}

func ValidateDivergences(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: %w", root, errMissingPackageRoot)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: %w", root, errMissingPackageRoot)
	}

	fixtures, err := LoadFixtures(root)
	if err != nil {
		return err
	}
	divergencePath := filepath.Join(root, "testdata", "divergences.md")
	fixturesByID := make(map[string]Fixture, len(fixtures))
	for _, fixture := range fixtures {
		fixturesByID[fixture.ID] = fixture
	}
	divergences, err := loadActiveDivergences(divergencePath)
	if err != nil {
		return err
	}
	for id, divergence := range divergences {
		fixture, ok := fixturesByID[id]
		if !ok {
			return fmt.Errorf("%s: divergence id %q does not match any fixture: %w", divergencePath, id, errUnknownDivergence)
		}
		if divergence.Source != fixture.Source {
			return fmt.Errorf("%s: divergence id %q source %q does not match fixture source %q: %w", divergencePath, id, divergence.Source, fixture.Source, errDivergenceSourceMismatch)
		}
		if err := validateDivergenceNativeWitness(divergencePath, divergence, fixturesByID); err != nil {
			return err
		}
	}
	return nil
}

func validateDivergenceNativeWitness(path string, divergence divergenceEntry, fixturesByID map[string]Fixture) error {
	if divergence.Owner != "datetimeformat" {
		return nil
	}
	if divergence.NativeWitness == "" {
		return fmt.Errorf("%s: divergence id %q missing native_witness: %w", path, divergence.ID, errMissingDivergenceWitness)
	}
	witness, ok := fixturesByID[divergence.NativeWitness]
	if !ok {
		return fmt.Errorf("%s: divergence id %q native_witness %q does not match any fixture: %w", path, divergence.ID, divergence.NativeWitness, errUnknownDivergenceWitness)
	}
	if fixtureSourceKind(witness.Source) != "node" {
		return fmt.Errorf("%s: divergence id %q native_witness %q source %q is not node: %w", path, divergence.ID, divergence.NativeWitness, witness.Source, errInvalidDivergenceWitness)
	}
	if !fixtureHasNativeExpectation(witness) {
		return fmt.Errorf("%s: divergence id %q native_witness %q has no observable expectation: %w", path, divergence.ID, divergence.NativeWitness, errInvalidDivergenceWitness)
	}
	return nil
}

func fixtureHasNativeExpectation(fixture Fixture) bool {
	return fixture.Expected != nil ||
		fixture.ExpectedOK != nil ||
		len(fixture.ExpectedLocales) > 0 ||
		len(fixture.ExpectedParts) > 0 ||
		fixture.ExpectedRange != nil ||
		len(fixture.ExpectedRangeParts) > 0 ||
		fixture.ExpectedComparison != nil ||
		len(fixture.ExpectedResolved) > 0 ||
		len(fixture.ExpectedSegments) > 0 ||
		fixture.ErrorCode != ""
}
