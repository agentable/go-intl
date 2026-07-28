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
type Notation string

type Options struct {
    Style                *string
    Currency             *string
    CurrencyDisplay      *string
    CurrencySign         *string
    Unit                 *string
    UnitDisplay          *string
    MinimumIntegerDigits *int
    MinimumFractionDigits *int
    MaximumFractionDigits *int
    MinimumSignificantDigits *int
    MaximumSignificantDigits *int
    RoundingIncrement    *int
    RoundingPriority     *string
    RoundingMode         *string
    TrailingZeroDisplay  *string
    Notation             *string
    CompactDisplay       *string
    UseGrouping          *string
    SignDisplay          *string
    LocaleMatcher        *string
    NumberingSystem      *string
}

func New(locales locale.List, opts Options) (*NumberFormat, error)

type Value struct{ /* opaque */ }

func Int(v int64) Value
func Uint(v uint64) Value
func Float(v float64) Value
func BigInt(v *big.Int) Value
func Decimal(s string) (Value, error)

func (f *NumberFormat) Format(v Value) string
func (f *NumberFormat) FormatToParts(v Value) []Part
func (f *NumberFormat) FormatRange(start, end Value) (string, error)
func (f *NumberFormat) FormatRangeToParts(start, end Value) ([]RangePart, error)

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
> **Rejected**:
> - public `Format(v any)` / `FormatToParts(v any)` - Mix unsupported input, malformed string, NaN display and user errors in the same hot path.
> - public `BigFloat(*big.Float)` - ECMA-402 has Number, BigInt, and string-to-decimal bridges; arbitrary-precision floating point is a Go convenience shape with no native Intl owner. Callers that need exact decimal magnitude should use `Decimal`.

### 1.2 Typed Options(Close §8 Q2)

**DECISION**: Active scope public APIs **MUST** take a single typed `Options` value. `numberformat.New(mustLocaleList("en-US"), numberformat.Options{Currency: gointl.String("USD"), MaximumFractionDigits: gointl.Int(2)})`.

> **Why**:
> 1. The correct value should be discoverable at the IDE and compiler level. `CurrencyStyle`, `CompactNotation`, and `HalfEvenRoundingMode` remain the ten-year vocabulary, and callers pass them through presence-aware `gointl.String(...)` so explicit empty values are still validated instead of disappearing as zero values.
> 2. `Options` is a common value, which can be inspected comparatively, bridged serializably, and cache key can be stably generated; functional option closure cannot do these.
> 3. Scalar fields whose zero value is meaningful or whose explicit empty value must still be validated use pointers to express presence, so `Style: gointl.String(numberformat.CurrencyStyle)`, `LocaleMatcher: gointl.String(numberformat.LookupLocaleMatcher)`, `Currency: gointl.String("USD")`, `Unit: gointl.String("meter")`, `MinimumFractionDigits: gointl.Int(0)`, `MinimumIntegerDigits: gointl.Int(1)`, `RoundingMode: gointl.String(numberformat.HalfEvenRoundingMode)`, `RoundingIncrement: gointl.Int(1)`, and `NumberingSystem: gointl.String("latn")` can be distinguished from unset.
>
> **Rejected**:
> - **Functional Options**: Hidden state, not serializable, not statically discoverable, and let the cache key depend on the closure execution result.
> - **Small set-bit option sub-value**: Input and `ResolvedOptions` use two sets of presence expressions, allowing the caller to learn them twice.
> - **Builder chain**: The verification timing is scattered, which conflicts with the goal of "one-time aggregation during construction".

**MUST** Rules:

1. The `Options` field name **MUST** correspond to the Go form of the ECMA-402 §15.4.1 option name.
2. Enumerated string option fields **MUST** use package named types and constants as vocabulary, but the public `Options` fields are `*string` so nil means omitted and `gointl.String("")` remains an explicit invalid value.
3. The constructor owns string normalization and validation for identifier option fields. `Options.Currency` uses `*string`, is normalized to uppercase during construction, and validates any explicit value even when `Style != CurrencyStyle`; `Options.Unit` uses `*string`, validates any explicit value, and **MUST NOT** normalize case. ECMA-402 unit identifiers are exact canonical lowercase strings; `"METER"` must be rejected like native `Intl.NumberFormat`.
4. String and scalar input options **MUST** use pointer fields to express the explicitly set state: `Style`, `LocaleMatcher`, `NumberingSystem`, `Currency`, `CurrencyDisplay`, `CurrencySign`, `Unit`, `UnitDisplay`, `Notation`, `CompactDisplay`, `UseGrouping`, `SignDisplay`, `RoundingMode`, `RoundingPriority`, `TrailingZeroDisplay`, all digit integer fields, and `RoundingIncrement`; the constructor must copy pointees into the internal config and must not save caller pointers.
5. `Options{}` The zero value **MUST** be equivalent to the ECMA-402 default option.
6. `New` accepts a `Options` value; multiple options merge is not an ECMA-402 behavior, nor is it an exposed Go API shape.
7. The verification is concentrated on `New`.

### 1.3 Calling example

```go
nf, err := numberformat.New(mustLocaleList("zh-Hant-TW"),
    numberformat.Options{
        Notation:       gointl.String(numberformat.CompactNotation),
        CompactDisplay: gointl.String(numberformat.ShortCompactDisplay),
    })
// nf.Format(numberformat.Int(98765)) == "99,000"
```

```go
nf, _ := numberformat.New(mustLocaleList("en-US"),
    numberformat.Options{
        Style:    gointl.String(numberformat.CurrencyStyle),
        Currency: gointl.String("USD"),
    })
// nf.Format(numberformat.Float(1234.5)) == "$1,234.50"
```

### 1.4 Option parameter type — typed ECMA-402 values

**MUST** Rules:

1. Enum-like options taken from the union of ECMA-402 string literals (`style` / `notation` / `compactDisplay` / `currencyDisplay` / `currencySign` / `unitDisplay` / `signDisplay` / `useGrouping` / `roundingPriority` / `roundingMode` / `trailingZeroDisplay`) **MUST** have package named types and constants as vocabulary, while constructor option fields use `*string`; nil means omitted and `gointl.String("")` is invalid.
2. `numberingSystem` uses `*string`, because it is a Unicode extension type whose omitted/default state and explicit empty value are distinct.
3. `Currency` and `Unit` are direct ECMA-402 identifier option values carried as `*string`. Currency case normalization and unit validation belong to `New`; unit lowercase fallback must not be done in the constructor.
4. The `false` value of `useGrouping` is expressed by `UseGroupingFalse` on the Go side, and the underlying layer is still serialized as `"false"`.
5. Verification **MUST** be completed centrally in `New`. Failure to wrap `ErrInvalidOption` and display the value passed in by the user.

> **Why**: typed value still retains ECMA-402 strings as wire/resolved form, but improves the call point from "guessing strings" to "selecting explicit constants". The conformance fixture and messageformat-go adapter can do a mapping at the boundary without downgrading the internal public API to JSON form.

---

## 2. Option pipeline (InitializeNumberFormat)

The public API inside `New` must retain the observable semantics of `formatjs/packages/ecma402-abstract/NumberFormat/InitializeNumberFormat.ts`, but the Go side directly consumes typed `Options` without going through `GetOption(map[string]any)`:

```text
1. localeMatcher       := typed `*Options.LocaleMatcher`, where nil means omitted/default best-fit and `gointl.String("")` is invalid
2. numberingSystem := typed `*Options.NumberingSystem`, where nil means omitted and `gointl.String("")` is an invalid explicit Unicode type
3. resolvedLocale      := ResolveLocale(availableLocales,requestedLocales,localeMatcher,relevantExt={"nu"},...)
4. SetNumberFormatUnitOptions(nf, opts)            // style + presence-aware currency/{...}/unit/{...}
5. notation            := typed `*Options.Notation`, where nil means default standard and `gointl.String("")` is invalid
6. mnfdDefault, mxfdDefault is determined by style:
    decimal       => 0,3
    percent       => 0,0
    currency      => CurrencyDigits(currency, currencyData), idem
    unit          => 0,3
Compact and scientific/engineering are adjusted in SetNumberFormatDigitOptions
7. SetNumberFormatDigitOptions(nf, opts, mnfdDefault, mxfdDefault, notation)
8. compactDisplay := typed `*Options.CompactDisplay`, where nil means default short and `gointl.String("")` is invalid
9. useGrouping := typed `*Options.UseGrouping`, where nil means default auto or compact min2 and `gointl.String("")` is invalid
10. signDisplay        := typed `*Options.SignDisplay`, where nil means default auto and `gointl.String("")` is invalid
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

> **Rejected**: `bojanz/currency` data is not CLDR straight out and is incompatible with fixture byte equality target.

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
5. The `morePrecision` / `lessPrecision` tie-break **MUST** compare the two candidates' ECMA-402 `[[RoundingMagnitude]]` (the decimal place each rounds at), **not** their numeric distance to the source value. The fixed (fraction) candidate is "more precise" when it rounds at a deeper place: `-maximumFractionDigits < e - maximumSignificantDigits + 1`, where `e` is the decimal exponent of the significant candidate's most significant digit. `morePrecision` keeps the fixed candidate exactly when it is more precise, `lessPrecision` when it is not. Result: `NumberFormat(en, {maximumSignificantDigits:2, minimumFractionDigits:3, maximumFractionDigits:3, roundingPriority:"morePrecision"}).Format(1.2)` → `"1.200"`.

> **Why**: The 5-way branch represents the "precision first" semantics of V3 added to ECMA-402. Generated reference resolves `roundingType` according to this branch, and any offset on the Go side will break byte equality.

### 2.3 Digit Formatting Record

`internal/ecma402/numberformat.FormatNumericToString(d, ResolvedDigitOptions)` is the only runtime code path for ECMA-402 digit formatting. Both NumberFormat and PluralRules consume the same constructor-resolved record and the formatted string and rounded numeric value it returns.

**MUST** Rules:

1. `ResolvedDigitOptions` **MUST** contain one `DigitOptions` value plus the resolved `RoundingType`. `DigitOptions` owns the resolved digit bounds, increment, typed `decimal.RoundingMode`, priority, and trailing-zero policy. `SetNumberFormatDigitOptions` parses and validates the rounding mode and selects the branch once during construction; formatting must not infer either value again or fall back from invalid internal state.
2. `FormatNumericToString` **MUST** return an unlocalized, ungrouped ASCII decimal string, retaining only trailing zeros forced by resolved digit options. Lexical zeros in an input such as `"1.50"` are not part of the Intl mathematical value and must not become a second display owner. Grouping, local numeric notation, currency/unit/percent/compact wrapping can only be done in the formatter package.
3. `FormatNumericToString` **MUST** also return rounded numeric value for compact plural selection and PluralRules operands. NumberFormat range equality is decided later from the fully partitioned visible endpoint text, not from this rounded decimal value.
4. NumberFormat and PluralRules are **FORBIDDEN** to each duplicate fixed/significant/priority rounding code; any rounding or zero-padding fixes must fall in `internal/ecma402/numberformat` and be shared by both formatters.
5. `FormatNumericToString` **disables** decimal rounding or exponential scaling via `float64`, `strconv.ParseFloat`, `math.Log10`, `math.Pow10`; these operations must be done via [SPEC 21 §Decimal API](./21-number-math.md#decimal-api).
6. The returned rounded mathematical value **MUST** preserve a negative zero sign when a negative input rounds to zero. Text and range partitioning consume that sign; `-0.1` and `-0.2` with zero fraction digits produce the shared approximate result `~-0`, not `~0`.

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
    Currency                     *string        // nil unless style=currency
    CurrencyDisplay              *CurrencyDisplay
    CurrencySign                 *CurrencySign
    Unit                         *string        // nil unless style=unit
    UnitDisplay                  *UnitDisplay
MinimumIntegerDigits int // Always present
MinimumFractionDigits *int // nil when roundingType == "significantDigits"
MaximumFractionDigits *int // Same as above
MinimumSignificantDigits *int // nil when roundingType == "fractionDigits"
MaximumSignificantDigits *int // Same as above
    UseGrouping                  UseGrouping    // "always" | "auto" | "min2" | "false"
    Notation                     Notation       // "standard" | "scientific" | "engineering" | "compact"
    CompactDisplay               *CompactDisplay // nil unless notation=compact
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
3. `Currency`, `CurrencyDisplay`, `CurrencySign`, `Unit`, `UnitDisplay`, and `CompactDisplay` **MUST** be pointer fields with `omitempty`; `nil` means ECMA-402 did not create that resolved-options property. A valid-but-irrelevant input such as `Options{Currency: gointl.String("USD")}` with decimal style is validated at construction but omitted from `ResolvedOptions`.
4. The `Locale` field **MUST** be the parsing result after `New`’s internal `ResolveLocale` (including the `-u-nu-...` extension), which may be different from the input `loc`.
5. **MUST** return value type (non-pointer) and clone every pointer-backed scalar so that the caller cannot modify the internal state of the formatter.
6. JSON field names and `omitempty` behavior **MUST** comply with [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Intl.NumberFormat](./73-json-records.md#intlnumberformat). ECMA-402 Must-occur properties must not use `omitempty`; NumberFormat branch properties use `*T + omitempty` to express JavaScript property absence.

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
4. n := internal/ecma402/numberformat.FormatNumericToString(x, nf.ResolvedDigitOptions)
                                                 // String + Rounded
5. if x.specialValue: // NaN / ±Inf go to symbol map
       formattedString := getNaN(localeData) or getInfinity(localeData)
   else:
       formattedString := PartitionDigitParts(nf, n, exponent, getNumberingSystem(nf), localeData)
6. Packaging sign / currency / unit / percent / compact-suffix / exponent-symbol
```

**MUST** Rules:

1. `ToIntlMathematicalValue` **MUST** be implemented using [SPEC 21 §ToIntlMathematicalValue](./21-number-math.md#tointlmathematicalvalue); **FORBIDDEN** to convert values via `fmt.Sprintf("%v", value)`.
2. The `String` output by `FormatNumericToString` **MUST** retain trailing zeros forced by resolved digit options (for example `mnfd`) for subsequent OperandsRecord calculation of `v / w / f / t` (see [SPEC 40 §Operands](./40-pluralrules.md#operands)); it must not retain zeros solely because they appeared in the input literal.
3. `NaN / +Inf / -Inf` **MUST** be expressed through `apd.Decimal.Form`, and **FORBIDDEN** to be transferred through `math.IsNaN(float64(...))`.
4. The `[]Part` element `Type` output by `PartitionDigitParts` **MUST** qualify ECMA-402 §15.5.1 + Generated reference extension for a total of 16 enumeration strings: `integer | group | decimal | fraction | currency | percentSign | minusSign | plusSign | nan | infinity | unit | literal | exponentSeparator | exponentMinusSign | exponentInteger | compact | approximatelySign` (strictly aligned with `.references/formatjs/packages/ecma402-abstract/types/number.ts` `NumberFormatPartTypes`; **FORBIDDEN** to use `exponentSymbol`, the canonical name is `exponentSeparator`;`approximatelySign` only appears as part type when the formatting results of both ends of `FormatRange` are the same).
5. `ComputeExponent` **MUST** be a single shared engine (`internal/ecma402/numberformat.computeExponent`) behind scientific, engineering, and compact notation. It scales the value into its notation mantissa, rounds with the resolved digit options, and applies the ECMA-402 post-rounding carry recheck: when rounding pushes the mantissa up a magnitude (for example `999` with `maximumFractionDigits:0` rounds to `1000`), the exponent is re-derived from `magnitude + 1`. Scientific/engineering `ScientificExponent` and compact `ResolveCompactMagnitude` **MUST** both route through this engine; **FORBIDDEN** to keep a second exponent routine that computes the pre-rounding magnitude and skips the carry recheck. Result: scientific `Format(999)` with `maximumFractionDigits:0` → `"1E3"`, `999500` → `"1E6"`, engineering `999500` → `"1E6"`.
6. `Format` and `FormatToParts` **MUST** consume one package-private `[]Part` partition. `Format` only concatenates `Part.Value`; parallel text renderers and integer-only fast paths that duplicate sign, grouping, notation, localization, currency, or unit semantics are forbidden.

> **Why**: This is the key operator for conformance byte equality; any step skipped will be detected by the `generated-reference` `format_to_parts.test.ts` fixture.

### 4.1 Compact Notation and Plural Operand Contract

NumberFormat has a narrower compact-plural contract than public
`Intl.PluralRules`. It selects the compact suffix for the visible NumberFormat
output; it does not call a public `pluralrules.PluralRules` instance.

**MUST** Rules:

1. Under `notation = compact`, NumberFormat **MUST** use the rounded display decimal and compact exponent to select the CLDR compact suffix category:
   ```go
   ops := pluralop.GetOperands(formattedDisplayDecimal, exponent)
   rule, err := plural.Rule(dataLocale, "cardinal") // resolved in New
   cat := rule(ops)
   ```
2. `formattedDisplayDecimal` is the unlocalized decimal string after scaling the source value by `10^exponent` and applying the resolved digit options. Forced trailing zeros are retained.
3. `exponent` is the compact exponent selected from generated compact pattern data, including the rounded-carry case where a value moves into the next compact magnitude.
4. NumberFormat **MUST** resolve the exact generated cardinal rule for its constructor data locale in `New`. A missing rule is a constructor/data error; runtime formatting must not substitute English or an always-`other` rule. NumberFormat must not parse the plural DSL, copy generated plural rules, or hold a public `pluralrules.PluralRules` instance.
5. Public `pluralrules.PluralRules` compact notation is a different observable operation: it selects the public plural category from the source decimal string plus the compact exponent. That contract is owned by [SPEC 40 §Compact Operand Contract](./40-pluralrules.md#compact-operand-contract).

> **Why**: Compact suffix selection is a NumberFormat-internal formatting step. The stable boundary is "display decimal + compact exponent + generated CLDR rule"; public PluralRules has its own native-observable compact selection semantics.
>
> **Rejected**: Expose `SelectFormatted` or let NumberFormat hold `pluralrules.PluralRules` - this is a shadow of JS internal slot, not Go API.

### 4.2 Sign / Currency / Unit / Percent / Compact Packaging

**MUST** Rules:

1. `signDisplay = "negative"` and `"exceptZero"` are added in ES 2024 and must be implemented.
2. `currencyDisplay = "narrowSymbol"` **MUST** fall back to `"symbol"` when CLDR data lacks narrow form.
3. `currencySign = "accounting"` **MUST** use the CLDR accounting pattern; when the negative sub-pattern exists, the minus sign is consumed by the pattern, and when it does not exist, the explicit sign part is retained.
4. Compact suffix selection **MUST** first determine the plural category according to §4.1, and then check CLDR `numbers.json` `decimalFormats.{short|long}.decimalFormat[length].decimal-format-pattern.<category>`; when category is missing, fall back to `other`.
5. `useGrouping = "min2"` **MUST** only insert groups when the integer part is ≥ 5 bits (aligned generated-reference `useGrouping` implementation).

---

## 5. FormatRange / FormatRangeToParts <a id="5-formatrange--formatrangetoparts"></a>

**MUST** Rules:

1. `FormatRange(a, b)` **MUST** implement ECMA-402 §15.5.7 `FormatNumericRange`: first format both ends separately, then use the endpoint visible text to decide the approximate branch, and only then call `CollapseNumberRange` to merge the same prefix/suffix for visibly different endpoints.
2. Approximate range equality **MUST** compare the final visible endpoint text produced by the NumberFormat partition pipeline. A rounded-decimal shortcut is forbidden because notation, exponent, sign, currency, unit, compact, and literal parts can make two equal rounded numeric values visibly different.
3. `CollapseNumberRange` **MUST** consume the `NumberFormatPart{Type, Value}` sequence after approximate equality has failed. Prefix/suffix collapse remains package-local; **BANNED** sharing an abstract generic `CollapseRange[T]` with DateTimeFormat.
4. The shared range literal **MUST** come from the constructor-resolved CLDR number symbols `rangeSign`; formatter code must not hard-code the English en dash. When the start range already carries a sign part, the literal may insert spacing around the locale range sign to match native ICU readability.
5. Range source **MUST** be limited to ECMA-402 three values: `"startRange" | "shared" | "endRange"`;`approximatelySign` is a part type, not a source.
6. `FormatRange` / `FormatRangeToParts` **MUST** return `ErrInvalidValue` for `NaN` endpoints instead of signaling errors with empty strings or nil parts. Positive and negative infinity remain valid ECMA-402 mathematical values and must format through the normal parts pipeline.
7. `a > b` **MUST NOT** be locally normalized, transposed, rejected, or added `~`; numeric ranges are formatted in input order and then collapsed.
8. When the formatted endpoint visible text is the same, output shared `approximatelySign` part + shared digital parts (for example, when the maximum fraction digits is 0, `1.1–1.2` outputs `~1`).

> **Why**: NumberFormatPart and DateTimeFormatPart have different fields (`unit | currency | percentSign | exponentInteger` vs `era | year | month | ...`). Although the collapse algorithm has the same structure (removing suffixes), it works on different part fields; Generated references are also implemented separately.
>
> **Rejected**: Abstract general `CollapseRange[T Part]` generic function - one more layer of indirection, and the "equivalence" semantics of `T` are different between the two packages.

---

## 6. Input type support

**MUST** rules (corresponding to ECMA-402 §15.5.1):

1. Public hot path is **FORBIDDEN** to accept `any`; the caller must construct `numberformat.Value` first, and then call `Format`, `FormatToParts`, `FormatRange` or `FormatRangeToParts`.
2. `Decimal` accepts ECMA-402 `StringNumericLiteral`, such as `"1234.5"` / `"NaN"` / `"Infinity"`; the resulting `Value` represents the mathematical value, so `"1"`, `"1.0"`, and `"1.00"` have the same value and differ in visible zeros only when resolved digit options require them. The range method consumes the constructed `Value`.
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

> **Why**: The `internalSlot` mode of `generated-reference` (checking slot every time `format()`) corresponds to "materialization during construction and read-only during runtime" on Go. This is consistent with the CLAUDE.md "constructor-eager / Format-no-error" rule.

---

## 8. Error model

**MUST** Rules:

1. Construction-time errors **MUST** be the wrapped form of `gointl.ErrInvalidOption`, and expose field names, user-input values, locale and expected-value guidance through `*gointl.Error`.
2. Sentinel **MUST** match root `gointl.ErrInvalidOption`(SPEC 12); this package does not establish a separate independent error category.
3. **BANNED** `panic` any user path; `MustNew` does not exist (user can wrap it in the caller).
4. Runtime fallback (NaN / Infinity / string parsing failure) **must not** return an error, but directly output the fallback string.

```go
// Error form example (signature)
err := ecma402.InvalidOptionErrorExpected("numberformat", "currency", code, loc.String(), "a well-formed ISO 4217 currency code")
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

- **BANNED** Introduce `golang.org/x/text/message` as NumberFormat implementation - missing currency/unit/compact notation, incompatible with fixture byte equality.
- **BANNED** Introduce `bojanz/currency` as currency data source - non-CLDR direct output, and comes with ISO 4217 historical data that conflicts with our CLDR nail version.
- **BANNED** Return `error` (ECMA-402 fallback for invalid input) in `Format` / `FormatToParts` paths.
- **BANNED** Use `Format` with `fmt.Sprintf("%v", value)` on hot path - trailing-zero behavior is unaligned + ~150 ns per allocation.
- **FORBIDDEN** NumberFormat parses plural DSL, copies plural rule tables, or selects compact suffix by exposing a `pluralrules.PluralRules` instance; can only call `internal/plural.GetOperands` and use `internal/cldr/plural` generated rules.
- **FORBIDDEN** Extract `CollapseNumberRange` into a cross-package generic - Part fields are different.
- **FORBIDDEN** Calling CLDR domain accessors on the `Format` path; CLDR data must be materialized on `New`.
- **BANNED** Builder chaining API(`numberformat.NewBuilder().Currency("USD").Build()`); active scope exposes only `New(mustLocaleList("en-US"), Options{...})` or equivalent `locale.List` variables.
- **BANNED** Pointer configuration API (`numberformat.New(locales, &Options{...})`) and functional options; the only public configuration form after turning off §8 Q2 is the typed `Options` value.
- **FORBIDDEN** Self-developed `BigDecimal`; all math layers go through [SPEC 21 §Decimal API](./21-number-math.md#decimal-api)(`apd/v3` backend).

---

## 11. Acceptance Ledger

SPEC 20 is accepted by observable behavior, not by step-order traces. The
implementation can simplify internal sequencing when these contracts remain
green.

| Contract | Evidence | Status |
|----------|----------|--------|
| FormatJS `format`, `formatToParts`, `formatRange`, and `formatRangeToParts` fixtures are byte-equal except accepted divergence/XFAIL records. | `numberformat/conformance_unified_test.go`; `numberformat/testdata/conformance/formatjs/*.json`; `task conformance:verify` | Satisfied |
| Compact `zh-TW` output stays source-owned by the generated FormatJS lane, including `format(98765) == "9.9\u842c"`. | `numberformat/testdata/conformance/formatjs/notation-compact-zh-tw-test-ts.json` | Satisfied |
| Constructor, invalid option, NaN, decimal parse, accounting sign, compact long, and rounding-priority behavior are covered by package tests and Node/manual fixtures. | `numberformat/format_test.go`; `numberformat/resolved_options_test.go`; `numberformat/range_test.go`; `numberformat/testdata/conformance/node-v26/*.json`; `numberformat/testdata/conformance/manual/*.json` | Satisfied |
| Resolved optional scalar fields preserve ECMA-402 absence semantics with pointers. | `numberformat/resolved_options_test.go`; `numberformat/conformance_unified_test.go` | Satisfied |
| NumberFormat compact suffix selection is independent from public PluralRules compact source-decimal selection. | `compact_contract_test.go`; SPEC 40 Node compact fixtures | Satisfied |
| NumberFormat keeps decimal rounding centralized in `internal/ecma402/numberformat` and does not expose public compact-plural helpers. | `internal/ecma402/numberformat/*`; absence of `SelectFormatted` / `ResolvePlural` in Go source | Satisfied |
| Race and vet gates pass for the package. | `go test -race ./numberformat/...`; `go vet ./numberformat/...` | Required verification |
| Benchmark telemetry remains per-surface and non-blocking. | `numberformat/benchmark_test.go`; `SPECS/71-benchmark.md`; `task bench` | Satisfied |

The older `TestInitializeNumberFormat_StepOrder` acceptance item is not
retained: no such test exists, and a trace of internal option steps would lock
implementation mechanics rather than ECMA-402-observable behavior.

---

## 11.1 Verification Evidence Boundary

The scientific/engineering rounding-overflow carry (single `computeExponent`,
§4 rule 5) and the `roundingPriority` `RoundingMagnitude` tie-break (§2.2 rule
5) are locked by hand-written Node-witnessed tests because the active FormatJS
extractor does not structurally reduce those source shapes. These behavior tests
are the owning evidence; the absence of generated fixtures does not create a
second implementation or a weaker runtime contract.

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
