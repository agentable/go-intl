# SPEC 71 — Benchmark Strategy & Performance Telemetry

> **Status:** Revised (2026-05-31)
> **Type:** Rule + Flow — defines how go-intl measures hot-path performance without turning benchmark numbers into merge gates.
> **Authority:** This spec records benchmark layout, baseline selection, non-blocking telemetry, and `benchstat` report flow. SPEC 70 records correctness gates; this spec records performance evidence.

---

## Overview

go-intl’s performance signal consists of two layers of benchmarks:

1. **PerCall** (including `New`): reflects the ECMA-402 constructor path, that is, the cost of calling a method immediately after `new Intl.<Constructor>(locales, options)`.
2. **Cached** (only method call): reflects the pure hot-path after formatter has been constructed.

baseline = `golang.org/x/text/message` + `golang.org/x/text/feature/plural` (stdlib-grade Go implementation); **FORBIDDEN** using embedded JS engines as a baseline because cross-process maintenance and runtime overhead pollute the signal.

> **Why**: ECMA-402 has both construction time cost and method hot-path cost. A two-tier baseline makes option resolution / locale negotiation and formatting main paths visible separately.
> **Rejected**: root one-shot / global-cache benchmark as the main signal - JavaScript `Intl` namespace does not have per-locale session or root one-shot helpers, this kind of API is not a long-term public surface.

---

## 1. Benchmark Layout

### 1.1 File location

| Package | File | Content |
|----|------|------|
| `numberformat/` | `numberformat/benchmark_test.go` | NumberFormat PerCall + Cached |
| `datetimeformat/` | `datetimeformat/benchmark_test.go` | DateTimeFormat PerCall + Cached |
| `pluralrules/` | `pluralrules/benchmark_test.go` | PluralRules Cardinal/Ordinal Cached(PerCall can be omitted) |
| `durationformat/` | `durationformat/benchmark_test.go` | DurationFormat PerCall + Cached, including digital formatting |
| `locale/` | `locale/benchmark_test.go` | `locale.Parse(string)` boundary analysis + `locale.New(string, Options{})` construction |
| baseline | `<package>/benchmark_baseline_test.go` | `x/text/message` / `x/text/feature/plural` comparison for NumberFormat / PluralRules only |

**Rules**:

1. benchmark **MUST** be packaged in the same package (`*_test.go`) as conformance test, and no additional `benchmarks/` top-level directory is required.
2. The baseline benchmark **MUST** be located in `_baseline_test.go` (independent file), so that `go test -bench=Baseline` can be run alone.
3. **FORBIDDEN** Writing benchmarks in the main package (violates Go testing conventions, and cannot be used to isolate baseline with `_test.go` build tag).

> **Why**: The same package benchmark makes godoc / IDE jump seamless, consistent with stdlib habits.
> **Rejected**: a top-level `bench/` directory; the package-local benchmark layout is simpler and gives no less signal.

### 1.2 Naming Convention

```go
// PerCall (including structure)
func BenchmarkNumberFormat_Decimal_PerCall(b *testing.B)
func BenchmarkNumberFormat_Currency_PerCall(b *testing.B)

// Cached (pure hot-path)
func BenchmarkNumberFormat_Decimal_Cached(b *testing.B)
func BenchmarkNumberFormat_Currency_Cached(b *testing.B)
func BenchmarkNumberFormat_Compact_Cached(b *testing.B)
func BenchmarkNumberFormat_FormatToParts_Cached(b *testing.B)

// PluralRules
func BenchmarkPluralRules_Cardinal_Cached(b *testing.B)
func BenchmarkPluralRules_Ordinal_Cached(b *testing.B)
func BenchmarkPluralRules_SelectRange_Cached(b *testing.B)

// DateTimeFormat
func BenchmarkDateTimeFormat_DateStyleShort_PerCall(b *testing.B)
func BenchmarkDateTimeFormat_DateStyleShort_Cached(b *testing.B)
func BenchmarkDateTimeFormat_DateTimeRange_Cached(b *testing.B)
func BenchmarkDateTimeFormat_FormatToParts_Cached(b *testing.B)

// DurationFormat
func BenchmarkDurationFormat_Short_PerCall(b *testing.B)
func BenchmarkDurationFormat_Short_Cached(b *testing.B)
func BenchmarkDurationFormat_Digital_Cached(b *testing.B)

// Construction period (measured separately)
func BenchmarkNumberFormat_New(b *testing.B)
func BenchmarkDateTimeFormat_New(b *testing.B)

// Baseline (independent file)
func BenchmarkBaseline_XText_Decimal(b *testing.B)
func BenchmarkBaseline_XText_Plural_Cardinal(b *testing.B)
```

**Rules**:

1. The benchmark function name **MUST** be the `Benchmark<Type>_<Feature>_<Layer>` three-part formula; **implicit abbreviations are prohibited** (such as `BenchmarkNF`).
2. The baseline benchmark **MUST** be prefixed with `BenchmarkBaseline_` to visually distinguish it from the production benchmark.
3. PerCall benchmark **MUST** contain a complete constructor `New(...)` call; Cached benchmark **MUST** be constructed before `b.Loop()`.

### 1.3 b.Loop mandatory use

```go
func BenchmarkNumberFormat_Decimal_Cached(b *testing.B) {
    b.ReportAllocs()
    nf, _ := numberformat.New(en, numberformat.Options{Style: numberformat.DecimalStyle})
    for b.Loop() {
        _ = nf.Format(numberformat.Int(42))
    }
}
```

**Rules**:

1. **MUST** use Go 1.24+ `b.Loop()` (project baseline Go 1.27.0). **FORBIDDEN** to use the old-style `for i := 0; i < b.N; i++` loop.
2. Cached benchmark **MUST** complete the construction before `b.Loop()`; **FORBIDDEN** to `New(...)` in the loop body (that is the scope of PerCall).
3. All benchmarks **MUST** call `b.ReportAllocs()`, reporting B/op + allocs/op outside of ns/op.
4. **FORBIDDEN** Calling `time.Now()` / `runtime.GC()` and other perturbation measurements within the benchmark.

> **Why**: `b.Loop()` is the standard benchmark loop API of Go 1.24+, which automatically handles keepalive and inline, and has fewer pitfalls than the `b.N` loop.
> **Rejected**: Old style `for i := 0; i < b.N; i++` - new benchmark uses `b.Loop()` uniformly.

### 1.4 Root facade measurement

The root `github.com/agentable/go-intl` package is an aggregate facade over all active constructor packages. Import-cost measurements must keep that aggregate signal separate from per-surface package signals.

Required signals for import-cost or binary-size work:

| Signal | Root facade | Per-surface |
|--------|-------------|-------------|
| Dependency graph | `go list -deps .` | `go list -deps ./numberformat`, `go list -deps ./datetimeformat`, or the touched constructor package |
| Binary-size smoke | a harness that imports `github.com/agentable/go-intl` and labels the result as root facade cost | a harness that imports only the touched constructor package and labels the package path |
| CLDR linker smoke | `task build:size` when generated data shape changes | same command as a data-layer smoke signal, never as a substitute for a per-surface harness |

**Rules**:

1. Root facade dependency counts and binary-size numbers **MUST** be marked as aggregate facade cost.
2. Formatter package reports **MUST** use the corresponding sub-package path, and the root facade number must not be written as `numberformat` / `datetimeformat` and other single surface costs.
3. Root facade measurements **MUST NOT** be reduced by removing root constructor aliases or constructor imports; SPEC 60 owns the namespace shape.

> **Why**: The root package intentionally imports every active constructor package so aliases mirror JavaScript `Intl` constructor properties. Per-surface production services should not inherit that aggregate measurement unless they import the root facade.
>
> **Rejected**: using root package benchmark or binary-size numbers as evidence for one formatter. That hides the direct-import production path and steers optimization toward API removal instead of real package boundaries.

---

## 2. Baseline Selection

### 2.1 Main baseline:`golang.org/x/text`

| comparison item | x/text API | go-intl API | baseline file |
|-------|-----------|-------------|--------------|
| Decimal | `message.NewPrinter(tag).Sprintf("%v", n)` | `numberformat.New(locale.List{loc}, opts).Format(numberformat.Int(n))` | `numberformat/benchmark_baseline_test.go` |
| Percent | `number.Percent(0.5)` by `message.Printf` | `numberformat.New(locale.List{loc}, numberformat.Options{Style: numberformat.PercentStyle}).Format(numberformat.Float(0.5))` | Same as above |
| PluralCardinal | `plural.Cardinal.Matches(tag, ...)` | `pluralrules.New(locale.List{loc}, opts).Select(pluralrules.Int(n))` | `pluralrules/benchmark_baseline_test.go` |
| PluralOrdinal | `plural.Ordinal.Matches(tag, ...)` | `pluralrules.New(locale.List{loc}, pluralrules.Options{Type: pluralrules.Ordinal}).Select(pluralrules.Int(n))` | Same as above |

**Rules**:

1. The main baseline **MUST** be `golang.org/x/text/message` + `golang.org/x/text/feature/plural`; both have appeared in SPEC 50 / SPEC 40 (as a comparison object, **not** production dependent).
2. baseline only covers **decimal / percent / plural cardinal / plural ordinal** four levels; currency / unit / compact / DateTimeFormat / Range. Since x/text has no corresponding API, it does not participate in the baseline comparison and is changed to "absolute throughput + memory allocation" from the baseline.
3. baseline and go-intl run with the same input (same locale, same input value); **It is prohibited** to use different inputs to create misleading comparisons.

> **Why**: x/text is Go native, zero tool chain, and officially maintained; benchmark can be written by importing in the same module, which is consistent with go-intl deployment assumption (same process).
> **Rejected**: a cross-process JS baseline; IPC overhead pollutes the signal and the maintenance cost is not justified.
> **Rejected**: readable polyfill ops/sec (README §"Benchmark Results") - it already exists, but in screenshot status, it is not running in CI; as a reference wall, it **not** enters CI gate.

### 2.2 Disable embedded JS baseline

**FORBIDDEN** Running `Intl.NumberFormat` as baseline through embedded JS engine. Reason:

1. JS host integration is a candidate for consumer-driven expansion and does not overlap with the active scope timeline.
2. The startup overhead of embedded JS engine is ≥ 100ms, the benchmark signal is dominated by startup, and the steady state cannot be measured.
3. Maintenance cost = JS engine tool chain + version pinning + JS-Go marshalling verification; ROI does not match the go-intl active scope target.
4. Byte equality has been covered by the SPEC 70 fixture; the performance comparison "same order of magnitude" target is sufficient to be guaranteed by x/text baseline + absolute throughput.

> **Why**: Embedded JS baseline brings 4 second-order problems (startup / tool chain / version pinning / marshalling), in exchange for 1 doubtful gain, YAGNI does the opposite.
> **Rejected**: Embedded JS baseline - Explicitly rejected, any second phase PR proposal **MUST** first revise this SPEC §2.2.

### 2.3 Reference target (compared to baseline)

The following expressions are investigation priority and magnitude judgment, not CI blocking conditions.

| Dimension | Reference reading (compared to baseline) | Remarks |
|-----|---------------------|------|
| `BenchmarkNumberFormat_Decimal_Cached` vs `BenchmarkBaseline_XText_Decimal` | If go-intl is slower than baseline / 1.5, investigate first | Amortized construction after cache hit |
| `BenchmarkNumberFormat_Decimal_PerCall` vs baseline | If go-intl is slower than baseline/2.0, investigate first | Contains structure, allowing higher cost |
| `BenchmarkPluralRules_Cardinal_Cached` vs baseline | If go-intl is slower than baseline / 1.5, investigate first | The codegen path should be close to stdlib |
| `BenchmarkDateTimeFormat_*` | Only `b.ReportAllocs()`, no baseline | x/text has no corresponding API |
| `BenchmarkDurationFormat_*` | Only `b.ReportAllocs()`, no baseline | ECMA-402 duration composition has no x/text equivalent |
| `BenchmarkNumberFormat_Compact_Cached` | only `b.ReportAllocs()`, no baseline | x/text no compact |
| `BenchmarkNumberFormat_Currency_Cached` | Only `b.ReportAllocs()`, no baseline | x/text/currency form does not correspond |

> **Why**: The "same order of magnitude" goal (within 2×) is enough for go-intl to have no obvious disadvantage compared to x/text in the production environment; it is unrealistic to tie stdlib (go-intl has more ECMA-402 spec behaviors).
> **Rejected**: Encode these reference targets into test failure conditions - machine, Go version, CPU governor and temporary load can all make a single benchmark noisy; wrong red lights are more likely to drive wrong designs than no red lights.

---

## 3. Non-Blocking Telemetry Policy

Benchmark numbers are evidence, not gates. A benchmark result may justify investigation, a profile, or a focused optimization PR; it must not be the sole reason a merge fails.

**Rules**:

1. `task verify`, package tests, and CI correctness gates ** must not ** fail solely because `ns/op`, `B/op`, or `allocs/op` crossed a number.
2. Do not add `TestPerf_*`, `*_perf_test.go`, `go test -tags perf`, or `testing.Benchmark(...)` budget assertions.
3. If performance evidence changes an architectural decision, record the command, Go version, GOARCH, CPU, package path, and before/after numbers in the change summary or owning SPEC.
4. ECMA-402 correctness, supported-locale truthfulness, and Go API clarity outrank benchmark gains. A faster implementation that changes observable Intl behavior is a bug.
5. Keep root facade, per-surface formatter, and CLDR linker measurements separate. Root facade measurements are aggregate namespace cost, not a single formatter cost.

> **Why**: Performance still matters, but single-machine benchmark gates turn noisy numbers into design pressure. go-intl needs stable Intl behavior for years; benchmark telemetry should guide where to profile, not override the contract.
>
> **Rejected**: nightly or PR benchmark blocking jobs. They look objective, but CPU variance and Go version drift make them poor merge criteria for this library.

---

## 4. benchstat Workflow

### 4.1 Tools

`benchstat` is the CLI tool of `golang.org/x/perf/cmd/benchstat`, **not** go.mod dependent.

**Rules**:

1. CI runner **MUST** install the pinned `benchstat` version declared by `Taskfile.yml` through `go install golang.org/x/perf/cmd/benchstat@<version>` (independent command, not go.mod require).
2. **FORBIDDEN** to add `golang.org/x/perf` to `go.mod`; the active scope keeps runtime and test dependencies tightly closed.
3. The active pin is `golang.org/x/perf/cmd/benchstat@v0.0.0-20260615155930-9e4b9ddef5b6`. Refreshing it is a release-maintenance act and must update this SPEC and `Taskfile.yml` together.

### 4.2 Report format

`task bench` Subtask **Required**:

1. Run the full set of benchmarks ≥ 5 times (default `-count=5`).
2. Output `.tmp/bench-current.txt`, the file header contains Go version, GOARCH, and CPU (when available).
3. When the caller provides `BASELINE=<file>`, call `benchstat "$BASELINE" .tmp/bench-current.txt` to generate the report.
4. Report format **MUST** contain: benchmark name, N, ns/op delta, p-value, B/op delta, allocs/op delta.
5. PR-CI can post the report to the PR comment area, but the report must not set failing status.
6. main/nightly can retain report artifacts ≥ 30 days, but the report must not be used as a standalone block.

Committed baseline artifacts live under `testdata/bench/` and use
`baseline-<go-version>-<goos>-<goarch>.txt`. The current Apple Silicon
baseline is `testdata/bench/baseline-go1.27-darwin-arm64.txt`, captured with:

```bash
go test -run '^$' -bench=. -benchmem ./numberformat ./pluralrules
```

Use it with `task bench BASELINE=testdata/bench/baseline-go1.27-darwin-arm64.txt`
when reviewing performance regressions on comparable hardware. Refreshing the
file is a release-maintenance act, not a correctness gate.

---

## 5. Forbidden

- **BANNED** Old-style `for i := 0; i < b.N; i++` loops (requires `b.Loop()`).
- **BANNED** from calling formatter `New(...)` within a Cached benchmark (that is PerCall scope).
- **FORBIDDEN** Add `golang.org/x/perf` to `go.mod` (benchstat is CLI, not go-import dependent).
- **BANNED** Baseline via embedded JS engine (SPEC §2.2).
- **FORBIDDEN** active scope introduces `clipperhouse/uax29`, `bojanz/currency` and other consumer-driven expansion candidates as baseline - baseline only `x/text`.
- **Disable** any benchmark number as standalone blocking merge rule.
- **BANNED** `perf` build tag, `TestPerf_*`, `testing.Benchmark(...)` budget assertion, or `task bench:gate`.
- **FORBIDDEN** The baseline benchmark is the same file as the production benchmark (must be `_baseline_test.go` independent).
- **Disabled** Calling `time.Now()` / `runtime.GC()` perturbation measurements within the benchmark.
- **BANNED** Improve root facade aggregate cost by removing root constructor aliases or hiding imports.

---

## 6. Acceptance Criteria

### Layout

- [ ] Active benchmarked formatter packages have `benchmark_test.go`; `benchmark_baseline_test.go` exists only for NumberFormat / PluralRules.
- [ ] benchmark naming follows the `Benchmark<Type>_<Feature>_<Layer>` three-part formula.
- [ ] All benchmarks use `b.Loop()` and `b.ReportAllocs()`.
- [ ] baseline benchmark prefix `BenchmarkBaseline_`.

### Baseline

- [ ] `BenchmarkBaseline_XText_Decimal` exists in `numberformat/`.
- [ ] `BenchmarkBaseline_XText_Plural_Cardinal` / `Ordinal` exists in `pluralrules/`.
- [ ] `go.mod` does not contain embedded JS engine related dependencies.

### Telemetry

- [ ] No `perf` build tag, `TestPerf_*`, `testing.Benchmark(...)` budget assertion, or `task bench:gate` exists.
- [ ] Performance evidence that changes architecture records command, Go version, GOARCH, CPU, package path, and before/after numbers.
- [ ] Root facade, per-surface formatter, and CLDR linker measurements are labeled separately.

### Reporting

- [ ] `task bench` subtask exists, run ≥ 5 times benchmark + benchstat report.
- [ ] Committed baseline artifacts live under `testdata/bench/baseline-<go-version>-<goos>-<goarch>.txt`.
- [ ] PR-CI may post benchstat reports, but must not fail on benchmark numbers alone.
- [ ] main/nightly may retain benchmark artifacts ≥ 30 days, but must not block on benchmark numbers alone.

### Tooling

- [ ] benchstat is installed through the pinned `go install` version in `Taskfile.yml`; **not** present in `go.mod`.
- [ ] CI runner specifications (CPU / Go version / GOARCH) are recorded in the benchstat report header.
- [ ] Import-cost reports list `go list -deps .` separately from each touched per-surface package.
- [ ] Binary-size reports label root facade harnesses separately from per-surface harnesses.

---

## References

- SPECS/00 §7(consumer-driven expansion optimization anchor point)
- SPEC 60(root `Intl` namespace; forbids root one-shot helpers as benchmark ownership)
- SPEC 70 §4(correctness gates and performance telemetry boundary)
- `.references/formatjs/packages/intl-numberformat/benchmark/benchmark.ts:1-94` (double-layer benchmark inspiration)
- `golang.org/x/text/message`(main baseline)
- `golang.org/x/text/feature/plural`(main baseline)
- `golang.org/x/perf/cmd/benchstat`(CLI tool)
