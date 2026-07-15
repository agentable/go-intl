package conformance

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	errUnknownDivergence           = errors.New("unknown divergence id")
	errMissingPackageRoot          = errors.New("missing package root")
	errDuplicateDivergenceID       = errors.New("duplicate divergence id")
	errDuplicateDivergenceField    = errors.New("duplicate divergence field")
	errUnknownDivergenceField      = errors.New("unknown divergence field")
	errMalformedDivergenceLine     = errors.New("malformed divergence line")
	errMissingDivergenceField      = errors.New("missing divergence field")
	errInvalidDivergenceStatus     = errors.New("invalid divergence status")
	errInvalidDivergenceReviewDate = errors.New("invalid divergence review date")
	errDivergenceSourceMismatch    = errors.New("divergence source mismatch")
	errMissingDivergenceWitness    = errors.New("missing divergence native witness")
	errUnknownDivergenceWitness    = errors.New("unknown divergence native witness")
	errInvalidDivergenceWitness    = errors.New("invalid divergence native witness")
)

type missingDivergenceFieldError struct {
	Path  string
	ID    string
	Field divergenceField
}

func (e *missingDivergenceFieldError) Error() string {
	return fmt.Sprintf("%s: divergence id %q missing %s: %v", e.Path, e.ID, e.Field, errMissingDivergenceField)
}

func (e *missingDivergenceFieldError) Unwrap() error {
	return errMissingDivergenceField
}

type divergenceField string

const (
	divergenceFieldID            divergenceField = "id"
	divergenceFieldStatus        divergenceField = "status"
	divergenceFieldSource        divergenceField = "source"
	divergenceFieldOwner         divergenceField = "owner"
	divergenceFieldReason        divergenceField = "reason"
	divergenceFieldNativeWitness divergenceField = "native_witness"
	divergenceFieldReviewAfter   divergenceField = "review_after"
	divergenceFieldRemovalPath   divergenceField = "removal_path"
)

type divergenceStatus string

const (
	divergenceStatusAccepted divergenceStatus = "accepted"
	divergenceStatusResolved divergenceStatus = "resolved"
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
	currentFields := map[divergenceField]struct{}{}
	started := false
	flush := func() error {
		if !started {
			current = divergenceEntry{}
			return nil
		}
		if err := addActiveDivergence(path, entries, seen, current); err != nil {
			return err
		}
		current = divergenceEntry{}
		currentFields = map[divergenceField]struct{}{}
		started = false
		return nil
	}
	lineNumber := 0
	for line := range strings.SplitSeq(string(data), "\n") {
		lineNumber++
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("%s:%d: %q: %w", path, lineNumber, line, errMalformedDivergenceLine)
		}
		field := divergenceField(strings.TrimSpace(key))
		if !knownDivergenceField(field) {
			return nil, fmt.Errorf("%s:%d: field %q: %w", path, lineNumber, key, errUnknownDivergenceField)
		}
		value = strings.TrimSpace(value)
		if field == divergenceFieldID && started {
			if err := flush(); err != nil {
				return nil, err
			}
		}
		if _, ok := currentFields[field]; ok {
			return nil, fmt.Errorf("%s:%d: divergence id %q field %q: %w", path, lineNumber, current.ID, field, errDuplicateDivergenceField)
		}
		if field != divergenceFieldID && !started {
			return nil, &missingDivergenceFieldError{Path: path, Field: divergenceFieldID}
		}
		started = true
		currentFields[field] = struct{}{}
		switch field {
		case divergenceFieldID:
			current.ID = value
		case divergenceFieldStatus:
			current.Status = divergenceStatus(value)
		case divergenceFieldSource:
			current.Source = value
		case divergenceFieldOwner:
			current.Owner = value
		case divergenceFieldReason:
			current.Reason = value
		case divergenceFieldNativeWitness:
			current.NativeWitness = value
		case divergenceFieldReviewAfter:
			current.ReviewAfter = value
		case divergenceFieldRemovalPath:
			current.RemovalPath = value
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return entries, nil
}

func knownDivergenceField(field divergenceField) bool {
	switch field {
	case divergenceFieldID,
		divergenceFieldStatus,
		divergenceFieldSource,
		divergenceFieldOwner,
		divergenceFieldReason,
		divergenceFieldNativeWitness,
		divergenceFieldReviewAfter,
		divergenceFieldRemovalPath:
		return true
	default:
		return false
	}
}

func addActiveDivergence(path string, entries map[string]divergenceEntry, seen map[string]struct{}, entry divergenceEntry) error {
	if _, ok := seen[entry.ID]; ok {
		return fmt.Errorf("%s: duplicate divergence id %q: %w", path, entry.ID, errDuplicateDivergenceID)
	}
	seen[entry.ID] = struct{}{}
	if field := missingDivergenceField(entry); field != "" {
		return &missingDivergenceFieldError{Path: path, ID: entry.ID, Field: field}
	}
	if entry.Status != divergenceStatusAccepted && entry.Status != divergenceStatusResolved {
		return fmt.Errorf("%s: divergence id %q status %q: %w", path, entry.ID, entry.Status, errInvalidDivergenceStatus)
	}
	if _, err := time.Parse(time.DateOnly, entry.ReviewAfter); err != nil {
		return fmt.Errorf("%s: divergence id %q review_after %q: %w", path, entry.ID, entry.ReviewAfter, errInvalidDivergenceReviewDate)
	}
	if entry.Status == divergenceStatusResolved {
		return nil
	}
	entries[entry.ID] = entry
	return nil
}

func missingDivergenceField(entry divergenceEntry) divergenceField {
	switch {
	case entry.ID == "":
		return divergenceFieldID
	case entry.Status == "":
		return divergenceFieldStatus
	case entry.Source == "":
		return divergenceFieldSource
	case entry.Owner == "":
		return divergenceFieldOwner
	case entry.Reason == "":
		return divergenceFieldReason
	case entry.ReviewAfter == "":
		return divergenceFieldReviewAfter
	case entry.RemovalPath == "":
		return divergenceFieldRemovalPath
	default:
		return ""
	}
}

type divergenceEntry struct {
	ID            string
	Source        string
	Owner         string
	Status        divergenceStatus
	Reason        string
	NativeWitness string
	ReviewAfter   string
	RemovalPath   string
}

func ValidateDivergences(root string) error {
	if err := validatePackageRoot(root); err != nil {
		return err
	}

	fixtures, err := LoadFixtures(root)
	if err != nil {
		return err
	}
	divergencePath := divergenceLedgerPath(root)
	divergences, err := loadActiveDivergences(divergencePath)
	if err != nil {
		return err
	}
	return validateDivergenceEntries(divergencePath, fixtures, divergences)
}

func validatePackageRoot(root string) error {
	// The conformance command intentionally validates a maintainer-selected package root.
	info, err := os.Stat(root) // #nosec G703 -- inspecting the selected package root is the API contract.
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: %w", root, errMissingPackageRoot)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: %w", root, errMissingPackageRoot)
	}
	return nil
}

func validateDivergenceEntries(path string, fixtures []Fixture, divergences map[string]divergenceEntry) error {
	fixtureIndex := fixturesByID(fixtures)
	for id, divergence := range divergences {
		fixture, ok := fixtureIndex[id]
		if !ok {
			return fmt.Errorf("%s: divergence id %q does not match any fixture: %w", path, id, errUnknownDivergence)
		}
		if divergence.Source != fixture.Source {
			return fmt.Errorf("%s: divergence id %q source %q does not match fixture source %q: %w", path, id, divergence.Source, fixture.Source, errDivergenceSourceMismatch)
		}
		if err := validateDivergenceNativeWitness(path, divergence, fixtureIndex); err != nil {
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
	witness, witnessStatus := classifyNativeWitness(fixturesByID, divergence.NativeWitness)
	switch witnessStatus {
	case nativeWitnessUnknown:
		return fmt.Errorf("%s: divergence id %q native_witness %q does not match any fixture: %w", path, divergence.ID, divergence.NativeWitness, errUnknownDivergenceWitness)
	case nativeWitnessNotNode:
		return fmt.Errorf("%s: divergence id %q native_witness %q source %q is not node: %w", path, divergence.ID, divergence.NativeWitness, witness.Source, errInvalidDivergenceWitness)
	case nativeWitnessNoExpectation:
		return fmt.Errorf("%s: divergence id %q native_witness %q has no observable expectation: %w", path, divergence.ID, divergence.NativeWitness, errInvalidDivergenceWitness)
	case nativeWitnessValid:
		return nil
	}
	return nil
}
