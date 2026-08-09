# SPEC 41 — ListFormat

> **Status:** Draft (2026-05-14)
> **Owner:** `listformat/` + `internal/cldr/list` + `tools/gen-cldr/`
> **Reference contract:** `.references/ecma402/spec/listformat.html` first, then `formatjs/packages/intl-listformat/`

## Overview

Defines the active `listformat.ListFormat` public API, constructor option model, CLDR list-pattern data contract, `format`, `formatToParts`, and `supportedLocalesOf` behavior for `Intl.ListFormat`.

This SPEC owns only `Intl.ListFormat`. `Intl.RelativeTimeFormat` is owned by SPEC 42 and `Intl.DurationFormat` is owned by SPEC 43. It does not promote unrelated constructors or add root one-shot helpers.

This SPEC does not redefine:

- Locale parsing and matching -> [SPEC 10](./10-locale.md), [SPEC 11](./11-locale-matching.md)
- CLDR profile and generator architecture -> [SPEC 50](./50-cldr-data.md)
- Root namespace alias policy -> [SPEC 60](./60-facade.md)
- Fixture and divergence policy -> [SPEC 70](./70-conformance.md)

---

## 1. Public API

```go
package listformat

type LocaleMatcher string
type Type string
type Style string
type PartType string

const (
    LookupLocaleMatcher  LocaleMatcher = "lookup"
    BestFitLocaleMatcher LocaleMatcher = "best fit"

    Conjunction Type = "conjunction"
    Disjunction Type = "disjunction"
    Unit        Type = "unit"

    LongStyle   Style = "long"
    ShortStyle  Style = "short"
    NarrowStyle Style = "narrow"

    PartElement PartType = "element"
    PartLiteral PartType = "literal"
)

type Options struct {
    LocaleMatcher *string
    Type          *string
    Style         *string
}

type ResolvedOptions struct {
    Locale locale.Locale
    Type   Type
    Style  Style
}

type Part struct {
    Type  PartType
    Value string
}

type ListFormat struct{ /* immutable resolved options + selected templates */ }

func New(locales locale.List, opts Options) (*ListFormat, error)
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
func (f *ListFormat) Format(list []string) string
func (f *ListFormat) FormatToParts(list []string) []Part
func (f *ListFormat) ResolvedOptions() ResolvedOptions
```

MUST rules:

1. `New` accepts one `Options` value. `New(locales, Options{})` matches JavaScript `new Intl.ListFormat(locales, {})` or omitted options.
2. `Options{}` defaults to `type="conjunction"` and `style="long"`.
3. String options are presence-aware: `nil` means omitted, while a non-nil pointer is an explicit option value. Explicit empty strings are invalid option values instead of silently selecting defaults.
4. `ListFormat` is immutable after construction. All methods on `*ListFormat` must be safe for concurrent callers.
5. `ResolvedOptions` returns a value snapshot.
6. `Format` and `FormatToParts` accept `[]string`. JavaScript's non-string iterable error path is represented by Go's static type boundary, not by accepting `[]any`.
7. `SupportedLocalesOf` belongs to `listformat`, not the root package.
8. JSON field names and `omitempty` behavior follow [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Other Constructors](./73-json-records.md#other-constructors).

> **Why**: `Intl.ListFormat.prototype.format` accepts an iterable and then rejects non-string elements at runtime. Go can encode the same boundary as `[]string`, avoiding `any` without changing observable list formatting semantics.
>
> **Rejected**: `Format(list ...string)` as the primary API. Variadic calls are convenient for hand-written examples but make forwarding an existing slice less explicit and diverge from the "one list argument" JavaScript owner shape.

---

## 2. Constructor and Options

`New` must follow ECMA-402 `Intl.ListFormat(locales, options)` after the Go boundary has parsed `locales` into one `locale.Locale`.

Pipeline:

1. Validate at most one options object.
2. Read `localeMatcher`, default `best fit`, allowed `lookup | best fit`; `nil` means omitted and `gointl.String("")` is invalid.
3. Resolve locale against `internal/cldr/list.SupportedLocales()` with no relevant Unicode extension keys.
4. Read `type`, default `conjunction`, allowed `conjunction | disjunction | unit`; `nil` means omitted and `gointl.String("")` is invalid.
5. Read `style`, default `long`, allowed `long | short | narrow`; `nil` means omitted and `gointl.String("")` is invalid.
6. Load the selected CLDR template set for resolved data locale, type, and style.

MUST rules:

1. Invalid `localeMatcher`, `type`, or `style` returns an error wrapping `ErrInvalidOption`.
2. Constructor errors must include the option name, user value, and locale when useful.
3. `ShortStyle` and `NarrowStyle` are valid for every type. Do not preserve old MDN/Generated reference prose that says short/narrow only pair with unit; ECMA-402 current spec permits all `type` and `style` combinations when data exists.
4. Locale resolution must use `localematcher.ResolveLocale` / `FilterLocalesWithMaximizer` patterns already used by existing constructor packages.

---

## 3. CLDR Data

`tools/gen-cldr` must extract CLDR list pattern data from `cldr-misc-full/main/<locale>/listPatterns.json` for every tag in `tools/locale-profile.json`'s `locales` list.

Data shape:

```go
package cldr

type ListPatterns struct {
    Pair   string
    Start  string
    Middle string
    End    string
}
```

Mapping:

| ECMA-402 type/style | CLDR key |
|---------------------|----------|
| `conjunction` / `long` | `listPattern-type-standard` |
| `conjunction` / `short` | `listPattern-type-standard-short` |
| `conjunction` / `narrow` | `listPattern-type-standard-narrow` |
| `disjunction` / `long` | `listPattern-type-or` |
| `disjunction` / `short` | `listPattern-type-or-short` |
| `disjunction` / `narrow` | `listPattern-type-or-narrow` |
| `unit` / `long` | `listPattern-type-unit` |
| `unit` / `short` | `listPattern-type-unit-short` |
| `unit` / `narrow` | `listPattern-type-unit-narrow` |

MUST rules:

1. Generated list supported locales must be derived from actual list pattern payload maps.
2. Each generated template string must be a syntactically valid placeholder pattern and contain `{0}` and `{1}` exactly once.
3. Runtime formatting must never read CLDR JSON files.
4. Accessors must go through `internal/cldr/list`; public `listformat` code must not read generated package variables directly.
5. If a CLDR locale lacks a short or narrow variant, generation may fall back to the corresponding long variant only if this fallback is documented in a divergence or generator test.

> **Why**: ListFormat output is almost entirely data-driven. A hand-written list of supported locales would claim support before payloads exist; deriving support from payload maps keeps `SupportedLocalesOf` honest.

---

## 4. Formatting Algorithm

`Format` and `FormatToParts` must implement ECMA-402 `CreatePartsFromList`, `FormatList`, and `FormatListToParts`.

MUST rules:

1. Empty input returns `""` and an empty parts slice.
2. One element returns that element and a single `{Type: PartElement, Value: element}` part.
3. Two elements use the selected `Pair` template.
4. Three elements fold from the end: last element, then `End`, then `Start`.
5. Four or more elements fold from the end: last element, `End`, zero or more `Middle`, then `Start`.
6. `Format(list)` must equal the concatenation of `FormatToParts(list)[i].Value`.
7. Template decomposition must use `internal/ecma402.PartitionPattern`; do not hand-parse placeholders in `listformat`.
8. Part records must use only `element` and `literal`.

The implementation may use a single deterministic template set per locale/type/style. ECMA-402 permits implementation-defined selection among multiple template records for context-sensitive languages; this project does not add that complexity until CLDR extraction exposes multiple alternatives.

---

## 5. Static Supported Locales

```go
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
```

MUST rules:

1. Use `internal/cldr/list.SupportedLocales()` as the supported set.
2. Call `localematcher.FilterLocalesWithMaximizer`.
3. Accept one `Options` value; `Options{}` represents omitted static-method options.
4. Read only `LocaleMatcher`; ignore formatting options for this static method. `nil` means omitted and an explicit empty string is invalid.
5. Invalid locale matcher returns `ErrInvalidOption`.

---

## 6. Errors

```go
var ErrInvalidOption error
```

MUST rules:

1. `ErrInvalidOption` classifies invalid constructor options and invalid `SupportedLocalesOf` options.
2. Formatting methods do not return errors because `[]string` eliminates JavaScript's non-string iterable failure path.
3. Errors must be matchable through `errors.Is`.
4. Constructor and `SupportedLocalesOf` errors expose `*gointl.Error` and follow SPEC 12's `expected ...; got ...` text rule.

---

## 7. Root Namespace

After `listformat` package tests, generated data checks, README, and conformance fixtures pass, the root package may add:

```go
type ListFormat = listformat.ListFormat
```

The root package must not add `NewListFormat`, `FormatList`, cache controls, or one-shot list helpers.

---

## 8. Testing

MUST rules:

1. Use stdlib `testing`, table-driven tests, and `t.Parallel()` unless shared generated-output state prevents it.
2. Add focused unit tests for constructor defaults, invalid options, empty/one/two/many-element formatting, parts joining, and supported locales.
3. Add generator tests for CLDR list pattern extraction and generated supported locales.
4. Add generated-reference conformance fixtures under `listformat/testdata/conformance/formatjs/`.
5. Accepted output mismatches must go to `listformat/testdata/divergences.md` or `xfail.json`, never by removing generated fixture cases.

Acceptance checks:

- [ ] `go test -race ./listformat/...`
- [ ] `(cd tools/gen-cldr && go test ./...)`
- [ ] `task data:check`
- [ ] `task test`

---

## Forbidden

- No root-level `FormatList` or `NewListFormat`.
- No runtime CLDR JSON loading.
- No public cache controls.
- No `[]any` or `interface{}` formatting API.
- No hand-written `SupportedLocalesOf` locale list.
- No implementation of other ECMA-402 constructors as part of this package.
