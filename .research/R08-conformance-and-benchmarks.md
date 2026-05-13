---
id: R08
title: 一致性测试与基准测试 — 跨项目模式与 go-intl 落地建议
task: r08 (ANALYSIS.md §8 / task.md §r08)
date: 2026-05-08
status: draft
scope:
  - 从 formatjs Vitest 测试中机械抽取 fixture 的可行性
  - FormatJS fixture 的格式统一
  - 已知 FormatJS 输出分歧的处理流程（divergences.md）
  - 基准测试 baseline 选择（`golang.org/x/text/message` vs FormatJS 公开数据）
  - 基准分层（含构造 vs 仅热路径）
  - CI 一致性策略与 `task verify` 集成
tags: [conformance, benchmark, fixture, divergence, ci, formatjs, icu, x-text]
---

## Executive Summary

| # | 主题 | 推荐结论 | 置信度 |
|---|-----|---------|--------|
| 1 | Fixture 主格式 | `testdata/conformance/<source>/<file>.json`，每行一个 `{id, locale, options, input, expected}` 对象，FormatToParts 用 `expectedParts: []`，FormatRange 用 `expectedRange: ""` | High |
| 2 | formatjs 抽取通道 | 简单 `expect(nf.format(x)).toBe(...)` 完全可机械化；`it.each(table)` 风格直接转表；Vitest `.snap` 通过 `parse-snapshot` 工具脚本提取；Date 字面量与回调需手写映射 | High |
| 3 | 已知分歧处理 | 每个 package 维护 `testdata/divergences.md`，结构化条目（source/case/our/reference/rationale）；CI 默认严格、列入 divergences 的条目作为豁免 | High |
| 4 | 基准 baseline | 主 baseline = `golang.org/x/text/message` + `golang.org/x/text/number`（Go 原生、可对比）；次 baseline = formatjs polyfill ops/sec 数据（截图，不在 CI 跑） | High |
| 6 | 基准分层 | 同时跑 `New(...).Format(...)`（含构造）和 `f.Format(...)`（缓存命中）两条曲线，对应 façade 与 per-formatter 两种使用模式 | High |
| 7 | CI 策略 | `task verify` 强制 conformance 通过；divergences.md 由 spec PR 显式审批；性能回归通过阈值断言（仿 `intl-localematcher/tests/benchmark.test.ts`），非阻塞但 PR-报警 | Medium |

> 总体方向：**fixture 多源 + 单一 Go 内部表达 + 显式分歧登记**，benchmark **双 baseline + 双层级**，CI 把"输出正确性"作为阻塞、把"性能回归"作为可观测信号。

---

## 1. Cross-cut · Fixture 抽取可行性

### 1.1 formatjs Vitest 测试的四种断言形态

通过遍历 `.references/formatjs/packages/intl-numberformat/tests/`、`intl-datetimeformat/tests/`、`intl-pluralrules/tests/`、`intl-locale/tests/`、`intl-localematcher/tests/`，可以归纳出四类 assertion，对机械抽取的友好程度依次递减：

| 形态 | 示例位置 | 抽取难度 | 抽取策略 |
|-----|---------|---------|---------|
| (A) 直接表驱动 (`it.each(table)`) | `intl-numberformat/tests/percent/percentTest.ts`、`intl-numberformat/tests/currency/`、`intl-numberformat/tests/unit/`、`intl-numberformat/tests/decimal/` | 极易 | 直接读 `testCombos` 数组，转 JSON |
| (B) 预定义 `tests` 数组 + 多 locale 字段 | `intl-datetimeformat/tests/format.test.ts`（`{options, ko, en}` 结构）、`intl-locale/tests/likely-subtags.test.ts`（`Record<string,string>`） | 易 | 解开 `Record` 或多 locale 字段为 `{locale, options, input, expected}` 行 |
| (C) `__snapshots__/*.snap` | `intl-numberformat/tests/percent/__snapshots__/en.test.ts.snap`、其他 `signDisplay-*.test.ts.snap`、`notation-compact-*.test.ts.snap` | 中 | 用脚本解析 Vitest `exports[\`<title>\`] = \`...\`` 格式；title 含 `notation=compact, signDisplay=always, compactDisplay=long 1` 之类的描述需要规整化 |
| (D) 内联 `expect(...)`（Date 字面量、回调、错误检测） | `intl-numberformat/tests/misc.test.ts`（61 处）、`intl-numberformat/tests/format_to_parts.test.ts`（50 处）、`intl-numberformat/tests/currency-code.test.ts`、`intl-pluralrules/tests/index.test.ts`（103 处）、`intl-datetimeformat/tests/format-range.test.ts`、`intl-datetimeformat/tests/offset-timezone.test.ts` | 难 | 手工归类，按 spec 含义重写为 fixture；错误用例归到 `errors.json` 单列 |

**抽取产出量级（基于 grep `expect(` 计数）：**

- `intl-numberformat/tests/format_to_parts.test.ts`: 50 处
- `intl-numberformat/tests/misc.test.ts`: 61 处
- `intl-pluralrules/tests/index.test.ts`: 103 处
- `intl-numberformat/tests/percent/__snapshots__/en.test.ts.snap` 等 snapshot 单文件 ≈ 数百 entry
- 全部 NumberFormat snapshot（en/zh/de/fr/ar/...）合计 4–6 K 条 fixture（数量级估算，需脚本核实）

### 1.2 不可机械抽取的部分

| 类别 | 来源 | 处理 |
|-----|------|------|
| Date 字面量 (`new Date(2020, 0, 1)`) | datetimeformat 全部测试 | 在 fixture 中用 ISO-8601 字符串（含偏移）+ `time.Time` 反序列化层。例：`{"input": "2020-01-01T00:00:00Z"}`，Go 端 `time.Parse(time.RFC3339, …)` |
| `__addLocaleData(en)` 注入 | NumberFormat polyfill 测试 | Go 端不需要；CLDR 数据已 `embed`，fixture 不带 locale-data 注入字段 |
| 错误断言（`expect(...).toThrow(...)`） | currency-code、misc、value-tonumber | 单独写到 `testdata/conformance/<source>/errors.json`，与 `errors.Is` / sentinel 对照 |
| `defaultRichTextElements` / React-only 用例 | `packages/intl/tests/` | 不抽取，归 messageformat-go |
| `Intl.NumberFormat.supportedLocalesOf` 在 `tests/legacy.test.ts` | 多个 polyfill | 转换为 `SupportedLocales(requested, supported)` 表 |

### 1.3 fixture 单元结构（建议）

每个 fixture 文件是 JSON 数组，每条元素结构如下（具体字段按 formatter 取舍）：

```json
{
  "id": "intl-numberformat/percent/en/notation=compact_signDisplay=always_long",
  "source": "formatjs:packages/intl-numberformat/tests/percent/__snapshots__/en.test.ts.snap",
  "locale": "en",
  "options": {"style": "percent", "notation": "compact", "signDisplay": "always", "compactDisplay": "long"},
  "input": 1.0,
  "expected": "+1 million",
  "expectedParts": [{"type": "plusSign", "value": "+"}, {"type": "integer", "value": "1"}, ...]
}
```

`source` 字段是审计来源；`id` 用于在 divergences.md 引用。Range 测试再加 `endInput` / `expectedRange` / `expectedRangeParts`。

---

## 2. Cross-cut · FormatJS fixture format

| Source | 角色 | Phase 1 用途 |
|--------|-----|-------------|
| formatjs Vitest 抽取 | **主 conformance**（数千条） | `task test` 必过 |
| `conformance-tests/icu4j/ICUConformanceTest.java` | **离线参考**（locale canonicalization、ScriptCode normalization、PluralRules selectRange、LocaleMatcher） | 不集成到 CI；定期手动 cross-check，结果手写为 JSON fixture |

> 置信度 **High** — FormatJS 是仓库内唯一实现参考，fixture provenance 单一更容易审计。

---

## 3. Cross-cut · 已知 formatjs vs V8 / ICU 分歧

### 3.1 分歧的来源类别

| 类别 | 典型例 | go-intl 处置 |
|-----|-------|-------------|
| Compact-notation 阈值 | formatjs `1.2K` 起算点与某些 locale ICU 阈值不同（见 NumberFormat README §4） | 跟 formatjs，记 divergences |
| Unit 复数形式 | `unit: 'hour', unitDisplay: 'long'` 在低使用 locale 下 V8 偶发 fallback 字串 | 跟 formatjs 主 fixture，标 v74/v75/v78 不一致 |
| Timezone abbreviation | `(中欧标准时间)` vs `(GMT+1)` — Node `dateStrings` v74→v78 个别 locale 改写 | 跟 v76.1，旧/新版本进 divergences |
| Locale canonicalization 大小写 | `en-arab-US` ⇒ `en-Arab-US`（ICU4J 已验证，见 `ICUConformanceTest.java:42-72`） | 强制对齐（属于 spec 必现） |
| `selectRange` 跨 plural-form | `(1, 1.5)` ⇒ `'one'` / `'few'` 取决 CLDR `pluralRanges.xml` | Phase 1 跟 CLDR 内置数据；偏移由 R05 报告管理 |
| Date 24h vs 12h locale 默认 | `en-US` 默认 12h，`en-GB` 默认 24h；formatjs 与 V8 在某些 locale 边界（`en-IN`）不一致 | 跟 formatjs，记 divergences |

### 3.2 divergences.md 结构（建议）

每个 formatter package 的 `testdata/divergences.md` 列举条目，每条字段：

| 字段 | 含义 |
|-----|------|
| `id` | 与 fixture `id` 字段一一对应 |
| `source` | `formatjs:` 或 `node:v76.1:` 或 `icu4j:` |
| `our` | go-intl 实际输出 |
| `reference` | 对照实现输出 |
| `category` | 上表中的类别 |
| `rationale` | 为何接受此分歧（CLDR 数据版本/spec 实现定义/已知 V8 bug） |
| `review_after` | 下次评审锚点（CLDR pin bump / Go 1.27 / 季度回顾） |

CI fixture runner 读取 divergences.md 的 `id` 列表，命中即跳过断言；任何不在列表中的失败都阻塞 `task verify`。

> 这个流程对齐 SPECS/00 §2.1 末段的"divergence handling"，实质化为 PR-审批契约。

---

## 4. Cross-cut · Benchmark Baseline 选择

### 4.1 候选基线对比

| 候选 | 可获取 | 输入语义对齐 | 维护成本 | 评分 |
|-----|--------|-------------|---------|------|
| `golang.org/x/text/message` + `golang.org/x/text/number` | go.dev 直接 import | NumberFormat decimal/currency/percent 大致对齐；compact/unit/scientific 部分覆盖 | 0（同 module 内即可写 benchmark） | **主 baseline** |
| formatjs polyfill ops/sec（`intl-numberformat/benchmark/README.md` 表格） | `.references/formatjs/packages/intl-numberformat/benchmark/` 已有数据 | 一致（同 spec） | 静态截图，不在 CI 跑 | **次 baseline（参考墙）** |
| Node `Intl.NumberFormat`（V8 native） | 需要 Node 进程 + JSON IPC | 完全一致 | 高（Bench harness、CI Node 工具链） | **不引入（YAGNI）** |
| `go-typescript` embed 后的 V8 Intl | 实验性、未必有 Phase 1 接入 | 一致 | 极高 | **拒绝**（不合 KISS） |

> 参考 formatjs 自家的对比方法：`intl-numberformat/benchmark/benchmark.ts:3-94` 用 tinybench 跑 polyfill 与 native 对照；其结果（README §"Benchmark Results"）显示 polyfill 比 V8 慢 6–21 倍。Go 端我们对 `x/text` 的目标是**同数量级**（2× 以内），不追平 V8 native。

### 4.2 ECMA-402 vs `x/text/message` 语义差距

`x/text/message` 不直接覆盖 ECMA-402 全部 option。基线对比仅覆盖三档：

| 比较项 | `x/text` API | go-intl API |
|-------|-------------|-------------|
| Decimal | `message.NewPrinter(tag).Sprintf("%v", n)` | `intl.FormatNumber(loc, n)` |
| Percent | `number.Percent(0.5)` 经 `message.Printf` | `intl.FormatNumber(loc, 0.5, intl.WithStyle("percent"))` |
| Currency | `number.NewFormat(number.Decimal, ...)` 不足；用 `golang.org/x/text/currency` 拼接 | `intl.FormatNumber(loc, 100, intl.WithCurrency("USD"))` |
| Compact / Unit / Scientific | 无对应 | go-intl 独占（无 baseline） |
| DateTimeFormat | `time.Time.Format(layout)` + `x/text/message` 拼接 | `intl.FormatDate(loc, t, ...)` |
| PluralRules | `golang.org/x/text/feature/plural`（见 R05） | `intl.SelectPlural(loc, n)` |

**结论**：基线只覆盖 NumberFormat decimal/percent/currency 与 PluralRules cardinal/ordinal；其余维度（DateTimeFormat、unit、compact）以"绝对吞吐 + 内存分配数"自基准。

### 4.3 基准目标与放置

| 维度 | 目标 |
|-----|------|
| `BenchmarkNumberFormat_Decimal_New` | go-intl 的 ops/sec ≥ `x/text/message` 的 1/2，b.Loop() 单值整数（参照 formatjs benchmark.ts:9 `testValues = [59, 0, ...]`） |
| `BenchmarkNumberFormat_Decimal_Cached` | 缓存命中时 ≥ `x/text/message` 的 1/1.5（构造摊销后）|
| `BenchmarkPluralRules_Cardinal_Cached` | ≥ `x/text/feature/plural` 的 1/1.5 |
| `BenchmarkDateTimeFormat_Short_Cached` | 仅 b.ReportAllocs()，无外部 baseline |

放置：`numberformat/benchmark_test.go`、`pluralrules/benchmark_test.go`、`datetimeformat/benchmark_test.go`，同包；不单独建 `benchmarks/` 顶层目录（YAGNI，且与 Go 测试惯例一致）。

---

## 5. Cross-cut · Benchmark 分层（含构造 vs 仅热路径）

### 5.1 formatjs/V8 的层级

formatjs `intl-numberformat/benchmark/benchmark.ts` 在 module 顶层一次性创建所有 NumberFormat 实例（`benchmark.ts:12-27`），bench 闭包里只调 `nf.format(val)`。这意味着 formatjs 的对外报告 **只衡量热路径**，不计构造成本。但实际生产中（如 React 渲染），开发者经常每次都 `new Intl.NumberFormat(...)`，因此 formatjs 又在 README 列出"hot path 优化"作为关键卖点。

### 5.2 go-intl 的双层设计

go-intl façade 本身就有缓存（见 R07），所以基准必须同时反映两条曲线：

| 曲线 | bench 名 | 含义 |
|-----|---------|------|
| 含构造 | `BenchmarkFormatNumber_PerCall` | `intl.FormatNumber(loc, n)` —— 每次都走 façade 的 cache lookup（首次 miss + 后续 hit 摊销） |
| 含构造（冷） | `BenchmarkNumberFormat_New` | `numberformat.New(loc, ...)` 单独基准 —— 只衡量 option 解析与 ResolvedOptions 落定 |
| 仅热路径 | `BenchmarkNumberFormat_Format` | 预先 `f := numberformat.New(...)`，循环 `f.Format(n)` —— 与 formatjs benchmark.ts 同口径 |
| 仅热路径（FormatToParts） | `BenchmarkNumberFormat_FormatToParts` | parts 切片分配（关注 `b.ReportAllocs()`） |

> 这与 R07 §3 推荐的"双轨 façade（per-call + persistent）"对齐：façade 缓存命中是 per-call 的真实热路径，per-formatter 调用是 persistent 的真实热路径。两者都该 publish。

### 5.3 性能回归阈值（仿 intl-localematcher）

`intl-localematcher/tests/benchmark.test.ts:21-55` 用 `expect(ms).toBeLessThan(0.01)` 等阈值断言把 benchmark 转成回归保护。go-intl 可以同样：

```go
func TestPerformanceRegression_Decimal(t *testing.T) {
    t.Parallel()
    f := numberformat.New(loc, intl.WithStyle("decimal"))
    res := testing.Benchmark(func(b *testing.B) {
        for b.Loop() { f.Format(42) }
    })
    if ns := res.NsPerOp(); ns > 5000 {
        t.Errorf("decimal hot path regressed: %d ns/op (budget 5000)", ns)
    }
}
```

> 这种断言放在 `*_test.go`（非 `*_bench_test.go`），通过 `task verify` 间接跑；阈值在 SPEC 中显式登记，CI 失败 = block。**置信度 Medium** — 阈值需要在 Phase 1 实现完成后用真实数据校准。

---

## 6. Cross-cut · CI 一致性策略

### 6.1 三级闸门（建议）

| 级别 | 作用 | 触发 | 行为 |
|-----|------|-----|------|
| **Gate 1: conformance** | 所有 fixture 必过（已在 divergences.md 登记的除外） | `task verify` ⇒ `go test -race -p 1 ./...` | 阻塞 |
| **Gate 2: divergence audit** | divergences.md 内容与代码一致（每条 `id` 都能找到对应 fixture） | `task verify` ⇒ 自定义 lint | 阻塞 |
| **Gate 3: performance regression** | hot-path benchmarks 不超过登记阈值 | `task verify` ⇒ `go test -bench=.` 后断言 | **PR 评论**（非阻塞），主分支 nightly 阻塞 |

### 6.2 与 `task verify` 集成

go-intl 当前 `Taskfile.yml` 已有 `verify` task（CLAUDE.md `## Commands` 区段）。建议串成：

```text
task verify
  ├── deps         (go mod tidy 验证)
  ├── fmt          (go fmt)
  ├── vet          (go vet)
  ├── lint         (golangci-lint v2)
  ├── test         (go test -race -p 1 ./...)  ← Gate 1 + Gate 2
  ├── vuln         (govulncheck)
  └── bench-check  (go test -bench -benchtime=1x -run=^$ ./... | check-thresholds)  ← Gate 3
```

`bench-check` 子任务读取 `bench.budgets.yaml`（per-package thresholds），任何 ns/op > budget 视为回归。该子任务在 PR-CI 中作为 `continue-on-error: true` 步骤，主分支夜间任务则严格。

### 6.3 fixture 来源审计（自动化）

`testdata/conformance/<source>/` 必须配套 `MANIFEST.json`，记录每个 fixture 文件的：

- `extracted_from`: formatjs commit SHA + 测试文件路径
- `extracted_at`: ISO 8601 时间戳
- `extractor_version`: 抽取脚本版本

PR 修改 fixture 必须同步更新 manifest；CI lint 阻塞缺失 manifest 的提交。这避免"fixture 漂移"。

> 置信度 **Medium** — manifest 机制相对重；初版可用 git blame 替代，等 PR 数量上来再加 lint。

---

## 7. Per-project Landing Recommendations

### 7.1 给 `numberformat/` 包

- `testdata/conformance/formatjs/`：抽取 `intl-numberformat/tests/` 全部 13 个测试文件 + 全部 `__snapshots__/`，分布到子目录 `decimal/`、`currency/`、`unit/`、`percent/`、`compact/`、`signDisplay/`、`format-range/`、`format-to-parts/`。
- `testdata/divergences.md`：占位条目 — currency precision 在 CLDR vs ISO 4217 选择导致的 0.5% 数据点差异（具体由实现阶段确认）。
- benchmark：8 个固定测试用例对齐 formatjs `benchmark.ts:9-27`（decimal、percent、currency、unit-long、significantDigits、fractionDigits、time-values-0-59、formatToParts），分 `_PerCall` / `_Cached` 两层。

### 7.2 给 `datetimeformat/` 包

- 必须为 Date 字面量建立 fixture-to-time 转换层：`{"input": "1980-07-25T00:35:33+00:00"}` ⇒ `time.Time` UTC ⇒ format。
- 抽取 `intl-datetimeformat/tests/format.test.ts`、`format-range.test.ts`、`offset-timezone.test.ts`、`abstract/skeleton.test.ts`。
- `testdata/divergences.md` 预期热点：timezone abbreviation（CLDR `Mídúl Yúrop Fíksd Taim`）在不同 locale 下版本敏感。

### 7.3 给 `pluralrules/` 包

- 抽取 `intl-pluralrules/tests/index.test.ts` 的 103 条 `expect`，转 `{locale, type, value, expected}` 行。
- selectRange fixture 单独一文件 `range.json`：`{startLocale, type, start, end, expected}`。
- benchmark：cardinal vs ordinal × 整数 / decimal / bigint / compact-mantissa 共 12 个 sub-bench。

### 7.4 给 `locale/` 包

- 抽取 `intl-locale/tests/likely-subtags.test.ts`（`Record<string,string>` 直接转 JSON）、`minimize.test.ts`、`intl-localematcher/tests/locale-match-fixtures.json`（直接复用，schema 已是 `{description, requested, supported, expected}`）。
- ICU4J `ICUConformanceTest.java:42-72` 的 8 条 canonicalization 用例 + script normalization 用例手抄成 fixture。
- benchmark：`intl-localematcher/tests/benchmark.test.ts` 的 6 层（exact / fallback / maximized / es-419 / sr-Latn-BA / fuzzy）作为 baseline，go-intl 复刻并按 microsecond 阈值断言。

### 7.5 给 façade（根包）

- 不写独立 conformance fixture（依赖 per-formatter）。
- benchmark：复刻 R07 §3 的 cache 命中 / miss 对比 — `BenchmarkFormatNumber_FaCacheHit` vs `BenchmarkFormatNumber_FaCacheMiss`，同时与 `numberformat.New(...).Format(...)` 对比验证 façade 不引入显著开销。

### 7.6 给 `messageformat-go` 集成

- messageformat-go 测试套件目前自含 ECMA-402 测试（pkg/functions/*_test.go）；迁移到 go-intl 后，**保留** messageformat-go 的 function-context 集成测试，**删除** 重复的 ECMA-402 conformance 测试（go-intl 已覆盖）。
- 双方共享 fixture 集：`messageformat-go/internal/testdata/intl-bridge/*.json` 引用 go-intl 的 `testdata/conformance/` 路径（symlink 或 go submodule 引用）— 避免 fixture fork。

---

## 8. Decision Matrix

| 决策 | 选项 | 推荐 | 置信度 | 关键证据 |
|-----|------|------|--------|---------|
| 主 fixture source | formatjs / Node localizationData / 双源 | **双源（formatjs 主，Node 跨版本验证）** | High | SPECS/00 §2.1 已明确双源；formatjs 覆盖广，Node 覆盖跨 ICU 版本 |
| Node 锚定 ICU 版本 | v74.2 / v75.1 / **v76.1** / v78.2 | **v76.1**（CLDR 46） | High | CLDR 46 与 `golang.org/x/text` 当前 CLDR pin 接近；v78.2 太前沿 |
| Snapshot 抽取方法 | 手工 / **脚本解析** / 重写为 it.each | **脚本解析 `.snap`** | High | Vitest snapshot 格式机器可读（`exports[\`title\`] = \`...\`;`），见 `intl-numberformat/tests/percent/__snapshots__/en.test.ts.snap:3-22` |
| 错误用例 fixture | 与正向用例同文件 / **errors.json 单列** | **errors.json 单列** | High | Go errors.Is/As 校验路径与字符串校验路径不同，分文件减小 harness 复杂度 |
| 主 benchmark baseline | x/text / formatjs ops/sec / V8 / 无 | **`x/text/message` + `x/text/number` + `x/text/feature/plural`** | High | Go 原生、零工具链成本；语义对齐三档 |
| 基准分层 | 仅热路径 / 仅含构造 / **两条** | **两条都跑** | High | 与 R07 双轨 façade 对齐；formatjs `benchmark.ts:12-27` 也只测热路径，但 README 写明这是优化目标 |
| 性能回归 CI | 阻塞 / **阈值告警** / 不集成 | **阈值告警（PR 非阻塞，main 阻塞）** | Medium | 类似 `intl-localematcher/tests/benchmark.test.ts`；阈值需用真实数据校准 |
| ICU4J 集成 | CI 自动化 / **离线参考** / 不用 | **离线参考** | High | Java 工具链开销大；定期手动 cross-check 即可，结果手抄 fixture |
| divergences.md 评审 | 自动登记 / **PR 显式审批** / 无 | **PR 显式审批** | High | divergence 涉及 spec 解释权，必须人审；CI 校验 `id` 完整性 |

---

## 9. Code Block Index

> 仅列签名/调用样例，便于实现期参考；非生产代码。

### 9.1 fixture loader 调用样例

```go
// numberformat/conformance_test.go
func TestNumberFormat_Conformance(t *testing.T) {
    t.Parallel()
    cases := loadConformance(t, "formatjs/percent/en.json")
    for _, c := range cases {
        t.Run(c.ID, func(t *testing.T) { /* assert */ })
    }
}
```

### 9.2 divergence skip 样例

```go
// internal/conformance/runner.go
type Runner struct{ skipped map[string]string /* id → rationale */ }
func (r *Runner) Run(t *testing.T, c Case, fn func(Case) (string, error)) { /* skip if c.ID in r.skipped */ }
```

### 9.3 benchmark 双层样例

```go
// numberformat/benchmark_test.go
func BenchmarkNumberFormat_Decimal_PerCall(b *testing.B)  { for b.Loop() { intl.FormatNumber(en, 42) } }
func BenchmarkNumberFormat_Decimal_Cached(b *testing.B)   { f := numberformat.New(en); b.ResetTimer(); for b.Loop() { f.Format(42) } }
```

### 9.4 性能阈值断言样例

```go
// 仿 intl-localematcher/tests/benchmark.test.ts:21-55
func TestPerf_DecimalCached(t *testing.T) { /* run benchmark, assert NsPerOp() < budget */ }
```

### 9.5 baseline 对照样例

```go
// numberformat/benchmark_baseline_test.go
func BenchmarkBaseline_XText_Decimal(b *testing.B) { p := message.NewPrinter(language.English); for b.Loop() { p.Sprintf("%v", 42) } }
```

---

## 10. Citation List

### formatjs（主 conformance 来源）

- `.references/formatjs/packages/intl-numberformat/tests/format_to_parts.test.ts` — 50 处 `expect`，FormatToParts 主测试
- `.references/formatjs/packages/intl-numberformat/tests/misc.test.ts` — 61 处 `expect`，杂项 NumberFormat 行为
- `.references/formatjs/packages/intl-numberformat/tests/currency-code.test.ts` — 货币代码校验
- `.references/formatjs/packages/intl-numberformat/tests/percent/percentTest.ts` — `it.each(testCombos)` 表驱动模板
- `.references/formatjs/packages/intl-numberformat/tests/percent/__snapshots__/en.test.ts.snap:3-22` — Vitest snapshot 格式样例
- `.references/formatjs/packages/intl-numberformat/tests/notation-compact-ko-KR.test.ts`、`notation-compact-zh-TW.test.ts` — compact-notation locale 测试
- `.references/formatjs/packages/intl-numberformat/tests/legacy.test.ts`、`fast-path-optimizations.test.ts`、`value-tonumber.test.ts`、`signDisplay-currency-zh-TW.test.ts`、`signDisplay-zh-TW.test.ts`、`unit-zh-TW.test.ts` — locale-specific 用例
- `.references/formatjs/packages/intl-numberformat/tests/{currency,decimal,unit}/` — 子目录表驱动测试
- `.references/formatjs/packages/intl-datetimeformat/tests/format.test.ts` — `tests` 数组 `{options, ko, en}` 结构
- `.references/formatjs/packages/intl-datetimeformat/tests/format-range.test.ts` — formatRange 与 formatRangeToParts
- `.references/formatjs/packages/intl-datetimeformat/tests/offset-timezone.test.ts` — UTC 偏移时区
- `.references/formatjs/packages/intl-datetimeformat/tests/abstract/skeleton.test.ts` — splitRangePattern 单元测试
- `.references/formatjs/packages/intl-pluralrules/tests/index.test.ts` — 103 处 `expect`，cardinal/ordinal/bigint
- `.references/formatjs/packages/intl-locale/tests/likely-subtags.test.ts` — `Record<string,string>` maximize 表
- `.references/formatjs/packages/intl-locale/tests/minimize.test.ts` — minimize 表
- `.references/formatjs/packages/intl-localematcher/tests/conformance.test.ts` — JSON fixture loop pattern
- `.references/formatjs/packages/intl-localematcher/tests/locale-match-fixtures.json:1-40` — `{description, requested, supported, expected}` schema
- `.references/formatjs/packages/intl-localematcher/tests/benchmark.test.ts:21-55` — 性能回归阈值断言模式

### formatjs（benchmark 参考）

- `.references/formatjs/packages/intl-numberformat/benchmark/benchmark.ts:1-94` — tinybench polyfill vs native 对照
- `.references/formatjs/packages/intl-numberformat/benchmark/README.md:62-138` — 优化前后 ops/sec 对比、关键观察、Decimal.js fast-path 优化思路
- `.references/formatjs/benchmarks/cli-comparison/` — CLI 性能对比目录（Phase 3+ 用）

### formatjs（ICU4J 离线参考）

- `.references/formatjs/conformance-tests/icu4j/ICUConformanceTest.java:1-80` — locale canonicalization、script normalization、PluralRules selectRange、LocaleMatcher 7 大测试入口

### go-intl 项目内部

- `SPECS/00-vision-and-scope.md:69-94` — §2.1 test fixture policy（双源、porting flow、divergence handling）
- `task.md:329-376` — r08 任务定义与检查点
- `ANALYSIS.md` §8 — 模块研究指引

### 外部 baseline（go.dev / 文档）

- `golang.org/x/text/message` — 主 baseline（NumberFormat decimal/percent/currency 部分覆盖）
- `golang.org/x/text/number` — Decimal/Percent baseline
- `golang.org/x/text/feature/plural` — PluralRules baseline（详见 R05）
- `golang.org/x/text/currency` — 货币符号 baseline（go-intl currency 抽取的辅助）

### 同源研究报告

- `.research/R07-facade-and-caching.md` — façade + 缓存 + 错误模型；R08 的双层 benchmark 与 R07 的双轨 façade 对应
- `.research/R03-numberformat.md`（已完成） — Decimal 选型与 NumberFormat option pipeline
- `.research/R04-datetimeformat.md`（已完成） — skeleton / timezone / calendar 决策
- `.research/R05-pluralrules.md`（已完成） — codegen vs interpreter 与 selectRange 数据流
- `.research/R06-cldr-data-strategy.md`（已完成） — CLDR/ICU 版本 pin 决策（与本报告 §3.3 互验）
