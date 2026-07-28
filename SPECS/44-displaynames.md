# 44 — DisplayNames

Status: active
Owns: `displaynames` package public API, type/style/fallback resolution, CLDR localenames data binding.

References:
- ECMA-402: `.references/ecma402/spec/displaynames.html`
- readable polyfill: `.references/formatjs/packages/intl-displaynames/`

---

## 1. Public Surface

```go
package displaynames

type Type string
const (
    Language      Type = "language"
    Region        Type = "region"
    Script        Type = "script"
    Currency      Type = "currency"
    Calendar      Type = "calendar"
    DateTimeField Type = "dateTimeField"
)

type Style string
const (
    LongStyle   Style = "long"
    ShortStyle  Style = "short"
    NarrowStyle Style = "narrow"
)

type Fallback string
const (
    CodeFallback Fallback = "code"
    NoneFallback Fallback = "none"
)

type LanguageDisplay string
const (
    DialectLanguageDisplay  LanguageDisplay = "dialect"
    StandardLanguageDisplay LanguageDisplay = "standard"
)

type Options struct {
    LocaleMatcher   *string
    Type            *string
    Style           *string
    Fallback        *string
    LanguageDisplay *string
}

type DisplayNames struct{ /* immutable resolved options + lookup data */ }

func New(locales locale.List, opts Options) (*DisplayNames, error)
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
func (d *DisplayNames) Of(code string) (string, bool, error)
func (d *DisplayNames) ResolvedOptions() ResolvedOptions
```

MUST rules:

1. `New` accepts one `Options` value. `Options.Type` is required; `nil` means the required option is missing, and `gointl.String("")` returns `ErrInvalidOption` with name `"type"`.
2. String options use presence-aware `*string` fields. `nil` means omitted/default: `LocaleMatcher=BestFitLocaleMatcher`, `Style=LongStyle`, `Fallback=CodeFallback`, and `LanguageDisplay=DialectLanguageDisplay`. Explicit empty strings are invalid for `localeMatcher`, `type`, `style`, `fallback`, and `languageDisplay`. `ResolvedOptions.LanguageDisplay` is reported only when `Type=Language`; for other types it must be absent (`nil`), mirroring ECMA-402 §12.4.2 step 22.
3. `Of` returns `(name, true, nil)` for a successful lookup. When the code is unknown and `Fallback=CodeFallback`, return the input code canonicalized per type with `true`. When unknown and `Fallback=NoneFallback`, return `("", false, nil)`.
4. `Of` validates the code shape per type before lookup (Unicode `unicode_language_id` for language codes, ISO 3166 region, ISO 15924 script, ISO 4217 currency, BCP 47 calendar, ECMA-402 dateTimeField). Language codes allow language/script/region/variant subtags only; extensions, private-use tags, grandfathered tags, and extlang forms are invalid for `DisplayNames.of`. Invalid shape returns an error wrapping `ErrInvalidCode` and carries the constructor-resolved locale in the structured error record.
5. `DisplayNames` is immutable after construction and safe for concurrent use.

> **Why `(string, bool, error)`**: ECMA-402 returns `string | undefined` for successful validation but throws `RangeError` for invalid code shape. Go needs both pieces: `(string, bool)` bridges `string | undefined`, and `error` bridges the throw path.

---

## 2. Resolved Options

```go
type ResolvedOptions struct {
    Locale          locale.Locale
    Style           Style
    Type            Type
    Fallback        Fallback
    LanguageDisplay *LanguageDisplay
}
```

`LanguageDisplay` is reported only when `Type=Language`, mirroring ECMA-402 `InitializeDisplayNames` step 22. The pointer distinguishes a JavaScript-omitted property from the concrete `"dialect"` value. ResolvedOptions JSON field names and `omitempty` behavior follow [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Other Constructors](./73-json-records.md#other-constructors).

---

## 3. CLDR Data

Display-name data comes from generated CLDR payloads in `internal/cldr/displaynames/data.go`:

| Type | CLDR source | Notes |
|------|-------------|-------|
| Language | `cldr-localenames-full/main/<locale>/languages.json` | `-alt-short` keys map to `Style=ShortStyle`, `-alt-narrow` keys map to `Style=NarrowStyle`; both fall back to long. `LanguageDisplay=StandardLanguageDisplay` rebuilds region-suffixed tags (e.g. `en-GB`) by composing the bare language and territory through `localeDisplayPattern`. Script-suffixed tags fall through to the dialect entry. |
| Region | `cldr-localenames-full/main/<locale>/territories.json` | `-alt-short` mapped to ShortStyle. Numeric UN M.49 region codes are returned as-is when they are not a CLDR-carried macro-region (e.g. `of("840")` → `"840"`), matching V8/Node — go-intl does not synthesize an M.49→alpha-2 alias table. |
| Script | `cldr-localenames-full/main/<locale>/scripts.json` | |
| Currency | `cldr-numbers-full/main/<locale>/currencies.json`. All styles return the localized `displayName` (singular noun, exposed through `cldr.Locale.CurrencyCanonicalName`); `of("USD")` is `"US Dollar"` for long/short/narrow alike, matching V8/Node. Currency *symbols* (`$`, `symbol-alt-narrow`) are not exposed through DisplayNames — that surface belongs to NumberFormat. | |
| Calendar | `cldr-localenames-full/main/<locale>/localeDisplayNames.json#types.calendar`. ECMA-402 calendar keys are aliased to CLDR keys at lookup time (`gregory` → `gregorian`, `ethioaa` → `ethiopic-amete-alem`). | |
| DateTimeField | `cldr-dates-full/main/<locale>/dateFields.json` `fields.<field>.displayName`. CLDR field keys are normalized to ECMA-402 names (`week` → `weekOfYear`, `zone` → `timeZoneName`); `-short` and `-narrow` suffixes feed the corresponding styles. | |

Lookup walks only the resolved data locale's truncation parent chain (`en-US` → `en`). Missing names remain missing so the public `Fallback` option can return the canonical code or absence; runtime lookup never borrows a name from an unrelated locale.

Generation MUST:

1. Live in `tools/gen-cldr/cldr/displaynames.go`, `tools/gen-cldr/extract/displaynames.go`, and `tools/gen-cldr/codegen/displaynames.go`.
2. Use the project-wide locale profile (`tools/locale-profile.json` → `locales`).
3. Emit a single generated file at `internal/cldr/displaynames/data.go` that wires `displayNamesData()` and `displayNamesSupportedLocales()` through a `sync.Once` lazy initializer.

---

## 4. Errors

- `gointl.ErrInvalidOption`: missing `Type`, invalid enum value, or unsupported `LocaleMatcher`.
- `ErrInvalidCode`: invalid `Of` input code shape for the resolved display-name type.

`Of` returns structured `*gointl.Error` values for invalid input code shape. Missing localized data is not an error; it surfaces through the `(string, bool)` return according to `Fallback`. Public error text follows SPEC 12's `expected ...; got ...` rule.

---

## 5. Static Supported Locales

```go
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
```

MUST rules:

1. Use `internal/cldr/displaynames.SupportedLocales()` as the supported set.
2. Call `localematcher.FilterLocalesWithMaximizer`.
3. Accept one `Options` value; `Options{}` represents omitted static-method options.
4. Read only `LocaleMatcher`; ignore other options for this static method. `nil` means omitted and an explicit empty string is invalid.

---

## 6. Typed Bridges and Divergences

| JavaScript | Go | Class |
|------------|----|-------|
| `displayNames.of(code) -> string \| undefined` or throws `RangeError` | `Of(code) (string, bool, error)` | Typed bridge |
| `new Intl.DisplayNames(locales, options)` with optional second arg | `New(locales locale.List, opts Options)`; use `Options{}` for the omitted or empty object case | Typed bridge |

Accepted divergences live in `displaynames/testdata/divergences.md` when added.
