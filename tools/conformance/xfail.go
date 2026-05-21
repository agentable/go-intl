package conformance

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

func validateXFailFile(root string, fixtures []Fixture, now time.Time) error {
	path := filepath.Join(root, "testdata", "xfail.json")
	entries, err := loadXFails(path)
	if err != nil {
		return err
	}
	return validateXFailEntries(path, fixtures, entries, now)
}

func validateXFailEntries(path string, fixtures []Fixture, entries []xfail, now time.Time) error {
	if len(entries) == 0 {
		return nil
	}
	fixtureIDs := make(map[string]struct{}, len(fixtures))
	for _, fixture := range fixtures {
		fixtureIDs[fixture.ID] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.ID == "" {
			return fmt.Errorf("%s: missing required field %q: %w", path, "id", errMissingXFailField)
		}
		if _, ok := seen[entry.ID]; ok {
			return fmt.Errorf("%s: duplicate fixture %q: %w", path, entry.ID, errDuplicateXFailID)
		}
		seen[entry.ID] = struct{}{}
		if _, ok := fixtureIDs[entry.ID]; !ok {
			return fmt.Errorf("%s: fixture %q does not match any fixture: %w", path, entry.ID, errUnknownXFailID)
		}
		if field := missingXFailField(entry); field != "" {
			return fmt.Errorf("%s: fixture %q missing required field %q: %w", path, entry.ID, field, errMissingXFailField)
		}
		expiresAt, err := entry.expiresAt()
		if err != nil {
			return fmt.Errorf("%s: fixture %q invalid expires_at %q: %w", path, entry.ID, entry.ExpiresAt, err)
		}
		if !expiresAt.After(now) {
			return fmt.Errorf("%s: fixture %q expired on %s: %w", path, entry.ID, entry.ExpiresAt, errExpiredXFail)
		}
	}
	return nil
}

func (x xfail) active(now time.Time) bool {
	expiresAt, err := x.expiresAt()
	return err == nil && expiresAt.After(now)
}

func (x xfail) expiresAt() (time.Time, error) {
	return time.Parse(time.DateOnly, x.ExpiresAt)
}

func missingXFailField(entry xfail) string {
	switch {
	case entry.Reason == "":
		return "reason"
	case entry.ExpiresAt == "":
		return "expires_at"
	case entry.TrackingIssue == "":
		return "tracking_issue"
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
