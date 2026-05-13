package conformance

import (
	"testing"
	"time"
)

func RunFixtures(t *testing.T, root string, run func(*testing.T, Fixture)) {
	t.Helper()

	fixtures, err := LoadFixtures(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.ID, func(t *testing.T) {
			t.Parallel()
			if reason, ok := SkipReason(root, fixture.ID, time.Now()); ok {
				t.Skipf("skipping %s: %s", fixture.ID, reason)
			}
			run(t, fixture)
		})
	}
}
