# SPEC 40 — PluralRules

> **Status:** Draft (2026-05-08)
> **Owner:** `pluralrules/` + `internal/ecma402/pluralrules/` + `tools/gen-plural-rules/`
> **Reference contract:** `.references/ecma402/spec/pluralrules.html` first, then `formatjs/packages/intl-pluralrules/` + `formatjs/packages/ecma402-abstract/PluralRules/`

## Overview

定义 `pluralrules.PluralRules` 公开 API、`OperandsRecord` 类型(SSOT)、cardinal / ordinal 规则的 codegen 策略、typed select / range 算法、与 NumberFormat 的 compact operand 契约。

本 SPEC 不重定义:
- `Locale` 结构 → 见 [SPEC 10 §Locale 结构](./10-locale.md#locale-结构)
- `Decimal` 类型与运算 → 见 [SPEC 21 §Decimal API](./21-number-math.md#decimal-api)
- digit option 舍入与补零实现 → 见 [SPEC 20 §Digit Formatting SSOT](./20-numberformat.md#23-digit-formatting-ssot)
- NumberFormat compact path → 见 [SPEC 20 §Compact Notation](./20-numberformat.md#41-compact-notation-与-pluralrules-契约)
- CLDR 数据生成器架构 → 见 [SPEC 50 §Codegen](./50-cldr-data.md#codegen)
- 抽象操作总入口 → 见 [SPEC 12 §Abstract Ops](./12-abstract-operations.md)

---

## 1. 公开 API

### 1.1 类型

```go
package pluralrules

type Type uint8
const (
    Cardinal Type = iota
    Ordinal
)

type Category uint8
const (
    Zero Category = iota
    One
    Two
    Few
    Many
    Other
)

func (c Category) String() string  // "zero" | "one" | "two" | "few" | "many" | "other"

type PluralRules struct{ /* 不可变;含 resolved + 编译后的 plural 函数指针 */ }

type Options struct {
    Type                 Type
    MinimumIntegerDigits int
    FractionDigits       FractionDigitOptions
    SignificantDigits    SignificantDigitOptions
    RoundingIncrement    int
    RoundingMode         RoundingMode
    RoundingPriority     RoundingPriority
    TrailingZeroDisplay  TrailingZeroDisplay
    Notation             Notation
    CompactDisplay       CompactDisplay
}

type FractionDigitOptions struct{ /* opaque */ }
type SignificantDigitOptions struct{ /* opaque */ }

func FractionDigits(minimum, maximum int) FractionDigitOptions
func MinimumFractionDigits(n int) FractionDigitOptions
func MaximumFractionDigits(n int) FractionDigitOptions
func SignificantDigits(minimum, maximum int) SignificantDigitOptions
func MinimumSignificantDigits(n int) SignificantDigitOptions
func MaximumSignificantDigits(n int) SignificantDigitOptions

func New(loc locale.Locale, opts ...Options) (*PluralRules, error)

func (r *PluralRules) SelectInt(v int) Category
func (r *PluralRules) SelectInt64(v int64) Category
func (r *PluralRules) SelectUint(v uint) Category
func (r *PluralRules) SelectUint64(v uint64) Category
func (r *PluralRules) SelectFloat64(v float64) (Category, error)
func (r *PluralRules) SelectDecimal(v string) (Category, error)

func (r *PluralRules) SelectRangeInt(start, end int) Category
func (r *PluralRules) SelectRangeInt64(start, end int64) Category
func (r *PluralRules) SelectRangeUint(start, end uint) Category
func (r *PluralRules) SelectRangeUint64(start, end uint64) Category
func (r *PluralRules) SelectRangeFloat64(start, end float64) (Category, error)
func (r *PluralRules) SelectRangeDecimal(start, end string) (Category, error)
func (r *PluralRules) ResolvedOptions() ResolvedOptions
```

**MUST** 规则:

1. `New` **必须**在构造期完成全部选项校验,失败返回 `error`。
2. `New` **最多接受一个** `Options` 值。`New(loc)` 等价 JS 省略 options;`New(loc, Options{})` 等价 JS 传空 options object;传入多个 `Options` 必须返回 `ErrInvalidOption`。
3. 整数 typed select / range 方法 **必须不返回** `error`;它们没有运行时解析失败路径。
4. `SelectFloat64` / `SelectDecimal` 与对应 range 方法 **必须**对非有限数、非法十进制字符串返回 `ErrInvalidValue` 包装错误,而不是把用户输入错误静默映射为 `Other`。
5. `PluralRules` 是不可变值;`*PluralRules` 上的所有方法 **必须**并发安全。
6. `Options` 是唯一公开配置值;禁止恢复 functional options 或多 options merge。
7. `Category.String()` 返回值 **必须**与 ECMA-402 字符串表示一致(便于 messageformat-go 的 `:plural` 函数直接 case 分支)。

### 1.2 选项

```go
type Options struct {
    Type                 Type
    MinimumIntegerDigits int
    FractionDigits       FractionDigitOptions
    SignificantDigits    SignificantDigitOptions
    RoundingIncrement    int
    RoundingMode         RoundingMode
    RoundingPriority     RoundingPriority
    TrailingZeroDisplay  TrailingZeroDisplay
    Notation             Notation
    CompactDisplay       CompactDisplay
}

func FractionDigits(minimum, maximum int) FractionDigitOptions
func MinimumFractionDigits(n int) FractionDigitOptions
func MaximumFractionDigits(n int) FractionDigitOptions
func SignificantDigits(minimum, maximum int) SignificantDigitOptions
func MinimumSignificantDigits(n int) SignificantDigitOptions
func MaximumSignificantDigits(n int) SignificantDigitOptions
```

**MUST** 规则:

1. 公开选项必须覆盖 ECMA-402 `Intl.PluralRules` number-format digit surface: notation、compactDisplay、fraction/significant digits、roundingIncrement、roundingMode、roundingPriority、trailingZeroDisplay。
2. `Type` **必须**是 Go typed enum,零值 `Cardinal` 是默认值;禁止公开裸字符串枚举。
3. `New` 最多接受一个 `Options` 值。传入多个 `Options` 是调用点错误,必须返回 `ErrInvalidOption`。
4. `FractionDigitOptions` / `SignificantDigitOptions` 必须用构造函数创建,从而保留“未设置 / 显式设置 minimum / 显式设置 maximum”的差异。

> **Why typed Options**:
> 1. Go 调用者应在编译期看到可选值边界;`"ordinal"` 这类字符串只属于 JS option object,不是 Go API 的自然形状。
> 2. 一个 `Options` 值比 functional options 更容易比较、缓存、文档化,也不会把状态隐藏在闭包执行顺序里。
> 3. messageformat-go 若持有 ICU 字符串,应在 adapter 边界做一次显式映射;不能把上游字符串透传压力转嫁给 go-intl 的长期公共 API。
>
> **Rejected**: `WithType(string)` —— 把拼写错误推迟到运行时,且让长期 API 被 JS object model 牵引。
> **Rejected**: 同时接受 string 与 typed enum —— 双轨入参 = 校验路径分叉 + cache key 分叉。

#### 1.2.1 Digit option pipeline

PluralRules 的 digit options 是 NumberFormat digit pipeline 的子集。`pluralrules` 包 **必须**把 integer/fraction/significant digits 与 rounding options 映射为 `internal/ecma402/numberformat.DigitOptions`,并调用 `internal/ecma402/numberformat.FormatNumericToString` 得到 `ResolvePlural` 使用的 formatted string 和 rounded numeric value。

**MUST** 规则:

1. PluralRules 默认 digit options **必须**是 `minimumIntegerDigits=1`、`minimumFractionDigits=0`、`maximumFractionDigits=3`、`roundingIncrement=1`、`roundingMode="halfExpand"`、`roundingPriority="auto"`、`trailingZeroDisplay="auto"`。
2. `Select` / `SelectRange` **必须**从共享 digit formatter 的输出字符串计算 operands;**禁止**在 `pluralrules` 包内复制 `trimFraction`、`padMinimumIntegerDigits` 或 fixed rounding 逻辑。
3. 负数选择 category 时 **必须**去掉格式化字符串前缀 `-`,因为 ECMA-402 operands 基于绝对值;zero scale 与 trailing zero 仍由共享 formatter 保留。

> **Why**: PluralRules 对 digit options 的可观测面不是最终字符串,而是 operands 的 `v/w/f/t`。这些字段依赖同一套 trailing-zero 规则;共享 formatter 是防止 NumberFormat 与 PluralRules 语义分叉的最小边界。

#### 1.2.2 typed 常量与 ResolvedOptions

```go
type ResolvedOptions struct {
    Locale          locale.Locale
    Type            Type
    PluralCategories []Category
    // ... digit options
}
```

### 1.3 调用示例

```go
pr, err := pluralrules.New(locale.MustParse("en"))
// pr.SelectInt(1)   == pluralrules.One
// pr.SelectInt(2)   == pluralrules.Other
// pr.SelectDecimal("0.5") == pluralrules.Other, nil

pr, _ := pluralrules.New(locale.MustParse("en"), pluralrules.Options{Type: pluralrules.Ordinal})
// pr.SelectInt(1) == pluralrules.One   ("1st")
// pr.SelectInt(2) == pluralrules.Two   ("2nd")
// pr.SelectInt(3) == pluralrules.Few   ("3rd")
// pr.SelectInt(4) == pluralrules.Other ("4th")

cat := pr.SelectRangeInt(1, 5) // Other
```

---

## 2. OperandsRecord <a id="operandsrecord"></a> <a id="operands"></a>

`OperandsRecord` 是 PluralRules 与 NumberFormat 共享的 ECMA-402 操作数表示。**SSOT 在本 SPEC**;实际 Go 类型定义位于 `internal/ecma402/pluralrules/operands.go`,本包与 `internal/ecma402/numberformat/` 共享。

### 2.1 类型

```go
package pluralrules // (实际位于 internal/ecma402/pluralrules)

type OperandsRecord struct {
    N OperandValue // |x| 原始量级数值(用于 plural rule 中的 "n" 引用)
    I OperandValue // integer digits, |trunc(x)|
    V int      // number of fraction digits(含 trailing 0)
    W int      // number of fraction digits(去 trailing 0)
    F OperandValue // fraction digits as integer(含 trailing 0)
    T OperandValue // fraction digits as integer(去 trailing 0)
    C int      // compact exponent(== E)
    E int      // compact/scientific exponent
}
```

**MUST** 规则:

1. `OperandsRecord` 字段 **必须**与 ECMA-402 §16.5.1 GetOperands + ES 2024 `c/e` 扩展一致。
2. `C` 与 `E` **必须**始终相等(ECMA-402 显式约束:c 是 e 的别名);存两份字段为可读性,但赋值时一并写。
3. `N/I/F/T` **必须**使用精确 `OperandValue`;禁止把 CLDR 离散规则降级为 `float64`、`uint64` 或公开 `*big.Int`。
4. `OperandsRecord` 类型 **必须**位于 `internal/ecma402/pluralrules/operands.go`;`pluralrules` 与 `numberformat` 包通过该路径共享 import。

> **Why**:
> 1. ECMA-402 OperandsRecord 是 plural rule DSL 的输入域,字段名与字段语义不可改。
> 2. `OperandValue` 保留十进制 digit view,让 `%`、整数比较和区间判断对超过 IEEE-754 safe integer 的输入仍然准确。
> 3. 共享路径 `internal/ecma402/pluralrules/operands.go` 而非 `pluralrules/operands.go`:NumberFormat 的 compact path 需要从 `internal/` 路径访问,避免循环依赖。
>
> **Rejected**:
> - `float64` 承载 `N` —— plural rule 是离散规则,二进制浮点会在大整数和小数边界制造错误 category。
> - `uint64` / `*big.Int` 直接作为字段类型 —— 会把底层表示泄漏给 generated rule call sites。
> - `OperandsRecord` 放在 `pluralrules/` 包公开层 —— NumberFormat 不应 import `pluralrules` 公开包(分层倒置)。

### 2.2 GetOperands

```go
// internal/ecma402/pluralrules/operands.go
func GetOperands(formatted string, exponent int) OperandsRecord
```

**MUST** 规则:

1. `GetOperands` 实现 **必须**与 formatjs `GetOperands.ts` 输出语义一致(对 trailing zero 行为)。
2. `i / v / w / f / t` **必须**从 `formatted` 字符串(已经 `FormatNumericToString` 输出,含 `mnfd` 强制的 trailing 0)抽取:
   - `v` = 小数串长度(含 trailing 0)
   - `w` = 去 trailing 0 后小数串长度
   - `f` = 小数串作整数
   - `t` = 去 trailing 0 后小数串作整数
   - `i` = 整数部分作整数(`|trunc(x)|`)
3. `n` **必须**是 `formatted` 的绝对十进制值,以 `OperandValue` 保存;禁止通过 `strconv.ParseFloat` 得到。
4. `c / e` **必须**等于参数 `exponent`。
5. 非有限数不进入 `GetOperands`;公开 `SelectFloat64` 必须在边界返回 `ErrInvalidValue`。

> **Why**: 从 `formatted` 字符串抽 `v/w/f/t` 而非从 `decimal.Decimal` 抽,是因为 trailing zero 是"格式化器决定的可观测属性"(`mnfd=2` 时 `1` → `"1.00"` 即 `v=2, w=0`),不是数学属性。FormatJS 从字符串抽,语义对齐。
>
> **Note**: 本 SPEC 明确 SSOT 是 `formatted` 字符串(formatjs 行为);trailing 完全看字符串。这避免 decimal backend 与显示 digit view 双路径分叉。

---

## 3. CLDR Plural Rules Codegen

### 3.1 决策:codegen 到 Go 源

**决定**:active scope PluralRules 实现 **必须**通过 codegen 从 CLDR JSON `cldr-core/supplemental/plurals.json` + `ordinals.json` + `pluralRanges.json` 生成 Go 函数与分类表到 `internal/cldr/plural/` 和 `internal/cldr/plurals.go`。

> **Why**:
> 1. formatjs `intl-pluralrules/scripts/plural-rules-compiler.ts` 已在 TS 端把 plural DSL 编译成直接执行的函数;Go 端保持同一产物形态,不在运行时解释 DSL。
> 2. 与 SPEC 50 §"embed-only / 无运行时 I/O" 一致 —— 数据即 Go 源。
> 3. 编译后函数零分配查询 vs runtime interpreter 多 2–3 倍开销。
> 4. `golang.org/x/text/feature/plural` 不可用(详见 §3.2)。
>
> **Rejected**:
> - **Runtime interpreter**:DSL 解析 + tree-walk 在 hot path 显著慢于编译后函数;且需要写一个稳健 DSL parser。
> - **复用 `golang.org/x/text/feature/plural`**:见 §3.2。

### 3.2 拒绝复用 `golang.org/x/text/feature/plural` <a id="rejected-x-text-plural"></a>

**MUST** 规则:

1. **禁止** `pluralrules` / `internal/ecma402/pluralrules/` / `tools/gen-plural-rules/` 任何位置 import `golang.org/x/text/feature/plural`。
2. 拒绝理由(写入此 SPEC 与 CLAUDE.md "Forbidden"):
   - **CLDR 数据基线不可控**:`golang.org/x/text/feature/plural` 的数据版本与 go-intl 钉定的 CLDR 48.1.0 / formatjs 基线不同步。CLDR 33+ 重写了 Welsh/Cymraeg、Hebrew、Polish 复数规则;CLDR 41+ 改了 Russian ordinal。数据基线不一致即与 formatjs 字节不等。
   - **缺 c/e 操作数**:`plural.Select(t, scale, digits)` 仅二参,无法承载 `Intl.NumberFormat` notation=compact 的 `e` 操作数。少数语言(波兰语、捷克语)的 `1 thousand` vs `1K` 复数类别区分必须有 `e`。
   - **包标注 UNDER CONSTRUCTION**:pkg.go.dev 显式注明 "This package is UNDER CONSTRUCTION ...";不能作为生产依赖。

> **Why**: 我们的目标年(2026)下 `x/text/feature/plural` 既不准确也不完整,且没有可见路径得到修复(CLDR 数据更新由 x/text 团队节奏控制,与 ECMA-402 conformance 解耦)。

### 3.3 Codegen 工具(`tools/gen-plural-rules/`)

**MUST** 规则:

1. codegen 入口 **必须**位于 `tools/gen-plural-rules/main.go`,独立 Go module(独立 `go.mod`,不污染主 module 依赖图)。
2. codegen **必须**保持 stdlib-only:用 `encoding/json` 读取 CLDR JSON,用确定性字符串构造输出,最后经 `go/format` 格式化;**禁止** `dave/jennifer` 或其他 codegen 框架。
3. 输入 **必须**是 pinned CLDR JSON: `cldr-core/supplemental/plurals.json` + `ordinals.json` + `pluralRanges.json`,版本钉化通过 [SPEC 50 §Version Pin](./50-cldr-data.md#version-pin)。
4. 输出位置:
   - `internal/cldr/plural/cardinal_rules.go`
   - `internal/cldr/plural/ordinal_rules.go`
   - `internal/cldr/plural/range_rules.go`
   - `internal/cldr/plural/categories.go`
   - `internal/cldr/plurals.go`
5. 每个 locale 输出一个独立 Go 函数;`CardinalRule` / `OrdinalRule` switch 索引到具体函数。
6. **禁止**输出未使用的操作数表达式(对应 formatjs `should-emit-*.ts`,只生成 rule 实际引用的操作数判定)。

```go
// 生成器内部形态(示意,无实现)
package main

func parsePluralRules(path string) (cardinal, ordinal map[string][]Rule, err error)
func parsePluralRanges(path string) (map[string]map[RangeKey]Category, error)
func renderRuleFile(kind string, rules map[string][]Rule) string
func renderRangeFile(ranges map[string]map[RangeKey]Category) string
func renderCategoriesFile(cardinal, ordinal map[string][]Rule) string
```

### 3.4 Codegen 输出形态

**MUST** 规则:

1. 每个 locale 一个小写规则函数,命名为 `cardinalEn` / `cardinalEnUS` / `ordinalEn` 这一类稳定 Go 标识符。
2. 函数体 **必须**是 if-else 链,直接对应 CLDR DSL 转译,无运行时解析:
   ```go
   // cardinal_rules.go (片段;由 tools/gen-plural-rules 生成)
   func CardinalRule(loc string) (func(ecma402pr.OperandsRecord) ecma402pr.Category, bool) {
       switch loc {
       case "en":
           return cardinalEn, true
       }
       return nil, false
   }

   func cardinalEn(o ecma402pr.OperandsRecord) ecma402pr.Category {
       if o.I == 1 && o.V == 0 {
           return ecma402pr.One
       }
       return ecma402pr.Other
   }

   func cardinalPl(o ecma402pr.OperandsRecord) ecma402pr.Category {
       if o.I == 1 && o.V == 0 {
           return ecma402pr.One
       }
       if o.V == 0 && o.I%10 >= 2 && o.I%10 <= 4 && (o.I%100 < 12 || o.I%100 > 14) {
           return ecma402pr.Few
       }
       if o.V == 0 && o.I != 1 && (o.I%10 == 0 || o.I%10 == 1) {
           return ecma402pr.Many
       }
       return ecma402pr.Other
   }
   ```
3. 函数 **必须**全部并发安全(纯读 OperandsRecord,无可变状态)。
4. 每个 codegen 文件头部 **必须**含 `// Code generated by tools/gen-plural-rules; DO NOT EDIT.` 标头 + CLDR 版本号;**禁止**生成时间戳,保证同一输入输出 byte-stable。

### 3.5 PluralRanges 数据 <a id="plural-ranges"></a>

**MUST** 规则:

1. `pluralRanges.json` 数据 **必须**与 `plurals.json` / `ordinals.json` 同流水线 codegen,输出到 `internal/cldr/plural/range_rules.go`。
2. 形态:
   ```go
   type rangeKey struct {
       Start, End ecma402pr.Category
   }

   func Range(loc, typ string, start, end ecma402pr.Category) (ecma402pr.Category, bool) {
       if typ != "cardinal" {
           return ecma402pr.Other, false
       }
       ranges, ok := cardinalRanges[loc]
       if !ok {
           return ecma402pr.Other, false
       }
       result, ok := ranges[rangeKey{start, end}]
       return result, ok
   }

   var cardinalRanges = map[string]map[rangeKey]ecma402pr.Category{
       // ...
   }
   ```
3. accessor:`Range(loc, typ string, start, end Category) (Category, bool)`,bool 表示是否命中表(用于 fallback 决策)。

---

## 4. Compact Operand Contract <a id="compact-operand-contract"></a> <a id="selectformatted"></a>

NumberFormat compact-notation 路径 **不**通过公开 `pluralrules.PluralRules` 实例。它直接复用 `internal/ecma402/pluralrules.GetOperands` 和 codegen 输出的 `internal/cldr/plural.CardinalRule`,避免公开包之间产生循环或虚假的 internal slot API。

### 4.1 签名

```go
package pluralrules

func GetOperands(formatted string, exponent int) OperandsRecord
```

```go
package plural

func CardinalRule(loc string) (func(ecma402pr.OperandsRecord) ecma402pr.Category, bool)
```

### 4.2 算法

**MUST** 规则:

1. NumberFormat compact suffix 选择 **必须**:
   ```text
   1. value, exponent := ComputeExponent(...)
   2. formatted       := FormatNumericToString(value / 10^exponent).String
   3. ops             := GetOperands(formatted, exponent)
   4. rule            := plural.CardinalRule(localeTag)
   5. category        := rule(ops)
   6. pattern         := CompactPattern(numberingSystem, compactDisplay, exponent, category)
   ```
2. `formatted` **必须**是显示数的已舍入十进制字符串,并保留 trailing zero;`GetOperands` 从它派生 `n / i / v / w / f / t`,并把 `c/e` 设置为 compact exponent。
3. **必须**与 [SPEC 20 §4.1 Compact Notation 与 PluralRules 契约](./20-numberformat.md#41-compact-notation-与-pluralrules-契约) 文字一致:
   - 输入:`(formattedDisplayDecimal, exponent)` —— `formattedDisplayDecimal` 是"显示数"(已除 `10^exponent`),`exponent` 是 compact 指数。
   - `OperandsRecord.C` 与 `OperandsRecord.E` **必须**等于 `exponent`。
4. **公开 API 不增加**:`SelectFormatted` / `ResolvePlural` 不作为 public surface 出现;NumberFormat 通过 internal 包和 generated rule 函数组合完成选择。

> **Why**: compact plural 的稳定点是 operand 生成与 generated CLDR rule,不是一个 JS-style `SelectFormatted` 方法。Go 里把公开 formatter 实例绕回另一个公开 formatter 只会制造状态与缓存的假依赖。
>
> **Rejected**: 公开 `PluralRules.SelectFormatted` —— 这不是 ECMA-402 公共 API,也不是 Go 用户需要的语言。
> **Rejected**: 让 NumberFormat 复制 plural DSL 或规则表 —— 规则只允许来自 `internal/cldr/plural` codegen。

---

## 5. SelectRange 算法

**MUST** 规则(对应 ECMA-402 §16.5.4 三步算法 + formatjs `ResolvePluralRange.ts`):

```text
function SelectRangeInt(start, end int) Category:
    sCat := SelectInt(start)
    eCat := SelectInt(end)

    // 步骤 1:已格式化字符串相等 → 返回 sCat(避免 "1–1" 走 range 表的边角)
    sFmt := FormatNumericToString(start)
    eFmt := FormatNumericToString(end)
    if sFmt == eFmt:
        return sCat

    // 步骤 2:locale 没有 pluralRanges 数据 → 回落 eCat(end-class 兜底)
    rangeMap, ok := pluralRanges[localeData]
    if !ok:
        return eCat

    // 步骤 3:查 pluralRanges["${sCat}_${eCat}"];未命中回落 eCat
    cat, ok := rangeMap[{sCat, eCat}]
    if !ok:
        return eCat
    return cat
```

**MUST** 规则:

1. 三步顺序 **必须**严格按 ECMA-402 §16.5.4。
2. "已格式化字符串相等"判定 **必须**通过 `FormatNumericToString(start) == FormatNumericToString(end)` 字符串比较,**不**通过 `decimal.Cmp`(数学相等不等价于格式化相等,例:`1.00` vs `1` 数学相等但字符串不等)。
3. locale 无 `pluralRanges` 数据时 **必须**回落 `eCat`(end-class),**禁止**回落 `sCat` 或返回 error。
4. `rangeMap` 未命中 `(sCat, eCat)` **必须**回落 `eCat`,**禁止**自动尝试 `(sCat, Other)` 或 `(Other, eCat)` 等启发式 fallback。

> **Why**: 步骤 1 短路"1–1"是 FormatJS 的固化行为;步骤 2 回落 end-class 是 ECMA-402 §16.5.4 显式规定;启发式 fallback 会引入与 FormatJS 不一致。
>
> **Rejected**: 数学比较短路 step 1 —— 与 formatjs 行为不一致。
> **Rejected**: 报错(无 ranges 数据)—— 与 `Select` 不返 error 一致性冲突。

---

## 6. ResolvedOptions

```go
type ResolvedOptions struct {
    Locale                   locale.Locale
    Type                     Type
    MinimumIntegerDigits     int
    MinimumFractionDigits    int
    MaximumFractionDigits    int
    MinimumSignificantDigits int
    MaximumSignificantDigits int
    PluralCategories         []Category
    Notation                 Notation
    CompactDisplay           CompactDisplay
    RoundingIncrement        int
    RoundingMode             RoundingMode
    RoundingPriority         RoundingPriority
    TrailingZeroDisplay      TrailingZeroDisplay
}
```

**MUST** 规则:

1. 字段顺序 **必须**与 ECMA-402 §16.4.5 spec 顺序一致。
2. `PluralCategories` **必须**返回 locale 实际定义的类别(从 codegen 的 `cardinal_rules.go` / `ordinal_rules.go` per-locale 表反查),**禁止**返回所有 6 类硬编码列表。
3. **必须**返回值类型(非指针),并发安全。

---

## 7. 错误模型

**MUST** 规则:

1. **必须**重导出 sentinel:`ErrInvalidOption` 与 `ErrInvalidValue`。
2. 构造期错误 **必须**用 `fmt.Errorf("pluralrules: ...: %w", ErrInvalidOption)` 包装。
3. **禁止** `panic` 任何用户路径。
4. 运行时无效输入(非有限 float、非法 decimal 字符串)**必须**返回 `ErrInvalidValue` 包装错误;整数方法没有无效输入路径。

---

## 8. 性能目标

**MUST** 规则:

1. `SelectInt64` cached **必须** ≤ 200 ns/op(对齐 [SPEC 71 §阈值](./71-benchmark.md#thresholds))。
2. 整数 select hot path **必须**零分配。
3. `SelectRangeInt64` cached **必须** ≤ 400 ns/op(两次 select + map 查表)。

> **Why**: messageformat-go `:plural` 函数对每条含复数变量的消息 N 次 `Select`;200 ns 是 messageformat-go 性能 SLA 的关键预算。

---

## 9. Forbidden

- **禁止** import `golang.org/x/text/feature/plural`(任何位置)—— 数据基线不可控 + 缺 c/e + UNDER CONSTRUCTION。
- **禁止** 引入 `dave/jennifer` 或其他 codegen 框架 —— 当前规模用 stdlib JSON 读取 + 确定性字符串输出 + `go/format` 即可。
- **禁止** runtime DSL interpreter —— 必须 codegen。
- **禁止** `OperandsRecord.N/I/F/T` 使用 `float64` 或公开 big-number 类型 —— 必须用精确 `OperandValue`。
- **禁止** `OperandsRecord` 类型放在 `pluralrules/` 公开包 —— 必须位于 `internal/ecma402/pluralrules/operands.go`。
- **禁止** 恢复公共 `Select(any)` / `SelectRange(any, any)` coercion API。
- **禁止** `SelectFormatted` / `ResolvePlural` 暴露为公开 API —— NumberFormat compact path 通过 internal operand builder 与 generated cardinal rule 完成选择。
- **禁止** NumberFormat 复制 plural DSL 或规则表 —— plural category 只允许来自 `internal/cldr/plural` codegen。
- **禁止** 数学比较短路 `SelectRange` step 1(必须字符串比较)。
- **禁止** `SelectRange` 启发式 fallback(`(sCat, Other)` / `(Other, eCat)` 等)—— 必须严格回落 `eCat`。
- **禁止** `PluralCategories` 硬编码所有 6 类列表 —— 必须从 codegen 数据反查。
- **禁止** `panic` 任何用户路径。
- **禁止** codegen 输出未使用的操作数表达式(should-emit 优化)。

---

## 10. Acceptance Criteria

- [ ] `tools/gen-plural-rules/` 是独立 Go module(`tools/gen-plural-rules/go.mod` 存在,主 module 不引入)。
- [ ] codegen 工具输入 CLDR `plurals.json` / `ordinals.json` / `pluralRanges.json` v48.1.0,输出 `internal/cldr/plural/cardinal_rules.go`、`ordinal_rules.go`、`range_rules.go`、`categories.go` 与 `internal/cldr/plurals.go`,字节稳定(同一输入 + 同一工具版本 = 同一输出)。
- [ ] 输出文件头含 `// Code generated by tools/gen-plural-rules; DO NOT EDIT.` + CLDR 版本号。
- [ ] codegen 输出在 SPEC 40 锁定 CLDR 版本(48.1.0)下,与 `formatjs/packages/intl-pluralrules/tests/index.test.ts` 中已机械抽取的 `.select()` fixture 字节相等;未抽取的 `selectRange`、`supportedLocalesOf` 或复杂 Vitest shape 必须由 `.skip-list.json` 审计或单独手写 fixture 覆盖。
- [ ] codegen 工具实现 **不** import `dave/jennifer`(`grep -r "dave/jennifer" tools/gen-plural-rules/` 空)。
- [ ] `pluralrules`、`internal/ecma402/pluralrules`、`tools/gen-plural-rules` 均不 import `golang.org/x/text/feature/plural`(`grep -r "x/text/feature/plural" .` 在仓库主 module 空)。
- [ ] `pluralrules.New(locale.MustParse("en")).SelectInt(1) == One`。
- [ ] `pluralrules.New(locale.MustParse("en")).SelectInt(2) == Other`。
- [ ] `pluralrules.New(locale.MustParse("en"), Options{Type: Ordinal}).SelectInt(1) == One`(1st)。
- [ ] `pluralrules.New(locale.MustParse("en"), Options{Type: Ordinal}).SelectInt(2) == Two`(2nd)。
- [ ] `pluralrules.New(locale.MustParse("en"), Options{Type: Ordinal}).SelectInt(3) == Few`(3rd)。
- [ ] `pluralrules.New(locale.MustParse("en"), Options{Type: Ordinal}).SelectInt(4) == Other`(4th)。
- [ ] `pluralrules.New(locale.MustParse("pl")).SelectInt(1) == One`。
- [ ] `pluralrules.New(locale.MustParse("pl")).SelectInt(2) == Few`。
- [ ] `pluralrules.New(locale.MustParse("pl")).SelectInt(5) == Many`。
- [ ] `pluralrules.New(locale.MustParse("ar")).SelectInt(0) == Zero`(阿拉伯语 zero 类别)。
- [ ] `pluralrules.New(locale.MustParse("en")).SelectRangeInt(1, 5) == Other`。
- [ ] `pluralrules.New(locale.MustParse("en")).SelectRangeInt(1, 1) == One`(step 1 字符串相等短路)。
- [ ] `pluralrules.New(locale.MustParse("zh")).SelectRangeInt(1, 5) == Other`(zh cardinal 仅 Other,所有 range 回落 Other)。
- [ ] NumberFormat compact path 通过 `internal/ecma402/pluralrules.GetOperands` + `internal/cldr/plural.CardinalRule` 选择 plural category,与 formatjs `format_to_parts.ts:262/304/316/331` 行为字节相等(用 `numberformat.New(loc-pl, numberformat.Options{Notation: numberformat.CompactNotation}).FormatInt(1500)` 做端到端测试)。
- [ ] `pluralrules.New(loc).SelectFormatted(...)` 与 `internal/ecma402/pluralrules.ResolvePlural` 均不存在;compact plural 选择不是公开 API。
- [ ] `OperandsRecord` 类型位于 `internal/ecma402/pluralrules/operands.go`(单文件 SSOT);`pluralrules` 包不重定义。
- [ ] `OperandsRecord` 字段集合精确为 `{N,I,F,T OperandValue; V,W,C,E int}`(8 字段,通过反射测试断言)。
- [ ] `ResolvedOptions().PluralCategories` 对 `en` cardinal 返回 `[One, Other]`,对 `ar` cardinal 返回 `[Zero, One, Two, Few, Many, Other]`。
- [ ] `go test -race ./pluralrules/...` 通过(含 `TestPluralRules_ConcurrentSelect` 100 goroutine × 1000 调用)。
- [ ] `go vet ./pluralrules/...` 干净。
- [ ] `task data:verify` 通过,确认 CLDR 生成数据与当前仓库 byte-equal。
- [ ] `pluralrules.New(loc).SelectFloat64(math.NaN())` 返回 `ErrInvalidValue`,不 panic。
- [ ] benchmark `BenchmarkPluralRules_Select_Cardinal_Int64_EN` ≤ 200 ns/op(SPEC 71 阈值)。
- [ ] benchmark `BenchmarkPluralRules_SelectRange_EN` ≤ 400 ns/op。
- [ ] codegen 输出在 CLDR bump(48.1.0 → 49.0)时,通过 `task data:diff` 检测变更并 block 未审查增量。

---

## 11. References

### Primary

- `.references/formatjs/packages/intl-pluralrules/scripts/plural-rules-compiler.ts` — plural DSL → JS 函数串(Go 端输出等价 Go 函数)
- `.references/formatjs/packages/intl-pluralrules/index.ts` — `PluralRuleSelect` / `PluralRuleSelectRange`
- `.references/formatjs/packages/intl-pluralrules/tests/index.test.ts` — 主 conformance fixture
- `.references/formatjs/packages/ecma402-abstract/PluralRules/GetOperands.ts` — `OperandsRecord` 类型定义
- `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePlural.ts` — Format-then-GetOperands
- `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePluralRange.ts` — selectRange 三步
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/format_to_parts.ts:262/304/316/331` — compact 路径喂 plural

### Secondary

- CLDR release notes — 复数规则调整记录(用作版本钉化对齐)
- `pkg.go.dev/golang.org/x/text/feature/plural` — 反例:UNDER CONSTRUCTION + 无 c/e
- `golang.org/x/text/internal/cldr` — 反例:数据基线不由 go-intl 控制
- `.references/intl/intl.go` — translate-agent/intl(无 PluralRules 实现,但 codegen 模式可参考)
- `.references/ext/src/ecma402/plural_rules.c` — PHP/ICU `PluralRules::forLocale` 路径
- CLDR `cldr-core/supplemental/plurals.json` + `ordinals.json` — cardinal / ordinal 权威源
- CLDR `cldr-core/supplemental/pluralRanges.json` — pluralRanges 权威源

### Project Cross-References

- [SPEC 12 §Abstract Ops](./12-abstract-operations.md) — shared decimal boundary / `ErrInvalidOption`
- [SPEC 10 §Locale 结构](./10-locale.md#locale-结构) — `Locale` 入参类型
- [SPEC 20 §Compact Notation](./20-numberformat.md#41-compact-notation-与-pluralrules-契约) — NumberFormat compact path operand 契约(本 SPEC 与之文字一致)
- [SPEC 21 §Decimal API](./21-number-math.md#decimal-api) — `decimal.Decimal` 入参类型
- [SPEC 50 §Codegen](./50-cldr-data.md#codegen) — `tools/gen-cldr` 与 `tools/gen-plural-rules` 均使用 stdlib-only deterministic codegen 约束
- [SPEC 50 §Version Pin](./50-cldr-data.md#version-pin) — CLDR 版本锁(48.1.0)
- [SPEC 60](./60-facade.md) — root namespace ownership; root `intl.SelectPlural*` one-shot helpers are outside the long-term public surface.
- [SPEC 71 §阈值](./71-benchmark.md#thresholds) — 性能基线(200 ns/op)
