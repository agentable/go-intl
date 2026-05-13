---
task: R03
title: NumberFormat 算法、Decimal 选型与选项 API 研究
date: 2026-05-08
status: draft
authors: [research-agent]
references:
  - .references/formatjs/packages/intl-numberformat/
  - .references/formatjs/packages/ecma402-abstract/NumberFormat/
  - .references/formatjs/packages/bigdecimal/
  - .references/intl/intl.go (translate-agent/intl)
  - .references/ext/src/ecma402/currency.c
related-research:
  - R05-pluralrules.md (操作数契约对齐)
  - R04-datetimeformat.md (formatRange 通用化)
---

# R03 — NumberFormat 算法、Decimal 选型与选项 API

## Executive Summary

本报告回答 SPECS/00-vision-and-scope.md §8 中关于 `numberformat` 的三个未决问题，
并梳理与 `pluralrules` 的协同契约（compact-notation → PluralRules.Select 的输入）。
核心结论与置信度如下：

| 决策 | 推荐 | 置信度 | 关键依据 |
|------|------|------|---------|
| `internal/decimal` 底层类型 | `cockroachdb/apd/v3` | High | formatjs/bigdecimal 的 SpecialValue 与 GDA `Form` 一一映射；apd 提供 `Log10`/`Floor`/`Ceil`/`Quantize`/`Round`，覆盖 ToRawFixed / ToRawPrecision / ComputeExponent；并发安全 `Context` |
| `numberformat` 选项 API 风格 | Functional Options + 显式 set 标记 | Medium-High | Go 生态主流；`不传 = ECMA-402 默认`语义；零值不可与 `0` 默认值冲突的字段需显式标记 |
| 货币精度数据来源 | CLDR `currencyData.json` 派生（生成期 codegen） | High | formatjs CurrencyDigits 注入；PHP 经 ICU 由 CLDR 取；ISO 4217 不覆盖 CLDR 个别覆盖项 |
| formatRange 通用化 | NumberFormat 与 DateTimeFormat 各自实现 Collapse | High | `CollapseNumberRange` 操作 NumberFormatPart 枚举；DateTimeFormat 自有 collapse；不可共享 |
| 与 PluralRules 操作数契约 | Format-then-Select；传入 `roundedNumber × 10^exponent`；`c/e` 由 NumberFormat 计算并随 OperandsRecord 透传 | High | formatjs `format_to_parts.ts:262/304/316/331` |

## Background

ECMA-402 的 `Intl.NumberFormat` 在 ES 2025 已扩展到 `decimal/percent/currency/unit/compact/scientific/engineering`
七种 style/notation 组合，并加入 V3（rounding modes、roundingPriority、roundingIncrement、trailingZeroDisplay、useGrouping=auto/min2/always/false）。
项目目标是与 formatjs 输出**字节相等**，因此实现路径必须对齐 formatjs 的抽象算子序列与 BigDecimal 内部表示语义。

go-intl 仅有标准库 `math/big` 与 `golang.org/x/text` 作为基础设施，没有现成的 ECMA-402 数值表示：
JS `Number` ≈ IEEE-754 binary64，可隐式转换 BigInt；Go 端必须显式选定一个能承载
`NaN / +Inf / -Inf / Finite(coeff, exp)` 的 Decimal 类型，并在性能、并发、API 表面三方面与 formatjs 算法贴合。

## Cross-cut Topics

### 1. Decimal 表示选型

`@formatjs/bigdecimal` 的核心结构（参见 `.references/formatjs/packages/bigdecimal/src/index.ts`）：

```ts
class BigDecimal {
  readonly mantissa: bigint
  readonly exponent: number
  readonly specialValue?: 'NaN' | 'POSITIVE_INFINITY' | 'NEGATIVE_INFINITY'
  // value = mantissa × 10^exponent；special 优先
}
```

ECMA-402 的 `ToIntlMathematicalValue`、`ToRawPrecision`、`ToRawFixed`、`ComputeExponent` 都依赖：

1. NaN / ±∞ 的显式分支；
2. 任意精度十进制系数（mantissa）+ 十进制指数（exponent）；
3. `floor(log10(|x|))` 的精确十进制结果；
4. 半偶/半上/半下舍入到指定指数；
5. `Quantize` 到 `10^k`（用于 `roundingIncrement`）。

候选库横向比较（每项均交叉对照 formatjs/bigdecimal、ICU4C `decNumber`、translate-agent/intl 与 PHP ext/intl）：

| 维度 | `cockroachdb/apd/v3` | `shopspring/decimal` | `math/big.Float` | `math/big.Rat` |
|------|---------------------|----------------------|------------------|----------------|
| 内部表示 | `Negative + Coeff(big.Int) + Exponent + Form` | `value(big.Int) × 10^Exp` | binary mantissa + exp（IEEE 风格） | 分数 num/den |
| NaN / ±∞ | `Form` 枚举 = Finite/Infinite/NaN/NaNSignaling | 不支持，构造时 panic | 不支持有限/非有限的 IEEE 语义但十进制兼容差 | 不支持 |
| 规范遵循 | IEEE 754-2008 GDA（与 ICU `decNumber` 同源） | 自定 fixed-point | IEEE-754 binary | 数学纯有理数 |
| `Log10` / `Floor` / `Ceil` | 原生 `Log`/`Log10`/`Floor`/`Ceil` | 缺 `Log10`，需自实现 | 仅 binary log | 无 |
| `Quantize` | 原生 `Quantize` | 通过 `Truncate(places)` 近似 | 不直接 | 无 |
| 舍入模式 | `Rounder` 完整覆盖 ECMA-402 9 种 | 4 种 | binary 模式 | 无 |
| 并发 | `Context` 复用安全 | 全局可变默认 | `Float` 非并发安全 | 同 Float |
| 维护活跃度 | 持续维护、稳定 v3 | 维护，但 issue 反馈缓慢 | stdlib | stdlib |
| 0.1 + 0.2 字节相等 | 通过 | 通过 | **不通过**（binary） | 通过 |
| ECMA-402 `roundingIncrement = 5000` | `Quantize`+`Rounder` 覆盖 | 需要手写 | 需要手写 | 需要手写 |
| 与 formatjs/bigdecimal 表面相似度 | 高（mantissa/exponent/Form ↔ specialValue） | 中（无 Special） | 低 | 低 |

**推荐：`cockroachdb/apd/v3`**（High）

`internal/decimal` 包仅暴露 ECMA-402 抽象操作所需的窄接口，apd 作为后端实现。该包向上不暴露 apd 类型，便于将来切换。
不引入 `shopspring/decimal`、`big.Float`、`big.Rat` 作为公共依赖，理由：

- `shopspring/decimal` 缺 NaN/±∞，不能承载 `ToIntlMathematicalValue("NaN")` 的输入；构造期 panic 与项目"无 panic"红线冲突。
- `math/big.Float` 是二进制浮点，0.1+0.2 在转十进制串时会偏离 formatjs 输出，破坏字节相等。
- `math/big.Rat` 缺乏 `Log10`、`Floor` 与定向舍入，用其完成 ToRawPrecision 等价物代价过高。

风险与落点：

- apd 的 `Log10` 用 `Context.Precision` 控制内部精度；ComputeExponent 必须把内部精度设到足以覆盖 ECMA-402 `mxfd` 上界（ES 2025 = 100），实测 `Precision = 200` 安全。
- `Context` 是值类型，可按 goroutine/Format 调用复制；NumberFormat 实例持一个不可变 baseline `Context`，实际格式化时复制后修改 `Rounding`。

### 2. NumberFormat 选项管线

formatjs 在 `.references/formatjs/packages/ecma402-abstract/NumberFormat/InitializeNumberFormat.ts` 给出权威序列：

```text
opts := CoerceOptionsToObject(options)
1. localeMatcher        := GetOption(opts,"localeMatcher",string,["lookup","best fit"],"best fit")
2. numberingSystem      := GetOption(opts,"numberingSystem",string,nil,nil)  // 校验 type=numbers
3. resolvedLocale       := ResolveLocale(availableLocales,requestedLocales,localeMatcher,relevantExt={"nu"},...)
4. SetNumberFormatUnitOptions(nf, opts)            // style + currency/{...}/unit/{...}
5. notation             := GetOption(...,"standard|scientific|engineering|compact","standard")
6. mnfdDefault, mxfdDefault 由 style 决定：
    decimal       => 0,3
    percent       => 0,0
    currency      => CurrencyDigits(currency, currencyData), idem
    unit          => 0,3
   compact 与 scientific/engineering 在 SetNumberFormatDigitOptions 中再调整
7. SetNumberFormatDigitOptions(nf, opts, mnfdDefault, mxfdDefault, notation)
   - roundingPriority = auto | morePrecision | lessPrecision
   - roundingIncrement ∈ VALID_ROUNDING_INCREMENTS (1,2,5,10,20,25,50,100,200,250,500,1000,2000,2500,5000)
   - roundingMode      ∈ ceil|floor|expand|trunc|halfCeil|halfFloor|halfExpand|halfTrunc|halfEven
   - trailingZeroDisplay ∈ auto|stripIfInteger
8. compactDisplay       := GetOption(...) 仅 notation=compact 生效
9. useGrouping          := normalize 至 always|auto|min2|false
10. signDisplay         := auto|always|exceptZero|negative|never
```

`SetNumberFormatDigitOptions`（`.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatDigitOptions.ts`）的 5 路分支：

```text
hasSd = mnsd|mxsd 至少一个被设置
hasFd = mnfd|mxfd 至少一个被设置
need  = hasSd || hasFd
priority = roundingPriority

case priority='morePrecision' or 'lessPrecision': roundingType=morePrecision/lessPrecision
case hasSd                                       : roundingType=significantDigits
case hasFd                                       : roundingType=fractionDigits
case notation='compact'                          : roundingType=compactRounding
otherwise                                        : roundingType=fractionDigits with mnfd=mnfdDefault, mxfd=mxfdDefault
roundingIncrement!=1 ⇒ roundingType 强制 fractionDigits 且 mnfd=mxfd
```

go-intl 在 `numberformat/options.go` 应严格按上述分支实现；`Resolve` 方法返回的 `ResolvedOptions` 与 formatjs `resolvedOptions()` 字段对齐：

```go
type ResolvedOptions struct {
    Locale                       locale.Locale
    NumberingSystem              string
    Style                        Style          // decimal|percent|currency|unit
    Currency                     string         // style=currency
    CurrencyDisplay              CurrencyDisplay
    CurrencySign                 CurrencySign
    Unit                         string
    UnitDisplay                  UnitDisplay
    MinimumIntegerDigits         int
    MinimumFractionDigits        int            // 仅在 roundingType≠significantDigits 时呈现
    MaximumFractionDigits        int
    MinimumSignificantDigits     int            // 仅在 roundingType≠fractionDigits 时呈现
    MaximumSignificantDigits     int
    UseGrouping                  UseGrouping    // always|auto|min2|false
    Notation                     Notation       // standard|scientific|engineering|compact
    CompactDisplay               CompactDisplay // short|long
    SignDisplay                  SignDisplay
    RoundingIncrement            int
    RoundingMode                 RoundingMode
    RoundingPriority             RoundingPriority
    TrailingZeroDisplay          TrailingZeroDisplay
}
```

`PartitionNumberPattern` 的格式化主流程（`.references/formatjs/packages/ecma402-abstract/NumberFormat/PartitionNumberPattern.ts` + `format_to_parts.ts`）：

```text
1. x := ToIntlMathematicalValue(value)
2. exponent := 0
3. if notation=scientific|engineering|compact:
       exponent := ComputeExponent(nf, x, localeData)
       x        := x ÷ 10^exponent
4. n := FormatNumericToString(nf, x)            // RawString + RoundedNumber
5. if x.specialValue:                           // NaN / ±Inf 走 symbol map
       formattedString := getNaN(localeData) or getInfinity(localeData)
   else:
       formattedString := PartitionDigitParts(nf, n, exponent, getNumberingSystem(nf), localeData)
6. 包装 sign / currency / unit / percent / compact-suffix / exponent-symbol
```

### 3. 选项 API 风格选型

候选三种风格，三方对照（formatjs config 对象、translate-agent/intl `type Options struct`、Go 生态 functional options）：

| 维度 | Functional Options | Config Struct | Builder（链式） |
|------|--------------------|----------------|----------------|
| Go 习惯度 | 高（grpc/zap/cobra/sql 主流） | 中（pgx、translate-agent/intl 用） | 低（少见，xorm 等） |
| "未传 = ECMA-402 默认" 语义表达 | 显式（option 内部用 `*T` 或 set 位） | 需要单独的 `*T` 或 `Has*` 字段 | 与 functional 等价 |
| 与 `ResolvedOptions` 解耦 | 解耦 | 易把"输入"与"解析后"混淆 | 解耦 |
| 校验时机 | 构造期一次性聚合 | 构造期 | 链式过程难以聚合 |
| 与 messageformat-go 互操作 | option fn 难以从 `map[string]any` 来回转换 | struct + json/mapstructure 直转 | 同 builder |
| 演进 | 加新 Option fn 不破坏 ABI | 加新字段非 zero 起步即可 | 加新方法 |
| 文档/示例可读性 | 高（"option 即句子"） | 中 | 中 |

go-intl 的 ECMA-402 字段中存在几类**零值与默认值不同**的字段：`MinimumFractionDigits=0` 与"未设置"语义不同；
`UseGrouping` 的 zero-value `""` 必须解析为 ECMA-402 默认 `auto`，而 `false` 是合法用户选择。
两种风格都能解决：

- functional：`WithMinimumFractionDigits(int)` 内部置 `set` flag。
- struct：每个会出现"零值合法"的字段用 `*int` / `Optional[T]`。

**推荐：Functional Options**（Medium-High）。理由：

1. 与 `intl.FormatNumber(loc, v, opts ...Option)` 的一行式 façade 同构（SPECS 00 §6.2 Progressive Disclosure）。
2. 校验在构造函数内集中执行；reify 后的内部 struct 不外暴。
3. 对 messageformat-go 的 `map[string]any` 桥接放在 `numberformat.OptionsFromMap` 适配器，不污染主 API。
4. translate-agent/intl 选 struct 的根因是 DateTimeFormat 的字段几乎都有"未设置/具体枚举值"两态枚举（`Year`/`Month` 等），用 byte iota + 0=undefined 即可天然区分；NumberFormat 字段语义更杂（int 默认值非零、bool 三态），functional options 的显式 set 表达更自然。

`numberformat` 公开形态（仅签名，不含实现）：

```go
type Option func(*config)

func WithCurrency(code string) Option
func WithCurrencyDisplay(d CurrencyDisplay) Option
func WithMinimumFractionDigits(n int) Option
func WithMaximumFractionDigits(n int) Option
func WithRoundingPriority(p RoundingPriority) Option
func WithRoundingIncrement(inc int) Option
func WithRoundingMode(m RoundingMode) Option
func WithNotation(n Notation) Option
func WithCompactDisplay(d CompactDisplay) Option
func WithUseGrouping(g UseGrouping) Option
func WithSignDisplay(s SignDisplay) Option
func WithTrailingZeroDisplay(t TrailingZeroDisplay) Option

type Formatter struct{ /* 不可变；包含 resolved + apd.Context 副本 */ }

func New(loc locale.Locale, opts ...Option) (*Formatter, error)
func (f *Formatter) Format(v any) string
func (f *Formatter) FormatToParts(v any) []Part
func (f *Formatter) FormatRange(a, b any) string
func (f *Formatter) FormatRangeToParts(a, b any) []Part
func (f *Formatter) ResolvedOptions() ResolvedOptions
```

调用示例（3–5 行，不含实现）：

```go
nf, err := numberformat.New(locale.MustParse("zh-Hant-TW"),
    numberformat.WithNotation(numberformat.Compact),
    numberformat.WithCompactDisplay(numberformat.CompactShort))
// nf.Format(98765) == "9.9萬"
```

### 4. 货币精度来源

`.references/formatjs/packages/ecma402-abstract/NumberFormat/CurrencyDigits.ts` 接受外部注入的
`currencyDigitsData: Record<string, number>`，数据由 CLDR `supplemental/currencyData.json` 在生成期注入。
`.references/ext/src/ecma402/currency.c` 经 `ucurr_openISOCurrencies` → ICU → CLDR；不内嵌 ISO 4217。
ECMA-402 §15.5 显式要求"locale-aware default"，CLDR 是唯一覆盖率全且与 ICU/formatjs 对齐的数据源。

**推荐**：`internal/cldr/currency` 在代码生成阶段从 CLDR `currencyData.fractions` 抽取 `(code → defaultFractionDigits, cashDigits, rounding, cashRounding)`，
`numberformat` 读取 `defaultFractionDigits`（或 `cashDigits` 当未来扩展 cash 模式）。不引入 ISO 4217 静态表。

### 5. formatRange 与 CollapseNumberRange

`.references/formatjs/packages/ecma402-abstract/NumberFormat/CollapseNumberRange.ts` 直接消费
`NumberFormatPart{type:"unit"|"currency"|"percentSign"|"exponentMinusSign"|"exponentInteger"|...}` 枚举，
DateTimeFormat 的 part 类型完全不同（`era|year|month|...|relatedYear`），
两者的 collapse 算法虽然结构同构（两端 part 序列里去除可合并的固定前后缀），但工作在各自的 part 域上。
formatjs 也是各自实现：`packages/ecma402-abstract/NumberFormat/CollapseNumberRange.ts` 与
`packages/ecma402-abstract/DateTimeFormat/PartitionDateTimeRangePattern.ts`。

**推荐**：go-intl 不抽象通用 `CollapseRange`；`numberformat/collapse.go` 与 `datetimeformat/collapse.go` 各自实现，
共享内容仅限于 `internal/ecma402` 中两个文件都用到的 `RangeKind = startRange|shared|endRange|approximateSign` 字符串常量，
以避免拼写漂移（与 formatjs 一致）。

### 6. 与 PluralRules 的操作数契约（与 R05 对齐）

formatjs `format_to_parts.ts:262` / `:304` / `:316` / `:331`：

```text
// compact / scientific / engineering 路径
selectPlural(pl, numberResult.roundedNumber.times(getPowerOf10(exponent)).toNumber(), { type: 'cardinal' })
```

注意：传给 `pl.select()` 的是**回放到原始量级**的 `Number`（即 `roundedNumber × 10^exponent`），
**不是**已经除以 `10^exponent` 后的"显示数"。`PluralRules` 内部会再调用 `ResolvePlural`，
其中 `GetOperands` 拿到的 `formattedString` 由 `PartitionNumberPattern` 重新格式化，
`CompactExponent` 与 `e` 操作数则由 `ComputeExponentForMagnitude` 在 PluralRules 一侧独立计算。

go-intl 的对应契约（在 `numberformat.partitionPattern` 内部，对应 R05 §3）：

```go
// 伪代码：仅契约示意
sel := f.plural.SelectFormatted(roundedDecimal, exponent)  // sel ∈ {zero,one,two,few,many,other}
// PluralRules 内部用 (roundedDecimal, exponent) 重建 OperandsRecord 并调用编译后的 plural fn
```

`pluralrules` 不应假设 NumberFormat 已经把数值除回去；`numberformat` 也不应自己取操作数。
`internal/ecma402.OperandsRecord` 是两包共享的 SSOT 类型，定义见 R05 §2。

## Project Landings

### formatjs / `@formatjs/intl-numberformat`

`.references/formatjs/packages/intl-numberformat/src/get_internal_slots.ts` + `core.ts`：
NumberFormat 持 `localeData` + `availableLocales` 的闭包；`format`/`formatToParts`/`formatRange`/`formatRangeToParts`/`resolvedOptions` 都直接命中 `internalSlot`。
go-intl 对应：`Formatter` 不可变，所有 `localeData` 由 `internal/cldr` accessor 在 `New` 时拉取并冻结到 `Formatter` 内部。

### translate-agent/intl

`.references/intl/intl.go`：仅 DateTimeFormat。NumberFormat 没有先例，所以选项 API 风格没有 Go 圈内同领域参照。
但其 `type Options struct{Era, Year, Month, Day}` 的"byte iota = 0 表示 undefined"模式在 NumberFormat 不适用，
因为 NumberFormat 的字段域是 `int`/`bool` 等存在合法零值的类型——这是不选 struct API 的关键论据。

### PHP / ext/intl

`.references/ext/src/ecma402/currency.c` 把 ISO 4217 校验完全交给 ICU；其经 `ucurr_isAvailable` 验证、
`ucurr_getDefaultFractionDigits` 取精度。等价于 go-intl 用 CLDR 派生的 codegen 表。

## Decision Matrix

| 决策点 | 推荐 | 备选 | 拒绝 |
|--------|------|------|------|
| Decimal 后端 | `cockroachdb/apd/v3`（GDA / Form / Log10） | `math/big.Float`（仅作 IEEE 转换辅助） | `shopspring/decimal`（无 NaN/Inf，construct panic）；`math/big.Rat`（无 Log10/定向舍入） |
| 选项 API | Functional Options + 显式 set | Config struct + `*T` 字段 | Builder（与 façade 不一致） |
| 货币精度 | CLDR `currencyData.fractions` codegen | ICU runtime（不引入 ICU） | ISO 4217 静态表（与 CLDR/ICU/formatjs 不一致） |
| Range collapse | NumberFormat / DateTimeFormat 各自实现 | 抽象 `CollapseRange[T]` 泛型 | 跨包共享同一函数（part 域不同） |
| Plural 联动 | NumberFormat 把 `(roundedDecimal, exponent)` 给 PluralRules | NumberFormat 自取操作数 | NumberFormat 提前除 10^exp 再传 |

## Code Block Index

| Block | 来源 | 说明 |
|-------|------|------|
| `BigDecimal{mantissa,exponent,specialValue}` | formatjs/bigdecimal/src/index.ts | TS 端权威表示 |
| `InitializeNumberFormat` 步骤 | formatjs/ecma402-abstract/.../InitializeNumberFormat.ts | 选项管线 |
| `SetNumberFormatDigitOptions` 5 路分支 | 同包 | priority/sd/fd/compact 决策表 |
| `PartitionNumberPattern` 主流程 | 同包 + format_to_parts.ts | 格式化引擎 |
| `ResolvedOptions` 字段表 | 本报告 | go-intl 对齐 formatjs `resolvedOptions()` |
| Functional Options 签名 | 本报告 | 推荐 API |
| `selectPlural(...roundedNumber.times(getPowerOf10(exponent)).toNumber(),...)` | format_to_parts.ts:262 etc. | 操作数契约（与 R05 一致） |

## Citations

1. `.references/formatjs/packages/bigdecimal/src/index.ts` — `mantissa/exponent/specialValue` 表示
2. `.references/formatjs/packages/ecma402-abstract/NumberFormat/InitializeNumberFormat.ts` — 选项管线步骤
3. `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatDigitOptions.ts` — 5 路分支与 `VALID_ROUNDING_INCREMENTS`
4. `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatUnitOptions.ts` — currency/unit 校验
5. `.references/formatjs/packages/ecma402-abstract/NumberFormat/PartitionNumberPattern.ts` — 主流程
6. `.references/formatjs/packages/ecma402-abstract/NumberFormat/format_to_parts.ts:262` — Compact 路径将 `roundedNumber × 10^exponent` 喂给 PluralRules
7. `.references/formatjs/packages/ecma402-abstract/NumberFormat/CollapseNumberRange.ts` — Number-specific collapse
8. `.references/formatjs/packages/ecma402-abstract/NumberFormat/CurrencyDigits.ts` — 注入式 currencyData
9. `.references/formatjs/packages/intl-numberformat/tests/notation-compact-zh-TW.test.ts` — `format(98765)='9.9萬'`
10. `.references/intl/intl.go` — translate-agent/intl 选项风格（DateTimeFormat 范例）
11. `.references/ext/src/ecma402/currency.c` — PHP 经 ICU/CLDR 取货币精度
12. `github.com/cockroachdb/apd/v3` — GDA、Form、Context、Log10、Quantize、Rounder 文档
15. `github.com/shopspring/decimal` — 缺 NaN/Inf 与 Log10
16. `pkg.go.dev/math/big` — Float/Rat 的语义边界
17. SPECS/00-vision-and-scope.md §8 — Open Questions
