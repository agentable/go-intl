# SPEC 21 — Number Math & Decimal

> **Status:** Draft (2026-05-08)
> **Priority:** High(NumberFormat / PluralRules / DateTimeFormat 跨年份计算共用的数学层;阻塞 SPEC 20 / 40)
> **Authority:** 本 SPEC 是 `internal/decimal` 包、`Decimal` 类型、`ToIntlMathematicalValue`、九种 ECMA-402 rounding mode、`RoundingPriority` / `RoundingIncrement` / `TrailingZeroDisplay` 算法的 SSOT。**关闭 SPEC 00 §8 Q1(Decimal 后端选型)**。

---

## Overview

ECMA-402 §6.4 的 `ToIntlMathematicalValue` 把任意 JS 值(`Number` / `BigInt` / `String`)归一化为内部"数学值"概念,可表达 `NaN`、`±Infinity`、任意精度十进制有限数 `Finite(coeff, exp)`。该值随后被 `ToRawPrecision` / `ToRawFixed` / `ComputeExponent` / `ApplyUnsignedRoundingMode` 等 abstract op 消费。

formatjs 用 `@formatjs/bigdecimal` 自实现该数据结构;Go 不需要重写,因为 `cockroachdb/apd/v3`(IEEE 754-2008 GDA Decimal)的 `Form` 枚举(`Finite` / `Infinite` / `NaN` / `NaNSignaling`)与 formatjs `BigDecimal.specialValue` 一一对应,且原生提供 `Log10` / `Floor` / `Ceil` / `Quantize` / `Round` / 8 种 GDA rounding mode(覆盖 ECMA-402 全部 9 种)。

本 SPEC 决定:

1. 后端选 **`cockroachdb/apd/v3`**;**拒绝** `shopspring/decimal`(无 NaN/Inf 构造 panic)与 `ericlagergren/decimal`(v3 alpha 长期停滞)。
2. 在 `internal/decimal/` 提供 ECMA-402 抽象层窄接口的 `Decimal` 类型;apd 作为后端,但**不**通过公共 API 暴露 apd 类型(便于将来切换)。
3. 九种 ECMA-402 rounding mode → apd 的 8 种 GDA mode 映射 + `halfFloor` 自实现补丁。
4. `RoundingPriority` / `RoundingIncrement` / `TrailingZeroDisplay` 三个 ES 2025 V3 算法在本包内实现(SPEC 20 / 40 仅消费)。
5. 实现 `MathematicalValue` 接口(SPEC 12 §6 定义),把 `*Decimal` 注入抽象层。

本 SPEC **不**定义:NumberFormat 选项管线(SPEC 20)、PluralRules 规则编译(SPEC 40)、CLDR 货币精度数据(SPEC 50)、`MathematicalValue` interface 本身(SPEC 12 §6 已定义)。

---

## 1. 后端选型

### 1.1 决策:`cockroachdb/apd/v3`

```text
require github.com/cockroachdb/apd/v3 v3.x.x
```

| 候选 | 决策 | 关键理由 |
|------|------|---------|
| **`github.com/cockroachdb/apd/v3`** | ✅ 选定 | IEEE 754-2008 GDA;`Form` 枚举对应 formatjs `specialValue`;原生 `Log10` / `Quantize` / `Round`;`Context` 并发安全;Apache-2.0;持续维护(R09 §gh 信号:790 stars / 2026-03-23 last commit) |
| `shopspring/decimal` | ❌ 拒绝 | 无 NaN/Inf 表示;`NewFromString("NaN")` panic,与 CLAUDE.md "no panic in production" 红线冲突;缺 `Log10` 原生 |
| `ericlagergren/decimal` | ❌ 拒绝 | v3 长期 alpha;2024-04 后无新提交(R09 §gh);ABI 不稳定 |
| `math/big.Float` | ❌ 拒绝 | 二进制浮点;`0.1 + 0.2` 转十进制串与 formatjs 输出**字节不等**(违反 SPEC 70 conformance) |
| `math/big.Rat` | ❌ 拒绝 | 纯有理数;无 `Log10` 与定向舍入;ToRawPrecision 实现代价过高 |

> **Why `cockroachdb/apd/v3`**:
> 1. **GDA 与 formatjs 同构** —— `apd.Form` 枚举(`Finite=0` / `Infinite=1` / `NaN=2` / `NaNSignaling=3`)直接对应 formatjs `BigDecimal.specialValue`(`undefined` / `'POSITIVE_INFINITY'` / `'NEGATIVE_INFINITY'` / `'NaN'`);移植 `ToIntlMathematicalValue` 是 1:1 翻译。
> 2. **原生覆盖 ECMA-402 操作** —— `Log10` / `Floor` / `Ceil` / `Quantize` / `Round` 全部内置;不需要为 ToRawPrecision / ComputeExponent 自实现。
> 3. **并发安全** —— `apd.Context` 是值类型,可按 goroutine 复制;不像 `big.Float` 是共享可变状态。
> 4. **维护活跃** —— cockroachdb 主仓,持续提交,Apache-2.0 兼容 go-intl 许可证。
>
> **Rejected `shopspring/decimal` 详情**:
> - ❌ 无 NaN / +Inf / -Inf 表示;构造时直接 `panic("decimal: NaN not supported")`,违反 "no panic" 红线。
> - ❌ ECMA-402 `ToIntlMathematicalValue("NaN")` 必须能返回 NaN-shaped 值(formatjs `BigDecimal.NaN` 单例);用 `shopspring` 必须自己包一层"sentinel value",代码量与 apd 相当但失去 GDA 同构优势。
> - ❌ 缺 `Log10` 原生;ComputeExponent 需要自己实现 `floor(log10(|x|))`,精度边界难控。
>
> **Rejected `math/big.Float`**:
> - ❌ 二进制 IEEE-754 mantissa;`big.Float.Text('e', -1)` 输出 `0.30000000000000004` 而 formatjs 输出 `"0.3"`,**字节不等**。
> - ❌ ECMA-402 conformance 测试要求 `0.1 + 0.2 → "0.3"` 在 NumberFormat decimal style 下;`big.Float` 永远不能通过。
>
> **Rejected `ericlagergren/decimal`**:
> - ❌ v3 长期 alpha;2024-04 后停滞(R09 §gh);ABI 在 minor release 间反复变。
> - ❌ active scope 不能用一个不稳定后端作为基础设施。

### 1.2 包边界

```text
internal/decimal/
├── decimal.go        ← Decimal 类型(包 apd.Decimal)+ 构造 / 比较 / 算术
├── from.go           ← ToIntlMathematicalValue(value any) (Decimal, error)
├── rounding.go       ← 9 种 RoundingMode + ApplyUnsignedRoundingMode
├── quantize.go       ← Quantize / RoundingIncrement
├── log10.go          ← Log10Floor(用于 ComputeExponent)
├── trailing_zero.go  ← TrailingZeroDisplay 算法
├── priority.go       ← RoundingPriority(auto / morePrecision / lessPrecision)
└── math_value.go     ← 实现 SPEC 12 §6 MathematicalValue 接口
```

> **Why 私有包**:用户**不**直接构造 Decimal;`internal/` 强制隐藏 apd 依赖,让"切换 Decimal 后端"成为单点修改。
>
> **Why 不暴露 `apd.Decimal`**:`numberformat.Options{RoundingMode: ...}` 不应把后端泄漏到公共 API;一旦未来切到 `ericlagergren/decimal/v4`,公共 API 破坏。统一包一层 `numberformat.RoundingMode` 类型,值域 verbatim ECMA-402。
>
> **Why 文件按 abstract op 切分**:每个文件对应 ECMA-402 一节,便于 fixture 移植(formatjs `bigdecimal/tests/` 与 `ecma402-abstract/NumberFormat/tests/`)。

---

## 2. Decimal 类型

<a id="decimal-类型"></a>
<a id="decimal-api"></a>

### 2.1 类型定义

```go
// internal/decimal/decimal.go(签名)
package decimal

import "github.com/cockroachdb/apd/v3"

// Decimal 是 ECMA-402 §6.4 ToIntlMathematicalValue 的归一化数值表示。
// 包装 apd.Decimal;不直接暴露 apd 类型,便于将来切换后端。
//
// 表示语义:
//   - Form == Finite:     value = (-1)^Negative × Coeff × 10^Exponent
//   - Form == Infinite:   ±∞,Negative 决定符号
//   - Form == NaN:        Quiet NaN(ECMA-402 ToIntlMathematicalValue 不区分 quiet/signaling)
//   - Form == NaNSignaling: 内部使用,不会从 ToIntlMathematicalValue 输出
type Decimal struct {
    inner apd.Decimal // 不导出:封装 apd 实现细节
}

// Form 是数值的种类。verbatim 镜像 apd.Form 但作为公共 API。
type Form uint8

const (
    Finite       Form = iota // 普通有限值
    Infinite                 // ±∞
    NaN                      // 非数值
    NaNSignaling             // signaling NaN(内部)
)

// Form 返回 d 的种类。
func (d Decimal) Form() Form

// Sign 返回 -1 / 0 / +1(NaN 返回 0,Inf 按 Negative 字段)。
func (d Decimal) Sign() int

// IsZero 仅当 d 为 +0 或 -0(Form=Finite, Coeff=0)时为 true。
func (d Decimal) IsZero() bool

// IsNaN / IsInf / IsFinite 是 Form 检查的语法糖。
func (d Decimal) IsNaN() bool
func (d Decimal) IsInf() bool
func (d Decimal) IsFinite() bool

// Exponent 返回十进制指数(Finite 形态);其他 form 返回 0。
func (d Decimal) Exponent() int32

// Coeff 返回十进制系数的绝对值字符串(big.Int.String());NaN/Inf 返回空串。
func (d Decimal) Coeff() string

// Negative 返回符号位(Inf / Finite 都有意义)。
func (d Decimal) Negative() bool
```

> **Why 字段全部通过 method 暴露**:`apd.Decimal` 公开字段(`Negative` / `Coefficient` / `Exponent` / `Form`)允许直接 mutation;我们的 `Decimal` 用值语义 + 只读方法,客户端不能 mutate 中间状态。
>
> **Why `Form` 重新定义而非 `apd.Form` alias**:`type Form = apd.Form` 在 godoc 显示后端类型名,泄漏实现;独立常量 + 转换函数(internal)更清晰。
>
> **Rejected `Decimal` 是 `*apd.Decimal` 别名**:违反值语义偏好;`decimal.New(...)` 返回 `*Decimal` 强制堆分配,heap allocation 在 hot path 是显著开销(R03 §performance)。

### 2.2 构造

```go
// internal/decimal/decimal.go(签名)

// Sentinel 实例(包级单例,不可变)。
var (
    Zero        = Decimal{} // form=Finite, coeff=0, exp=0
    NaNValue    Decimal     // form=NaN
    PosInfinity Decimal     // form=Infinite, negative=false
    NegInfinity Decimal     // form=Infinite, negative=true
)

// New 从 (negative, coeff, exp) 构造 Finite Decimal。
// coeff 是非负 big.Int(系数绝对值)。
func New(negative bool, coeff *big.Int, exp int32) Decimal

// FromInt64 从 int64 构造 Finite Decimal(零分配快路径)。
func FromInt64(n int64) Decimal

// FromFloat64 从 float64 构造 Decimal。
//   - NaN  → NaNValue
//   - ±Inf → PosInfinity / NegInfinity
//   - 其他 → 经 strconv.FormatFloat('g', -1) → ParseString 转换
//          (避免 IEEE-754 二进制误差)
func FromFloat64(f float64) Decimal

// ParseString 从 ECMA-402 StringNumericLiteral 文法解析(§6.4.1)。
//   - "NaN" → NaNValue
//   - "Infinity" / "+Infinity" / "-Infinity" → ±PosInfinity
//   - 其他十进制 / 十六进制 / 二进制 / 八进制 → Finite
//   - 不合法 → 返回 ErrInvalidDecimal
func ParseString(s string) (Decimal, error)
```

> **Why `FromInt64` 单独**:NumberFormat hot path 大多接受 `int` / `int64`;经 `apd.New(int64, exp)` 直构造比 `ParseString(strconv.FormatInt(...))` 快 ~10×(R03 §performance)。
>
> **Why `FromFloat64` 走字符串中转而非 `apd.NewFromFloat`**:ECMA-402 `ToIntlMathematicalValue(float64)` 要求 spec §6.4 "convert via shortest round-trip string";`apd.NewFromFloat` 直接读 float64 二进制位,会保留 `0.30000000000000004` 这类伪精度。

### 2.3 算术

<a id="decimal-cmp"></a>

```go
// internal/decimal/decimal.go(签名)

// Add / Sub / Mul / Div 返回新 Decimal(不修改 receiver)。
// NaN 传染:任一操作数为 NaN 时返回 NaNValue。
// Inf 算术按 IEEE-754:Inf - Inf = NaN;0 × Inf = NaN;Inf / Inf = NaN。
func (d Decimal) Add(other Decimal) Decimal
func (d Decimal) Sub(other Decimal) Decimal
func (d Decimal) Mul(other Decimal) Decimal
func (d Decimal) Div(other Decimal) (Decimal, error)  // 0/0 返回 NaN;非零/0 返回 ±Inf

// Neg 返回 -d(NaN 仍为 NaN)。
func (d Decimal) Neg() Decimal

// Abs 返回 |d|。
func (d Decimal) Abs() Decimal

// Cmp 比较 d 与 other:
//   -1 = d <  other
//    0 = d == other(NaN-aware:任一为 NaN 返回 -1?见下)
//   +1 = d >  other
//
// NaN 处理:任一操作数为 NaN 时返回 ErrNaNComparison(IEEE-754 推荐);
// 调用方需先 IsNaN 检查再 Cmp。
func (d Decimal) Cmp(other Decimal) (int, error)

// Equal 是 NaN-aware 等价:NaN == NaN 返回 false(IEEE-754),其他走 Cmp。
// Equal 不返回错误(NaN 直接 false)。
func (d Decimal) Equal(other Decimal) bool

// PowerOf10 返回 10^n;n 可负;NaN 输入返回 NaN。
func PowerOf10(n int) Decimal
```

> **Why `Cmp` 返回 error 而 `Equal` 不**:
> - `Cmp` 是 3-valued (`<` / `=` / `>`);NaN 没有 3-valued 答案,IEEE-754 推荐"unordered"——Go 用 error 表达。
> - `Equal` 是 2-valued bool;NaN ≠ NaN(IEEE-754),自然映射到 false。
>
> **Why 不实现 Go `==` (Decimal 不可比较)**:`apd.Decimal` 内部含 `big.Int`(非 comparable);嵌入后 Decimal 也不是。强制走 `Equal()`。

### 2.4 字符串往返

```go
// internal/decimal/decimal.go(签名)

// String 返回 ECMA-402 StringNumericLiteral 兼容输出。
//   - NaN  → "NaN"
//   - ±Inf → "Infinity" / "-Infinity"
//   - Finite → 不带 trailing 0 的最短可逆表示(apd.Decimal.Text('G'))
func (d Decimal) String() string

// Text 返回指定格式输出(底层走 apd.Decimal.Text)。
//   format 'e' / 'E' / 'f' / 'g' / 'G',与 strconv 一致。
//   prec 是小数位数(format='f')或有效位数(format='g')。
func (d Decimal) Text(format byte, prec int) string
```

---

<a id="tointlmathematicalvalue"></a>

## 3. ToIntlMathematicalValue

### 3.1 ECMA-402 §6.4 算法

```text
ToIntlMathematicalValue(value):
  1. primValue := ToPrimitive(value, hint=number)
  2. 若 primValue 是 BigInt:返回 (BigInt 值的精确 Decimal,Form=Finite)
  3. 若 primValue 是 String:
       a. str := StringToNumber(primValue)  # ES §7.1.4.1 数字字面量
       b. 若 str = NaN:返回 NaN
       c. 若 str = ±Infinity:返回 ±Infinity
       d. 否则返回 Finite(精确十进制)
  4. 若 primValue 是 Number:
       a. 若 IsNaN(primValue):返回 NaN
       b. 若 ±Infinity:返回 ±Infinity
       c. 若 -0:返回 -0(保留符号位)
       d. 否则:把 primValue 经"最短 round-trip 字符串"转 Decimal
  5. 否则:抛 TypeError
```

### 3.2 Go 签名

```go
// internal/decimal/from.go(签名)

// ToIntlMathematicalValue 实现 ECMA-402 §6.4。
//
// 接受输入(由 NumberFormat / PluralRules 在边界处统一调用):
//   - int / int8 / int16 / int32 / int64
//   - uint / uint8 / uint16 / uint32 / uint64
//   - float32 / float64
//   - *big.Int / *big.Float / *big.Rat
//   - string(StringNumericLiteral)
//   - Decimal(直通)
//   - fmt.Stringer(取 String() 后递归)
//
// 不接受:nil / bool / 复合类型 / 任意 struct。
func ToIntlMathematicalValue(value any) (Decimal, error)
```

调用示例:

```go
d, _ := decimal.ToIntlMathematicalValue(int64(98765))   // Finite, coeff=98765, exp=0
d, _ = decimal.ToIntlMathematicalValue("3.14")          // Finite, coeff=314, exp=-2
d, _ = decimal.ToIntlMathematicalValue(math.Inf(+1))    // PosInfinity
d, _ = decimal.ToIntlMathematicalValue(math.NaN())      // NaNValue
```

### 3.3 性能目标

| 输入类型 | 目标 |
|---------|------|
| `int64` 快路径 | ≤ 50 ns / op,**零堆分配**(`FromInt64` 无 allocator) |
| `float64`(非 NaN/Inf) | ≤ 250 ns / op |
| `string`("3.14"-级) | ≤ 500 ns / op |
| `*big.Int` | ≤ 800 ns / op |

> **Why int64 ≤ 50 ns**:NumberFormat hot path 在 messageformat-go 单元测试每 ms 调用 1000+ 次;`Format` 总开销目标 < 800 ns(SPEC 71);ToIntlMathematicalValue 占 5-10% 预算。
>
> **Why 零堆分配**:`Decimal` 是值类型;`FromInt64(int64)` 内部用 `apd.Decimal.SetInt64` 写到栈上 receiver,不触发 `big.Int.New`。

---

<a id="rounding-modes"></a>

## 4. Rounding Modes

### 4.1 九种 ECMA-402 模式

ECMA-402 §15.5.5 定义九种 rounding mode;与 apd 对照:

| ECMA-402 名 | apd Rounding 常量 | 说明 |
|------------|------------------|------|
| `ceil` | `apd.RoundCeiling` | 向 +∞ |
| `floor` | `apd.RoundFloor` | 向 -∞ |
| `expand` | `apd.RoundUp` | 远离零 |
| `trunc` | `apd.RoundDown` | 向零 |
| `halfCeil` | (无 apd 直接对应) | 半数情况向 +∞;**自实现** |
| `halfFloor` | (无 apd 直接对应) | 半数情况向 -∞;**自实现** |
| `halfExpand` | `apd.RoundHalfUp` | 半数情况远离零(默认) |
| `halfTrunc` | `apd.RoundHalfDown` | 半数情况向零 |
| `halfEven` | `apd.RoundHalfEven` | 半数情况到偶数(银行家) |

> **Why apd 缺 `halfCeil` / `halfFloor`**:apd 实现 GDA 标准 8 种 mode;ECMA-402 V3(2022)新增 `halfCeil` / `halfFloor` 两种(用于金融四舍五入到正方向 / 负方向)。
>
> **Why 不向 apd 提 PR**:维护方已表示这是 ECMA-402 specific 不属于 GDA;且 apd `Rounder` 接口允许我们扩展。

### 4.2 Go 类型

```go
// internal/decimal/rounding.go(签名)

// RoundingMode 是 ECMA-402 §15.5.5 的九种 mode 之一。
// String() 输出 spec verbatim 名(用于 ResolvedOptions)。
type RoundingMode int

const (
    RoundCeil       RoundingMode = iota // "ceil"
    RoundFloor                          // "floor"
    RoundExpand                         // "expand"
    RoundTrunc                          // "trunc"
    RoundHalfCeil                       // "halfCeil"   ← 自实现
    RoundHalfFloor                      // "halfFloor"  ← 自实现
    RoundHalfExpand                     // "halfExpand" (默认)
    RoundHalfTrunc                      // "halfTrunc"
    RoundHalfEven                       // "halfEven"
)

func (m RoundingMode) String() string

// ParseRoundingMode 把 ECMA-402 字符串转 RoundingMode。
// 大小写敏感(spec verbatim)。
func ParseRoundingMode(s string) (RoundingMode, error)
```

### 4.3 ApplyUnsignedRoundingMode

ECMA-402 §15.5.7 `ApplyUnsignedRoundingMode(x, r1, r2, unsignedRoundingMode)`:把 `x ∈ (r1, r2)` 按 mode 决定 round 到 r1 还是 r2。

```go
// internal/decimal/rounding.go(签名)

// ApplyUnsignedRoundingMode 实现 ECMA-402 §15.5.7。
//   x  : 待舍入十进制
//   r1 : 下界(向零方向)
//   r2 : 上界(远离零方向)
//   m  : RoundingMode(已转 unsigned —— 见 GetUnsignedRoundingMode)
// 返回 r1 或 r2(其一)。
func ApplyUnsignedRoundingMode(x, r1, r2 Decimal, m RoundingMode) Decimal

// GetUnsignedRoundingMode(m, sign) 把 signed mode 转 unsigned(spec §15.5.6)。
//   - ceil  + 负号 → halfDown 风格的 unsigned
//   - floor + 正号 → halfDown 风格
//   - 其余 mode 不变
func GetUnsignedRoundingMode(m RoundingMode, isNegative bool) RoundingMode
```

> **Why 按 spec 名 `ApplyUnsignedRoundingMode` verbatim**:formatjs `ecma402-abstract/NumberFormat/ApplyUnsignedRoundingMode.ts` 1:1 实现;移植成本最低。
>
> **Why 不直接调 `apd.Decimal.Quantize` + `apd.Rounder`**:apd 的 Rounder 接口不暴露"我现在在哪两个邻居 r1 / r2 之间"的中间值,无法实现 `halfCeil` / `halfFloor` 的"看符号选边"逻辑;必须自实现 `ApplyUnsignedRoundingMode`,内部用 apd 算 `r1` / `r2` 然后自己选。

### 4.4 RoundingPriority

```go
// internal/decimal/priority.go(签名)

// RoundingPriority 是 ES 2025 V3 字段;决定 minSD/mxSD 与 minFD/mxFD
// 同时设置时的优先级。
type RoundingPriority int

const (
    PriorityAuto         RoundingPriority = iota // "auto"          (默认)
    PriorityMorePrecision                        // "morePrecision"
    PriorityLessPrecision                        // "lessPrecision"
)

// ApplyRoundingPriority 在 SetNumberFormatDigitOptions 内调用,决定
// roundingType ∈ {fractionDigits, significantDigits, morePrecision, lessPrecision}。
//
// 输入:
//   hasSD = mnsd|mxsd 至少一个被设置
//   hasFD = mnfd|mxfd 至少一个被设置
//   priority = PriorityAuto / MorePrecision / LessPrecision
// 返回 RoundingType(供 PartitionNumberPattern 路由)。
func ApplyRoundingPriority(hasSD, hasFD bool, priority RoundingPriority) RoundingType

type RoundingType int
const (
    RoundingFractionDigits    RoundingType = iota
    RoundingSignificantDigits
    RoundingMorePrecision
    RoundingLessPrecision
    RoundingCompact           // notation=compact 默认
)
```

### 4.5 RoundingIncrement

```go
// internal/decimal/quantize.go(签名)

// ValidRoundingIncrements 是 ECMA-402 §15.5.4 允许的 17 个值。
// 其他值 → ErrInvalidRoundingIncrement。
var ValidRoundingIncrements = []int{
    1, 2, 5, 10, 20, 25, 50,
    100, 200, 250, 500,
    1000, 2000, 2500, 5000,
}

// IsValidRoundingIncrement 校验。
func IsValidRoundingIncrement(inc int) bool

// QuantizeToIncrement(x, increment, exp, mode) 把 x 舍入到最近的
// (increment × 10^exp) 倍数。
//   x          : 待舍入值
//   increment  : 必须 ∈ ValidRoundingIncrements
//   exp        : 量级(由 mxfd 决定)
//   mode       : RoundingMode
func QuantizeToIncrement(x Decimal, increment int, exp int32, mode RoundingMode) Decimal
```

> **Why 静态校验 `ValidRoundingIncrements`**:ECMA-402 §15.5.4 显式列举,任何其他值是 RangeError;在 `numberformat.New` 边界一次性 `IsValidRoundingIncrement` 检查,避免 `Format` 时再判。
>
> **Why `QuantizeToIncrement` 内部不调 `apd.Quantize`**:apd 的 Quantize 是"量化到 10^k",本算子是"量化到 (increment × 10^exp) 的整数倍";需要先 `x / increment` 再 quantize 再乘回 increment。

### 4.6 TrailingZeroDisplay

```go
// internal/decimal/trailing_zero.go(签名)

type TrailingZeroDisplay int
const (
    TrailingZeroAuto           TrailingZeroDisplay = iota // "auto"           (默认,保留)
    TrailingZeroStripIfInteger                            // "stripIfInteger"  (整数时去尾零)
)

// ApplyTrailingZeroDisplay 在 ToRawFixed / ToRawPrecision 输出后调用。
//   formatted : 已舍入的字符串(如 "3.00" 或 "3.14")
//   isInteger : 数学值是否整数
//   display   : 用户选项
// 返回处理后的字符串(可能截断 trailing zero)。
func ApplyTrailingZeroDisplay(formatted string, isInteger bool, display TrailingZeroDisplay) string
```

> **Why 字符串后处理而非数值层面**:trailing zero 是显示概念,不是数学概念;`Decimal{coeff=3, exp=-2}` 与 `Decimal{coeff=300, exp=-4}` 数学上相等但 trailing-zero 表现不同。在 ToRawFixed 把数值固化为字符串后处理最自然(formatjs 同方案)。

---

## 5. Log10Floor & ComputeExponent

### 5.1 用途

`ComputeExponent`(ECMA-402 §15.5.3)在 scientific / engineering / compact notation 中决定数值的指数:

```text
ComputeExponent(nf, x):
  if x == 0: return 0
  magnitude := floor(log10(|x|))
  exponent  := ComputeExponentForMagnitude(nf, magnitude)
  if exponent < 0:
      mv := x × 10^(-exponent)
  else:
      mv := x × 10^exponent  # 错!应是 x ÷ 10^exponent
  return exponent
```

需要 `floor(log10(|x|))` 的精确十进制结果(不能用 `math.Log10(float64(...))`,精度损失)。

### 5.2 Go 签名

```go
// internal/decimal/log10.go(签名)

// Log10Floor 返回 floor(log10(|x|)) 的精确整数结果。
// x 必须 Form == Finite 且 != 0;否则返回 ErrLog10Domain。
//
// 内部用 apd.BaseContext.Precision = 200(覆盖 ECMA-402 mxfd 上界 100)
// 调用 apd.Log10 然后 Floor;结果为 int32(永远在 ECMA-402 数值域内)。
func Log10Floor(x Decimal) (int32, error)
```

> **Why `apd.Log10` + Precision=200**:R03 §1.3 实测;ECMA-402 mxfd 上界 100,数值最大约 10^100;Log10 内部精度 200 足以保证 `floor` 边界正确。

---

## 6. MathematicalValue 接口实现

### 6.1 SPEC 12 §6 接口

[SPEC 12 §6](./12-abstract-operations.md#6-math-value-boundary) 定义抽象层接口(在 `internal/ecma402/types/math.go`):

```go
// 由 SPEC 12 §6 SSOT 定义,本 SPEC 仅实现。
package types
type MathematicalValue interface {
    // (具体方法集由 SPEC 12 拥有;本 SPEC 不重复)
}
```

### 6.2 实现绑定

```go
// internal/decimal/math_value.go(签名)
package decimal

import "github.com/agentable/go-intl/internal/ecma402/types"

// 编译期断言:Decimal 实现 SPEC 12 §6 MathematicalValue 接口。
var _ types.MathematicalValue = Decimal{}

// 实现细节略;每个方法从 d.Form() / d.Coeff() / d.Exponent() 派生。
```

> **Why 在本包注入实现而非在 SPEC 12 包**:SPEC 12 不能 `import internal/decimal`(SPEC 12 §1.4 forbidden);实现绑定必须在依赖图下游(本 SPEC)。
>
> **Why 编译期断言 `var _`**:接口签名变化时立即编译失败,避免运行时 nil interface panic;Go 习惯用法。

---

## 7. 错误处理

### 7.1 哨兵

```go
// internal/decimal/errors.go(签名)
package decimal

import "errors"

var (
    // ErrInvalidDecimal: ParseString 输入非法十进制字面量。
    ErrInvalidDecimal = errors.New("decimal: invalid numeric literal")

    // ErrInvalidRoundingIncrement: 不在 ValidRoundingIncrements 列表中。
    ErrInvalidRoundingIncrement = errors.New("decimal: invalid rounding increment")

    // ErrNaNComparison: Cmp 时任一操作数为 NaN。
    ErrNaNComparison = errors.New("decimal: NaN in comparison")

    // ErrLog10Domain: Log10Floor 输入 0 或非 Finite。
    ErrLog10Domain = errors.New("decimal: log10 of zero or non-finite")
)
```

### 7.2 与 numberformat / pluralrules 错误协调

`numberformat.ErrInvalidOption` 在边界包装本包错误:

```go
// numberformat/options.go(SPEC 20 内的 wrap 模式)
if !decimal.IsValidRoundingIncrement(inc) {
    return fmt.Errorf("numberformat: %w: roundingIncrement=%d", decimal.ErrInvalidRoundingIncrement, inc)
}
```

> **Why 不在本包返回 wrap 后的错误**:SSOT —— 本包的错误是 raw sentinel;wrapping 由 SPEC 20 / 40 在边界处加 context。

---

## 8. Forbidden

### 8.1 ❌ 不要直接暴露 `apd.Decimal` 类型

```go
// ❌ 错误:public API 泄漏 apd 类型
package numberformat
func WithRoundingMode(m apd.Rounder) Option

// ✅ 正确:用 decimal.RoundingMode 抽象
package numberformat
func WithRoundingMode(m numberformat.RoundingMode) Option  // numberformat.RoundingMode = decimal.RoundingMode
```

> **Why**:切换 Decimal 后端时所有 public API 跟着破坏;`internal/decimal` 一层隔离允许后端透明替换。

### 8.2 ❌ 不要用 `shopspring/decimal`

```go
// ❌ 错误:无 NaN/Inf,违反 "no panic" 红线
import "github.com/shopspring/decimal"
d := decimal.NewFromString("NaN") // panic!

// ✅ 正确:apd 原生支持
d, _ := decimal.ParseString("NaN")  // d.IsNaN() == true
```

### 8.3 ❌ 不要在 hot path 用 `fmt.Sprintf` 转数值

```go
// ❌ 错误:每次 Format 调用 ~150 ns 分配
s := fmt.Sprintf("%v", anyValue)
d, _ := decimal.ParseString(s)

// ✅ 正确:类型 switch + 专用构造函数
d, err := decimal.ToIntlMathematicalValue(anyValue)
```

> **Why**:`fmt.Sprintf("%v", float64)` 用 `%g` 格式;ECMA-402 ToIntlMathematicalValue 要求"最短 round-trip";二者不一致 + Sprintf 在 hot path 是显著分配。

### 8.4 ❌ 不要用 `math/big.Float` 作为 ECMA-402 数值

```go
// ❌ 错误:0.1 + 0.2 字节不等
f := new(big.Float).SetFloat64(0.1)
f.Add(f, big.NewFloat(0.2))
fmt.Println(f.Text('g', -1))  // "0.30000000000000004"

// ✅ 正确:十进制 Decimal
a, _ := decimal.ToIntlMathematicalValue("0.1")
b, _ := decimal.ToIntlMathematicalValue("0.2")
sum := a.Add(b)
fmt.Println(sum.String())  // "0.3"
```

> **Why**:formatjs 用十进制 BigDecimal;`big.Float` 是 IEEE-754 二进制,conformance 测试必然失败。

### 8.5 ❌ 不要在 `Cmp` 用 NaN 时静默返回 0

```go
// ❌ 错误:违反 IEEE-754 "NaN unordered"
func (d Decimal) Cmp(other Decimal) int {
    if d.IsNaN() || other.IsNaN() { return 0 }  // 假装相等
    // ...
}

// ✅ 正确:返回 error 强制调用方处理
func (d Decimal) Cmp(other Decimal) (int, error) {
    if d.IsNaN() || other.IsNaN() { return 0, ErrNaNComparison }
    // ...
}
```

### 8.6 ❌ 不要在 `Decimal` 上实现 Go `==`

```go
// ❌ 错误:Decimal 嵌入 apd.Decimal 含 big.Int(非 comparable)
if d1 == d2 { /* compile error 或行为未定义 */ }

// ✅ 正确:走 Equal()
if d1.Equal(d2) { /* ... */ }
```

### 8.7 ❌ 不要把 `RoundingPriority` 算法放在 SPEC 20

```go
// ❌ 错误:SPEC 20 重复实现 priority 决策
package numberformat
func setNumberFormatDigitOptions(...) {
    if priority == "morePrecision" { /* ... */ }  // 重复
}

// ✅ 正确:调用 SPEC 21 §4.4 ApplyRoundingPriority
package numberformat
rt := decimal.ApplyRoundingPriority(hasSD, hasFD, opts.RoundingPriority)
```

> **Why**:`RoundingPriority` 决策算法是数学层关注点;SPEC 20 是消费者。SSOT 在本 SPEC。

### 8.8 ❌ 不要在 `internal/decimal` 包导入 `internal/ecma402` 实现

```go
// ❌ 错误:循环依赖(internal/ecma402 §1.4 也禁止反向)
import "github.com/agentable/go-intl/internal/ecma402"
func Foo() { ecma402.ToIntlMathematicalValue(...) }

// ✅ 正确:只导入 internal/ecma402/types(纯类型包)
import "github.com/agentable/go-intl/internal/ecma402/types"
var _ types.MathematicalValue = Decimal{}
```

> **Why**:SPEC 12 §1.4 与 SPEC 21 §1.2 闭合 —— 两个包对方向严格分层;只通过 `types/` 子包(纯类型,无逻辑)交换接口。

---

## 9. Acceptance Criteria

### 后端

- [ ] `go.mod` 包含 `github.com/cockroachdb/apd/v3`,**不**包含 `github.com/shopspring/decimal` 或 `github.com/ericlagergren/decimal`。
- [ ] `internal/decimal/` 子目录下分文件 `decimal.go` / `from.go` / `rounding.go` / `quantize.go` / `log10.go` / `trailing_zero.go` / `priority.go` / `math_value.go` / `errors.go`。
- [ ] `internal/decimal` 包的公共 API **不**暴露任何 `apd.*` 类型(`grep -r "apd\." | grep -v "internal/decimal/" | grep -v "_test.go"` 返回空)。

### Decimal 类型

- [ ] `Decimal` 是值类型(struct);`Decimal{}` 是 `Form=Finite, Coeff=0, Exp=0`(等价 +0)。
- [ ] `Form` 枚举值 `Finite=0` / `Infinite=1` / `NaN=2` / `NaNSignaling=3`(可由 apd 转换)。
- [ ] 包级单例 `Zero` / `NaNValue` / `PosInfinity` / `NegInfinity` 不可被 mutate。
- [ ] `IsNaN` / `IsInf` / `IsFinite` 与 `Form()` 方法语义一致。

### 构造

- [ ] `New(negative, coeff, exp) Decimal` 接受 `*big.Int` 系数。
- [ ] `FromInt64(int64) Decimal` 在 `BenchmarkFromInt64` 下 ≤ 50 ns/op,**0 allocs/op**。
- [ ] `FromFloat64(NaN)` 返回 `NaNValue`(IsNaN()=true)。
- [ ] `FromFloat64(±Inf)` 返回 `±PosInfinity`。
- [ ] `FromFloat64(0.1).Add(FromFloat64(0.2)).String() == "0.3"`(十进制不偏离)。
- [ ] `ParseString("NaN")` / `"Infinity"` / `"-Infinity"` 各自正确。
- [ ] `ParseString("foo")` 返回 `ErrInvalidDecimal` wrap 错误。

### 算术

- [ ] `Add` / `Sub` / `Mul` / `Div` 不修改 receiver;返回新 Decimal。
- [ ] NaN 传染:任一操作数 NaN → 结果 NaN。
- [ ] Inf 算术按 IEEE-754:`Inf - Inf = NaN`、`0 × Inf = NaN`、`Inf / Inf = NaN`。
- [ ] `Div` 在 `0 / 0` 返回 NaN,`非零 / 0` 返回 `±Inf`,**不** panic。
- [ ] `Cmp(NaN, _) → (0, ErrNaNComparison)`。
- [ ] `Equal(NaN, NaN) == false`(IEEE-754)。
- [ ] `PowerOf10(n) == 10^n` 对 `-100 ≤ n ≤ 100` 全部精确。

### ToIntlMathematicalValue

- [ ] `ToIntlMathematicalValue(int64(98765)).String() == "98765"`。
- [ ] `ToIntlMathematicalValue("3.14").Exponent() == -2` 且 `Coeff() == "314"`。
- [ ] `ToIntlMathematicalValue(math.NaN()).IsNaN() == true`。
- [ ] `ToIntlMathematicalValue(math.Inf(+1)).IsInf() == true` 且 `Negative() == false`。
- [ ] `ToIntlMathematicalValue(nil)` 返回 `ErrInvalidDecimal`。
- [ ] `BenchmarkToIntlMathematicalValue_Int64` ≤ 50 ns/op,0 allocs。
- [ ] `BenchmarkToIntlMathematicalValue_String_3p14` ≤ 500 ns/op。
- [ ] formatjs `bigdecimal/tests/` 全部 fixture 在 `internal/decimal/from_test.go` 通过。

### Rounding Modes

- [ ] `RoundingMode` 9 个常量;`String()` 输出 spec verbatim 名(`"halfCeil"` 不是 `"half-ceil"`)。
- [ ] `ParseRoundingMode("halfExpand") == RoundHalfExpand`,`ParseRoundingMode("HALFEXPAND")` 失败(大小写敏感)。
- [ ] `ApplyUnsignedRoundingMode` 在 `halfCeil` / `halfFloor` 下结果与 formatjs `ApplyUnsignedRoundingMode.test.ts` byte-equal。
- [ ] `GetUnsignedRoundingMode` 与 spec §15.5.6 verbatim 表对齐。
- [ ] formatjs `ecma402-abstract/NumberFormat/tests/ApplyUnsignedRoundingMode.test.ts` 全部 fixture 通过。

### RoundingPriority

- [ ] `ApplyRoundingPriority` 5 个分支(MorePrecision / LessPrecision / hasSD / hasFD / Compact / 默认 fractionDigits)与 formatjs `SetNumberFormatDigitOptions.ts` 对齐。
- [ ] `RoundingType` 5 值(FractionDigits / SignificantDigits / MorePrecision / LessPrecision / Compact)。

### RoundingIncrement

- [ ] `ValidRoundingIncrements` verbatim 17 值。
- [ ] `IsValidRoundingIncrement(3) == false`,`IsValidRoundingIncrement(50) == true`。
- [ ] `QuantizeToIncrement(123.456, 25, -2, RoundHalfExpand).String() == "123.5"`(125/25=5;round half expand;25×0.05=1.25?见 fixture)。
- [ ] formatjs `ecma402-abstract/NumberFormat/tests/Quantize.test.ts` fixture 通过(若存在)。

### TrailingZeroDisplay

- [ ] `ApplyTrailingZeroDisplay("3.00", true, TrailingZeroStripIfInteger) == "3"`。
- [ ] `ApplyTrailingZeroDisplay("3.14", false, TrailingZeroStripIfInteger) == "3.14"`(非整数不 strip)。
- [ ] `ApplyTrailingZeroDisplay("3.00", true, TrailingZeroAuto) == "3.00"`(auto 保留)。

### Log10Floor

- [ ] `Log10Floor(FromInt64(98765)) == 4`(98765 ∈ [10^4, 10^5))。
- [ ] `Log10Floor(FromInt64(0))` 返回 `ErrLog10Domain`。
- [ ] `Log10Floor(NaNValue)` 返回 `ErrLog10Domain`。

### MathematicalValue 接口

- [ ] `var _ types.MathematicalValue = Decimal{}` 编译通过。
- [ ] `internal/decimal` **不**导入 `internal/ecma402`(只导入 `internal/ecma402/types`);`grep -r "internal/ecma402\"" internal/decimal/` 应为空。

### 错误

- [ ] `errors.Is(err, ErrInvalidDecimal)` 在 `ParseString` 失败下为 true。
- [ ] `errors.Is(err, ErrInvalidRoundingIncrement)` 在 `IsValidRoundingIncrement(false)` 后调用方 wrap 时为 true。
- [ ] 包内**无** `panic` 调用。

### 测试

- [ ] formatjs `bigdecimal/tests/` 全部 fixture 移植到 `internal/decimal/testdata/` 并通过。
- [ ] formatjs `ecma402-abstract/NumberFormat/tests/{ApplyUnsignedRoundingMode,SetNumberFormatDigitOptions}.test.ts` fixture 通过。
- [ ] 所有测试用 `t.Parallel()`。
- [ ] `BenchmarkFromInt64` / `BenchmarkToIntlMathematicalValue_Int64` 在 `task test:bench` 跑,记录到 SPEC 71 §benchmark。
- [ ] 至少 1 个 `Example*` 函数演示 `ToIntlMathematicalValue` + `String`。

---

## References

### Specification

- [ECMA-402 §6.4 — Number Format](https://tc39.es/ecma402/#sec-numbers)(`ToIntlMathematicalValue`)
- [ECMA-402 §15.5 — NumberFormat Digit Options](https://tc39.es/ecma402/#sec-numberformat-digitoptions)
- [ECMA-402 §15.5.5 — Rounding Modes](https://tc39.es/ecma402/#sec-rounding-modes)
- [ECMA-402 §15.5.6 — GetUnsignedRoundingMode](https://tc39.es/ecma402/#sec-getunsignedroundingmode)
- [ECMA-402 §15.5.7 — ApplyUnsignedRoundingMode](https://tc39.es/ecma402/#sec-applyunsignedroundingmode)
- [IEEE 754-2008 §General Decimal Arithmetic](https://standards.ieee.org/standard/754-2008.html)

### Reference implementations

- `.references/formatjs/packages/bigdecimal/src/index.ts` —— `BigDecimal{mantissa, exponent, specialValue}` 与 `add` / `sub` / `mul` / `div` / `quantize` / `log10`
- `.references/formatjs/packages/bigdecimal/tests/` —— fixture
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatDigitOptions.ts` —— `RoundingPriority` / `RoundingIncrement` / `roundingType` 五路分支
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ApplyUnsignedRoundingMode.ts` —— `halfCeil` / `halfFloor` 自实现路径
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/GetUnsignedRoundingMode.ts`
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ToIntlMathematicalValue.ts` —— spec §6.4 实现
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ToRawFixed.ts` / `ToRawPrecision.ts` —— RoundingMode 消费者
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ComputeExponent.ts` —— `Log10Floor` 消费者

### Library survey

- `github.com/cockroachdb/apd/v3` —— 选定后端;Apache-2.0;`Form` / `Context` / `Log10` / `Quantize` / `Round` / 8 种 GDA mode
- `github.com/shopspring/decimal` —— ❌ 拒绝;无 NaN/Inf;构造 panic
- `github.com/ericlagergren/decimal` —— ❌ 拒绝;v3 alpha 长期停滞(2024-04 后无新提交)
- `math/big.Float` —— ❌ 拒绝;二进制 IEEE-754;conformance byte-equality 不通过
- `math/big.Rat` —— ❌ 拒绝;无 Log10 / 定向舍入

### Cross-SPEC

- [SPEC 00 §8 Q1 — Decimal 后端选型](./00-vision-and-scope.md#8-open-questions)(本 SPEC 关闭)
- [SPEC 12 §6 — MathematicalValue 接口](./12-abstract-operations.md#6-math-value-boundary) —— 本 SPEC `Decimal` 实现该接口
- [SPEC 12 §1 — Package Layout(forbidden import)](./12-abstract-operations.md#1-package-layout) —— 本 SPEC §1.2 与之闭合
- [SPEC 20 §Format Pipeline](./20-numberformat.md) —— 本 SPEC 是其数学层
- [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract) —— compact notation 通过 `Decimal` 与格式化字符串构造 OperandsRecord
- [SPEC 50 §6 — Data Access API](./50-cldr-data.md#6-data-access-api) —— 货币默认精度数据(`CurrencyDigits`)由 SPEC 50 注入,本 SPEC 不重复定义
- [SPEC 71 §Benchmark](./71-benchmark.md) —— 本 SPEC §3.3 / §9 性能目标对应

### Research

- `.research/R03-numberformat.md` §1 —— Decimal 选型横向比较(Form / NaN / Log10 / Quantize / 并发 / 字节相等)
- `.research/R09-dependencies.md` §4.1 —— `cockroachdb/apd/v3` 维护信号(stars / commits / license)、拒绝清单印证

---

> 本 SPEC 是 `internal/decimal` 与 ECMA-402 数学层的 SSOT。新增 ECMA-402 rounding mode(spec 罕见)或 `apd/v3` 升级触发本 SPEC 修订;后端切换(若发生)在本 SPEC §1.1 决策表更新。
