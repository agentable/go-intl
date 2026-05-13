# SPEC 50 — CLDR Data & Codegen

> **Status:** Draft (2026-05-08)
> **Priority:** High(全部 formatter 数据底层;阻塞 SPEC 10 / 20 / 30 / 31 / 32 / 40)
> **Authority:** 本 SPEC 是 `internal/cldr/` 包结构、CLDR / ICU / tzdata 版本钉、active scope locale 范围、`tools/gen-cldr/` 代码生成器架构的 SSOT。**关闭 SPEC 00 §8 Q4(CLDR 版本)**。

---

## Overview

`internal/cldr/` 是 go-intl 全部 formatter 的数据底层。它把 [Unicode CLDR](https://cldr.unicode.org/) 的 locale-aware 表(数字符号、货币精度、日期模式、复数规则、时区显示名、likely subtags、locale matching 数据)在生成期编译为 Go 字面量,运行时通过 O(1) 访问器供 formatter 包消费。`tools/gen-cldr/` 是该包的代码生成器,作为独立 Go module 维护。

本 SPEC 定义:数据源选型(直接消费 `unicode-org/cldr-json`)、嵌入策略(Go 字面量,**不**走 `//go:embed *.json`)、active scope locale 范围(~100 modern tier 1)、版本钉(`cldr=48.1.0` / `icu=78` / `tzdata=2025b`)、生成器位置(`tools/gen-cldr/` 独立 `go.mod`)、文件布局与访问器接口、升级流程。

---

## 1. Data Source

### 1.1 选型

CLDR 数据 **必须**直接消费 [`unicode-org/cldr-json`](https://github.com/unicode-org/cldr-json) 的 npm 镜像包(`cldr-bcp47` / `cldr-core` / `cldr-dates-full` / `cldr-localenames-full` / `cldr-misc-full` / `cldr-numbers-full` / `cldr-segments-full` / `cldr-units-full`),与 formatjs 同源。

> **Why**:
> 1. **同源** —— 与 formatjs 锁同一 CLDR 版本,conformance 失败必然源于代码差异而非数据差异,调试路径清晰。
> 2. **稳定** —— `cldr-json` npm 包发布节奏与 CLDR 版本一致,每年两次(春/秋)。
> 3. **形态匹配** —— JSON 直接对接 Go `encoding/json`(仅在生成期使用),无需 LDML XML 解析器。
>
> **Rejected**:
> - **`golang.org/x/text/cldr`**:数据版本与 go-intl / formatjs 的 CLDR 钉版不构成同一 conformance 基线;它的设计目标是 `x/text` 内部工具消费,不暴露给生产代码;数据形态是 Go struct,不是 JSON,改造成本高。
> - **formatjs 中间产物**(`@formatjs_generated/cldr.locale/` 等):是 TS 源 + Bazel 编译产物,**不发到 npm**(formatjs `knowledge-base/001-repo-layout.md` 明示 "Generated files are compiled and packaged, not checked into git");要用必须先在 CI 跑 formatjs Bazel 流水线,运维成本不可接受。
> - **ICU CGO 绑定**(`goccy/go-icu` / `goodsign/icu4go`):前者 GitHub 404 不存在,后者无活跃维护;CGO 与 SPEC 00 §1.1 "不依赖 ICU C/C++" 冲突。

### 1.2 抽取范围(active scope) <a id="schema"></a>

每个 `internal/cldr/<file>.go` 对应一个抽取入口:

| `internal/cldr/<file>.go` | CLDR 来源 | 提取字段 |
|---------------------------|-----------|---------|
| `numbers.go` | `cldr-numbers-full/main/<locale>/numbers.json` | symbols(decimal/group/percent/plus/minus/NaN/Infinity)、decimalFormats、percentFormats、currencyFormats、scientificFormats、numberingSystems |
| `dates.go` | `cldr-dates-full/main/<locale>/ca-gregorian.json` | era / month / weekday / quarter / dayPeriod 名(stand-alone × format,wide / abbreviated / narrow),dateFormats、timeFormats、dateTimeFormats、availableFormats、intervalFormats |
| `metazones.go` | `cldr-core/supplemental/metaZones.json` + `cldr-dates-full/main/<locale>/timeZoneNames.json` | zone → metazone 映射、metazone 显示名(long / short × generic / standard / daylight)、exemplarCity |
| `timezones.go` | `cldr-core/supplemental/metaZones.json` + IANA `backward` | IANA link → canonical zone 映射 |
| `collations.go` | `cldr-bcp47/bcp47/collation.json` | canonical collation identifiers, excluding deprecated and non-sort internal values |
| `plural/*.go` | `cldr-core/supplemental/plurals.json` + `ordinals.json` + `pluralRanges.json` | cardinal rules、ordinal rules、pluralRanges(本文件由 SPEC 40 codegen 输出;本 SPEC 仅约定文件位置) |
| `preference.go` | `cldr-core/supplemental/timeData.json` + `weekData.json` + `calendarPreferenceData.json` | 每 region 的 hourCycle 偏好、firstDay / weekend / minDays、calendar 偏好 |
| `likely_subtags.go` | `cldr-core/supplemental/likelySubtags.json` | language → maximize 后的 (script, region) 映射 |
| `locale_matching.go` | `cldr-core/supplemental/languageMatching.json` | paradigmLocales、matchVariables、distance 表(SPEC 11 BestFitMatcher 数据) |
| `regions.go` | `cldr-core/supplemental/territoryContainment.json` | matchVariables 区域展开($enUS / $cnsar / $americas / $maghreb 等) |
| `currencies.go` | `cldr-numbers-full/main/<locale>/currencies.json` + `cldr-core/supplemental/currencyData.json` | 货币显示名(long / short / narrow)、复数形式、defaultFractionDigits、cashDigits、rounding |
| `units.go` | `cldr-units-full/main/<locale>/units.json` | unit 复数模式(long / short / narrow)、compoundUnitPatterns |
| `locales.go` | `cldr-core/availableLocales.json` | 支持的 locale 列表 + `AvailableLocales()` 访问器 |
| `supported.go` | 派生自 generated runtime maps and generator constants | formatter-specific supported locale 列表 + `Intl.supportedValuesOf` value accessors |
| `strings.go` | (派生) | 共享去重字符串表(`const _data string`) |

### 1.3 标识符来源决策

| 标识符 | 来源 | 备注 |
|--------|------|------|
| 货币(ISO 4217 编码 + 精度) | CLDR `currencyData.json` | **不**引入独立 ISO 4217 表(避免双源同步);**不**引入 `bojanz/currency`(自维护 CLDR 派生表,版本独立,与 `internal/cldr/VERSION` 漂移) |
| 时区(IANA zone) | `time/tzdata`(transition 表)+ CLDR `metaZones.json`(显示名) | tzdata 由 SPEC 32 注入;本 SPEC 仅生成 `metazones.go` 显示名表 |
| 排序 collation 标识符 | CLDR BCP47 `collation.json` | 生成期过滤 deprecated、`ducet`、`search`、`standard`;运行时不得手写列表 |
| 单位标识符(sanctioned list) | ECMA-402 spec hardcode 进 `internal/ecma402/numberformat/constants.go` | spec 列表权威;CLDR 提供模式但不提供 sanctioned list |

> **Why**: 货币精度走 CLDR 是 R03 决策;`bojanz/currency` 自维护 CLDR 派生表会引入"两份 CLDR 数据"——任何升级都要双向校对。spec sanctioned units 是 normative,任何 CLDR-driven 探测都可能与 spec 偏离。
>
> **Rejected**:
> - 引入 ISO 4217 静态表(与 CLDR / ICU / formatjs 不一致,出现 conformance 反向 divergence)。
> - `bojanz/currency` 作为 ECMA-402 数据源(版本路径分裂,见 R09 §4.5)。
> - CLDR 探测 sanctioned units 列表(spec 是权威,CLDR 表更广)。

---

## 2. Embedding Strategy

### 2.1 Go 字面量,不走 `//go:embed *.json`

CLDR 数据 **必须**编译为 Go 字面量(常量字符串、map literal、slice literal),由 `tools/gen-cldr/` 生成。**禁止**:

- `//go:embed *.json` + 运行时 `encoding/json.Unmarshal`。
- `//go:embed *.bin` + 自定义二进制解码器(consumer-driven expansion 才考虑)。
- 启动时网络拉取或文件 I/O。

> **Why**:
> 1. **运行时零成本** —— Go 字面量在 `.rodata` 段,O(1) 访问;JSON parse 100 个 locale × 几个数据文件 ≈ 100–500 ms 冷启动开销不可接受(CLI / Lambda / 短连接服务尤其敏感)。
> 2. **去除 JSON 依赖** —— 运行时不引入 `encoding/json`(仅生成期使用);减小 binary 与依赖图。
> 3. **Go 生态先例** —— `golang.org/x/text/internal/data`、`x/text/currency`、`translate-agent/intl` 全部走这条路。
> 4. **体积可控** —— 实测 `translate-agent/intl/internal/cldr/data.go` 单文件 400 KB / 3203 行,涵盖几百个 locale 的 era/month/weekday;字符串去重后 binary 比 JSON 小 30–50%。
> 5. **运行时契约清晰** —— CLDR 数据是生成的 Go source;hot path 不做文件 I/O,不解析 JSON。
>
> **Rejected**:
> - **`//go:embed *.json` + 运行时反序列化**:冷启动开销 + JSON / 反射依赖 + 体积更大。
> - **自定义二进制 + 解码器**(formatjs 在 tz_data 用过 base36 + `|` 分隔):active scope 不必要;只在 binary 体积压倒一切(如 WASM)时考虑(consumer-driven expansion)。
> - **运行时网络拉取**(cdn / s3 上托管):Go 二进制单体哲学 + offline 部署需求矛盾。

### 2.2 共享字符串表

为减小 binary 体积,**应当**使用单一 `const _data string` 巨型字符串 + `[start:end]` 索引访问的去重模式:

```go
// internal/cldr/strings.go(签名;由 codegen 生成)
package cldr

// 由 tools/gen-cldr/codegen/stringtable.go 输出。
const _data = "" +
    "JanuaryFebruaryMarchAprilMayJuneJulyAugustSeptemberOctober..." +
    // (拼接所有 locale 的 era/month/weekday/dayPeriod/...)

// 通过 [start:end] 取出对应字符串(零分配)。
type sliceRef struct {
    start, length uint32
}
```

> **Why**: `translate-agent/intl/internal/cldr/data.go` 验证可行;Go 编译期把 `const _data` 直接放进 `.rodata`;`_data[start:end]` 是 string slice 操作,零分配。
>
> **Rejected**:
> - 每个字段一个独立 `var s = "..."`:重复字符串无法去重(`"January"` 在 100 个 locale 重复约 10 次);binary 增大约 30%。
> - `var stringTable = []string{...}`:slice header 开销;`s[i]` 是 indirect 访问,失去常量折叠。

### 2.3 体积估算与目标

| 数据类型 | 范围 | 预估体积 |
|---------|------|---------|
| era / month / weekday 名 | translate-agent/intl 覆盖范围(~100 locale × Gregorian) | 400 KB |
| dayPeriod / availableFormats / intervalFormats | + dateStyle/timeStyle 与 skeleton 矩阵 | +300 KB |
| metaZones(zone → metazone → 显示名) | ~430 zones × ~50 显示名 locale | +200 KB |
| number symbols / currency / unit pattern | 每 locale 30–50 行模式串 | +250 KB |
| currency display names + plurals | 200 货币 × ~40 locale | +200 KB |
| compact / scientific 数字 pattern | small set per locale | +50 KB |
| plurals 编译产物(SPEC 40 codegen) | 200 locale × ~10 行 Go 代码 | +50 KB |
| **active scope 合计(~100 locale)** | | **≤ 1.5 MB** |

体积 **必须**通过 `task build:size` 监控;active scope 上限 1.5 MB。超出需 SPEC 修订。

### 2.4 单文件 vs 分文件

CLDR 数据 **必须**按 §1.2 表分文件生成(每类一个 `.go` 文件)。每个 locale **可以**进一步分到独立 `data_<locale>.go`(如 `data_en_US.go`),以利 Go 并行编译。

> **Why**:
> 1. translate-agent/intl 的 400 KB 单文件在 `go build` 中约 1 秒编译;100+ locale 全部塞进单文件可能导致 5+ 秒。
> 2. 分文件后 `go build` 并行编译每个文件,减少冷启动 build 时间。
> 3. IDE 跳转可定位到具体 locale 的具体常量。
>
> **Rejected**:
> - **单一 `data.go` 容纳全部 locale**:编译耗时 + IDE 加载缓慢。

---

## 3. active scope Locale Scope

### 3.1 全量嵌入 ~100 modern tier 1 locale

active scope **必须**全量嵌入约 100 个 CLDR modern tier 1 / 2 locale,**禁止**:

- per-locale tree-shaking(formatjs `__addLocaleData` 风格)。
- sub-locale 懒加载(运行时 file I/O 拉取 `zh-Hans` 数据)。
- build tag 拆分(intl_full / intl_minimal,**consumer-driven expansion** 才考虑)。

候选清单(实施期由 `tools/gen-cldr/locale_list.go` 维护;**SPEC 不固定具体清单**,仅给出方向):

```text
# 必须包含(active scope 强约束)
en, en-US, en-GB, en-CA, en-AU
zh, zh-Hans, zh-Hans-CN, zh-Hant, zh-Hant-TW, zh-Hant-HK
ja, ja-JP
ko, ko-KR
fr, fr-FR, fr-CA
de, de-DE, de-AT
es, es-ES, es-419, es-MX
pt, pt-BR, pt-PT
ru, ru-RU
ar, ar-SA, ar-EG
hi, hi-IN
it, it-IT
nl, nl-NL
pl, pl-PL
tr, tr-TR
th, th-TH
vi, vi-VN
id, id-ID

# 应当包含(覆盖 Tier 1 区域)
sv, da, nb, fi, cs, hu, ro, sk, uk, el, he, fa,
ms, bn, ta, te, ml, mr, gu, kn,
... (合计 ~100 个)
```

### 3.2 体积约束

总生成代码 **必须** ≤ 1.5 MB。CI 通过 `task build:size` 校验。

> **Why**:
> 1. **YAGNI** —— active scope 主消费者(messageformat-go / go-test)无 "我只用 en" 硬约束;tree-shaking 引入实现复杂度(per-locale init 函数 + 全局 map 注册)无收益。
> 2. **Go 没有 `__addLocaleData` 等动态注册** —— 实现该模式需 `init()` + 全局可变 map,违反 SPEC 00 §1 "no implicit state"。
> 3. **Go dead-code elimination** —— Go linker 对未引用的 `var` / `func` 已经能消除,但对 `const string` 引用不可静态消除(因为 `localeDataMap[localeKey]` 这种动态访问)。这一限制在 active scope 通过 "全量嵌入,体积可控" 解决,consumer-driven expansion 通过 build tag 解决。
> 4. **单 binary 部署** —— 不需要部署 sidecar 数据文件;Lambda / Docker 友好。
>
> **Rejected**:
> - **per-locale 子包**(`numberformat/locale-data/zh.js` 风格):Go 没有 `__addLocaleData`;子包间 import 关系污染依赖图。
> - **active scope 就引入 build tag**(`intl_full` / `intl_minimal`):增加 generator 复杂度;主消费者无需求。

### 3.3 consumer-driven expansion 占位

consumer-driven expansion **可以**引入 build tag 分级:

- 默认(无 tag):~100 常用 locale,~1.5 MB(同 active scope)。
- `intl_full`:全部 ~500 locale,~6 MB。
- `intl_minimal`:仅 `root` + `en`,~50 KB。

实现路径:generator 输出 `data_default.go` / `data_full.go` / `data_minimal.go`,各自加 `//go:build` 头。**本 SPEC 不强制 consumer-driven expansion 时间**;SPEC 80 占位。

---

## 4. Version Pinning <a id="version-pin"></a>

### 4.1 决策(关闭 SPEC 00 §8 Q4)

`internal/cldr/VERSION` **必须**单文件存放三个版本号:

```text
cldr=48.1.0
icu=78
tzdata=2025b
```

> **Why**:
> 1. **与 formatjs 同步** —— formatjs 当前主分支 `package.json` 锁 `cldr-*: 48.1.0`(对应 ICU 78);go-intl "与 formatjs 字节级一致" 是 SPEC 00 §1 首要目标。钉早期版本 = 把 "我们对、formatjs 错" 的反向 divergence 制度化。
> 2. **更新窗口** —— CLDR 48 / ICU 78 发布于 2025-10,到公开发布时已稳定 6+ 个月。
> 3. **tzdata 2025b** —— Go 1.26.2 内置 tzdata 版本;与 CLDR / ICU 钉同一时间点的 tzdata 释出。
>
> **Rejected**:
> - **CLDR 47 / ICU 76**(2024-10):比 formatjs 晚一个版本,反向 diverge。
> - **CLDR 44 / ICU 74**(2023-11):落后两年,与 formatjs 字节相等失败率高。
> - **跟随 `golang.org/x/text/cldr` 版本**:数据基线不由 go-intl 控制,会把 conformance 目标交给外部发布节奏。

### 4.2 升级流程

CLDR / ICU / tzdata 任一版本变更 **必须**:

1. 更新 `internal/cldr/VERSION` 三行。
2. 运行 `task data:update` 重新生成 `internal/cldr/*.go`。
3. 运行 `task data:diff` 输出与上一版的字段级 diff。
4. CI 在 main 上 **必须** block 未审查的数据增量(diff 出现在 PR 评审环节)。
5. 同步更新 conformance fixture(`<package>/testdata/conformance/`)如需要。

> **Why**: CLDR 升级常引入百行级数据变更(新 locale、新 dayPeriod、新 currency);静默替换会让一次 PR 改动不可审查。`task data:diff` + CI block 强制评审。

### 4.3 哈希一致性校验

`internal/cldr/VERSION` 中的版本号 **必须**与 `internal/cldr/*.go` 中编码的数据来源对应。CI **必须**通过 `task data:verify` 校验(重新生成与 git 中文件 byte-equal)。

> **Why**: 防止手改 `internal/cldr/*.go` 而忘记更新 `VERSION`;反之亦然。

---

## 5. Generator (`tools/gen-cldr/`) <a id="codegen"></a>

### 5.1 独立 Go module

`tools/gen-cldr/` **必须**是独立 Go module(独立 `go.mod`),**不**污染主 module 依赖图。

```text
go-intl/
├── go.mod              # 主 module,运行时依赖最小:x/text + apd + tzdata + go-cmp
└── tools/
    └── gen-cldr/
        ├── go.mod      # 独立 module;生成期可依赖 encoding/json、第三方解析工具
        ├── main.go     # cmd 入口;CLI flags / config
        ├── run.go      # generator orchestration
        ├── locale_list.go
        ├── extract/
        │   ├── dates.go
        │   ├── numbers.go
        │   ├── metazones.go
        │   ├── likely_subtags.go
        │   ├── matching.go       # SPEC 11 协同
        │   └── locales.go
        ├── codegen/
        │   ├── stringtable.go    # 共享去重字符串表生成
        │   ├── format.go         # gofmt 输出
        │   ├── golang_literal.go # Go 字面量序列化(支持 map / slice / struct)
        │   ├── render.go         # consumer-driven expansion render orchestration
        │   ├── dates.go
        │   ├── numbers.go
        │   ├── metazones.go
        │   └── matching.go
        └── cldr/
            ├── fetch.go          # 拉取 cldr-json npm 包(或 GitHub release)
            ├── source.go         # JSON source loader
            ├── version.go        # 校验 cldr-json 版本与 internal/cldr/VERSION 一致
            ├── dates.go
            ├── numbers.go
            ├── metazones.go
            ├── units.go
            ├── matching.go
            └── preference.go
```

> **Why**:
> 1. **依赖隔离** —— 主 module 不引入 `encoding/json` 重度使用、第三方 CLDR JSON 解析工具(若有);保持运行时依赖图干净。
> 2. **并行演进** —— generator 升级与库代码升级可独立 PR;CI 仅在 generator PR 上跑生成器测试。
> 3. **YAGNI** —— generator 是单向工具(CLDR JSON → Go literal);不需要被外部消费,独立 module 不增加 Go users 心智负担。
>
> **Rejected**:
> - **同主 module**(`internal/gen/`):translate-agent/intl 的方案,但 go-intl generator 需要更多依赖(JSON parser / CLI / fetch),会污染主 module。
> - **独立仓库**(`agentable/go-intl-gen-cldr`):同步 PR 流复杂(SPEC + 数据 + 库三个仓库);CLDR 升级路径割裂。

### 5.2 Codegen 工具栈

`tools/gen-cldr/codegen/` **必须**保持 stdlib-only:用确定性字符串/字面量构造输出,最后经 `go/format` 格式化。**禁止** `dave/jennifer` 或其他 codegen 框架。

```go
// tools/gen-cldr/codegen/golang_literal.go(签名;非完整实现)
package codegen

import (
    "go/format"
    "io"
)

// EmitLiteral 把任意 Go 值序列化为合法 Go 源码字面量(map / slice / struct)。
// 输出走 go/format.Source 保证 gofmt 一致。
func EmitLiteral(w io.Writer, value any) error

// FormatFile 把完整 Go 源文件格式化为 gofmt 后的源码。
func FormatFile(src []byte) ([]byte, error)
```

> **Why**:
> 1. **stdlib 已够用** —— 当前生成物是确定性 Go 字面量与小型 accessor 函数;字符串构造 + `go/format` 比 AST/template 框架更直接。
> 2. **`dave/jennifer` 停滞** —— 2024-09 后无新 commit;新代码不应引入"事实停滞"工具(R09 §4.7)。
> 3. **更少概念,更少漂移** —— 生成器只需要描述数据如何落盘,不需要维护一套额外的 Go AST DSL。
>
> **Rejected**:
> - `dave/jennifer`:见上。
> - `agentable/gendog`:ralphy 内部框架,可在 generator 中评估,但 R06 codegen 路径仅需 "字面量字符串拼接 + Go fmt",stdlib 已够;YAGNI 原则不引入额外框架。

### 5.3 Generator 入口

```go
// tools/gen-cldr/main.go(签名)
package main

type Config struct {
    CLDRDir string  // 本地 cldr-json 解压根目录
    OutDir  string  // internal/cldr 路径(默认 ../../internal/cldr)
    Version string  // CLDR 版本号(从 internal/cldr/VERSION 读取并校验)
}

type Extractor interface {
    Name() string
    Extract(ctx context.Context, src *cldr.Source) (codegen.Module, error)
}

func Run(ctx context.Context, cfg Config, extractors []Extractor) error
```

调用形式:`task data:update`(在 Taskfile 中暴露,**不**在 `task verify` 默认触发,避免 CI 网络依赖)。

### 5.4 输入与输出契约

- **输入**:本地 `cldr-json` 目录(由 CI / 开发者通过 `npm install cldr-*@48.1.0` 或 GitHub release 解压获取)。
- **输出**:`internal/cldr/*.go`(单一目录,不写入仓库其他位置)。
- **副作用**:`internal/cldr/VERSION` **不**由 generator 修改;由人手更新触发 generator。

> **Why**: generator 副作用最小化;`VERSION` 由 PR 作者明确意图后修改,generator 仅消费它。

### 5.5 Locale list extraction

`tools/gen-cldr/extract/locales.go` 负责把 CLDR available locale 与 active allowlist 合并为生成期 locale 记录。

**MUST** 规则:

1. 生成的 locale tag 列表 **必须**确定性排序,并且 `und` **必须**位于第 0 位。`Locale(0)` / `Undefined` 的稳定性依赖该位置。
2. 排序、key 收集、move-to-front 等集合操作 **必须**优先使用 stdlib `maps` / `slices` helper;禁止引入私有 collection helper 或第三方 collection 库。
3. `und` 首位调整 **必须**在排序后显式完成,不能依赖 map iteration 顺序、allowlist 输入顺序或 CLDR JSON 原始顺序。

> **Why**: locale 索引写入 generated Go data,任何非确定性顺序都会造成 byte-diff 噪声,甚至破坏 `cldr.Undefined` 的哨兵语义。stdlib `maps` / `slices` 已经覆盖该需求,不需要维护自定义搬移逻辑。

---

## 6. Data Access API

### 6.1 SSOT

`internal/cldr/` 是所有 formatter 包的数据接入点;访问器 **必须** O(1) 查表,**不**分配。formatter 包不直接读 `internal/cldr/*.go` 文件级 var,**必须**走访问器函数。

```go
// internal/cldr/cldr.go(签名;访问器子集)
package cldr

// Locale 是不透明的 locale 句柄;内部是 dataLocale 索引。
type Locale uint16

// ResolveLocale 把 language.Tag 解析为 dataLocale(BCP 47 → 内部索引)。
// 第二返回值 false 表示该 tag 不在 CLDR 数据集中(此时调用方走 SPEC 11 best-fit fallback)。
func ResolveLocale(tag language.Tag) (Locale, bool)

// AvailableLocales 返回 CLDR availableLocales universe.
func AvailableLocales() []string
// NumberSupportedLocales / DateSupportedLocales 返回实际生成 formatter payload 的 locale 列表(SPEC 11 ResolveLocale 入参)。
func NumberSupportedLocales() []string
func DateSupportedLocales() []string

// Intl.supportedValuesOf 数据访问器(SPEC 60 消费)。
func SupportedCalendars() []string
func SupportedCollations() []string
func SupportedCurrencies() []string
func SupportedNumberingSystems() []string
func SupportedTimeZones() []string

// 数字符号与模式
type NumberSymbols struct {
    Decimal, Group, Percent, Plus, Minus, NaN, Infinity, ApproxSign string
}
func (l Locale) NumberSymbols(numberingSystem string) NumberSymbols
func (l Locale) DecimalPattern(ns string) string
func (l Locale) PercentPattern(ns string) string
func (l Locale) CurrencyPattern(ns, sign string) string
func (l Locale) CompactPattern(ns, display string, exp int, plural string) string

// 日期 / 时间(SPEC 30 / 31 / 32 消费)
type CalendarNames struct{ Eras, Months, Weekdays, DayPeriods, Quarters []string }
func (l Locale) CalendarNames(calendar, width, context string) CalendarNames
func (l Locale) DateFormat(style string) string
func (l Locale) TimeFormat(style string) string
func (l Locale) AvailableFormats() map[string]string  // skeleton -> pattern
func (l Locale) IntervalFormats() IntervalFormats

// 时区(SPEC 32 消费)
func ZoneToMetazone(zone string) string
func (l Locale) MetazoneName(metazone, kind string) string  // kind: "long-generic" / "short-standard" / ...

// 货币(SPEC 20 消费)
func CurrencyDigits(code string) (defaultDigits, cashDigits int, rounding int)
func (l Locale) CurrencyDisplayName(code, plural string) string

// 复数(SPEC 40 消费)— 由 SPEC 40 codegen 输出
func (l Locale) Cardinal(operand Operand) Form
func (l Locale) Ordinal(operand Operand) Form

// Locale Preference(SPEC 10 GetCalendars 等消费)
func (l Locale) HourCyclePreference() []string
func (l Locale) FirstDayOfWeek() time.Weekday
func (l Locale) Weekend() []time.Weekday
func (l Locale) MinimalDaysInFirstWeek() int
func (l Locale) CalendarPreference() []string

// Likely Subtags(SPEC 10 Maximize / Minimize 消费)
func MaximizeSubtags(language, script, region string) (lang, scr, reg string, ok bool)
func MinimizeSubtags(language, script, region string) (lang, scr, reg string, ok bool)

// Locale Matching(SPEC 11 BestFitMatcher 消费)
func MatchingDistance(desired, supported string) int
func ParadigmLocales() []string
func MatchVariables() map[string][]string
```

**MUST** 规则:

1. `AvailableLocales()` 表示 CLDR availability universe;formatter locale matching **必须**使用 formatter-specific 列表,即 `NumberSupportedLocales()` 或 `DateSupportedLocales()`。
2. `NumberSupportedLocales()` / `DateSupportedLocales()` **必须**由 generated runtime data maps 派生,并且每个 tag 都能通过 `ResolveLocale` 找到非 `Undefined` locale。
3. `internal/cldrmatch` **禁止**维护手写 `numberLocales` / `dateLocales` 列表;新增 formatter 数据后必须在 CLDR generator 里增加对应 supported-locale accessor。
4. `SupportedCollations()` **必须**由 `cldr-bcp47` 生成,不得在 runtime 手写 `emoji` / `eor` 等临时列表。

> **Why**: `availableLocales.json` 可能包含某些 formatter 尚未生成完整 numbers/date 数据的 locale。supported locale 列表必须来自实际生成的数据,否则 `ResolveLocale` 可以命中一个没有 formatter payload 的 locale,导致 fallback 行为与 formatjs 漂移。

### 6.2 zero-allocation 约束

所有访问器 **必须** 零分配(返回 `string` 是 string slice into `_data`,返回 `[]string` 通过 Go 编译器对常量数组的优化处理;返回 `map[string]string` 通过 package-level `var` 持有)。

CI 通过 `go test -benchmem -run=^$ -bench=BenchmarkAccessor ./internal/cldr/` 监控,任何访问器 ≥ 1 alloc/op 阻断 PR。

> **Why**: hot path 性能。NumberFormat / DateTimeFormat 在 `Format` 调用中读取 ~10 个 CLDR 字段;每读取一个分配一个 `string` 头会让性能阈值(SPEC 71)失败。
>
> **Rejected**:
> - 返回 copy-on-read(深拷贝):每次 `Format` 调用产生 GC 压力。
> - 接口返回(`type DataReader interface { ... }`):接口方法调用比直接 method dispatch 慢,且无收益。

### 6.3 公共可见性

`internal/cldr/` 通过 `internal/` 路径段强制私有。formatter 公共包 **不**重导出 `cldr.Locale` / `cldr.NumberSymbols` 等类型;通过 `internal/ecma402` 抽象层间接消费。

CLDR / ICU / tzdata pins **必须**保留在 `internal/cldr.Version()` 与 `internal/cldr/VERSION` 中供内部测试、审计和报告使用;根包 **不得**公开 `intl.Version()` 或其他诊断 API,因为 ECMA-402 `Intl` 命名空间没有对应成员。

---

## 7. Testing

### 7.1 Generator 自测

`tools/gen-cldr/` **必须**自带 `gen_test.go`,断言生成结果中"en/zh/ar 的 era/month 字段不为空"等基本契约:

```go
// tools/gen-cldr/gen_test.go(签名)
func TestExtractDates_BasicLocales(t *testing.T) {
    t.Parallel()
    // 跑 extract,断言 en / zh-Hans / ar 的 month 名长度 >= 12 等
}
```

### 7.2 快照测试

`internal/cldr/snapshot_test.go` **必须**比对当前 git 中的 `internal/cldr/*.go` 与重新生成的输出,CI 拒绝差异(避免手改):

```go
// internal/cldr/snapshot_test.go(签名)
func TestGenerated_NoDrift(t *testing.T) {
    t.Parallel()
    // 调 tools/gen-cldr 重新生成到临时目录;byte-compare 与 git 中文件
}
```

### 7.3 Conformance fixture 集成

`internal/cldr/` 数据接入正确性 **必须**通过下游 formatter 的 conformance 测试间接验证(SPEC 70)。本 SPEC 不重复 fixture 设计。

---

## Forbidden

- **`//go:embed *.json` + 运行时反序列化**:违反 §2.1 嵌入策略。
  - ✅ Do: `tools/gen-cldr/` 在生成期序列化为 Go 字面量。
  - ❌ Don't: `//go:embed numbers.json` + `var _ = json.Unmarshal(numbersJSON, ...)`。

- **运行时网络拉取 CLDR 数据**(从 CDN / S3):破坏 offline 部署、违反 SPEC 00 §1 单 binary 哲学。
  - ✅ Do: 全量嵌入(active scope)/ build tag 拆分(consumer-driven expansion)。
  - ❌ Don't: `http.Get("https://cdn.example/cldr-data/...")`。

- **启动时反序列化**(包括 `init()` 中 JSON parse):破坏冷启动时间。
  - ✅ Do: Go 字面量在 `.rodata` 段,O(1) 访问。
  - ❌ Don't: `func init() { json.Unmarshal(_embedded, &globalData) }`。

- **sub-locale 懒加载**(运行时 file I/O 拉取 `zh-Hans` 数据):违反 SPEC 00 §5.3 "Locale data is universal in active scope"。
  - ✅ Do: 全量嵌入。
  - ❌ Don't: `func loadLocaleData(loc string) { os.Open(...) }`。

- **`golang.org/x/text/cldr`**:数据基线不由 go-intl 控制,与 formatjs 钉版可能 diverge。
  - ✅ Do: 直接消费 `unicode-org/cldr-json`。
  - ❌ Don't: `import "golang.org/x/text/cldr"`(在 `internal/cldr/` 与 `tools/gen-cldr/` 中)。

- **formatjs 中间产物**(`@formatjs_generated/*`):不发到 npm,运维不可行。
  - ✅ Do: 直接消费上游 `cldr-json`。
  - ❌ Don't: 在 `tools/gen-cldr/` 中 `npm install @formatjs_generated/...`。

- **`dave/jennifer` codegen**:2024-09 后停滞;当前规模不需要 codegen 框架。
  - ✅ Do: stdlib JSON 读取 + 确定性字符串输出 + `go/format`。
  - ❌ Don't: `import "github.com/dave/jennifer/jen"`(在 `tools/gen-cldr/`)。

- **`bojanz/currency` 作为 ECMA-402 数据源**:自维护 CLDR 派生表,与 `internal/cldr/VERSION` 漂移。
  - ✅ Do: CLDR `currencyData.json` 直接生成 `internal/cldr/currencies.go`。
  - ❌ Don't: `import "github.com/bojanz/currency"`(在 `numberformat/` 实现中)。

- **per-locale tree-shaking 在 active scope**(formatjs `__addLocaleData` 风格):无 Go 等价机制 + 主消费者无需求。
  - ✅ Do: 全量嵌入(active scope);build tag 分级(consumer-driven expansion)。
  - ❌ Don't: active scope 引入 `init()` 注册全局 map。

- **`internal/cldr/*.go` 手改而不更新 `VERSION`**:破坏数据来源可审计。
  - ✅ Do: `VERSION` + `task data:update` 两步流程,CI 校验 byte-equal。
  - ❌ Don't: 直接编辑 `internal/cldr/numbers.go`。

- **生成器位于主 module**(`internal/gen/`):污染主 module 依赖图。
  - ✅ Do: `tools/gen-cldr/` 独立 `go.mod`。
  - ❌ Don't: 把 generator 文件放在 `internal/gen/`。

---

## Acceptance Criteria

### 数据源

- [ ] `internal/cldr/VERSION` 文件存在,内容为 `cldr=48.1.0\nicu=78\ntzdata=2025b\n`(三行)。
- [ ] `tools/gen-cldr/cldr/version.go` 启动时校验本地 `cldr-json` 目录的 `package.json` 版本 = `VERSION` 中的 `cldr=` 值,不一致则 fail。
- [ ] `internal/cldr/` 内**不**包含 `*.json` 文件;**不**出现 `//go:embed` 指令(grep `go:embed` in `internal/cldr/*.go` 返回 0)。

### 包结构

- [ ] `internal/cldr/` 文件清单与 §1.2 表格一致(`numbers.go` / `dates.go` / `metazones.go` / `timezones.go` / `plurals.go` / `plural/*.go` / `preference.go` / `likely_subtags.go` / `locale_matching.go` / `regions.go` / `currencies.go` / `units.go` / `locales.go` / `supported.go` / `strings.go` 等)。
- [ ] `internal/cldr/` 通过 `internal/` 路径段强制 Go 工具链拒绝外部 import。
- [ ] `internal/cldr.Version()` 返回 `VERSION` 内容;根包不公开版本诊断函数。

### Generator

- [ ] `tools/gen-cldr/go.mod` 存在(独立 Go module)。
- [ ] `tools/gen-cldr/main.go` 入口存在;`task data:update` 在 `Taskfile.yml` 中暴露并触发 generator。
- [ ] `tools/gen-cldr/codegen/` 仅 import stdlib,使用确定性字符串输出 + `go/format`;**不** import `github.com/dave/jennifer` 或其他 codegen 框架。
- [ ] `tools/gen-cldr/extract/locales.go` 输出 locale tag 列表确定性排序,且 `und` 永远位于第 0 位。
- [ ] `tools/gen-cldr/` 的局部验证从嵌套 module 根目录执行:`cd tools/gen-cldr && go test ./...` 与 `go vet ./...`;根 module 的 `./tools/gen-cldr/...` pattern 不作为验收入口。

### 嵌入与体积

- [ ] `internal/cldr/strings.go` 含单一 `const _data string` 共享去重字符串表;访问器通过 `[start:end]` 切片访问。
- [ ] `task build:size` 报告 go-intl 主 binary 比无 CLDR 数据基线增长 ≤ 1.5 MB。
- [ ] active scope 嵌入约 100 个 modern tier 1 / 2 locale(具体清单由 `tools/gen-cldr/locale_list.go` 维护)。

### 升级流程

- [ ] `task data:update` 重新生成 `internal/cldr/*.go`(任意运行后无 git diff,除 `VERSION` 已先变更)。
- [ ] `task data:diff` 输出与上次 commit 的字段级 diff(供 PR 审查)。
- [ ] `task data:verify`(CI)校验 `internal/cldr/*.go` 与重新生成结果 byte-equal,不一致 fail。
- [ ] CI 在 main 上 block 数据增量 PR 直到 reviewer 批准。

### 访问器

- [ ] §6.1 列出的访问器全部声明(`ResolveLocale` / `NumberSymbols` / `CalendarNames` / `IntervalFormats` / `ZoneToMetazone` / `MetazoneName` / `CurrencyDigits` / `MaximizeSubtags` / `MatchingDistance` 等)。
- [ ] `NumberSupportedLocales()` 与 `DateSupportedLocales()` 非空,不含 `und`,且每个 tag 都有对应 generated number/date payload。
- [ ] `SupportedCalendars()` / `SupportedCollations()` / `SupportedCurrencies()` / `SupportedNumberingSystems()` / `SupportedTimeZones()` 返回 canonical、sorted、unique values,并且数据来自 generated runtime maps 或 ECMA-402 generator constants。
- [ ] `SupportedCalendars()` 包含 `iso8601`;`SupportedNumberingSystems()` 包含 ECMA-402 simple digit numbering systems 全表,即使当前 profile 未生成对应 CLDR symbol payload。
- [ ] `go test -benchmem -run=^$ -bench=BenchmarkAccessor ./internal/cldr/` 报告所有访问器 0 alloc/op。

### 测试

- [ ] `tools/gen-cldr/gen_test.go` 通过;断言基本 locale(en / zh-Hans / ar)的 era / month 字段非空。
- [ ] `internal/cldr/snapshot_test.go` 通过;CI 校验生成结果与 git 一致。
- [ ] formatjs `tests/likely-subtags.test.ts` / `minimize.test.ts` 全部 fixture 在 `internal/cldr/likely_subtags_test.go` 通过。
- [ ] formatjs `intl-localematcher/tests/best-fit-matcher.test.ts` 全部 fixture 在 SPEC 11 测试中通过(本 SPEC 提供 `MatchingDistance` 数据)。

---

## References

### Specification

- [Unicode CLDR](https://cldr.unicode.org/) —— 数据源。
- [CLDR JSON Distribution](https://github.com/unicode-org/cldr-json) —— `cldr-bcp47` / `cldr-core` / `cldr-dates-full` / `cldr-numbers-full` / `cldr-units-full` 等 npm 包。
- [ECMA-402 §6 — Locale and Currency Identifiers](https://tc39.es/ecma402/#locale-and-currency-identifiers) —— 标识符规范。

### Reference implementations

- `.references/formatjs/package.json` —— `cldr-*: 48.1.0` 锁定证据。
- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` —— CLDR JSON import 清单与抽取逻辑。
- `.references/formatjs/packages/intl-numberformat/scripts/extract-currencies.ts` —— currency 抽取。
- `.references/formatjs/knowledge-base/001-repo-layout.md` —— `@formatjs_generated/*` 包架构。
- `.references/intl/internal/cldr/data.go` —— 400 KB / 3203 行 Go 字面量先例(translate-agent/intl)。
- `.references/intl/internal/gen/main.go` —— text/template 模板生成器先例。

### Cross-SPEC

- [SPEC 00 §5.3 — Data strategy](./00-vision-and-scope.md#53-data-strategy)
- [SPEC 10 §Maximize / Minimize](./10-locale.md) —— 消费 `MaximizeSubtags` / `MinimizeSubtags`。
- [SPEC 11 §BestFitMatcher Data](./11-locale-matching.md) —— 消费 `MatchingDistance` / `ParadigmLocales` / `MatchVariables`。
- [SPEC 20 §Currency Data](./20-numberformat.md) —— 消费 `CurrencyDigits` / `CurrencyDisplayName`。
- [SPEC 30 §DateTimeFormat Core](./30-datetimeformat.md) —— 消费 `CalendarNames` / `DateFormat` / `TimeFormat` / `AvailableFormats` / `IntervalFormats`。
- [SPEC 31 §Skeleton Resolution](./31-datetimeformat-skeleton.md) —— 消费 `AvailableFormats` / `IntervalFormats`。
- [SPEC 32 §TimeZone & Calendar Data](./32-datetimeformat-tz.md) —— 消费 `metazones.go`,与 `time/tzdata` 集成。
- [SPEC 40 §PluralRules](./40-pluralrules.md) —— `plurals.go` 与 `plural/*.go` 由 SPEC 40 codegen 输出到本目录。

### Research

- `.research/R06-cldr-data-strategy.md` —— 数据源选型、嵌入策略、版本钉、generator 架构。
- `.research/R09-dependencies.md` §4.3 / §4.5 —— 数据源 / 货币库选型 gh 印证。

---

> 本 SPEC 是 CLDR 数据底层的 SSOT。版本钉变更触发 SPEC 修订;active scope locale 清单变更通过 `tools/gen-cldr/locale_list.go` 维护(不触发 SPEC 修订),但体积上限与全量嵌入策略不可改。
