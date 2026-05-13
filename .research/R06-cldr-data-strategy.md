---
id: R06
title: CLDR 数据采集、生成与嵌入策略调研
task: r06
date: 2026-05-08
status: draft
scope:
  - CLDR 数据源选型（Unicode cldr-json / x/text/cldr / formatjs 中间产物）
  - 嵌入策略（//go:embed JSON vs Go 字面量生成）
  - per-locale tree-shaking 时机
  - CLDR / ICU 版本钉决策（v74.2 / v75.1 / v76.1 / v78.2）
  - 货币、时区、单位标识符的来源
  - 生成器架构（tools/gen-cldr/）
tags: [cldr, codegen, data-embedding, version-pin]
---

# R06 — CLDR 数据采集、生成与嵌入策略调研

## 执行摘要

| 决策点 | 推荐 | 置信度 | 关键依据 |
|--------|------|--------|----------|
| CLDR 数据源 | 直接消费 Unicode `cldr-json`（npm 镜像 `cldr-dates-full` / `cldr-numbers-full` / `cldr-misc-full` / `cldr-core` / `cldr-bcp47`），与 formatjs 同源 | 高 | formatjs 在 `package.json` 显式 pin `cldr-*: 48.1.0`；`golang.org/x/text/cldr` 提供 LDML XML 但停在更早版本，且不暴露给生产消费 |
| 嵌入策略 | Go 字面量（生成 `internal/cldr/*.go`），不走 `//go:embed JSON`。文本数据共享一个 `const data = "..."` 全局字符串 + 索引表 | 高 | translate-agent 的 `data.go` 是 400 KB 单文件、3203 行 Go 字面量，运行时零解析；`//go:embed` JSON 在 100 多个 locale 下需要冷启动反序列化 |
| Phase 1 locale 范围 | 全量嵌入约 100 个常用 locale；预留 build tag `intl_full`（500 locale）与 `intl_minimal`（仅 en）作为 Phase 4 的优化路径 | 中 | translate-agent 已经全量嵌入（400 KB），还在可接受范围；formatjs 的 per-locale 拆包在 Go 上的 ROI 受 binary 编辑成本（每个 locale 一个 init 函数）限制 |
| CLDR / ICU 版本钉 | **CLDR 48.1.0 / ICU 78** | 高 | formatjs 当前主分支锁定 `cldr-*: 48.1.0`，与 ICU 78 对齐；钉早期版本会在与 formatjs 对齐时出现"我们对、formatjs 错"的反向 divergence |
| 标识符来源 | 货币：CLDR `currencyData` + ISO 4217（CLDR 已包含）；时区：IANA + CLDR `metaZones`；单位：CLDR `units` + UN/CEFACT 子集（`units-constants.ts` 的 sanctioned 列表） | 高 | formatjs 三类标识符全走 CLDR，未引入独立 ISO/IANA fetch；统一来源减少版本漂移 |
| 生成器位置 | 仓库内子包 `tools/gen-cldr/`，独立 `go.mod`（不污染主 module 的依赖图） | 高 | translate-agent 用 `internal/gen/`（同 module，依赖少）；我们因为依赖 npm 包链路稍重，但仍建议同仓库以同步 PR 流；分独立 module 避免 `cmd/...` 污染主 binary |

## 1. CLDR 数据源选型

### 1.1 三条路径对比

**路径 A：直接读 Unicode `cldr-json` 仓库 / npm 镜像**

formatjs 走的就是这条路<!-- ref: formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts:7-30 -->：

```ts
// 提炼自 .references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts:7-30
import DateFields from 'cldr-dates-full/main/en/ca-gregorian.json' with {type: 'json'}
import NumberFields from 'cldr-numbers-full/main/en/numbers.json' with {type: 'json'}
import AVAILABLE_LOCALES from 'cldr-core/availableLocales.json' with {type: 'json'}
import rawTimeData from 'cldr-core/supplemental/timeData.json' with {type: 'json'}
import rawCalendarPreferenceData from 'cldr-core/supplemental/calendarPreferenceData.json'
import TimeZoneNames from 'cldr-dates-full/main/en/timeZoneNames.json' with {type: 'json'}
import metaZones from 'cldr-core/supplemental/metaZones.json' with {type: 'json'}
```

formatjs 的 `package.json` 锁定如下 npm 包：`cldr-bcp47`、`cldr-core`、`cldr-dates-full`、`cldr-localenames-full`、`cldr-misc-full`、`cldr-numbers-full`、`cldr-segments-full`、`cldr-units-full`，均锁在 `48.1.0`<!-- ref: formatjs/package.json -->。

**路径 B：复用 `golang.org/x/text/cldr`**

`golang.org/x/text/cldr` 提供 LDML XML 解析，但：
- 停留在 CLDR 32 左右（最新提交 2018 年），与 formatjs 的 48 相差约 7 年。
- 不直接暴露给生产代码——它的设计目标是"x/text 内部工具消费"。
- 数据形态是 Go struct，不是 JSON；改成 JSON 路径需要二次序列化。

**路径 C：消费 formatjs 生成的中间产物**

formatjs 生成 `@formatjs_generated/cldr.locale/`、`cldr.number/`、`tz/` 等 npm 包<!-- ref: formatjs/knowledge-base/001-repo-layout.md "Generated Data Packages" -->。我们可以直接拉取这些 generated 包做转换。

但这些包是 TypeScript 源 + Bazel 编译产物，并不发到 npm（"Generated files are compiled and packaged ... not checked into git"）。要用必须先在我们 CI 里跑一遍 formatjs 的 Bazel 流水线，运维成本极高。

### 1.2 决策

**推荐路径 A（直接消费 cldr-json）。** 依据：

1. **同源** — 与 formatjs 锁同一 CLDR 版本，conformance 失败必然源于代码差异而非数据差异，调试路径清晰。
2. **稳定** — `cldr-json` 的 npm 包发布节奏与 CLDR 版本一致，每年两次（春/秋）。
3. **可控** — 我们的 generator 自己控制抽取规则；与 formatjs 在数据形态上不耦合（formatjs 改字段名不影响我们）。

**不推荐路径 B（`x/text/cldr`）。** 数据滞后 + 不暴露 + 形态不匹配。但其 LDML 解析逻辑可作为 generator 内部参考（避免重写 LDML XML/JSON parser）。

**不推荐路径 C（formatjs 中间产物）。** 运维不可行。

### 1.3 抽取范围（Phase 1）

参照 SPEC 00 §5.1 列出的 `internal/cldr/` 文件，每个文件对应抽取入口：

| `internal/cldr/` 文件 | CLDR 来源 | 字段 |
|----------------------|-----------|------|
| `numbers.go` | `cldr-numbers-full/main/<locale>/numbers.json` | symbols（小数点/千分位/分组）、currency / unit / compact pattern、percent pattern |
| `dates.go` | `cldr-dates-full/main/<locale>/ca-gregorian.json` | era/month/weekday/dayPeriod 名、formats、intervalFormats、availableFormats |
| `metazones.go` | `cldr-core/supplemental/metaZones.json` + `cldr-dates-full/main/<locale>/timeZoneNames.json` | zone → metazone 映射、metazone 显示名 |
| `plurals.go` | `cldr-core/supplemental/plurals.xml` + `pluralRanges.xml` | cardinal / ordinal / pluralRanges |
| `preference.go` | `cldr-core/supplemental/timeData.json` + `weekData.json` + `calendarPreferenceData.json` + `likelySubtags.json` | hourCycle 偏好、firstDay / weekend / minDays、calendar 偏好、likely subtags |
| `currencies.go` | `cldr-numbers-full/main/<locale>/currencies.json` + `cldr-core/supplemental/currencyData.json` | 货币显示名、复数形式、货币精度 |
| `units.go` | `cldr-units-full/main/<locale>/units.json` | unit 复数形式、长短窄各形态 |

## 2. 嵌入策略

### 2.1 三种形态

**形态 A：`//go:embed` JSON，运行时反序列化**

代价：
- 冷启动：100 个 locale × 几个数据文件 × JSON parse ≈ 100–500 ms。
- 内存：JSON parse 后保留 map[string]any 或 strongly-typed struct，约 5–10 MB。
- Binary：JSON 文件直接打包，约 5–20 MB。

收益：
- 数据更新只换 JSON 文件，无需重新生成 Go 代码（理论上）。
- 实际上仍需重新构建（embed 是编译时绑定）。

**形态 B：Go 字面量生成**

translate-agent 走的就是这条路<!-- ref: intl/internal/cldr/data.go -->：

```go
// 提炼自 .references/intl/internal/cldr/data.go
package cldr

// Code generated by "earthly +generate". DO NOT EDIT.
const data = "..."  // ~400 KB Unicode 文本，所有 era/month/weekday 名拼接
```

文件大小：400 KB（3203 行）<!-- ref: intl/internal/cldr/data.go -->。其结构是把所有去重后的字符串拼成一个 `const data` 巨型字符串，用 `[start:end]` 切片访问。

translate-agent 的 generator (`internal/gen/cldr_data.go.tmpl`<!-- ref: intl/internal/gen/cldr_data.go.tmpl -->) 用 `text/template` 做 codegen，模板中：

```go
// 模板片段
var EraLookup = map[string]Era{
{{-  range $locale, $era := .Eras }}
  "{{ $locale }}": { ... },
{{- end }}
}
```

代价：
- Generator 复杂度（一次性投入）。
- Binary：与 JSON 接近，但去掉了 JSON 的引号 / 大括号开销，估计 60–80%。
- 编译时间：3203 行 Go 编译比 JSON parse 快得多，但比短代码慢——translate-agent 的 `data.go` 单文件编译约 1 s。

收益：
- 运行时零解析，零 JSON 依赖。
- 字符串去重后 binary 体积比 JSON 小（实测 translate-agent 的 400 KB 涵盖了几百个 locale 的几千条文本）。
- IDE 跳转能定位到具体常量。

**形态 C：嵌入二进制 + 自定义解码器**

formatjs 走 base36 + `|` 分隔符的紧凑编码（在 tz_data 部分用过<!-- ref: formatjs/packages/intl-datetimeformat/scripts/packer.ts:3-19 -->），但其 cldr.locale 数据本身是 TS 源，不是二进制。Node 的 ICU 走二进制（`icudt*.dat`），但通过 native 加载。

### 2.2 决策

**推荐形态 B（Go 字面量生成）。** 依据：

1. **运行时零成本** — 没有 JSON parse 开销；所有数据在 `.rodata` 段，O(1) 访问。
2. **Go 生态先例** — `golang.org/x/text/internal/data`、`golang.org/x/text/currency`、translate-agent 全部走这条路。
3. **体积可控** — translate-agent 验证了 400 KB 涵盖几百个 locale 的 era/month/weekday 数据；NumberFormat / DateTimeFormat 的额外数据（formats / intervalFormats / dayPeriods / metazones）估计 +500 KB–1 MB；总量在 1–2 MB 量级。
4. **冷启动友好** — 服务端短连接、CLI、AWS Lambda 场景下，无 JSON parse 关键路径。

**不推荐形态 A（go:embed JSON）。** 依据：

1. 冷启动开销不可接受（CLI 工具尤其敏感）。
2. JSON parse 需要 `encoding/json` 依赖与运行时反射。
3. 体积没有显著优势（实测 JSON 比 Go 字面量大 30–50%，因为多了引号 / 字段名 / 缩进）。

**形态 C（自定义二进制）暂不必要。** Phase 4 才考虑——只有当 binary 体积压倒一切时（如 WASM 部署）才值得增加解码器复杂度。

### 2.3 体积估算（基于 translate-agent 实测）

`du -sh /Users/lincheng/work/golang/go-intl/.references/intl/internal/cldr/`：**432 KB**。

明细：
- `data.go`：400 KB / 3203 行（去重的 Unicode 字符串 + lookup map）<!-- ref: intl/internal/cldr/data.go -->
- `locale.go`：20 KB / 504 行（language.MustParseBase 列表）<!-- ref: intl/internal/cldr/locale.go -->
- `numbering.go`：4 KB / 180 行（numbering system 选择逻辑）<!-- ref: intl/internal/cldr/numbering.go -->
- `cldr.go`：1 KB / 52 行（lookup 函数）<!-- ref: intl/internal/cldr/cldr.go -->
- `fmt.go`：1.4 KB / 76 行
- `internal/symbols/symbols.go`：5.3 KB

translate-agent 的覆盖范围：era / month / weekday 名 × 数百 locale × Gregorian 日历 + Persian + Buddhist 子集。

**对 go-intl Phase 1 的外推估算：**

| 数据类型 | 范围 | 预估体积 |
|---------|------|---------|
| era/month/weekday 名 | translate-agent 覆盖范围 | 400 KB |
| dayPeriod / availableFormats / intervalFormats | + dateStyle/timeStyle 与 skeleton 矩阵 | +300 KB |
| metaZones（zone → metazone → 显示名） | ~430 zones × ~50 显示名 locale | +200 KB |
| number symbols / currency / unit pattern | 每 locale 30–50 行模式串 | +250 KB |
| currency display names + plurals | 200 货币 × ~40 locale | +200 KB |
| compact / scientific 数字 pattern | small set per locale | +50 KB |
| plurals 编译产物 | 200 locale × ~10 行 Go 代码 | +50 KB |
| 合计（Phase 1，约 100 locale） | | **~1.5 MB** |

如果扩展到全部 ~500 locale，估算 **~6 MB**。Go 1.26 binary 大小通常 5–20 MB，1.5 MB CLDR 增量是可接受的；6 MB 增量就需要考虑 build tag 拆分。

## 3. per-locale tree-shaking

formatjs 通过 `__addLocaleData` 调用 + Bazel 拆包做"按需加载"<!-- ref: formatjs/knowledge-base/001-repo-layout.md "Tree-shakeable" -->，每个 locale 在 npm 上是独立子路径（`@formatjs/intl-datetimeformat/locale-data/zh.js`）。

translate-agent 全量嵌入，不拆分<!-- ref: intl/internal/cldr/data.go 的 const data 是单文件 -->。

V8 / ICU 通过 `small-icu` / `full-icu` 构建模式拆分<!-- ref: node/doc/api/intl.md:33-110 -->。

### 3.1 决策

**Phase 1：全量嵌入（约 100 个常用 locale）。** 依据：
- 1.5 MB 增量在 v1 可接受。
- 实现简单（generator 不需要 build tag 分支）。
- `messageformat-go` 的当前用户没有"我只用 en"的硬约束。

**Phase 4：build tag 分级。** 设计：
- 默认（无 tag）：~100 个常用 locale，~1.5 MB。
- `intl_full`：全部 ~500 locale，~6 MB。
- `intl_minimal`：仅 `root` + `en`，~50 KB。

实现路径：generator 输出 `data_default.go` / `data_full.go` / `data_minimal.go` 三个文件，文件头部加 `//go:build`。Phase 4 引入即可，Phase 1 不需要预设。

不推荐 formatjs 风格的"per-locale 子包"（如 `numberformat/locale-data/zh`）：

- Go 没有 `__addLocaleData` 这种动态注册机制；动态注册需要 `init()` 函数 + 全局 map 写入，不利于纯函数化。
- 子包之间的 import 关系会污染依赖图。
- Go 的 dead-code elimination 已经能消除未使用的 const string，build tag 即可。

## 4. CLDR / ICU 版本钉

### 4.1 候选与对应关系

| ICU 版本 | CLDR 版本 | 发布时间 |
|---------|-----------|----------|
| 74.2 | CLDR 44 | 2023-11 |
| 75.1 | CLDR 45 | 2024-04 |
| 76.1 | CLDR 46 | 2024-10 |
| 78.2 | CLDR 48 | 2025-10 |

formatjs 当前主分支锁定 `cldr-*: 48.1.0`<!-- ref: formatjs/package.json -->，对应 ICU 78。

### 4.2 决策

**钉 CLDR 48.1.0 / ICU 78。** 写入 `internal/cldr/VERSION`。

依据：
1. **与 formatjs 同步** — 我们的输出"与 formatjs 字节级一致"是首要目标；钉 v44 / v45 / v46 等于在 conformance 上把"我们对、formatjs 错"的反向 divergence 制度化，违反 SPEC 00 §2 "formatjs 是主参考"。
2. **更新窗口** — CLDR 48 / ICU 78 发布于 2025-10，到 v1.0 发布前预计稳定 6+ 个月。

`internal/cldr/VERSION`：
```
cldr=48.1.0
icu=78
tzdata=2025b
```

任何修改这三个版本都是 SPEC 影响行为，需经评审。

## 5. 标识符来源

### 5.1 货币（ISO 4217）

formatjs 通过 `cldr-numbers-full` 的 `currencies.json` 取显示名与复数形式，通过 `cldr-core/supplemental/currencyData.json` 取精度（最小/最大小数位）<!-- ref: formatjs/packages/intl-numberformat/scripts/extract-currencies.ts、currency-digits.ts -->。

CLDR `currencyData.json` 内嵌 ISO 4217 编码与精度，二者同步。

PHP ext/intl 直接从 ICU 取。

> 决策：从 CLDR 取。不引入独立 ISO 4217 fetch（避免双源同步问题）。

### 5.2 时区（IANA）

formatjs 走 `cldr-core/supplemental/metaZones.json` + 自带 zdump<!-- ref: formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts:137-163 -->。

> 决策：transition 表用 `time/tzdata`（IANA 直接，与 Go runtime 同步），metazone 映射 / 显示名用 CLDR。详见 R04 §2。

### 5.3 单位（UN/CEFACT 子集）

formatjs 在 `units-constants.ts` 维护 sanctioned 列表（meter / kilometer / second / hour / liter 等约 50 项），与 ECMA-402 spec 列表一致<!-- ref: formatjs/packages/intl-numberformat/scripts/units-constants.ts -->。`cldr-units-full/main/<locale>/units.json` 提供模式与复数形式。

> 决策：sanctioned 列表 hardcode 进 `internal/ecma402/numberformat/`（spec 列表，不与 CLDR 走 in-out-of-sync），模式与显示名走 CLDR。

## 6. 生成器架构

### 6.1 三种参考的形态

**translate-agent**：`internal/gen/`（Go 子包，`go run` 触发）+ `internal/gen/cldr_data.go.tmpl`（text/template）+ Earthly 容器编排<!-- ref: intl/internal/gen/main.go、gen.go、cldr_data.go.tmpl -->。优点：单语言（Go），CI 简单。缺点：CLDR 解析逻辑用 Go 重写，与 formatjs 不能共享。

**formatjs**：`packages/intl-*/scripts/extract-*.ts`（TypeScript）+ Bazel 编排 + 输出 `@formatjs_generated/*` 包<!-- ref: formatjs/knowledge-base/001-repo-layout.md "Generated Data Packages" -->。优点：与 polyfill 用同一语言；缺点：Bazel 构建链对 Go 项目过重。

**V8 / ICU**：完全离线编译产物，与项目无关。

### 6.2 决策

**生成器位置：仓库内子包 `tools/gen-cldr/`，独立 `go.mod`。**

```
go-intl/
├── go.mod              # 主 module，不含 CLDR 抽取依赖
└── tools/
    └── gen-cldr/
        ├── go.mod      # 独立 module，可依赖 CLDR JSON 解析、第三方工具
        ├── main.go
        ├── extract/
        │   ├── dates.go
        │   ├── numbers.go
        │   ├── plurals.go
        │   ├── timezones.go
        │   └── identifiers.go
        ├── codegen/
        │   ├── stringtable.go     # 去重字符串 + 索引表生成
        │   ├── locale_lookup.go   # map[locale]Index 生成
        │   └── format.go          # gofmt 输出
        └── cldr/
            ├── fetch.go           # 拉 cldr-json npm 包或 GitHub release
            └── parse.go           # JSON → Go struct
```

**关键设计：**

1. **独立 `go.mod`** — 主 module 不引入 CLDR 解析依赖（如 `encoding/json` 重度使用、`gjson`、`fast-glob` 等），保持依赖图干净。
2. **输入：本地 `cldr-json` 目录** — 由 CI 拉取（`npm install cldr-dates-full@48.1.0` 或 GitHub release），避免每次跑都打 npm。
3. **输出：`internal/cldr/*.go`** — generator 只写文件，不直接修改 `internal/cldr/` 之外的代码。
4. **`task gen-cldr`** — 在 `Taskfile.yml` 中暴露入口，CI 失败时人工运行；不在 `task verify` 中默认触发（避免 CI 网络依赖）。
5. **`internal/cldr/VERSION`** — 单一事实源；generator 启动时读它定位 npm 包版本。

> 落地接口建议：
>
> ```go
> // 提炼自落地建议（tools/gen-cldr/）
> type Config struct {
>     CLDRDir string  // 本地 cldr-json 解压根目录
>     OutDir  string  // internal/cldr 路径
>     Version string  // CLDR 版本号校验
> }
>
> type Extractor interface {
>     Name() string
>     Extract(ctx context.Context, src *cldr.Source) (codegen.Module, error)
> }
>
> func Run(ctx context.Context, cfg Config, extractors []Extractor) error
> ```

## 7. 对本项目的落地建议

### 7.1 数据源链路

```
cldr-json npm 包 (48.1.0)
        │
        ▼
tools/gen-cldr/  (独立 go.mod, 只在本地/CI 跑)
        │
        ▼
internal/cldr/*.go   (生成的 Go 字面量, 提交进仓库)
        │
        ▼
内部消费者: internal/ecma402/, locale/, numberformat/, datetimeformat/, pluralrules/
```

### 7.2 文件布局（最终态）

```
internal/cldr/
├── VERSION                    # cldr=48.1.0 / icu=78 / tzdata=2025b
├── README.md                  # 数据来源、生成方法
├── strings.go                 # 共享去重字符串表（const _data string）
├── locales.go                 # 支持的 locale 列表 + AvailableLocales()
├── numbers.go                 # number symbols / patterns
├── dates.go                   # era/month/weekday/dayPeriod/formats
├── intervals.go               # intervalFormats
├── metazones.go               # zone → metazone → display name
├── currencies.go              # currency display names + digits
├── units.go                   # unit pattern + plurals
├── plurals.go                 # 编译后的 cardinal / ordinal 选择函数
├── preference.go              # weekData / timeData / likelySubtags
└── doc.go                     # 包文档
```

### 7.3 数据访问 API（简化提炼）

```go
// 提炼自落地建议
package cldr

// Locale 是一个不透明的 locale 句柄；内部是 dataLocale 索引。
type Locale uint16

func ResolveLocale(tag language.Tag) (Locale, bool)

// 数字符号与模式
type NumberSymbols struct { Decimal, Group, Percent, Plus, Minus, NaN, Infinity, ApproxSign string }
func (l Locale) NumberSymbols(ns string) NumberSymbols
func (l Locale) DecimalPattern(ns string) string
func (l Locale) PercentPattern(ns string) string
func (l Locale) CurrencyPattern(ns string) string
func (l Locale) CompactPattern(ns string, exp int, plural string) string

// 日期/时间
type CalendarNames struct { Eras, Months, Weekdays, DayPeriods []string }
func (l Locale) CalendarNames(calendar, width, context string) CalendarNames
func (l Locale) DateFormat(style string) string
func (l Locale) TimeFormat(style string) string
func (l Locale) AvailableFormats() map[string]string  // skeleton -> pattern
func (l Locale) IntervalFormats() IntervalFormats

// 时区
func ZoneToMetazone(zone string) string
func (l Locale) MetazoneName(metazone, kind string) string  // kind: long/short, generic/standard/daylight

// 货币
func CurrencyDigits(code string) (min, max int)
func (l Locale) CurrencyDisplayName(code, plural string) string

// 复数
func (l Locale) Cardinal(operand Operand) Form
func (l Locale) Ordinal(operand Operand) Form
```

所有访问器都是 `O(1)` 查表，不分配。

### 7.4 测试策略

- **生成器自测**：`tools/gen-cldr/` 自带 `gen_test.go`，断言生成结果中"en/zh/ar 的 era/month 字段不为空"等基本契约。
- **快照测试**：`internal/cldr/snapshot_test.go` 比对当前 git 中的 `internal/cldr/*.go` 与重新生成的输出，CI 拒绝差异（避免手改）。
- **conformance fixture**：把 formatjs `tests/locale-data/` 吃进 `testdata/`，pluralrules / numberformat / datetimeformat 各自的 conformance test 直接消费。

## 8. 决策矩阵

| 主题 | 推荐 | 备选 | 否决 | 依据 |
|------|------|------|------|------|
| 数据源 | Unicode `cldr-json` npm 包 | formatjs 中间产物 | `golang.org/x/text/cldr`（数据滞后） | §1.2 |
| 嵌入策略 | Go 字面量（generator codegen） | 自定义二进制 + 解码器（Phase 4 候选） | `//go:embed` JSON | §2.2 |
| 体积估算 | Phase 1 ~1.5 MB（~100 locale） | 全量 ~6 MB（intl_full tag） | 必须按 locale 拆 npm 风格子包 | §2.3, §3.1 |
| Tree-shaking | 默认全量；预留 `intl_full` / `intl_minimal` build tag（Phase 4） | per-locale 子包动态注册 | Phase 1 就拆包 | §3.1 |
| 版本钉 | **CLDR 48.1.0 / ICU 78**（与 formatjs 同步） | CLDR 47 / ICU 76 | CLDR 44 / ICU 74（与 formatjs 反向 diverge） | §4.2 |
| 货币标识符 | CLDR `currencyData` | 独立 ISO 4217 表 | 第三方包 | §5.1 |
| 时区数据 | `time/tzdata`（transition）+ CLDR `metaZones`（显示名） | 自带 zdump 流水线（Docker 编排） | 系统 zoneinfo | R04 §2.2, §5.2 |
| 单位标识符 | spec sanctioned 列表 hardcode + CLDR 模式 | 全部从 CLDR 抽（含 spec 之外的单位） | UN/CEFACT 直接取 | §5.3 |
| Generator 位置 | `tools/gen-cldr/`（独立 `go.mod`） | `internal/gen/`（同 module） | 独立仓库 | §6.2 |
| Generator 触发 | `task gen-cldr` 手动 / CI（不在 `task verify`） | 每次 `task verify` 都跑 | git pre-commit | §6.2 |

## 9. 代码块索引

| 章节 | 代码块 | 来源 |
|------|--------|------|
| §1.1 | formatjs `cldr-*` import 清单 | formatjs/scripts/extract-dates.ts |
| §2.1 | translate-agent `const data = "..."` 模式 | intl/internal/cldr/data.go + internal/gen/cldr_data.go.tmpl |
| §6.2 | `tools/gen-cldr/` 目录骨架 + `Config / Extractor / Run` 接口签名 | 落地建议 |
| §7.3 | `cldr.Locale` + 各类访问器签名（`NumberSymbols / CalendarNames / IntervalFormats / Cardinal / Ordinal`） | 落地建议 |

## 10. 引用清单

### formatjs（主参考）
- `.references/formatjs/package.json` — `cldr-bcp47/core/dates-full/localenames-full/misc-full/numbers-full/segments-full/units-full: 48.1.0`
- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts:7-30,137-200` — CLDR JSON import 清单与抽取逻辑
- `.references/formatjs/packages/intl-numberformat/scripts/extract-numbers.ts` — number 数据抽取
- `.references/formatjs/packages/intl-numberformat/scripts/extract-currencies.ts` — currency 抽取
- `.references/formatjs/packages/intl-numberformat/scripts/currency-digits.ts` — currencyData → 精度表
- `.references/formatjs/packages/intl-numberformat/scripts/extract-units.ts` — units 抽取
- `.references/formatjs/packages/intl-numberformat/scripts/units-constants.ts` — sanctioned 单位列表
- `.references/formatjs/packages/intl-pluralrules/scripts/plural-rules-compiler.ts` — pluralRules 编译
- `.references/formatjs/packages/intl-datetimeformat/scripts/packer.ts` — base36 + `|` 紧凑编码
- `.references/formatjs/knowledge-base/001-repo-layout.md` — `@formatjs_generated/*` 包架构、数据源组织
- `.references/formatjs/knowledge-base/001a-bazel-toolchain.md` — `formatjs_generated_package()` 宏

### translate-agent/intl（Go 先例）
- `.references/intl/internal/cldr/data.go` — 400 KB / 3203 行单文件 Go 字面量
- `.references/intl/internal/cldr/locale.go:1-100` — language.MustParseBase 列表（504 行）
- `.references/intl/internal/cldr/numbering.go:1-181` — numbering system 选择
- `.references/intl/internal/cldr/cldr.go:1-52` — MonthNames / EraName lookup 函数
- `.references/intl/internal/symbols/symbols.go` — symbol 字典 (5.3 KB)
- `.references/intl/internal/gen/main.go:1-40` — generator 入口
- `.references/intl/internal/gen/gen.go:1-60` — Generator / Conf / Gen 函数
- `.references/intl/internal/gen/cldr_data.go.tmpl:1-50` — text/template 模板
- 体积实测：`du -sh internal/cldr/` = 432 KB

### 项目内部
- `SPECS/00-vision-and-scope.md:189-198` — §5.3 数据策略（无 runtime JSON parse、universal locales、版本 pin）
- `SPECS/00-vision-and-scope.md:249-258` — §8 开放问题第 4 项（CLDR 版本钉）
- `ANALYSIS.md §6` — 数据源选择、嵌入策略、tree-shaking、版本钉、标识符来源、generator 架构
