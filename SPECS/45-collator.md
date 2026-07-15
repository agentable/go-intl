# 45 — Collator

Status: active
Owns: `collator` package public API, mapping of ECMA-402 options to `golang.org/x/text/collate`, locale resolution against CLDR collation data.

References:
- ECMA-402 Collator constructor, options, resolved options, and `compare` method.
- Locale matching and Unicode extension handling are shared with SPEC 11.
- Underlying comparison backend: `golang.org/x/text/collate`.

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
    LocaleMatcher     *string
    Usage             *string
    Sensitivity       *string
    CaseFirst         *string
    Numeric           *bool
    IgnorePunctuation *bool
    Collation         *string
}

type Collator struct{ /* immutable resolved options + locale tag + collate options */ }

func New(locales locale.List, opts Options) (*Collator, error)
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
func (c *Collator) Compare(x, y string) int
func (c *Collator) ResolvedOptions() ResolvedOptions
```

MUST rules:

1. `Compare` returns a negative, zero, or positive integer, matching `bytes.Compare`, `strings.Compare`, and the comparator shape that `slices.SortFunc` expects.
2. `LocaleMatcher`, `Usage`, `Sensitivity`, `CaseFirst`, `Numeric`, `IgnorePunctuation`, and `Collation` use pointer presence fields so that zero `Options{}` means "no caller preference", while `LocaleMatcher: gointl.String(collator.LookupLocaleMatcher)` selects lookup, `Usage: gointl.String(collator.SearchUsage)` requests search tailoring, `Sensitivity: gointl.String(collator.BaseSensitivity)` selects a comparison strength, `CaseFirst: gointl.String(collator.FalseCaseFirst)` explicitly overrides locale `kf`, `Numeric: gointl.Bool(false)` can explicitly override locale `kn`, and `Collation: gointl.String("")` is an explicit invalid option instead of omitted input. Constructors copy pointee values into internal config.
2a. `New` accepts one `Options` value; `Options{}` represents omitted or empty JS options.
3. `New` MUST construct and freeze one `*collate.Collator` backend from the resolved data locale and options. `Compare` MUST serialize access to that cached backend because `golang.org/x/text/collate.Collator` mutates private iterators while comparing; the wrapping `*Collator` remains safe for concurrent use.
4. `Collator` is immutable after construction except for the private comparison mutex.
5. `Collation` and `CaseFirst` are resolved through ECMA-402 locale negotiation. Backend-supported `co` values are applied through the private `collate` tag and reflected in `ResolvedOptions().Collation`; well-formed unsupported `co` and locale `kf` extension values fall back to the active default resolved options. Explicit `Options.CaseFirst = gointl.String(upper|lower)` returns `ErrUnsupportedOption` because it is an option value the active backend cannot apply truthfully.

### 1.1 Current support tier

Current tier: **locale-scoped backend capability** for collation specializations,
with narrowed gaps for the rows below.

The active backend advertises collation specializations through
`golang.org/x/text/collate.Supported()` Unicode-extension tags. The root
`Intl.supportedValuesOf("collation")` equivalent returns the sorted global set
of those backend-applied specializations, excluding ECMA-402 reserved values
`default`, `standard`, and `search`; the current supported set includes
`phonebk`. Constructor locale data remains locale-scoped: `de-u-co-phonebk` and
`Options{Collation: gointl.String("phonebk")}` for German resolve to
`ResolvedOptions().Collation == "phonebk"`, while the same well-formed value for
a locale without backend support falls back to `"default"`.

The dependency evidence for keeping these gaps narrow lives in
`reports/golang.org-x-text.md`.
native-engine backend-proof fixtures under `collator/testdata/conformance/node-v26/`
cover the default sort behavior the active backend can already apply; option
contract fixtures plus XFAIL entries cover behavior that must not be advertised
until the backend proves it.

Supported option precedence:

- `kn` is active. Locale `kn=true` sets numeric comparison unless an explicit `Options.Numeric` value overrides it.
- `Options.Numeric: gointl.Bool(false)` is an explicit caller preference and must override locale `-u-kn-true`.
- `kf` locale extensions remain a capability gap. Unsupported `co` values may be parsed as locale negotiation inputs, but they must not be reported as supported unless the backend applies the requested behavior.

| Gap | Current behavior | Rationale | review_after | Removal path |
|-----|------------------|-----------|--------------|--------------|
| `usage = "search"` | Constructor returns `ErrUnsupportedOption`. | Search collation must not pretend to be sort collation; ECMA-402 says search data may have different behavior. | 2026-09-30 | Identify a CLDR/x/text-backed search tailoring path, add native comparison fixtures, then accept `SearchUsage`. |
| explicit `caseFirst = "upper" \| "lower"` | Constructor returns `ErrUnsupportedOption`. | The active backend cannot yet control case-level direction truthfully, and explicit options should not pretend to be applied. | 2026-09-30 | Add backend support or a documented dependency report, then verify resolved options and ordering fixtures. |
| `ignorePunctuation = true` | Constructor returns `ErrUnsupportedOption`. | UCA alternate-shifted uses collation variable weights; deleting Unicode punctuation/space categories changes semantics and the active backend does not implement shifted handling. | 2026-09-30 | Add a backend with proved alternate-shifted behavior, then verify native ordering/resolved fixtures before accepting `true`. |
| explicit collation option reflected in `ResolvedOptions().Locale` | Supported explicit collation options apply to comparison and `ResolvedOptions().Collation`, while the resolved locale tag remains the base matched locale. | ECMA-402 ResolveLocale and FormatJS clear supported locale keywords when an option value overrides them; Node v26 reflects this option in the resolved locale tag, so the native fixture stays XFAIL as an observed engine divergence. | 2026-09-30 | Keep the Node witness under XFAIL unless the normative spec changes; do not change the shared resolver merely to mirror this native tag detail. |

---

## 2. Option to `collate` mapping

| ECMA-402 option | x/text/collate option | Notes |
|-----------------|-----------------------|-------|
| `sensitivity = "base"` | `collate.IgnoreCase`, `collate.IgnoreDiacritics` | |
| `sensitivity = "accent"` | `collate.IgnoreCase` | |
| `sensitivity = "case"` | `collate.IgnoreDiacritics` | |
| `sensitivity = "variant"` | (no options) | Default for `usage="sort"`. |
| `numeric = true` | `collate.Numeric` | |
| `numeric = false` with locale `kn=true` | (no options) | Explicit option overrides locale extension; resolved locale drops the `kn` extension when false is selected. |
| `caseFirst = "false"` | (no options) | Default case order. |
| explicit `caseFirst = "upper" \| "lower"` | constructor error | Returns `ErrUnsupportedOption`; active collation backend cannot apply case-level direction. |
| locale `kf=upper\|lower` | (no options) | Unsupported locale extension value falls back to default `caseFirst=false`. |
| `ignorePunctuation = true` | constructor error | Returns `ErrUnsupportedOption`; `x/text` v0.40.0 stubs UCA alternate-shifted handling and Unicode category deletion is not an equivalent implementation. |
| `usage = "search"` | constructor error | Current implementation gap; returns `ErrUnsupportedOption` until real search tailoring exists. |
| backend-supported `collation = "<value>"` | BCP 47 `co=<value>` on the private `collate` tag | Locale-scoped backend support; currently proves German `phonebk` through manual and Node witness fixtures. |
| well-formed unsupported `collation = "<value>"` | (no options) | Negotiation input; unsupported values fall back to resolved collation `"default"`. |
| malformed `collation = "<value>"` | constructor error | Returns `ErrInvalidOption`; invalid Unicode locale extension type syntax is caller-fixable input. |

Rows marked constructor error are active backend refusals, not accepted divergences. Accepted collation support requires backend behavior, supported-values evidence, and resolved-options fixtures in the same change.

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

- `Usage = SortUsage`; explicit `SearchUsage` returns `ErrUnsupportedOption`.
- `Sensitivity = VariantSensitivity` when omitted.
- `CaseFirst = FalseCaseFirst`; explicit `upper` / `lower` returns `ErrUnsupportedOption`, while locale `kf=upper|lower` falls back to `false`.
- `Collation = "default"` unless active locale-scoped backend data adopts a requested `co` value; supported German `phonebk` resolves to `"phonebk"`.
- `Numeric` defaults to `false`; `IgnorePunctuation` defaults to `false` and explicit `true` returns `ErrUnsupportedOption` until the backend can execute UCA alternate-shifted behavior.
- JSON field names and presence behavior follow [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Other Constructors](./73-json-records.md#other-constructors); `collation` is always present because ECMA-402 reports `"default"` when no backend specialization applies.

---

## 4. Errors

- `gointl.ErrInvalidOption`: invalid `LocaleMatcher`, `Usage`, `Sensitivity`, or `CaseFirst`.
- `gointl.ErrUnsupportedOption`: valid but unimplemented explicit `usage=search`, `caseFirst=upper|lower`, or `ignorePunctuation=true`. Well-formed unsupported collation requests are fallback inputs, not errors.

Constructor and `SupportedLocalesOf` failures expose `*gointl.Error` and follow SPEC 12's `expected ...; got ...` text rule. `Compare` does not return errors. Strings that fail UTF-8 validation are compared by replacement-rune behavior of `x/text/collate`.

---

## 5. Static Supported Locales

```go
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
```

MUST rules:

1. Use `internal/collation.SupportedLocales()` as the supported set (drops `und` and Unicode-extension forms returned by `x/text/collate.Supported()`). The package lives outside `internal/cldr/` because its data is sourced from `golang.org/x/text/collate`, not CLDR.
2. Call `localematcher.FilterLocalesWithMaximizer`.
3. Preserve requested Unicode extension forms when the base locale is supported. `SupportedLocalesOf` is a locale availability check, not a collation-tailoring support check.
4. Accept one `Options` value; `Options{}` represents omitted static-method options.
5. Read only `LocaleMatcher`; `nil` means omitted and an explicit empty string is invalid.
6. Use `internal/collation.SupportedCollationsForLocale` as Collator `co` locale data so specialization support remains tied to the backend locale that advertised it.

---

## 6. Typed Bridges and Gap Boundaries

| JavaScript | Go | Class |
|------------|----|-------|
| `collator.compare(x, y) -> -1 \| 0 \| 1` | `Compare(x, y) int` with any negative / positive value | Typed bridge (matches `slices.SortFunc`). |
| `caseFirst` reflected in tailoring | `upper` / `lower` rejected with `ErrUnsupportedOption` | Narrowed implementation gap; see §1.1. |
| `collation` reflected in tailoring | backend-supported locale-scoped values such as German `phonebk` are applied through `co=<value>` | Implemented behavior. |
| `ignorePunctuation` reflected in comparison | `true` is rejected with `ErrUnsupportedOption` | Narrowed implementation gap; category deletion is not UCA alternate-shifted. |
| `usage = "search"` distinct tailoring | rejected with `ErrUnsupportedOption` | Narrowed implementation gap; see §1.1. |

Accepted divergences must be enumerated in `collator/testdata/divergences.md` when they are added.
