# SPEC 43 — DurationFormat

> **Status:** Draft (2026-05-14)
> **Owner:** `durationformat/` + `internal/cldr` units/number/list/plural payloads + `tools/gen-cldr/`
> **Reference contract:** `.references/ecma402/spec/durationformat.html` first, then `formatjs/packages/intl-durationformat/` + `formatjs/packages/ecma402-abstract/DurationFormat/`

## Overview

Defines the active `durationformat.DurationFormat` public API, constructor option model, duration record bridge, CLDR unit-pattern data contract, `format`, `formatToParts`, and `supportedLocalesOf` behavior for `Intl.DurationFormat`.

DurationFormat composes existing `NumberFormat`, `ListFormat`, and `PluralRules` semantics. It must not reimplement number formatting, unit display, list joining, digit substitution, or plural category selection.

This SPEC does not redefine:

- Locale parsing and matching -> [SPEC 10](./10-locale.md), [SPEC 11](./11-locale-matching.md)
- Number formatting and parts -> [SPEC 20](./20-numberformat.md)
- Decimal mathematical values -> [SPEC 21](./21-number-math.md)
- Plural category selection -> [SPEC 40](./40-pluralrules.md)
- List joining and list parts -> [SPEC 41](./41-listformat.md)
- CLDR profile and generator architecture -> [SPEC 50](./50-cldr-data.md)
- Root namespace alias policy -> [SPEC 60](./60-facade.md)

---

## 1. Public API

```go
package durationformat

type LocaleMatcher string
type Style string
type UnitStyle string
type Display string
type Unit string
type PartType string

type Options struct {
    LocaleMatcher       *string
    NumberingSystem     *string
    Style               *string
    Years               *string
    YearsDisplay        *string
    Months              *string
    MonthsDisplay       *string
    Weeks               *string
    WeeksDisplay        *string
    Days                *string
    DaysDisplay         *string
    Hours               *string
    HoursDisplay        *string
    Minutes             *string
    MinutesDisplay      *string
    Seconds             *string
    SecondsDisplay      *string
    Milliseconds        *string
    MillisecondsDisplay *string
    Microseconds        *string
    MicrosecondsDisplay *string
    Nanoseconds         *string
    NanosecondsDisplay  *string
    FractionalDigits    *int
}

type Duration struct {
    Years        float64
    Months       float64
    Weeks        float64
    Days         float64
    Hours        float64
    Minutes      float64
    Seconds      float64
    Milliseconds float64
    Microseconds float64
    Nanoseconds  float64
}

type Part struct {
    Type  PartType
    Value string
    Unit  Unit // empty for literal parts
}

const (
    // ECMA-402 §18.5.5 PartitionDurationFormatPattern combines list-pattern
    // literals, digital separators, and embedded NumberFormat partition
    // records. DurationFormat's embedded NumberFormat uses Style="unit"
    // (long/short/narrow) or Style="decimal" (2-digit/numeric); other
    // notations are not exposed, so only the constants below can appear.
    // Constants stay scoped to this package (do not reuse
    // `numberformat.PartType` — see CLAUDE.md).
    PartLiteral   PartType = "literal"
    PartInteger   PartType = "integer"
    PartGroup     PartType = "group"
    PartDecimal   PartType = "decimal"
    PartFraction  PartType = "fraction"
    PartPlusSign  PartType = "plusSign"
    PartMinusSign PartType = "minusSign"
    PartInfinity  PartType = "infinity"
    PartNaN       PartType = "nan"
    PartUnit      PartType = "unit"
)

func New(locales locale.List, opts Options) (*DurationFormat, error)
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
func (f *DurationFormat) ResolvedOptions() ResolvedOptions
func (f *DurationFormat) Format(duration Duration) (string, error)
func (f *DurationFormat) FormatToParts(duration Duration) ([]Part, error)
```

MUST rules:

1. `New` accepts one `Options` value. `New(locales, Options{})` matches JavaScript `new Intl.DurationFormat(locales, {})` or omitted options.
2. `Options{}` defaults to `style="short"`; explicit empty `style` is invalid.
3. `DurationFormat` is immutable after construction. All methods on `*DurationFormat` must be safe for concurrent callers.
4. `ResolvedOptions` returns a value snapshot and uses ECMA-402 option names and values.
5. Go uses a typed `Duration` struct whose `float64` fields mirror ECMAScript Number instead of accepting `any` or parsing Temporal strings. Zero-valued `Duration{}` is a valid typed bridge and formats to the result implied by resolved display defaults.
6. Formatting methods return errors for invalid duration values where JavaScript would throw `TypeError` or `RangeError`.
7. JSON field names and `omitempty` behavior follow [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Other Constructors](./73-json-records.md#other-constructors).

---

## 2. Constructor and Options

Pipeline:

1. Validate at most one options object.
2. Read `localeMatcher`, default `best fit`, allowed `lookup | best fit`; `nil` means omitted and `gointl.String("")` is invalid.
3. Validate optional `numberingSystem` as a Unicode numbering-system identifier.
4. Resolve locale against the DurationFormat supported set with relevant extension key `nu`.
5. Read `style`, default `short`, allowed `long | short | narrow | digital`; `nil` means omitted and `gointl.String("")` is invalid.
6. Resolve unit options for years through nanoseconds using ECMA-402 `GetDurationUnitOptions`; each unit style and unit display option is a presence-aware `*string`, where `nil` means omitted/default and `gointl.String("")` is invalid.
7. Read `fractionalDigits`, allowed integer 0 through 9, default omitted.
8. Load the locale's numeric time separator from CLDR number symbols, falling back to `":"` if the payload is empty.
9. Materialize the embedded `NumberFormat` and `ListFormat` instances implied by the resolved options, including sign-hidden variants and fractional numeric variants, so cached formatting does not repeat locale negotiation, option validation, or CLDR data lookup.

MUST rules:

1. Invalid `localeMatcher`, `numberingSystem`, `style`, unit style, unit display, or `fractionalDigits` returns an error wrapping `ErrInvalidOption`; explicit empty `style`, unit style, or unit display is invalid rather than treated as omitted.
2. `numberingSystem` option must override the `nu` locale extension during locale resolution.
3. `style="digital"` must select numeric hour/minute/second defaults and short date-unit defaults according to ECMA-402.
4. `fractionalDigits` must distinguish omitted from explicit zero; `FractionalDigits: gointl.Int(0)` is an explicit option, and constructors copy the pointee value.
5. Unit option validation must reject numeric/non-numeric mixing and illegal fractional display combinations according to ECMA-402 `ValidateDurationUnitStyle`.

---

## 3. Duration Value Semantics

MUST rules:

1. Each public duration field must be a finite integral ECMAScript Number. NaN, infinities, and non-integral values return `ErrInvalidValue`; negative zero projects to integer zero.
2. Each field must be projected once at the formatting boundary into the exact integer represented by its `float64`. No later sign, limit, rollup, or NumberFormat path may narrow that value through `int64` or `float64`.
3. All duration fields must have the same sign after zero fields are ignored. Mixed positive and negative fields return `ErrInvalidValue`.
4. `years`, `months`, and `weeks` absolute values must be less than 2^32.
5. The normalized seconds value for days through nanoseconds must be less than 2^53 seconds.
6. Fractional rollup for milliseconds, microseconds, and nanoseconds must use exact decimal or integer math after boundary projection.
7. Fractional rollup must carry into the parent unit. For example, `1s + 1000ms` formats as `2s` when milliseconds are fractional.

> **Why**: ECMA-402 accepts Number fields but computes the duration record as mathematical integers. Exact projection preserves represented values such as `2^53 + 2` and `1e20`; binary-float rollup breaks observable `roundingMode: "trunc"` behavior.

---

## 4. CLDR Data

`tools/gen-cldr` must extract:

- duration units `year`, `month`, `week`, `day`, `hour`, `minute`, `second`, `millisecond`, `microsecond`, and `nanosecond` from `cldr-units-full`
- number-symbol `timeSeparator` from `cldr-numbers-full`
- generated supported locales for unit-pattern payloads

MUST rules:

1. DurationFormat support requires unit pattern data, number symbol data, list pattern data, and plural rule data. A locale is supported only when all four payload families are present.
2. `durationformat.SupportedLocalesOf` must derive its set from the intersection of the generated `internal/cldr/unit`, `internal/cldr/number`, `internal/cldr/list`, and `internal/cldr/plural` supported-locale accessors; a hand-written list is forbidden.
3. Runtime formatting must never read CLDR JSON files.

---

## 5. Formatting and Parts

Formatting pipeline:

1. Validate each Number field and project it once into the private exact duration record.
2. Walk units in ECMA-402 order: years, months, weeks, days, hours, minutes, seconds, milliseconds, microseconds, nanoseconds.
3. For non-numeric unit styles, use the constructor-resolved `NumberFormat` with `style="unit"` and the resolved unit display.
4. For the first numeric or two-digit time unit, format the remaining numeric time sequence with CLDR separators.
5. When the next smaller unit is fractional, add exact fractional digits to the current unit, use `roundingMode="trunc"`, and stop.
6. Join unit groups with the constructor-resolved `ListFormat{type:"unit"}`; `style="digital"` joins non-digital groups using short list style.

MUST rules:

1. `Format` returns the concatenation of `FormatToParts` values.
2. Number parts must preserve number part type strings from `numberformat`.
3. Number parts carry the singular duration unit.
4. Literal parts carry an empty unit unless they come from a unit formatter part.
5. Digital separators are literal parts.

---

## 6. Static Supported Locales

```go
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
```

MUST rules:

1. Use the intersection of generated unit, number, list, and plural supported locale sets. Do not derive from `tools/locale-profile.json` or from one payload family alone.
2. Call `localematcher.FilterLocalesWithMaximizer`.
3. Accept one `Options` value; `Options{}` represents omitted static-method options.
4. Read only `LocaleMatcher`; invalid values, including an explicit empty string, return `ErrInvalidOption`.

---

## 7. Errors

```go
var ErrInvalidOption error
var ErrInvalidValue error
```

MUST rules:

1. `ErrInvalidOption` classifies invalid constructor and `SupportedLocalesOf` options.
2. `ErrInvalidValue` classifies invalid duration records and invalid numeric formatting values.
3. Errors must be matchable through `errors.Is`.
4. Public errors expose `*gointl.Error` and follow SPEC 12's `expected ...; got ...` text rule.

---

## 8. Root Namespace

After `durationformat` package tests, generated data checks, README, and conformance fixtures pass, the root package may add:

```go
type DurationFormat = durationformat.DurationFormat
```

The root package must not add `NewDurationFormat`, `FormatDuration`, cache controls, or one-shot duration helpers.

---

## 9. Testing

MUST rules:

1. Use stdlib `testing`, table-driven tests, and `t.Parallel()` unless shared generated-output state prevents it.
2. Add focused tests for constructor defaults, invalid options, supported locales, default formatting, digital formatting, parts, exact fractional rollup, invalid duration values, Number boundaries, and wide integral values.
3. Add generator tests for CLDR unit extraction, `timeSeparator`, and generated supported locales.
4. Add generated-reference conformance fixtures under `durationformat/testdata/conformance/formatjs/`.
5. Accepted output mismatches must go to `durationformat/testdata/divergences.md` or `xfail.json`, never by removing generated fixture cases.

Acceptance checks:

- [ ] `go test -race ./durationformat/...`
- [ ] `(cd tools/gen-cldr && go test ./...)`
- [ ] `(cd tools/gen-fixtures-from-formatjs && go test ./...)`
- [ ] `task conformance:verify`
- [ ] `task data:check`
- [ ] `task test`

---

## Forbidden

- No root-level `FormatDuration` or `NewDurationFormat`.
- No runtime CLDR JSON loading.
- No `time.Duration` replacement for the public `Duration` record; ECMA-402 durations include years, months, and weeks.
- No public `any` duration input or Temporal string parser.
- No `float64` math or `int64` narrowing after the public Number boundary is projected into the private exact record.
- No hand-written supported-locale list.
