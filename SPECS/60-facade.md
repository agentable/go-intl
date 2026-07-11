# SPEC 60 — Root `Intl` Namespace

> **Status:** Revised (2026-05-31)
> **Type:** Consumer API Spec — defines the public entry surface for the root `go-intl` package.
> **Authority:** ECMA-402 `.references/ecma402/spec/intl.html` is the normative source. This spec records the current root package contract. SPECS 10/20/30/40/41/42/43/44/45/46 record the active constructor package contracts.

---

## Overview

The root `github.com/agentable/go-intl` package represents the implemented JavaScript `Intl` namespace surface as closely as Go allows. It must not pretend that unimplemented constructors exist.

ECMA-402 says the `Intl` object:

1. is an ordinary namespace object,
2. is not a function object,
3. has no `[[Construct]]` internal method,
4. has no `[[Call]]` internal method,
5. exposes constructor properties such as `Intl.Locale`, `Intl.NumberFormat`, `Intl.DateTimeFormat`, and `Intl.PluralRules`,
6. exposes static common functions such as `Intl.getCanonicalLocales` and `Intl.supportedValuesOf`.

Therefore the root Go package is **not** a per-locale `Intl` session, not a third-party helper facade clone, and not a holder for root one-shot formatting helpers.

Formatter construction and behavior live in the active constructor packages:

- `locale`
- `numberformat`
- `datetimeformat`
- `pluralrules`
- `listformat`
- `relativetimeformat`
- `durationformat`
- `displaynames`
- `collator`
- `segmenter`

The root package exposes active constructor aliases as the Go bridge for ECMA-402 constructor properties, plus static common namespace functions. It must not contain formatting logic.

All ten ECMA-402 active constructors are aliased on the root package once each package passes its own conformance gate.

---

## 1. Root Package Shape

### 1.1 Public namespace functions

```go
package gointl

func GetCanonicalLocales(locales locale.List) locale.List
func SupportedCalendars() []string
func SupportedCollations() []string
func SupportedCurrencies() []string
func SupportedNumberingSystems() []string
func SupportedTimeZones() []string
func SupportedUnits() []string
```

Mapping:

| Go | ECMA-402 |
|----|----------|
| `GetCanonicalLocales` | `Intl.getCanonicalLocales(locales)` after locale parsing has occurred at the Go boundary |
| `SupportedCalendars` | `Intl.supportedValuesOf("calendar")` |
| `SupportedCollations` | `Intl.supportedValuesOf("collation")` |
| `SupportedCurrencies` | `Intl.supportedValuesOf("currency")` |
| `SupportedNumberingSystems` | `Intl.supportedValuesOf("numberingSystem")` |
| `SupportedTimeZones` | `Intl.supportedValuesOf("timeZone")` |
| `SupportedUnits` | `Intl.supportedValuesOf("unit")` |

`GetCanonicalLocales` accepts `locale.Locale` values because Go should parse raw strings once at the boundary. If callers need to canonicalize raw tags, they call `locale.Parse` or `locale.ParseList` first. This is the Go typed bridge for ECMA-402 `CanonicalizeLocaleList`.

This is the only public locale-list canonicalization entrypoint. The lower-level canonical locale-list operation stays in `internal/ecma402` for constructor initialization and `SupportedLocalesOf`; package-level abstract-operation helpers must not be reintroduced.

Implementation organization stays package-scoped: `intl.go` owns the root
constructor aliases and `GetCanonicalLocales`, while `supported.go` owns the
typed `Intl.supportedValuesOf` accessors. This is a file organization
convention, not a new public namespace.

### 1.2 Active constructor aliases

The root package **MUST** expose type aliases matching ECMA-402 constructor properties for every active constructor:

```go
type Locale = locale.Locale
type NumberFormat = numberformat.NumberFormat
type DateTimeFormat = datetimeformat.DateTimeFormat
type PluralRules = pluralrules.PluralRules
type ListFormat = listformat.ListFormat
type RelativeTimeFormat = relativetimeformat.RelativeTimeFormat
type DurationFormat = durationformat.DurationFormat
type DisplayNames = displaynames.DisplayNames
type Collator = collator.Collator
type Segmenter = segmenter.Segmenter
```

Construction still belongs to the constructor packages:

```go
locales := mustLocaleList("en-US")
nf, err := numberformat.New(locales, numberformat.Options{})
dtf, err := datetimeformat.New(locales, datetimeformat.Options{})
pr, err := pluralrules.New(locales, pluralrules.Options{})
lf, err := listformat.New(locales, listformat.Options{})
rtf, err := relativetimeformat.New(locales, relativetimeformat.Options{})
df, err := durationformat.New(locales, durationformat.Options{})
dn, err := displaynames.New(locales, displaynames.Options{Type: gointl.String(displaynames.Language)})
col, err := collator.New(locales, collator.Options{})
seg, err := segmenter.New(locales, segmenter.Options{})
```

> **Why**: JavaScript can write `new Intl.NumberFormat(...)` because `Intl.NumberFormat` is a constructor property. Go cannot expose a package-level property that acts like a subpackage constructor without either hiding the real package or introducing a misleading factory. Type aliases are the current Go bridge for constructor-property parity: they preserve the visible namespace relationship while the constructor packages preserve typed options and package ownership.
>
> **Rejected**: root `NewNumberFormat`, `NewDateTimeFormat`, `NewPluralRules`, `NewListFormat`, `NewRelativeTimeFormat`, or `NewDurationFormat`. Those names are Go inventions, not ECMA-402 names, and duplicate constructor package APIs.
>
> **Rejected**: deleting root constructor aliases because direct constructor packages already exist. Direct packages are the production construction path, but the root package still represents the ECMA-402 `Intl` namespace object; removing constructor properties to reduce aggregate imports would make the facade less faithful.

### 1.3 Import-cost policy

Because the root package aliases every active constructor type, it imports every active constructor package. This is intentional namespace fidelity, but it makes the root package an aggregate import.

Rules:

1. Performance-sensitive applications that need one formatter should import that formatter subpackage directly. The optional scalar pointer helpers live in the zero-dependency leaf package `github.com/agentable/go-intl/option` (`option.Int`/`option.Bool`/`option.String`); single-formatter services set options through it without importing the aggregate root. The root re-exports them as `gointl.Int`/`gointl.Bool`/`gointl.String` for namespace fidelity.
2. Root package benchmarks, build-size reports, and dependency graph reports must label the result as aggregate facade cost.
3. Per-surface measurements must be reported separately from root package measurements.
4. Constructor aliases must not be removed from the root package to reduce aggregate import cost or make dependency reports look smaller.
5. An active constructor alias can disappear only if the corresponding constructor leaves the active ECMA-402 surface, or if a future Go API design preserves constructor-property parity more faithfully than aliases and updates this SPEC first.
6. Root one-shot helpers and public cache controls remain forbidden; they are not an acceptable response to facade import cost.

> **Why**: JavaScript `Intl` exposes constructor properties on one namespace object. Go can mirror that shape only by importing the constructor packages that provide the aliased types. Hiding those imports would make the documented namespace shape misleading, while treating the root result as a single-formatter cost would mislead performance work.
>
> **Rejected**: trimming root constructor aliases to make `go list -deps .` look smaller. The production answer for one formatter is direct subpackage import, not a less honest root facade.

---

## 2. Forbidden Root APIs

The following symbols are outside the long-term public surface:

- `type Intl struct`
- `func New(...) *Intl`
- `type Option`
- `func WithTimeZone(...) Option`
- root typed one-shot helpers:
  - `FormatNumberInt`
  - `FormatNumberInt64`
  - `FormatNumberUint`
  - `FormatNumberUint64`
  - `FormatNumberFloat64`
  - `FormatNumberDecimal`
  - `FormatDate`
  - `FormatTime`
  - `FormatRange`
  - `SelectPluralInt`
  - `SelectPluralInt64`
  - `SelectPluralUint`
  - `SelectPluralUint64`
  - `SelectPluralFloat64`
  - `SelectPluralDecimal`
  - `FormatList`
  - `FormatRelativeTime`
  - `FormatDuration`
- public cache controls:
  - `WithCache`
  - `WithoutCache`
  - `ResetGlobalCache`
  - `Cache`
- root diagnostic APIs such as `Version()`; CLDR / ICU / tzdata pins are implementation metadata, not ECMA-402 `Intl` namespace members.

> **Why**: JavaScript `Intl` is not a constructor and has no per-locale session object. Root one-shot helpers are convenience wrappers, not namespace members. They make the root package drift toward an application helper facade, which is not the ECMA-402 contract.

Convenience helpers may live in examples or consumer code. They must not define the core API.

---

## 3. `GetCanonicalLocales`

`GetCanonicalLocales` is the root namespace equivalent of `Intl.getCanonicalLocales`.

Required behavior:

1. Preserve input order after canonicalization.
2. Remove duplicates by canonical locale string.
3. Return canonical `locale.Locale` values.
4. Do not perform locale availability matching.
5. Do not apply formatter-specific `localeMatcher`.

Raw string validation remains in `locale.Parse`:

```go
loc, err := locale.Parse("zh-hans-cn")
if err != nil {
    return err
}
canonical := gointl.GetCanonicalLocales(locale.List{loc})
```

> **Why**: ECMA-402 `CanonicalizeLocaleList` handles JavaScript dynamic values, strings, and `Intl.Locale` objects. Go public APIs should not accept `any` to simulate that dynamic boundary. The semantic operation is still the same after `locale.Parse` has converted strings to typed values.

---

## 4. Supported Value Accessors

The root supported-value accessors are the Go typed equivalents of
`Intl.supportedValuesOf`. Go uses one named function per ECMA-402 key rather
than a string-key dispatch function.

Supported accessors:

| Function | Source |
|----------|--------|
| `SupportedCalendars` | generated CLDR calendar identifiers plus ECMA-402 required constants such as `iso8601` |
| `SupportedCollations` | active `collator` backend collation identifiers that can be truthfully applied through locale-scoped Collator collation requests, currently sourced from `golang.org/x/text/collate.Supported()` |
| `SupportedCurrencies` | generated CLDR / ISO 4217 currency identifiers |
| `SupportedNumberingSystems` | ECMA-402 simple digit numbering systems plus generated CLDR numbering-system identifiers |
| `SupportedTimeZones` | generated primary IANA time-zone identifiers |
| `SupportedUnits` | ECMA-402 sanctioned single unit identifiers |

Required behavior:

1. Return sorted, unique, canonical string values.
2. Return independent slices; callers may mutate the returned slice without changing later calls.
3. Source all values from generated data, ECMA-402 constants, or generator manifests; no ad hoc runtime lists.
4. `SupportedNumberingSystems` must include the ECMA-402 simple digit universe even when the current CLDR profile has only `latn` symbol payloads; digit substitution remains supported by the abstract-operation layer.

Implementation rules:

1. Keep these functions in the root `gointl` package; do not move them to public `cldr`, `ecma402`, or `supported` packages.
2. Keep source packages private: CLDR-backed values come from `internal/cldr`, active collation capability comes from `internal/collation`, and sanctioned unit identifiers come from `internal/ecma402`.
3. Contract tests may verify the root package uses those owned data sources, but must not require the accessors to live in `intl.go`.

---

## 5. Constructor Package Static Methods

ECMA-402 constructor static methods remain owned by their constructor packages:

```go
numberformat.SupportedLocalesOf(locales locale.List, opts numberformat.Options) (locale.List, error)
datetimeformat.SupportedLocalesOf(locales locale.List, opts datetimeformat.Options) (locale.List, error)
pluralrules.SupportedLocalesOf(locales locale.List, opts pluralrules.Options) (locale.List, error)
listformat.SupportedLocalesOf(locales locale.List, opts listformat.Options) (locale.List, error)
relativetimeformat.SupportedLocalesOf(locales locale.List, opts relativetimeformat.Options) (locale.List, error)
durationformat.SupportedLocalesOf(locales locale.List, opts durationformat.Options) (locale.List, error)
displaynames.SupportedLocalesOf(locales locale.List, opts displaynames.Options) (locale.List, error)
collator.SupportedLocalesOf(locales locale.List, opts collator.Options) (locale.List, error)
segmenter.SupportedLocalesOf(locales locale.List, opts segmenter.Options) (locale.List, error)
```

Mapping:

| Go | ECMA-402 |
|----|----------|
| `numberformat.SupportedLocalesOf` | `Intl.NumberFormat.supportedLocalesOf` |
| `datetimeformat.SupportedLocalesOf` | `Intl.DateTimeFormat.supportedLocalesOf` |
| `pluralrules.SupportedLocalesOf` | `Intl.PluralRules.supportedLocalesOf` |
| `listformat.SupportedLocalesOf` | `Intl.ListFormat.supportedLocalesOf` |
| `relativetimeformat.SupportedLocalesOf` | `Intl.RelativeTimeFormat.supportedLocalesOf` |
| `durationformat.SupportedLocalesOf` | `Intl.DurationFormat.supportedLocalesOf` |
| `displaynames.SupportedLocalesOf` | `Intl.DisplayNames.supportedLocalesOf` |
| `collator.SupportedLocalesOf` | `Intl.Collator.supportedLocalesOf` |
| `segmenter.SupportedLocalesOf` | `Intl.Segmenter.supportedLocalesOf` |

The root package must not duplicate these methods. This preserves package ownership and avoids root-level formatter option re-exports.

`SupportedLocalesOf` accepts at most one options object. The only option it reads is `localeMatcher`; invalid values and multiple option objects return errors matching root `ErrInvalidOption`.

---

## 6. Error Model

The root package owns the public Intl error vocabulary. Formatter packages construct and return those same categories while keeping construction and formatting behavior in their own packages.

Rules:

1. Root exposes `ErrorKind`, `Error`, and the seven category sentinels: `ErrInvalidOption`, `ErrUnsupportedOption`, `ErrInvalidValue`, `ErrInvalidCode`, `ErrInvalidKey`, `ErrUnsupportedLocale`, and `ErrUnsupportedBackend`.
2. Formatter constructor and runtime errors match these root sentinels through `errors.Is`; callers do not need formatter-owned category sentinels.
3. `Error` exposes `Kind`, `Owner`, `Name`, `Value`, `Locale`, optional `Expected`, and wrapped `Err` context for `errors.AsType`.
4. `ErrUnsupportedOption`, `ErrUnsupportedLocale`, and `ErrUnsupportedBackend` also match `errors.ErrUnsupported`.
5. Human-facing error text uses `expected ...; got ...` and must not expose ECMA-402 abstract-operation names.
6. No fallback formatting output is produced by the root package.

---

## 7. Acceptance Criteria

- [ ] `go doc github.com/agentable/go-intl` presents the package as the `Intl` namespace, not a constructor.
- [ ] The root package has no `Intl` struct and no root `New`.
- [ ] The root package exposes no typed one-shot formatting helpers.
- [ ] The root package exposes active constructor aliases as the Go bridge for ECMA-402 `Intl` constructor properties.
- [ ] Constructor aliases are not removed for import-cost or build-size reasons; aggregate facade cost is measured and documented separately.
- [ ] The root package exposes no public cache controls.
- [ ] The root package exposes the canonical Intl error categories and no formatter-specific error aliases.
- [ ] README documents direct constructor package imports as the preferred production path for services that need one formatter.
- [ ] Root package build-size and dependency graph measurements are labeled as aggregate facade cost and kept separate from per-surface package measurements.
- [ ] `GetCanonicalLocales` preserves order, canonicalizes, and deduplicates.
- [ ] `GetCanonicalLocales` is the only public locale-list canonicalization helper; root and `locale` do not expose ECMA-402 abstract-operation helpers.
- [ ] Supported-value accessors cover exactly the ECMA-402 keys in this spec and return generated canonical values.
- [ ] Supported-value accessors live in the root package, conventionally in `supported.go`, without creating public data-layer packages.
- [ ] The root package does not expose `SupportedValueKey`, `SupportedValue*` constants, or `SupportedValuesOf`.
- [ ] `numberformat`, `datetimeformat`, `pluralrules`, `listformat`, `relativetimeformat`, `durationformat`, `displaynames`, `collator`, and `segmenter` own their own `SupportedLocalesOf` functions.
- [ ] `go list -deps ./...` confirms `go-intl` does not depend on `messageformat-go`.

---

## References

- `.references/ecma402/spec/intl.html`
- `.references/ecma402/spec/negotiation.html`
- SPEC 10 — Locale
- SPEC 11 — Locale Matching
- SPEC 20 — NumberFormat
- SPEC 30 — DateTimeFormat
- SPEC 40 — PluralRules
- SPEC 41 — ListFormat
- SPEC 42 — RelativeTimeFormat
- SPEC 43 — DurationFormat
- SPEC 44 — DisplayNames
- SPEC 45 — Collator
- SPEC 46 — Segmenter
- SPEC 50 — CLDR Data
