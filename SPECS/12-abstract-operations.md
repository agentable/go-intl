# SPEC 12 — ECMA-402 Abstract Operations

> **Status:** Revised (2026-05-12)
> **Priority:** High
> **Authority:** This SPEC defines the current `internal/ecma402/` contract. Normative source: `.references/ecma402/spec/`. FormatJS is an implementation reference, not the authority.

## Overview

`internal/ecma402/` contains ECMA-402 algorithms, validators, constants, and shared types that production formatter paths actually use. Go public APIs use typed `Options`, but the observable constructor pipeline must remain equivalent to ECMA-402 `ResolveOptions`, `GetOption`, `GetNumberOption`, relevant extension key processing, and resolved internal-slot initialization.

Do not copy FormatJS's JavaScript object model mechanically. Do preserve the ECMA-402 semantics:

- option names and allowed values,
- defaults,
- option conflict errors,
- `localeMatcher`,
- relevant Unicode extension keys,
- resolved options,
- TypeError/RangeError-equivalent error boundaries.

Unused audit helpers are still forbidden. Production-used ECMA-402 option operations are not.

## 1. Package Layout

Current layout:

```text
internal/ecma402/
├── constants.go            # sanctioned unit constants
├── errors.go               # ErrInvalidOption sentinel
├── identifier.go           # currency/unit identifier validation
├── math_value.go           # ToIntlMathematicalValue boundary
├── partition.go            # PartitionPattern
├── timezone.go             # IsValidTimeZoneName and offset validation
├── options.go              # production-used ECMA-402 option-resolution helpers
├── types/
│   ├── math.go             # MathematicalValue interface
│   └── part.go             # Part / Pattern
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
2. Subpackages may import root `internal/ecma402` and `internal/ecma402/types`; subpackages must not import sibling formatter subpackages.
3. `types/` contains only cross-formatter data contracts, not algorithms.
4. Shared code must have at least one production caller. Tests alone do not justify an abstract operation.
5. `internal/ecma402/numberformat.FormatNumericToString` is allowed to be consumed by both NumberFormat and PluralRules because ECMA-402 plural operand rounding is defined in terms of the NumberFormat digit pipeline subset and requires both the formatted string and rounded numeric value.

## 2. Naming Policy

Surviving algorithms keep spec or reference names when that name remains honest:

| Behavior | Go name |
|----------|---------|
| pattern partition | `PartitionPattern` |
| currency syntax | `IsWellFormedCurrencyCode` |
| sanctioned unit syntax | `IsSanctionedSimpleUnitIdentifier` / `IsWellFormedUnitIdentifier` |
| time-zone validity | `IsValidTimeZoneName` |
| mathematical value conversion | `ToIntlMathematicalValue` |
| option resolution | `GetOption`-family helpers or typed equivalents that preserve ECMA-402 semantics |
| digit rounding and padding stage | `numberformat.FormatNumericToString` |
| date skeleton parsing | `datetimeformat.ParseSkeleton`-family internals |
| plural operand construction | `pluralrules.Operands`-family internals |

Public formatter packages keep Go typed APIs, but exported names and behavior must map to ECMA-402 constructors, methods, options, resolved options, parts, or range sources.

## 3. Option Validation

The option validation contract starts in public formatter constructors:

- `numberformat.New(locale.Locale, Options)`
- `datetimeformat.New(locale.Locale, Options)`
- `pluralrules.New(locale.Locale, Options)`

Those constructors validate enum values, numeric ranges, identifiers, and unsupported locale/data combinations before a formatter is returned. After construction, hot-path methods do not return option errors.

The abstract layer should provide reusable option and syntax helpers when they encode normative ECMA-402 behavior:

- `ErrInvalidOption`
- `GetOption` / typed option-selection equivalent
- `GetNumberOption` / typed numeric range equivalent
- `IsWellFormedCurrencyCode`
- `IsWellFormedUnitIdentifier`
- `IsValidTimeZoneName`
- `numberformat.FormatNumericToString`

It must not reintroduce a generic `map[string]any` option pipeline unless a production path needs JavaScript-value coercion. Typed Go `Options` can feed the same abstract rules without using dynamic maps.

## 4. Error Model

`internal/ecma402` exposes one shared sentinel:

```go
var ErrInvalidOption = errors.New("ecma402: invalid option")
```

It classifies caller-fixable option failures: invalid enum values, out-of-range numeric options, malformed identifiers, unsupported units, and invalid time-zone names. Public formatter packages may re-export it through package-local names.

Errors should wrap the sentinel and include the option name, value, and locale where available:

```go
return fmt.Errorf("numberformat: currency %q for locale %s: %w", code, loc, ErrInvalidOption)
```

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

`internal/ecma402/types.MathematicalValue` is the narrow interface shared across the abstract layer. The concrete decimal backend lives in `internal/decimal`.

`ToIntlMathematicalValue` may delegate to `internal/decimal`; callers should not duplicate numeric coercion or parse decimal strings ad hoc.

## Forbidden

- Reintroducing unused JS object option helpers.
- Adding pending-symbol audit tests.
- Adding placeholder `Internal struct{}` types that no production path consumes.
- Simulating JS weak-map internal slots.
- Returning option errors from hot-path formatting methods after a formatter has been constructed.
- Importing CLDR generated tables from the abstract layer.
- Cross-importing formatter abstract subpackages.

## Acceptance Criteria

- `internal/ecma402` contains only production-used algorithms, validators, constants, and shared types.
- `internal/ecma402/types/` contains `Part`, `Pattern`, and `MathematicalValue`, with no option discriminator.
- `internal/ecma402/errors.go` declares `ErrInvalidOption` only.
- Any `GetOption` / `GetNumberOption` / typed equivalent has a production caller and is covered by constructor error/resolved-options tests.
- No `audit_test.go` in `internal/ecma402` or `internal/cldr` maintains pending implementation rows.
- `go test ./internal/ecma402/...` passes.
- `go list -deps ./internal/ecma402/...` does not contain formatter sibling cycles.
- Formatter packages validate typed options in constructors and wrap `ErrInvalidOption` for caller-fixable failures.

## References

- `.references/ecma402/spec/conventions.html` — abstract operation and completion conventions.
- `.references/ecma402/spec/negotiation.html` — `CanonicalizeLocaleList`, `ResolveOptions`, `ResolveLocale`, `FilterLocales`.
- `.references/ecma402/spec/numberformat.html` — NumberFormat constructor, digit options, `FormatNumericToString`.
- `.references/ecma402/spec/datetimeformat.html` — DateTimeFormat constructor, time-zone handling, pattern partitioning.
- `.references/ecma402/spec/pluralrules.html` — PluralRules constructor and plural resolution.
- `.references/formatjs/packages/ecma402-abstract/` — readable implementation reference for algorithms that still exist here.
- `.references/formatjs/packages/ecma402-abstract/NumberFormat/` — digit pipeline and rounding reference.
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/` — skeleton and matcher reference.
