---
task: R05
title: PluralRules 编译策略、x/text 复用与 selectRange 研究
date: 2026-05-08
status: draft
authors: [research-agent]
references:
  - .references/formatjs/packages/intl-pluralrules/
  - .references/formatjs/packages/ecma402-abstract/PluralRules/
  - golang.org/x/text/feature/plural
  - .references/intl/intl.go (translate-agent/intl)
  - .references/ext/src/ecma402/plural_rules.c
related-research:
  - R03-numberformat.md (操作数契约对齐)
---

# R05 — PluralRules 编译策略、x/text 复用与 selectRange

## Executive Summary

本报告回答 SPECS/00-vision-and-scope.md §8 中关于 `pluralrules` 的三类未决问题：
（1）CLDR plural DSL 的 Go 端落地策略——codegen 还是 runtime interpreter；
（2）`golang.org/x/text/feature/plural` 是否可复用；
（3）`selectRange` 与 compact-notation 操作数的端到端数据流。

| 决策 | 推荐 | 置信度 | 关键依据 |
|------|------|------|---------|
| CLDR plural DSL 落地 | **生成期 codegen 到 Go 源码** | High | formatjs intl-pluralrules 走 TS-AST→JS 字符串；与 SPECS 00 §5.3 "embed-only / 无运行时 I/O" 一致；translate-agent/intl 已用同模式生成 CLDR 数据 |
| 复用 `golang.org/x/text/feature/plural` | **不复用** | High | CLDR 32（当前 47+）；缺 c/e（compact/scientific）操作数；包标注 "UNDER CONSTRUCTION"；不能承载 ECMA-402 v3 的 PluralRules |
| selectRange 数据来源 | CLDR `pluralRanges.xml` 与 `pluralRules.xml` 在生成期同流入 `internal/cldr/plural` | High | formatjs `PluralRuleSelectRange` 用 `${start}_${end}` 表查；缺 range 表则回落 end 类别 |
| 与 NumberFormat 操作数契约 | Format-then-Select；`(roundedDecimal, exponent)` 是 PluralRules 输入；`c/e` 由 PluralRules 一侧通过 `ComputeExponentForMagnitude` 产出 | High | formatjs `format_to_parts.ts:262`；V8 `js-plural-rules.cc:222-226` |
| `Select` 接口形态 | `Select(decimal.Decimal) Category` + `SelectRange(start, end decimal.Decimal) Category` + `SelectFormatted(roundedDecimal, exponent)`（包内可见） | Medium-High | 暴露 ECMA-402 公共 API；compact 联动用包内可见入口 |

## Background

ECMA-402 `Intl.PluralRules` 定义于 ES 2017，2024 年扩展 `selectRange`。CLDR `plurals.xml` 用一阶布尔代数 DSL
描述 `zero / one / two / few / many / other` 六类规则；`pluralRanges.xml` 单独列出
"start 类别 + end 类别 → 范围类别" 的映射表。Plural 规则不仅服务 `Intl.PluralRules`，
也在 `Intl.NumberFormat` 的 compact-notation 后缀选择上被调用——这意味着 PluralRules 的输入契约
直接影响 NumberFormat 字节相等。

go-intl 的限制（来自 SPECS 00）：
（a）数据走 `embed`，运行时不做文件 I/O 与 JSON 解析；
（b）不引入 ICU 等 native 依赖；
（c）输出与 formatjs 字节相等；
（d）`ResolveLocale` 路径与其他 Intl 包共享 `internal/ecma402` 一份。

## Cross-cut Topics

### 1. CLDR plural DSL：codegen vs interpreter vs 复用 x/text

CLDR plural DSL 的形式（节选 `pluralRules.xml` `en` cardinal）：

```text
one:    i = 1 and v = 0
other:  (everything else)
```

操作数集合（参见 `.references/formatjs/packages/ecma402-abstract/PluralRules/GetOperands.ts`）：

```ts
interface OperandsRecord {
  Number:                            number  // |x|
  IntegerDigits:                     number  // i
  NumberOfFractionDigits:            number  // v
  NumberOfFractionDigitsWithoutTrailing: number // w
  FractionDigits:                    number  // f
  FractionDigitsWithoutTrailing:     number  // t
  CompactExponent:                   number  // e (== c)
}
```

候选三策略：

| 维度 | (A) 生成期 codegen 到 Go | (B) 运行时 interpreter（解析 DSL 树） | (C) 复用 `x/text/feature/plural` |
|------|-------------------------|-------------------------------------|--------------------------------|
| CLDR 版本对齐 | 与生成器一致（每次 release 重生成） | 同 (A) | 锁死 CLDR 32（2017-09），无升级路径 |
| 运行时成本 | 一次函数调用 | DSL 解析 + tree-walk | 包内查表（无 c/e） |
| 包尺寸 | 中等（per-locale 函数） | 小（一份 interpreter + per-locale 数据） | 已在依赖 |
| 与 SPECS 00 §5.3 "embed-only / 无 I/O" 一致 | 一致（数据即 Go 源） | 一致（数据嵌入） | 一致 |
| compact 操作数 c/e | ✅ codegen 时可生成 `func(num, isOrdinal, exponent)` | ✅ 运行时把 e 注入操作数 | ❌ 无该输入 |
| 与 formatjs 字节相等 | ✅（移植 plural-rules-compiler） | ✅ 算法层面 | ❌（CLDR 32 已与 47 不同；en-Cyrl-AT 等新增 locale 缺） |
| 翻新成本 | TypeScript-AST 思路移植到 Go AST（go/ast）；中等 | 写一个稳健的 DSL parser；中等 | 0 |
| 维护负担 | 每次 CLDR release 跑生成器 | 每次 CLDR release 替换数据 | 受 x/text 升级节奏限制 |
| 与 translate-agent/intl 模式 | 一致（已用 codegen 嵌入 CLDR） | 不一致 | 不一致 |

**推荐：策略 A（codegen 到 Go 源）**（High）。

理由：

1. formatjs `.references/formatjs/packages/intl-pluralrules/scripts/plural-rules-compiler.ts` 在生成期将
   CLDR 规则用 TypeScript-AST 编译为字符串化 JS 函数。Go 端可对应使用 `go/ast` + `go/format`
   生成 `func(o OperandsRecord) Category` 闭包，逐 locale 一个 func，挂在
   `var rules = map[language.Tag]func(OperandsRecord) Category{...}` 上。
2. 只生成实际用到的操作数表达式（formatjs `should-emit-*.ts` 系列做这件事）；既保证 codegen 输出小，
   又使每个 locale 的判定路径最短。
3. 与 SPECS 00 §5.3 的 "embed-only / 无 I/O" 与 "数据生成一次"原则同构。
4. 与 R03 推荐的 NumberFormat codegen（locale 数据派生）共用一条 `tools/gen-cldr` 流水线。
5. 不复用 `x/text/feature/plural` 详见 §1.1。

### 1.1 为什么 `golang.org/x/text/feature/plural` 不可用

直接证据：

- 包 doc：`Package plural provides utilities for handling linguistic plurals in text. This package is UNDER CONSTRUCTION ...`
- CLDR 数据基线：`golang.org/x/text/internal/cldr` 内置数据来自 CLDR 32（2017-09）。当前 CLDR 47（2025）。
  CLDR 33+ 重写了 Welsh/Cymraeg、Hebrew、Polish 等若干语言的复数规则；CLDR 41+ 加了 Russian
  ordinal 修正。停在 32 即输出与 formatjs 不一致。
- API 缺口：`plural.Select(t language.Tag, scale int, digits string)` 仅接受 `(scale, digits)` 两参，
  没有 compact-notation 的 `c/e` 操作数。`Intl.NumberFormat` notation=compact 在 `getCompactDisplay`
  环节必须查 `pluralCategory(amount, exponent)`，无 `e` 操作数则不能区分 `1 thousand` 与 `1K` 在
  少数语言（如波兰语、捷克语）下的复数类别。

结论：x/text/feature/plural 在我们目标年（2026）下既不准确也不完整，不能作为依赖。

### 2. 操作数与 Plural 函数签名（与 R03 对齐）

`internal/ecma402.OperandsRecord`（建议放在 `internal/ecma402` 而非 `pluralrules` 包，理由是
`numberformat` 的 `format_to_parts.go` 与 `pluralrules` 都消费它）：

```go
type OperandsRecord struct {
    Number                              decimal.Decimal // |x| 原始量级
    IntegerDigits                       int             // i
    NumberOfFractionDigits              int             // v
    NumberOfFractionDigitsWithoutTrailing int           // w
    FractionDigits                      uint64          // f
    FractionDigitsWithoutTrailing       uint64          // t
    CompactExponent                     int             // c == e
}
```

生成器输出（一个 locale 一个 `func`，签名固定）：

```go
// generated_plurals.go (片段；由 tools/gen-cldr/plural 生成)
func plural_en_cardinal(o internalecma402.OperandsRecord) Category {
    if o.IntegerDigits == 1 && o.NumberOfFractionDigits == 0 {
        return One
    }
    return Other
}
```

`pluralrules` 公开 API：

```go
type Type uint8       // Cardinal | Ordinal
type Category uint8   // Zero | One | Two | Few | Many | Other

type Option func(*config)
func WithType(t Type) Option
func WithMinimumIntegerDigits(n int) Option
func WithMinimumFractionDigits(n int) Option
func WithMaximumFractionDigits(n int) Option
func WithMinimumSignificantDigits(n int) Option
func WithMaximumSignificantDigits(n int) Option
func WithRoundingPriority(p numberformat.RoundingPriority) Option
func WithRoundingMode(m numberformat.RoundingMode) Option

type Rules struct{ /* 不可变；包含 resolved + 编译后的 plural func */ }

func New(loc locale.Locale, opts ...Option) (*Rules, error)
func (r *Rules) Select(v any) Category
func (r *Rules) SelectRange(start, end any) Category
func (r *Rules) ResolvedOptions() ResolvedOptions

// 包内可见，供 numberformat compact 路径调用
// 入参为 NumberFormat 内部已经 ToRawPrecision/ToRawFixed 后的 decimal + 计算好的 exponent
func (r *Rules) selectFormatted(rounded decimal.Decimal, exponent int) Category
```

### 3. selectRange 数据流

formatjs `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePluralRange.ts` +
`.references/formatjs/packages/intl-pluralrules/index.ts` 中 `PluralRuleSelectRange`：

```text
xp := PluralRuleSelect(localeData, type, x, isOrdinal)
yp := PluralRuleSelect(localeData, type, y, isOrdinal)
// 1. 已格式化字符串相等 → 返回 xp（避免 "1–1" 走 range 表的边角）
if FormatNumericToString(x) == FormatNumericToString(y) { return xp }
// 2. 若 locale 没有 pluralRanges 数据 → 回落到 yp（end-class 兜底）
if !pluralRanges[localeData] { return yp }
// 3. 查 pluralRanges["${xp}_${yp}"] 或 pluralRanges["${xp}_other"] 等
return pluralRanges[localeData][`${xp}_${yp}`] ?? yp
```

Go 端落地：

- `tools/gen-cldr/plural` 同时读取 CLDR `plurals.xml`（→ rules）与 `pluralRanges.xml`（→ ranges），
  emit 到 `internal/cldr/plural/generated_ranges.go`：

```go
// generated_ranges.go (片段)
var pluralRanges_en = map[rangeKey]Category{
    {Start: One, End: Other}: Other,
    // ...
}
```

- `pluralrules.SelectRange` 流程严格按 formatjs 的 3 步算法（先 stringify-eq 短路、再缺数据回落、再表查）。
- `internal/ecma402.PluralRuleSelectRange` 公共算子供 R04（DateTimeFormat range 类别选择）共用——
  虽然 DateTimeFormat 不直接调 PluralRules，但 pluralRanges 表的 `(start, end) → category` 形态
  是只在 plural 用，故仍只在 `pluralrules` 内可见。

### 4. 与 NumberFormat 协同：Format-then-Select 契约

`format_to_parts.ts:262/304/316/331` 关键行：

```ts
selectPlural(pl, numberResult.roundedNumber.times(getPowerOf10(exponent)).toNumber(),
             { type: 'cardinal' })
```

`ResolvePlural`（`.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePlural.ts`）：

```text
n           := abs(num)
formatted   := FormatNumericToString(internalNumberFormat, n)
if notation==compact:
    e := ComputeExponentForMagnitude(numberFormat, log10(n))   // PluralRules 自己算 e
operands    := GetOperands(formatted.formattedString, e)
return PluralRuleSelect(localeData, type, operands, isOrdinal)
```


ICU 端把 `FormattedNumber`（含 minor unit、小数位、舍入信息）整体传入 `select`，
内部抽 `OperandsRecord` 后选规则。两实现语义等价。

go-intl 在 `numberformat.partitionPattern` 中，对 compact/scientific/engineering 路径：

```go
// 合同：rounded 已经是 ToRawPrecision/ToRawFixed 之后的十进制；
// exponent 是 ComputeExponent 决定的 compact/scientific 指数；
// 不要预先除以 10^exponent 后再交给 plural；plural 要看到原始量级。
cat := f.plural.selectFormatted(rounded, exponent)
```

`pluralrules.selectFormatted` 内部用 `(rounded, exponent)` 重建 `OperandsRecord`：
`Number = rounded × 10^exponent`、`CompactExponent = exponent`、其余从 `rounded` 的小数串抽取。

### 5. 操作数计算：在 Decimal 上运行 vs 字符串运行

formatjs `GetOperands` 接受 `formattedString: string`。这是 JS 实现的妥协（BigDecimal 与 number 互操作复杂）。
Go 端因为 R03 推荐 `cockroachdb/apd/v3`，`Decimal` 拥有 `Coeff(big.Int)` + `Exponent`，
可以**在 Decimal 上直接抽操作数**（避免一次 string→int 解析），并在 trailing-zero 处理上更可靠：

| 操作数 | 从字符串（formatjs） | 从 apd.Decimal |
|--------|---------------------|---------------|
| `i` | 取整数部分位数 | `Coeff` 与 `Exponent` 推导（`numDigits(Coeff) + Exponent` 截到非负） |
| `v` | 小数位数（含 trailing 0） | `-Exponent` 取下界（trailing 由"原始格式化串"里的 trailing 决定） |
| `w` | 去 trailing 0 后小数位 | `Coeff.Trailing0` 计数 |
| `f` | 小数串原值 | `Coeff mod 10^v` |
| `t` | 去 trailing 0 后小数值 | `Coeff / 10^trailing0 mod 10^w` |
| `e/c` | `numberFormat` 计算后注入 | 同 |

**推荐**：`internal/ecma402.GetOperands` 对外签名 `func(rounded decimal.Decimal, formatted string, exponent int) OperandsRecord`，
其中 `formatted` 仅用于跟 formatjs 对齐 trailing-zero 行为（formatjs 的 trailing 由 `FormatNumericToString` 输出决定，
不仅看数学值；e.g. `mnfd=2` 时 `1` 显示成 `1.00`，`v=2`、`w=0`），其余从 Decimal 抽。

### 6. 内部缓存与复用 ResolveLocale

formatjs `intl-pluralrules` 不缓存编译后的规则；规则是 import-time 静态。
我们用 codegen 后所有 locale 的 plural fn 都是包级 `var`，天然复用，无须 cache。
`ResolveLocale` 走 `internal/ecma402.ResolveLocale`（与 numberformat 同源），
按 `[best-fit|lookup]` 在 `availablePluralLocales`（codegen 产出）里匹配。

## Project Landings

### formatjs / `@formatjs/intl-pluralrules`

`packages/intl-pluralrules/scripts/plural-rules-compiler.ts`：TS-AST → JS 函数串。
`packages/intl-pluralrules/index.ts` 的 `PluralRuleSelect/PluralRuleSelectRange`。
go-intl 把"AST 编译器"用 `go/ast` 重写，输出到 `internal/cldr/plural/generated_*.go`。

### Node / V8

`deps/v8/src/objects/js-plural-rules.cc`：完全委托 ICU。`select(FormattedNumber)` 是 ECMA-402
"Format-then-Select" 契约的另一存在性证明。go-intl 不走 ICU 但语义对齐。

### translate-agent/intl

`.references/intl/intl.go` 没实现 PluralRules，但其 CLDR codegen 思路（`make` + Go 源生成）
可直接复用为 `tools/gen-cldr/plural` 的脚手架。

### `golang.org/x/text/feature/plural`

仅作为反例引用：标注 UNDER CONSTRUCTION、CLDR 32、缺 c/e。我们不能依赖。

### PHP / ext/intl

`.references/ext/src/ecma402/plural_rules.c` 用 ICU `PluralRules::forLocale`。
`selectRange` 在 ICU 64+ 提供。go-intl 自实现 selectRange，等价语义。

## Decision Matrix

| 决策点 | 推荐 | 备选 | 拒绝 |
|--------|------|------|------|
| Plural DSL 落地 | 生成期 codegen（go/ast 端口 plural-rules-compiler.ts） | 运行时 interpreter | 复用 x/text/feature/plural |
| 数据来源 | CLDR `plurals.xml` + `pluralRanges.xml` 同流水线 | 仅 plurals，pluralRanges 单独 | 引入 ICU |
| 操作数提取 | 从 `apd.Decimal` 直接抽，再用 `formatted` 串校准 trailing-zero | 全部从字符串抽（formatjs 路径） | 全部从 float 抽（精度不足） |
| selectRange 缺数据兜底 | 回落 end-class（与 formatjs 一致） | 报错 | 回落 start-class |
| 与 NumberFormat 接口 | 包内可见 `selectFormatted(rounded, exponent)` | 让 NumberFormat 自取 OperandsRecord | NumberFormat 预先除 10^exp |

## Code Block Index

| Block | 来源 | 说明 |
|-------|------|------|
| `OperandsRecord` 字段 | formatjs/PluralRules/GetOperands.ts | 操作数集合 |
| `plural_en_cardinal` 生成示例 | 本报告 | codegen 输出形态 |
| `pluralrules` 公开 API 签名 | 本报告 | go-intl 对齐 |
| `pluralRanges_en` map 形态 | 本报告 | range 表 codegen |
| formatjs `selectPlural(...roundedNumber.times(getPowerOf10(exponent)).toNumber(),...)` | format_to_parts.ts:262 | Format-then-Select 契约（与 R03 一致） |
| V8 `formatted_number = fmt->formatDouble(...); icu_plural_rules->select(formatted_number,...)` | js-plural-rules.cc:222-226 | ICU 同语义 |
| `ResolvePluralRange` 三步 | formatjs/PluralRules/ResolvePluralRange.ts | selectRange 算法 |

## Citations

1. `.references/formatjs/packages/intl-pluralrules/scripts/plural-rules-compiler.ts` — TS-AST 编译器
2. `.references/formatjs/packages/intl-pluralrules/index.ts` — `PluralRuleSelect`、`PluralRuleSelectRange`
3. `.references/formatjs/packages/ecma402-abstract/PluralRules/GetOperands.ts` — `OperandsRecord`
4. `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePlural.ts` — Format-then-GetOperands
5. `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePluralRange.ts` — selectRange 三步
6. `.references/formatjs/packages/ecma402-abstract/NumberFormat/format_to_parts.ts:262/304/316/331` — `roundedNumber × 10^exponent` 喂 plural
7. `.references/formatjs/packages/ecma402-abstract/NumberFormat/ComputeExponent.ts` — `e/c` 来源
8. `pkg.go.dev/golang.org/x/text/feature/plural` — UNDER CONSTRUCTION + 无 c/e 操作数
9. `golang.org/x/text/internal/cldr` — CLDR 32 数据基线
11. `.references/intl/intl.go` — translate-agent/intl 的 codegen 模式
12. `.references/ext/src/ecma402/plural_rules.c` — ICU `PluralRules::forLocale`
13. SPECS/00-vision-and-scope.md §5.3, §8 — embed-only 数据策略与未决问题
14. CLDR 47 release notes — 复数规则在 32→47 间的多次调整
15. R03-numberformat.md §6 — 操作数契约（同步引用）
