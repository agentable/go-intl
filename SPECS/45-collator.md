# 45 — Collator

Status: active
Owns: `collator` package public API, mapping of ECMA-402 options to `golang.org/x/text/collate`, locale resolution against CLDR collation data.

References:
- ECMA-402: `.references/ecma402/spec/collator.html`
- FormatJS polyfill: `.references/formatjs/packages/intl-localematcher/` (only locale matching; FormatJS does not polyfill Collator)
- Underlying engine: `golang.org/x/text/collate`

---

## 1. Public Surface

```go
package collator

type Usage string
const (
    SortUsage   Usage = "sort"
    SearchUsage Usage = "search"
)

type Sensitivity string
const (
    BaseSensitivity    Sensitivity = "base"
    AccentSensitivity  Sensitivity = "accent"
    CaseSensitivity    Sensitivity = "case"
    VariantSensitivity Sensitivity = "variant"
)

type CaseFirst string
const (
    UpperCaseFirst CaseFirst = "upper"
    LowerCaseFirst CaseFirst = "lower"
    FalseCaseFirst CaseFirst = "false"
)

type Options struct {
    LocaleMatcher     LocaleMatcher
    Usage             Usage
    Sensitivity       Sensitivity
    CaseFirst         CaseFirst
    Numeric           *bool
    IgnorePunctuation *bool
    Collation         string
}

type Collator struct{ /* immutable resolved options + locale tag + collate options */ }

func New(locales locale.List, opts Options) (*Collator, error)
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
func (c *Collator) Compare(x, y string) int
func (c *Collator) ResolvedOptions() ResolvedOptions
```

MUST rules:

1. `Compare` returns a negative, zero, or positive integer, matching `bytes.Compare`, `strings.Compare`, and the comparator shape that `slices.SortFunc` expects.
2. `Numeric` and `IgnorePunctuation` use `*bool` presence fields so that zero `Options{}` means "no caller preference", while `Numeric: gointl.Bool(false)` can explicitly override locale `kn`. Constructors copy pointee values into internal config.
2a. `New` accepts one `Options` value; `Options{}` represents omitted or empty JS options.
3. Each `Compare` call MUST acquire a fresh `*collate.Collator` (via `collate.New`). `golang.org/x/text/collate.Collator` is NOT safe for concurrent use; the wrapping `*Collator` is.
4. `Collator` is immutable after construction.
5. `Collation` and `CaseFirst` may be supplied through the Unicode `co` and `kf` extensions only after the implementation can truthfully apply them. In the active scope, unsupported collation/case-first requests on the matched locale return `ErrUnsupportedOption`; extensions on fallback candidates that are not selected must not affect construction. This is a narrowed implementation gap, not a permanent API philosophy.

### 1.1 Current support tier

Current tier: **narrowed implementation gap** for the rows below.

Because explicit collation tailoring is currently rejected, the root
`Intl.supportedValuesOf("collation")` equivalent and unrestricted
`Locale.GetCollations()` return an empty list. `Locale` instances with an
explicit `co` extension still return that explicit value from `GetCollations()`
because ECMA-402 locale info reflects the locale tag before constructor support
is considered.

The dependency evidence for keeping these gaps narrow lives in
`reports/golang.org-x-text.md`.

| Gap | Current behavior | Rationale | review_after | Removal path |
|-----|------------------|-----------|--------------|--------------|
| `usage = "search"` | Constructor returns `ErrUnsupportedOption`. | Search collation must not pretend to be sort collation; ECMA-402 says search data may have different behavior. | 2026-09-30 | Identify a CLDR/x/text-backed search tailoring path, add Node comparison fixtures, then accept `SearchUsage`. |
| `caseFirst = "upper" \| "lower"` | Constructor returns `ErrUnsupportedOption`. | The active backend cannot yet control case-level direction truthfully. | 2026-09-30 | Add backend support or a documented dependency report, then verify resolved options and ordering fixtures. |
| non-default `collation` | Constructor returns `ErrUnsupportedOption`. | Advertising CLDR collation identifiers without applying tailoring would overpromise. | 2026-09-30 | Map supported CLDR collation identifiers to backend behavior, add supportedValues/resolvedOptions fixtures, and keep unsupported identifiers rejected. |

---

## 2. Option to `collate` mapping

| ECMA-402 option | x/text/collate option | Notes |
|-----------------|-----------------------|-------|
| `sensitivity = "base"` | `collate.IgnoreCase`, `collate.IgnoreDiacritics` | |
| `sensitivity = "accent"` | `collate.IgnoreCase` | |
| `sensitivity = "case"` | `collate.IgnoreDiacritics` | |
| `sensitivity = "variant"` | (no options) | Default for `usage="sort"`. |
| `numeric = true` | `collate.Numeric` | |
| `caseFirst = "false"` | (no options) | Default case order. |
| `caseFirst = "upper" \| "lower"` | constructor error | Returns `ErrUnsupportedOption`; active collation backend cannot apply case-level direction. |
| `ignorePunctuation = true` | BCP 47 `ka=shifted` on the private `collate` tag | Uses `golang.org/x/text/collate` UCA alternate-shifted handling; does not rewrite input strings. |
| `usage = "search"` | constructor error | Current implementation gap; returns `ErrUnsupportedOption` until real search tailoring exists. |
| well-formed `collation = "<value>"` | constructor error | Current implementation gap; returns `ErrUnsupportedOption` until requested collation tailoring can be applied. |
| malformed `collation = "<value>"` | constructor error | Returns `ErrInvalidOption`; invalid Unicode locale extension type syntax is caller-fixable input. |

Rows marked constructor error are current implementation gaps, not accepted divergences. They must move to implemented behavior or an explicit dependency report before `collator` can claim complete ECMA-402 option coverage.

---

## 3. Resolved Options

```go
type ResolvedOptions struct {
    Locale            locale.Locale
    Usage             Usage
    Sensitivity       Sensitivity
    CaseFirst         CaseFirst
    Collation         string
    Numeric           bool
    IgnorePunctuation bool
}
```

Defaults at construction:

- `Usage = SortUsage`
- `Usage = SortUsage`; explicit `SearchUsage` returns `ErrUnsupportedOption`.
- `Sensitivity = VariantSensitivity` when omitted.
- `CaseFirst = FalseCaseFirst`; explicit `upper` / `lower` or locale `kf=upper|lower` returns `ErrUnsupportedOption`.
- `Collation = "default"`; explicit `default` and locale `co=default|standard|search` resolve to default collation, while real tailoring requests such as `co=phonebk` return `ErrUnsupportedOption` until tailoring is real.
- `Numeric` defaults to `false`; `IgnorePunctuation` defaults to `false` and explicit `true` is reflected in comparison and resolved options.
- JSON field names and `omitempty` behavior follow [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Other Constructors](./73-json-records.md#other-constructors).

---

## 4. Errors

- `gointl.ErrInvalidOption`: invalid `LocaleMatcher`, `Usage`, `Sensitivity`, or `CaseFirst`.
- `gointl.ErrUnsupportedOption`: valid but unimplemented `usage=search`, `caseFirst=upper|lower`, or collation tailoring.

Constructor and `SupportedLocalesOf` failures expose `*gointl.Error` and follow SPEC 12's `expected ...; got ...` text rule. `Compare` does not return errors. Strings that fail UTF-8 validation are compared by replacement-rune behavior of `x/text/collate`.

---

## 5. Static Supported Locales

```go
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
```

MUST rules:

1. Use `internal/collation.SupportedLocales()` as the supported set (drops `und` and Unicode-extension forms returned by `x/text/collate.Supported()`). The package lives outside `internal/cldr/` because its data is sourced from `golang.org/x/text/collate`, not CLDR.
2. Call `localematcher.FilterLocalesWithMaximizer`.
3. Do not advertise requested Unicode extension forms that the active constructor support boundary rejects: non-default `co` tailoring and `kf=upper|lower`.
4. Accept one `Options` value; `Options{}` represents omitted static-method options.
5. Read only `LocaleMatcher`.

---

## 6. Typed Bridges and Gap Boundaries

| JavaScript | Go | Class |
|------------|----|-------|
| `collator.compare(x, y) -> -1 \| 0 \| 1` | `Compare(x, y) int` with any negative / positive value | Typed bridge (matches `slices.SortFunc`). |
| `caseFirst` reflected in tailoring | `upper` / `lower` rejected with `ErrUnsupportedOption` | Narrowed implementation gap; see §1.1. |
| `ignorePunctuation` reflected in comparison | `true` maps to `x/text/collate` alternate-shifted handling through `ka=shifted` | Implemented behavior. |
| `usage = "search"` distinct tailoring | rejected with `ErrUnsupportedOption` | Narrowed implementation gap; see §1.1. |

All future accepted divergences must be enumerated in `collator/testdata/divergences.md` when they are added.
