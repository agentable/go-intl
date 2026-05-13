# SPEC 20 — NumberFormat

> **Status:** Draft (2026-05-08)
> **Closes Open Question:** SPECS/00 §8 Q2 (Functional Options vs Config Struct)
> **Owner:** `numberformat/`
> **Reference contract:** `.references/ecma402/spec/numberformat.html` first, then `formatjs/packages/intl-numberformat/` + `formatjs/packages/ecma402-abstract/NumberFormat/`

## Overview

定义 `numberformat.NumberFormat` 公开 API、构造期校验、typed format / parts / range 行为,以及 NumberFormat ↔ PluralRules 在 compact notation 上的内部契约。active scope 必须支持 ECMA-402 §15(ES 2025)全部 style/notation 组合:`decimal | percent | currency | unit` × `standard | scientific | engineering | compact`。

本 SPEC 不重定义:
- `Decimal` 类型与七种 rounding mode → 见 [SPEC 21 §Decimal API](./21-number-math.md#decimal-api)
- `Locale` 与 `Locale.NumberingSystem()` → 见 [SPEC 10 §Locale 结构](./10-locale.md#locale-结构)
- `OperandsRecord` → 见 [SPEC 40 §OperandsRecord](./40-pluralrules.md#operandsrecord)
- CLDR 数据 schema(`numbers.json` / `currencies.json` / `units.json`)→ 见 [SPEC 50 §Schema](./50-cldr-data.md#schema)
- 抽象操作总入口 → 见 [SPEC 12 §Abstract Ops](./12-abstract-operations.md)

---

## 1. 公开 API

### 1.1 构造与 Option

```go
package numberformat

type NumberFormat struct{ /* 不可变;包含 resolved + apd.Context 副本 + plural 句柄 */ }

type Style string
type Currency string
type Unit string
type Notation string

type Options struct {
    Style                Style
    Currency             Currency
    CurrencyDisplay      CurrencyDisplay
    CurrencySign         CurrencySign
    Unit                 Unit
    UnitDisplay          UnitDisplay
    MinimumIntegerDigits int
    FractionDigits       FractionDigitOptions
    SignificantDigits    SignificantDigitOptions
    RoundingIncrement    int
    RoundingPriority     RoundingPriority
    RoundingMode         RoundingMode
    TrailingZeroDisplay  TrailingZeroDisplay
    Notation             Notation
    CompactDisplay       CompactDisplay
    UseGrouping          UseGrouping
    SignDisplay          SignDisplay
    LocaleMatcher        LocaleMatcher
    NumberingSystem      string
}

func New(loc locale.Locale, opts ...Options) (*NumberFormat, error)

func (f *NumberFormat) FormatInt(v int) string
func (f *NumberFormat) FormatInt64(v int64) string
func (f *NumberFormat) FormatUint(v uint) string
func (f *NumberFormat) FormatUint64(v uint64) string
func (f *NumberFormat) FormatFloat64(v float64) string
func (f *NumberFormat) FormatDecimal(v string) (string, error)

func (f *NumberFormat) FormatIntToParts(v int) []Part
func (f *NumberFormat) FormatInt64ToParts(v int64) []Part
func (f *NumberFormat) FormatUintToParts(v uint) []Part
func (f *NumberFormat) FormatUint64ToParts(v uint64) []Part
func (f *NumberFormat) FormatFloat64ToParts(v float64) []Part
func (f *NumberFormat) FormatDecimalToParts(v string) ([]Part, error)

func (f *NumberFormat) FormatRangeInt(a, b int) string
func (f *NumberFormat) FormatRangeInt64(a, b int64) string
func (f *NumberFormat) FormatRangeUint(a, b uint) string
func (f *NumberFormat) FormatRangeUint64(a, b uint64) string
func (f *NumberFormat) FormatRangeFloat64(a, b float64) (string, error)
func (f *NumberFormat) FormatRangeDecimal(a, b string) (string, error)

func (f *NumberFormat) FormatRangeIntToParts(a, b int) []RangePart
func (f *NumberFormat) FormatRangeInt64ToParts(a, b int64) []RangePart
func (f *NumberFormat) FormatRangeUintToParts(a, b uint) []RangePart
func (f *NumberFormat) FormatRangeUint64ToParts(a, b uint64) []RangePart
func (f *NumberFormat) FormatRangeFloat64ToParts(a, b float64) ([]RangePart, error)
func (f *NumberFormat) FormatRangeDecimalToParts(a, b string) ([]RangePart, error)

func (f *NumberFormat) ResolvedOptions() ResolvedOptions
```

**MUST** 规则:

1. `New` **必须**在构造期完成全部选项校验,失败即返回 `error`。
2. `New` **最多接受一个** `Options` 值。`New(loc)` 等价 JS 省略 options;`New(loc, Options{})` 等价 JS 传空 options object;传入多个 `Options` 必须返回 `ErrInvalidOption`。
3. 整数和 float typed 方法 **必须不返回** `error`;`float64` 的 NaN / Infinity 是合法 NumberFormat 输入并按 locale-specific symbol 格式化。
4. Decimal-string 方法 **必须**对 malformed decimal 返回 `ErrInvalidValue`;`"NaN"` / `"Infinity"` / `"-Infinity"` 是合法 decimal 字符串。
5. `NumberFormat` 是不可变值;`*NumberFormat` 上的所有方法**必须**并发安全。
6. `ResolvedOptions` **必须**返回不可变快照(值类型);多次调用返回相等结果。

> **Why**: 构造期校验集中处理配置错误;运行时数值类型边界由方法名表达。JS 端 `format(NaN)` 不抛,所以 `FormatFloat64(math.NaN())` 仍返回 `"NaN"`;但 Go 的 decimal string 解析失败是调用者错误,必须通过 `ErrInvalidValue` 暴露。
>
> **Rejected**: 公开 `Format(v any)` / `FormatToParts(v any)` —— 把 unsupported input、malformed string、NaN display 和用户错误混在同一个热路径里。

### 1.2 Typed Options(关闭 §8 Q2)

**决定**:active scope 公开 API **必须**采用单一 typed `Options` 值。`numberformat.New(loc, numberformat.Options{Currency: numberformat.CurrencyCode("USD"), FractionDigits: numberformat.MaximumFractionDigits(2)})`。

> **Why**:
> 1. 正确值应当在 IDE 和编译器层面可发现。`CurrencyStyle`、`CompactNotation`、`HalfEvenRoundingMode` 比裸字符串更适合作为十年公共 API。
> 2. `Options` 是普通值,可比较地审查、可序列化地桥接、可稳定生成 cache key;functional option 闭包不能做到这些。
> 3. 字段语义复杂时,用小的 option 子值表达三态。`FractionDigitOptions` / `SignificantDigitOptions` 内部持有 set bit,因此 `MinimumFractionDigits(0)` 与未设置可区分。
>
> **Rejected**:
> - **Functional Options**:隐藏状态、不可序列化、不可静态发现,且让 cache key 依赖闭包执行结果。
> - **Config struct + `*T` 字段**:把每个字段变成指针噪音,让简单调用读起来像内部配置文件。
> - **Builder 链式**:校验时机分散,与"构造期一次性聚合"目标冲突。

**MUST** 规则:

1. `Options` 字段名 **必须**对应 ECMA-402 §15.4.1 选项名的 Go 形态。
2. 枚举型字段 **必须**使用包内命名类型和常量;禁止调用者传裸字符串来选择核心枚举。
3. `CurrencyCode(code string)` **必须**规范化 ISO 4217 currency code;`UnitIdentifier(unit string)` **不得**规范化大小写。ECMA-402 unit identifiers are exact canonical lowercase strings;`"METER"` must be rejected like native `Intl.NumberFormat`.
4. `FractionDigits` / `SignificantDigits` 子值 **必须**内部记录显式设置状态,以区分"显式传入零值"与"未传"。
5. `Options{}` 零值 **必须**等价 ECMA-402 默认选项。
6. `New` 最多接受一个 `Options` 值;多 options merge 不是 ECMA-402 行为,必须拒绝。
7. 校验集中在 `New`。

### 1.3 调用示例

```go
nf, err := numberformat.New(locale.MustParse("zh-Hant-TW"),
    numberformat.Options{
        Notation:       numberformat.CompactNotation,
        CompactDisplay: numberformat.ShortCompactDisplay,
    })
// nf.FormatInt(98765) == "9.9萬"
```

```go
nf, _ := numberformat.New(locale.MustParse("en-US"),
    numberformat.Options{
        Style:    numberformat.CurrencyStyle,
        Currency: numberformat.CurrencyCode("USD"),
    })
// nf.FormatFloat64(1234.5) == "$1,234.50"
```

### 1.4 Option 参数类型 — typed ECMA-402 values

**MUST** 规则:

1. 所有取自 ECMA-402 字符串字面量并集的选项(`style` / `notation` / `compactDisplay` / `currencyDisplay` / `currencySign` / `unitDisplay` / `signDisplay` / `useGrouping` / `roundingPriority` / `roundingMode` / `trailingZeroDisplay` / `localeMatcher`)**必须**以命名类型承载,底层 kind 保持 `string`,序列化形态与 JS resolvedOptions 字符串一致。
2. `numberingSystem` 仍是 `string`,因为它是 Unicode extension type,不是小枚举。
3. `Currency` **必须**通过 `CurrencyCode` 构造,避免调用点散落大小写规则;`Unit` **必须**传 canonical ECMA-402 unit identifier,不得在构造器内做 lowercase fallback。
4. `useGrouping` 的 `false` 值在 Go 端由 `UseGroupingFalse` 表达,底层仍序列化为 `"false"`。
5. 校验 **必须**在 `New` 内集中完成,失败包装 `ErrInvalidOption` 并显示用户传入值。

> **Why**: typed value 仍保留 ECMA-402 字符串作为 wire/resolved 形态,但把调用点从"猜字符串"提升为"选择明确常量"。conformance fixture 和 messageformat-go adapter 可以在边界做一次映射,不把内部公共 API 降级成 JSON 形态。

---

## 2. 选项管线(InitializeNumberFormat)

公开 API 在 `New` 内部 **必须**保留 `formatjs/packages/ecma402-abstract/NumberFormat/InitializeNumberFormat.ts` 的可观测语义,但 Go 端直接消费 typed `Options`,不经过 `GetOption(map[string]any)`:

```text
1. localeMatcher        := typed Options.LocaleMatcher, default best-fit
2. numberingSystem      := typed Options.NumberingSystem, 校验 Unicode type=numbers
3. resolvedLocale       := ResolveLocale(availableLocales,requestedLocales,localeMatcher,relevantExt={"nu"},...)
4. SetNumberFormatUnitOptions(nf, opts)            // style + currency/{...}/unit/{...}
5. notation             := typed Options.Notation, default standard
6. mnfdDefault, mxfdDefault 由 style 决定:
    decimal       => 0,3
    percent       => 0,0
    currency      => CurrencyDigits(currency, currencyData), idem
    unit          => 0,3
   compact 与 scientific/engineering 在 SetNumberFormatDigitOptions 中再调整
7. SetNumberFormatDigitOptions(nf, opts, mnfdDefault, mxfdDefault, notation)
8. compactDisplay       := typed Options.CompactDisplay, 仅 notation=compact 生效
9. useGrouping          := normalize 至 always|auto|min2|false
10. signDisplay         := auto|always|exceptZero|negative|never
```

> **Why**: 步骤顺序是 ECMA-402 算法可观测行为(不同步骤产生不同错误信息);改顺序即破坏字节相等。

### 2.1 SetNumberFormatUnitOptions

**MUST** 规则:

1. `style="currency"` 时 **必须**校验 `currency` 选项存在且为 3 字母 ISO 4217 形式,失败返回 `ErrInvalidOption`,错误消息含 currency code + locale。
2. 货币精度 **必须**通过 `internal/cldr.CurrencyDigits(code)` 取(数据来自 CLDR `supplemental/currencyData.json` `fractions` 节点 codegen 产出)。**禁止**内嵌 ISO 4217 静态表。**禁止**引入 `bojanz/currency`。
3. `style="unit"` 时 **必须**校验 `unit` 选项是 sanctioned simple unit 或 `<numerator>-per-<denominator>` 复合形式;sanctioned 列表来自 CLDR `units-constants` codegen。
4. `currencyDisplay` ∈ `code | symbol | narrowSymbol | name`;`currencyDisplay="symbol"` 是 ECMA-402 默认值。
5. `currencySign` ∈ `standard | accounting`,默认 `standard`。
6. `unitDisplay` ∈ `short | narrow | long`,默认 `short`。

> **Rejected**: `bojanz/currency` 数据非 CLDR 直出,与 formatjs 字节相等目标不兼容。

### 2.2 SetNumberFormatDigitOptions

**MUST** 规则(对应 `formatjs/.../SetNumberFormatDigitOptions.ts` 5 路分支):

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

1. `roundingIncrement` **必须**仅接受 `VALID_ROUNDING_INCREMENTS`(`1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000, 2000, 2500, 5000`),其它值返回 `ErrInvalidOption`。
2. `roundingMode` **必须**接受 ECMA-402 全部 9 种:`ceil | floor | expand | trunc | halfCeil | halfFloor | halfExpand | halfTrunc | halfEven`(算法实现见 [SPEC 21 §Rounding Modes](./21-number-math.md#rounding-modes))。
3. `trailingZeroDisplay` ∈ `auto | stripIfInteger`,默认 `auto`。
4. `roundingPriority` ∈ `auto | morePrecision | lessPrecision`,默认 `auto`。

> **Why**: 5 路分支表征 V3 加进 ECMA-402 后的"精度优先"语义。FormatJS 按此分支决议 `roundingType`,Go 端任何偏移都会破坏字节相等。

### 2.3 Digit Formatting SSOT

`internal/ecma402/numberformat.FormatNumericToString(d, DigitOptions)` 是 ECMA-402 digit formatting 的唯一运行时代码路径。NumberFormat 和 PluralRules 都消费它返回的 formatted string 与 rounded numeric value。

**MUST** 规则:

1. `DigitOptions` **必须**包含 `minimumIntegerDigits`、fraction digits、significant digits、`roundingIncrement`、`roundingMode`、`roundingPriority`、`trailingZeroDisplay` 的 resolved 状态;公共 formatter 包只负责把自己的 config 映射为该结构。
2. `FormatNumericToString` **必须**返回未本地化、未分组的 ASCII 十进制字符串,保留由 digit options 强制产生的 trailing zero。分组、本地数字符号、currency/unit/percent/compact 包装只能在 formatter 包完成。
3. `FormatNumericToString` **必须**同时返回 rounded numeric value,供 compact plural、range equality、PluralRules operands 使用。
4. NumberFormat 与 PluralRules **禁止**各自复制 fixed/significant/priority rounding 代码;任何舍入或补零修复必须落在 `internal/ecma402/numberformat` 并由两个 formatter 共享。
5. `FormatNumericToString` **禁止**通过 `float64`、`strconv.ParseFloat`、`math.Log10`、`math.Pow10` 做十进制舍入或指数缩放;这些操作必须通过 [SPEC 21 §Decimal API](./21-number-math.md#decimal-api)。

> **Why**: ECMA-402 PluralRules 复用 NumberFormat digit options 的语义;两包各写一套 rounding 会在 trailing-zero、`roundingPriority` 和 zero scale 上漂移。把 display string stage 收敛为一个函数,可以让 conformance fixture 同时约束 NumberFormat 输出和 PluralRules operands。

---

## 3. ResolvedOptions

```go
// 命名 string 类型(底层 kind 必须是 string,与 JS resolvedOptions 字符串字面值一一对应;
// 见 §1.4 第 4 条)。
type (
    Style               string  // "decimal" | "percent" | "currency" | "unit"
    CurrencyDisplay     string  // "code" | "symbol" | "narrowSymbol" | "name"
    CurrencySign        string  // "standard" | "accounting"
    UnitDisplay         string  // "short" | "narrow" | "long"
    UseGrouping         string  // "always" | "auto" | "min2" | "false"
    Notation            string  // "standard" | "scientific" | "engineering" | "compact"
    CompactDisplay      string  // "short" | "long"
    SignDisplay         string  // "auto" | "always" | "exceptZero" | "negative" | "never"
    RoundingMode        string  // "ceil" | "floor" | "expand" | "trunc" | "halfCeil" | "halfFloor" | "halfExpand" | "halfTrunc" | "halfEven"
    RoundingPriority    string  // "auto" | "morePrecision" | "lessPrecision"
    TrailingZeroDisplay string  // "auto" | "stripIfInteger"
    LocaleMatcher       string  // "lookup" | "best fit"
)

type ResolvedOptions struct {
    Locale                       locale.Locale
    NumberingSystem              string
    Style                        Style          // "decimal" | "percent" | "currency" | "unit"
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
    UseGrouping                  UseGrouping    // "always" | "auto" | "min2" | "false"
    Notation                     Notation       // "standard" | "scientific" | "engineering" | "compact"
    CompactDisplay               CompactDisplay // "short" | "long"
    SignDisplay                  SignDisplay
    RoundingIncrement            int
    RoundingMode                 RoundingMode
    RoundingPriority             RoundingPriority
    TrailingZeroDisplay          TrailingZeroDisplay
}
```

**MUST** 规则:

1. 字段顺序 **必须**与 ECMA-402 §15.4.5 spec 顺序一致(便于 conformance 测试逐字段对齐)。
2. `MinimumFractionDigits` / `MaximumFractionDigits` **必须**仅在 `roundingType ≠ significantDigits` 时呈现(零值 = 未呈现);`MinimumSignificantDigits` / `MaximumSignificantDigits` **必须**仅在 `roundingType ≠ fractionDigits` 时呈现。
3. `Locale` 字段 **必须**是 `New` 内部 `ResolveLocale` 后的解析结果(含 `-u-nu-...` 扩展),与输入 `loc` 可能不同。
4. **必须**返回值类型(非指针),保证调用方无法修改 formatter 内部状态。

> **Why**: ECMA-402 `resolvedOptions()` 是规范规定的可观测面,字段缺失或顺序错都被 conformance 测试视为失败。

> **Rejected**: 用 `map[string]any` 表达 ResolvedOptions —— 失去类型安全,且 messageformat-go 桥接需要二次断言。

---

## 4. 格式化主流程(PartitionNumberPattern)

`Format` 与 `FormatToParts` 共享同一主流程 `PartitionNumberPattern`,定义于 `formatjs/.../PartitionNumberPattern.ts` + `format_to_parts.ts`:

```text
1. x := ToIntlMathematicalValue(value)
2. exponent := 0
3. if notation=scientific|engineering|compact:
       exponent := ComputeExponent(nf, x, localeData)
       x        := x ÷ 10^exponent
4. n := internal/ecma402/numberformat.FormatNumericToString(x, nf.DigitOptions)
                                                 // String + Rounded
5. if x.specialValue:                           // NaN / ±Inf 走 symbol map
       formattedString := getNaN(localeData) or getInfinity(localeData)
   else:
       formattedString := PartitionDigitParts(nf, n, exponent, getNumberingSystem(nf), localeData)
6. 包装 sign / currency / unit / percent / compact-suffix / exponent-symbol
```

**MUST** 规则:

1. `ToIntlMathematicalValue` **必须**使用 [SPEC 21 §ToIntlMathematicalValue](./21-number-math.md#tointlmathematicalvalue) 实现;**禁止**通过 `fmt.Sprintf("%v", value)` 转换数值。
2. `FormatNumericToString` 输出的 `String` **必须**保留 trailing zero(由 `mnfd` 强制),供后续 OperandsRecord 计算 `v / w / f / t`(见 [SPEC 40 §Operands](./40-pluralrules.md#operands))。
3. `NaN / +Inf / -Inf` **必须**通过 `apd.Decimal.Form` 表达,**禁止** `math.IsNaN(float64(...))` 中转。
4. `PartitionDigitParts` 输出的 `[]Part` 元素 `Type` **必须**限定 ECMA-402 §15.5.1 + formatjs 扩展共 16 种枚举字符串:`integer | group | decimal | fraction | currency | percentSign | minusSign | plusSign | nan | infinity | unit | literal | exponentSeparator | exponentMinusSign | exponentInteger | compact | approximatelySign`(与 `.references/formatjs/packages/ecma402-abstract/types/number.ts` `NumberFormatPartTypes` 严格对齐;**禁止**使用 `exponentSymbol`,canonical 名称是 `exponentSeparator`;`approximatelySign` 只在 `FormatRange` 两端格式化结果相同时作为 part type 出现)。

> **Why**: 这是 conformance 字节相等的关键算子;任何步骤跳过都会被 `formatjs` `format_to_parts.test.ts` fixture 检出。

### 4.1 Compact Notation 与 PluralRules 契约

**MUST(与 [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract) 文字一致)**:

1. NumberFormat 在 `notation = compact` 路径下,**必须**用显示数的已舍入字符串和 compact exponent 选择 plural category:
   ```go
   ops := ecma402pr.GetOperands(formattedDisplayDecimal, exponent)
   rule, ok := plural.CardinalRule(localeTag)
   cat := rule(ops)
   ```
2. 传入参数 `(formattedDisplayDecimal string, exponent int)` 语义 **必须**是:
   - `formattedDisplayDecimal` = `FormatNumericToString` 输出的已舍入十进制字符串,**已经**除以 `10^exponent`(即"显示数"),并保留 trailing zero。
   - `exponent` = `ComputeExponent` 决定的 compact/scientific 指数。
3. **禁止** NumberFormat 自己解析 plural DSL 或持有公开 `pluralrules.PluralRules` 实例;规则函数只来自 `internal/cldr/plural`。
4. `internal/ecma402/pluralrules.GetOperands` 是 operand SSOT:从 `formattedDisplayDecimal` 派生 `n / i / v / w / f / t`,并把 `c / e` 设置为 `exponent`(见 [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract))。

> **Why**: 跨包契约。compact formatting 的稳定边界是"显示数 + compact exponent + generated CLDR rule";这保留 CLDR `c/e` 操作数,同时避免公开 formatter 之间相互依赖。
>
> **Rejected**: 公开 `SelectFormatted` 或让 NumberFormat 持有 `pluralrules.PluralRules` —— 这是 JS internal slot 的影子,不是 Go API。

### 4.2 Sign / Currency / Unit / Percent / Compact 包装

**MUST** 规则:

1. `signDisplay = "negative"` 与 `"exceptZero"` 在 ES 2024 加入,**必须**实现。
2. `currencyDisplay = "narrowSymbol"` **必须**回落到 `"symbol"` 当 CLDR 数据缺 narrow 形式。
3. `currencySign = "accounting"` **必须**使用 CLDR accounting pattern;负数子模式存在时由 pattern 消耗 minus sign,不存在时保留显式 sign part。
4. compact suffix 选择 **必须**先按 §4.1 决定 plural category,再查 CLDR `numbers.json` `decimalFormats.{short|long}.decimalFormat[length].decimal-format-pattern.<category>`;缺 category 时回落 `other`。
5. `useGrouping = "min2"` **必须**仅在整数部分 ≥ 5 位时插分组(对齐 formatjs `useGrouping` 实现)。

---

## 5. FormatRange / FormatRangeToParts <a id="5-formatrange--formatrangetoparts"></a>

**MUST** 规则:

1. `FormatRange(a, b)` **必须**实现 ECMA-402 §15.5.7 `FormatNumericRange`:先分别格式化两端,再调用 `CollapseNumberRange` 合并相同前缀/后缀。
2. `CollapseNumberRange` **必须**消费 `NumberFormatPart{Type, Value}` 序列,逐个判等;判等基准是 **per-package 字段**(`Type` 与 `Value` 都相等)。**禁止**抽象通用 `CollapseRange[T]` 与 DateTimeFormat 共享。
3. 两端 `Decimal` 比较 **必须**通过 [SPEC 21 §Decimal.Cmp](./21-number-math.md#decimal-cmp);**禁止**通过 `float64` 转换。
4. Range source **必须**限定为 ECMA-402 三值:`"startRange" | "shared" | "endRange"`;`approximatelySign` 是 part type,不是 source。
5. `FormatRangeFloat64` / `FormatRangeFloat64ToParts` **必须**在任一端是 `NaN` 时返回 `ErrInvalidValue` 包装错误,对齐 ECMA-402 `RangeError`。
6. `a > b` **必须不**被本地归一化、调换或加 `~`;按入参顺序格式化并 collapse range parts。
7. 当 `FormatNumeric(a) == FormatNumeric(b)` 时,输出 shared `approximatelySign` part + shared 数字 parts(例如最大 fraction digits 为 0 时 `1.1–1.2` 输出 `~1`)。

> **Why**: NumberFormatPart 与 DateTimeFormatPart 字段域不同(`unit | currency | percentSign | exponentInteger` vs `era | year | month | ...`),collapse 算法虽然结构同构(去前后缀)但工作在不同 part 域上;formatjs 也是各自实现。
>
> **Rejected**: 抽象通用 `CollapseRange[T Part]` 泛型函数 —— 多一层间接,且 `T` 的"等价"语义两包不同。

---

## 6. 输入类型支持

**MUST** 规则(对应 ECMA-402 §15.5.1):

1. 公共 hot path **禁止**接受 `any`;调用者必须选择 `FormatInt*`、`FormatUint*`、`FormatFloat64` 或 `FormatDecimal`。
2. `FormatDecimal` / `FormatDecimalToParts` / decimal range 方法接受 ECMA-402 `StringNumericLiteral`,如 `"1234.5"` / `"NaN"` / `"Infinity"`。
3. malformed decimal string **必须**返回 `ErrInvalidValue`;禁止静默回落到 `"NaN"`。
4. conformance fixture 可在包内保留未导出的 `formatValue(any)` adapter,但它不得出现在公开 API、README 或 root package 中。

> **Why**: `fmt.Sprintf("%v", float64)` 用 `%g` 格式,trailing-zero 与 ECMA-402 ToIntlMathematicalValue 不一致;在 hot path 走 Sprintf 是显著性能损失(每次 ~150 ns 分配)。

---

## 7. Internal Slot 与缓存

**MUST** 规则:

1. NumberFormat 内部状态 **必须**全部在 `New` 时计算并冻结;`New` 返回的 `*NumberFormat` 是不可变快照。
2. CLDR 数据指针(`numberSymbols / patterns / compactDecimalFormats`)**必须**在 `New` 时通过 `internal/cldr.NumbersFor(loc)` 一次拉取并保存到 `NumberFormat` 内部 slot;`Format` 路径**禁止**再调用 cldr accessor。
3. PluralRules 句柄 **必须**在 `New` 时 lazy 构造(仅 `notation = compact | scientific | engineering` 时构造);`Format` 路径**禁止**重新构造 PluralRules。
4. `apd.Context` **必须**作为不可变 baseline 持有;`Format` 调用时复制后修改 `Rounding` 字段(避免 race)。

> **Why**: `formatjs` 的 `internalSlot` 模式(每次 `format()` 都查 slot)在 Go 上对应"构造期物化、运行期只读"。这与 CLAUDE.md "constructor-eager / Format-no-error" 规则一致。

---

## 8. 错误模型

**MUST** 规则:

1. 构造期错误 **必须**是 `ErrInvalidOption` 的 wrapped 形式,消息含字段名 + 用户传入值 + locale,如 `numberformat: invalid currency "XYZ" for locale "en-US": invalid option`。
2. Sentinel **必须**重导出 `internal/ecma402.ErrInvalidOption`(SPEC 12);本包不另立 sentinel。
3. **禁止** `panic` 任何用户路径;`MustNew` 不存在(用户可在调用方包装)。
4. 运行时 fallback(NaN / Infinity / 字符串解析失败)**必须不**返回 error,直接出 fallback string。

```go
// 错误形态示例(签名)
err := fmt.Errorf("numberformat: invalid currency %q for locale %q: %w",
    code, loc.String(), ecma402.ErrInvalidOption)
```

---

## 9. 性能目标

**MUST** 规则:

1. `New(...)` 对常见 locale + decimal style **必须** ≤ 5 μs/op(P50)。
2. Cached `Format(int64)` decimal style **必须** ≤ 800 ns/op(对齐 [SPEC 71 §阈值](./71-benchmark.md#thresholds))。
3. `Format(int64)` 在 hot path **必须**零分配,除 string return value 外。

> **Why**: messageformat-go 在 `:number` 内调用,每条消息可能 N 次 `Format`;次于 1 μs 才能保留消息层 SLA。

---

## 10. Forbidden

- **禁止** 引入 `golang.org/x/text/message` 作为 NumberFormat 实现 —— 缺 currency/unit/compact notation,与 formatjs 字节相等不兼容。
- **禁止** 引入 `bojanz/currency` 作为货币数据源 —— 非 CLDR 直出,且自带 ISO 4217 历史数据与我们 CLDR 钉版冲突。
- **禁止** 在 `Format` / `FormatToParts` 路径返回 `error`(无效输入用 ECMA-402 fallback)。
- **禁止** `Format` 在 hot path 用 `fmt.Sprintf("%v", value)` —— trailing-zero 行为不对齐 + 每次 ~150 ns 分配。
- **禁止** NumberFormat 解析 plural DSL、复制 plural rule 表、或通过公开 `pluralrules.PluralRules` 实例选择 compact suffix;只能调用 `internal/ecma402/pluralrules.GetOperands` 并使用 `internal/cldr/plural` generated rule。
- **禁止** 把 `CollapseNumberRange` 抽成跨包泛型 —— Part 域不同。
- **禁止** 在 `Format` 路径调用 `internal/cldr.*` accessor;CLDR 数据必须在 `New` 时物化。
- **禁止** Builder 链式 API(`numberformat.NewBuilder().Currency("USD").Build()`);active scope 公开形态只有 `New(loc, Options{...})`。
- **禁止** 指针配置 API(`numberformat.New(loc, &Options{...})`)和 functional options;关闭 §8 Q2 后唯一公开配置形态是 typed `Options` 值。
- **禁止** 自研 `BigDecimal`;数学层全部走 [SPEC 21 §Decimal API](./21-number-math.md#decimal-api)(`apd/v3` 后端)。

---

## 11. Acceptance Criteria

- [ ] `formatjs/packages/intl-numberformat/tests/format_to_parts.test.ts` 全部 fixture 在 `numberformat/conformance_test.go` 通过(byte-equality)。
- [ ] `formatjs/packages/intl-numberformat/tests/notation-compact-zh-TW.test.ts` 通过(`format(98765) == "9.9萬"`)。
- [ ] `formatjs/packages/intl-numberformat/tests/format_range.test.ts` 全部 fixture 在 `FormatRange` 通过。
- [ ] `go test -race ./numberformat/...` 通过(含 `TestNumberFormat_ConcurrentFormat` 100 goroutine × 1000 调用)。
- [ ] `go vet ./numberformat/...` 干净。
- [ ] `New(...)` 对未知 currency 返回 `ErrInvalidOption` wrapped error,消息含 currency + locale。
- [ ] `FormatFloat64(math.NaN())` 返回 locale-specific NaN string,不返回 error,不 panic。
- [ ] `FormatDecimal("abc")` 返回 `ErrInvalidValue`。
- [ ] `ResolvedOptions().MinimumFractionDigits` 在 `Options{SignificantDigits: MinimumSignificantDigits(2)}` 单独传入时呈现零值(未呈现);整数零值与"已设置 0"通过内部 set bit 区分。
- [ ] `roundingPriority = "morePrecision" | "lessPrecision"` 在同时传入 fraction 与 significant digit 选项时可观测,不被当作 unsupported option。
- [ ] `Options{CurrencySign: AccountingCurrencySign}` 对 `en-US` 负 USD 输出 `($12.00)`。
- [ ] `Options{CompactDisplay: LongCompactDisplay}` 对 `en` `1500` + `MaximumFractionDigits(1)` 输出 `1.5 thousand`。
- [ ] compact plural 契约用例:`numberformat.New(loc-pl, numberformat.Options{Notation: numberformat.CompactNotation}).FormatInt(1500)` 在 `pl-PL` 下与 `formatjs` 输出一致(plural category `few` 后缀)。
- [ ] 选项管线步骤顺序通过 `internal/ecma402/numberformat.TestInitializeNumberFormat_StepOrder` 测试(每步打 trace,与 formatjs 调用序列对齐)。
- [ ] benchmark `BenchmarkNumberFormat_Cached_Decimal_Int64` ≤ 800 ns/op(SPEC 71 阈值)。
- [ ] benchmark `BenchmarkNumberFormat_New` ≤ 5 μs/op(常见 locale + decimal style,P50)。

---

## 12. References

### Primary

- `.references/formatjs/packages/intl-numberformat/` — 公开 API 形态、`format / formatToParts / formatRange / resolvedOptions` 行为
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/InitializeNumberFormat.ts` — 选项管线
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatDigitOptions.ts` — 5 路分支
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatUnitOptions.ts` — currency/unit 校验
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/PartitionNumberPattern.ts` — 主流程
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/format_to_parts.ts` — Compact 路径(`:262/304/316/331` 行 plural 调用)
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/CollapseNumberRange.ts` — Number-specific collapse
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/CurrencyDigits.ts` — currencyData 注入

- `.references/intl/intl.go` — translate-agent/intl(Go 先例,DateTimeFormat-only,NumberFormat 无先例)
- `.references/ext/src/ecma402/currency.c` — PHP/ICU 货币数据路径

### Project Cross-References

- [SPEC 12 §Abstract Ops](./12-abstract-operations.md) — shared validators / digit pipeline / `ErrInvalidOption`
- [SPEC 10 §Locale 结构](./10-locale.md#locale-结构) — `Locale.NumberingSystem()`
- [SPEC 21 §Decimal API](./21-number-math.md#decimal-api) — `Decimal` / `apd/v3` 后端 / 七种 rounding mode / RoundingPriority / RoundingIncrement / TrailingZeroDisplay 算法
- [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract) — compact plural operand 契约
- [SPEC 50 §Schema](./50-cldr-data.md#schema) — `numbers.json` / `currencies.json` / `units.json` 数据形态
- [SPEC 60](./60-facade.md) — root namespace ownership; root `intl.FormatNumber*` one-shot helpers are outside the long-term public surface.
- [SPEC 71 §阈值](./71-benchmark.md#thresholds) — 性能基线
