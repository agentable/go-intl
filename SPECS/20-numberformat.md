# SPEC 20 — NumberFormat

> **Status:** Draft (2026-05-08)
> **Closes Open Question:** SPECS/00 §8 Q2 (Functional Options vs Config Struct)
> **Owner:** `numberformat/`
> **Reference contract:** `.references/ecma402/spec/numberformat.html` first, then `formatjs/packages/intl-numberformat/` + `formatjs/packages/ecma402-abstract/NumberFormat/`

## Overview

Defines the `numberformat.NumberFormat` public API, construction-time validation, typed format / parts / range behavior, and the internal contract of NumberFormat ↔ PluralRules on compact notation. The active scope must support all style/notation combinations of ECMA-402 §15 (ES 2025): `decimal | percent | currency | unit` × `standard | scientific | engineering | compact`.

This SPEC does not redefine:
- `Decimal` type and seven rounding modes → see [SPEC 21 §Decimal API](./21-number-math.md#decimal-api)
- `Locale` and `Locale.NumberingSystem()` → see [SPEC 10 §Locale structure](./10-locale.md)
- `OperandsRecord` → see [SPEC 40 §OperandsRecord](./40-pluralrules.md#operandsrecord)
- CLDR data schema(`numbers.json` / `currencies.json` / `units.json`)→ see [SPEC 50 §Schema](./50-cldr-data.md#schema)
- General entrance to abstract operations → See [SPEC 12 §Abstract Ops](./12-abstract-operations.md)

---

## 1. Public API

### 1.1 Construction and Option

```go
package numberformat

type NumberFormat struct{ /* Immutable; contains resolved + apd.Context copy + plural handle */ }

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
    MinimumIntegerDigits *int
    MinimumFractionDigits *int
    MaximumFractionDigits *int
    MinimumSignificantDigits *int
    MaximumSignificantDigits *int
    RoundingIncrement    *int
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

func New(locales locale.List, opts Options) (*NumberFormat, error)

type Value struct{ /* opaque */ }

func Int(v int64) Value
func Uint(v uint64) Value
func Float(v float64) Value
func BigInt(v *big.Int) Value
func BigFloat(v *big.Float) Value
func Decimal(s string) (Value, error)

func (f *NumberFormat) Format(v Value) string
func (f *NumberFormat) FormatToParts(v Value) []Part
func (f *NumberFormat) FormatRange(start, end Value) string
func (f *NumberFormat) FormatRangeToParts(start, end Value) []RangePart

func (f *NumberFormat) ResolvedOptions() ResolvedOptions
```

**MUST** Rules:

1. `New` **MUST** complete all option verification during the construction period, and `error` will be returned if it fails.
2. `New` accepts a `Options` value. `New(locales, Options{})` is equivalent to JS passing an empty options object or omitting options; multiple options objects are not Go API shapes and are rejected by the compiler.
3. `Format` / `FormatToParts` / range methods **MUST** accept opaque `Value`; the caller expresses ECMA-402 numeric input through typed constructors.
4. `Decimal` **MUST** return `ErrInvalidValue` for malformed decimal; `"NaN"` / `"Infinity"` / `"-Infinity"` are legal decimal strings.
5. `NumberFormat` is an immutable value; all methods on `*NumberFormat` must be concurrency-safe.
6. `ResolvedOptions` **MUST** return an immutable snapshot (value type); multiple calls return equal results.

> **Why**: Construction-time verification centrally handles configuration errors; runtime numeric type boundaries are expressed by `Value` constructors. The JS side `format(NaN)` does not throw, so `Format(numberformat.Float(math.NaN()))` still returns `"NaN"`; but the failure of Go's decimal string parsing is a caller error and must be exposed through `Decimal`'s `ErrInvalidValue`.
>
> **Rejected**: public `Format(v any)` / `FormatToParts(v any)` - Mix unsupported input, malformed string, NaN display and user errors in the same hot path.

### 1.2 Typed Options(Close §8 Q2)

**DECISION**: Active scope public APIs **MUST** take a single typed `Options` value. `numberformat.New(locale.MustParseList("en-US"), numberformat.Options{Currency: numberformat.CurrencyCode("USD"), MaximumFractionDigits: gointl.Int(2)})`.

> **Why**:
> 1. The correct value should be discoverable at the IDE and compiler level. `CurrencyStyle`, `CompactNotation`, `HalfEvenRoundingMode` are better suited as ten-year public APIs than bare strings.
> 2. `Options` is a common value, which can be inspected comparatively, bridged serializably, and cache key can be stably generated; functional option closure cannot do these.
> 3. The scalar field can be omitted and a pointer is used to express presence, so `MinimumFractionDigits: gointl.Int(0)`, `MinimumIntegerDigits: gointl.Int(1)`, `RoundingIncrement: gointl.Int(1)` can be distinguished from unset.
>
> **Rejected**:
> - **Functional Options**: Hidden state, not serializable, not statically discoverable, and let the cache key depend on the closure execution result.
> - **Small set-bit option sub-value**: Input and `ResolvedOptions` use two sets of presence expressions, allowing the caller to learn them twice.
> - **Builder chain**: The verification timing is scattered, which conflicts with the goal of "one-time aggregation during construction".

**MUST** Rules:

1. The `Options` field name **MUST** correspond to the Go form of the ECMA-402 §15.4.1 option name.
2. Enumerated fields **MUST** use named types and constants in the package; callers are prohibited from passing bare strings to select core enumerations.
3. `CurrencyCode(code string)` **MUST** normalize ISO 4217 currency code; `UnitIdentifier(unit string)` **MUST NOT** normalize case. ECMA-402 unit identifiers are exact canonical lowercase strings; `"METER"` must be rejected like native `Intl.NumberFormat`.
4. `MinimumIntegerDigits`, fraction digits, significant digits, `RoundingIncrement` **MUST** use the `*int` field to express the explicitly set state; the constructor must copy pointee into the internal config and must not save the caller pointer.
5. `Options{}` The zero value **MUST** be equivalent to the ECMA-402 default option.
6. `New` accepts a `Options` value; multiple options merge is not an ECMA-402 behavior, nor is it an exposed Go API shape.
7. The verification is concentrated on `New`.

### 1.3 Calling example

```go
nf, err := numberformat.New(locale.MustParseList("zh-Hant-TW"),
    numberformat.Options{
        Notation:       numberformat.CompactNotation,
        CompactDisplay: numberformat.ShortCompactDisplay,
    })
// nf.Format(numberformat.Int(98765)) == "99,000"
```

```go
nf, _ := numberformat.New(locale.MustParseList("en-US"),
    numberformat.Options{
        Style:    numberformat.CurrencyStyle,
        Currency: numberformat.CurrencyCode("USD"),
    })
// nf.Format(numberformat.Float(1234.5)) == "$1,234.50"
```

### 1.4 Option parameter type — typed ECMA-402 values

**MUST** Rules:

1. All options taken from the union of ECMA-402 string literals (`style` / `notation` / `compactDisplay` / `currencyDisplay` / `currencySign` / `unitDisplay` / `signDisplay` / `useGrouping` / `roundingPriority` / `roundingMode` / `trailingZeroDisplay` / `localeMatcher`) **MUST** be carried in a named type, the underlying kind remains `string`, and the serialization form is consistent with the JS resolvedOptions string.
2. `numberingSystem` is still `string`, because it is a Unicode extension type, not a small enumeration.
3. `Currency` **MUST** be constructed through `CurrencyCode` to avoid scattered case rules at the call point; `Unit` **MUST** pass the canonical ECMA-402 unit identifier, and lowercase fallback must not be done in the constructor.
4. The `false` value of `useGrouping` is expressed by `UseGroupingFalse` on the Go side, and the underlying layer is still serialized as `"false"`.
5. Verification **MUST** be completed centrally in `New`. Failure to wrap `ErrInvalidOption` and display the value passed in by the user.

> **Why**: typed value still retains ECMA-402 strings as wire/resolved form, but improves the call point from "guessing strings" to "selecting explicit constants". The conformance fixture and messageformat-go adapter can do a mapping at the boundary without downgrading the internal public API to JSON form.

---

## 2. Option pipeline (InitializeNumberFormat)

The public API inside `New` must retain the observable semantics of `formatjs/packages/ecma402-abstract/NumberFormat/InitializeNumberFormat.ts`, but the Go side directly consumes typed `Options` without going through `GetOption(map[string]any)`:

```text
1. localeMatcher       := typed Options.LocaleMatcher, default best-fit
2. numberingSystem := typed Options.NumberingSystem, check Unicode type=numbers
3. resolvedLocale      := ResolveLocale(availableLocales,requestedLocales,localeMatcher,relevantExt={"nu"},...)
4. SetNumberFormatUnitOptions(nf, opts)            // style + currency/{...}/unit/{...}
5. notation            := typed Options.Notation, default standard
6. mnfdDefault, mxfdDefault is determined by style:
    decimal       => 0,3
    percent       => 0,0
    currency      => CurrencyDigits(currency, currencyData), idem
    unit          => 0,3
Compact and scientific/engineering are adjusted in SetNumberFormatDigitOptions
7. SetNumberFormatDigitOptions(nf, opts, mnfdDefault, mxfdDefault, notation)
8. compactDisplay := typed Options.CompactDisplay, only notation=compact takes effect
9. useGrouping := normalize to always|auto|min2|false
10. signDisplay        := auto|always|exceptZero|negative|never
```

> **Why**: The order of steps is the observable behavior of the ECMA-402 algorithm (different steps produce different error messages); changing the order will destroy byte equality.

### 2.1 SetNumberFormatUnitOptions

**MUST** Rules:

1. When `style="currency"` is used, it must be verified that the `currency` option exists and is in the 3-letter ISO 4217 form. Failure to return `ErrInvalidOption`, and the error message contains currency code + locale.
2. Currency precision **MUST** be obtained through `internal/cldr/currency.Digits(code)` (data comes from CLDR `supplemental/currencyData.json` `fractions` node codegen output). **FORBIDDEN** Embedding ISO 4217 static tables. **BANNED** Introduction of `bojanz/currency`.
3. `style="unit"` **MUST** be verified when the `unit` option is sanctioned simple unit or `<numerator>-per-<denominator>` compound form; the sanctioned list comes from the CLDR `units-constants` codegen.
4. `currencyDisplay` ∈ `code | symbol | narrowSymbol | name`;`currencyDisplay="symbol"` is the ECMA-402 default value.
5. `currencySign` ∈ `standard | accounting`, default `standard`.
6. `unitDisplay` ∈ `short | narrow | long`, default `short`.

> **Rejected**: `bojanz/currency` data is not CLDR straight out and is incompatible with FormatJS byte equality target.

### 2.2 SetNumberFormatDigitOptions

**MUST** rules (corresponding to `formatjs/.../SetNumberFormatDigitOptions.ts` 5-way branch):

```text
hasSd = mnsd|mxsd At least one is set
hasFd = mnfd|mxfd at least one is set
need  = hasSd || hasFd
priority = roundingPriority

case priority='morePrecision' or 'lessPrecision': roundingType=morePrecision/lessPrecision
case hasSd                                      : roundingType=significantDigits
case hasFd                                      : roundingType=fractionDigits
case notation='compact'                         : roundingType=compactRounding
otherwise                                       : roundingType=fractionDigits with mnfd=mnfdDefault, mxfd=mxfdDefault
roundingIncrement!=1 ⇒ roundingType forces fractionDigits and mnfd=mxfd
```

1. `roundingIncrement` **MUST** only accept `VALID_ROUNDING_INCREMENTS`(`1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000, 2000, 2500, 5000`), other values return `ErrInvalidOption`.
2. `roundingMode` **MUST** accept all 9 types of ECMA-402: `ceil | floor | expand | trunc | halfCeil | halfFloor | halfExpand | halfTrunc | halfEven` (see [SPEC 21 §Rounding Modes](./21-number-math.md#rounding-modes) for algorithm implementation).
3. `trailingZeroDisplay` ∈ `auto | stripIfInteger`, default `auto`.
4. `roundingPriority` ∈ `auto | morePrecision | lessPrecision`, default `auto`.

> **Why**: The 5-way branch represents the "precision first" semantics of V3 added to ECMA-402. FormatJS resolves `roundingType` according to this branch, and any offset on the Go side will break byte equality.

### 2.3 Digit Formatting Record

`internal/ecma402/numberformat.FormatNumericToString(d, DigitOptions)` is the only runtime code path for ECMA-402 digit formatting. Both NumberFormat and PluralRules consume the formatted string and rounded numeric value it returns.

**MUST** Rules:

1. `DigitOptions` **MUST** contain the resolved status of `minimumIntegerDigits`, fraction digits, significant digits, `roundingIncrement`, `roundingMode`, `roundingPriority`, `trailingZeroDisplay`; the public formatter package is only responsible for mapping its own config to this structure.
2. `FormatNumericToString` **MUST** return an unlocalized, ungrouped ASCII decimal string, retaining the trailing zeros forced by digit options. Grouping, local numeric notation, currency/unit/percent/compact wrapping can only be done in the formatter package.
3. `FormatNumericToString` **MUST** also return rounded numeric value for use by compact plural, range equality, and PluralRules operands.
4. NumberFormat and PluralRules are **FORBIDDEN** to each duplicate fixed/significant/priority rounding code; any rounding or zero-padding fixes must fall in `internal/ecma402/numberformat` and be shared by both formatters.
5. `FormatNumericToString` **disables** decimal rounding or exponential scaling via `float64`, `strconv.ParseFloat`, `math.Log10`, `math.Pow10`; these operations must be done via [SPEC 21 §Decimal API](./21-number-math.md#decimal-api).

> **Why**: ECMA-402 PluralRules reuses the semantics of NumberFormat digit options; writing a set of rounding in each package will drift on trailing-zero, `roundingPriority` and zero scale. Converging the display string stage into a function allows the conformance fixture to constrain both the NumberFormat output and the PluralRules operands.

---

## 3. ResolvedOptions

```go
// Named string type (the underlying kind must be string, corresponding to the JS resolvedOptions string literal value one-to-one;
// See §1.4 Article 4).
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
MinimumIntegerDigits int // Always present
MinimumFractionDigits *int // nil when roundingType == "significantDigits"
MaximumFractionDigits *int // Same as above
MinimumSignificantDigits *int // nil when roundingType == "fractionDigits"
MaximumSignificantDigits *int // Same as above
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

**MUST** Rules:

1. The field order **MUST** be consistent with the ECMA-402 §15.4.5 spec order (to facilitate field-by-field alignment for conformance testing).
2. `MinimumFractionDigits` / `MaximumFractionDigits` / `MinimumSignificantDigits` / `MaximumSignificantDigits` **MUST** be expressed in `*int`, `nil` means that the attribute is not rendered in `resolvedOptions()`. Hidden rules: `roundingType=="significantDigits"` → frac pair `nil`;`roundingType=="fractionDigits"` → sig pair `nil`;`roundingType=="morePrecision" | "lessPrecision"` → both pairs are presented. The zero value `0` always means "set to 0" and no longer conflicts with "not rendered".
3. The `Locale` field **MUST** be the parsing result after `New`’s internal `ResolveLocale` (including the `-u-nu-...` extension), which may be different from the input `loc`.
4. **MUST** return value type (non-pointer) to ensure that the caller cannot modify the internal state of the formatter.
5. JSON field names and `omitempty` behavior **MUST** comply with [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Intl.NumberFormat](./73-json-records.md#intlnumberformat). ECMA-402 Must-occur properties must not use `omitempty`; branch properties use `*T + omitempty` or value `omitempty` to express JavaScript property absence.

> **Why**: ECMA-402 `resolvedOptions()` is an observable surface stipulated in the specification. Missing fields or wrong order are considered failures by the conformance test.

> **Rejected**: Use `map[string]any` to express ResolvedOptions - type safety is lost, and the messageformat-go bridge requires a secondary assertion.

---

## 4. Formatting main process (PartitionNumberPattern)

`Format` and `FormatToParts` share the same main process `PartitionNumberPattern`, defined in `formatjs/.../PartitionNumberPattern.ts` + `format_to_parts.ts`:

```text
1. x := ToIntlMathematicalValue(value)
2. exponent := 0
3. if notation=scientific|engineering|compact:
       exponent := ComputeExponent(nf, x, localeData)
       x       := x ÷ 10^exponent
4. n := internal/ecma402/numberformat.FormatNumericToString(x, nf.DigitOptions)
                                                 // String + Rounded
5. if x.specialValue: // NaN / ±Inf go to symbol map
       formattedString := getNaN(localeData) or getInfinity(localeData)
   else:
       formattedString := PartitionDigitParts(nf, n, exponent, getNumberingSystem(nf), localeData)
6. Packaging sign / currency / unit / percent / compact-suffix / exponent-symbol
```

**MUST** Rules:

1. `ToIntlMathematicalValue` **MUST** be implemented using [SPEC 21 §ToIntlMathematicalValue](./21-number-math.md#tointlmathematicalvalue); **FORBIDDEN** to convert values via `fmt.Sprintf("%v", value)`.
2. The `String` output by `FormatNumericToString` **MUST** retain trailing zero (forced by `mnfd`) for subsequent OperandsRecord calculation of `v / w / f / t` (see [SPEC 40 §Operands](./40-pluralrules.md#operands)).
3. `NaN / +Inf / -Inf` **MUST** be expressed through `apd.Decimal.Form`, and **FORBIDDEN** to be transferred through `math.IsNaN(float64(...))`.
4. The `[]Part` element `Type` output by `PartitionDigitParts` **MUST** qualify ECMA-402 §15.5.1 + FormatJS extension for a total of 16 enumeration strings: `integer | group | decimal | fraction | currency | percentSign | minusSign | plusSign | nan | infinity | unit | literal | exponentSeparator | exponentMinusSign | exponentInteger | compact | approximatelySign` (strictly aligned with `.references/formatjs/packages/ecma402-abstract/types/number.ts` `NumberFormatPartTypes`; **FORBIDDEN** to use `exponentSymbol`, the canonical name is `exponentSeparator`;`approximatelySign` only appears as part type when the formatting results of both ends of `FormatRange` are the same).

> **Why**: This is the key operator for conformance byte equality; any step skipped will be detected by the `FormatJS` `format_to_parts.test.ts` fixture.

### 4.1 Compact Notation and PluralRules Contract

**MUST (Same text as [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract))**:

1. NumberFormat Under the `notation = compact` path, **MUST** use the rounded string showing the number and compact exponent to select the plural category:
   ```go
   ops := ecma402pr.GetOperands(formattedDisplayDecimal, exponent)
   rule, ok := plural.CardinalRule(localeTag)
   cat := rule(ops)
   ```
2. The semantics of the passed-in parameter `(formattedDisplayDecimal string, exponent int)` **MUST** be:
- `formattedDisplayDecimal` = `FormatNumericToString` The rounded decimal string output has been divided by `10^exponent` (the "display number"), and trailing zero is retained.
- compact/scientific exponent determined by `exponent` = `ComputeExponent`.
3. **FORBIDDEN** NumberFormat parses plural DSL itself or holds a public `pluralrules.PluralRules` instance; rule functions only come from `internal/cldr/plural`.
4. `internal/ecma402/pluralrules.GetOperands` is the implementation of operand records: derive `n / i / v / w / f / t` from `formattedDisplayDecimal`, and set `c / e` to `exponent` (see [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract)).

> **Why**: Cross-package contract. The stable boundary of compact formatting is "display number + compact exponent + generated CLDR rule"; this preserves the CLDR `c/e` operand while avoiding exposing interdependencies between formatters.
>
> **Rejected**: Expose `SelectFormatted` or let NumberFormat hold `pluralrules.PluralRules` - this is a shadow of JS internal slot, not Go API.

### 4.2 Sign / Currency / Unit / Percent / Compact Packaging

**MUST** Rules:

1. `signDisplay = "negative"` and `"exceptZero"` are added in ES 2024 and must be implemented.
2. `currencyDisplay = "narrowSymbol"` **MUST** fall back to `"symbol"` when CLDR data lacks narrow form.
3. `currencySign = "accounting"` **MUST** use the CLDR accounting pattern; when the negative sub-pattern exists, the minus sign is consumed by the pattern, and when it does not exist, the explicit sign part is retained.
4. Compact suffix selection **MUST** first determine the plural category according to §4.1, and then check CLDR `numbers.json` `decimalFormats.{short|long}.decimalFormat[length].decimal-format-pattern.<category>`; when category is missing, fall back to `other`.
5. `useGrouping = "min2"` **MUST** only insert groups when the integer part is ≥ 5 bits (aligned FormatJS `useGrouping` implementation).

---

## 5. FormatRange / FormatRangeToParts <a id="5-formatrange--formatrangetoparts"></a>

**MUST** Rules:

1. `FormatRange(a, b)` **MUST** implement ECMA-402 §15.5.7 `FormatNumericRange`: First format both ends separately, and then call `CollapseNumberRange` to merge the same prefix/suffix.
2. `CollapseNumberRange` **MUST** consume the `NumberFormatPart{Type, Value}` sequence, and the equality is determined one by one; the criterion for equality is the **per-package field** (`Type` and `Value` are both equal). **BANNED** The abstract generic `CollapseRange[T]` is shared with DateTimeFormat.
3. Both ends of `Decimal` comparison **MUST** pass [SPEC 21 §Decimal.Cmp](./21-number-math.md#decimal-cmp); **FORBIDDEN** to pass `float64` conversion.
4. Range source **MUST** be limited to ECMA-402 three values: `"startRange" | "shared" | "endRange"`;`approximatelySign` is a part type, not a source.
5. `FormatRange` / `FormatRangeToParts` **MUST** use limited `Value` endpoints; non-limited endpoints output empty strings or nil parts, and must not panic.
6. `a > b` **must not** be locally normalized, transposed or added `~`; formatted in the order of input parameters and collapse range parts.
7. When `FormatNumeric(a) == FormatNumeric(b)`, output shared `approximatelySign` part + shared digital parts (for example, when the maximum fraction digits is 0, `1.1–1.2` outputs `~1`).

> **Why**: NumberFormatPart and DateTimeFormatPart have different fields (`unit | currency | percentSign | exponentInteger` vs `era | year | month | ...`). Although the collapse algorithm has the same structure (removing suffixes), it works on different part fields; FormatJS is also implemented separately.
>
> **Rejected**: Abstract general `CollapseRange[T Part]` generic function - one more layer of indirection, and the "equivalence" semantics of `T` are different between the two packages.

---

## 6. Input type support

**MUST** rules (corresponding to ECMA-402 §15.5.1):

1. Public hot path is **FORBIDDEN** to accept `any`; the caller must construct `numberformat.Value` first, and then call `Format`, `FormatToParts`, `FormatRange` or `FormatRangeToParts`.
2. `Decimal` accepts ECMA-402 `StringNumericLiteral`, such as `"1234.5"` / `"NaN"` / `"Infinity"`; the range method consumes the constructed `Value`.
3. malformed decimal string **MUST** return `ErrInvalidValue`; silent fallback to `"NaN"` is prohibited.
4. The conformance fixture can keep the unexported `formatValue(any)` adapter in the package, but it must not appear in the public API, README or root package.

> **Why**: `fmt.Sprintf("%v", float64)` uses `%g` format, trailing-zero is inconsistent with ECMA-402 ToIntlMathematicalValue; walking Sprintf on the hot path is a significant performance penalty (~150 ns per allocation).

---

## 7. Internal Slot and cache

**MUST** Rules:

1. NumberFormat internal state **MUST** all be calculated and frozen at the time of `New`; the `*NumberFormat` returned by `New` is an immutable snapshot.
2. The CLDR number-domain data (`numberSymbols / patterns / compactDecimalFormats`) **MUST** be pulled once through the resolved `internal/cldr/number.Locale` accessors in `New` and saved to the `NumberFormat` internal slot; the `Format` path **is prohibited** from calling CLDR accessors again.
3. The PluralRules handle **MUST** be lazy constructed in `New` (only constructed in `notation = compact | scientific | engineering`); the `Format` path **forbids** to reconstruct PluralRules.
4. `apd.Context` **MUST** be held as an immutable baseline; the `Rounding` field is modified after copying when `Format` is called (to avoid races).

> **Why**: The `internalSlot` mode of `FormatJS` (checking slot every time `format()`) corresponds to "materialization during construction and read-only during runtime" on Go. This is consistent with the CLAUDE.md "constructor-eager / Format-no-error" rule.

---

## 8. Error model

**MUST** Rules:

1. Construction-time errors **MUST** be the wrapped form of `gointl.ErrInvalidOption`, and expose field names, user-input values, locale and expected-value guidance through `*gointl.Error`.
2. Sentinel **MUST** match root `gointl.ErrInvalidOption`(SPEC 12); this package does not establish a separate independent error category.
3. **BANNED** `panic` any user path; `MustNew` does not exist (user can wrap it in the caller).
4. Runtime fallback (NaN / Infinity / string parsing failure) **must not** return an error, but directly output the fallback string.

```go
// Error form example (signature)
err := ecma402.InvalidOptionError("numberformat", "currency", code, loc.String(), gointl.ErrInvalidOption)
```

---

## 9. Performance Telemetry

Benchmark numbers guide profiling and prioritization; they do not override ECMA-402 correctness or act as standalone merge blockers(SPEC 71).

**MUST** Rules:

1. `BenchmarkNumberFormat_New` and cached decimal benchmarks stay in `task bench` telemetry.
2. `Format(int64)` hot-path allocation counts are tracked with `b.ReportAllocs()`.
3. Performance work must not change digit rounding ownership, parts semantics, or public API shape.

> **Why**: messageformat-go is called within `:number`. Each message may be called N times `Format`; less than 1 μs to retain the message layer SLA.

---

## 10. Forbidden

- **BANNED** Introduce `golang.org/x/text/message` as NumberFormat implementation - missing currency/unit/compact notation, incompatible with FormatJS byte equality.
- **BANNED** Introduce `bojanz/currency` as currency data source - non-CLDR direct output, and comes with ISO 4217 historical data that conflicts with our CLDR nail version.
- **BANNED** Return `error` (ECMA-402 fallback for invalid input) in `Format` / `FormatToParts` paths.
- **BANNED** Use `Format` with `fmt.Sprintf("%v", value)` on hot path - trailing-zero behavior is unaligned + ~150 ns per allocation.
- **FORBIDDEN** NumberFormat parses plural DSL, copies plural rule tables, or selects compact suffix by exposing a `pluralrules.PluralRules` instance; can only call `internal/ecma402/pluralrules.GetOperands` and use `internal/cldr/plural` generated rules.
- **FORBIDDEN** Extract `CollapseNumberRange` into a cross-package generic - Part fields are different.
- **FORBIDDEN** Calling CLDR domain accessors on the `Format` path; CLDR data must be materialized on `New`.
- **BANNED** Builder chaining API(`numberformat.NewBuilder().Currency("USD").Build()`); active scope exposes only `New(locale.MustParseList("en-US"), Options{...})` or equivalent `locale.List` variables.
- **BANNED** Pointer configuration API (`numberformat.New(locales, &Options{...})`) and functional options; the only public configuration form after turning off §8 Q2 is the typed `Options` value.
- **FORBIDDEN** Self-developed `BigDecimal`; all math layers go through [SPEC 21 §Decimal API](./21-number-math.md#decimal-api)(`apd/v3` backend).

---

## 11. Acceptance Criteria

- [ ] `formatjs/packages/intl-numberformat/tests/format_to_parts.test.ts` All fixtures in `numberformat/conformance_unified_test.go` pass (byte-equality).
- [ ] `formatjs/packages/intl-numberformat/tests/notation-compact-zh-TW.test.ts` passes (`format(98765) == "9.9\u842c"`).
- [ ] `formatjs/packages/intl-numberformat/tests/format_range.test.ts` All fixtures passed in `FormatRange`.
- [ ] `go test -race ./numberformat/...` passed (including `TestNumberFormat_ConcurrentFormat` 100 goroutine × 1000 calls).
- [ ] `go vet ./numberformat/...` clean.
- [ ] `New(...)` returns `ErrInvalidOption` wrapped error for unknown currency, the message contains currency + locale.
- [ ] `Format(numberformat.Float(math.NaN()))` returns locale-specific NaN string, does not return error, and does not panic.
- [ ] `numberformat.Decimal("abc")` returns `ErrInvalidValue`.
- [ ] `ResolvedOptions().MinimumFractionDigits` when `Options{MinimumSignificantDigits: gointl.Int(2)}` is passed in alone `== nil`(roundingType=="significantDigits" hides frac); symmetrically, `Options{MaximumFractionDigits: gointl.Int(2)}` when passed in alone is `MinimumSignificantDigits == nil`. Pointer types are unambiguous between the two states.
- [ ] `roundingPriority = "morePrecision" | "lessPrecision"` is observable when the fraction and significant digit options are passed in at the same time, and is not treated as an unsupported option.
- [ ] `Options{CurrencySign: AccountingCurrencySign}` negative USD output `($12.00)` for `en-US`.
- [ ] `Options{CompactDisplay: LongCompactDisplay}` outputs `1.5 thousand` for `en` `1500` + `MaximumFractionDigits: gointl.Int(1)`.
- [ ] compact plural contract use case: `numberformat.New(locale.MustParseList("pl-PL"), numberformat.Options{Notation: numberformat.CompactNotation}).Format(numberformat.Int(1500))` is consistent with `FormatJS` output under `pl-PL` (plural category `few` suffix).
- [ ] The sequence of options pipeline steps passes the `internal/ecma402/numberformat.TestInitializeNumberFormat_StepOrder` test (a trace is generated at each step, aligned with the FormatJS call sequence).
- [ ] `BenchmarkNumberFormat_Decimal_Cached` and `BenchmarkNumberFormat_New` appear in non-blocking `task bench` telemetry.
- [ ] Benchmark reports label NumberFormat as a per-surface package, not root facade cost.

---

## 12. References

### Primary

- `.references/formatjs/packages/intl-numberformat/` — public API form, `format / formatToParts / formatRange / resolvedOptions` behavior
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/InitializeNumberFormat.ts` — option pipeline
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatDigitOptions.ts` — 5-way branch
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatUnitOptions.ts` — currency/unit verification
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/PartitionNumberPattern.ts` — Main process
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/format_to_parts.ts` — Compact path (`:262/304/316/331` line plural call)
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/CollapseNumberRange.ts` — Number-specific collapse
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/CurrencyDigits.ts` — currencyData injection

- `.references/intl/intl.go` — translate-agent/intl(Go precedent, DateTimeFormat-only, NumberFormat no precedent)
- `.references/ext/src/ecma402/currency.c` — PHP/ICU currency data path

### Project Cross-References

- [SPEC 12 §Abstract Ops](./12-abstract-operations.md) — shared validators / digit pipeline / `ErrInvalidOption`
- [SPEC 10 §Locale structure](./10-locale.md) — `Locale.NumberingSystem()`
- [SPEC 21 §Decimal API](./21-number-math.md#decimal-api) — `Decimal` / `apd/v3` backend / seven rounding modes / RoundingPriority / RoundingIncrement / TrailingZeroDisplay algorithm
- [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract) — compact plural operand contract
- [SPEC 50 §Schema](./50-cldr-data.md#schema) — `numbers.json` / `currencies.json` / `units.json` data shape
- [SPEC 60](./60-facade.md) — root namespace ownership; root `intl.FormatNumber*` one-shot helpers are outside the long-term public surface.
- [SPEC 71](./71-benchmark.md) — non-blocking performance telemetry
