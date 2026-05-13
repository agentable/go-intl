//go:build perf

package numberformat

import (
	"sync"
	"testing"
)

func TestPerf_NumberFormat_Decimal_Cached(t *testing.T) {
	checkBenchmarkBudget(t, BenchmarkNumberFormat_Decimal_Cached, 800, -1)
}

func TestPerf_NumberFormat_Currency_Cached(t *testing.T) {
	checkBenchmarkBudget(t, BenchmarkNumberFormat_Currency_Cached, 1200, -1)
}

func TestPerf_NumberFormat_Compact_Cached(t *testing.T) {
	checkBenchmarkBudget(t, BenchmarkNumberFormat_Compact_Cached, 1500, -1)
}

func TestPerf_NumberFormat_FormatToParts_Cached(t *testing.T) {
	checkBenchmarkBudget(t, BenchmarkNumberFormat_FormatToParts_Cached, 1800, -1)
}

func TestPerf_NumberFormat_New(t *testing.T) {
	checkBenchmarkBudget(t, BenchmarkNumberFormat_New, 5000, -1)
}

var benchmarkBudgetMu sync.Mutex

func checkBenchmarkBudget(t *testing.T, fn func(*testing.B), nsBudget int64, allocBudget int64) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping perf gate in short mode")
	}
	if raceEnabled {
		t.Skip("skipping perf gate under race detector")
	}
	benchmarkBudgetMu.Lock()
	defer benchmarkBudgetMu.Unlock()
	result := testing.Benchmark(fn)
	if ns := result.NsPerOp(); ns > nsBudget {
		t.Fatalf("benchmark regressed: %d ns/op (budget %d)", ns, nsBudget)
	}
	if allocBudget >= 0 && result.AllocsPerOp() > allocBudget {
		t.Fatalf("benchmark allocations regressed: %d allocs/op (budget %d)", result.AllocsPerOp(), allocBudget)
	}
}
