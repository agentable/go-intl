---
id: R01
title: ECMA-402 抽象操作层调研 — 命名策略、依赖图与 Go 包布局
task: r01
date: 2026-05-08
status: draft
scope:
  - ECMA-402 抽象操作的依赖图与"叶子操作 vs Initialize* 入口"分层
  - GetOption 选项校验流水线在所有 formatter 间的复用方式
  - WeakMap 内部槽（internal slots）在 TypeScript 与 Go 中的语义对齐
  - 命名策略：保留 spec 原名 vs Go 习惯命名
  - 哪些行为属于 ECMA-402 规范、哪些是实现优化
  - `internal/ecma402/` 候选 Go 包布局
tags: [ecma402, abstract-ops, internal-package, naming, formatjs, php-ext-intl]
---

# R01 — ECMA-402 抽象操作层调研

## 执行摘要

| 决策点 | 推荐方案 | 置信度 | 依据 |
|--------|----------|--------|------|
| 命名策略 | 抽象操作函数名严格沿用 ECMA-402 spec 原名（PascalCase 即 Go 大写导出形式天然兼容），保留在 `internal/ecma402/`；公共 API（`numberformat.New` 等）使用 Go 习惯命名 | High | FormatJS 与 PHP ext/intl 均锚定 spec 原名；SPEC 00 §5.2 第 3 条已要求"函数名匹配 spec" |
| 抽象操作分层 | 分两层：`internal/ecma402/`（根，叶子工具）与 `internal/ecma402/{numberformat,datetimeformat,pluralrules}/`（formatter 专属操作） | High | 与 formatjs 三个 Bazel 子包（NumberFormat/、DateTimeFormat/、PluralRules/）一一对应，依赖方向单向 |
| 内部槽实现 | 直接用 Go struct 字段；不再使用 WeakMap-style 间接表 | High | TS 的 WeakMap 是为模拟 [[Slot]] 的 ES 私有性，Go 没有该约束；`internal/` 包对外不可见已等价 |
| 选项校验流水线 | `GetOption[T]` 泛型 + `(string|bool, []T, T)` 显式契约，避免 `any`；保留 `RangeError`/`TypeError` 两类错误为 `ErrInvalidOption` 的不同包装 | High | formatjs/GetOption.ts 的 `'string' \| 'boolean'` 二元参数即语义边界；Go 1.26 generics 足以覆盖 |
| 类型放置 | `internal/ecma402/types/`（与 formatjs 的 `types/` 子包同构），含选项枚举、`Pattern`、`Part`、`MathematicalValue` | Medium | formatjs 把跨 formatter 复用的 type 单独成包，避免循环依赖；Go 也需要类似结构 |
| `Decimal` 与 `MathematicalValue` 边界 | `internal/ecma402` 只定义 `MathematicalValue` 接口；具体实现下沉到 `internal/decimal` | High | `ToIntlMathematicalValue` 是 spec 的边界，实现可换；与 SPEC 00 §8 Q1 解耦 |

**前提**：本报告聚焦 SPECS §8 Q2（命名策略）的子问题——它在 §8 中没有显式列出但散见于 §5.2，本报告将其视为"待入 SPEC 12 — Abstract Operations"的开放项并给出推荐。

## 1. 抽象操作的依赖分层

四个参考项目都以"叶子工具 → mid-tier 校验 → formatter Initialize* 入口"的三层模型组织抽象操作，但物理边界差异显著。

### 1.1 formatjs 的物理分层

formatjs `ecma402-abstract` 是 1 个 npm 包但 3 个 Bazel 复合子包（`root`、`types/`、`NumberFormat/` 等），通过 `#packages/ecma402-abstract/*.js` 子路径导入；它**不发布 npm**。<!-- ref: formatjs/CLAUDE.md, formatjs/knowledge-base/001a-bazel-toolchain.md:Composite Subpackages -->

```text
ecma402-abstract/                ← Layer 1（leaf 工具 + 共享类型）
├── GetOption.ts                 ← 唯一的 string/boolean 校验入口
├── DefaultNumberOption.ts       ← 数字类型校验
├── GetNumberOption.ts           ← 委托给 DefaultNumberOption
├── GetStringOrBooleanOption.ts  ← 处理 useGrouping 这种联合
├── GetOptionsObject.ts          ← 选项对象与 undefined 兼容
├── CoerceOptionsToObject.ts     ← undefined → {}
├── PartitionPattern.ts          ← 解析 "{key}" 模板
├── ToIntlMathematicalValue.ts   ← 输入归一化到 Decimal
├── CanonicalizeLocaleList.ts    ← 委托给 Intl.getCanonicalLocales
├── SupportedLocales.ts          ← lookup/best-fit 的统一入口
├── Is*Identifier.ts             ← 货币/单位/时区/计算系统校验
├── CanonicalizeTimeZoneName.ts
├── utils.ts                     ← setInternalSlot / getInternalSlot WeakMap
├── constants.ts, data.ts        ← 共享常量
├── types/                       ← 跨 formatter 类型（不发布）
└── {NumberFormat,DateTimeFormat,PluralRules,...}/   ← Layer 2（formatter 专属）
    ├── Initialize*.ts           ← formatter 构造入口
    ├── Partition*Pattern.ts     ← 输出阶段
    ├── Format*.ts / Format*ToParts.ts
    └── 其他算法（ToRawFixed、SetNumberFormatDigitOptions 等）
```

`NumberFormat/` 子包专门收纳 `InitializeNumberFormat.ts`、`SetNumberFormatDigitOptions.ts`、`PartitionNumberPattern.ts`、`ToRawFixed.ts`、`ToRawPrecision.ts`、`ApplyUnsignedRoundingMode.ts`、`ComputeExponent.ts`、`CurrencyDigits.ts` 等十余个文件。<!-- ref: formatjs/packages/ecma402-abstract/NumberFormat/ -->

依赖方向严格单向：`NumberFormat/` 可以 `import` 根目录的 `GetOption` / `utils`，但根目录不反向 `import` 任何 formatter 子目录。这与 SPEC 00 §5.2 第 1 条"`internal/*` 永远不依赖公共包"的方向语义相同。

### 1.2 PHP `ext/intl` 的分层

PHP 把所有抽象操作集中在 `src/ecma402/` 下，每类单独成对（`util.{c,h}`、`error.{c,h}`、`category.{c,h}`、`locale.{c,h}`、`hour_cycle.{c,h}`、`time_zone.{c,h}` …），所有函数前缀 `ecma402_*`。<!-- ref: ext/src/ecma402/util.h, ext/src/ecma402/error.h, ext/src/ecma402/category.h, ext/src/ecma402/locale.h -->

错误模型是关键差异：

```c
typedef struct ecma402_errorStatus {
    ecma402_errorType ecma;   // ECMA-402 RangeError / TypeError 枚举
    UErrorCode icu;           // ICU 错误码
    char *errorMessage;
} ecma402_errorStatus;
```

PHP 把"ECMA-402 抽象错误类型"和"ICU 底层错误"作为独立字段共存——这是因为 PHP 既要给 PHP 用户抛 `IntlException`，又要透传 ICU 诊断。Go 不会有 ICU，但"`ErrInvalidOption` 是 ECMA-402 RangeError，`ErrUnsupportedLocale` 是 ECMA-402 TypeError"的语义区分仍然有意义。

### 1.4 translate-agent/intl 的分层

`translate-agent/intl` 没有抽象操作层——它直接基于 `language.Tag` 写 DateTimeFormat 业务，连 `GetOption` 等价物都没有，因为它只支持一种 formatter。这恰好提示我们："如果 `go-intl` 只做单 formatter，可以省掉 `internal/ecma402/`；但 SPEC 00 要求支持 4 个 formatter，必须设独立层"。<!-- ref: intl/intl.go:1-200 -->

### 1.5 推荐 Go 分层

```text
internal/
├── ecma402/                     ← 叶子工具（formatter 无关）
│   ├── option.go                ← GetOption / GetNumberOption / DefaultNumberOption / GetStringOrBooleanOption / GetOptionsObject / CoerceOptionsToObject
│   ├── partition.go             ← PartitionPattern
│   ├── identifier.go            ← IsWellFormedCurrencyCode / IsValidTimeZoneName / IsSanctionedSimpleUnitIdentifier / IsWellFormedUnitIdentifier
│   ├── timezone.go              ← CanonicalizeTimeZoneName
│   ├── locale_list.go           ← CanonicalizeLocaleList / SupportedLocales
│   ├── math_value.go            ← ToIntlMathematicalValue（接口边界）
│   ├── types/                   ← Pattern / Part / MathematicalValue 接口、共享枚举
│   └── constants.go             ← SANCTIONED_UNITS、字段常量
├── ecma402/numberformat/
│   ├── initialize.go            ← InitializeNumberFormat
│   ├── digit_options.go         ← SetNumberFormatDigitOptions
│   ├── unit_options.go          ← SetNumberFormatUnitOptions
│   ├── partition.go             ← PartitionNumberPattern / PartitionNumberRangePattern
│   ├── format.go                ← FormatNumeric / FormatNumericToParts / FormatNumericRange / FormatNumericRangeToParts
│   ├── raw.go                   ← ToRawFixed / ToRawPrecision
│   ├── rounding.go              ← ApplyUnsignedRoundingMode / GetUnsignedRoundingMode
│   ├── exponent.go              ← ComputeExponent / ComputeExponentForMagnitude
│   └── currency.go              ← CurrencyDigits
├── ecma402/datetimeformat/
│   ├── initialize.go            ← InitializeDateTimeFormat / ToDateTimeOptions
│   ├── skeleton.go              ← skeleton 解析
│   ├── format_matcher.go        ← BasicFormatMatcher / BestFitFormatMatcher
│   ├── partition.go             ← PartitionDateTimePattern / PartitionDateTimeRangePattern
│   └── format.go                ← FormatDateTime / FormatDateTimeRange / FormatDateTimePattern / 各 ToParts
├── ecma402/pluralrules/
│   ├── initialize.go            ← InitializePluralRules
│   ├── operands.go              ← GetOperands
│   └── resolve.go               ← ResolvePlural / ResolvePluralRange
└── decimal/                     ← MathematicalValue 实现（SPEC 21 决定）
```

四个子目录对应 formatjs 四个 Bazel 子包；`internal/ecma402/types/` 对应 formatjs `types/` 复合子包。这一布局的优点：

- **可审计**：每个 Go 文件可一对一映射到 formatjs 的 `.ts` 文件，方便 reviewer 比对算法。
- **依赖单向**：`numberformat/` 等子包只能 `import "go-intl/internal/ecma402"` 和 `internal/ecma402/types`，不互相 import；与 SPEC 00 §5.2 一致。
- **未来扩展**：Phase 3 加入 `relativetimeformat/` / `listformat/` 等只需要新增同级子目录。

## 2. 选项校验流水线

### 2.1 三方对比

| 项目 | 入口签名 | 错误类型 | 数字类型扩展 | 字符串/布尔联合 |
|------|---------|---------|------------|---------------|
| formatjs | `GetOption<T,K,F>(opts, prop, 'string'\|'boolean', values?, fallback): ...` | `RangeError`/`TypeError` | `DefaultNumberOption` 单独函数 | `GetStringOrBooleanOption` 单独函数 |
| PHP ext/intl | 用 zend 参数解析宏（`Z_PARAM_*`），ECMA-402 的 `GetOption` 语义被分散到各 setter | `IntlException` | `Z_PARAM_LONG` 配合手动校验 | 各 setter 内联处理 |
| translate-agent | 无 | 无 | 无 | 无 |

formatjs 的 `GetOption.ts` 用 `'string'|'boolean'` 二元参数 + 可选 `values` 集合 + `fallback` 三段契约，是最贴近 spec 原文的写法。<!-- ref: formatjs/packages/ecma402-abstract/GetOption.ts:11-38 -->

```typescript
// formatjs 签名摘录
export function GetOption<T extends object, K extends keyof T, F>(
  opts: T,
  prop: K,
  type: 'string' | 'boolean',
  values: readonly T[K][] | undefined,
  fallback: F,
): Exclude<T[K], undefined> | F
```

### 2.2 推荐 Go 形式

Go 1.26 generics + `comparable` 约束足以重现该契约，避免 `any`：

```go
// internal/ecma402/option.go（仅展示签名）
type OptionType int
const (
    OptionString OptionType = iota
    OptionBoolean
)

// GetOption 对应 spec sec-getoption。values 为 nil 时不做枚举校验。
func GetOption[T comparable](
    opts map[string]any,
    prop string,
    typ OptionType,
    values []T,
    fallback T,
) (T, error)

// DefaultNumberOption / GetNumberOption 单独签名（与 spec 同名）。
func DefaultNumberOption(value any, minimum, maximum, fallback int) (int, error)
func GetNumberOption(opts map[string]any, prop string, minimum, maximum, fallback int) (int, error)
```

调用侧示例（仅 3 行，不是实现）：

```go
// 在 InitializeNumberFormat 中
style, err := ecma402.GetOption(rawOpts, "style", ecma402.OptionString,
    []string{"decimal", "percent", "currency", "unit"}, "decimal")
```

但是 Go 包并不会让 formatter 用户传 `map[string]any`——这是构造器内部的归一化形式。公共 API（`numberformat.New`）用 functional options 或 config struct（SPEC 00 §8 Q2，由 SPEC 20 决定），构造器在内部 build 出 `map[string]any` 喂给 `Initialize*`。**这一点是关键设计契约**：抽象操作层只接受归一化输入，公共层负责 Go 习惯到 spec 输入的转换。

### 2.3 错误模型

formatjs 的两类错误（`TypeError`、`RangeError`）在 Go 应映射为：

- `ErrInvalidOption`（root `errors.go`）= ECMA-402 `RangeError`，常用于值不在枚举内或越界。
- `ErrInvalidOptionType`（root `errors.go`）= ECMA-402 `TypeError`，常用于 `opts` 不是对象、`type !== 'boolean'|'string'`。

Wrap 时提供 spec sentinel：`fmt.Errorf("ecma402: option %q value %q not in %v: %w", prop, value, values, ErrInvalidOption)`。这样既能 `errors.Is` 匹配，又向上保留 spec 类型信息——与 PHP ext/intl 的 `ecma402_errorStatus.ecma` 字段语义对齐。

## 3. 内部槽（Internal Slots）实现

### 3.1 TS WeakMap 模式

formatjs 用 `WeakMap<Instance, InternalSlots>` 模拟 ES 的 `[[Field]]`：<!-- ref: formatjs/packages/ecma402-abstract/utils.ts:15-78 -->

```typescript
export function setInternalSlot<I extends object, S extends object, F extends keyof S>(
  map: WeakMap<I, S>, pl: I, field: F, value: NonNullable<S>[F],
): void

export function getInternalSlot<I extends object, S extends object, F extends keyof S>(
  map: WeakMap<I, S>, pl: I, field: F,
): S[F]
```

每个 formatter 模块（`intl-numberformat`、`intl-locale` 等）声明 `const internalSlotMap = new WeakMap()`，其 `index.ts` 在构造时调用 `setInternalSlot`，运行时通过 `getInternalSlot` 取值。

**为什么 TS 这样做？** ES 没有"私有字段对外不可见"的运行期机制（直到 ES2022 `#field`，但 polyfill 必须兼容老版本）。WeakMap 提供"实例 → 私有数据"的映射，且 GC 安全。

### 3.2 PHP 对比

PHP `ecma402_locale` 是平铺的 C struct，所有字段（`baseName`、`calendar`、`hourCycle` …）作为 `char*` 直接存活在栈/堆对象上。<!-- ref: ext/src/ecma402/locale.h -->

### 3.3 Go 推荐：直接 struct 字段

Go 没有"对外可见的实例 + 隐藏内部状态"的间接需要：`internal/` 包对外不可见已经达成 spec WeakMap 的封装目标。因此每个 formatter 的"internal slots"就是 Go struct 的非导出字段：

```go
// numberformat/numberformat.go
type NumberFormat struct {
    locale          locale.Locale  // [[Locale]]
    style           string         // [[Style]]
    currency        string         // [[Currency]]
    minIntDigits    int            // [[MinimumIntegerDigits]]
    // ... 其余 [[InternalSlot]] 全部成 unexported field
}
```

无需 `WeakMap`、无需 `sync.Map`、无需独立的 slot helper：构造器返回 `*NumberFormat`，`Format` 方法直接读字段。这是与 PHP 一致的"struct 内联"模型，仅有的不同是 Go 没有 ICU，所以字段会更"逻辑"而非"句柄"。

**唯一例外**：locale 的"惰性 getter"——`Intl.Locale.getCalendars()` 在 formatjs 中**不是**预存到 internal slot 而是每次调用时计算。Go 也应保持惰性，但实现是普通 method（详见 R02 §3）。

## 4. 命名策略

### 4.1 现状

四方一致采用 spec 原名（PascalCase），无人改写：

| spec 名称 | formatjs 文件 | PHP 函数 |
|----------|--------------|---------|
| `GetOption` | `GetOption.ts` 中导出的 `GetOption` | 分散到各 setter，无直接对应 |
| `CanonicalizeLocaleList` | `CanonicalizeLocaleList.ts` | `ecma402_canonicalizeLocaleList` |
| `PartitionPattern` | `PartitionPattern.ts` | 无（PHP 不实现 formatToParts） |
| `BestAvailableLocale` | `BestAvailableLocale.ts` | `ecma402_bestAvailableLocale` |

PHP 的 `ecma402_*` 前缀是 C 命名空间约束所致。在 Go 里，`internal/ecma402` 包路径已经提供了等价命名空间，函数名直接用 `GetOption`（导出）即可。

### 4.2 Go 习惯化的代价

如果偏离 spec 名（如把 `GetOption` 改成 `OptionValue` 或 `ParseOption`）会造成：

1. **审计成本**：reviewer 需要在 spec / formatjs 与 Go 代码之间维护映射表。
2. **bug 通信成本**：spec issue 引用 `GetOption` 时，调研者要二次翻译。
3. **跨语言移植抖动**：未来 Phase 3 增加 `RelativeTimeFormat` 时，复用 formatjs 的 `InitializeRelativeTimeFormat` 算法仍然要按 spec 名字搜索。

### 4.3 推荐

**对 `internal/ecma402/`**：所有函数名 verbatim 沿用 spec/formatjs 名（`GetOption`、`CanonicalizeLocaleList`、`InitializeNumberFormat`、`PartitionPattern`、`ToRawFixed` …）。Go 的"导出函数 PascalCase"恰好与 spec 大写头一致，无需任何转写。

**对公共 API（`numberformat.New`、`locale.Parse`、`intl.FormatNumber`）**：使用 Go 习惯命名。`Intl.NumberFormat` → `numberformat.New`，`Intl.Locale.maximize()` → `(Locale).Maximize() Locale`。

**置信度：High**。这与 SPEC 00 §5.2 第 3 条"函数名匹配 spec"已写明的方向一致；现需在 SPEC 12（待新建）中正式记入。

## 5. 规范行为 vs 实现优化

研究中发现以下行为容易被误当成必须照抄的实现细节：

| 行为 | 来源 | go-intl 选择 |
|------|------|-------------|
| `BestFitFormatMatcher` 的字段权重表 | spec 留 implementation-defined（注释说 "best fit may use any algorithm"） | 见 R04，可移植 formatjs 的表，或更换 |
| `decimal-cache` 复用 `BigDecimal` 实例 | formatjs 的性能优化 | Go 不需要——复用由 GC 管理 |
| `setInternalSlot` 抛 `TypeError` 'has not been initialized' | formatjs 的运行时保护，不是 spec | Go 用 nil 接收者检查或构造时强制 |
| `RangeError` vs `TypeError` 的精确划分 | spec normative | 必须保留语义（用两个 sentinel error） |
| `IsValidTimeZoneName` 大小写不敏感比较 | spec normative | 必须实现（见 `IsValidTimeZoneName.ts:9-13`） |
| `CanonicalizeTimeZoneName` 把 `Etc/UTC` 归并为 `UTC` | spec normative | 必须实现 |
| `IsSanctionedSimpleUnitIdentifier` 的硬编码列表 | spec 在 ECMA-402 sec-iswellformedunitidentifier 显式列出 | 必须保留 verbatim 列表 |

判别准则：**spec 文本里写明的 → 必须实现；spec 标为 implementation-defined 或属于运行时内部行为 → 可以替换**。

## 6. 对本项目的落地建议

### 6.1 包布局（汇总）

按 §1.5 给出的双层布局即可。重点强调：

- `internal/ecma402` 根目录只放**与 formatter 无关**的工具；任何带 `NumberFormat`/`DateTimeFormat`/`PluralRules` 名字的函数都放进对应子目录。
- `internal/ecma402/types/` 放跨 formatter 共享类型（`Part`、`Pattern`、`MathematicalValue` 接口）；每个 formatter 自己的选项 struct 放在该 formatter 的子目录里（`internal/ecma402/numberformat/options.go`）或 formatter 公共包（`numberformat/options.go`）。
- `internal/decimal` 与 `internal/ecma402` 同级；`ecma402` 通过 `MathematicalValue` 接口隔离 decimal 实现。

### 6.2 接口契约（关键签名草案）

```go
// internal/ecma402/types/types.go
type Part struct {
    Type  string // "literal" | "integer" | "fraction" | "decimal" | "currency" | ...
    Value string
}

type Pattern []Part
type MathematicalValue interface {
    IsNaN() bool
    IsInfinity() bool
    Sign() int
    // 由 internal/decimal 实现
}
```

```go
// internal/ecma402/option.go  ← 只展示签名
func GetOption[T comparable](opts map[string]any, prop string, typ OptionType,
    values []T, fallback T) (T, error)
func GetNumberOption(opts map[string]any, prop string, min, max, fallback int) (int, error)
func GetStringOrBooleanOption[T comparable](opts map[string]any, prop string,
    stringValues []T, trueValue, falsyValue T) (T, error)
```

```go
// internal/ecma402/numberformat/initialize.go  ← 只展示签名
type Internal struct { /* unexported [[Slot]]s */ }

func Initialize(loc locale.Locale, opts map[string]any) (*Internal, error)
func PartitionPattern(internal *Internal, x ecma402types.MathematicalValue) ecma402types.Pattern
```

### 6.3 与 SPEC 00 的耦合

- SPEC 00 §5.1（包布局）已为 `internal/ecma402/` 留出位置，本报告的子目录划分是兼容细化，**不需要改 SPEC 00**。
- SPEC 00 §5.2 第 3 条"函数名匹配 spec"——本报告把它正式化为命名策略推荐。
- SPEC 00 §8 Q2 / Q5（`numberformat.New` 选项 API、最佳匹配实现）由其他 SPEC（20 / 11）裁决；本报告不越俎。

### 6.4 验证策略

- **单元测试**：每个 spec 抽象操作配 `*_test.go`，移植 formatjs 在对应 `tests/` 下的输入输出对（`GetOption.test.ts` / `PartitionPattern.test.ts` / `IsValidTimeZoneName.test.ts` 等）。
- **冒烟测试**：在 `internal/ecma402` 根写一个 `audit_test.go`，列出所有应实现的 spec 抽象操作并断言对应 Go 函数已声明（防止漏函数）。

## 7. 决策矩阵

| 决策 | 推荐方案 | 备选方案 | 否决方案 | 依据 |
|------|---------|---------|---------|------|
| 抽象操作层物理位置 | `internal/ecma402/` 双层（根 + 4 子目录） | 单层平铺 `internal/ecma402/`；按字母 | 不分层 | formatjs 的 Bazel 子包验证可行；维护成本最低 |
| Init/Format 算法位置 | 子目录（`internal/ecma402/numberformat/`） | 公共 formatter 包（`numberformat/`） | 全部内联进公共包 | 公共包要走 Go 习惯 API；底层算法独立利于审计 |
| 函数命名 | spec 原名 verbatim | 全 Go 习惯化（`OptionValue` 等） | 混合 | 跨语言审计成本最优 |
| 选项校验入口 | 单一泛型 `GetOption[T]` | 拆为 `GetStringOption`/`GetBoolOption` | 反射 | Go 1.26 generics 已经够用，且贴近 spec 单一入口 |
| 内部槽存储 | struct 非导出字段 | sync.Map | WeakMap-like 反射间接表 | Go 没有 spec 强制的"间接私有"约束；`internal/` 已等价 |
| 错误类型映射 | 双 sentinel（ErrInvalidOption ≈ RangeError、ErrInvalidOptionType ≈ TypeError） | 单 sentinel | panic | 与 PHP ext/intl 双字段错误结构语义相同 |
| `MathematicalValue` 边界 | `internal/ecma402/types` 接口 | 直接用 `internal/decimal.Decimal` 具体类型 | 用 `*big.Float` 裸类型 | 隔离 SPEC 21 决议（Decimal 选型）；不绑定实现 |
| 选项归一化形式 | 抽象层接受 `map[string]any` | 抽象层接受 typed options struct | 反射 typed options | 与 formatjs 一致；公共层负责 functional options → map 转换 |

## 8. 代码块索引

| 位置 | 主题 | 类型 |
|------|------|------|
| §1.1 | formatjs `ecma402-abstract/` 目录布局 | 目录结构 |
| §1.5 | go-intl `internal/ecma402/` 推荐布局 | 目录结构 |
| §2.1 | formatjs `GetOption` 泛型签名 | TypeScript 签名 |
| §2.2 | go-intl `GetOption[T]` Go 签名 | Go 签名 |
| §2.2 | `GetOption` 调用示例 | Go 调用片段 |
| §3.1 | formatjs `setInternalSlot`/`getInternalSlot` 签名 | TypeScript 签名 |
| §3.3 | go-intl `NumberFormat` struct 内部槽形式 | Go 结构体片段 |
| §6.2 | `Part` / `Pattern` / `MathematicalValue` 接口 | Go 类型签名 |
| §6.2 | `GetNumberOption` / `GetStringOrBooleanOption` 签名 | Go 签名 |
| §6.2 | `Initialize` / `PartitionPattern` 子包签名 | Go 签名 |

## 9. 引用清单

### formatjs（主参考）

- `.references/formatjs/packages/ecma402-abstract/` — 整个抽象操作层 npm 子包
  - `GetOption.ts`、`GetNumberOption.ts`、`DefaultNumberOption.ts`、`GetStringOrBooleanOption.ts`、`GetOptionsObject.ts`、`CoerceOptionsToObject.ts`
  - `PartitionPattern.ts`
  - `ToIntlMathematicalValue.ts`
  - `CanonicalizeLocaleList.ts`、`SupportedLocales.ts`
  - `IsWellFormedCurrencyCode.ts`、`IsValidTimeZoneName.ts`、`CanonicalizeTimeZoneName.ts`
  - `IsSanctionedSimpleUnitIdentifier.ts`、`IsWellFormedUnitIdentifier.ts`
  - `utils.ts`（WeakMap 内部槽、`memoize` 等）
  - `constants.ts`、`data.ts`
  - 子目录：`NumberFormat/`、`DateTimeFormat/`、`PluralRules/`、`DisplayNames/`、`DurationFormat/`、`RelativeTimeFormat/`
  - `types/`（共享类型子包）
- `.references/formatjs/CLAUDE.md` 与 `knowledge-base/001a-bazel-toolchain.md`、`knowledge-base/002-ts-package-dependency-hierarchy.md` — 描述 `ecma402-abstract` 不发布 npm、3 个 Bazel 复合子包、Layer 1 在依赖图中的位置

### PHP ext/intl（scope 检查）

- `.references/ext/src/ecma402/util.h`、`util.cpp` — 数组去重、ASCII 工具
- `.references/ext/src/ecma402/error.{c,h}` — `ecma402_errorStatus` 双字段错误结构
- `.references/ext/src/ecma402/category.{c,h}` — 分类常量
- `.references/ext/src/ecma402/locale.{cpp,h}` — `ecma402_locale` struct + 全套 helpers

### translate-agent/intl（Go prior art）

- `.references/intl/intl.go` — 单 formatter（DateTimeFormat），无独立抽象操作层

### 内部文档

- `SPECS/00-vision-and-scope.md` §5.1 / §5.2 / §8 — 包布局、层级规则、开放问题
- `ANALYSIS.md` §1 — 调研方向（抽象操作层）
- `task.md` §r01 — 调研任务定义
