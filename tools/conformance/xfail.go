package conformance

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	errMissingXFailField = errors.New("missing xfail field")
	errExpiredXFail      = errors.New("expired xfail")
	errDuplicateXFailID  = errors.New("duplicate xfail id")
	errUnknownXFailID    = errors.New("unknown xfail id")
)

type xfail struct {
	ID            string `json:"id"`
	Reason        string `json:"reason"`
	ExpiresAt     string `json:"expires_at"`
	TrackingIssue string `json:"tracking_issue"`
}

type xfailField string

const (
	xfailFieldID            xfailField = "id"
	xfailFieldReason        xfailField = "reason"
	xfailFieldExpiresAt     xfailField = "expires_at"
	xfailFieldTrackingIssue xfailField = "tracking_issue"
)

func validateXFailEntries(path string, fixtures []Fixture, entries []xfail, now time.Time) error {
	if len(entries) == 0 {
		return nil
	}
	fixtureIndex := fixturesByID(fixtures)
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.ID == "" {
			return fmt.Errorf("%s: missing required field %q: %w", path, xfailFieldID, errMissingXFailField)
		}
		if _, ok := seen[entry.ID]; ok {
			return fmt.Errorf("%s: duplicate fixture %q: %w", path, entry.ID, errDuplicateXFailID)
		}
		seen[entry.ID] = struct{}{}
		if _, ok := fixtureIndex[entry.ID]; !ok {
			return fmt.Errorf("%s: fixture %q does not match any fixture: %w", path, entry.ID, errUnknownXFailID)
		}
		if field := missingXFailField(entry); field != "" {
			return fmt.Errorf("%s: fixture %q missing required field %q: %w", path, entry.ID, field, errMissingXFailField)
		}
		expired, err := entry.expired(now)
		if err != nil {
			return fmt.Errorf("%s: fixture %q invalid %s %q: %w", path, entry.ID, xfailFieldExpiresAt, entry.ExpiresAt, err)
		}
		if expired {
			return fmt.Errorf("%s: fixture %q expired on %s: %w", path, entry.ID, entry.ExpiresAt, errExpiredXFail)
		}
	}
	return nil
}

func (x xfail) expired(now time.Time) (bool, error) {
	expiresAt, err := time.Parse(time.DateOnly, x.ExpiresAt)
	if err != nil {
		return false, err
	}
	return !expiresAt.After(now), nil
}

func missingXFailField(entry xfail) xfailField {
	switch {
	case entry.Reason == "":
		return xfailFieldReason
	case entry.ExpiresAt == "":
		return xfailFieldExpiresAt
	case entry.TrackingIssue == "":
		return xfailFieldTrackingIssue
	default:
		return ""
	}
}

func loadXFails(path string) ([]xfail, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []xfail
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return entries, nil
}
