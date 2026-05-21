package gointl

import (
	"os"
	"strings"
	"testing"
)

func TestTaskfileConformanceTargets(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if !strings.Contains(content, "conformance:verify:") {
		t.Fatal("Taskfile.yml missing conformance:verify")
	}
	if strings.Contains(content, "conformance:fixtures:") || strings.Contains(content, "conformance:divergences:") {
		t.Fatal("Taskfile.yml should keep conformance validation in one target")
	}
	if !strings.Contains(content, "go run ./tools/check-conformance") {
		t.Fatal("conformance:verify must use the unified check-conformance tool")
	}
	if strings.Contains(strings.ToLower(content), "icu4j") || strings.Contains(strings.ToLower(content), "java") {
		t.Fatal("Taskfile.yml must not invoke ICU4J or Java")
	}
}

func TestTaskfileBenchmarkTargets(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	for _, target := range []string{"benchstat:", "bench:run:", "bench:"} {
		if !strings.Contains(content, target) {
			t.Fatalf("Taskfile.yml missing %s", target)
		}
	}
	if !strings.Contains(content, "golang.org/x/perf/cmd/benchstat") {
		t.Fatal("Taskfile.yml must install benchstat as a CLI")
	}
	if !strings.Contains(content, "go test -run '^$' -bench=.") {
		t.Fatal("Taskfile.yml must run Go benchmarks")
	}
	if strings.Contains(content, "bench:gate:") {
		t.Fatal("Taskfile.yml must keep benchmarks as telemetry, not expose a benchmark gate")
	}
	if strings.Contains(content, "go test -tags perf") {
		t.Fatal("Taskfile.yml must not run perf-tag budget tests")
	}
}

func TestTaskfileLintRunsSerially(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if strings.Contains(content, "  lint:\n    desc: Run all linters\n    deps:") {
		t.Fatal("lint must not run tidy-lint and golangci-lint as parallel deps")
	}
	if !strings.Contains(content, "  lint:\n    desc: Run all linters\n    cmds:\n      - task: tidy-lint\n      - task: golangci-lint") {
		t.Fatal("lint must run tidy-lint before golangci-lint")
	}
}

func TestTaskfileLintBootstrapBoundary(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if !strings.Contains(content, "golangci-lint:preflight:") {
		t.Fatal("Taskfile.yml must expose a golangci-lint preflight target")
	}
	if !strings.Contains(content, "golangci-lint bootstrap failed") {
		t.Fatal("Taskfile.yml must label golangci-lint bootstrap failures distinctly")
	}
	if !strings.Contains(content, "task: golangci-lint:preflight") {
		t.Fatal("golangci-lint execution must depend on the bootstrap preflight")
	}
}

func TestTaskfileLintBootstrapUsesGoInstall(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if !strings.Contains(content, "github.com/golangci/golangci-lint/v2/cmd/golangci-lint") {
		t.Fatal("golangci-lint bootstrap must install the pinned Go module")
	}
	if strings.Contains(content, "raw.githubusercontent.com/golangci/golangci-lint/master/install.sh") {
		t.Fatal("golangci-lint bootstrap must not depend on the upstream shell installer")
	}
}

func TestTaskfileLintPreflightReadsInstalledBinary(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if !strings.Contains(content, "actual=$({{.GOLANGCI_LINT_BINARY}} version") {
		t.Fatal("golangci-lint preflight must read the installed binary after bootstrap")
	}
	if strings.Contains(content, `if [ "{{.GOLANGCI_LINT_VERSION}}" != "{{.REQUIRED_GOLANGCI_LINT_VERSION}}" ]; then
          echo "golangci-lint bootstrap failed`) {
		t.Fatal("golangci-lint preflight must not use the stale task variable after install")
	}
}

func TestTaskfileDataContractTarget(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if !strings.Contains(content, "data:contract:") {
		t.Fatal("Taskfile.yml must expose a lightweight data:contract target")
	}
	if !strings.Contains(content, "task: data:contract") {
		t.Fatal("task verify must include the lightweight data contract")
	}
}

func TestTaskfileBuildSizeUsesSizecheckTelemetry(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if !strings.Contains(content, "build:size:") {
		t.Fatal("Taskfile.yml must expose build:size telemetry")
	}
	if !strings.Contains(content, "go run ./tools/sizecheck") {
		t.Fatal("build:size must use tools/sizecheck for direct formatter, root facade, and CLDR measurements")
	}
	if !strings.Contains(content, "build:size:cold:") {
		t.Fatal("Taskfile.yml must expose cold build-size telemetry")
	}
	if !strings.Contains(content, "go run ./tools/sizecheck -cold") {
		t.Fatal("build:size:cold must clear the Go build cache before measuring compile time")
	}
}

func readTaskfile(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
