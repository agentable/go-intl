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
	if !strings.Contains(content, "conformance:witness:") {
		t.Fatal("Taskfile.yml missing conformance:witness")
	}
	if !strings.Contains(content, "cd tools/gen-fixtures-from-formatjs") ||
		!strings.Contains(content, "go run . -node \"$node_path\" -out ../..") ||
		!strings.Contains(content, "./tools/node-witness") ||
		!strings.Contains(content, `node_path="$(command -v node)"`) ||
		!strings.Contains(content, "requires node on PATH") ||
		!strings.Contains(content, "Node witness diff:") ||
		!strings.Contains(content, "git diff -- '*/testdata/conformance/node-v*/' testdata/native/") ||
		!strings.Contains(content, "task conformance:verify") {
		t.Fatal("conformance:witness must refresh Node fixtures, print their diff, and point back to conformance verification")
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

func TestTaskfileLintUsesSharedDeps(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if !strings.Contains(content, "  lint:\n    desc: Run all linters\n    deps:\n      - golangci-lint\n      - tidy-lint") {
		t.Fatal("lint must run the shared golangci-lint and tidy-lint dependencies")
	}
	if strings.Contains(content, "  lint:\n    desc: Run all linters\n    cmds:\n      - task: tidy-lint") {
		t.Fatal("lint must use the shared dependency form")
	}
}

func TestTaskfileLintSelectsExactVersionBinary(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	for _, want := range []string{
		"GOLANGCI_LINT_LOCAL_BINARY",
		`for candidate in "$(command -v golangci-lint 2>/dev/null || true)" "{{.GOLANGCI_LINT_LOCAL_BINARY}}"`,
		`echo "{{.GOLANGCI_LINT_LOCAL_BINARY}}"`,
		"  golangci-lint:\n    desc: Run golangci-lint\n    deps:\n      - install-golangci-lint",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("Taskfile.yml missing exact-version golangci-lint wiring %q", want)
		}
	}
}

func TestTaskfileLintBootstrapUsesPinnedInstaller(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if !strings.Contains(content, "raw.githubusercontent.com/golangci/golangci-lint/master/install.sh") {
		t.Fatal("golangci-lint bootstrap must use the pinned upstream installer")
	}
	if !strings.Contains(content, `| sh -s -- -b "{{.GOBIN}}" v{{.REQUIRED_GOLANGCI_LINT_VERSION}}`) {
		t.Fatal("golangci-lint bootstrap must install the required version into GOBIN")
	}
	if strings.Contains(content, "github.com/golangci/golangci-lint/v2/cmd/golangci-lint") {
		t.Fatal("golangci-lint bootstrap must not compile the linter from source")
	}
}

func TestTaskfileLintHasNoDeadPreflight(t *testing.T) {
	t.Parallel()

	content := readTaskfile(t)
	if strings.Contains(content, "golangci-lint:preflight:") {
		t.Fatal("Taskfile.yml must not keep an unused golangci-lint preflight target")
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
	if !strings.Contains(content, "SupportedCodesDoesNotDecodeOtherBlobs") ||
		!strings.Contains(content, "SupportedCalendarsReturnsCopy") ||
		!strings.Contains(content, "GeneratedNumberingSystemExtrasHaveRuntimePayload") ||
		!strings.Contains(content, "RuntimeNumberingSystemPayloadsAreAdvertised") ||
		!strings.Contains(content, "go test ./internal/numbering") ||
		!strings.Contains(content, "SupportedCollationsReturnsCopy") ||
		!strings.Contains(content, "TestSupportedLocales(ExcludesTailoredLocales|ReturnsSnapshot)$") ||
		!strings.Contains(content, "./internal/collation") ||
		!strings.Contains(content, "./internal/segmentation") ||
		!strings.Contains(content, "RetiredRootCLDRHasNoGoFiles") ||
		!strings.Contains(content, "./internal/cldr/currency ./internal/cldr/date ./internal/cldr/displaynames ./internal/cldr/list ./internal/cldr/number ./internal/cldr/relativetime ./internal/cldr/timezone ./internal/cldr/unit") {
		t.Fatal("data:contract must include leaf CLDR narrow-index guards")
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
