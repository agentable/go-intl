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
├── constants.go            # sanctioned unit constants
├── constructor_locale.go   # shared constructor ResolveLocale wrapper
├── decimal.go              # decimal-string bridge parsing and finite checks
├── doc.go                  # package boundary
├── errors.go               # ErrInvalidOption and structured OptionError helpers
├── identifier.go           # currency/unit identifier validation
├── partition.go            # PartitionPattern
├── options.go              # production-used ECMA-402 option-resolution helpers
├── types.go                # MathematicalValue, Part, and Pattern contracts
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
3. `types.go` contains only cross-formatter data contracts, not algorithms.
4. Shared code must have at least one production caller. Tests alone do not justify an abstract operation.
5. `internal/ecma402/numberformat.FormatNumericToString` is allowed to be consumed by both NumberFormat and PluralRules because ECMA-402 plural operand rounding is defined in terms of the NumberFormat digit pipeline subset and requires both the formatted string and rounded numeric value.

## 2. Naming Policy

Surviving algorithms keep spec or reference names when that name remains honest:

| Behavior | Go name |
|----------|---------|
| pattern partition | `PartitionPattern` |
| currency syntax | `IsWellFormedCurrencyCode` |
| sanctioned unit syntax | `IsSanctionedSimpleUnitIdentifier` / `IsWellFormedUnitIdentifier` |
| mathematical value conversion | `ToIntlMathematicalValue`, `ParseDecimalInput`, `ParseFiniteDecimalInput`, `RequireFiniteDecimalInput` |
| constructor locale negotiation | `ResolveConstructorLocale` typed wrapper around locale-list, matcher, default-locale, and relevant-extension processing |
| option resolution | `options.One`, `LocaleMatcherAlgorithm`, `LocaleMatcherOption`, `SupportedLocalesOf`, `InvalidStringOption`, `InvalidIntegerOption`, `GetOption`-family helpers or typed equivalents that preserve ECMA-402 semantics |
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
- `LocaleMatcherOption`
- `ResolveConstructorLocale`
- `SupportedLocalesOf`
- `InvalidStringOption`
- `InvalidIntegerOption`
- `GetOption` / typed option-selection equivalent
- `GetNumberOption` / typed numeric range equivalent
- `IsWellFormedCurrencyCode`
- `IsWellFormedUnitIdentifier`
- `numberformat.FormatNumericToString`

It must not reintroduce a generic `map[string]any` option pipeline unless a production path needs JavaScript-value coercion. Typed Go `Options` can feed the same abstract rules without using dynamic maps.

`ResolveConstructorLocale` is the only shared constructor-locale wrapper. It may combine `RequestedLocaleStrings`, `LocaleMatcherAlgorithm`, `DefaultLocale`, and `internal/localematcher.ResolveLocale`, then parse the resolved locale into a Go `locale.Locale`. It must not own formatter data fallback, formatter-specific relevant-extension defaults, unsupported-option errors, CLDR accessor selection, pattern selection, digit resolution, time-zone handling, or embedded formatter construction.

Time-zone name and offset validation live in `internal/tz` per [SPEC 32](./32-datetimeformat-tz.md). They are data-coupled to DateTimeFormat rather than formatter-independent abstract helpers.

## 4. Error Model

`internal/ecma402` exposes the root invalid-option category sentinel and wraps structured error details through the root `gointl.Error` type. The implementation uses `internal/intlerr` to avoid import cycles, while public callers see only `gointl` sentinels and `*gointl.Error`:

```go
var ErrInvalidOption = intlerr.ErrInvalidOption
```

It classifies caller-fixable option failures: invalid enum values, out-of-range numeric options, malformed identifiers, unsupported units, and invalid time-zone names. The public root package exposes the canonical seven category sentinels:

```go
var (
    ErrInvalidOption error
    ErrUnsupportedOption error
    ErrInvalidValue error
    ErrInvalidCode error
    ErrInvalidKey error
    ErrUnsupportedLocale error
    ErrUnsupportedBackend error
)
```

Errors must wrap the sentinel and include the option name, value, owner, and locale where available. Shared validation paths use `OptionError`, `InvalidOptionError`, and `UnsupportedOptionError` so callers can keep using `errors.Is` for root sentinels and `errors.AsType` for structured context:

```go
return ecma402.InvalidOptionError("numberformat", "currency", code, loc.String(), ecma402.ErrInvalidOption)
```

`OptionError` is an alias for the internal implementation of the public root `gointl.Error` type. The public error-detail bridge is:

```go
type ErrorKind string

const (
    InvalidOption      ErrorKind = "invalidOption"
    UnsupportedOption  ErrorKind = "unsupportedOption"
    InvalidValue       ErrorKind = "invalidValue"
    InvalidCode        ErrorKind = "invalidCode"
    InvalidKey         ErrorKind = "invalidKey"
    UnsupportedLocale  ErrorKind = "unsupportedLocale"
    UnsupportedBackend ErrorKind = "unsupportedBackend"
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
3. `ErrUnsupportedOption`, `ErrUnsupportedLocale`, and `ErrUnsupportedBackend` must also match `errors.ErrUnsupported`.
4. `Owner` is the owning Intl package or root namespace name, such as `numberformat`, `datetimeformat`, `displaynames`, or `intl`.
5. `Name` is the rejected option, argument, key, code, or field name.
6. `Value` is the rejected value after public-boundary normalization.
7. `Locale` is empty unless the failure depends on a resolved or requested locale.
8. `Expected` is optional human guidance; when empty, `Error()` derives generic guidance from `Kind` and `Name`.
9. The string returned by `Error()` is for humans only; tests and consumers must not branch on it.
10. Public error text must use the three-part teaching shape: the failing owner/name/value/locale, `expected ...`, and `got ...`.
11. Public error text must not expose ECMA-402 abstract operation names such as `GetOption`, `PartitionPattern`, `ResolveLocale`, `FormatNumericToString`, or `ToIntlMathematicalValue`.

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

`internal/ecma402.MathematicalValue` is the narrow interface shared across the abstract layer. The concrete decimal backend lives in `internal/decimal`.

`ToIntlMathematicalValue` may delegate to `internal/decimal`; callers should not duplicate numeric coercion or parse decimal strings ad hoc. Public decimal-string bridges must use `ParseDecimalInput` when `NaN` / infinities are legal and `ParseFiniteDecimalInput` or `RequireFiniteDecimalInput` when the native operation rejects non-finite values.

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
- `internal/ecma402/types.go` contains `Part`, `Pattern`, and `MathematicalValue`, with no option discriminator.
- `internal/ecma402/errors.go` declares `ErrInvalidOption` plus shared option-error context helpers.
- Root `errors.go` exposes `ErrorKind`, `Error`, and the seven category sentinels; `internal/intlerr/errors_test.go` proves `errors.Is`, `errors.AsType`, and `expected ...; got ...` text behavior.
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
