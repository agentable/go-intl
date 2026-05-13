# SPEC 71 — Benchmark Strategy & Performance Gates

> **Status:** Draft (2026-05-08)
> **Type:** Rule + Flow — defines how go-intl measures hot-path performance and gates regressions in CI.
> **Authority:** This spec is SSOT for benchmark layout, baseline selection, performance thresholds, and `benchstat` report flow. SPEC 70 owns conformance Gate 3 invocation; this spec owns the *content* of Gate 3.

---

## Overview

go-intl 的性能信号由两层基准构成:

1. **PerCall**(含 `New`):反映 ECMA-402 构造器路径,即 `new Intl.<Constructor>(locales, options)` 后立即调用一次方法的成本。
2. **Cached**(仅方法调用):反映 formatter 已构造后的纯 hot-path。

baseline = `golang.org/x/text/message` + `golang.org/x/text/feature/plural`(stdlib-grade Go 实现);**禁止** 通过 embedded JS engines 作 baseline(R08 §4.1 已论证 ROI 不足)。

> **Why**: ECMA-402 同时有构造期成本与方法 hot-path 成本。双层基准让 option resolution / locale negotiation 与格式化主路径分别可见。
> **Rejected**: root one-shot / global-cache benchmark 作为主信号——JavaScript `Intl` namespace 没有 per-locale session 或 root one-shot helpers,这类 API 不属于长期公开 surface。

---

## 1. Benchmark Layout

### 1.1 文件位置

| 包 | 文件 | 内容 |
|----|------|------|
| `numberformat/` | `numberformat/benchmark_test.go` | NumberFormat PerCall + Cached |
| `datetimeformat/` | `datetimeformat/benchmark_test.go` | DateTimeFormat PerCall + Cached |
| `pluralrules/` | `pluralrules/benchmark_test.go` | PluralRules Cardinal/Ordinal Cached(PerCall 可省) |
| `locale/` | `locale/benchmark_test.go` | `locale.Parse(string)` 边界解析 + `locale.New(language.Tag, Options{})` 构造 + `BestFitMatcher` |
| baseline | `<package>/benchmark_baseline_test.go` | `x/text/message` / `x/text/feature/plural` 对照 |

**必须遵守**:

1. benchmark **必须**与 conformance test 同包(`*_test.go`),不另建 `benchmarks/` 顶层目录。
2. baseline benchmark **必须**位于 `_baseline_test.go`(独立文件),便于 `go test -bench=Baseline` 单独跑。
3. **禁止** 在 main package 编写 benchmark(违反 Go 测试惯例,且无法用 `_test.go` build tag 隔离 baseline)。

> **Why**: 同包 benchmark 让 godoc / IDE 跳转无缝,与 stdlib 习惯一致。
> **Rejected**: `bench/` 顶层目录——R08 §4.3 已论证;额外目录无收益。

### 1.2 命名约定

```go
// PerCall(含构造)
func BenchmarkNumberFormat_Decimal_PerCall(b *testing.B)
func BenchmarkNumberFormat_Currency_PerCall(b *testing.B)

// Cached(纯 hot-path)
func BenchmarkNumberFormat_Decimal_Cached(b *testing.B)
func BenchmarkNumberFormat_Currency_Cached(b *testing.B)
func BenchmarkNumberFormat_Compact_Cached(b *testing.B)
func BenchmarkNumberFormat_FormatToParts_Cached(b *testing.B)

// PluralRules
func BenchmarkPluralRules_Cardinal_Cached(b *testing.B)
func BenchmarkPluralRules_Ordinal_Cached(b *testing.B)
func BenchmarkPluralRules_SelectRange_Cached(b *testing.B)

// DateTimeFormat
func BenchmarkDateTimeFormat_DateStyleShort_Cached(b *testing.B)
func BenchmarkDateTimeFormat_DateTimeRange_Cached(b *testing.B)
func BenchmarkDateTimeFormat_FormatToParts_Cached(b *testing.B)

// 构造期(单独度量)
func BenchmarkNumberFormat_New(b *testing.B)
func BenchmarkDateTimeFormat_New(b *testing.B)

// Baseline(独立文件)
func BenchmarkBaseline_XText_Decimal(b *testing.B)
func BenchmarkBaseline_XText_Plural_Cardinal(b *testing.B)
```

**必须遵守**:

1. benchmark 函数命名 **必须**为 `Benchmark<Type>_<Feature>_<Layer>` 三段式;**禁止** 隐式简写(如 `BenchmarkNF`)。
2. baseline benchmark **必须**前缀 `BenchmarkBaseline_`,与生产 benchmark 视觉区分。
3. PerCall benchmark **必须**包含完整 `numberformat.New(...)` 调用;Cached benchmark **必须**在 b.Loop() 之前完成构造。

### 1.3 b.Loop 强制使用

```go
func BenchmarkNumberFormat_Decimal_Cached(b *testing.B) {
    b.ReportAllocs()
    nf, _ := numberformat.New(en, numberformat.Options{Style: numberformat.DecimalStyle})
    b.Loop()
    for b.Loop() {
        _ = nf.FormatInt(42)
    }
}
```

**必须遵守**:

1. **必须**使用 Go 1.24+ `b.Loop()`(项目基线 Go 1.26.2)。**禁止** 用旧式 `for i := 0; i < b.N; i++` 循环。
2. Cached benchmark **必须**在 `b.Loop()` 之前完成构造;**禁止**在循环体内 `New(...)`(那是 PerCall 的范畴)。
3. 所有 benchmark **必须**调用 `b.ReportAllocs()`,在 ns/op 之外报告 B/op + allocs/op。
4. **禁止** 在 benchmark 内调用 `time.Now()` / `runtime.GC()` 等扰动测量。

> **Why**: `b.Loop()` 是 Go 1.24+ 标准 benchmark loop API,自动处理 keepalive 与去 inline,比 `b.N` 循环少踩坑。
> **Rejected**: 旧式 `for i := 0; i < b.N; i++`——新 benchmark 统一使用 `b.Loop()`。

---

## 2. Baseline Selection

### 2.1 主 baseline:`golang.org/x/text`

| 比较项 | x/text API | go-intl API | baseline 文件 |
|-------|-----------|-------------|--------------|
| Decimal | `message.NewPrinter(tag).Sprintf("%v", n)` | `numberformat.New(loc, opts).FormatInt(n)` | `numberformat/benchmark_baseline_test.go` |
| Percent | `number.Percent(0.5)` 经 `message.Printf` | `numberformat.New(loc, numberformat.Options{Style: numberformat.PercentStyle}).FormatFloat64(0.5)` | 同上 |
| PluralCardinal | `plural.Cardinal.Matches(tag, ...)` | `pluralrules.New(loc, opts).SelectInt(n)` | `pluralrules/benchmark_baseline_test.go` |
| PluralOrdinal | `plural.Ordinal.Matches(tag, ...)` | `pluralrules.New(loc, pluralrules.Options{Type: pluralrules.Ordinal}).SelectInt(n)` | 同上 |

**必须遵守**:

1. 主 baseline **必须**是 `golang.org/x/text/message` + `golang.org/x/text/feature/plural`;两者都已在 SPEC 50 / SPEC 40 出现(作为对比对象,**非**生产依赖)。
2. baseline 仅覆盖 **decimal / percent / plural cardinal / plural ordinal** 四档;currency / unit / compact / DateTimeFormat / Range 由于 x/text 无对应 API,**不**参与 baseline 对比,改为"绝对吞吐 + 内存分配"自基准。
3. baseline 与 go-intl 同跑同输入(同 locale、同 input 值);**禁止** 用不同 input 制造误导对比。

> **Why**: x/text 是 Go 原生、零工具链、官方维护;同 module 内 import 即可写 benchmark,与 go-intl 部署假设(同进程)一致。
> **Rejected**: Cross-process JS baseline——R08 §4.1 已论证维护成本;且 IPC 跨进程开销污染信号。
> **Rejected**: formatjs polyfill ops/sec(README §"Benchmark Results")——它已有,但截图状态,不在 CI 跑;作为参考墙,**不**进入 CI gate。

### 2.2 禁用 embedded JS baseline

**禁止** 通过 embedded JS engine 跑 `Intl.NumberFormat` 作 baseline。理由:

1. R09 §4.8:go-typescript 是 consumer-driven expansion 候选,与 active scope 时间线不重叠。
2. embedded JS engine 启动开销 ≥ 100ms,benchmark 信号被启动支配,无法测量稳态。
3. 维护成本 = JS engine 工具链 + 版本钉化 + JS-Go marshalling 校验;ROI 与 go-intl active scope 目标不匹配。
4. 字节相等性已由 SPEC 70 fixture 覆盖;性能对比"同数量级"目标足以由 x/text baseline + 绝对吞吐保证。

> **Why**: Embedded JS baseline 带来 4 个二阶问题(启动 / 工具链 / 版本钉化 / marshalling),换 1 个可疑收益,YAGNI 反此。
> **Rejected**: Embedded JS baseline——明确拒绝,任何二期 PR 提议都 **必须**先修订本 SPEC §2.2。

### 2.3 性能目标(对比 baseline)

| 维度 | 目标(对比 baseline) | 备注 |
|-----|---------------------|------|
| `BenchmarkNumberFormat_Decimal_Cached` vs `BenchmarkBaseline_XText_Decimal` | go-intl ≥ baseline / 1.5 | 缓存命中后摊销构造 |
| `BenchmarkNumberFormat_Decimal_PerCall` vs baseline | go-intl ≥ baseline / 2.0 | 含构造,允许 2× 慢 |
| `BenchmarkPluralRules_Cardinal_Cached` vs baseline | go-intl ≥ baseline / 1.5 | codegen 路径应接近 stdlib |
| `BenchmarkDateTimeFormat_*` | 仅 `b.ReportAllocs()`,无 baseline | x/text 无对应 API |
| `BenchmarkNumberFormat_Compact_Cached` | 仅 `b.ReportAllocs()`,无 baseline | x/text 无 compact |
| `BenchmarkNumberFormat_Currency_Cached` | 仅 `b.ReportAllocs()`,无 baseline | x/text/currency 形态不对应 |

> **Why**: "同数量级"目标(2× 以内)足以让 go-intl 在生产环境与 x/text 无明显劣势;追平 stdlib 是不切实际的(go-intl 多了 ECMA-402 spec 行为)。
> **Rejected**: 1× 追平——会让 SPEC 持续摩擦实现细节,反而推迟 active scope 发布。

---

## 3. Performance Thresholds (Gate 3)

### 3.1 阈值表

下列阈值在 PR 仅 warn,在 main 分支 nightly job block(SPEC 70 §4 Gate 3)。

| Benchmark | 阈值(ns/op) | 来源 / 校准 |
|-----------|--------------|-------------|
| `BenchmarkPluralRules_Cardinal_Cached` | ≤ **200** | codegen 直查表,接近 stdlib 量级 |
| `BenchmarkPluralRules_Ordinal_Cached` | ≤ **250** | 同上,但 ordinal 规则更长 |
| `BenchmarkPluralRules_SelectRange_Cached` | ≤ **400** | 双 form lookup + range table |
| `BenchmarkNumberFormat_Decimal_Cached` | ≤ **800** | 数字格式化主流形态 |
| `BenchmarkNumberFormat_Currency_Cached` | ≤ **1200** | + currency 符号查表 |
| `BenchmarkNumberFormat_Compact_Cached` | ≤ **1500** | + plural lookup + compact 表 |
| `BenchmarkNumberFormat_FormatToParts_Cached` | ≤ **1800** | + parts slice 分配 |
| `BenchmarkDateTimeFormat_DateStyleShort_Cached` | ≤ **2000**(2 μs) | skeleton resolution + tz |
| `BenchmarkDateTimeFormat_DateTimeRange_Cached` | ≤ **3500** | range pattern + collapse |
| `BenchmarkDateTimeFormat_FormatToParts_Cached` | ≤ **3000** | parts slice 分配 |
| `BenchmarkNumberFormat_New` | ≤ **5000** | 构造期含 option 校验 + ResolvedOptions |
| `BenchmarkDateTimeFormat_New` | ≤ **8000** | + skeleton 解析 + tz cache |

**必须遵守**:

1. 上述阈值在 **实施期** 完成首轮 benchmark 后**必须** 校准——以 baseline 实测值为锚点,go-intl 目标值按 §2.3 比例倒推。校准结果 **必须**通过 PR 修订本 SPEC §3.1 表格固化(并删除"实施期校准"标注)。
2. PR-CI **必须**仅 warn(评论 PR diff);main nightly job **必须**用 `benchstat` 跑 ≥ 5 次 benchmark 计算 statistically significant regression(p < 0.05)后 block。
3. 单 PR 性能 diff 阈值: **+5%**(warn)/ **+15%**(main block)。
4. 阈值表中的绝对值是 **环境依赖** 的——`benchstat` 报告 **必须**记录 CI runner 规格(CPU / Go version / GOARCH)。
5. 任何 benchmark 新增 **必须**配套阈值表条目;**禁止** "无阈值的孤儿 benchmark"。

> **Why**: 阈值绝对值是为开发者提供锚点而非教条——CI runner 异构环境下绝对值难以稳定,因此用 `benchstat` 统计判定 + main block 防止"飞秒级"摩擦。
> **Rejected**: PR 直接阻塞 +5% regression——单 PR 噪音 ≥ 5% 是常态,会让 reviewer 习惯性 force-merge,规则失效。

### 3.2 内存分配预算

下列 benchmark **必须**校验 `b.ReportAllocs()` 输出:

| Benchmark | allocs/op 阈值 | 备注 |
|-----------|---------------|------|
| `BenchmarkNumberFormat_Decimal_Cached` | ≤ **3** | 输出 string + 内部 buffer + 1 spare |
| `BenchmarkNumberFormat_FormatToParts_Cached` | ≤ **8** | parts slice + 每 part 1 alloc |
| `BenchmarkPluralRules_Cardinal_Cached` | ≤ **0**(零分配) | 直查表,纯枚举返回 |
| `BenchmarkDateTimeFormat_DateStyleShort_Cached` | ≤ **5** | 日期 buffer + tz lookup |

**必须遵守**:

1. 内存预算 **必须**与 ns/op 阈值同时校验;违反任一即 main block。
2. PluralRules `Select` **必须**零分配——它是 messageformat-go 高频调用,任何分配都放大到消息渲染热路径。

> **Why**: ns/op 与 allocs/op 是双信号——无 alloc 的快路径偶尔会被 GC 拖慢,只看 ns/op 会漏报内存退化。
> **Rejected**: 仅 ns/op——formatjs polyfill 的关键卖点之一就是 hot-path 零分配,go-intl 必须保持竞争力。

---

## 4. benchstat Workflow

### 4.1 工具

`benchstat` 是 `golang.org/x/perf/cmd/benchstat` 的 CLI 工具,**非** go.mod 依赖。

**必须遵守**:

1. CI runner **必须**通过 `go install golang.org/x/perf/cmd/benchstat@latest` 安装(独立命令,非 go.mod require)。
2. **禁止** 把 `golang.org/x/perf` 加入 `go.mod`(R09 §1 已论证 active scope 直接依赖严格收口)。
3. benchstat 版本 **应当** 钉死在 `Taskfile.yml`(`benchstat@v0.0.0-2025xxxxxxxx`),防止 CI 行为漂移。

### 4.2 报告格式

`task bench` 子任务 **必须**:

1. 跑全套 benchmark ≥ 5 次(默认 `-count=5`)。
2. 输出 `bench-current.txt`(当前分支)与 `bench-baseline.txt`(对比分支,默认 main)。
3. 调用 `benchstat bench-baseline.txt bench-current.txt` 生成报告。
4. 报告格式 **必须**包含:基准名、N、ns/op delta、p-value、B/op delta、allocs/op delta。
5. PR-CI **必须**把报告 post 到 PR 评论区(GitHub `actions/github-script` 或等价机制)。
6. main nightly job **必须**保留报告 artifact ≥ 30 天。

### 4.3 阈值断言示例

```go
// 仿 intl-localematcher/tests/benchmark.test.ts:21-55
// 仅在 main 分支跑;PR 不阻塞。
func TestPerf_Decimal_Cached(t *testing.T) {
    if testing.Short() { t.Skip() }
    nf, _ := numberformat.New(en, numberformat.Options{Style: numberformat.DecimalStyle})
    res := testing.Benchmark(func(b *testing.B) {
        for b.Loop() { _ = nf.FormatInt(42) }
    })
    const budget = 800
    if ns := res.NsPerOp(); ns > budget {
        t.Errorf("decimal cached regressed: %d ns/op (budget %d)", ns, budget)
    }
}
```

**必须遵守**:

1. 阈值断言 test 函数 **必须**前缀 `TestPerf_`,与 conformance test 视觉区分。
2. 阈值断言 **必须**支持 `testing.Short()` 跳过——`go test -short` 不跑 perf gate。
3. 单测试函数仅断言一个 benchmark;**禁止** 多 benchmark 共享一个测试函数(失败定位困难)。

> **Why**: PR 不跑阈值断言(走 benchstat 评论);main nightly 才跑阈值断言(防 noisy 摩擦)。
> **Rejected**: PR 直接 `TestPerf_*` 阻塞——单次运行噪音大,benchstat 多次取统计才有信号。

---

## 5. Forbidden

- **禁止** 旧式 `for i := 0; i < b.N; i++` 循环(必须 `b.Loop()`)。
- **禁止** 在 Cached benchmark 内调用 `numberformat.New(...)`(那是 PerCall 范畴)。
- **禁止** 把 `golang.org/x/perf` 加入 `go.mod`(benchstat 是 CLI,非 go-import 依赖)。
- **禁止** 通过 embedded JS engine 作 baseline(SPEC §2.2)。
- **禁止** active scope 引入 `clipperhouse/uax29`、`bojanz/currency` 等 consumer-driven expansion 候选作 baseline——baseline 仅 `x/text`。
- **禁止** 单 PR 直接以 +5% regression 阻塞(必须走 benchstat 统计 + main nightly)。
- **禁止** 阈值表条目缺失对应 benchmark 函数(孤儿阈值)。
- **禁止** benchmark 函数缺失对应阈值表条目(孤儿 benchmark)。
- **禁止** baseline benchmark 与生产 benchmark 同文件(必须 `_baseline_test.go` 独立)。
- **禁止** 在 benchmark 内调用 `time.Now()` / `runtime.GC()` 扰动测量。
- **禁止** PluralRules `Select` 在 hot-path 出现任何 heap allocation。

---

## 6. Acceptance Criteria

### Layout

- [ ] 每 formatter 包存在 `benchmark_test.go` 与 `benchmark_baseline_test.go`(后者仅 NumberFormat / PluralRules)。
- [ ] benchmark 命名遵循 `Benchmark<Type>_<Feature>_<Layer>` 三段式。
- [ ] 全部 benchmark 使用 `b.Loop()` 与 `b.ReportAllocs()`。
- [ ] baseline benchmark 前缀 `BenchmarkBaseline_`。

### Baseline

- [ ] `BenchmarkBaseline_XText_Decimal` 存在于 `numberformat/`。
- [ ] `BenchmarkBaseline_XText_Plural_Cardinal` / `Ordinal` 存在于 `pluralrules/`。
- [ ] `go.mod` 不含 embedded JS engine 相关依赖。

### Thresholds

- [ ] SPEC §3.1 阈值表在实施期完成首轮校准,并通过 PR 固化(删除"实施期校准"标注)。
- [ ] 每条阈值表条目对应一个 benchmark 函数,反之亦然(无孤儿)。
- [ ] PluralRules `Select` 在 `b.ReportAllocs()` 下报告 `0 allocs/op`。

### CI Integration

- [ ] `task bench` 子任务存在,跑 ≥ 5 次 benchmark + benchstat 报告。
- [ ] PR-CI 把 benchstat 报告 post 到 PR 评论(warn,不阻塞)。
- [ ] `task bench` 输出 benchstat 报告;CI 在 main 分支检测回归 ≥ 5%(p < 0.05)时 block。
- [ ] main nightly job 保留 benchstat artifact ≥ 30 天。

### Tooling

- [ ] benchstat 通过 `go install` 安装;**不**在 `go.mod` 出现。
- [ ] CI runner 规格(CPU / Go version / GOARCH)记录在 benchstat 报告头部。

---

## References

- SPECS/00 §7(consumer-driven expansion 优化锚点)
- SPEC 60(root `Intl` namespace; forbids root one-shot helpers as benchmark ownership)
- SPEC 70 §4(Gate 3 触发本 SPEC 阈值)
- `.research/R08-conformance-and-benchmarks.md` §4–§5
- `.research/R09-dependencies.md` §4.9(test-only deps)、§4.8(consumer-driven expansion 候选,baseline 不含)
- `.references/formatjs/packages/intl-numberformat/benchmark/benchmark.ts:1-94`(双层基准灵感)
- `.references/formatjs/packages/intl-localematcher/tests/benchmark.test.ts:21-55`(阈值断言模式)
- `golang.org/x/text/message`(主 baseline)
- `golang.org/x/text/feature/plural`(主 baseline)
- `golang.org/x/perf/cmd/benchstat`(CLI 工具)
