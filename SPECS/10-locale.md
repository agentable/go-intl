# SPEC 10 — Locale

> **Status:** Revised (2026-10-21)
> **Priority:** High (all formatter input types; blocking SPEC 11 / 20 / 30 / 40 / 60)
> **Authority:** ECMA-402 `.references/ecma402/spec/locale.html` is a normative source. This SPEC records the current of `locale.Locale` type, ECMA-402 `Intl.Locale` alignment, parsing and normalization, `Maximize` / `Minimize`, read-only property getter, `String` / `Equal` / `MarshalText` / `UnmarshalText` Go contracts.

---

## Overview

`locale.Locale` is the immutable Go representation of `Intl.Locale` and the
input identity for formatter constructors. `internal/localeid` owns Unicode
syntax and canonical strings, using pinned CLDR alias data. Registry coverage
never determines grammatical validity. `language.Tag` is an explicit, fallible
interop projection, not identity storage.

This SPEC defines the `Locale` structure, constructor (`New` / `FromTag` / `Parse`), locale-list helper (`ParseList`), normalization (`Maximize` / `Minimize`), getter materialization strategy, string round trip, comparability. **Do not** define best-fit matching algorithm (SPEC 11), CLDR data format (SPEC 50), formatter internal slot (SPEC 12 / 20 / 30 / 40).

---

## 1. Locale Structure <a id="locale-structure"></a>

### 1.1 Decision: canonical identity and immutable extension state

Locale stores a canonical private `localeid.Tag` without its separately owned
Unicode extension, immutable canonical keyword/attribute state, and cached full
canonical text. The zero value is `und`. L/S/R/variant getters project the tag;
keyword getters project the extension. There is no mutable parallel identity.

`BaseName()` omits extensions. `Tag() (language.Tag, error)` includes the **full**
identity and succeeds only when `FromTag(tag, Options{})` preserves `Equal`.
Lossy or unsupported x/text conversion returns `ErrInvalidValue`. For example,
`Parse("und-Gara")` succeeds while its `Tag()` bridge can fail. `String`/`Parse`
and text/JSON round trips work for every valid Locale. See
[the dependency decision](../reports/golang.org-x-text.md).

### 1.2 Canonical keyword state

The private extension state owns one canonical keyword map and Unicode attributes.
Map presence distinguishes an absent key from a present empty type; string output,
getters and candidate lists project that same state. It is built only during
construction and is immutable afterwards, so copied Locale values remain safe.

String getters return the type, using `""` for both absent and present-empty values
as the existing Go bridge. `Numeric()` is true only for a present empty `kn` type
(the canonical spelling of `true`); absent, `false`, and unknown types return false.
`String()` retains the distinction, and candidate lists return a singleton for a
present keyword even when its type is empty. Constructor option pointers separately
distinguish omission from an explicit option value.

---

## 2. Construction & Parsing

### 2.1 Entry signature

```go
// locale/locale.go (signature)
package locale

// Parse Parse Locale from BCP 47 string. Supports -u- extended keys ca/co/hc/kf/kn/nu/fw.
// Parsing failure returns an error matching gointl.ErrInvalidValue; boundary unification occurs here in one go.
func Parse(s string) (Locale, error)

// New constructs Locale from BCP 47 locale identifier, aligned with Options Intl.Locale(tag, options).
// The set extension field will enter the String() canonical output that returns Locale.
func New(tag string, opts Options) (Locale, error)

// FromTag is an explicit Go bridge from language.Tag to Intl.Locale.
func FromTag(tag language.Tag, opts Options) (Locale, error)

func ParseList(tags ...string) (List, error)

type Options struct {
    Language        *string
    Script          *string
    Region          *string
    Variants        *string
    Calendar        *string
    Collation       *string
    HourCycle       *string
    CaseFirst       *string
    Numeric         *bool
    NumberingSystem *string
    FirstDayOfWeek  *string
}
```

`language.Tag` is the only x/text type exposed publicly. `FromTag` converts its
input through the canonical Locale boundary; `Tag` returns `(language.Tag,
error)` and includes Unicode extensions. Callers normally use Locale directly.

BCP 47 language, script, region, and variant subtag shape checks live in
`internal/localeid`; constructor options, locale parsing, and internal code
canonicalization must not grow parallel grammar copies. Locale info methods that
interpret Unicode subdivision values (`rg` / `sd` region prefixes), locale
matching available-locale aliases, and CLDR accessors that interpret script or
region subtags (`DateLocaleData` hour-cycle lookup and DisplayNames
language-region composition) also reuse `internal/localeid` rather than carrying
package-local ASCII/length grammar.

Calling example:

```go
loc, err := locale.Parse("en-US-u-hc-h23-ca-buddhist")
fmt.Println(loc.HourCycle())   // "h23"
fmt.Println(loc.Calendar())    // "buddhist"

loc2, err := locale.New("ja", locale.Options{Calendar: gointl.String("japanese")})
fmt.Println(loc2.String())   // "ja-u-ca-japanese"
```

### 2.2 Parsing behavioral contracts

`Parse` checks the original Unicode language identifier before Unicode extension
separation and alias replacement. The language, optional
script, optional region, and variants must have Unicode subtag shapes in that
order. BCP 47 compatibility syntax such as `i-klingon`, `en-GB-oed`, and `zh-cmn`
is rejected rather than accepted through its canonical replacement. `New` and
`ParseList` use the same boundary and retain the original input in invalid-value
errors. Underscores are rejected throughout the original input, including
extensions and private use, before the backend can normalize them to hyphens.
Acceptance is covered by `locale/locale_test.go:TestParseInvalidLocale`.

The same raw check rejects case-insensitive duplicate variants in both the base
language identifier and a transformed language identifier, and duplicate
extension singletons. It stops interpreting subtags at private use. Repeated
Unicode keys keep their first value; repeated transformed field keys remain
grammatically valid. These are not duplicate singletons. The check shares the
subtag grammar in `internal/localeid` and does not perform alias selection or
replace the extension parsers.

| Input | Behavior |
|------|------|
| `"en-US"` | OK;`Tag = en-US`, all extended fields are empty |
| `"en-US-u-hc-h23"` | OK; `HourCycle="h23"`, `String()` retains `-u-hc-h23`, a successful `Tag()` bridge also retains the Unicode extension |
| `"en-US-u-ca-buddhist-hc-h23"` | OK; The field order is normalized according to spec (see §3.2) |
| `"en-US-u-ca-islamicc"` | OK; the deprecated CLDR value canonicalizes to `Calendar="islamic-civil"` |
| `"en-US-u-ca-islamic-civil"` | OK; the canonical value remains unchanged |
| `"en-US-u-ca-gregorian"` | Returns an error matching `gointl.ErrInvalidValue`; the nine-character subtag is not a well-formed Unicode type |
| `"en-US-u-ms-imperial"` | OK; key-aware CLDR canonicalization produces `-u-ms-uksystem` |
| `"xx-INVALID"` | Accepted and canonicalized to `xx-invalid`; unknown language and variant subtags are valid identity. |
| `""` | Returns an error matching `gointl.ErrInvalidValue` (empty string is illegal) |
| `"en-US-u-kn"`(no value table true) | OK;`Numeric=true` |
| `"en-US-u-kn-false"` | OK;`Numeric=false` |
| `"en-US-u-kf-false"` | OK;`CaseFirst="false"` (spec literal is legal) |
| `"en-US-u-fw-0"` | Returns an error matching `gointl.ErrInvalidValue`; numeric weekday aliases belong to the `FirstDayOfWeek` constructor option and are not well-formed Unicode type subtags |

> **Why**:
> 1. **Unique boundary** - Parse is BCP 47 boundary, the public API no longer accepts raw `string` locale (SPEC 11 / 20 / 30 / 40 input parameters all `Locale`).
> 2. **Key-aware canonicalization** - ECMA-402 `CanonicalizeUValue` canonicalizes a value for its specific Unicode key. The pinned CLDR BCP47 data therefore owns mappings such as `ca=islamicc` → `islamic-civil` and `ms=imperial` → `uksystem`; the same spelling under an unrelated key is not rewritten. Syntax validation runs before alias lookup, so an alias cannot make malformed input valid.
> 3. **No public panic parser** - Test and example code can define local helpers, but production API exposes `Parse` / `ParseList` error returns only.
>
> **Rejected**:
> - **Receive `string` at formatter layer** (`numberformat.New("en-US", ...)`): Formatter constructors receive `locale.List`; callers parse strings at the boundary. `language.Tag` enters only through `locale.FromTag`.
> - **`Parse(s) Locale`(panic on error)**: spec RangeError must return error, violating CLAUDE.md "no panic in production code".

### 2.3 Construction phase verification

`Parse` validates tag syntax and retains every canonical Unicode keyword, including unknown legal types and keys with an empty type. It does not apply constructor-option enums to tags. `New` validates only explicit options before replacing those keywords:

| Field | Validation Rules |
|------|---------|
| `Language` / `Script` / `Region` | When the pointer is non-nil: the explicit string must be non-empty and form a valid BCP 47 language identifier with the retained subtags |
| `Calendar` | When the pointer is non-nil: the explicit string must be non-empty and match Unicode type syntax (one or more 3–8 character alphanumeric subtags); **does not** verify whether it exists in CLDR (formatter layer is responsible) |
| `Collation` | When the pointer is non-nil: same as above |
| `HourCycle` | When the pointer is non-nil: an exact, case-sensitive member of {`"h11"`, `"h12"`, `"h23"`, `"h24"`} |
| `CaseFirst` | When the pointer is non-nil: an exact, case-sensitive member of {`"upper"`, `"lower"`, `"false"`} |
| `Numeric` | `nil` means omitted; non-nil `true` writes `-u-kn`; non-nil `false` writes `-u-kn-false` |
| `NumberingSystem` | When the pointer is non-nil: the explicit string must be non-empty and match BCP 47 type subtag syntax |
| `FirstDayOfWeek` | A supplied value must be non-empty. Apply ECMA-402 `WeekdayToUValue` to numeric aliases `"0"` through `"7"`, then validate Unicode locale type syntax. Unknown well-formed types such as `"funday"` are valid; this is not a weekday enum. `Parse` never applies option coercion to a tag. |

Verification failure returns an error matching the root error category. Parsing failed to match `gointl.ErrInvalidValue`; constructor option failed to match `gointl.ErrInvalidOption`.

> **Why**: One-time verification during construction, the formatter layer no longer verifies the syntax twice; but whether it can be parsed in CLDR (such as `Calendar="vulcan"`) will be reported by the formatter when searching for data (`numberformat`/`datetimeformat` is sentinel by itself).
>
> **Rejected**: Check the existence of CLDR data during the construction period (hierarchy violation: the construction period does not rely on CLDR; only the formatter layer requires CLDR data).

---

## 3. String Round-trip & Canonicalization

### 3.1 `String()` Contract

```go
// (Locale).String() returns canonical BCP 47 representation, including -u- extension.
// Reciprocal with Parse: mustLocale(loc.String()).Equal(loc) == true.
func (l Locale) String() string
```

### 3.2 Canonicalization rules

`String()` Output **Required**:

1. **subtag order**: `language[-script][-region][-variants...]`; variants are sorted after CLDR aliases.
2. **`-u-` extended keys in dictionary order**: `ca` < `co` < `fw` < `hc` < `kf` < `kn` < `nu`.
3. **`-u-` extended value lowercase** (`Calendar="GREGORY"` output `-u-ca-gregory`).
4. **Absent keywords are omitted; present empty types remain as bare keys**, for example `en-u-ca`; constructor option pointers with explicit `""` are rejected before `String()` canonicalization.
5. **Numeric=true outputs `-u-kn`** (no value table true, consistent with spec); explicit Numeric=false outputs `-u-kn-false`; omitted Numeric does not output.
6. **CaseFirst=`"false"` outputs `-u-kf-false`** (literal legal value).
7. **Key-aware Unicode type canonicalization** follows pinned CLDR BCP47 data after syntax validation (`ca=islamicc` → `islamic-civil`, `ms=imperial` → `uksystem`).
8. **Duplicate Unicode extension keys are first-wins**, matching ECMA-402 `UnicodeExtensionComponents` and native engine behavior (`en-u-ca-buddhist-ca-gregory` canonicalizes to `en-u-ca-buddhist`).
9. **Extension singletons are sorted; Unicode extension insertion precedes later singletons and private use** (`en-u-ca-gregory-x-private`, not `en-x-private-u-ca-gregory`).

Explicit unknown subtags are identity fields: `en-Zzzz-ZZ` has Script `Zzzz`,
Region `ZZ`, and BaseName `en-Zzzz-ZZ`. Language/script/region options replace
only the selected field, preserving the others and all suffixes. Getters project
the existing canonical tag; Locale does not store a second mutable identity.
`locale/language_identity_test.go:TestExplicitUnknownLanguageSubtags` and
`locale/construct_test.go:TestNewLanguageOptionsPreserveExplicitUnknownSubtags`
cover getters, JSON, option overrides, and suffix preservation.

Example:

| Construction | `String()` Output |
|------|-----------------|
| `Parse("en-US")` | `"en-US"` |
| `Parse("en-US-u-hc-h23")` | `"en-US-u-hc-h23"` |
| `Parse("en-us-U-Ca-Islamicc-Hc-H23")` | `"en-US-u-ca-islamic-civil-hc-h23"` |
| `Parse("en-US-u-kn")` | `"en-US-u-kn"` |
| `Parse("en-US-u-kn-false")` | `"en-US-u-kn-false"` |
| `New("ja", Options{Calendar: gointl.String("japanese"), HourCycle: gointl.String("h23")})` | `"ja-u-ca-japanese-hc-h23"` |

### 3.3 `MarshalText` / `UnmarshalText`

```go
// (Locale).MarshalText() / UnmarshalText() implement encoding.TextMarshaler / TextUnmarshaler.
// Equivalent to String() / Parse(), used for JSON / YAML / config files.
func (l Locale) MarshalText() ([]byte, error)
func (l *Locale) UnmarshalText(text []byte) error
```

> **Why**: `encoding.TextMarshaler` is the Go standard interface; after implementation, `json.Marshal(loc)` / `yaml.Marshal(loc)` will automatically use canonical String(), without the need for the consumer to write a custom marshaler.
>
> **Rejected**: `MarshalJSON` / `UnmarshalJSON` both coexist (redundant, `encoding/json/v2` automatically falls back to TextMarshaler).

### 3.4 JSON host-boundary records

`locale.Locale` JSON encoding is the canonical BCP 47 string via `MarshalText`. Locale info records use ECMA-402 field names:

```go
type WeekInfo struct {
    FirstDay time.Weekday   `json:"firstDay"`
    Weekend  []time.Weekday `json:"weekend"`
}

type TextInfo struct {
    Direction *string `json:"direction,omitempty"`
}
```

`WeekInfo.MarshalJSON` must emit ECMA-402 weekday numbers, Monday=1 through Sunday=7, instead of Go's `time.Weekday` values where Sunday=0. `TextInfo` marshals as `{"direction":"ltr"}` or `{"direction":"rtl"}` when CLDR has a known direction; an unknown direction is nil and marshals as `{}`. Locale JSON field names follow [SPEC 73 §Part and Locale Info Records](./73-json-records.md#3-part-and-locale-info-records).

> **Why**: JS host bindings and API adapters expect `Intl.Locale.prototype.getWeekInfo()` records, not Go enum ordinals. Keeping Go's `time.Weekday` in memory and ECMA-402 numbers on the wire preserves both sides.
>
> **Rejected**: Exposing a second public `WeekInfoJSON` type. One record with a custom marshal boundary is simpler and avoids parallel APIs.

---

## 4. Maximize & Minimize <a id="4-maximize--minimize"></a>

### 4.1 Entry signature

```go
// (Locale).Maximize adds likely subtags (language → script + region inference).
// Corresponds to ECMA-402 sec-addlikelysubtags + spec sec-intl.locale.prototype.maximize.
func (l Locale) Maximize() Locale

// (Locale).Minimize removes inferred subtags, and is the inverse of Maximize.
// Corresponds to sec-removelikelysubtags + sec-intl.locale.prototype.minimize.
func (l Locale) Minimize() Locale
```

Call example:

```go
loc := mustLocale("zh-Hant")
fmt.Println(loc.Maximize().String())  // "zh-Hant-TW"
fmt.Println(loc.Maximize().Minimize().String())  // "zh-TW"
```

### 4.2 Implementation strategy

Both operations consume the pinned `cldrlocale.MaximizeSubtags` data. Maximize
uses CLDR fallback order while preserving explicit language, script and region.
The unknown identifiers `Zzzz` and `ZZ` are treated as missing only at the
likely-subtag inference boundary, `cldrlocale.MaximizeSubtags`. Identity getters
retain them. For example, `en-Zzzz-ZZ` maximizes to `en-Latn-US` and minimizes to
`en`, preserving variants and extensions in both operations. These results follow
the pinned CLDR inference contract, independently of host ICU differences.
Minimize first obtains that maximal LSR, then tries its language alone,
language-region, and language-script. The first candidate that maximizes to the
same LSR wins; otherwise the maximal LSR remains. This selects the stable
region-preferred policy, for example `zh-Hant-TW` and `zh-Hant` both become
`zh-TW`. The maximal language is used even when the original input was `und`.

This follows ICU's candidate-equivalence algorithm in
`.references/node/deps/icu-small/source/common/loclikelysubtags.cpp` rather than
an inverse map chosen by source-key sort order. Sharing CLDR input alone does
not make two minimization algorithms equivalent. There is no separate generated
minimize table. Variants and all extensions are copied unchanged; only LSR changes.
The implementation does not use x/text's separate likely-subtag data.

### 4.3 Extended fields reserved

`Maximize` / `Minimize` **MUST** preserve the complete language-identifier suffix: variants, transformed extensions, Unicode attributes and keys (including fields not modeled by the seven getters), and private use. Only language, script, and region may change.

```go
loc := mustLocale("zh-u-hc-h23-ca-chinese")
m := loc.Maximize()
// m.HourCycle() == "h23" // Reserved
// m.Calendar() == "chinese" // Reserved
// m.String() == "zh-Hans-CN-u-ca-chinese-hc-h23"
```

> **Why**: Consistent with generated-reference `intl-locale/index.ts` `maximize()` implementation; extended fields are explicitly selected by the user and Maximize should not silently discard them.

---

## 5. Getter Materialization (Close SPEC 00 §8 Q3)

### 5.1 Decision: Simple field pre-parsing, candidate list method is lazy

| Type | Field/Method | Materialization Timing |
|------|------------|---------|
| **Simple getter**(spec `Intl.Locale.prototype.<name>` getter) | `Calendar()` / `Collation()` / `HourCycle()` / `CaseFirst()` / `Numeric()` / `NumberingSystem()` / `FirstDayOfWeek()` | **Prepared on construction** (read directly from the BCP 47 `-u-` extended key; zero additional cost) |
| **Candidate list method** (spec `Intl.Locale.prototype.get<Name>s()` method) | `GetCalendars()` / `GetCollations()` / `GetHourCycles()` / `GetNumberingSystems()` / `GetTimeZones()` / `GetWeekInfo()` / `GetTextInfo()` | **Calculated on each call**; `GetCollations()` projects explicit locale state, while the other methods read their owning generated data; not cached |

### 5.2 Candidate list method signature

```go
// locale/info.go(signature)

// GetCalendars returns a list of calendars preferred by this locale (sorted by priority).
// Corresponds to ECMA-402 sec-intl.locale.prototype.getCalendars.
func (l Locale) GetCalendars() []string

// GetCollations returns Locale.Collation() as a singleton when it is present.
// Without an explicit collation it returns the ECMA-402 no-match fallback
// ["emoji", "eor"]; no locale-specific Collator support is inferred.
func (l Locale) GetCollations() []string

// GetHourCycles returns hour cycle preferences (by priority).
func (l Locale) GetHourCycles() []string

// GetNumberingSystems returns numbering system preferences.
func (l Locale) GetNumberingSystems() []string

// GetTimeZones returns the sorted primary IANA identifiers assigned to the
// explicit region subtag; it returns nil when the locale has no region.
// An explicit region with no assigned zones returns a non-nil empty list.
func (l Locale) GetTimeZones() []string

// WeekInfo corresponds to ECMA-402 sec-week-info-of-locale.
type WeekInfo struct {
FirstDay time.Weekday // First day (default time.Monday)
Weekend []time.Weekday // Weekend (common [Saturday, Sunday])
}
func (l Locale) GetWeekInfo() WeekInfo

// TextInfo corresponds to ECMA-402 sec-text-info-of-locale.
type TextInfo struct {
    Direction *string // "ltr" | "rtl"; nil when unknown
}
func (l Locale) GetTextInfo() TextInfo
```

Call example:

```go
loc := mustLocale("ar-SA")
fmt.Println(loc.GetWeekInfo().FirstDay)    // time.Sunday
direction := loc.GetTextInfo().Direction
fmt.Println(direction != nil && *direction == "rtl") // true
fmt.Println(loc.GetCalendars())            // [gregory] (current active date data)
```

An explicit `ca` value is returned by `GetCalendars` before active-data preferences
are consulted. Without it, the list is restricted to the calendars currently
supported by the date payload. For weekday options,
`Options{FirstDayOfWeek: gointl.String("funday")}` is valid, and `"0"` normalizes
to `sun`; a tag containing `-u-fw-0` remains malformed.

> **Why**(Close SPEC 00 §8 Q3):
> 1. **Minimum construction period cost** - Simple fields are easily obtained during BCP 47 parsing, zero overhead + fields are read directly (O(1)).
> 2. **Low frequency of candidate list** - `getCalendars()`, etc. are only used in meta-information query scenarios (such as "display calendars supported by this locale"); most formatters call `loc.Calendar()` (singular getter) without adjusting `getCalendars()`.
> 3. **Value-preserving semantics** - `Locale` struct does not contain lazy cache fields; no `*Locale` pointer, `sync.Once`, internal mutex is required.
> 4. **YAGNI** - If the benchmark shows that a method is a hot spot, add `sync.Once` cache.
> 5. **Exactly consistent with Generated reference** - `getCalendars()` of `intl-locale/index.ts` is also not cached, and `maximize() + region table lookup` is used every time `calendarsOfLocale(this)` is called.
>
> **Rejected**:
> - **All pre-parsed (the candidate list is also stored in the struct)**: Increase the struct size by about 100+ bytes (7 `[]string`/`map`/`*WeekInfo` pointers); most callers do not read these fields, which is a waste of space.
> - **All lazy (simple getter re-parses BCP 47 each time)**: hot path performance loss, and `loc.Calendar()` is a high-frequency read.
> - **`sync.Once` cache candidate list**: `*Locale` must be used to mutate internally; value semantics are lost; and the CLDR table lookup itself is O(1), and the cache benefit is minimal.

### 5.3 Candidate list implementation path

Each candidate list method **MUST** use its owning generated data. Calendar,
hour-cycle, week, and text preferences use the locale/date CLDR domains.
`GetTimeZones` is different: ECMA-402 `TimeZonesOfLocale` reads only the
locale's explicit region, returns nil when no region exists, and projects the
sorted primary identifiers assigned by the generated `internal/tz` registry's
IANA `zone.tab` records. It must not maximize a missing region or maintain a
locale-layer country table.

An explicit unknown region such as `ZZ` is still present: `GetTimeZones` returns
an empty list, not nil, and region preference does not substitute subdivision or
likely-region values for it. Generated calendar/week data supplies its ordinary
no-data fallback. `GetTextInfo` likewise does not maximize an explicit `Zzzz`
script; the pinned script metadata marks its direction unknown, so Direction is
nil and omitted from JSON. These distinctions are covered by `locale/info_test.go`.

Subdivision region interpretation uses generated territory aliases. Legal but
unregistered regions such as `000` are retained and follow ordinary CLDR data
fallback; getters never call a Must parser on user-derived values. `rg` overrides
preferences only when that domain has data; otherwise the base region, `sd`,
likely region, and world fallback retain their existing precedence.


### 5.4 Explicit Calendar fields take precedence

If the calendar keyword is present, `GetCalendars()` returns that type as a singleton, including an empty type. The same presence rule applies to collation, hour-cycle and numbering-system candidate lists. String getters retain their Go zero-value bridge for absent/empty; `String()` and candidate lists preserve presence.

```go
loc := mustLocale("en-US-u-ca-buddhist")
fmt.Println(loc.GetCalendars())  // ["buddhist"]
```

---

## 6. Equality & Comparability

### 6.1 Decision: Explicit `Equal` + `String` comparison

```go
// Equal compares the complete canonical string.
func (l Locale) Equal(other Locale) bool

// (Locale).String returns canonical BCP 47; the same Locale String() results are consistent multiple times.
func (l Locale) String() string
```

### 6.2 `==` Incomparable

`Locale` is deliberately not Go-comparable. Use `Equal` or canonical `String`
keys; the fallible x/text bridge is not the identity comparison mechanism.

```go
a, _ := locale.Parse("en-US")
b, _ := locale.Parse("en-US")
// a == b // Compilation error or undefined behavior, use prohibited
fmt.Println(a.Equal(b))         // true
fmt.Println(a.String() == b.String()) // true (alternative)
```

> **Why**:
> 1. **`Locale` does not promise comparable** - extension state and future fields can evolve, and `==` is not written into the API contract.
> 2. **`Equal` explicit** is safer and clearer than `==`; the caller knows at a glance that the comparison is based on field semantics.
> 3. **`String()` comparison** is used as a fallback for map key: if the user needs `map[Locale]V`, he must change it to `map[string]V` + `loc.String()`.
>
> **Rejected**:
> - **Requirement `Locale` is `comparable`**: Exposing Go comparison semantics as a long-term API will limit extension state evolution.
> - **`reflect.DeepEqual`**: slow, possible misjudgment (slice order, zero value field).

### 6.3 Sorting

Sorting **should** be in lexicographic order `Locale.String()`; callers use `slices.SortFunc(locs, func(a, b Locale) int { return strings.Compare(a.String(), b.String()) })`. This SPEC does not provide the `Less` method.

---

## 7. Options Object and Read-Only Locale

### 7.1 Constructor Options

`Options` is the Go typed bridge of `Intl.Locale(tag, options)`. It only takes effect in the constructor boundary; `Locale` is a read-only value after returning.

```go
// locale/locale.go (signature)
type Options struct {
    Language        *string
    Script          *string
    Region          *string
    Variants        *string
    Calendar        *string
    Collation       *string
    HourCycle       *string
    CaseFirst       *string
    Numeric         *bool
    NumberingSystem *string
    FirstDayOfWeek  *string
}
```

Call example:

```go
loc, err := locale.New("en", locale.Options{
    Calendar:        gointl.String("gregory"),
    HourCycle:       gointl.String("h23"),
    Numeric:         gointl.Bool(true),
    NumberingSystem: gointl.String("latn"),
})
```

`nil` string option pointers mean the property was omitted. A non-nil pointer to `""` means the caller explicitly supplied an empty string and must be rejected with an invalid-option error, matching `Intl.Locale` option presence semantics.

`Variants *string` replaces the variant sequence and preserves t/u/x and other
extensions. Nil preserves existing variants. Supplied values must be nonempty,
contain only grammatical variants, and contain no case-insensitive duplicates.
Case and alias normalization follow the same owner as Parse. For example,
`New("de-1901-u-nu-latn", Options{Variants: gointl.String("1996")})` yields
`de-1996-u-nu-latn`. Input pointers are copied during construction.
`Locale.Variants()` returns a detached slice.

### 7.2 Verification timing

`New` accepts a `Options` value. `Options{}` represents an empty options object for ECMA-402; multiple options objects are not Go API shapes and are rejected by the compile time.

`Options` verification must reuse the verification logic of §2.3 in `New`, and an error matching `gointl.ErrInvalidOption` or `gointl.ErrInvalidValue` will be returned when it fails. No panic API is added to the production path.

> **Why**:
> 1. **Uniform Boundary** - `Parse` / `New` is the boundary where locale user input enters the system, and errors should be returned centrally here.
> 2. **ECMA-402 alignment** - JavaScript `Intl.Locale(tag, options)` has only one options object; Go typed `Options` retains the same shape.
> 3. **No implicit panic** - locale construction returns errors; tests and examples use local helpers when a hard-coded tag must be valid.
>
> **Rejected**:
> - `With*` constructor options: Split a JS options object into a string of Go closures and introduce execution order semantics.
> - `(Locale).WithCalendar(...)` immutable setters: JavaScript `Intl.Locale` is a read-only accessor object; the setter API will make the Go surface wider than native Intl.
> - `Options` silently discards illegal values: difficult to troubleshoot, inconsistent with spec behavior.

---

## 8. Errors

```go
// locale errors classify through root gointl sentinels:
// - gointl.ErrInvalidValue for parse / field-value failures
// - gointl.ErrInvalidOption for constructor option failures
// Public caller-fixable failures also expose *gointl.Error.
```

Error message convention:

```go
return Locale{}, intlerr.New(intlerr.InvalidValue, "locale", "languageTag", input, "", intlerr.ErrInvalidValue)
return Locale{}, intlerr.New(intlerr.InvalidOption, "locale", "hourCycle", hc, "", intlerr.ErrInvalidOption)
```

`errors.Is(err, gointl.ErrInvalidValue)` should return true if parsing fails; `errors.Is(err, gointl.ErrInvalidOption)` should return true if constructor option verification fails. Exposed error text takes the shape `expected ...; got ...` and must not expose internal ECMA-402 abstract operation names.

---

## Forbidden

- **formatter public API receives raw `string` locale**: destroys type boundaries and allows parsing to fail until the `Format` calling layer is exposed.
  - ✅ Do: `numberformat.New(mustLocaleList("en-US"), numberformat.Options{})` for const-like examples, or `locale.ParseList` at application boundaries.
  - ❌ Don't: `numberformat.New("en-US", numberformat.Options{})`.

- **Duplicate locale parsers or alias tables**: use `internal/localeid` and its generated CLDR data; formatter packages must not invent their own grammar.

- **Reimplement or fork the likelySubtags algorithm**: reinvent maximize/minimize instead of consuming generated CLDR data.
- ✅ Do: use generated `cldrlocale.MaximizeSubtags` data and prove minimal candidates by maximize equivalence.
- ❌ Don't: depend on `language.Tag.LikelyScript()` / `LikelyRegion()` (data version is not the go-intl CLDR pin) or hand-maintain a `likely_subtags.go` patch layer.

- **Construction period panic**: Violation of CLAUDE.md "no panic in production code".
- ✅ Do: `Parse` / `New` returns `(Locale, error)`.
- ❌ Don't: `Parse` panics on invalid input or exposes a public `Must*` parser.

- **`Locale` is designed as a `*Locale` pointer type**: breaks value semantics.
- ✅ Do: `Locale` is a struct value type; re-call `locale.New(tag, Options{...})` or `locale.Parse(tag)` when the locale extension needs to be changed.
  - ❌ Don't: `func (l *Locale) WithCalendar(c string)`(in-place mutate).

- **`Locale` implements Go `==` comparable**: identity equality is an explicit canonical-string contract.
- ✅ Do: `loc1.Equal(loc2)` or `loc1.String() == loc2.String()`.
- ❌ Don't: `loc1 == loc2` (compilation error or undefined behavior).

- **Check the existence of CLDR data during construction** (such as whether `Calendar="vulcan"` is in CLDR): Hierarchy violation.
- ✅ Do: Only BCP 47 syntax is verified during construction; data existence is determined by formatter reporting an error during data search.
- ❌ Don't: `import "internal/cldr"` inside `Parse` checks the calendar name.

- **Candidate list method returns fixed `[]string` literal** (ignoring region preference): breaks spec consistency.
- ✅ Do: `GetCalendars()` walks `Maximize().region` + `internal/cldr` preference table.
  - ❌ Don't: `func (l Locale) GetCalendars() []string { return []string{"gregory"} }`.

- **`getCalendars` / `getWeekInfo` and other candidate list methods are cached into struct fields**: destroy value semantics.
- ✅ Do: Compute per call (O(1) CLDR table lookup).
- ❌ Don't: Add `cachedCalendars []string` + `sync.Once` to the `Locale` struct.

- **`String()` output is non-canonical** (not in dictionary order, not key-aware canonicalized): destroys round-trip.
- ✅ Do: `ca=islamicc` → `islamic-civil`, preserve the same spelling under unrelated keys, order `-u-` keys lexicographically, and retain present empty keyword types.
- ❌ Don't: Directly `fmt.Sprintf("%s-u-ca-%s-hc-%s", base, cal, hc)`.

---

## Acceptance Criteria

### Structure

- [ ] `locale.Locale` expresses `Intl.Locale` read-only object using unexported canonical identity + extension state, consistent with §1.1 getter.
- [ ] `Numeric()` getter returns `bool`; `Options.Numeric *bool` expresses omitted vs explicit boolean constructor input, while string option pointers express omitted vs explicit string constructor input.

### Analysis

- [ ] `Parse(s string) (Locale, error)` accepts BCP 47 strings (with `-u-` extension).
- [ ] `New(tag string, opts Options) (Locale, error)` accepts a `Options` value, `Options{}` represents an empty options object.
- [ ] `FromTag` and fallible `Tag` are the explicit x/text bridges.
- [ ] A successful `Tag()` round-trips through `FromTag` with complete identity; lossy conversion returns `ErrInvalidValue`.
- [ ] The only `golang.org/x/text` type allowed in an export signature is `language.Tag`; the secondary `x/text` type is not leaked.
- [ ] `ParseList(tags ...string)` constructs the formatter request list.
- [ ] If parsing fails, an error with `errors.Is(err, gointl.ErrInvalidValue)` being true is returned.
- [ ] Canonicalizes Unicode types by key from pinned CLDR BCP47 data: `ca=islamicc` → `islamic-civil`, `ms=imperial` → `uksystem`; `co=islamic-civil` remains unchanged and malformed `ca=gregorian` is rejected before lookup.
- [ ] `Parse("en-US-u-fw-0")` rejects malformed tag syntax, while `Options{FirstDayOfWeek: gointl.String("0")}` normalizes to `-u-fw-sun`; tag parsing never applies constructor-option coercion.
- [ ] Unicode extension keys have exactly two ASCII characters: alphanumeric followed by a letter. `00` and `a0` are invalid; `0a` and `aa` are valid, including uppercase input. Runtime and generator share `internal/localeid.IsUnicodeKey`.
- [ ] Tag types are grammar-checked; the §2.3 option rules apply only to explicit options.
- [ ] Explicit empty string option values (`Calendar`, `Collation`, `HourCycle`, `CaseFirst`, `NumberingSystem`, `FirstDayOfWeek`, and language identifier overrides) return invalid-option errors instead of being treated as omitted.

### String round trip

- [ ] `(Locale).String()` output canonical BCP 47:
- `-u-` extended keys are in lexicographic order (`ca` < `co` < `fw` < `hc` < `kf` < `kn` < `nu`).
- Absent keywords are omitted; present empty types are preserved; omitted `Numeric` is not output; explicit `Numeric=true` is output as `-u-kn`; explicit `Numeric=false` is output as `-u-kn-false`.
- Key-aware alias normalization (`Calendar="islamicc"` outputs `-u-ca-islamic-civil`).
- [ ] `Parse(loc.String()).Equal(loc) == true`(round-trip).
- [ ] `MarshalText` / `UnmarshalText` implements `encoding.TextMarshaler` / `TextUnmarshaler`.

### Maximize / Minimize

- [ ] `(Locale).Maximize() Locale` uses the pinned generated CLDR likely-subtag table and preserves caller-supplied subtags.
- [ ] `(Locale).Minimize() Locale` is idempotent, preserves maximal LSR, and uses the region-preferred candidate order.
- [ ] generated-reference `tests/likely-subtags.test.ts` and `tests/minimize.test.ts` All fixtures pass in `locale/canonical_test.go`.
- [ ] `Maximize` / `Minimize` and language/script/region constructor overrides preserve variants, transformed extensions, modeled and unmodeled Unicode extension content, and private use while replacing only language/script/region.

### Getter

- [ ] Simple fields (7 extended fields) are pre-parsed during construction in `Parse` / `New` (§5.1 table).
- [ ] Candidate list methods do not cache results into the struct. `GetCollations` projects the explicit locale collation as a singleton or returns the no-match fallback ["emoji", "eor"]; the other candidate-list methods read their owning generated data on each call. `GetTimeZones` uses explicit-region `internal/tz` records, including the full Canadian projection and `IN` → `Asia/Kolkata`.
- [ ] When the calendar keyword is present, `GetCalendars()` returns a single-element list, including an empty type.
- [ ] `WeekInfo` / `TextInfo` type signature is consistent with §5.2.

### Equality

- [ ] `(Locale).Equal(other) bool` compares the complete canonical string.
- [ ] `Locale` does not implement Go `==` comparable (explicitly prevented by a private zero-length function field).

### Options

- [ ] The `Options` field is consistent with §7.1, applied via `New(tag, options)`.
- [ ] Passing multiple `Options` values is impossible through the public Go signature.
- [ ] does not provide `With*` constructor options or `(Locale).With*` setters; the locale object remains read-only.

### Error

- [ ] `errors.Is(err, gointl.ErrInvalidValue)` returns true on all parsing/verification failures.
- [ ] `New` option failed to match `gointl.ErrInvalidOption`.
- [ ] `Parse` / `New` **not** panic (test covers various abnormal inputs).

### Test

- [ ] All cases of generated-reference `intl-locale/tests/index.test.ts` were ported to `locale/locale_test.go` and passed.
- [ ] Use `t.Parallel()` for all tests.
- [ ] At least 1 `Example*` function demonstrating `Parse` + `String()` round-trip.

---

## References

### Specification

- [ECMA-402 §14 — Intl.Locale Objects](https://tc39.es/ecma402/#locale-objects)
- [ECMA-402 §6.2.3 — CanonicalizeUnicodeLocaleId](https://tc39.es/ecma402/#sec-canonicalize-unicode-locale-id)
- [BCP 47 — Tags for Identifying Languages](https://www.rfc-editor.org/rfc/rfc5646)

### Reference implementations

- `.references/formatjs/packages/intl-locale/index.ts` —— `IntlLocaleOptions` / `RELEVANT_EXTENSION_KEYS` / `applyOptionsToTag` / `applyUnicodeExtensionToTag`
- `.references/formatjs/packages/intl-locale/preference-data.ts` —— region mapping calendar/hourCycle/firstDayOfWeek preference
- `.references/formatjs/packages/intl-locale/tests/index.test.ts` —— fixture
- `.references/ext/src/ecma402/locale.h` —— PHP `ecma402_locale` struct flat field struct precedent
- `.references/intl/intl.go` - translate-agent/intl is directly based on the Go precedent of `language.Tag`

### Cross-SPEC

- [SPEC 00 §4 — Locale Model](./00-vision-and-scope.md#4-locale-model)
- [SPEC 11 §ResolveLocale](./11-locale-matching.md) —— Consume `Locale.String()` with extended fields
- [SPEC 12 §Internal Slots](./12-abstract-operations.md#5-internal-slots) —— `[[Locale]]` slot expression at the abstraction layer
- [SPEC 50 §6 — Data Access API](./50-cldr-data.md#6-data-access-api) —— `MaximizeSubtags` / `CalendarPreference` / `WeekData` data
- [SPEC 60](./60-facade.md) —— root `GetCanonicalLocales` consumes canonical `Locale` values without adding locale availability matching


---

> This SPEC is a maintenance record for `locale.Locale`. A new ECMA-402 extension key (the spec rarely adds a new `-u-` key) triggers this SPEC revision; `x/text/language` behavior changes are exposed via fixture failures.

Calendar preference identifiers are canonicalized at generation, not repaired
by runtime parsing. `GetCalendars` retains region/world fallback, filters the
generated preference against supported calendars, and uses `gregory` when that
intersection is empty. Explicit `ca` presence bypasses that list selection.

Language identifier parsing and option replacement share `internal/localeid`.
Pinned CLDR language/script/territory/variant aliases canonicalize identity,
while Unicode and transformed extension aliases retain their key namespaces.
Subdivision aliases normalize rg/sd values. Unknown grammatical subtags survive;
no/nb remain distinct. Variants and t/u/x suffixes survive L/S/R replacement.

Hour-cycle info checks language-region time data before region time data for each
RegionPreference candidate (rg override, then base/sd/likely region). Explicit hc
remains a singleton. If no candidate has time data, the result is `["h23"]`, as
HourCyclesOfLocale requires; it does not borrow the world allowed list. Kernel
`HourCyclePreference(language, regions...)` owns this lookup and returns a copy.

Identity acceptance: `locale/language_identity_test.go`,
`locale/construct_test.go`, `internal/localeid/localeid_test.go`,
`internal/ecma402/displaynames_test.go`, and `intl_test.go` cover valid unknowns,
invalid duplicates/extension syntax, alias replacement, option overrides,
complete bridge failure/success, JSON, matching, and unavailable-data fallback.
Aliases come from `cldr-core/supplemental/aliases.json` and `cldr-bcp47/bcp47`;
`tools/gen-cldr/cldr/identity_alias.go` and `bcp47.go` load them, and
`codegen/unicode_alias.go` emits them into the shared private owner.
