package conformance

import (
	"path/filepath"
	"time"
)

func SkipReason(root, id string, now time.Time) (string, bool) {
	divergenceIDs, err := LoadDivergenceIDs(filepath.Join(root, "testdata", "divergences.md"))
	if err == nil {
		if _, ok := divergenceIDs[id]; ok {
			return "divergence listed in testdata/divergences.md", true
		}
	}
	xfails, err := LoadXFails(filepath.Join(root, "testdata", "xfail.json"))
	if err != nil {
		return "", false
	}
	for _, xfail := range xfails {
		if xfail.ID != id {
			continue
		}
		expiresAt, err := time.Parse(time.DateOnly, xfail.ExpiresAt)
		if err != nil || !expiresAt.After(now) {
			return "", false
		}
		return xfail.Reason, true
	}
	return "", false
}
