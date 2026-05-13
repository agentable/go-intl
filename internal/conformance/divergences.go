package conformance

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var errUnknownDivergence = errors.New("unknown divergence id")

func LoadDivergenceIDs(path string) (map[string]struct{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	ids := map[string]struct{}{}
	currentID := ""
	resolved := false
	flush := func() {
		if currentID != "" && !resolved {
			ids[currentID] = struct{}{}
		}
		currentID = ""
		resolved = false
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if value, ok := strings.CutPrefix(line, "id:"); ok {
			flush()
			currentID = strings.TrimSpace(value)
			continue
		}
		if value, ok := strings.CutPrefix(line, "status:"); ok {
			resolved = strings.TrimSpace(value) == "resolved"
		}
	}
	flush()
	return ids, nil
}

func ValidateDivergences(fixtureRoot, divergencePath string) error {
	fixtures, err := LoadFixtures(fixtureRoot)
	if err != nil {
		return err
	}
	fixtureIDs := map[string]struct{}{}
	for _, fixture := range fixtures {
		fixtureIDs[fixture.ID] = struct{}{}
	}
	divergenceIDs, err := LoadDivergenceIDs(divergencePath)
	if err != nil {
		return err
	}
	for id := range divergenceIDs {
		if _, ok := fixtureIDs[id]; !ok {
			return fmt.Errorf("%s: divergence id %q does not match any fixture: %w", divergencePath, id, errUnknownDivergence)
		}
	}
	return nil
}
