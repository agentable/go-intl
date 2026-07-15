package conformance

import (
	"testing"
	"time"
)

func RunFixtures(t *testing.T, root string, run func(*testing.T, Fixture)) {
	t.Helper()

	suite, err := loadRunSuite(root, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range suite.fixtures {
		t.Run(fixture.ID, func(t *testing.T) {
			t.Parallel()
			if reason, ok := suite.skipReasons[fixture.ID]; ok {
				t.Skipf("skipping %s: %s", fixture.ID, reason)
			}
			run(t, fixture)
		})
	}
}

type runSuite struct {
	fixtures    []Fixture
	skipReasons map[string]string
}

func loadRunSuite(root string, now time.Time) (runSuite, error) {
	if err := validatePackageRoot(root); err != nil {
		return runSuite{}, err
	}
	fixtures, err := LoadFixtures(root)
	if err != nil {
		return runSuite{}, err
	}
	divergences, err := loadActiveDivergences(divergenceLedgerPath(root))
	if err != nil {
		return runSuite{}, err
	}
	if err := validateDivergenceEntries(divergenceLedgerPath(root), fixtures, divergences); err != nil {
		return runSuite{}, err
	}
	xfailFile := xfailPath(root)
	xfails, err := loadXFails(xfailFile)
	if err != nil {
		return runSuite{}, err
	}
	if err := validateXFailEntries(xfailFile, fixtures, xfails, now); err != nil {
		return runSuite{}, err
	}
	reasons := make(map[string]string, len(divergences)+len(xfails))
	for id := range divergences {
		reasons[id] = "divergence listed in " + divergenceLedgerRelativePath()
	}
	for _, xfail := range xfails {
		reasons[xfail.ID] = xfail.Reason
	}
	return runSuite{fixtures: fixtures, skipReasons: reasons}, nil
}
