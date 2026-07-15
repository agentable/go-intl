# SPEC 40 — PluralRules

> **Status:** Draft (2026-05-08)
> **Owner:** `pluralrules/` + `internal/plural/` + `tools/gen-plural-rules/`
> **Reference contract:** `.references/ecma402/spec/pluralrules.html` first, then `formatjs/packages/intl-pluralrules/` + `formatjs/packages/ecma402-abstract/PluralRules/`

## Overview

Define `pluralrules.PluralRules` public API, `OperandsRecord` type record, codegen strategy for cardinal / ordinal rules, typed select / range algorithm, compact operand contract with NumberFormat.

This SPEC does not redefine:
- `Locale` structure → see [SPEC 10 §Locale structure](./10-locale.md)
- `Decimal` types and operations → see [SPEC 21 §Decimal API](./21-number-math.md#decimal-api)
- digit option rounding and zero padding implementation → see [SPEC 20 §Digit Formatting Record](./20-numberformat.md#23-digit-formatting-record)
- NumberFormat compact path → see [SPEC 20 §Compact Notation](./20-numberformat.md)
- CLDR data generator architecture → see [SPEC 50 §Codegen](./50-cldr-data.md#codegen)
- General entrance to abstract operations → See [SPEC 12 §Abstract Ops](./12-abstract-operations.md)

---

## 1. Public API

### 1.1 Type

```go
package pluralrules

type Type string
const (
    Cardinal Type = "cardinal"
    Ordinal  Type = "ordinal"
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

type PluralRules struct{ /* Immutable; including resolved + compiled plural function pointer */ }

type Options struct {
    LocaleMatcher        *string
    Type                 *string
    MinimumIntegerDigits *int
    MinimumFractionDigits *int
    MaximumFractionDigits *int
    MinimumSignificantDigits *int
    MaximumSignificantDigits *int
    RoundingIncrement    *int
    RoundingMode         *string
    RoundingPriority     *string
    TrailingZeroDisplay  *string
    Notation             *string
    CompactDisplay       *string
}

func New(locales locale.List, opts Options) (*PluralRules, error)

type Value struct{ /* opaque */ }

func Int(v int64) Value
func Uint(v uint64) Value
func Float(v float64) Value
func BigInt(v *big.Int) Value
func Decimal(s string) (Value, error)

func (r *PluralRules) Select(v Value) Category
func (r *PluralRules) SelectRange(start, end Value) (Category, error)
func (r *PluralRules) ResolvedOptions() ResolvedOptions
```

**MUST** Rules:

1. `New` **MUST** complete all option verification during the construction period, and `error` will be returned if it fails.
2. `New` accepts a `Options` value. `New(locales, Options{})` is equivalent to JS passing an empty options object or omitting options; multiple options objects are not Go API shapes and are rejected by the compiler.
3. `Select` / `SelectRange` **MUST** accept opaque `Value`; the caller expresses ECMA-402 numeric input through typed constructors.
4. `Select` **MUST** return `Other` for NaN and positive or negative infinity. `SelectRange` **MUST** return `ErrInvalidValue` only when an endpoint is NaN and **MUST** accept positive or negative infinity. `Decimal` **MUST** accept well-formed decimal strings plus `NaN`, `Infinity`, and `-Infinity`, and **MUST** return `ErrInvalidValue` for malformed strings. Finite decimal spellings denote mathematical values: input-only trailing zeros do not affect category or formatted range equality.
5. `PluralRules` is an immutable value; all methods on `*PluralRules` must be concurrency-safe.
6. `Options` is the only public configuration value; restoring functional options or multiple options merge is prohibited.
7. The `Category.String()` return value **MUST** be consistent with the ECMA-402 string representation (to facilitate direct case branching of the `:plural` function of messageformat-go).
8. `New` **MUST** resolve the exact generated cardinal or ordinal rule for its data locale and freeze that function on the formatter. Missing active-locale data is a constructor failure; `Select` is total only after successful construction and therefore returns `Category` without an error.

> **Rejected**: public `BigFloat(*big.Float)` - ECMA-402 has Number, BigInt, and string-to-decimal bridges; arbitrary-precision floating point is a Go convenience shape with no native Intl owner. Callers that need exact decimal operands should use `Decimal`.

### 1.2 Options

```go
type Options struct {
    LocaleMatcher        *string
    Type                 *string
    MinimumIntegerDigits *int
    MinimumFractionDigits *int
    MaximumFractionDigits *int
    MinimumSignificantDigits *int
    MaximumSignificantDigits *int
    RoundingIncrement    *int
    RoundingMode         *string
    RoundingPriority     *string
    TrailingZeroDisplay  *string
    Notation             *string
    CompactDisplay       *string
}
```

**MUST** Rules:

1. Exposed options must override ECMA-402 `Intl.PluralRules` number-format digit surface: notation, compactDisplay, fraction/significant digits, roundingIncrement, roundingMode, roundingPriority, trailingZeroDisplay.
2. String option fields **MUST** use package named types and constants as vocabulary, but the public `Options` fields are `*string`; nil means omitted/default and `gointl.String("")` remains an explicit invalid value.
3. `New` accepts a `Options` value. Passing in multiple `Options` is a compile-time call site error.
4. String and scalar option fields must express presence with pointers: `LocaleMatcher`, `Type`, `Notation`, `CompactDisplay`, `RoundingMode`, `RoundingPriority`, `TrailingZeroDisplay`, all digit integer fields, and `RoundingIncrement`; the call point uses `gointl.String(v)` or `gointl.Int(n)`, and the constructor copies pointee values into the internal config.

> **Why typed Options**:
> 1. Go callers should see optional value boundaries at compile time; `Ordinal`, `CompactNotation`, and `HalfEvenRoundingMode` remain package vocabulary while `gointl.String(...)` expresses ECMA-402 option presence.
> 2. A `Options` value is easier to compare, cache, and document than functional options, and it does not hide state in the closure execution order.
> 3. If messageformat-go holds an ICU string, an explicit mapping should be done at the adapter boundary; the pressure of transparent transmission of the upstream string cannot be transferred to go-intl's long-term public API.
>
> **Rejected**: `WithType(string)` - scatter verification timing and make cache keys depend on closure order.
> **Rejected**: Accepts both direct enum values and pointer strings - dual-rail input parameters = verification path bifurcation + cache key bifurcation.

#### 1.2.1 Digit option pipeline

The digit options of PluralRules are a subset of the NumberFormat digit pipeline. The `pluralrules` package **MUST** map integer/fraction/significant digits and rounding options to `internal/ecma402/numberformat.DigitOptions`, and call `internal/ecma402/numberformat.FormatNumericToString` to get the source decimal string and rounded numeric value used by `ResolvePlural`.

**MUST** Rules:

1. PluralRules default digit options **MUST** be `minimumIntegerDigits=1`, `minimumFractionDigits=0`, `maximumFractionDigits=3`, `roundingIncrement=1`, `roundingMode="halfExpand"`, `roundingPriority="auto"`, `trailingZeroDisplay="auto"`.
2. `Select` / `SelectRange` **MUST** compute operands from the source output string of a shared digit formatter; **MUST NOT** copy `trimFraction`, `padMinimumIntegerDigits` or fixed rounding logic within the `pluralrules` package.
3. When selecting category with a negative number, the format string prefix `-` must be removed because ECMA-402 operands are based on absolute values; trailing zeros introduced by resolved digit options are retained by the shared formatter, while input-only lexical zeros are not.

> **Why**: The observable surface of PluralRules for digit options is not the final string, but the `v/w/f/t` of operands. These fields rely on the trailing zeros produced by resolved digit options, not on the spelling of the input value; the shared formatter is the minimum boundary that prevents NumberFormat and PluralRules from bifurcating those semantics.

#### 1.2.2 typed constants and ResolvedOptions

```go
type ResolvedOptions struct {
    Locale         locale.Locale
    Type           Type
    Notation       Notation
    CompactDisplay *CompactDisplay
    // ... digit options, then PluralCategories and rounding options
}
```

The full field order is fixed in [§6 ResolvedOptions](#6-resolvedoptions); `notation` and `compactDisplay` sit at ECMA-402 §16.4.5 positions 3-4.

### 1.3 Calling example

```go
pr, err := pluralrules.New(mustLocaleList("en"), pluralrules.Options{})
// pr.Select(pluralrules.Int(1)) == pluralrules.One
// pr.Select(pluralrules.Int(2)) == pluralrules.Other
ordinal, err := pluralrules.New(mustLocaleList("en"),
    pluralrules.Options{Type: gointl.String(pluralrules.Ordinal)})
// ordinal.Select(pluralrules.Int(1)) == pluralrules.One

pr, _ := pluralrules.New(mustLocaleList("en"),
    pluralrules.Options{Type: gointl.String(pluralrules.Ordinal)})
// pr.Select(pluralrules.Int(1)) == pluralrules.One   ("1st")
// pr.Select(pluralrules.Int(2)) == pluralrules.Two   ("2nd")
// pr.Select(pluralrules.Int(3)) == pluralrules.Few   ("3rd")
// pr.Select(pluralrules.Int(4)) == pluralrules.Other ("4th")

cat, err := pr.SelectRange(pluralrules.Int(1), pluralrules.Int(5)) // Other, nil
```

---

## 2. OperandsRecord <a id="operandsrecord"></a> <a id="operands"></a>

`OperandsRecord` is the ECMA-402 operand representation shared by PluralRules and NumberFormat. The current contract is recorded in this SPEC; the actual Go type definitions are in `internal/plural/plural.go`, which the `pluralrules`, `numberformat`, and `internal/cldr/plural` packages all import directly (the earlier pure-alias `internal/ecma402/pluralrules` package was removed so one type keeps one import path).

### 2.1 Type

```go
package plural // (actually located at internal/plural, imported as pluralop)

type OperandsRecord struct {
N OperandValue // |x| original magnitude value (used for "n" reference in plural rule)
    I OperandValue // integer digits, |trunc(x)|
V int // number of fraction digits (including trailing 0)
W int // number of fraction digits (go to trailing 0)
F OperandValue // fraction digits as integer (including trailing 0)
T OperandValue // fraction digits as integer (go to trailing 0)
    C int      // compact exponent (== E)
    E int      // compact/scientific exponent
}
```

**MUST** Rules:

1. The `OperandsRecord` field **MUST** be consistent with ECMA-402 §16.5.1 GetOperands + ES 2024 `c/e` extension.
2. `C` and `E` **MUST** always be equal (ECMA-402 explicit constraint: c is an alias of e); saving two copies of the field is for readability, but writing them together when assigning values.
3. `N/I/F/T` **MUST** use exact `OperandValue`; downgrading CLDR discrete rules to `float64`, `uint64`, or public `*big.Int` is prohibited.
4. The `OperandsRecord` type **MUST** be located in `internal/plural/plural.go`;`pluralrules`, `numberformat`, and `internal/cldr/plural` share the same import path (`internal/plural`, aliased `pluralop`).

> **Why**:
> 1. ECMA-402 OperandsRecord is the input field of plural rule DSL, and the field name and field semantics cannot be changed.
> 2. `OperandValue` retains the decimal digit view, allowing `%`, integer comparison and interval judgment to remain accurate for inputs exceeding the IEEE-754 safe integer.
> 3. The shared path `internal/plural/plural.go` instead of `pluralrules/operands.go`:NumberFormat’s compact path needs to be accessed from the `internal/` path to avoid circular dependencies.
>
> **Rejected**:
> - `float64` carries `N` - plural rule is a discrete rule, and binary floating point will create errors at the boundaries of large integers and decimals.
> - `uint64` / `*big.Int` directly as field types - will leak the underlying representation to generated rule call sites.
> - `OperandsRecord` is placed in the `pluralrules/` package public layer - NumberFormat should not import the `pluralrules` public package (layered inversion).

### 2.2 GetOperands

```go
// internal/plural/plural.go
func GetOperands(formatted string, exponent int) OperandsRecord
```

**MUST** Rules:

1. `GetOperands` implementation **MUST** be consistent with generated-reference `GetOperands.ts` output semantics (for trailing zero behavior).
2. `i / v / w / f / t` **MUST** be extracted from the `formatted` string (already output by `FormatNumericToString`, including `mnfd` mandatory trailing 0):
- `v` = decimal string length (including trailing 0)
- `w` = the length of the decimal string after trailing 0
- `f` = convert a decimal string into an integer
- `t` = remove trailing 0 and convert the decimal string into an integer
- `i` = Use the integer part as an integer (`|trunc(x)|`)
3. `n` **MUST** be the absolute decimal value of `formatted`, which is saved as `OperandValue`; it is forbidden to obtain it through `strconv.ParseFloat`.
4. `c / e` **MUST** be equal to parameter `exponent`.
5. Non-finite numbers do not enter `GetOperands`; `ResolvePlural` short-circuits NaN and positive or negative infinity to category `Other` while retaining the corresponding formatted special-value string for range equality.

> **Why**: `v/w/f/t` is extracted from the `formatted` string instead of `decimal.Decimal` because trailing zero is an "observable attribute determined by the formatter" (`mnfd=2` when `1` → `"1.00"` is `v=2, w=0`), not a mathematical attribute. Generated reference extracts semantically aligned strings. Non-finite special values have no plural operands; ECMA-402 owns them in `ResolvePlural` before operand construction.
>
> **Note**: The record is based on the `formatted` string (reference behavior), after input spelling has already been reduced to an Intl mathematical value. This keeps one display-digit view: formatter options may introduce visible zeros; parser scale may not.

---

## 3. CLDR Plural Rules Codegen

### 3.1 Decision: codegen to Go source

**Decision**: active scope PluralRules implementation **MUST** generate Go functions and classification tables from CLDR JSON `cldr-core/supplemental/plurals.json` + `ordinals.json` + `pluralRanges.json` via codegen to `internal/cldr/plural/`.

> **Why**:
> 1. generated-reference `intl-pluralrules/scripts/plural-rules-compiler.ts` has compiled the plural DSL into a directly executed function on the TS side; the Go side maintains the same product form and does not interpret the DSL at runtime.
> 2. Consistent with SPEC 50 §"embed-only / no runtime I/O" - data is the Go source.
> 3. Compiled function zero-allocation query vs runtime interpreter has 2–3 times more overhead.
> 4. `golang.org/x/text/feature/plural` is not available (see §3.2 for details).
>
> **Rejected**:
> - **Runtime interpreter**: DSL parsing + tree-walk on the hot path is significantly slower than the compiled function; and a robust DSL parser needs to be written.
> - **Reuse `golang.org/x/text/feature/plural`**: See §3.2.

### 3.2 Reject reuse `golang.org/x/text/feature/plural` <a id="rejected-x-text-plural"></a>

**MUST** Rules:

1. **FORBIDDEN** runtime plural selection and `tools/gen-plural-rules/` import `golang.org/x/text/feature/plural`. The only accepted import is the `_test.go` benchmark baseline in `pluralrules/benchmark_baseline_test.go`, which exists solely as non-runtime comparison evidence.
2. Reason for rejection (write this SPEC and CLAUDE.md "Forbidden"):
- **CLDR data baseline is uncontrollable**: The data version of `golang.org/x/text/feature/plural` is out of sync with the CLDR 48.1.0/formatjs baseline pinned by go-intl. CLDR 33+ rewrote the plural rules of Welsh/Cymraeg, Hebrew, and Polish; CLDR 41+ changed the Russian ordinal. The data baseline is inconsistent, that is, not equal to fixture bytes.
- **Missing c/e operand**: `plural.Select(t, scale, digits)` has only two parameters and cannot carry the `e` operand of `Intl.NumberFormat` notation=compact. `1 thousand` vs `1K` for minority languages (Polish, Czech) must have `e` to distinguish plural categories.
- **Package annotation UNDER CONSTRUCTION**:pkg.go.dev explicitly indicates "This package is UNDER CONSTRUCTION..."; it cannot be used as a production dependency.

> **Why**: `x/text/feature/plural` is neither accurate nor complete under our target year (2026), and there is no visible path to be fixed (CLDR data updates are controlled by the x/text team cadence, decoupled from ECMA-402 conformance).

### 3.3 Codegen tool (`tools/gen-plural-rules/`)

**MUST** Rules:

1. The codegen entrance **MUST** be located in `tools/gen-plural-rules/main.go`, an independent Go module (independent `go.mod`, which does not pollute the main module dependency graph).
2. codegen **MUST** remain stdlib-only: read CLDR JSON with `encoding/json`, construct the output with deterministic strings, and finally format it with `go/format`; **disallow** `dave/jennifer` or other codegen frameworks.
3. The input **MUST** be pinned CLDR JSON: `cldr-core/supplemental/plurals.json` + `ordinals.json` + `pluralRanges.json`, version pinned through [SPEC 50 §Version Pin](./50-cldr-data.md#version-pin).
4. Output location (each emitted with the generated header; hand-written `rules.go` and `doc.go` in the same package are not generator outputs):
   - `internal/cldr/plural/cardinal_rules.go`
   - `internal/cldr/plural/ordinal_rules.go`
   - `internal/cldr/plural/range_rules.go`
   - `internal/cldr/plural/categories.go`
   - `internal/cldr/plural/supported.go`
5. Each locale outputs an independent Go function; `CardinalRule` / `OrdinalRule` switch index to a specific function.
6. **Disable** output of unused operand expressions (corresponding to generated-reference `should-emit-*.ts`, only the operand judgment actually referenced by rule is generated).
7. The active locale profile **MUST** have both a cardinal and ordinal source rule after CLDR parent fallback. Generation fails with the missing family and locale when either is absent; it must not silently omit that locale.

```go
// Internal form of generator (schematic, no implementation)
package main

func parsePluralRules(path string) (cardinal, ordinal map[string][]Rule, err error)
func parsePluralRanges(path string) (map[string]map[RangeKey]Category, error)
func renderRuleFile(kind string, rules map[string][]Rule) string
func renderRangeFile(ranges map[string]map[RangeKey]Category) string
func renderCategoriesFile(cardinal, ordinal map[string][]Rule) string
```

### 3.4 Codegen output form

**MUST** Rules:

1. Each locale has a lowercase regular function named `cardinalEn` / `cardinalEnUS` / `ordinalEn`, which are stable Go identifiers.
2. The function body **MUST** be an if-else chain, directly corresponding to CLDR DSL translation, without runtime analysis:
   ```go
// cardinal_rules.go (fragment; generated by tools/gen-plural-rules)
   func CardinalRule(loc string) (func(pluralop.OperandsRecord) pluralop.Category, bool) {
       switch loc {
       case "en":
           return cardinalEn, true
       }
       return nil, false
   }

   func cardinalEn(o pluralop.OperandsRecord) pluralop.Category {
       if o.I == 1 && o.V == 0 {
           return pluralop.One
       }
       return pluralop.Other
   }

   func cardinalPl(o pluralop.OperandsRecord) pluralop.Category {
       if o.I == 1 && o.V == 0 {
           return pluralop.One
       }
       if o.V == 0 && o.I%10 >= 2 && o.I%10 <= 4 && (o.I%100 < 12 || o.I%100 > 14) {
           return pluralop.Few
       }
       if o.V == 0 && o.I != 1 && (o.I%10 == 0 || o.I%10 == 1) {
           return pluralop.Many
       }
       return pluralop.Other
   }
   ```
3. Functions **MUST** be fully concurrency safe (pure reading OperandsRecord, no mutable state).
4. The header of each codegen file **MUST** contain the `// Code generated by tools/gen-plural-rules; DO NOT EDIT.` header + CLDR version number; it is prohibited to generate timestamps to ensure the same input and output byte-stable.
5. `CardinalRule` and `OrdinalRule` return `(nil, false)` for an unknown data locale. The hand-written `Rule(loc, typ)` wrapper turns that state into a constructor/data error. English or always-`other` fallback is forbidden because it would advertise fabricated locale behavior.

### 3.5 PluralRanges data <a id="plural-ranges"></a>

**MUST** Rules:

1. `pluralRanges.json` data **MUST** be in the same pipeline codegen as `plurals.json` / `ordinals.json`, and output to `internal/cldr/plural/range_rules.go`.
2. Form:
   ```go
   type rangeKey struct {
       Start, End pluralop.Category
   }

   func Range(loc, typ string, start, end pluralop.Category) (pluralop.Category, bool) {
       if typ != "cardinal" {
           return pluralop.Other, false
       }
       ranges, ok := cardinalRanges[loc]
       if !ok {
           return pluralop.Other, false
       }
       result, ok := ranges[rangeKey{start, end}]
       return result, ok
   }

   var cardinalRanges = map[string]map[rangeKey]pluralop.Category{
       // ...
   }
   ```
3. accessor:`Range(loc, typ string, start, end Category) (Category, bool)`, bool indicates whether the table is hit (used for fallback decision-making).

---

## 4. Compact Operand Contract <a id="compact-operand-contract"></a> <a id="selectformatted"></a>

Public `PluralRules` compact notation and NumberFormat compact suffix selection are separate observable contracts.

- Public `pluralrules.PluralRules.Select` uses the source decimal string plus the selected compact exponent.
- NumberFormat compact suffix selection uses the scaled display decimal plus the selected compact exponent; see [SPEC 20 §4.1 Compact Notation and Plural Operand Contract](./20-numberformat.md#41-compact-notation-and-plural-operand-contract).

Both contracts share `internal/plural.GetOperands` for operand construction, but they pass different formatted decimal strings by design.

### 4.1 Signature

```go
package plural // internal/plural, imported as pluralop

func GetOperands(formatted string, exponent int) OperandsRecord
```

```go
package plural

func CardinalRule(loc string) (func(pluralop.OperandsRecord) pluralop.Category, bool)
```

### 4.2 Algorithm

**MUST** Rules:

1. Public PluralRules compact selection **required**:
   ```text
   1. exponent       := ComputeCompactExponent(locale, compactDisplay, sourceValue, digitOptions)
   2. formatted      := FormatNumericToString(sourceValue).String
   3. ops            := GetOperands(formatted, exponent)
   4. category       := PluralRuleSelect(localeTag, type, ops)
   ```
2. Public PluralRules compact selection **MUST** follow native Node/V8 behavior when FormatJS diverges. Node v26 compact fixtures are the product witness for this boundary.
3. NumberFormat compact suffix selection **required**:
   ```text
   1. value, exponent := ComputeExponent(...)
   2. formatted      := FormatNumericToString(value / 10^exponent).String
   3. ops            := GetOperands(formatted, exponent)
   4. rule           := plural.CardinalRule(localeTag)
   5. category       := rule(ops)
   6. pattern        := CompactPattern(numberingSystem, compactDisplay, exponent, category)
   ```
4. In every compact path, `OperandsRecord.C` and `OperandsRecord.E` **MUST** equal the selected compact exponent.
5. **Public API does not increase**: `SelectFormatted` / `ResolvePlural` do not appear as public surface; NumberFormat completes suffix selection through the combination of internal package and generated rule function.

> **Why**: The stable point is operand generation and generated CLDR rule evaluation, not a JS-style `SelectFormatted` method. Public PluralRules and NumberFormat compact formatting pass different decimal strings because they expose different observable operations.
>
> **Rejected**: Public `PluralRules.SelectFormatted` - This is not an ECMA-402 public API, nor is it a language that Go users need.
> **Rejected**: Let NumberFormat copy plural DSL or rules table - rules are only allowed from `internal/cldr/plural` codegen.

---

## 5. SelectRange algorithm

**MUST** rules (corresponding to ECMA-402 §16.5.4 three-step algorithm + generated-reference `ResolvePluralRange.ts`):

```text
function SelectRange(start, end Value) Category:
    if start or end is NaN:
        return ErrInvalidValue

    sResult := ResolvePlural(start) // Infinity categories are Other
    eResult := ResolvePlural(end)

// Step 1: Formatted string equality → return sCat (avoid "1–1" going to the corner of the range table)
    if sResult.formatted == eResult.formatted:
        return sResult.category

// Step 2: locale does not have pluralRanges data → fall back to eCat(end-class)
    rangeMap, ok := pluralRanges[localeData]
    if !ok:
        return eResult.category

// Step 3: Check pluralRanges["${sCat}_${eCat}"]; fall back to eCat if missed
    cat, ok := rangeMap[{sResult.category, eResult.category}]
    if !ok:
        return eResult.category
    return cat
```

**MUST** Rules:

1. The three-step sequence **MUST** strictly follow ECMA-402 §16.5.4.
2. "Formatted string equality" determination **MUST** pass `FormatNumericToString(start) == FormatNumericToString(end)` string comparison, **not** pass `decimal.Cmp`. Mathematical equality alone is insufficient because digit options can make distinct values visibly equal (for example, `1.1` and `1.2` with `maximumFractionDigits=0`); conversely, `1`, `1.0`, and `1.00` are the same mathematical input and must not differ solely by parser scale.
3. When there is no `pluralRanges` data in the locale, it **MUST** fall back to `eCat`(end-class), and **is prohibited** from falling back to `sCat` or return an error.
4. `rangeMap` misses `(sCat, eCat)` **MUST** fallback to `eCat`, and is **disabled** from automatically trying heuristic fallbacks such as `(sCat, Other)` or `(Other, eCat)`.
5. A NaN start or end **MUST** return `ErrInvalidValue`. Positive and negative infinity **MUST** resolve to `Other` with their special-value formatted strings and continue through the same equality and range-category steps.

> **Why**: Step 1 short-circuiting "1–1" is the solidified behavior of Generated reference; step 2 falling back to end-class is explicitly specified by ECMA-402 §16.5.4; heuristic fallback will introduce inconsistencies with Generated reference.
>
> **Rejected**: Math comparison short-circuiting step 1 - inconsistent with reference behavior.
> **Rejected**: Error (no ranges data) - Consistency conflict with `Select` which does not return error.

---

## 6. ResolvedOptions

```go
type ResolvedOptions struct {
    Locale                   locale.Locale
    Type                     Type
    Notation                 Notation
    CompactDisplay           *CompactDisplay
    MinimumIntegerDigits     int
    MinimumFractionDigits    *int
    MaximumFractionDigits    *int
    MinimumSignificantDigits *int
    MaximumSignificantDigits *int
    PluralCategories         []Category
    RoundingIncrement        int
    RoundingMode             RoundingMode
    RoundingPriority         RoundingPriority
    TrailingZeroDisplay      TrailingZeroDisplay
}
```

**MUST** Rules:

1. The field order **MUST** be consistent with the ECMA-402 §16.4.5 spec order.
2. `PluralCategories` **MUST** return the category actually defined by the locale (reverse check from the `cardinal_rules.go` / `ordinal_rules.go` per-locale table of codegen), **disabled** from returning all 6 categories of hard-coded lists.
3. Digit resolved-option fields **MUST** use pointers and directly project the slots populated by `SetNumberFormatDigitOptions`. Fraction-digit properties are non-nil for fraction-digits, more-precision, and less-precision; significant-digit properties are non-nil for significant-digits, more-precision, and less-precision. The `auto` branches omit only the unused family. PluralRules must not maintain a second slot-presence policy.
4. `CompactDisplay` **MUST** be nil unless `Notation == CompactNotation`; constructor validation still rejects invalid `compactDisplay` values even when `Notation` is standard.
5. **MUST** return an immutable snapshot (value type); pointer-backed resolved scalars must not expose formatter state to callers.
6. JSON field names and `omitempty` behavior **MUST** comply with [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Other Constructors](./73-json-records.md#other-constructors).

---

## 7. Error model

**MUST** Rules:

1. **MUST** pass root sentinel classification: `gointl.ErrInvalidOption` and `gointl.ErrInvalidValue`; this package no longer re-exports formatter-owned sentinels.
2. Construction-time error **MUST** match root `gointl.ErrInvalidOption` and expose `*gointl.Error` structured context.
3. **BANNED** `panic` any user path.
4. A malformed decimal bridge string or a NaN `SelectRange` endpoint **MUST** return a structured error matching `gointl.ErrInvalidValue`. `Select` accepts all ECMA-402 mathematical values, and infinity is valid in `SelectRange`.
5. Public error text follows SPEC 12's `expected ...; got ...` rule and must not expose abstract operation names.

---

## 8. Performance Telemetry

Benchmark numbers guide profiling and prioritization; they do not override ECMA-402 correctness or act as standalone merge blockers(SPEC 71).

**MUST** Rules:

1. Cardinal, ordinal, and range benchmarks stay in `task bench` telemetry.
2. Integer select hot-path allocation counts are tracked with `b.ReportAllocs()`.
3. Performance work must not change operands, compact exponent handling, or CLDR range semantics.

> **Why**: The messageformat-go `:plural` function executes `Select` N times for each message containing complex variables; benchmark telemetry keeps this hot path visible without creating a merge gate.

---

## 9. Forbidden

- **FORBIDDEN** import `golang.org/x/text/feature/plural` from runtime plural selection or codegen; the benchmark-only baseline exception must stay in `_test.go` and must not drive behavior.
- **FORBIDDEN** Introducing `dave/jennifer` or other codegen frameworks - the current scale uses stdlib JSON reading + deterministic string output + `go/format`.
- **BANNED** runtime DSL interpreter - must codegen.
- **BANNED** `OperandsRecord.N/I/F/T` uses `float64` or exposes a big-number type - must use exact `OperandValue`.
- **BANNED** `OperandsRecord` types placed in `pluralrules/` public package - must be placed in `internal/plural/plural.go`.
- **BANNED** Restore the public `Select(any)` / `SelectRange(any, any)` coercion API.
- **BANNED** `SelectFormatted` / `ResolvePlural` exposed as public API - NumberFormat compact path selection via internal operand builder and generated cardinal rule.
- **FORBIDDEN** NumberFormat Copy plural DSL or rule table - plural category is only allowed from `internal/cldr/plural` codegen.
- **Disabled** Mathematical comparison short-circuit `SelectRange` step 1 (string comparison required).
- **BANNED** `SelectRange` heuristic fallback (`(sCat, Other)` / `(Other, eCat)` etc.) - must strictly fallback to `eCat`.
- **BANNED** `PluralCategories` hardcodes all 6 class lists - must be looked up from codegen data.
- **BANNED** `panic` any user path.
- **disable** codegen from outputting unused operand expressions (should-emit optimization).

---

## 10. Acceptance Ledger

SPEC 40 is accepted when the generator, public constructor behavior, compact
semantics, and conformance fixtures agree with the ECMA-402/CLDR boundary.

| Contract | Evidence | Status |
|----------|----------|--------|
| `tools/gen-plural-rules/` remains an independent module and emits CLDR 48.1.0 generated files with stable headers. | `tools/gen-plural-rules/go.mod`; `tools/gen-plural-rules/main.go`; `internal/cldr/plural/cardinal_rules.go`; `internal/cldr/plural/ordinal_rules.go`; `internal/cldr/plural/range_rules.go`; `internal/cldr/plural/categories.go`; `internal/cldr/plural/supported.go` | Satisfied |
| Every active locale has generated cardinal and ordinal rules; missing generation input and missing runtime rule lookup fail with locale/family context instead of defaulting. | `tools/gen-plural-rules/main_test.go`; `internal/cldr/plural/rules.go`; `internal/cldr/plural/rules_test.go`; constructor tests | Satisfied |
| Codegen does not depend on `dave/jennifer`; runtime and codegen do not depend on `golang.org/x/text/feature/plural`. | `rg "dave/jennifer" tools/gen-plural-rules`; `rg "x/text/feature/plural" internal/plural tools/gen-plural-rules pluralrules/*.go` | Satisfied |
| The `pluralrules/benchmark_baseline_test.go` `x/text/feature/plural` import is retained only as benchmark comparison evidence and is outside runtime/codegen acceptance. | `pluralrules/benchmark_baseline_test.go` | Accepted exception |
| Cardinal, ordinal, range, non-finite select/range behavior, NaN range errors, reversed range, rounded equality, and resolved category behavior are covered by package tests and generated fixtures. | `pluralrules/pluralrules_test.go`; `pluralrules/range_test.go`; `pluralrules/options_test.go`; `pluralrules/conformance_unified_test.go`; `pluralrules/testdata/conformance/formatjs/index-test-ts.json` | Satisfied |
| Public PluralRules compact notation follows Node v26 source-decimal-plus-exponent behavior, including million-scale `many` fixtures. | `pluralrules/testdata/conformance/node-v26/compact.json`; `pluralrules/testdata/xfail.json`; `compact_contract_test.go` | Satisfied |
| NumberFormat compact suffix selection remains internal and generated-rule based; no public `SelectFormatted` or internal `ResolvePlural` helper exists. | `numberformat/notation.go`; `internal/plural/plural.go`; absence of `SelectFormatted` / `ResolvePlural` in Go source | Satisfied |
| `OperandsRecord` remains the single internal operand record bridge. | `internal/plural/plural.go`; `pluralrules/pluralrules_test.go` | Satisfied |
| Race, vet, and generated-data byte stability gates pass for the package and generated CLDR output. | `go test -race ./pluralrules/...`; `go vet ./pluralrules/...`; `task data:check` | Required verification |
| PluralRules cardinal, ordinal, and range benchmarks remain per-surface and non-blocking. | `pluralrules/benchmark_test.go`; `SPECS/71-benchmark.md`; `task bench` | Satisfied |
| CLDR bumps are visible as generated-data diffs and must be reviewed with source CLDR version changes. | `internal/cldr/plural/*`; `tools/gen-plural-rules/main.go`; `task data:check` | Required on CLDR update |

---

## 11. References

### Primary

- `.references/formatjs/packages/intl-pluralrules/scripts/plural-rules-compiler.ts` — plural DSL → JS function string (the Go end outputs the equivalent Go function)
- `.references/formatjs/packages/intl-pluralrules/index.ts` — `PluralRuleSelect` / `PluralRuleSelectRange`
- `.references/formatjs/packages/intl-pluralrules/tests/index.test.ts` — main conformance fixture
- `.references/formatjs/packages/ecma402-abstract/PluralRules/GetOperands.ts` — `OperandsRecord` type definition
- `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePlural.ts` — Format-then-GetOperands
- `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePluralRange.ts` — selectRange three steps
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/format_to_parts.ts:262/304/316/331` — compact path feed plural

### Secondary

- CLDR release notes — Plural rule adjustment records (used for version pinning alignment)
- `pkg.go.dev/golang.org/x/text/feature/plural` — Counterexample: UNDER CONSTRUCTION + None c/e
- `golang.org/x/text/internal/cldr` — Counterexample: data baseline is not controlled by go-intl
- `.references/intl/intl.go` — translate-agent/intl (no PluralRules implementation, but codegen mode can be referred to)
- `.references/ext/src/ecma402/plural_rules.c` — PHP/ICU `PluralRules::forLocale` path
- CLDR `cldr-core/supplemental/plurals.json` + `ordinals.json` — cardinal / ordinal authoritative source
- CLDR `cldr-core/supplemental/pluralRanges.json` — authoritative source for pluralRanges

### Project Cross-References

- [SPEC 12 §Abstract Ops](./12-abstract-operations.md) — shared decimal boundary / `ErrInvalidOption`
- [SPEC 10 §Locale structure](./10-locale.md) — `Locale` input parameter type
- [SPEC 20 §Compact Notation](./20-numberformat.md) — NumberFormat compact path operand contract (this SPEC is consistent with its text)
- [SPEC 21 §Decimal API](./21-number-math.md#decimal-api) — `decimal.Decimal` input parameter type
- [SPEC 50 §Codegen](./50-cldr-data.md#codegen) — Both `tools/gen-cldr` and `tools/gen-plural-rules` use the stdlib-only deterministic codegen constraint
- [SPEC 50 §Version Pin](./50-cldr-data.md#version-pin) — CLDR version lock (48.1.0)
- [SPEC 60](./60-facade.md) — root namespace ownership; root `intl.SelectPlural*` one-shot helpers are outside the long-term public surface.
- [SPEC 71](./71-benchmark.md) — non-blocking performance telemetry
