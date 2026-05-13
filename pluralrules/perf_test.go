//go:build perf

package pluralrules

import "testing"

func TestPerf_PluralRules_Cardinal_Cached(t *testing.T) {
	t.Parallel()
	checkBenchmarkBudget(t, BenchmarkPluralRules_Cardinal_Cached, 200, 0)
}

func TestPerf_PluralRules_Ordinal_Cached(t *testing.T) {
	t.Parallel()
	checkBenchmarkBudget(t, BenchmarkPluralRules_Ordinal_Cached, 250, -1)
}

func TestPerf_PluralRules_SelectRange_Cached(t *testing.T) {
	t.Parallel()
	checkBenchmarkBudget(t, BenchmarkPluralRules_SelectRange_Cached, 400, -1)
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
