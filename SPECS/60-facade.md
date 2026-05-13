# SPEC 60 — Root `Intl` Namespace

> **Status:** Revised (2026-05-13)
> **Type:** Consumer API Spec — defines the public entry surface for the root `go-intl` package.
> **Authority:** This spec is the SSOT for the root package. SPECS 10/20/30/40 own the active constructor packages. ECMA-402 `.references/ecma402/spec/intl.html` is the normative source.

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

Therefore the root Go package is **not** a per-locale `Intl` session, not a `createIntl` clone, and not a holder for root one-shot formatting helpers.

Formatter construction and behavior live in the active constructor packages:

- `locale`
- `numberformat`
- `datetimeformat`
- `pluralrules`

The root package may provide discoverability aliases and static common namespace functions. It must not contain formatting logic.

Unimplemented ECMA-402 constructors (`Collator`, `DisplayNames`, `DurationFormat`, `ListFormat`, `RelativeTimeFormat`, `Segmenter`) are intentionally absent until their packages and SPECS exist. Absence is preferable to placeholder aliases that imply unsupported behavior.

---

## 1. Root Package Shape

### 1.1 Public namespace functions

```go
package gointl

type SupportedValueKey string

const (
    SupportedValueCalendar        SupportedValueKey = "calendar"
    SupportedValueCollation       SupportedValueKey = "collation"
    SupportedValueCurrency        SupportedValueKey = "currency"
    SupportedValueNumberingSystem SupportedValueKey = "numberingSystem"
    SupportedValueTimeZone        SupportedValueKey = "timeZone"
    SupportedValueUnit            SupportedValueKey = "unit"
)

func GetCanonicalLocales(locales ...locale.Locale) []locale.Locale
func SupportedValuesOf(key SupportedValueKey) ([]string, error)
```

Mapping:

| Go | ECMA-402 |
|----|----------|
| `GetCanonicalLocales` | `Intl.getCanonicalLocales(locales)` after locale parsing has occurred at the Go boundary |
| `SupportedValuesOf` | `Intl.supportedValuesOf(key)` |

`GetCanonicalLocales` accepts `locale.Locale` values because Go should parse raw strings once at the boundary. If callers need to canonicalize raw tags, they call `locale.Parse` or `locale.MustParse` first. This is the Go typed bridge for ECMA-402 `CanonicalizeLocaleList`.

### 1.2 Active constructor aliases

The root package may expose type aliases matching constructor properties for discoverability:

```go
type Locale = locale.Locale
type NumberFormat = numberformat.NumberFormat
type DateTimeFormat = datetimeformat.DateTimeFormat
type PluralRules = pluralrules.PluralRules
type PluralCategory = pluralrules.Category
```

Construction still belongs to the constructor packages:

```go
loc := locale.MustParse("en-US")
nf, err := numberformat.New(loc)
dtf, err := datetimeformat.New(loc)
pr, err := pluralrules.New(loc)
```

> **Why**: JavaScript can write `new Intl.NumberFormat(...)` because `Intl.NumberFormat` is a constructor property. Go cannot expose a package-level property that acts like a subpackage constructor without either hiding the real package or introducing a misleading factory. Aliases preserve the visible namespace relationship while the constructor packages preserve typed options and package ownership.
>
> **Rejected**: root `NewNumberFormat`, `NewDateTimeFormat`, or `NewPluralRules`. Those names are Go inventions, not ECMA-402 names, and duplicate constructor package APIs.

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
- public cache controls:
  - `WithCache`
  - `WithoutCache`
  - `ResetGlobalCache`
  - `Cache`
- root diagnostic APIs such as `Version()`; CLDR / ICU / tzdata pins are implementation metadata, not ECMA-402 `Intl` namespace members.

> **Why**: JavaScript `Intl` is not a constructor and has no per-locale session object. Root one-shot helpers are convenience wrappers, not namespace members. They make the root package drift toward FormatJS `createIntl`, which is not the ECMA-402 contract.

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
canonical := gointl.GetCanonicalLocales(loc)
```

> **Why**: ECMA-402 `CanonicalizeLocaleList` handles JavaScript dynamic values, strings, and `Intl.Locale` objects. Go public APIs should not accept `any` to simulate that dynamic boundary. The semantic operation is still the same after `locale.Parse` has converted strings to typed values.

---

## 4. `SupportedValuesOf`

`SupportedValuesOf` is the root namespace equivalent of `Intl.supportedValuesOf`.

Supported keys:

| Key | Source |
|-----|--------|
| `"calendar"` | generated CLDR calendar identifiers plus ECMA-402 required constants such as `iso8601` |
| `"collation"` | generated CLDR BCP47 canonical collation identifiers filtered to supported sort collations |
| `"currency"` | generated CLDR / ISO 4217 currency identifiers |
| `"numberingSystem"` | ECMA-402 simple digit numbering systems plus generated CLDR numbering-system identifiers |
| `"timeZone"` | generated primary IANA time-zone identifiers |
| `"unit"` | ECMA-402 sanctioned single unit identifiers |

Required behavior:

1. Return sorted, unique, canonical string values.
2. Return a package sentinel error for unsupported key values, corresponding to ECMA-402 `RangeError`.
3. Source all values from generated data, ECMA-402 constants, or generator manifests; no ad hoc runtime lists.
4. `numberingSystem` must include the ECMA-402 simple digit universe even when the current CLDR profile has only `latn` symbol payloads; digit substitution remains supported by the abstract-operation layer.

---

## 5. Constructor Package Static Methods

ECMA-402 constructor static methods remain owned by their constructor packages:

```go
numberformat.SupportedLocalesOf(locales []locale.Locale, opts ...numberformat.Options) ([]locale.Locale, error)
datetimeformat.SupportedLocalesOf(locales []locale.Locale, opts ...datetimeformat.Options) ([]locale.Locale, error)
pluralrules.SupportedLocalesOf(locales []locale.Locale, opts ...pluralrules.Options) ([]locale.Locale, error)
```

Mapping:

| Go | ECMA-402 |
|----|----------|
| `numberformat.SupportedLocalesOf` | `Intl.NumberFormat.supportedLocalesOf` |
| `datetimeformat.SupportedLocalesOf` | `Intl.DateTimeFormat.supportedLocalesOf` |
| `pluralrules.SupportedLocalesOf` | `Intl.PluralRules.supportedLocalesOf` |

The root package must not duplicate these methods. This preserves package ownership and avoids root-level formatter option re-exports.

`SupportedLocalesOf` accepts at most one options object. The only option it reads is `localeMatcher`; invalid values and multiple option objects return the owning package's `ErrInvalidOption` sentinel.

---

## 6. Error Model

Root namespace functions do not own formatter errors.

Rules:

1. `SupportedValuesOf` returns an error wrapping root `ErrInvalidKey` when the key is not one of the ECMA-402 supported strings.
2. Formatter constructor errors are owned by formatter packages.
3. Root package sentinels may alias lower-level sentinels for caller convenience, but they must not hide the owning package.
4. No fallback formatting output is produced by the root package.

---

## 7. Acceptance Criteria

- [ ] `go doc github.com/agentable/go-intl` presents the package as the `Intl` namespace, not a constructor.
- [ ] The root package has no `Intl` struct and no root `New`.
- [ ] The root package exposes no typed one-shot formatting helpers.
- [ ] The root package exposes no public cache controls.
- [ ] `GetCanonicalLocales` preserves order, canonicalizes, and deduplicates.
- [ ] `SupportedValuesOf` supports exactly the ECMA-402 keys in this spec and returns generated canonical values.
- [ ] `numberformat`, `datetimeformat`, and `pluralrules` own their own `SupportedLocalesOf` functions.
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
- SPEC 50 — CLDR Data
