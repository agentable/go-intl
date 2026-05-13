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
)

type XFail struct {
	ID            string `json:"id"`
	Reason        string `json:"reason"`
	ExpiresAt     string `json:"expires_at"`
	TrackingIssue string `json:"tracking_issue"`
}

func ValidateXFailFile(root string, now time.Time) error {
	path := filepath.Join(root, "testdata", "xfail.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var entries []XFail
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	for _, entry := range entries {
		if entry.ID == "" {
			return fmt.Errorf("%s: missing required field %q: %w", path, "id", errMissingXFailField)
		}
		if entry.Reason == "" {
			return fmt.Errorf("%s: fixture %q missing required field %q: %w", path, entry.ID, "reason", errMissingXFailField)
		}
		if entry.ExpiresAt == "" {
			return fmt.Errorf("%s: fixture %q missing required field %q: %w", path, entry.ID, "expires_at", errMissingXFailField)
		}
		if entry.TrackingIssue == "" {
			return fmt.Errorf("%s: fixture %q missing required field %q: %w", path, entry.ID, "tracking_issue", errMissingXFailField)
		}
		expiresAt, err := time.Parse(time.DateOnly, entry.ExpiresAt)
		if err != nil {
			return fmt.Errorf("%s: fixture %q invalid expires_at %q: %w", path, entry.ID, entry.ExpiresAt, err)
		}
		if !expiresAt.After(now) {
			return fmt.Errorf("%s: fixture %q expired on %s: %w", path, entry.ID, entry.ExpiresAt, errExpiredXFail)
		}
	}
	return nil
}

func LoadXFails(path string) ([]XFail, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []XFail
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return entries, nil
}
