# SPEC 12 — ECMA-402 Abstract Operations

> **Status:** Revised (2026-05-20)
> **Priority:** High
> **Authority:** This SPEC defines the current `internal/ecma402/` contract. Normative source: `.references/ecma402/spec/`. Generated references are an implementation reference, not the authority.

## Overview

`internal/ecma402/` contains ECMA-402 algorithms, validators, constants, and shared types that production formatter paths actually use. Go public APIs use typed `Options`, but the observable constructor pipeline must remain equivalent to ECMA-402 `ResolveOptions`, `GetOption`, `GetNumberOption`, relevant extension key processing, and resolved internal-slot initialization.

Do not copy Generated reference's JavaScript object model mechanically. Do preserve the ECMA-402 semantics:

- option names and allowed values,
- defaults,
- option conflict errors,
- `localeMatcher`,
- relevant Unicode extension keys,
- resolved options,
- TypeError/RangeError-equivalent error boundaries with structured context at public package boundaries.

Unused audit helpers are still forbidden. Production-used ECMA-402 option operations are not.

## 1. Package Layout

Current layout:

```text
internal/ecma402/
├── unit_identifiers.go     # sanctioned unit identifier accessors
├── constructor_locale.go   # shared constructor ResolveLocale wrapper
├── decimal.go              # decimal-string bridge parsing and finite checks
├── doc.go                  # package boundary
├── errors.go               # ErrInvalidOption and structured OptionError helpers
├── identifier.go           # currency/unit identifier validation
├── partition.go            # PartitionPattern
├── options.go              # production-used ECMA-402 option-resolution helpers
├── pattern.go              # Part and Pattern contracts
├── numberformat/
│   └── digits.go           # shared digit rounding pipeline
├── datetimeformat/
│   ├── matcher.go          # basic/best-fit format matcher
│   └── skeleton.go         # skeleton parser
└── pluralrules/
    ├── category.go         # plural categories
    └── operands.go         # exact CLDR operands
```

Rules:

1. Root `internal/ecma402` contains formatter-independent algorithms only.
2. Subpackages may import root `internal/ecma402`; subpackages must not import sibling formatter subpackages.
3. `pattern.go` contains only cross-formatter pattern contracts, not algorithms.
4. Shared code must have at least one production caller. Tests alone do not justify an abstract operation.
5. `internal/ecma402/numberformat.FormatNumericToString` is allowed to be consumed by both NumberFormat and PluralRules because ECMA-402 plural operand rounding is defined in terms of the NumberFormat digit pipeline subset and requires both the formatted string and rounded numeric value.

## 2. Naming Policy

Surviving algorithms keep spec or reference names when that name remains honest:

| Behavior | Go name |
|----------|---------|
| pattern partition | `PartitionPattern` |
| currency syntax | `IsWellFormedCurrencyCode` |
| sanctioned unit syntax | `IsSanctionedSimpleUnitIdentifier` / `IsWellFormedUnitIdentifier` |
| typed mathematical-value bridge | `NumericValue`, `DecimalNumericValue`, `Int64NumericValue`, `Uint64NumericValue`, `Float64NumericValue`, `BigIntNumericValue`, `ParseDecimalInput`, `ParseFiniteDecimalInput`, `RequireFiniteDecimalInput` |
| constructor locale negotiation | `ResolveConstructorLocale` typed wrapper around locale-list, matcher, default-locale, and relevant-extension processing |
| option resolution | `options.One`, `LocaleMatcherAlgorithm`, `LocaleMatcherOption` / `LocaleMatcherOptionInput`, `SupportedLocalesOf`, `InvalidStringOption`, `InvalidIntegerOption`, Unicode type option validators, `GetOption`-family helpers or typed equivalents that preserve ECMA-402 semantics |
| digit rounding and padding stage | `numberformat.FormatNumericToString` |
| date skeleton parsing | `datetimeformat.ParseSkeleton`-family internals |
| plural operand construction | `pluralrules.Operands`-family internals |

Public formatter packages keep Go typed APIs, but exported names and behavior must map to ECMA-402 constructors, methods, options, resolved options, parts, or range sources.

## 3. Option Validation

The option validation contract starts in public formatter constructors:

- `numberformat.New(locale.List, Options)`
- `datetimeformat.New(locale.List, Options)`
- `pluralrules.New(locale.List, Options)`
- `listformat.New(locale.List, Options)`
- `relativetimeformat.New(locale.List, Options)`
- `durationformat.New(locale.List, Options)`
- `displaynames.New(locale.List, Options)`
- `collator.New(locale.List, Options)`
- `segmenter.New(locale.List, Options)`

Those constructors validate enum values, numeric ranges, identifiers, and unsupported locale/data combinations before a formatter is returned. After construction, hot-path methods do not return option errors.

The abstract layer should provide reusable option and syntax helpers when they encode normative ECMA-402 behavior:

- `ErrInvalidOption`
- `options.One`
- `LocaleMatcherAlgorithm`
- `LocaleMatcherOption` / `LocaleMatcherOptionInput`
- `ResolveConstructorLocale`
- `SupportedLocalesOf`
- `InvalidStringOption`
- `InvalidIntegerOption`
- `GetOption` / typed option-selection equivalent
- `GetNumberOption` / typed numeric range equivalent
- `IsWellFormedUnicodeType`
- `ValidateUnicodeTypeOption` / `ValidateUnicodeTypeOptionInput`
- `IsWellFormedCurrencyCode`
- `IsWellFormedUnitIdentifier`
- `numberformat.FormatNumericToString`

It must not reintroduce a generic `map[string]any` option pipeline unless a production path needs JavaScript-value coercion. Typed Go `Options` can feed the same abstract rules without using dynamic maps.

When a public Go option needs to preserve ECMA-402's omitted-versus-present distinction, the constructor should use a pointer field, copy the pointed-to value into private config during construction, and then feed the private string or scalar into the shared validators. `nil` means the option was omitted; an explicit pointer to `""` is a caller-provided value and must be validated as such. Shared static-method helpers such as `SupportedLocalesOf` accept the pointer-backed option directly so formatter packages do not duplicate present-bit extraction.

`ResolveConstructorLocale` is the only shared constructor-locale wrapper. It may combine `RequestedLocaleStrings`, `LocaleMatcherAlgorithm`, `DefaultLocale`, and `internal/localematcher.ResolveLocale`, then parse the resolved locale into a Go `locale.Locale`. It must not own formatter data fallback, formatter-specific relevant-extension defaults, unsupported-option errors, CLDR accessor selection, pattern selection, digit resolution, time-zone handling, or embedded formatter construction.

Time-zone name and offset validation live in `internal/tz` per [SPEC 32](./32-datetimeformat-tz.md). They are data-coupled to DateTimeFormat rather than formatter-independent abstract helpers.

## 4. Error Model

`internal/ecma402` exposes the root invalid-option category sentinel and wraps structured error details through the root `gointl.Error` type. The implementation uses `internal/intlerr` to avoid import cycles, while public callers see only the four reachable `gointl` sentinels and `*gointl.Error`:

```go
var ErrInvalidOption = intlerr.ErrInvalidOption
```

It classifies caller-fixable option failures: invalid enum values, out-of-range numeric options, malformed identifiers, unsupported units, and invalid time-zone names. The public root package exposes exactly the categories produced by reachable public paths:

```go
var (
    ErrInvalidOption error
    ErrUnsupportedOption error
    ErrInvalidValue error
    ErrInvalidCode error
)
```

Errors must wrap the sentinel and include the option name, value, owner, locale, and expected-value guidance where available. When an option error translates a lower-level dependency, data, or embedded-formatter failure, the structured error must also preserve that cause so `errors.Is` can still reach it. Shared validation paths use `OptionError`, `InvalidOptionErrorExpected`, `UnsupportedOptionErrorExpected`, and the string/integer option helpers so callers can keep using `errors.Is` for root sentinels and `errors.AsType` for structured context:

```go
return ecma402.InvalidOptionErrorExpected("numberformat", "currency", code, loc.String(), "a well-formed ISO 4217 currency code", err)
```

`OptionError` is an alias for the internal implementation of the public root `gointl.Error` type. The public error-detail bridge is:

```go
type ErrorKind string

const (
    InvalidOption      ErrorKind = "invalidOption"
    UnsupportedOption  ErrorKind = "unsupportedOption"
    InvalidValue       ErrorKind = "invalidValue"
    InvalidCode        ErrorKind = "invalidCode"
)

type Error struct {
    Kind     ErrorKind
    Owner    string
    Name     string
    Value    string
    Locale   string
    Expected string
    Err      error
}
```

Required behavior:

1. `errors.Is(err, gointl.ErrInvalidOption)` and the other root category sentinels remain the stable branch points for caller code.
2. `detail, ok := errors.AsType[*gointl.Error](err)` exposes machine-readable context for host bindings, config UIs, and API adapters.
3. `ErrUnsupportedOption` must also match `errors.ErrUnsupported`; the other categories must not.
4. `errors.Is(err, underlying)` must keep working when a structured error maps an internal dependency, data, or embedded-formatter failure to a public Intl category.
5. `Owner` is the owning Intl package or root namespace name, such as `numberformat`, `datetimeformat`, `displaynames`, or `intl`.
6. `Name` is the rejected option, argument, key, code, or field name.
7. `Value` is the rejected value after public-boundary normalization.
8. `Locale` is empty unless the failure depends on a resolved or requested locale.
9. `Expected` is optional human guidance; when empty, `Error()` derives generic guidance from `Kind` and `Name`.
10. The string returned by `Error()` is for humans only; tests and consumers must not branch on it.
11. Public error text must use the three-part teaching shape: the failing owner/name/value/locale, `expected ...`, and `got ...`.
12. Public error text must not expose ECMA-402 abstract operation names such as `GetOption`, `PartitionPattern`, `ResolveLocale`, `FormatNumericToString`, or `ToIntlMathematicalValue`.

Formatter-owned runtime failures that are not constructor options, such as malformed decimal strings, invalid relative-time units, invalid display-name codes, or invalid duration records, construct the same structured error and wrap the matching root category sentinel.

There is no separate public `ErrInvalidOptionType` because the public API no longer accepts arbitrary option objects. Shape errors are compile-time problems or package-specific constructor errors.

## 5. Internal Slots

Go formatter structs hold resolved state directly. Do not simulate JavaScript internal slots with weak maps, `sync.Map`, or `map[string]any`.

Accepted shape:

```go
type NumberFormat struct {
    locale locale.Locale
    style  Style
    // resolved fields...
}
```

Rejected shape:

```go
var slots sync.Map // map[*NumberFormat]map[string]any
```

## 6. Math Value Boundary

`internal/ecma402.NumericValue` is the closed record shared across the abstract
layer. Package-local public `Value` constructors select an explicit Go bridge
(`Int`, `Uint`, `Float`, `BigInt`, or `Decimal`) and produce that record. The
concrete decimal representation and arithmetic live in `internal/decimal`.

The Go core does not accept `any` and does not emulate JavaScript
`ToPrimitive` coercion. Public decimal-string bridges use `ParseDecimalInput`
when `NaN` / infinities are legal and `ParseFiniteDecimalInput` or
`RequireFiniteDecimalInput` when the native operation rejects non-finite
values. Host adapters must choose one typed bridge before entering the core.

## Forbidden

- Reintroducing unused JS object option helpers.
- Adding pending-symbol audit tests.
- Adding placeholder `Internal struct{}` types that no production path consumes.
- Simulating JS weak-map internal slots.
- Returning option errors from hot-path formatting methods after a formatter has been constructed.
- Returning plain string-only errors for caller-fixable public failures when a root sentinel and `gointl.Error` can carry the same context.
- Matching errors by `Error()` text in tests or consumer-facing code.
- Importing CLDR generated tables from the abstract layer.
- Cross-importing formatter abstract subpackages.

## Acceptance Criteria

- `internal/ecma402` contains only production-used algorithms, validators, constants, and shared types.
- `internal/ecma402/pattern.go` contains `Part` and `Pattern`, with no option discriminator.
- `internal/ecma402/errors.go` declares `ErrInvalidOption` plus shared option-error context helpers.
- Root `errors.go` exposes `ErrorKind`, `Error`, and exactly four reachable category sentinels; `internal/intlerr/errors_test.go` proves `errors.Is`, `errors.AsType`, `errors.ErrUnsupported`, and `expected ...; got ...` text behavior.
- `ResolveConstructorLocale` has production constructor callers and internal tests for localeMatcher dispatch plus relevant-extension option override behavior.
- Any `GetOption` / `GetNumberOption` / typed equivalent has a production caller and is covered by constructor error/resolved-options tests.
- No `audit_test.go` in `internal/ecma402` or `internal/cldr` maintains pending implementation rows.
- `go test ./internal/ecma402/...` passes.
- `go list -deps ./internal/ecma402/...` does not contain formatter sibling cycles.
- Formatter packages validate typed options in constructors and wrap `gointl.ErrInvalidOption` for caller-fixable failures.

## References

- `.references/ecma402/spec/conventions.html` — abstract operation and completion conventions.
- `.references/ecma402/spec/negotiation.html` — `CanonicalizeLocaleList`, `ResolveOptions`, `ResolveLocale`, `FilterLocales`.
- `.references/ecma402/spec/numberformat.html` — NumberFormat constructor, digit options, `FormatNumericToString`.
- `.references/ecma402/spec/datetimeformat.html` — DateTimeFormat constructor, time-zone handling, pattern partitioning.
- `.references/ecma402/spec/pluralrules.html` — PluralRules constructor and plural resolution.
- `.references/formatjs/packages/ecma402-abstract/` — readable implementation reference for algorithms that still exist here.
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/` — digit pipeline and rounding reference.
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/` — skeleton and matcher reference.
