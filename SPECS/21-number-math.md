# SPEC 21 — Number Math & Decimal

> **Status:** Active
> **Authority:** ECMA-402 number-format algorithms are normative. This SPEC documents the current Go contract for the `internal/decimal` package, `Decimal` type, typed numeric bridges, nine ECMA-402 rounding modes, exact increment quantization, and trailing-zero arithmetic. ECMA-402 digit-option and rounding-branch resolution is owned by SPEC 20.

---

## Overview

ECMA-402's mathematical-value algorithms consume values that can express
`NaN`, `±Infinity`, and arbitrary-precision finite decimals. JavaScript first
normalizes dynamic values at its host boundary. Go instead exposes explicit
`Int`, `Uint`, `Float`, `BigInt`, and `Decimal` constructors; those constructors
produce one closed `NumericValue` record before formatter algorithms run.

Generated reference uses `@formatjs/bigdecimal` to implement this data structure. Go uses `cockroachdb/apd/v3` (IEEE 754-2008 GDA Decimal) for finite/special-value representation, while ECMA-402 rounding uses exact coefficient/exponent arithmetic. No finite `apd.Context.Precision` participates in selecting rounding neighbours or ties.

This SPEC decides:

1. Backend selection **`cockroachdb/apd/v3`**; **rejection** `shopspring/decimal` (no NaN/Inf construct panic) and `ericlagergren/decimal` (v3 alpha long-term stagnation).
2. The `Decimal` type that provides the ECMA-402 abstraction layer narrow interface in `internal/decimal/`; apd as a backend, but does not expose the apd type through the public API (to facilitate future switching).
3. The nine ECMA-402 rounding modes map to the five unsigned selection rules and are evaluated against exact adjacent values.
4. Exact rounding-mode selection and increment quantization are implemented in this package. `RoundingPriority` and the resolved rounding branch belong to `internal/ecma402/numberformat`; `internal/decimal` does not own formatter option policy.
5. Formatter packages carry numeric inputs through `internal/ecma402.NumericValue`; the concrete `internal/decimal.Decimal` value is consumed directly by shared number-format and plural-rules algorithms. There is no active `MathematicalValue` interface layer.

This SPEC **not** defines: NumberFormat option pipeline (SPEC 20), PluralRules rule compilation (SPEC 40), CLDR currency precision data (SPEC 50), or formatter public `Value` wrappers.

---

## 1. Backend selection

### 1.1 Decision:`cockroachdb/apd/v3`

```text
require github.com/cockroachdb/apd/v3 v3.x.x
```

| Candidate | Decision | Key Reasons |
|------|------|---------|
| **`github.com/cockroachdb/apd/v3`** | ✅ Selected | IEEE 754-2008 GDA value representation; `Form` enumeration corresponds to generated-reference `specialValue`; arbitrary-size decimal coefficient; Apache-2.0; maintained upstream. |
| `shopspring/decimal` | ❌ Reject | No NaN/Inf representation; `NewFromString("NaN")` panic, conflicts with CLAUDE.md "no panic in production" red line; missing `Log10` native |
| `ericlagergren/decimal` | ❌ Rejected | v3 remains long-term alpha; ABI stability is weaker for a formatter math core. |
| `math/big.Float` | ❌ Rejected | Binary floating point; `0.1 + 0.2` converted to decimal string and reference output **bytes are not equal** (violates SPEC 70 conformance) |
| `math/big.Rat` | ❌ Reject | Purely rational; no `Log10` with directional rounding; ToRawPrecision is too expensive to implement |

> **Why `cockroachdb/apd/v3`**:
> 1. **GDA matches the required value forms** - `apd.Form` enumeration (`Finite=0` / `Infinite=1` / `NaN=2` / `NaNSignaling=3`) directly corresponds to the generated reference's finite and special-value representation.
> 2. **Exact representation without precision policy** - apd owns the decimal coefficient, exponent, sign, and special forms. ECMA-402 neighbour selection remains exact for arbitrary-size `BigInt` and decimal-string inputs instead of inheriting a context precision.
> 3. **Backend boundary** - formatter code consumes the narrow `Decimal` wrapper. Coefficient/exponent operations needed by rounding stay private and do not expose apd through public APIs.
> 4. **Active maintenance** - cockroachdb main repository, continuous submission, Apache-2.0 compatible with go-intl license.
>
> **Rejected `shopspring/decimal` Details**:
> - ❌ No NaN / +Inf / -Inf representation; direct `panic("decimal: NaN not supported")` during construction, violating the "no panic" red line.
> - ❌ Formatter `Float(math.NaN())` and `Decimal("NaN")` bridges must preserve a NaN-shaped value; `shopspring` would require a separate sentinel representation.
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
├── decimal.go    ← Decimal wrapper and read-only value operations
├── from.go       ← exact typed construction
├── ops.go        ← exact arithmetic and sign operations
├── rounding.go   ← nine RoundingMode values + unsigned selection
├── quantize.go   ← exact increment quantization
├── log10.go      ← decimal magnitude
└── errors.go     ← decimal sentinels
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

// Decimal is the normalized numerical representation used by ECMA-402 math.
// Packaging apd.Decimal; does not directly expose the apd type to facilitate backend switching in the future.
//
// Representation semantics:
//   - Form == Finite:     value = (-1)^Negative × Coeff × 10^Exponent
// - Form == Infinite: ±∞,Negative determining symbol
// - Form == NaN: Quiet NaN
// - Form == NaNSignaling: signaling NaN retained only as a backend form
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

// Sign returns -1 / 0 / +1 (NaN returns 0, Inf according to the Negative field).
func (d Decimal) Sign() int

// IsZero is true only when d is +0 or -0 (Form=Finite, Coeff=0).
func (d Decimal) IsZero() bool

// IsNaN / IsInf / IsFinite are syntactic sugar for Form checks.
func (d Decimal) IsNaN() bool
func (d Decimal) IsInf() bool
func (d Decimal) IsFinite() bool

// Negative returns the sign bit (Inf / Finite are both meaningful).
func (d Decimal) Negative() bool

// Cmp compares finite decimal values. Callers must reject NaN/non-finite inputs
// before using it where ECMA-402 would throw.
func (d Decimal) Cmp(other Decimal) int
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
> **Why `FromFloat64` uses string conversion instead of `apd.NewFromFloat`**:
> a JavaScript Number bridge carries the shortest round-trip decimal value of
> its binary `float64`; `apd.NewFromFloat` instead exposes binary expansion
> digits such as `0.30000000000000004`.

### 2.3 Arithmetic

<a id="decimal-cmp"></a>

```go
// internal/decimal/decimal.go(signature)

// Cmp compares finite values. Callers own NaN/non-finite rejection.
func (d Decimal) Cmp(other Decimal) int

// Abs returns d without a sign.
func Abs(d Decimal) Decimal

// AbsDiffCmp compares |base-a| and |base-b| for rounding-priority selection.
func AbsDiffCmp(base, a, b Decimal) int

// MulInt returns d multiplied by n, preserving non-finite values.
func MulInt(d Decimal, n int64) Decimal

// Scale10 returns d multiplied by 10^exp by adjusting the decimal exponent.
func Scale10(d Decimal, exp int32) Decimal
```

`internal/decimal` is not a general-purpose arithmetic package. It exports only
the operations consumed by NumberFormat and PluralRules: comparison, sign/finite
inspection, decimal scaling, integer multiplication, and rounding helpers.
Additional arithmetic stays unexported until a formatter algorithm actually
needs it.

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

<a id="rounding-modes"></a>

## 4. Rounding Modes

### 4.1 Nine ECMA-402 modes

ECMA-402 §15.5.5 defines nine signed rounding modes. `UnsignedRoundingMode` maps them to the five selection behaviours consumed by `ApplyUnsignedRoundingMode`:

| ECMA-402 Name | Positive input | Negative input |
|------------|------------------|------|
| `ceil` | infinity | zero |
| `floor` | zero | infinity |
| `expand` | infinity | infinity |
| `trunc` | zero | zero |
| `halfCeil` | half-infinity | half-zero |
| `halfFloor` | half-zero | half-infinity |
| `halfExpand` | half-infinity | half-infinity |
| `halfTrunc` | half-zero | half-zero |
| `halfEven` | half-even | half-even |

The implementation does not delegate these decisions to an apd `Rounder`: ECMA-402 requires both adjacent mathematical values so `halfCeil`, `halfFloor`, and `halfEven` can select the correct endpoint.

### 4.2 Go Type

```go
// internal/decimal/rounding.go(signature)

// RoundingMode is one of the nine modes specified in ECMA-402 §15.5.5.
// String() output spec verbatim name (for ResolvedOptions).
type RoundingMode string

const (
    RoundCeil       RoundingMode = "ceil"
    RoundFloor      RoundingMode = "floor"
    RoundExpand     RoundingMode = "expand"
    RoundTrunc      RoundingMode = "trunc"
    RoundHalfCeil   RoundingMode = "halfCeil"
    RoundHalfFloor  RoundingMode = "halfFloor"
    RoundHalfExpand RoundingMode = "halfExpand"
    RoundHalfTrunc  RoundingMode = "halfTrunc"
    RoundHalfEven   RoundingMode = "halfEven"
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
// m: RoundingMode (converted to unsigned - see UnsignedRoundingMode)
// Return r1 or r2 (one).
func ApplyUnsignedRoundingMode(x, r1, r2 Decimal, m RoundingMode) Decimal

// UnsignedRoundingMode(m, sign) converts signed mode to unsigned(spec §15.5.6).
// - ceil + minus sign → halfDown style unsigned
// - floor + positive sign → halfDown style
// - The rest of the modes remain unchanged
func UnsignedRoundingMode(m RoundingMode, isNegative bool) RoundingMode
```

> **Why According to spec name `ApplyUnsignedRoundingMode` verbatim**:generated-reference `ecma402-abstract/NumberFormat/ApplyUnsignedRoundingMode.ts` 1:1 implementation; the transplantation cost is the lowest.
>
> **Why not use `apd.Decimal.Quantize` + `apd.Rounder`**: The Rounder interface does not expose both ECMA-402 neighbours. `QuantizeToIncrement` therefore derives `r1` and `r2` by exact integer quotient/remainder, and `ApplyUnsignedRoundingMode` compares exact aligned coefficients. This also prevents a backend context from truncating 100+ digit values before the rounding decision.

### 4.4 Rounding branch ownership

```go
// internal/ecma402/numberformat/options.go (signature)
type RoundingType string

const (
    RoundingTypeFractionDigits    RoundingType = "fractionDigits"
    RoundingTypeSignificantDigits RoundingType = "significantDigits"
    RoundingTypeMorePrecision     RoundingType = "morePrecision"
    RoundingTypeLessPrecision     RoundingType = "lessPrecision"
)

type ResolvedDigitOptions struct {
    DigitOptions
    RoundingType RoundingType
}
```

`internal/decimal` owns mathematical rounding modes and exact arithmetic only.
`internal/ecma402/numberformat.SetNumberFormatDigitOptions` owns the ECMA-402
option-presence/defaulting algorithm and freezes the chosen branch in
`ResolvedDigitOptions`. NumberFormat and PluralRules pass that complete record
to `FormatNumericToString`; runtime code must not reconstruct the branch from
digit fields or a string priority.

> **Rejected**: a second `internal/decimal.ApplyRoundingPriority` policy layer.
> Rounding priority is an ECMA-402 option-resolution decision, not decimal
> arithmetic, and duplicate inference lets constructor state and formatting
> disagree.

### 4.5 RoundingIncrement

```go
// internal/decimal/quantize.go(signature)

// roundingIncrements contains the 15 values allowed by ECMA-402 §15.5.4.
// Other values → ErrInvalidRoundingIncrement.
var roundingIncrements = [...]int{
    1, 2, 5, 10, 20, 25, 50,
    100, 200, 250, 500,
    1000, 2000, 2500, 5000,
}

// RoundingIncrements returns a caller-owned copy.
func RoundingIncrements() []int

// IsValidRoundingIncrement validates a single option value.
func IsValidRoundingIncrement(inc int) bool

// QuantizeToIncrement(x, increment, exp, mode) rounds x to the nearest
// (increment × 10^exp) multiple.
// x: value to be rounded
// increment: must be present in roundingIncrements
// exp: magnitude (determined by mxfd)
//   mode      : RoundingMode
func QuantizeToIncrement(x Decimal, increment int, exp int32, mode RoundingMode) Decimal
```

> **Why a static increment set**: ECMA-402 §15.5.4 explicitly enumerates the values; any other value is a RangeError. The formatter validates once at construction, while the accessor returns a copy so callers cannot mutate package state.
>
> **Why `QuantizeToIncrement` uses coefficient/exponent arithmetic**: The target is an integer multiple of `increment × 10^exp`, not merely a decimal exponent. Exact integer quotient/remainder identifies the lower and upper multiples without division precision, then exact aligned coefficients determine distance and half-even cardinality. Replacing a fixed precision with a larger constant is forbidden because public `BigInt` and decimal-string inputs have no matching digit limit.
>
> **Work bound for extreme exponents**: Exact rounding must not materialize a
> `10^n` coefficient, or otherwise allocate work proportional to `n`, when the
> input is already an exact multiple of the step or when the rounded result can
> be selected without expanding that power. This keeps compact decimal records
> such as `1e+1000000000` and `1e-1000000000` bounded by the digits that must
> actually appear in the input or result. The implementation owner is
> `internal/decimal/quantize.go`; the regression witness is
> `internal/decimal/quantize_test.go` (`TestQuantizeToIncrementBoundsWorkForExtremeExponents`).

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
// For finite positive x, floor(log10(x)) is NumDigits(coeff)-1+exponent.
// The result is checked before conversion to int32.
func Log10Floor(x Decimal) (int32, error)
```

> **Why digit count instead of logarithm**: A finite decimal already stores an integer coefficient and base-10 exponent. Their decimal digit count gives the exact floor directly, including arbitrary-size values and exact powers of ten, with no floating or context precision boundary.

---

## 6. NumericValue bridge

### 6.1 Shared numeric record

Formatter packages do not expose `decimal.Decimal` directly. Public packages
wrap numeric inputs in package-local `Value` types; those wrappers carry an
`internal/ecma402.NumericValue` record:

```go
package ecma402

type NumericValueKind uint8

const (
    NumericValueDecimal NumericValueKind = iota
    NumericValueInt64
    NumericValueUint64
)

type NumericValue struct {
    Decimal decimal.Decimal
    Kind    NumericValueKind
    Int64   int64
    Uint64  uint64
}
```

The kind field preserves integer fast paths without widening the public API.
The decimal field remains the general ECMA-402 mathematical value consumed by
shared digit formatting and plural operand logic.

### 6.2 Conversion owner

```go
package ecma402

func DecimalNumericValue(value decimal.Decimal) NumericValue
func Int64NumericValue(value int64) NumericValue
func Uint64NumericValue(value uint64) NumericValue
func Float64NumericValue(value float64) NumericValue
func BigIntNumericValue(value *big.Int) NumericValue
```

Package-local public `Value` constructors are the only formatter input boundary.
They select one typed constructor above; decimal strings first pass through the
operation-appropriate parser. The core has no dynamic coercion entrypoint.

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

// ErrNaNComparison: Reserved comparison-domain sentinel; current Cmp callers
// reject NaN/non-finite values before comparison instead of returning this.
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
type Options struct { RoundingMode apd.Rounder }

// ✅ Correct: expose the ECMA-402 string vocabulary; parse internally.
package numberformat
type RoundingMode string
type Options struct { RoundingMode *string }
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

### 8.3 ❌ Do not add a dynamic coercion path

```go
// ❌ Wrong: the core inherits host-specific coercion rules.
func Coerce(value any) decimal.Decimal

// ✅ Correct: the public package exposes explicit typed bridges.
v := numberformat.Int(42)
d, err := numberformat.Decimal("9007199254740993")
```

> **Why**: JavaScript coercion is a host-language operation. An explicit Go
> bridge keeps accepted types, errors, signed zero, and integer precision
> visible in the API and executes conversion once before formatting.

### 8.4 ❌ Do not use `math/big.Float` as an ECMA-402 value

```go
// ❌ Error: 0.1 + 0.2 bytes are not equal
f := new(big.Float).SetFloat64(0.1)
f.Add(f, big.NewFloat(0.2))
fmt.Println(f.Text('g', -1))  // "0.30000000000000004"

// ✅ Correct: exact decimal-string bridge
d, _ := decimal.ParseString("0.3")
fmt.Println(d.String())  // "0.3"
```

> **Why**: Generated reference uses decimal BigDecimal; `big.Float` is IEEE-754 binary, and the conformance test must fail.

### 8.5 ❌ Do not use `Cmp` as validation for NaN

```go
// ❌ Error: comparison is being used before the native operation's error boundary.
if start.Cmp(end) == 0 {
    return startCategory
}

// ✅ Correct: reject or route non-finite input at the ECMA-402 owner boundary first.
if err := ecma402.RequireFiniteDecimalInput(start); err != nil { return err }
if err := ecma402.RequireFiniteDecimalInput(end); err != nil { return err }
same := start.Cmp(end) == 0
```

### 8.6 ❌ Don’t implement Go `==` on top of `Decimal`

```go
// ❌ Error: Decimal embedded apd.Decimal contains big.Int (non-comparable)
if d1 == d2 { /* compile error or undefined behavior */ }

// ✅ Correct: use Cmp after the ECMA-402 owner has rejected NaN/non-finite values.
if d1.Cmp(d2) == 0 { /* ... */ }
```

### 8.7 ❌ Do not put formatter option policy in `internal/decimal`

```go
// ❌ Wrong owner: decimal math infers an ECMA-402 option branch.
package decimal
func ApplyRoundingPriority(...) RoundingType

// ✅ Correct owner: SetNumberFormatDigitOptions resolves one complete record.
package ecma402nf
resolved, invalid, bad := SetNumberFormatDigitOptions(config, mnfd, mxfd, notation)
```

> **Why**: Rounding priority depends on option presence, notation defaults, and
> resolved digit slots. Those are ECMA-402 formatter semantics, while
> `internal/decimal` only chooses between exact mathematical neighbours.

### 8.8 ❌ Do not import `internal/ecma402` in `internal/decimal` production code

```go
// ❌ Error: Circular dependency (internal/ecma402 §1.4 reverse direction is also prohibited)
import "github.com/agentable/go-intl/internal/ecma402"
func Foo() { _ = ecma402.NumericValue{} }

// ✅ Correct: Contract assertions are placed in the internal/ecma402 test
package ecma402

type NumericValue struct {
    Decimal decimal.Decimal
    Kind    NumericValueKind
}
```

> **Why**: SPEC 12 §1.4 and SPEC 21 §1.2 are closed - the production code dependency direction remains `internal/ecma402` → `internal/decimal`. `internal/decimal` must not import the abstract-operation layer.

---

## 9. Acceptance Criteria

### Backend

- [ ] `go.mod` contains `github.com/cockroachdb/apd/v3`, **not** contains `github.com/shopspring/decimal` or `github.com/ericlagergren/decimal`.
- [ ] `internal/decimal/` contains only decimal representation, conversion, arithmetic, magnitude, rounding, quantization, and error ownership; formatter option policy remains in `internal/ecma402/numberformat`.
- [ ] The public API of the `internal/decimal` package does not expose any `apd.*` types (`grep -r "apd\." | grep -v "internal/decimal/" | grep -v "_test.go"` returns null).

### Decimal type

- [ ] `Decimal` is a value type (struct); `Decimal{}` is `Form=Finite, Coeff=0, Exp=0` (equivalent to +0).
- [ ] `Form` enumeration values `Finite=0` / `Infinite=1` / `NaN=2` / `NaNSignaling=3` (can be converted by apd).
- [ ] Package-level singletons `Zero` / `NaNValue` / `PosInfinity` / `NegInfinity` cannot be mutated.
- [ ] `IsNaN` / `IsInf` / `IsFinite` are semantically consistent with the stored form.

### Construction

- [ ] `New(negative, coeff, exp) Decimal` accepts `*big.Int` coefficients.
- [ ] `FromInt64(int64) Decimal` appears in benchmark telemetry with allocation reporting.
- [ ] `FromFloat64(NaN)` returns `NaNValue`(IsNaN()=true).
- [ ] `FromFloat64(±Inf)` returns `±PosInfinity`.
- [ ] `FromFloat64(0.1).String() == "0.1"` and `FromFloat64(0.2).String() == "0.2"` (shortest round-trip decimal conversion).
- [ ] `ParseString("NaN")` / `"Infinity"` / `"-Infinity"` are each correct.
- [ ] `ParseString("foo")` returns `ErrInvalidDecimal` wrap error.

### Arithmetic

- [ ] `Cmp` compares finite values and callers reject NaN/non-finite inputs before using it for ECMA-402 equality.
- [ ] `Abs`, `AbsDiffCmp`, `MulInt`, and `Scale10` preserve value semantics and do not mutate caller-owned inputs.
- [ ] `Scale10` preserves NaN and infinities unchanged.

### Rounding Modes

- [ ] `RoundingMode` 9 constants; `String()` output spec verbatim name (`"halfCeil"` not `"half-ceil"`).
- [ ] `ParseRoundingMode("halfExpand") == RoundHalfExpand`,`ParseRoundingMode("HALFEXPAND")` failed (case sensitive).
- [ ] `ApplyUnsignedRoundingMode` results under `halfCeil` / `halfFloor` with generated-reference `ApplyUnsignedRoundingMode.test.ts` byte-equal.
- [ ] `UnsignedRoundingMode` aligns with spec §15.5.6 verbatim table.
- [ ] generated-reference `ecma402-abstract/NumberFormat/tests/ApplyUnsignedRoundingMode.test.ts` All fixtures pass.

### Resolved rounding branch

- [ ] `SetNumberFormatDigitOptions` freezes fraction, significant, more-precision, or less-precision routing in `ResolvedDigitOptions` and `FormatNumericToString` executes that field without re-inference.
- [ ] All nine rounding-mode strings are parsed to `decimal.RoundingMode` during option resolution; no runtime fallback substitutes `halfExpand` for invalid state.

### RoundingIncrement

- [ ] `RoundingIncrements()` returns the 15 sanctioned values and a caller cannot mutate package state through the returned slice.
- [ ] `IsValidRoundingIncrement(3) == false`,`IsValidRoundingIncrement(50) == true`.
- [ ] `QuantizeToIncrement(123.456, 25, -2, RoundHalfExpand).String() == "123.50"` (the target step is `25 × 10^-2 = 0.25`).
- [ ] 1-, 100-, 101-, 250-, and 1000-digit coefficients retain their low-order rounding information across positive and negative exponent alignment.
- [ ] No fixed `apd.Context.Precision` participates in quotient, neighbour, distance, or half-even selection.
- [ ] `QuantizeToIncrement` handles compact inputs with exponents `+1_000_000_000` and `-1_000_000_000` without constructing exponent-sized powers of ten; exact multiples remain unchanged and tiny values select zero or one target step according to the rounding mode.
- [ ] generated-reference `ecma402-abstract/NumberFormat/tests/Quantize.test.ts` fixture passed (if exists).

### TrailingZeroDisplay

- [ ] `ApplyTrailingZeroDisplay("3.00", true, TrailingZeroStripIfInteger) == "3"`.
- [ ] `ApplyTrailingZeroDisplay("3.14", false, TrailingZeroStripIfInteger) == "3.14"` (non-integer does not strip).
- [ ] `ApplyTrailingZeroDisplay("3.00", true, TrailingZeroAuto) == "3.00"`(auto reserved).

### Log10Floor

- [ ] `Log10Floor(FromInt64(98765)) == 4`(98765 ∈ [10^4, 10^5)).
- [ ] `Log10Floor(FromInt64(0))` returns `ErrLog10Domain`.
- [ ] `Log10Floor(NaNValue)` returns `ErrLog10Domain`.

### NumericValue bridge

- [ ] `internal/ecma402.NumericValue` preserves decimal, int64, and uint64 bridge kinds.
- [ ] `internal/decimal` production files **not** imported `internal/ecma402`;`rg '"github.com/agentable/go-intl/internal/ecma402"' internal/decimal --glob '!**/*_test.go'` should be empty.

### Error

- [ ] `errors.Is(err, ErrInvalidDecimal)` is true if `ParseString` fails.
- [ ] `errors.Is(err, ErrInvalidRoundingIncrement)` is true after invalid rounding increment is wrapped by the callee.
- [ ] There are **no** `panic` calls in the package.

### Test

- [ ] generated-reference `bigdecimal/tests/` All fixtures were ported to `internal/decimal/testdata/` and passed.
- [ ] generated-reference `ecma402-abstract/NumberFormat/tests/{ApplyUnsignedRoundingMode,SetNumberFormatDigitOptions}.test.ts` fixture passed.
- [ ] Use `t.Parallel()` for all tests.
- [ ] Relevant decimal construction benchmarks run through `task bench:run` and remain non-blocking telemetry.

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
- `.references/formatjs/packages/ecma402-abstract/ToIntlMathematicalValue.ts` —— JavaScript host-boundary coercion reference; not a Go core API
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ToRawFixed.ts` / `ToRawPrecision.ts` —— RoundingMode consumer
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/ComputeExponent.ts` —— `Log10Floor` Consumer

### Library survey

- `github.com/cockroachdb/apd/v3` —— selected representation backend; Apache-2.0; `Decimal` / `Form` / arbitrary-size coefficient
- `github.com/shopspring/decimal` —— ❌ reject; no NaN/Inf; construct panic
- `github.com/ericlagergren/decimal` —— ❌ Rejected; v3 alpha long-term stagnation (no new commits after 2024-04)
- `math/big.Float` —— ❌ reject; binary IEEE-754; conformance byte-equality fails
- `math/big.Rat` —— ❌ Reject; None Log10 / Directional Rounding

### Cross-SPEC

- [SPEC 00 §8 Q1 — Decimal backend selection](./00-vision-and-scope.md#8-open-questions)(This SPEC is closed)
- [SPEC 12 §Numeric value boundary](./12-abstract-operations.md#6-math-value-boundary) — Formatter bridges carry `internal/ecma402.NumericValue` records backed by `internal/decimal.Decimal`
- [SPEC 12 §1 — Package Layout(forbidden import)](./12-abstract-operations.md#1-package-layout) — This SPEC §1.2 is closed with
- [SPEC 20 §Format Pipeline](./20-numberformat.md) - This SPEC is its mathematical layer
- [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract) —— compact notation constructs OperandsRecord through `Decimal` and format string
- [SPEC 50 §6 — Data Access API](./50-cldr-data.md#6-data-access-api) ——Currency default precision data (`CurrencyDigits`) is injected by SPEC 50, this SPEC is not defined repeatedly
- [SPEC 71 §Benchmark](./71-benchmark.md) ——This SPEC §3.3 / §9 performance target corresponds


---

> This SPEC is a maintenance record of `internal/decimal` and the ECMA-402 math layer. Added ECMA-402 rounding mode (spec rare) or `apd/v3` upgrade triggers this SPEC revision; backend switching (if it occurs) is updated in this SPEC §1.1 decision table.
