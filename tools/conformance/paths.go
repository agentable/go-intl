package conformance

import "path/filepath"

const (
	conformanceTestdataDir = "testdata"
	conformanceFixturesDir = "conformance"
	divergenceLedgerFile   = "divergences.md"
	xfailFile              = "xfail.json"
)

func conformanceFixturesPath(root string) string {
	return filepath.Join(root, conformanceTestdataDir, conformanceFixturesDir)
}

func divergenceLedgerPath(root string) string {
	return filepath.Join(root, conformanceTestdataDir, divergenceLedgerFile)
}

func divergenceLedgerRelativePath() string {
	return filepath.ToSlash(filepath.Join(conformanceTestdataDir, divergenceLedgerFile))
}

func xfailPath(root string) string {
	return filepath.Join(root, conformanceTestdataDir, xfailFile)
}
