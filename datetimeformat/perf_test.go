//go:build perf

package datetimeformat

import "testing"

func TestPerf_DateTimeFormat_DateStyleShort_Cached(t *testing.T) {
	t.Parallel()
	checkBenchmarkBudget(t, BenchmarkDateTimeFormat_DateStyleShort_Cached, 2000, -1)
}

func TestPerf_DateTimeFormat_DateTimeRange_Cached(t *testing.T) {
	t.Parallel()
	checkBenchmarkBudget(t, BenchmarkDateTimeFormat_DateTimeRange_Cached, 3500, -1)
}

func TestPerf_DateTimeFormat_FormatToParts_Cached(t *testing.T) {
	t.Parallel()
	checkBenchmarkBudget(t, BenchmarkDateTimeFormat_FormatToParts_Cached, 3000, -1)
}

func TestPerf_DateTimeFormat_New(t *testing.T) {
	t.Parallel()
	checkBenchmarkBudget(t, BenchmarkDateTimeFormat_New, 8000, -1)
}

func checkBenchmarkBudget(t *testing.T, fn func(*testing.B), nsBudget int64, allocBudget int64) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping perf gate in short mode")
	}
	if raceEnabled {
		t.Skip("skipping perf gate under race detector")
	}
	result := testing.Benchmark(fn)
	if ns := result.NsPerOp(); ns > nsBudget {
		t.Fatalf("benchmark regressed: %d ns/op (budget %d)", ns, nsBudget)
	}
	if allocBudget >= 0 && result.AllocsPerOp() > allocBudget {
		t.Fatalf("benchmark allocations regressed: %d allocs/op (budget %d)", result.AllocsPerOp(), allocBudget)
	}
}
