# SPEC 21 — Number Math & Decimal

> **Status:** Draft (2026-05-08)
> **Priority:** High(NumberFormat / PluralRules / DateTimeFormat shared math layer across years; blocking SPEC 20 / 40)
> **Authority:** ECMA-402 number-format algorithms are normative. This SPEC documents the current Go contract for the `internal/decimal` package, `Decimal` type, `ToIntlMathematicalValue`, nine ECMA-402 rounding modes, `RoundingPriority` / `RoundingIncrement` / `TrailingZeroDisplay` algorithms.

---

## Overview

ECMA-402 §6.4's `ToIntlMathematicalValue` normalizes any JS value (`Number` / `BigInt` / `String`) to an internal "mathematical value" concept, which can express `NaN`, `±Infinity`, and arbitrary precision decimal finite number `Finite(coeff, exp)`. This value is then consumed by abstract ops such as `ToRawPrecision` / `ToRawFixed` / `ComputeExponent` / `ApplyUnsignedRoundingMode`.

Generated reference uses `@formatjs/bigdecimal` to self-implement this data structure; Go does not need to rewrite, because the `Form` enumeration (`Finite` / `Infinite` / `NaN` / `NaNSignaling`) of `cockroachdb/apd/v3` (IEEE 754-2008 GDA Decimal) corresponds one-to-one with generated-reference `BigDecimal.specialValue`, and natively provides `Log10` / `Floor` / `Ceil` / `Quantize` / `Round` / 8 GDA rounding modes (covering all 9 ECMA-402 modes).

This SPEC decides:

1. Backend selection **`cockroachdb/apd/v3`**; **rejection** `shopspring/decimal` (no NaN/Inf construct panic) and `ericlagergren/decimal` (v3 alpha long-term stagnation).
2. The `Decimal` type that provides the ECMA-402 abstraction layer narrow interface in `internal/decimal/`; apd as a backend, but does not expose the apd type through the public API (to facilitate future switching).
3. Nine ECMA-402 rounding modes → 8 GDA mode mappings of apd + `halfFloor` self-implemented patch.
4. `RoundingPriority` / `RoundingIncrement` / `TrailingZeroDisplay` Three ES 2025 V3 algorithms are implemented in this package (SPEC 20 / 40 only for consumption).
5. Implement the `MathematicalValue` interface (defined in SPEC 12 §6) and inject `*Decimal` into the abstraction layer.

This SPEC **not** defines: NumberFormat option pipeline (SPEC 20), PluralRules rule compilation (SPEC 40), CLDR currency precision data (SPEC 50), `MathematicalValue` interface itself (SPEC 12 §6 defined).

---

## 1. Backend selection

### 1.1 Decision:`cockroachdb/apd/v3`

```text
require github.com/cockroachdb/apd/v3 v3.x.x
```

| Candidate | Decision | Key Reasons |
|------|------|---------|
| **`github.com/cockroachdb/apd/v3`** | ✅ Selected | IEEE 754-2008 GDA; `Form` enumeration corresponds to generated-reference `specialValue`; native `Log10` / `Quantize` / `Round`; `Context` concurrency safety; Apache-2.0; maintained upstream. |
| `shopspring/decimal` | ❌ Reject | No NaN/Inf representation; `NewFromString("NaN")` panic, conflicts with CLAUDE.md "no panic in production" red line; missing `Log10` native |
| `ericlagergren/decimal` | ❌ Rejected | v3 remains long-term alpha; ABI stability is weaker for a formatter math core. |
| `math/big.Float` | ❌ Rejected | Binary floating point; `0.1 + 0.2` converted to decimal string and reference output **bytes are not equal** (violates SPEC 70 conformance) |
| `math/big.Rat` | ❌ Reject | Purely rational; no `Log10` with directional rounding; ToRawPrecision is too expensive to implement |

> **Why `cockroachdb/apd/v3`**:
> 1. **GDA is isomorphic to Generated reference** - `apd.Form` enumeration (`Finite=0` / `Infinite=1` / `NaN=2` / `NaNSignaling=3`) directly corresponds to generated-reference `BigDecimal.specialValue`(`undefined` / `'POSITIVE_INFINITY'` / `'NEGATIVE_INFINITY'` / `'NaN'`); ported `ToIntlMathematicalValue` is 1:1 Translate.
> 2. **Native override ECMA-402 operations** - `Log10` / `Floor` / `Ceil` / `Quantize` / `Round` all built-in; no need to self-implement for ToRawPrecision / ComputeExponent.
> 3. **Concurrency safety** - `apd.Context` is a value type and can be copied by goroutine; unlike `big.Float`, it is a shared mutable state.
> 4. **Active maintenance** - cockroachdb main repository, continuous submission, Apache-2.0 compatible with go-intl license.
>
> **Rejected `shopspring/decimal` Details**:
> - ❌ No NaN / +Inf / -Inf representation; direct `panic("decimal: NaN not supported")` during construction, violating the "no panic" red line.
> - ❌ ECMA-402 `ToIntlMathematicalValue("NaN")` must be able to return NaN-shaped values (generated-reference `BigDecimal.NaN` singleton); when using `shopspring`, you must wrap a layer of "sentinel value" yourself. The amount of code is equivalent to apd but it loses the isomorphic advantage of GDA.
> - ❌ Missing `Log10` native; ComputeExponent needs to implement `floor(log10(|x|))` by itself, and the accuracy boundary is difficult to control.
>
> **Rejected `math/big.Float`**:
> - ❌ Binary IEEE-754 mantissa;`big.Float.Text('e', -1)` outputs `0.30000000000000004` while reference outputs `"0.3"`, **bytes are not equal**.
> - ❌ ECMA-402 conformance test requires `0.1 + 0.2 → "0.3"` under NumberFormat decimal style; `big.Float` will never pass.
>
> **Rejected `ericlagergren/decimal`**:
> - ❌ v3 long-term alpha with repeated ABI churn between minor releases.
> - ❌ active scope cannot use an unstable backend as infrastructure.

### 1.2 Packet Boundary

```text
internal/decimal/
├── decimal.go ← Decimal type (package apd.Decimal) + construction / comparison / arithmetic
├── from.go           ← ToIntlMathematicalValue(value any) (Decimal, error)
├── rounding.go ← 9 types of RoundingMode + ApplyUnsignedRoundingMode
├── quantize.go       ← Quantize / RoundingIncrement
├── log10.go ← Log10Floor (for ComputeExponent)
├── trailing_zero.go ← TrailingZeroDisplay algorithm
├── priority.go       ← RoundingPriority(auto / morePrecision / lessPrecision)
└── math_value.go ← Implements SPEC 12 §6 MathematicalValue interface
```

> **Why private package**: Users **cannot** directly construct Decimal; `internal/` forces the apd dependency to be hidden, making "switching the Decimal backend" a single point of modification.
>
> **Why not expose `apd.Decimal`**:`numberformat.Options{RoundingMode: ...}` should not leak the backend to the public API; once the future switches to `ericlagergren/decimal/v4`, the public API is broken. Unified package layer `numberformat.RoundingMode` type, value range verbatim ECMA-402.
>
> **Why files are split according to abstract op**: Each file corresponds to an ECMA-402 section to facilitate fixture transplantation (generated-reference `bigdecimal/tests/` and `ecma402-abstract/NumberFormat/tests/`).

---

## 2. Decimal type

<a id="decimal-type"></a>
<a id="decimal-api"></a>

### 2.1 Type definition

```go
// internal/decimal/decimal.go(signature)
package decimal

import "github.com/cockroachdb/apd/v3"

// Decimal is the normalized numerical representation of ECMA-402 §6.4 ToIntlMathematicalValue.
// Packaging apd.Decimal; does not directly expose the apd type to facilitate backend switching in the future.
//
// Representation semantics:
//   - Form == Finite:     value = (-1)^Negative × Coeff × 10^Exponent
// - Form == Infinite: ±∞,Negative determining symbol
// - Form == NaN: Quiet NaN(ECMA-402 ToIntlMathematicalValue does not distinguish between quiet/signaling)
// - Form == NaNSignaling: used internally, will not be output from ToIntlMathematicalValue
type Decimal struct {
inner apd.Decimal // Do not export: encapsulate apd implementation details
}

// Form is the type of value. verbatim mirrors apd.Form but as a public API.
type Form uint8

const (
Finite Form = iota // Ordinary finite value
    Infinite                 // ±∞
NaN // Not a numeric value
NaNSignaling // signaling NaN (internal)
)

// Form returns the type of d.
func (d Decimal) Form() Form

// Sign returns -1 / 0 / +1 (NaN returns 0, Inf according to the Negative field).
func (d Decimal) Sign() int

// IsZero is true only when d is +0 or -0 (Form=Finite, Coeff=0).
func (d Decimal) IsZero() bool

// IsNaN / IsInf / IsFinite are syntactic sugar for Form checks.
func (d Decimal) IsNaN() bool
func (d Decimal) IsInf() bool
func (d Decimal) IsFinite() bool

// Exponent returns decimal exponent (Finite form); other forms return 0.
func (d Decimal) Exponent() int32

// Coeff returns the absolute value string of the decimal coefficient (big.Int.String()); NaN/Inf returns an empty string.
func (d Decimal) Coeff() string

// Negative returns the sign bit (Inf / Finite are both meaningful).
func (d Decimal) Negative() bool
```

> **Why fields are all exposed through methods**: `apd.Decimal` public fields (`Negative` / `Coefficient` / `Exponent` / `Form`) allow direct mutation; our `Decimal` uses value semantics + read-only method, and the client cannot mutate the intermediate state.
>
> **Why `Form` is redefined instead of `apd.Form` alias**:`type Form = apd.Form` shows backend type name in godoc, leaking implementation; independent constant + conversion function (internal) is more clear.
>
> **Rejected `Decimal` is an alias for `*apd.Decimal`**: Violates value semantic preference; `decimal.New(...)` returns `*Decimal` and forces heap allocation, which is unacceptable on the formatter hot path.

### 2.2 Construction

```go
// internal/decimal/decimal.go (signature)

// Sentinel instance (package-level singleton, immutable).
var (
    Zero        = Decimal{} // form=Finite, coeff=0, exp=0
    NaNValue    Decimal     // form=NaN
    PosInfinity Decimal     // form=Infinite, negative=false
    NegInfinity Decimal     // form=Infinite, negative=true
)

// New Constructs Finite Decimal from (negative, coeff, exp).
// coeff is non-negative big.Int (absolute value of coefficient).
func New(negative bool, coeff *big.Int, exp int32) Decimal

// FromInt64 Construct Finite Decimal from int64 (zero allocation fast path).
func FromInt64(n int64) Decimal

// FromUint64 constructs a finite Decimal from uint64 without a float bridge.
func FromUint64(n uint64) Decimal

// FromBigInt constructs a finite Decimal from a BigInt bridge value. Nil means 0,
// and the input is copied into Decimal-owned storage.
func FromBigInt(n *big.Int) Decimal

// FromFloat64 Constructs Decimal from float64.
//   - NaN  → NaNValue
//   - ±Inf → PosInfinity / NegInfinity
// - Others → converted by strconv.FormatFloat('g', -1) → ParseString
// (Avoid IEEE-754 binary errors)
func FromFloat64(f float64) Decimal

// ParseString is parsed from the ECMA-402 StringNumericLiteral grammar (§6.4.1).
//   - "NaN" → NaNValue
//   - "Infinity" / "+Infinity" / "-Infinity" → ±PosInfinity
// - Other decimal / hexadecimal / binary / octal → Finite
// - Illegal → Return ErrInvalidDecimal
func ParseString(s string) (Decimal, error)
```

> **Why integer constructors live in `internal/decimal`**: NumberFormat and PluralRules both expose typed `Int`, `Uint`, and `BigInt` bridges. The exact conversion rule belongs to the decimal owner, so signed, unsigned, nil BigInt, and copied BigInt storage cannot drift between formatter packages.
>
> **Why `FromFloat64` uses string conversion instead of `apd.NewFromFloat`**: ECMA-402 `ToIntlMathematicalValue(float64)` requires spec §6.4 "convert via shortest round-trip string"; `apd.NewFromFloat` directly reads float64 binary bits, and will retain pseudo-precision such as `0.30000000000000004`.

### 2.3 Arithmetic

<a id="decimal-cmp"></a>

```go
// internal/decimal/decimal.go(signature)

// Add / Sub / Mul / Div returns a new Decimal (without modifying the receiver).
// NaN contagion: Return NaNValue when any operand is NaN.
// Inf arithmetic follows IEEE-754: Inf - Inf = NaN; 0 × Inf = NaN; Inf / Inf = NaN.
func (d Decimal) Add(other Decimal) Decimal
func (d Decimal) Sub(other Decimal) Decimal
func (d Decimal) Mul(other Decimal) Decimal
func (d Decimal) Div(other Decimal) (Decimal, error) // 0/0 returns NaN; non-zero/0 returns ±Inf

// Neg returns -d(NaN is still NaN).
func (d Decimal) Neg() Decimal

// Abs returns |d|.
func (d Decimal) Abs() Decimal

// Cmp compares d with other:
//   -1 = d <  other
// 0 = d == other(NaN-aware: If any one is NaN, return -1? See below)
//   +1 = d >  other
//
// NaN processing: When any operand is NaN, ErrNaNComparison (IEEE-754 recommendation) is returned;
// The caller needs to check IsNaN first and then Cmp.
func (d Decimal) Cmp(other Decimal) (int, error)

// Equal is NaN-aware equivalent: NaN == NaN returns false (IEEE-754), otherwise go to Cmp.
// Equal does not return an error (NaN is directly false).
func (d Decimal) Equal(other Decimal) bool

// PowerOf10 returns 10^n; n can be negative; NaN input returns NaN.
func PowerOf10(n int) Decimal
```

> **Why `Cmp` returns error but `Equal` does not**:
> - `Cmp` is 3-valued (`<` / `=` / `>`); NaN has no 3-valued answer, IEEE-754 recommends "unordered" - Go expresses it with error.
> - `Equal` is a 2-valued bool;NaN ≠ NaN (IEEE-754), which naturally maps to false.
>
> **Why does not implement Go `==` (Decimal is not comparable)**: `apd.Decimal` contains `big.Int` (non-comparable) internally; after embedding, Decimal is not either. Forced to go `Equal()`.

### 2.4 String round trip

```go
// internal/decimal/decimal.go(signature)

// String returns ECMA-402 StringNumericLiteral compatible output.
//   - NaN  → "NaN"
//   - ±Inf → "Infinity" / "-Infinity"
// - Finite → the shortest reversible representation without trailing 0 (apd.Decimal.Text('G'))
func (d Decimal) String() string

// Text returns the specified format output (the bottom layer uses apd.Decimal.Text).
// format 'e' / 'E' / 'f' / 'g' / 'G', consistent with strconv.
// prec is the number of decimal places (format='f') or the number of significant digits (format='g').
func (d Decimal) Text(format byte, prec int) string
```

---

<a id="tointlmathematicalvalue"></a>

## 3. ToIntlMathematicalValue

### 3.1 ECMA-402 §6.4 Algorithm

```text
ToIntlMathematicalValue(value):
  1. primValue := ToPrimitive(value, hint=number)
2. If primValue is BigInt: return (accurate Decimal of BigInt value,Form=Finite)
3. If primValue is String:
a. str := StringToNumber(primValue) # ES §7.1.4.1 Number literal
b. If str = NaN: return NaN
c. If str = ±Infinity: return ±Infinity
d. Otherwise, return Finite (exact decimal)
4. If primValue is Number:
a. If IsNaN(primValue): return NaN
b. If ±Infinity: return ±Infinity
c. If -0: return -0 (sign bit reserved)
d. Otherwise: Convert primValue to Decimal via "the shortest round-trip string"
5. Otherwise: throw TypeError
```

### 3.2 Go signature

```go
// internal/decimal/from.go(signature)

// ToIntlMathematicalValue implements ECMA-402 §6.4.
//
// Accept input (called uniformly by NumberFormat / PluralRules at the boundary):
//   - int / int8 / int16 / int32 / int64
//   - uint / uint8 / uint16 / uint32 / uint64
//   - float32 / float64
//   - *big.Int / *big.Float / *big.Rat
//   - string(StringNumericLiteral)
// - Decimal (through)
// - fmt.Stringer (recurse after taking String())
//
// Not accepted: nil / bool / composite type / any struct.
func ToIntlMathematicalValue(value any) (Decimal, error)
```

Call example:

```go
d, _ := decimal.ToIntlMathematicalValue(int64(98765))   // Finite, coeff=98765, exp=0
d, _ = decimal.ToIntlMathematicalValue("3.14")          // Finite, coeff=314, exp=-2
d, _ = decimal.ToIntlMathematicalValue(math.Inf(+1))    // PosInfinity
d, _ = decimal.ToIntlMathematicalValue(math.NaN())      // NaNValue
```

### 3.3 Performance Target

| Input type | telemetry target |
|---------|------|
| `int64` fast path | track with `BenchmarkFromInt64` and allocation reporting |
| `float64`(not NaN/Inf) | track with decimal conversion benchmarks |
| `string`("3.14"-level) | track with parse benchmarks |
| `*big.Int` | track with constructor benchmarks |

> **Why int64 telemetry**:NumberFormat hot path is called 1000+ times per ms in the messageformat-go unit test; ToIntlMathematicalValue should stay visible in reports without becoming a standalone merge blocker (SPEC 71).
>
> **Why zero heap allocation**: `Decimal` is a value type; `FromInt64(int64)` internally uses `apd.Decimal.SetInt64` to write to the stack receiver and does not trigger `big.Int.New`.

---

<a id="rounding-modes"></a>

## 4. Rounding Modes

### 4.1 Nine ECMA-402 modes

ECMA-402 §15.5.5 defines nine rounding modes; compared with apd:

| ECMA-402 Name | apd Rounding Constant | Description |
|------------|------------------|------|
| `ceil` | `apd.RoundCeiling` | Towards +∞ |
| `floor` | `apd.RoundFloor` | Towards -∞ |
| `expand` | `apd.RoundUp` | Far from zero |
| `trunc` | `apd.RoundDown` | Towards zero |
| `halfCeil` | (no apd direct correspondence) | Half of the cases go to +∞;**Self-implementation** |
| `halfFloor` | (no apd direct correspondence) | Half of the cases go to -∞;**Self-implementation** |
| `halfExpand` | `apd.RoundHalfUp` | Far from zero half the time (default) |
| `halfTrunc` | `apd.RoundHalfDown` | Towards zero half of the time |
| `halfEven` | `apd.RoundHalfEven` | Half of the time to an even number (Banker) |

> **Why apd is missing `halfCeil` / `halfFloor`**: apd implements 8 modes of the GDA standard; ECMA-402 V3(2022) adds `halfCeil` / `halfFloor` (for financial rounding to the positive direction/negative direction).
>
> **Why not submit PR to apd**: The maintainer has stated that this is ECMA-402 specific and does not belong to GDA; and the apd `Rounder` interface allows us to extend it.

### 4.2 Go Type

```go
// internal/decimal/rounding.go(signature)

// RoundingMode is one of the nine modes specified in ECMA-402 §15.5.5.
// String() output spec verbatim name (for ResolvedOptions).
type RoundingMode int

const (
    RoundCeil       RoundingMode = iota // "ceil"
    RoundFloor                          // "floor"
    RoundExpand                         // "expand"
    RoundTrunc                          // "trunc"
RoundHalfCeil // "halfCeil" ← Self-implemented
RoundHalfFloor // "halfFloor" ← Self-implemented
RoundHalfExpand // "halfExpand" (default)
    RoundHalfTrunc                      // "halfTrunc"
    RoundHalfEven                       // "halfEven"
)

func (m RoundingMode) String() string

// ParseRoundingMode converts ECMA-402 string to RoundingMode.
// Case sensitive (spec verbatim).
func ParseRoundingMode(s string) (RoundingMode, error)
```

### 4.3 ApplyUnsignedRoundingMode

ECMA-402 §15.5.7 `ApplyUnsignedRoundingMode(x, r1, r2, unsignedRoundingMode)`: Press `x ∈ (r1, r2)` to determine whether to round to r1 or r2.

```go
// internal/decimal/rounding.go(signature)

// ApplyUnsignedRoundingMode implements ECMA-402 §15.5.7.
// x: decimal to be rounded
// r1: lower bound (towards zero)
// r2: upper bound (away from zero direction)
// m: RoundingMode (converted to unsigned - see GetUnsignedRoundingMode)
// Return r1 or r2 (one).
func ApplyUnsignedRoundingMode(x, r1, r2 Decimal, m RoundingMode) Decimal

// GetUnsignedRoundingMode(m, sign) converts signed mode to unsigned(spec §15.5.6).
// - ceil + minus sign → halfDown style unsigned
// - floor + positive sign → halfDown style
// - The rest of the modes remain unchanged
func GetUnsignedRoundingMode(m RoundingMode, isNegative bool) RoundingMode
```

> **Why According to spec name `ApplyUnsignedRoundingMode` verbatim**:generated-reference `ecma402-abstract/NumberFormat/ApplyUnsignedRoundingMode.ts` 1:1 implementation; the transplantation cost is the lowest.
>
> **Why not directly adjust `apd.Decimal.Quantize` + `apd.Rounder`**: The Rounder interface of apd does not expose the intermediate value of "between the two neighbors r1 / r2 I am now", and cannot implement the "select the edge based on the symbol" logic of `halfCeil` / `halfFloor`; must implement `ApplyUnsignedRoundingMode` by itself, and use apd internally to calculate `r1` / `r2` Then choose yourself.

### 4.4 RoundingPriority

```go
// internal/decimal/priority.go(signature)

// RoundingPriority is an ES 2025 V3 field; determines minSD/mxSD and minFD/mxFD
// The priority when setting simultaneously.
type RoundingPriority int

const (
PriorityAuto RoundingPriority = iota // "auto" (default)
    PriorityMorePrecision                        // "morePrecision"
    PriorityLessPrecision                        // "lessPrecision"
)

// ApplyRoundingPriority is called within SetNumberFormatDigitOptions to determine
// roundingType ∈ {fractionDigits, significantDigits, morePrecision, lessPrecision}.
//
// Input:
// hasSD = mnsd|mxsd At least one is set
// hasFD = mnfd|mxfd At least one is set
//   priority = PriorityAuto / MorePrecision / LessPrecision
// Return RoundingType (for PartitionNumberPattern routing).
func ApplyRoundingPriority(hasSD, hasFD bool, priority RoundingPriority) RoundingType

type RoundingType int
const (
    RoundingFractionDigits    RoundingType = iota
    RoundingSignificantDigits
    RoundingMorePrecision
    RoundingLessPrecision
RoundingCompact // notation=compact default
)
```

### 4.5 RoundingIncrement

```go
// internal/decimal/quantize.go(signature)

// ValidRoundingIncrements are among the 17 values allowed by ECMA-402 §15.5.4.
// Other values → ErrInvalidRoundingIncrement.
var ValidRoundingIncrements = []int{
    1, 2, 5, 10, 20, 25, 50,
    100, 200, 250, 500,
    1000, 2000, 2500, 5000,
}

// IsValidRoundingIncrement verification.
func IsValidRoundingIncrement(inc int) bool

// QuantizeToIncrement(x, increment, exp, mode) rounds x to the nearest
// (increment × 10^exp) multiple.
// x: value to be rounded
// increment: must ∈ ValidRoundingIncrements
// exp: magnitude (determined by mxfd)
//   mode      : RoundingMode
func QuantizeToIncrement(x Decimal, increment int, exp int32, mode RoundingMode) Decimal
```

> **Why static check `ValidRoundingIncrements`**: ECMA-402 §15.5.4 Explicit enumeration, any other value is RangeError; one-time `IsValidRoundingIncrement` check at the `numberformat.New` boundary to avoid re-evaluation when `Format` is used.
>
> **Why `QuantizeToIncrement` is not adjusted internally `apd.Quantize`**: Quantize of apd is "quantized to 10^k", and this operator is "quantized to an integer multiple of (increment × 10^exp)"; you need to `x / increment` first, then quantize and then multiply back to increment.

### 4.6 TrailingZeroDisplay

```go
// internal/decimal/trailing_zero.go(signature)

type TrailingZeroDisplay int
const (
TrailingZeroAuto TrailingZeroDisplay = iota // "auto" (default, reserved)
TrailingZeroStripIfInteger // "stripIfInteger" (remove trailing zeros when integer)
)

// ApplyTrailingZeroDisplay is called after ToRawFixed / ToRawPrecision output.
// formatted: rounded string (such as "3.00" or "3.14")
// isInteger: Whether the mathematical value is an integer
// display: user options
// Return the processed string (possibly truncated to trailing zero).
func ApplyTrailingZeroDisplay(formatted string, isInteger bool, display TrailingZeroDisplay) string
```

> **Why string post-processing rather than numerical level**: trailing zero is a display concept, not a mathematical concept; `Decimal{coeff=3, exp=-2}` and `Decimal{coeff=300, exp=-4}` are mathematically equivalent but trailing-zero behaves differently. It is most natural to process the value after solidifying it into a string in ToRawFixed (Generated reference has the same solution).

---

## 5. Log10Floor & ComputeExponent

### 5.1 Purpose

`ComputeExponent`(ECMA-402 §15.5.3) determines the exponent of a value in scientific / engineering / compact notation:

```text
ComputeExponent(nf, x):
  if x == 0: return 0
  magnitude := floor(log10(|x|))
  exponent := ComputeExponentForMagnitude(nf, magnitude)
  if exponent < 0:
      mv := x × 10^(-exponent)
  else:
mv := x × 10^exponent # Wrong! It should be x ÷ 10^exponent
  return exponent
```

The exact decimal result of `floor(log10(|x|))` is required (cannot use `math.Log10(float64(...))`, precision loss).

### 5.2 Go signature

```go
// internal/decimal/log10.go (signature)

// Log10Floor returns the exact integer result of floor(log10(|x|)).
// x must be Form == Finite and != 0; otherwise ErrLog10Domain will be returned.
//
// Internally used apd.BaseContext.Precision = 200 (covering the ECMA-402 mxfd upper bound of 100)
// Call apd.Log10 then Floor; the result is int32 (always within the ECMA-402 numeric domain).
func Log10Floor(x Decimal) (int32, error)
```

> **Why `apd.Log10` + Precision=200**: ECMA-402's `mxfd` upper bound is 100, so the maximum relevant value is about 10^100; Log10 internal precision 200 is enough to keep the `floor` boundary correct.

---

## 6. MathematicalValue interface implementation

### 6.1 SPEC 12 §6 Interface

[SPEC 12 §6](./12-abstract-operations.md#6-math-value-boundary) defines the abstraction layer interface (in `internal/ecma402/types.go`):

```go
// Documented by SPEC 12 §6, this SPEC is implementation only.
package ecma402
type MathematicalValue interface {
// (The specific method set is owned by SPEC 12; this SPEC is not repeated)
}
```

### 6.2 Implement binding

```go
// internal/ecma402/decimal_test.go (signature)
package ecma402

import "github.com/agentable/go-intl/internal/decimal"

// Compile-time assertion: Decimal implements the SPEC 12 §6 MathematicalValue interface.
var _ MathematicalValue = decimal.Decimal{}

// Implementation details are omitted; each method is derived from d.Form() / d.Coeff() / d.Exponent().
```

> **Why assertions are placed in SPEC 12 tests instead of `internal/decimal` production code**: `MathematicalValue` is now at root `internal/ecma402`;`internal/decimal` Production code must not reverse import the abstraction layer. Test assertions verify the contract but do not change the direction of production dependencies.
>
> **Why compile-time assertion `var _`**: Compilation fails immediately when the interface signature changes to avoid nil interface panic at runtime; Go idiom.

---

## 7. Error handling

### 7.1 Sentinel

```go
// internal/decimal/errors.go(signature)
package decimal

import "errors"

var (
// ErrInvalidDecimal: ParseString input illegal decimal literal.
    ErrInvalidDecimal = errors.New("decimal: invalid numeric literal")

// ErrInvalidRoundingIncrement: Not in the ValidRoundingIncrements list.
    ErrInvalidRoundingIncrement = errors.New("decimal: invalid rounding increment")

// ErrNaNComparison: When Cmp, any operand is NaN.
    ErrNaNComparison = errors.New("decimal: NaN in comparison")

// ErrLog10Domain: Log10Floor input is 0 or non-Finite.
    ErrLog10Domain = errors.New("decimal: log10 of zero or non-finite")
)
```

### 7.2 Reconcile with numberformat / pluralrules errors

`gointl.ErrInvalidOption` Error wrapping this package at formatter boundary:

```go
// numberformat/options.go (wrap mode in SPEC 20)
if !decimal.IsValidRoundingIncrement(inc) {
    return fmt.Errorf("numberformat: %w: roundingIncrement=%d", decimal.ErrInvalidRoundingIncrement, inc)
}
```

> **Why this package does not return the error after wrap**: Error boundary record - the error in this package is raw sentinel; wrapping by SPEC 20 / 40 adds context at the boundary.

---

## 8. Forbidden

### 8.1 ❌ Do not directly expose the `apd.Decimal` type

```go
// ❌ Error: public API leaks apd type
package numberformat
func WithRoundingMode(m apd.Rounder) Option

// ✅ Correct: use decimal.RoundingMode abstraction
package numberformat
func WithRoundingMode(m numberformat.RoundingMode) Option  // numberformat.RoundingMode = decimal.RoundingMode
```

> **Why**: When switching the Decimal backend, all public APIs are broken; `internal/decimal` provides a layer of isolation to allow transparent replacement of the backend.

### 8.2 ❌ Do not use `shopspring/decimal`

```go
// ❌ Error: No NaN/Inf, violation of "no panic" red line
import "github.com/shopspring/decimal"
d := decimal.NewFromString("NaN") // panic!

// ✅ Correct: apd native support
d, _ := decimal.ParseString("NaN")  // d.IsNaN() == true
```

### 8.3 ❌ Do not use `fmt.Sprintf` to convert value in hot path

```go
// ❌ Error: ~150 ns allocation per Format call
s := fmt.Sprintf("%v", anyValue)
d, _ := decimal.ParseString(s)

// ✅ Correct: type switch + dedicated constructor
d, err := decimal.ToIntlMathematicalValue(anyValue)
```

> **Why**: `fmt.Sprintf("%v", float64)` uses `%g` format; ECMA-402 ToIntlMathematicalValue requires "shortest round-trip"; both are inconsistent + Sprintf is significantly allocated on the hot path.

### 8.4 ❌ Do not use `math/big.Float` as an ECMA-402 value

```go
// ❌ Error: 0.1 + 0.2 bytes are not equal
f := new(big.Float).SetFloat64(0.1)
f.Add(f, big.NewFloat(0.2))
fmt.Println(f.Text('g', -1))  // "0.30000000000000004"

// ✅ Correct: Decimal
a, _ := decimal.ToIntlMathematicalValue("0.1")
b, _ := decimal.ToIntlMathematicalValue("0.2")
sum := a.Add(b)
fmt.Println(sum.String())  // "0.3"
```

> **Why**: Generated reference uses decimal BigDecimal; `big.Float` is IEEE-754 binary, and the conformance test must fail.

### 8.5 ❌ Do not silently return 0 when `Cmp` uses NaN

```go
// ❌ Error: Violation of IEEE-754 "NaN unordered"
func (d Decimal) Cmp(other Decimal) int {
if d.IsNaN() || other.IsNaN() { return 0 } // Pretend to be equal
    // ...
}

// ✅ Correct: Return error to force the caller to handle it
func (d Decimal) Cmp(other Decimal) (int, error) {
    if d.IsNaN() || other.IsNaN() { return 0, ErrNaNComparison }
    // ...
}
```

### 8.6 ❌ Don’t implement Go `==` on top of `Decimal`

```go
// ❌ Error: Decimal embedded apd.Decimal contains big.Int (non-comparable)
if d1 == d2 { /* compile error or undefined behavior */ }

// ✅ Correct: use Equal()
if d1.Equal(d2) { /* ... */ }
```

### 8.7 ❌ Do not put the `RoundingPriority` algorithm in SPEC 20

```go
// ❌ Error: SPEC 20 Repeated implementation of priority decision
package numberformat
func setNumberFormatDigitOptions(...) {
if priority == "morePrecision" { /* ... */ } // Repeat
}

// ✅ Correct: Call SPEC 21 §4.4 ApplyRoundingPriority
package numberformat
rt := decimal.ApplyRoundingPriority(hasSD, hasFD, opts.RoundingPriority)
```

> **Why**:`RoundingPriority` The decision algorithm is the focus of the mathematical layer; SPEC 20 is the consumer. Current deeds are recorded in this SPEC.

### 8.8 ❌ Do not import `internal/ecma402` in `internal/decimal` production code

```go
// ❌ Error: Circular dependency (internal/ecma402 §1.4 reverse direction is also prohibited)
import "github.com/agentable/go-intl/internal/ecma402"
func Foo() { ecma402.ToIntlMathematicalValue(...) }

// ✅ Correct: Contract assertions are placed in the internal/ecma402 test
package ecma402
import "github.com/agentable/go-intl/internal/decimal"
var _ MathematicalValue = decimal.Decimal{}
```

> **Why**: SPEC 12 §1.4 and SPEC 21 §1.2 are closed - the production code dependency direction remains `internal/ecma402` → `internal/decimal`, and contract checking is undertaken by the test.

---

## 9. Acceptance Criteria

### Backend

- [ ] `go.mod` contains `github.com/cockroachdb/apd/v3`, **not** contains `github.com/shopspring/decimal` or `github.com/ericlagergren/decimal`.
- [ ] `internal/decimal/` subdirectory is divided into files `decimal.go` / `from.go` / `rounding.go` / `quantize.go` / `log10.go` / `trailing_zero.go` / `priority.go` / `math_value.go` / `errors.go`.
- [ ] The public API of the `internal/decimal` package does not expose any `apd.*` types (`grep -r "apd\." | grep -v "internal/decimal/" | grep -v "_test.go"` returns null).

### Decimal type

- [ ] `Decimal` is a value type (struct); `Decimal{}` is `Form=Finite, Coeff=0, Exp=0` (equivalent to +0).
- [ ] `Form` enumeration values `Finite=0` / `Infinite=1` / `NaN=2` / `NaNSignaling=3` (can be converted by apd).
- [ ] Package-level singletons `Zero` / `NaNValue` / `PosInfinity` / `NegInfinity` cannot be mutated.
- [ ] `IsNaN` / `IsInf` / `IsFinite` are semantically consistent with the `Form()` method.

### Construction

- [ ] `New(negative, coeff, exp) Decimal` accepts `*big.Int` coefficients.
- [ ] `FromInt64(int64) Decimal` appears in benchmark telemetry with allocation reporting.
- [ ] `FromFloat64(NaN)` returns `NaNValue`(IsNaN()=true).
- [ ] `FromFloat64(±Inf)` returns `±PosInfinity`.
- [ ] `FromFloat64(0.1).Add(FromFloat64(0.2)).String() == "0.3"` (decimal not deviated).
- [ ] `ParseString("NaN")` / `"Infinity"` / `"-Infinity"` are each correct.
- [ ] `ParseString("foo")` returns `ErrInvalidDecimal` wrap error.

### Arithmetic

- [ ] `Add` / `Sub` / `Mul` / `Div` Do not modify receiver; return new Decimal.
- [ ] NaN contagion: any operand NaN → result NaN.
- [ ] Inf Arithmetic as per IEEE-754: `Inf - Inf = NaN`, `0 × Inf = NaN`, `Inf / Inf = NaN`.
- [ ] `Div` returns NaN in `0 / 0`, `nonzero / 0` returns `±Inf`, **not** panic.
- [ ] `Cmp(NaN, _) → (0, ErrNaNComparison)`.
- [ ] `Equal(NaN, NaN) == false`(IEEE-754).
- [ ] `PowerOf10(n) == 10^n` is all accurate to `-100 ≤ n ≤ 100`.

### ToIntlMathematicalValue

- [ ] `ToIntlMathematicalValue(int64(98765)).String() == "98765"`.
- [ ] `ToIntlMathematicalValue("3.14").Exponent() == -2` and `Coeff() == "314"`.
- [ ] `ToIntlMathematicalValue(math.NaN()).IsNaN() == true`.
- [ ] `ToIntlMathematicalValue(math.Inf(+1)).IsInf() == true` and `Negative() == false`.
- [ ] `ToIntlMathematicalValue(nil)` returns `ErrInvalidDecimal`.
- [ ] `BenchmarkToIntlMathematicalValue_Int64` appears in non-blocking benchmark telemetry.
- [ ] `BenchmarkToIntlMathematicalValue_String_3p14` appears in non-blocking benchmark telemetry.
- [ ] generated-reference `bigdecimal/tests/` All fixtures pass in `internal/decimal/from_test.go`.

### Rounding Modes

- [ ] `RoundingMode` 9 constants; `String()` output spec verbatim name (`"halfCeil"` not `"half-ceil"`).
- [ ] `ParseRoundingMode("halfExpand") == RoundHalfExpand`,`ParseRoundingMode("HALFEXPAND")` failed (case sensitive).
- [ ] `ApplyUnsignedRoundingMode` results under `halfCeil` / `halfFloor` with generated-reference `ApplyUnsignedRoundingMode.test.ts` byte-equal.
- [ ] `GetUnsignedRoundingMode` aligns with spec §15.5.6 verbatim table.
- [ ] generated-reference `ecma402-abstract/NumberFormat/tests/ApplyUnsignedRoundingMode.test.ts` All fixtures pass.

### RoundingPriority

- [ ] `ApplyRoundingPriority` 5 branches (MorePrecision / LessPrecision / hasSD / hasFD / Compact / default fractionDigits) aligned with generated-reference `SetNumberFormatDigitOptions.ts`.
- [ ] `RoundingType` 5 values (FractionDigits / SignificantDigits / MorePrecision / LessPrecision / Compact).

### RoundingIncrement

- [ ] `ValidRoundingIncrements` verbatim 17 value.
- [ ] `IsValidRoundingIncrement(3) == false`,`IsValidRoundingIncrement(50) == true`.
- [ ] `QuantizeToIncrement(123.456, 25, -2, RoundHalfExpand).String() == "123.5"`(125/25=5; round half expand; 25×0.05=1.25? See fixture).
- [ ] generated-reference `ecma402-abstract/NumberFormat/tests/Quantize.test.ts` fixture passed (if exists).

### TrailingZeroDisplay

- [ ] `ApplyTrailingZeroDisplay("3.00", true, TrailingZeroStripIfInteger) == "3"`.
- [ ] `ApplyTrailingZeroDisplay("3.14", false, TrailingZeroStripIfInteger) == "3.14"` (non-integer does not strip).
- [ ] `ApplyTrailingZeroDisplay("3.00", true, TrailingZeroAuto) == "3.00"`(auto reserved).

### Log10Floor

- [ ] `Log10Floor(FromInt64(98765)) == 4`(98765 ∈ [10^4, 10^5)).
- [ ] `Log10Floor(FromInt64(0))` returns `ErrLog10Domain`.
- [ ] `Log10Floor(NaNValue)` returns `ErrLog10Domain`.

### MathematicalValue interface

- [ ] `var _ MathematicalValue = decimal.Decimal{}` compiles in `internal/ecma402` test.
- [ ] `internal/decimal` production files **not** imported `internal/ecma402`;`rg '"github.com/agentable/go-intl/internal/ecma402"' internal/decimal --glob '!**/*_test.go'` should be empty.

### Error

- [ ] `errors.Is(err, ErrInvalidDecimal)` is true if `ParseString` fails.
- [ ] `errors.Is(err, ErrInvalidRoundingIncrement)` is true after invalid rounding increment is wrapped by the callee.
- [ ] There are **no** `panic` calls in the package.

### Test

- [ ] generated-reference `bigdecimal/tests/` All fixtures were ported to `internal/decimal/testdata/` and passed.
- [ ] generated-reference `ecma402-abstract/NumberFormat/tests/{ApplyUnsignedRoundingMode,SetNumberFormatDigitOptions}.test.ts` fixture passed.
- [ ] Use `t.Parallel()` for all tests.
- [ ] `BenchmarkFromInt64` / `BenchmarkToIntlMathematicalValue_Int64` run on `task test:bench`, recorded to SPEC 71 §benchmark.
- [ ] At least 1 `Example*` function demonstrating `ToIntlMathematicalValue` + `String`.

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

- `.references/formatjs/packages/bigdecimal/src/index.ts` —— `BigDecimal{mantissa, exponent, specialValue}` and `add` / `sub` / `mul` / `div` / `quantize` / `log10`
- `.references/formatjs/packages/bigdecimal/tests/` —— fixture
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatDigitOptions.ts` —— `RoundingPriority` / `RoundingIncrement` / `roundingType` five-way branch
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ApplyUnsignedRoundingMode.ts` —— `halfCeil` / `halfFloor` self-implementation path
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/GetUnsignedRoundingMode.ts`
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ToIntlMathematicalValue.ts` —— spec §6.4 implementation
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ToRawFixed.ts` / `ToRawPrecision.ts` —— RoundingMode consumer
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ComputeExponent.ts` —— `Log10Floor` Consumer

### Library survey

- `github.com/cockroachdb/apd/v3` —— selected backend;Apache-2.0;`Form` / `Context` / `Log10` / `Quantize` / `Round` / 8 GDA modes
- `github.com/shopspring/decimal` —— ❌ reject; no NaN/Inf; construct panic
- `github.com/ericlagergren/decimal` —— ❌ Rejected; v3 alpha long-term stagnation (no new commits after 2024-04)
- `math/big.Float` —— ❌ reject; binary IEEE-754; conformance byte-equality fails
- `math/big.Rat` —— ❌ Reject; None Log10 / Directional Rounding

### Cross-SPEC

- [SPEC 00 §8 Q1 — Decimal backend selection](./00-vision-and-scope.md#8-open-questions)(This SPEC is closed)
- [SPEC 12 §6 — MathematicalValue interface](./12-abstract-operations.md#6-math-value-boundary) — This SPEC `Decimal` implements this interface
- [SPEC 12 §1 — Package Layout(forbidden import)](./12-abstract-operations.md#1-package-layout) — This SPEC §1.2 is closed with
- [SPEC 20 §Format Pipeline](./20-numberformat.md) - This SPEC is its mathematical layer
- [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract) —— compact notation constructs OperandsRecord through `Decimal` and format string
- [SPEC 50 §6 — Data Access API](./50-cldr-data.md#6-data-access-api) ——Currency default precision data (`CurrencyDigits`) is injected by SPEC 50, this SPEC is not defined repeatedly
- [SPEC 71 §Benchmark](./71-benchmark.md) ——This SPEC §3.3 / §9 performance target corresponds


---

> This SPEC is a maintenance record of `internal/decimal` and the ECMA-402 math layer. Added ECMA-402 rounding mode (spec rare) or `apd/v3` upgrade triggers this SPEC revision; backend switching (if it occurs) is updated in this SPEC §1.1 decision table.
