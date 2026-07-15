# SPEC 10 — Locale

> **Status:** Revised (2026-05-20)
> **Priority:** High (all formatter input types; blocking SPEC 11 / 20 / 30 / 40 / 60)
> **Authority:** ECMA-402 `.references/ecma402/spec/locale.html` is a normative source. This SPEC records the current of `locale.Locale` type, ECMA-402 `Intl.Locale` alignment, parsing and normalization, `Maximize` / `Minimize`, read-only property getter, `String` / `Equal` / `MarshalText` / `UnmarshalText` Go contracts.

---

## Overview

`locale.Locale` is the input parameter type of all locale-aware operations in the go-intl public API. It is the Go representation of `Intl.Locale`, wrapping the BCP 47 parsing capabilities of [`golang.org/x/text/language.Tag`](https://pkg.go.dev/golang.org/x/text/language#Tag), overlaying ECMA-402/UTS #35 Unicode extension state through `internal/localeid`, and exposing properties required by the spec via read-only getters.

This SPEC defines the `Locale` structure, constructor (`New` / `FromTag` / `Parse`), locale-list helper (`ParseList`), normalization (`Maximize` / `Minimize`), getter materialization strategy, string round trip, comparability. **Do not** define best-fit matching algorithm (SPEC 11), CLDR data format (SPEC 50), formatter internal slot (SPEC 12 / 20 / 30 / 40).

---

## 1. Locale Structure <a id="locale-structure"></a>

### 1.1 Decision: Immutable `language.Tag` + extension state

```go
// locale/locale.go (signature)
package locale

// Locale is the Go representation of ECMA-402 Intl.Locale.
// The field is not exported; the caller must construct it through Parse/New and read it through getter.
type Locale struct {
    tag language.Tag
    ext extensions
}

func (l Locale) BaseName() string
func (l Locale) Tag() language.Tag
func (l Locale) Calendar() string
func (l Locale) Collation() string
func (l Locale) HourCycle() string
func (l Locale) CaseFirst() string
func (l Locale) Numeric() bool
func (l Locale) NumberingSystem() string
func (l Locale) Language() string
func (l Locale) Script() string
func (l Locale) Region() string
func (l Locale) Variants() []string
```

> **Why**:
> 1. **Reuse** - `language.Tag` has implemented BCP 47 parsing, normalization, `Base` / `Script` / `Region` / `Variants` access; the Go ecosystem (`x/text/message` / `x/text/currency` / `messageformat-go`) uniformly uses it as the underlying handle.
> 2. **Read-only model** - ECMA-402 `Intl.Locale.prototype` property is an accessor property and has no setter. The Go side must express the same immutable model with unexported fields + getters.
> 3. **Semi-transparent transmission** - `language.Tag` can read the string value of `-u-` extended key through `TypeForKey("ca")`, etc., **but does not expose** it as a typed field, does not perform `numeric` string ↔ bool conversion, and does not recognize the literal validity of `caseFirst="false"`. go-intl must hold extension state itself.
> 4. **Value type** - struct (not `*Locale`) maintains value semantics; the caller passes it by value, and `Locale` is an immutable snapshot.
>
> **Rejected**:
> - **`type Locale = language.Tag`(type alias)**: The extension field has nowhere to be placed, violating independent getter semantics such as spec `Intl.Locale.calendar`.
> - **Flat fields without `Tag`** (keep `language` / `script` / `region` by itself): Reinvent BCP 47 parsing, violating CLAUDE.md "no reinventing locale parsing".
> - **Exported field struct**: The caller can bypass the constructor and create an illegal state, violating the `Intl.Locale` read-only attribute model.
> - **`*Locale` pointer type**: Violates SPEC 00 §1 value type preference; `Equal` / `String` uses value semantics to be clearer.
> - **PHP style flat field struct** (put `language` / `script` / `region` / `calendar` / `hourCycle` all `string`): Give up the parsing ability of `language.Tag`.

### 1.2 Field type selection

| Field | Type | Default zero value meaning |
|------|------|-------------|
| `Calendar` | `string` | `""` = not specified, formatter takes region default |
| `Collation` | `string` | `""` = Not specified |
| `HourCycle` | `string` | `""` = not specified, formatter takes region default |
| `CaseFirst` | `string` | `""` = not specified; `"false"` literal is a legal value (spec mandatory) |
| `Numeric` | `bool` | `false` = spec default (numeric collation is not enabled) |
| `NumberingSystem` | `string` | `""` = not specified, formatter takes locale default |
| `FirstDayOfWeek` | `string` | `""` = not specified, formatter takes region default |

> **Why locale state uses `Numeric bool` not `*bool`**: What the reader needs is a resolved property value; the omitted vs explicit false of the constructor boundary is expressed by `Options.Numeric *bool` and does not leak to the `Locale` value object.
>
> **Rejected `Locale.Numeric *bool`**: Saving constructor presence state in a read-only value leaves the getter caller saddled with nil branches.

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
    Calendar        *string
    Collation       *string
    HourCycle       *string
    CaseFirst       *string
    Numeric         *bool
    NumberingSystem *string
    FirstDayOfWeek  *string
}
```

`language.Tag` is the only `golang.org/x/text` type allowed in the public API. `Locale.Tag()` and
`FromTag` together constitute the BCP 47 bridge of the Go ecosystem and must meet
`FromTag(l.Tag()).Tag() == l.Tag()`. Other `x/text` types, including
`display.Tags`, `collate.Collator`, `transform.Transformer`, `message.Printer`
etc., can only be left as implementation details within unexported fields, unexported functions or internal packages.
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

| Input | Behavior |
|------|------|
| `"en-US"` | OK;`Tag = en-US`, all extended fields are empty |
| `"en-US-u-hc-h23"` | OK;`HourCycle="h23"`,Tag reserved `-u-hc-h23` |
| `"en-US-u-ca-buddhist-hc-h23"` | OK; The field order is normalized according to spec (see §3.2) |
| `"en-US-u-ca-islamicc"` | OK; the deprecated CLDR value canonicalizes to `Calendar="islamic-civil"` |
| `"en-US-u-ca-islamic-civil"` | OK; the canonical value remains unchanged |
| `"en-US-u-ca-gregorian"` | Returns an error matching `gointl.ErrInvalidValue`; the nine-character subtag is not a well-formed Unicode type |
| `"en-US-u-ms-imperial"` | OK; key-aware CLDR canonicalization produces `-u-ms-uksystem` |
| `"xx-INVALID"` | Returns an error matching `gointl.ErrInvalidValue` |
| `""` | Returns an error matching `gointl.ErrInvalidValue` (empty string is illegal) |
| `"en-US-u-kn"`(no value table true) | OK;`Numeric=true` |
| `"en-US-u-kn-false"` | OK;`Numeric=false` |
| `"en-US-u-kf-false"` | OK;`CaseFirst="false"` (spec literal is legal) |

> **Why**:
> 1. **Unique boundary** - Parse is BCP 47 boundary, the public API no longer accepts raw `string` locale (SPEC 11 / 20 / 30 / 40 input parameters all `Locale`).
> 2. **Key-aware canonicalization** - ECMA-402 `CanonicalizeUValue` canonicalizes a value for its specific Unicode key. The pinned CLDR BCP47 data therefore owns mappings such as `ca=islamicc` → `islamic-civil` and `ms=imperial` → `uksystem`; the same spelling under an unrelated key is not rewritten. Syntax validation runs before alias lookup, so an alias cannot make malformed input valid.
> 3. **No public panic parser** - Test and example code can define local helpers, but production API exposes `Parse` / `ParseList` error returns only.
>
> **Rejected**:
> - **Receive `string` at formatter layer** (`numberformat.New("en-US", ...)`): Violation CLAUDE.md "Locale arguments are typed (`locale.Locale` or `language.Tag`) — never raw `string`. Parsing happens at the boundary, once."
> - **`Parse(s) Locale`(panic on error)**: spec RangeError must return error, violating CLAUDE.md "no panic in production code".

### 2.3 Construction phase verification

`Parse` / `New` **MUST** perform spec verification on the 7 extended fields:

| Field | Validation Rules |
|------|---------|
| `Language` / `Script` / `Region` | When the pointer is non-nil: the explicit string must be non-empty and form a valid BCP 47 language identifier with the retained subtags |
| `Calendar` | When the pointer is non-nil: the explicit string must be non-empty and match BCP 47 type sub-tag syntax (2–8 characters alphanumeric); **does not** verify whether it exists in CLDR (formatter layer is responsible) |
| `Collation` | When the pointer is non-nil: same as above |
| `HourCycle` | When the pointer is non-nil: required ∈ {`"h11"`, `"h12"`, `"h23"`, `"h24"`} |
| `CaseFirst` | When the pointer is non-nil: must ∈ {`"upper"`, `"lower"`, `"false"`} |
| `Numeric` | `nil` means omitted; non-nil `true` writes `-u-kn`; non-nil `false` writes `-u-kn-false` |
| `NumberingSystem` | When the pointer is non-nil: the explicit string must be non-empty and match BCP 47 type subtag syntax |
| `FirstDayOfWeek` | When the pointer is non-nil: the explicit string must be non-empty and ∈ {`"mon"`, `"tue"`, `"wed"`, `"thu"`, `"fri"`, `"sat"`, `"sun"`, `"0"`, `"1"`, `"2"`, `"3"`, `"4"`, `"5"`, `"6"`, `"7"`}; numbers normalize to ECMA-402 `WeekdayToUValue` |

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

1. **subtag order**:`language[-script][-region][-variants...]` (given by `language.Tag.String()`).
2. **`-u-` extended keys in dictionary order**: `ca` < `co` < `fw` < `hc` < `kf` < `kn` < `nu`.
3. **`-u-` extended value lowercase** (`Calendar="GREGORY"` output `-u-ca-gregory`).
4. **Empty internal extension fields are not output**; constructor option pointers with explicit `""` are rejected before `String()` canonicalization.
5. **Numeric=true outputs `-u-kn`** (no value table true, consistent with spec); explicit Numeric=false outputs `-u-kn-false`; omitted Numeric does not output.
6. **CaseFirst=`"false"` outputs `-u-kf-false`** (literal legal value).
7. **Key-aware Unicode type canonicalization** follows pinned CLDR BCP47 data after syntax validation (`ca=islamicc` → `islamic-civil`, `ms=imperial` → `uksystem`).
8. **Duplicate Unicode extension keys are first-wins**, matching ECMA-402 `UnicodeExtensionComponents` and native engine behavior (`en-u-ca-buddhist-ca-gregory` canonicalizes to `en-u-ca-buddhist`).
9. **Unicode extension insertion precedes private-use extension** (`en-u-ca-gregory-x-private`, not `en-x-private-u-ca-gregory`).

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
> **Rejected**: `MarshalJSON` / `UnmarshalJSON` both coexist (redundant, `encoding/json` automatically falls back to TextMarshaler).

### 3.4 JSON host-boundary records

`locale.Locale` JSON encoding is the canonical BCP 47 string via `MarshalText`. Locale info records use ECMA-402 field names:

```go
type WeekInfo struct {
    FirstDay time.Weekday   `json:"firstDay"`
    Weekend  []time.Weekday `json:"weekend"`
}

type TextInfo struct {
    Direction string `json:"direction"`
}
```

`WeekInfo.MarshalJSON` must emit ECMA-402 weekday numbers, Monday=1 through Sunday=7, instead of Go's `time.Weekday` values where Sunday=0. `TextInfo` marshals as `{"direction":"ltr"}` or `{"direction":"rtl"}`. Locale JSON field names follow [SPEC 73 §Part and Locale Info Records](./73-json-records.md#3-part-and-locale-info-records).

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
fmt.Println(loc.Maximize().Minimize().String())  // "zh-Hant"
```

### 4.2 Implementation strategy

`Maximize` / `Minimize` **MUST** use the generated CLDR `cldrlocale.MaximizeSubtags` / `cldrlocale.MinimizeSubtags` tables in `internal/cldr/locale` (see [SPEC 50 §6](./50-cldr-data.md#6-data-access-api)). They do **not** call `language.Tag.LikelyScript()` / `LikelyRegion()`, and there is no `internal/cldr/likely_subtags.go` patch layer.

Strategy:

1. `Maximize` splits the tag into `(language, script, region)` via `internal/localeid`, looks the triple up in `cldrlocale.MaximizeSubtags`, and rejoins the result. Unknown triples leave the tag unchanged. All seven Unicode extension fields are preserved.
2. `Minimize` is a deliberate **two-tier** lookup, not two competing algorithms:
   - Tier 1 is the precomputed CLDR `cldrlocale.MinimizeSubtags` table for known subtag triples.
   - Tier 2 is the general ECMA-402 `RemoveLikelySubtags` trial: it maximizes the input, then tries the `language`, `language+region`, and `language+script` candidates and keeps the first whose `Maximize` result equals the input's maximized form.
   Both tiers are driven by the same generated CLDR data (Tier 2 through `Maximize`), so they are consistent by construction and cannot drift; the two-tier design is documented in `locale/canonical.go`.
3. Conformance is verified against generated-reference `tests/likely-subtags.test.ts` fixtures.

> **Why**:
> 1. **Generated CLDR data** - The maximize/minimize tables are generated from pinned CLDR `likelySubtags.json`, keeping them aligned with the CLDR 48.1.0 conformance baseline rather than the possibly-lagging `x/text` built-in tables.
> 2. **One data source** - Driving Tier 2 through `Maximize` keeps both minimize tiers on the same generated table; collapsing them would drop the authoritative precomputed table for modest gain.
> 3. **Conformance takes priority** - SPEC 00 §2 requires byte-level fixture consistency, and the fixture is the truth table.
>
> **Rejected**:
> - **Depend on `language.Tag.LikelyScript()` / `LikelyRegion()`**: its data version is not the go-intl CLDR pin, so fixtures would force accepted divergence.
> - **Self-implemented likelySubtags algorithm**: Violates CLAUDE.md "no reinventing locale parsing".

### 4.3 Extended fields reserved

`Maximize` / `Minimize` **MUST** preserve all 7 extension fields (`Calendar` / `HourCycle` / ...); only modify the `Tag` part.

```go
loc := mustLocale("zh-u-hc-h23-ca-chinese")
m := loc.Maximize()
// m.HourCycle() == "h23" // Reserved
// m.Calendar() == "chinese" // Reserved
// m.Tag() == zh-Hans-CN // Inference
```

> **Why**: Consistent with generated-reference `intl-locale/index.ts` `maximize()` implementation; extended fields are explicitly selected by the user and Maximize should not silently discard them.

---

## 5. Getter Materialization (Close SPEC 00 §8 Q3)

### 5.1 Decision: Simple field pre-parsing, candidate list method is lazy

| Type | Field/Method | Materialization Timing |
|------|------------|---------|
| **Simple getter**(spec `Intl.Locale.prototype.<name>` getter) | `Calendar()` / `Collation()` / `HourCycle()` / `CaseFirst()` / `Numeric()` / `NumberingSystem()` / `FirstDayOfWeek()` | **Prepared on construction** (read directly from the BCP 47 `-u-` extended key; zero additional cost) |
| **Candidate list method** (spec `Intl.Locale.prototype.get<Name>s()` method) | `GetCalendars()` / `GetCollations()` / `GetHourCycles()` / `GetNumberingSystems()` / `GetTimeZones()` / `GetWeekInfo()` / `GetTextInfo()` | **Calculated on each call** from its owning generated data; not cached |

### 5.2 Candidate list method signature

```go
// locale/info.go(signature)

// GetCalendars returns a list of calendars preferred by this locale (sorted by priority).
// Corresponds to ECMA-402 sec-intl.locale.prototype.getCalendars.
func (l Locale) GetCalendars() []string

// GetCollations returns the collation list supported by active Collator; explicit Locale.Collation() only returns this value when it is not empty.
// ECMA-402 AvailableCanonicalCollations is implementation-defined; this method cannot select CLDR candidates
// The collation identifier is treated as an implemented Collator tailoring capability.
func (l Locale) GetCollations() []string

// GetHourCycles returns hour cycle preferences (by priority).
func (l Locale) GetHourCycles() []string

// GetNumberingSystems returns numbering system preferences.
func (l Locale) GetNumberingSystems() []string

// GetTimeZones returns the sorted primary IANA identifiers assigned to the
// explicit region subtag; it returns nil when the locale has no region.
func (l Locale) GetTimeZones() []string

// WeekInfo corresponds to ECMA-402 sec-week-info-of-locale.
type WeekInfo struct {
FirstDay time.Weekday // First day (default time.Monday)
Weekend []time.Weekday // Weekend (common [Saturday, Sunday])
}
func (l Locale) GetWeekInfo() WeekInfo

// TextInfo corresponds to ECMA-402 sec-text-info-of-locale.
type TextInfo struct {
    Direction string // "ltr" | "rtl"
}
func (l Locale) GetTextInfo() TextInfo
```

Call example:

```go
loc := mustLocale("ar-SA")
fmt.Println(loc.GetWeekInfo().FirstDay)    // time.Sunday
fmt.Println(loc.GetTextInfo().Direction)   // "rtl"
fmt.Println(loc.GetCalendars())            // ["islamic-umalqura", "gregory", "islamic", ...]
```

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

```go
// locale/info.go (fragment; non-complete implementation)
func (l Locale) GetCalendars() []string {
region := l.Maximize().Tag.Region().String() // Example: "SA"
    cldrLoc, _ := cldr.ResolveLocale(l.Tag)
    return cldrLoc.CalendarPreference()  // CLDR calendarPreferenceData
}
```

### 5.4 Explicit Calendar fields take precedence

If `Locale.Calendar()` is non-empty, `GetCalendars()` **MUST** return a single-element list (spec behavior).

```go
loc := mustLocale("en-US-u-ca-buddhist")
fmt.Println(loc.GetCalendars())  // ["buddhist"]
```

---

## 6. Equality & Comparability

### 6.1 Decision: Explicit `Equal` + `String` comparison

```go
// (Locale).Equal compares Tag.String() with 7 extension fields.
func (l Locale) Equal(other Locale) bool

// (Locale).String returns canonical BCP 47; the same Locale String() results are consistent multiple times.
func (l Locale) String() string
```

### 6.2 `==` Incomparable

`Locale` **not** guaranteed to be comparable to Go `==`; implementation uses unexported zero-length `func` fields to actively prevent comparisons to avoid callers
ECMA-402 locale object is used as a stable map key. `language.Tag` bridge can be compared according to its own Go type capabilities,
Therefore `FromTag(l.Tag()).Tag() == l.Tag()` is allowed and required to test lock; `Locale` itself must still pass
`Equal` or `String()` comparison.

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

Variants are not constructor options. They are part of the BCP 47 language tag:
`locale.Parse("sl-rozaj")` and `locale.New("sl-rozaj", locale.Options{})`
preserve them, and `Locale.Variants()` exposes them as a read-only view.

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

- **Rewrite BCP 47 parsing**: Violation of CLAUDE.md "no reinventing locale parsing".
- ✅ Do: Internally superimpose `-u-` on top of `language.Parse(s)` for extended processing.
- ❌ Don't: Write `parseBCP47(s string)` yourself.

- **Reimplement or fork the likelySubtags algorithm**: reinvent maximize/minimize instead of consuming generated CLDR data.
- ✅ Do: call the generated `cldrlocale.MaximizeSubtags` / `cldrlocale.MinimizeSubtags` tables (regenerated from pinned CLDR `likelySubtags.json`).
- ❌ Don't: depend on `language.Tag.LikelyScript()` / `LikelyRegion()` (data version is not the go-intl CLDR pin) or hand-maintain a `likely_subtags.go` patch layer.

- **Construction period panic**: Violation of CLAUDE.md "no panic in production code".
- ✅ Do: `Parse` / `New` returns `(Locale, error)`.
- ❌ Don't: `Parse` panics on invalid input or exposes a public `Must*` parser.

- **`Locale` is designed as a `*Locale` pointer type**: breaks value semantics.
- ✅ Do: `Locale` is a struct value type; re-call `locale.New(tag, Options{...})` or `locale.Parse(tag)` when the locale extension needs to be changed.
  - ❌ Don't: `func (l *Locale) WithCalendar(c string)`(in-place mutate).

- **`Locale` implements Go `==` comparable**: `language.Tag` is not comparable internally, nor is the embedded type after embedding.
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
- ✅ Do: `ca=islamicc` → `islamic-civil`, preserve the same spelling under unrelated keys, order `-u-` keys lexicographically, and omit empty fields.
- ❌ Don't: Directly `fmt.Sprintf("%s-u-ca-%s-hc-%s", base, cal, hc)`.

---

## Acceptance Criteria

### Structure

- [ ] `locale.Locale` expresses `Intl.Locale` read-only object using unexported `language.Tag` + extension state, consistent with §1.1 getter.
- [ ] `Numeric()` getter returns `bool`; `Options.Numeric *bool` expresses omitted vs explicit boolean constructor input, while string option pointers express omitted vs explicit string constructor input.

### Analysis

- [ ] `Parse(s string) (Locale, error)` accepts BCP 47 strings (with `-u-` extension).
- [ ] `New(tag string, opts Options) (Locale, error)` accepts a `Options` value, `Options{}` represents an empty options object.
- [ ] `FromTag(tag language.Tag, opts Options) (Locale, error)` is the only public `language.Tag` bridge.
- [ ] `FromTag(l.Tag()).Tag() == l.Tag()` holds for all legal `Locale`.
- [ ] The only `golang.org/x/text` type allowed in an export signature is `language.Tag`; the secondary `x/text` type is not leaked.
- [ ] `ParseList(tags ...string)` constructs the formatter request list.
- [ ] If parsing fails, an error with `errors.Is(err, gointl.ErrInvalidValue)` being true is returned.
- [ ] Canonicalizes Unicode types by key from pinned CLDR BCP47 data: `ca=islamicc` → `islamic-civil`, `ms=imperial` → `uksystem`; `co=islamic-civil` remains unchanged and malformed `ca=gregorian` is rejected before lookup.
- [ ] The 7 extended fields are spec checked during construction (§2.3 table).
- [ ] Explicit empty string option values (`Calendar`, `Collation`, `HourCycle`, `CaseFirst`, `NumberingSystem`, `FirstDayOfWeek`, and language identifier overrides) return invalid-option errors instead of being treated as omitted.

### String round trip

- [ ] `(Locale).String()` output canonical BCP 47:
- `-u-` extended keys are in lexicographic order (`ca` < `co` < `fw` < `hc` < `kf` < `kn` < `nu`).
- Empty internal fields are omitted; omitted `Numeric` is not output; explicit `Numeric=true` is output as `-u-kn`; explicit `Numeric=false` is output as `-u-kn-false`.
- Key-aware alias normalization (`Calendar="islamicc"` outputs `-u-ca-islamic-civil`).
- [ ] `Parse(loc.String()).Equal(loc) == true`(round-trip).
- [ ] `MarshalText` / `UnmarshalText` implements `encoding.TextMarshaler` / `TextUnmarshaler`.

### Maximize / Minimize

- [ ] `(Locale).Maximize() Locale` uses the pinned generated CLDR likely-subtag table and preserves caller-supplied subtags.
- [ ] `(Locale).Minimize() Locale` is the inverse of Maximize.
- [ ] generated-reference `tests/likely-subtags.test.ts` and `tests/minimize.test.ts` All fixtures pass in `locale/canonical_test.go`.
- [ ] `Maximize` / `Minimize` 7 extension fields reserved.

### Getter

- [ ] Simple fields (7 extended fields) are pre-parsed during construction in `Parse` / `New` (§5.1 table).
- [ ] Candidate list methods `GetCalendars` / `GetCollations` / `GetHourCycles` / `GetNumberingSystems` / `GetTimeZones` / `GetWeekInfo` / `GetTextInfo` read their owning generated data on each call and do not cache it into the struct. `GetTimeZones` uses explicit-region `internal/tz` records, including the full Canadian projection and `IN` → `Asia/Kolkata`.
- [ ] When explicit `Calendar` is not empty, `GetCalendars()` returns a single-element list.
- [ ] `WeekInfo` / `TextInfo` type signature is consistent with §5.2.

### Equality

- [ ] `(Locale).Equal(other) bool` semantic comparison by field (`Tag.String()` + 7 extended fields).
- [ ] `Locale` does not implement Go `==` comparable (the embedded `language.Tag` is not comparable).

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
