# SPEC 42 — RelativeTimeFormat

> **Status:** Draft (2026-05-14)
> **Owner:** `relativetimeformat/` + `internal/cldr/relativetime` + `tools/gen-cldr/`
> **Reference contract:** `.references/ecma402/spec/relativetimeformat.html` first, then `formatjs/packages/intl-relativetimeformat/` + `formatjs/packages/ecma402-abstract/RelativeTimeFormat/`

## Overview

Defines the active `relativetimeformat.RelativeTimeFormat` public API, constructor option model, CLDR relative field data contract, `format`, `formatToParts`, and `supportedLocalesOf` behavior for `Intl.RelativeTimeFormat`.

RelativeTimeFormat composes existing `NumberFormat` and `PluralRules` semantics. It must not reimplement decimal rounding, digit substitution, number parts, plural operands, or plural category selection.

This SPEC does not redefine:

- Locale parsing and matching -> [SPEC 10](./10-locale.md), [SPEC 11](./11-locale-matching.md)
- Number formatting and parts -> [SPEC 20](./20-numberformat.md)
- Decimal mathematical values -> [SPEC 21](./21-number-math.md)
- Plural category selection -> [SPEC 40](./40-pluralrules.md)
- CLDR profile and generator architecture -> [SPEC 50](./50-cldr-data.md)
- Root namespace alias policy -> [SPEC 60](./60-facade.md)

---

## 1. Public API

```go
package relativetimeformat

type LocaleMatcher string
type Style string
type Numeric string
type Unit string
type PartType string

const (
    LookupLocaleMatcher  LocaleMatcher = "lookup"
    BestFitLocaleMatcher LocaleMatcher = "best fit"

    LongStyle   Style = "long"
    ShortStyle  Style = "short"
    NarrowStyle Style = "narrow"

    NumericAlways Numeric = "always"
    NumericAuto   Numeric = "auto"

    Second  Unit = "second"
    Minute  Unit = "minute"
    Hour    Unit = "hour"
    Day     Unit = "day"
    Week    Unit = "week"
    Month   Unit = "month"
    Quarter Unit = "quarter"
    Year    Unit = "year"

    // ECMA-402 §17.5.6 PartitionRelativeTimePattern emits a `literal` part for
    // the pattern shell and re-uses NumberFormat's partition records for the
    // value. RelativeTimeFormat's embedded NumberFormat uses Style="decimal"
    // with no Notation override, so only the number-style-neutral part types
    // listed below can appear at runtime. Constants stay scoped to this
    // package (do not reuse `numberformat.PartType` — see CLAUDE.md).
    PartLiteral   PartType = "literal"
    PartInteger   PartType = "integer"
    PartGroup     PartType = "group"
    PartDecimal   PartType = "decimal"
    PartFraction  PartType = "fraction"
    PartPlusSign  PartType = "plusSign"
    PartMinusSign PartType = "minusSign"
    PartInfinity  PartType = "infinity"
    PartNaN       PartType = "nan"
)

type Options struct {
    LocaleMatcher   *string
    NumberingSystem *string
    Style           *string
    Numeric         *string
}

type ResolvedOptions struct {
    Locale          locale.Locale
    Style           Style
    Numeric         Numeric
    NumberingSystem string
}

type Part struct {
    Type  PartType
    Value string
    Unit  Unit // empty for literal parts
}

type RelativeTimeFormat struct{ /* immutable resolved options + fields + number/plural formatters */ }

func New(locales locale.List, opts Options) (*RelativeTimeFormat, error)
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
func (f *RelativeTimeFormat) ResolvedOptions() ResolvedOptions

type Value struct{ /* opaque */ }

func Int(value int64) Value
func Uint(value uint64) Value
func Float(value float64) Value
func Decimal(value string) (Value, error)

func (f *RelativeTimeFormat) Format(value Value, unit Unit) (string, error)
func (f *RelativeTimeFormat) FormatToParts(value Value, unit Unit) ([]Part, error)
```

MUST rules:

1. `New` accepts a single `Options` value. `New(locales, Options{})` is the ECMA-402 "empty options object" call.
2. `Options{}` defaults to `style="long"` and `numeric="always"`; explicit empty `style` or `numeric` strings are invalid.
3. `RelativeTimeFormat` is immutable after construction. All methods on `*RelativeTimeFormat` must be safe for concurrent callers.
4. Integer and unsigned typed values return errors only for invalid units.
5. `Float` values reject NaN and infinities at `Format` / `FormatToParts` with `ErrInvalidValue`, matching ECMA-402 `RangeError`.
6. `Decimal` rejects malformed or non-finite decimal strings with `ErrInvalidValue`.
7. `ResolvedOptions` returns a value snapshot.
8. JSON field names and `omitempty` behavior follow [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Other Constructors](./73-json-records.md#other-constructors).

> **Why**: JavaScript accepts one dynamic `Number` plus a unit string. Go needs typed numeric bridges, but all bridges still share the same ECMA-402 partitioning algorithm and unit validation.
>
> **Rejected**: public `FormatInt*`, `FormatUint*`, `FormatFloat64*`, and `FormatDecimal*` method families - they encode Go overload mechanics into the method namespace and split one native `format(value, unit)` operation into many public verbs.

---

## 2. Constructor and Options

Pipeline:

1. Validate at most one options object.
2. Read `localeMatcher`, default `best fit`, allowed `lookup | best fit`; `nil` means omitted and `gointl.String("")` is invalid.
3. Validate optional `numberingSystem` as a Unicode numbering-system identifier.
4. Resolve locale against the RelativeTimeFormat supported set with relevant extension key `nu`.
5. Read `style`, default `long`, allowed `long | short | narrow`; `nil` means omitted and `gointl.String("")` is invalid.
6. Read `numeric`, default `always`, allowed `always | auto`; `nil` means omitted and `gointl.String("")` is invalid.
7. Construct or retain internal `numberformat.NumberFormat` using the resolved locale and numbering system.
8. Construct or retain internal cardinal `pluralrules.PluralRules` using the resolved locale.
9. Load CLDR relative field data for the resolved data locale.

MUST rules:

1. Invalid `localeMatcher`, `numberingSystem`, `style`, or `numeric` returns an error wrapping `ErrInvalidOption`.
2. `numberingSystem` option must override the `nu` locale extension during locale resolution, matching ECMA-402 `ResolveLocale`.
3. The internal NumberFormat must format absolute numeric values.
4. The internal PluralRules must select plural category from the original signed value as ECMA-402 `ResolvePlural` does.
5. Constructor support for a locale requires relative time field data, number data, and plural rule data. Do not claim support if any of those payload families is missing.

---

## 3. CLDR Data

`tools/gen-cldr` must extract CLDR relative field data from `cldr-dates-full/main/<locale>/dateFields.json` for every tag in `tools/locale-profile.json`'s `locales` list.

Required units:

```text
second, minute, hour, day, week, month, quarter, year
```

Required style variants:

```text
long:   <unit>
short:  <unit>-short
narrow: <unit>-narrow
```

Each field may contain:

- `future` plural-pattern map
- `past` plural-pattern map
- exact relative literal keys such as `-1`, `0`, or `1`

MUST rules:

1. Generated relative time supported locales must be derived from actual relative field payload maps.
2. Generation must store only the field names needed by ECMA-402 RelativeTimeFormat.
3. Runtime formatting must never read CLDR JSON files.
4. Accessors must go through `internal/cldr/relativetime`.
5. Short and narrow style lookup must fall back to long field data at runtime when the variant is absent.

> **Why**: RelativeTimeFormat is a data composition layer over number and plural primitives. Keeping relative fields small avoids pulling unrelated calendar data into a formatter that only needs fields.

---

## 4. Unit and Value Semantics

### 4.1 Unit normalization

`SingularRelativeTimeUnit` must accept:

```text
second/seconds, minute/minutes, hour/hours, day/days,
week/weeks, month/months, quarter/quarters, year/years
```

It returns the singular form. Any other value returns an error wrapping `ErrInvalidValue`.

### 4.2 Numeric sign

MUST rules:

1. Negative values, including negative zero, select `past`.
2. Positive values and positive zero select `future`.
3. Number formatting uses the absolute value.
4. Plural category selection follows ECMA-402 and existing `pluralrules` behavior.

### 4.3 `numeric=auto`

When `NumericAuto` is selected, the formatter must check for an exact CLDR relative literal key before numeric formatting. If a literal exists, `Format*` returns it and `Format*ToParts` returns a single literal part with empty unit.

MUST rules:

1. Exact literal matching is allowed only for values representable as CLDR relative keys.
2. If no exact literal exists, fall back to numeric past/future formatting.
3. Do not invent fuzzy calendar logic such as "last week" from dates; RelativeTimeFormat formats the caller's numeric offset only.

---

## 5. Formatting and Parts

Formatting pipeline:

1. Reject non-finite or malformed numeric input.
2. Normalize unit to singular.
3. Select style field, with short/narrow fallback to long.
4. If `numeric=auto`, return exact literal when present.
5. Select past or future pattern from sign.
6. Format absolute value with internal NumberFormat.
7. Select plural category with internal PluralRules.
8. Apply the selected pattern using `internal/ecma402.PartitionPattern`.

MUST rules:

1. `Format*` returns the concatenation of `Format*ToParts` values.
2. Numeric parts must preserve number part type strings from `numberformat` when possible.
3. Numeric parts carry the singular unit.
4. Literal parts carry an empty unit.
5. Pattern assembly must use `internal/ecma402.PartitionPattern`; do not hand-parse `{0}`.

---

## 6. Static Supported Locales

```go
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
```

MUST rules:

1. Use the intersection of generated RelativeTimeFormat, NumberFormat, and PluralRules supported locale sets. Do not derive from `tools/locale-profile.json` or from one payload family alone.
2. Call `localematcher.FilterLocalesWithMaximizer`.
3. Accept one `Options` value; `Options{}` represents omitted static-method options.
4. Read only `LocaleMatcher`; ignore formatting options for this static method. `nil` means omitted and an explicit empty string is invalid.
5. Invalid locale matcher returns `ErrInvalidOption`.

---

## 7. Errors

```go
var ErrInvalidOption error
var ErrInvalidValue error
```

MUST rules:

1. `ErrInvalidOption` classifies constructor and `SupportedLocalesOf` option failures.
2. `ErrInvalidValue` classifies invalid runtime unit and numeric input.
3. Errors must be matchable through `errors.Is`.
4. Do not hide constructor failures by falling back to English output.
5. Public errors expose `*gointl.Error` and follow SPEC 12's `expected ...; got ...` text rule.

---

## 8. Root Namespace

After `relativetimeformat` package tests, generated data checks, README, and conformance fixtures pass, the root package may add:

```go
type RelativeTimeFormat = relativetimeformat.RelativeTimeFormat
```

The root package must not add `NewRelativeTimeFormat`, `FormatRelativeTime`, cache controls, or one-shot relative-time helpers.

---

## 9. Testing

MUST rules:

1. Use stdlib `testing`, table-driven tests, and `t.Parallel()` unless shared generated-output state prevents it.
2. Add focused unit tests for constructor defaults, invalid options, invalid units, non-finite values, style fallback, `numeric=auto`, parts joining, and supported locales.
3. Add generator tests for CLDR relative field extraction and generated supported locales.
4. Add generated-reference conformance fixtures under `relativetimeformat/testdata/conformance/formatjs/`.
5. Accepted output mismatches must go to `relativetimeformat/testdata/divergences.md` or `xfail.json`.

Acceptance checks:

- [ ] `go test -race ./relativetimeformat/...`
- [ ] `go test -race ./numberformat/... ./pluralrules/...`
- [ ] `(cd tools/gen-cldr && go test ./...)`
- [ ] `task data:check`
- [ ] `task test`

---

## Forbidden

- No root-level `FormatRelativeTime` or `NewRelativeTimeFormat`.
- No runtime CLDR JSON loading.
- No public cache controls.
- No reimplementation of NumberFormat digit formatting or PluralRules category selection.
- No hand-written supported-locale list.
- No date arithmetic or calendar-relative helpers.
- No implementation of other ECMA-402 constructors as part of this package.
