# SPEC 30 — DateTimeFormat Core

> **Status:** Draft (2026-05-08)
> **Owner:** `datetimeformat/`
> **Reference contract:** `.references/ecma402/spec/datetimeformat.html` first, then `formatjs/packages/intl-datetimeformat/` + `formatjs/packages/ecma402-abstract/DateTimeFormat/`

## Overview

Defines `datetimeformat.DateTimeFormat` public API, constructor validation, ECMA-402 locale/options negotiation, time zone context handling for `time.Time`, `Format`/`FormatToParts`/`FormatRange`/`FormatRangeToParts`/`ResolvedOptions` behavior. Active generated pattern data is centered around Gregorian/ISO-8601; well-formed but unimplemented calendar requests participate in `ResolveLocale` and fall back to generated calendar data instead of becoming constructor errors.

This SPEC does not redefine:
- Skeleton parsing with BestFitFormatMatcher → see [SPEC 31 §Skeleton](./31-datetimeformat-skeleton.md#skeleton)
- Time zone data injection with metaZones → see [SPEC 32 §TZ Data](./32-datetimeformat-tz.md#tz-data)
- Calendar data schema → see [SPEC 32 §Calendar Data](./32-datetimeformat-tz.md#calendar-data)
- `Locale` structure → see [SPEC 10 §Locale structure](./10-locale.md)
- General entrance to abstract operations → See [SPEC 12 §Abstract Ops](./12-abstract-operations.md)

---

## 1. Public API

### 1.1 Construction and Option

```go
package datetimeformat

type DateTimeFormat struct{ /* Immutable; contains resolved + cached *time.Location + selectedPattern + cldr data slice */ }

type Options struct {
    Calendar               *string
    NumberingSystem        *string
    LocaleMatcher          *string
    FormatMatcher          *string
    TimeZone               *string
    TimeZoneName           *string
    Weekday                *string
    Era                    *string
    Year                   *string
    Month                  *string
    Day                    *string
    DayPeriod              *string
    Hour                   *string
    Minute                 *string
    Second                 *string
    HourCycle              *string
    Hour12                 *bool
    DateStyle              *string
    TimeStyle              *string
    FractionalSecondDigits *int
}

func New(locales locale.List, opts Options) (*DateTimeFormat, error)

func (f *DateTimeFormat) Format(t time.Time) string
func (f *DateTimeFormat) FormatToParts(t time.Time) []Part
func (f *DateTimeFormat) FormatRange(start, end time.Time) (string, error)
func (f *DateTimeFormat) FormatRangeToParts(start, end time.Time) ([]RangePart, error)
func (f *DateTimeFormat) ResolvedOptions() ResolvedOptions
```

**MUST** Rules:

1. `New` **MUST** complete all option syntax verification, locale/options negotiation, and `*time.Location` parsing during the construction period, and `error` will be returned on failure.
2. `New` accepts a `Options` value. `New(locales, Options{})` is equivalent to JS passing an empty options object or omitting options; multiple options objects are not Go API shapes and are rejected by the compiler.
3. Ordinary success paths for `Format` / `FormatToParts` / `FormatRange` / `FormatRangeToParts` **MUST NOT** return option errors. `FormatRange` / `FormatRangeToParts` return `ErrInvalidValue` for caller-fixable runtime range errors such as `start > end`.
4. `DateTimeFormat` is an immutable value; all methods on `*DateTimeFormat` must be concurrency safe.
5. Formatter options **MUST** adopt the typed `Options` value (same as SPEC 20), and functional options are prohibited from being used as a common main path.
6. `ResolvedOptions` **MUST** return an immutable snapshot (value type); the results of multiple calls are equal.

> **Why**: Centralized error handling during construction; `Format` does not return error and is byte-checked by conformance fixtures. The parsing of `*time.Location` (`time.LoadLocation`) cannot be redone in the `Format` phase (violating the hot path zero allocation rule).

### 1.2 Input type

**MUST** rules (corresponding to ECMA-402 §13.5.1 `time.Time` input parameter):

1. `Format`/`FormatToParts`/`FormatRange`/`FormatRangeToParts` input parameter **MUST** be `time.Time` (not `*time.Time`, not `int64` milliseconds); **disable** to accept `any` entry.
2. **MUST** call `t = t.Round(0)` immediately at the method entry to strip off the monotonic clock, and then go through the main formatting process.
3. `Location()` that comes with `time.Time` **cannot** be used as the display time zone for each call. ECMA-402 resolves `timeZone` to `[[TimeZone]]` during construction; when `Options.TimeZone` is not passed, the Go equivalent of `internal/tz.Default()` snapshot `SystemTimeZoneIdentifier()` must be passed. `internal/tz` is the only package that allows reading the host's default timezone, and tests override it via an internal provider.
4. The zero value of `time.Time{}` is instant that can be represented by Go, not JavaScript invalid Date. It must be formatted as normal `time.Time` via formatter `[[TimeZone]]`; rewriting it as Unix epoch or error is prohibited.
5. `FormatRange` / `FormatRangeToParts` normalize each instant independently before display-time-zone conversion. They do not compare or reorder endpoints: a later first argument is valid and remains the `startRange` source.

```go
// Entry normalization (signature, no implementation)
func (f *DateTimeFormat) Format(t time.Time) string {
t = t.Round(0) // Strip monotonic
loc := f.timeZone // Already cached on New; never from a single input Location()
    instant := t.UTC().UnixMilli()
// Use (instant, loc) to use PartitionDateTimePattern
}
```

> **Why**:
> 1. `time.Time` is Go idiom, interoperable with `time.Now()` / `time.Date()` / time package.
> 2. ECMA-402 model: input = instant, formatter `[[TimeZone]]` = output time zone; the input `Location` does not affect the output. Go idiom puts the time zone on `time.Time`, we must determine the formatter time zone during construction.
> 3. `t.Round(0)` is recommended by Go pkg.go.dev; otherwise `Format(t1) != Format(t1.Round(0))` will have observable differences (monotonic clock affects equality testing).
>
> **Rejected**:
> - **Accept `any` entry** (emulates JS `Date | number | string`) - Go type safety is weakened; and the messageformat-go bridge already holds `time.Time`.
> - **Do not strip monotonic** - Output jitter, `Format(t) == Format(t)` is true but `t.Equal(t2) ⇒ Format(t) == Format(t2)` is not true.
> - ** `t.Location()` that follows each input when TimeZone is not specified ** - This changes the formatter from an ECMA-402 fixed `[[TimeZone]]` object to a per-input display-zone helper and must be rejected.
> - **Use `t.Year()/Month()/Day()` directly and ignore `Options.TimeZone`** —— This departs from ECMA-402 and breaks fixture byte equality.

### 1.3 Calling example

```go
df, err := datetimeformat.New(locale.List{mustLocale("zh-CN")},
    datetimeformat.Options{
        DateStyle: gointl.String(string(datetimeformat.FullDateTimeStyle)),
        TimeZone:  gointl.String("Asia/Shanghai"),
    })
out := df.Format(time.Now())
// out == "Friday, May 8, 2026"
```

```go
df, _ := datetimeformat.New(locale.List{mustLocale("en-US")},
    datetimeformat.Options{
        Month: gointl.String(string(datetimeformat.ShortMonthStyle)),
        Day:   gointl.String(string(datetimeformat.NumericFieldStyle)),
        Year:  gointl.String(string(datetimeformat.NumericFieldStyle)),
    })
// df.Format(t) == "May 8, 2026"
```

### 1.4 Option parameter type — typed ECMA-402 values

**MUST** Rules:

1. String-valued ECMA-402 option bag entries (`localeMatcher` / `calendar` / `numberingSystem` / `timeZone` / `hourCycle` / `formatMatcher` / `weekday` / `year` / `month` / `day` / `dayPeriod` / `hour` / `minute` / `second` / `timeZoneName` / `dateStyle` / `timeStyle`) **MUST** be represented as `*string` in `Options`. `nil` means the option was omitted; `gointl.String("")` is present and invalid except for `timeZone`, where it is present and unsupported.
2. Formatter-owned named string types (`HourCycle`, `FormatMatcher`, `FieldStyle`, `NumericStyle`, `MonthStyle`, `TimeZoneName`, and `Style`) remain the legal-value vocabulary and the resolved-options vocabulary. Call sites pass them through `gointl.String`, and constructors copy the pointee into private config before validation.
3. Verification **MUST** be completed centrally in `New`. Failures wrap `ErrInvalidOption` or `ErrUnsupportedOption` and display the value passed in by the user.
4. `Hour12` is a special case (JS `boolean`): Go side **MUST** use `*bool` to distinguish "not passed" from "explicit false"; the call point uses `gointl.Bool(false)`.
5. `FractionalSecondDigits` **MUST** use `*int` to distinguish untransmitted and explicit illegal values; the call point uses `gointl.Int(n)`.
6. `TimeZone` **MUST** use `*string`: `nil` means the ECMA-402 `undefined` branch and selects the system default, while `gointl.String("")` is an explicit unsupported time-zone identifier and returns a constructor error.

> **Why**: typed value still retains ECMA-402 strings as wire/resolved forms, but lets Go call sites express legal values through constants. The conformance fixture and messageformat-go adapter do a mapping at the boundary and do not downgrade the public API to JSON form.
>
> **Rejected**: functional options + bare string - hidden state, not serializable, difficult to statically discover.

---

## 2. Options and ResolvedOptions

### 2.1 Options field

The `Options` (internal config struct) field corresponds to ECMA-402 §13.4.1 `InitializeDateTimeFormat`:

```go
// All named types underlying kind are string (JS literal alignment; see §1.4).
type (
    LocaleMatcher string  // "lookup" | "best fit"
    HourCycle     string  // "h11" | "h12" | "h23" | "h24"
    FormatMatcher string  // "basic" | "best fit"
    FieldStyle    string  // "narrow" | "short" | "long"
    NumericStyle  string  // "numeric" | "2-digit"
    MonthStyle    string  // "numeric" | "2-digit" | "narrow" | "short" | "long"
    TimeZoneName  string  // "short" | "long" | "shortOffset" | "longOffset" | "shortGeneric" | "longGeneric"
    Style         string  // "full" | "long" | "medium" | "short"
)

type Options struct {
    Calendar               *string // BCP 47 -u-ca-* literal; nil means omitted; gointl.String("") is invalid
    NumberingSystem        *string // nil means omitted; gointl.String("") is invalid
    LocaleMatcher          *string // nil means omitted/default best fit; gointl.String("") is invalid
    FormatMatcher          *string // nil means omitted/default best fit; gointl.String("") is invalid
    TimeZone               *string // IANA name or "+HH:MM"; nil selects the system default
    TimeZoneName           *string
    Weekday                *string
    Era                    *string
    Year                   *string
    Month                  *string
    Day                    *string
    DayPeriod              *string
    Hour                   *string
    Minute                 *string
    Second                 *string
    HourCycle              *string
    Hour12                 *bool // Distinguish between "not passed" and "explicit false"
    FractionalSecondDigits *int // 1..3
    DateStyle              *string
    TimeStyle              *string
}
```

`Options{}` is equivalent to the ECMA-402 empty options object. Pointer-backed fields carry option presence; the constructor copies pointee values into private config before validation, so caller mutation after construction cannot affect formatter behavior. Empty strings never mean "not passed"; omitted input is represented only by nil.

**MUST** Rules:

1. **MUST** check that `DateStyle` / `TimeStyle` are mutually exclusive with specific fields (`Weekday`/`Year`/`Month`/`...`); if set at the same time, return `ErrInvalidOption` (aligned with ECMA-402 §13.1.1.1 exception).
2. `Hour12` **MUST** be expressed through `*bool` to distinguish "not passed" from "explicit false".
3. `FractionalSecondDigits` **MUST** check ∈ `{1, 2, 3}`, otherwise return `ErrInvalidOption`.
4. `TimeZone` **MUST** distinguish omitted and explicit empty values. Omitted `nil` uses `internal/tz.Default()`; explicit empty `gointl.String("")` follows the same unsupported-option error path as any unavailable named time zone.
5. `NumberingSystem` and `Calendar` **MUST** only verify Unicode type syntax. Well-formed but unsupported values are selected by `ResolveLocale` when present in locale data, or fall back to default values when not present.
6. `FormatMatcher`, `TimeZoneName`, component fields, `HourCycle`, `DateStyle`, and `TimeStyle` **MUST** reject present empty strings through the same constructor option grammar as other ECMA-402 string options.

### 2.2 HourCycle linkage (§13.1.1.1) <a id="22-hourcycle-linkage-13111"></a>

**MUST** Rules:

1. The resolved hour-cycle value **MUST** be linked with `Locale.HourCycle()`(BCP 47 `-u-hc-...`) + explicit `Options.HourCycle` + `Options.Hour12`, according to ECMA-402 §13.1.1.1 step resolution:
   ```text
   if Hour12 != nil:
       resolved.HourCycle := Hour12 ? "h11"|"h12" (locale default 12-system location): "h23"|"h24"
       (The specific choice between the two is determined by dataLocale by default)
   else if HourCycle is set explicitly:
       resolved.HourCycle := HourCycle
   else if Locale.HourCycle() != "":
       resolved.HourCycle := Locale.HourCycle()
   else:
       resolved.HourCycle := dataLocale default (taken from CLDR `timeData.json` `preferred`)
   ```
2. **Disable** to let `Hour12 = false` overwrite `Locale.HourCycle() = h11` by default (must follow the above priority).
3. The simultaneous existence of `Options{HourCycle: gointl.String(string(H11HourCycle)), Hour12: gointl.Bool(false)}` MUST let `Hour12` take precedence over `HourCycle` (ECMA-402 §13.1.1.1).

> **Why**: HourCycle is a high-error field in the reference fixture corpus and must strictly follow §13.1.1.1.

### 2.3 ResolvedOptions

```go
type ResolvedOptions struct {
    Locale                 locale.Locale
    Calendar               string // generated-data-backed calendar identifier
    NumberingSystem        string
    TimeZone               string // IANA canonical name or "+HH:MM"
    HourCycle              *HourCycle
    Hour12                 *bool
    Weekday                *FieldStyle
    Era                    *FieldStyle
    Year                   *NumericStyle
    Month                  *MonthStyle
    Day                    *NumericStyle
    DayPeriod              *FieldStyle
    Hour                   *NumericStyle
    Minute                 *NumericStyle
    Second                 *NumericStyle
    FractionalSecondDigits *int
    TimeZoneName           *TimeZoneName
    DateStyle              *Style
    TimeStyle              *Style
}
```

**MUST** Rules:

1. The field order **MUST** be consistent with the ECMA-402 §13.4.5 spec order.
2. `TimeZone` **MUST** return the IANA canonical name (such as `"America/New_York"`), even if the input is link(`"US/Eastern"`); obtained through [SPEC 32 §CanonicalLink](./32-datetimeformat-tz.md#canonicallink).
3. `HourCycle` and `Hour12` **MUST** both be nil when no hour field is present, and both be non-nil when `hour` or `timeStyle` makes an hour field observable. `Hour12` is derived from the resolved hour cycle: `h11`/`h12` => `true`; `h23`/`h24` => `false`.
4. Branch-only resolved options (`HourCycle`, `Hour12`, component fields, `FractionalSecondDigits`, `TimeZoneName`, `DateStyle`, and `TimeStyle`) **MUST** use pointers; nil is the Go bridge for ECMA-402 property absence.
5. **MUST** return an immutable snapshot (value type) and clone pointer-backed scalar fields so caller mutation cannot affect later calls.
6. JSON field names and `omitempty` behavior **MUST** comply with [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy) and [SPEC 73 §Intl.DateTimeFormat](./73-json-records.md#intldatetimeformat). Nil pointer fields represent ECMA-402 property absence for branch-only resolved fields.

---

## 3. Calendar Resolution(active scope)

### 3.1 Support tier

Current tier: **narrowed implementation gap**.

| Field | Value |
|-------|-------|
| Current behavior | Generated Gregorian calendar data is active; `iso8601` is the ECMA-402 bridge over the same Gregorian local-time projection. |
| Rationale | ECMA-402 treats `calendar` as a Unicode extension negotiation input. The formatter stays truthful by reporting only the resolved calendar actually adopted by `ResolveLocale`, while `SupportedCalendars()` advertises only generated calendar data. |
| Native contract | Native deep fixtures must cover time-zone-name forms, metazone standard/daylight names, UTC-offset time zones including the `<24:00` boundary, interval ranges, range parts, resolved options, and every active DateTimeFormat divergence's `native_witness` before any DateTimeFormat expansion is accepted. |
| review_after | 2026-09-30 or the next CLDR / ICU calendar-data upgrade, whichever comes first. |
| Removal path | Generate non-Gregorian calendar payloads, implement calendar arithmetic equivalent to ECMA-402 `ToLocalTime`, add generated-reference and native fixtures, then expand `SupportedCalendars()`. |

This section defines the safe boundary: `ResolveLocale` may ignore unsupported calendar requests, and `SupportedCalendars()` remains the capability advertisement for calendars the formatter can actually report.

### 3.2 Current behavior

Active generated pattern data currently covers Gregorian/ISO-8601 observable behavior. Well-formed but unimplemented values in `Options.Calendar` or locale `-u-ca-*` must not return `gointl.ErrUnsupportedOption`; they must flow through `ResolveLocale`, fail to match locale calendar data, and resolve to the active default calendar.

> **Why**:
> 1. ECMA-402 constructor allows unsupported but well-formed calendar to fall back through locale data negotiation; go-intl follows that boundary and keeps the promise truthful by exposing the adopted calendar in `ResolvedOptions`.
> 2. The core difficulty of non-Gregorian pattern/calculation is the state machine of ICU `Calendar::computeFields` (~5000 lines of C++); the active generated payload does not contain equivalent data.
> 3. messageformat-go’s current `:datetime` function only supports Gregorian.
>
> **Excluded until implemented**:
> - **Buddhist** —— requires generated schema, year offset semantics, and conformance fixtures before it can be advertised.
> - **Persian** —— requires a Solar Hijri algorithm decision plus generated calendar data.
> - **Japanese** —— requires era table versioning and conformance around era transitions.

**MUST** Rules:

1. The `Calendar` field type **MUST** be `string` (exposing BCP 47 `-u-ca-...` form), and **not** be changed to a controlled enumeration (`type Calendar int`) - the string semantics directly corresponds to BCP 47, and is reserved for consumer-driven expansion.
2. When `Locale.Calendar()` / `Options.Calendar` is not empty, Unicode type syntax verification must be done first. Calendars outside the active locale data are ignored by `ResolveLocale` and must not be reported in `ResolvedOptions().Calendar`.
3. `SupportedCalendars()` **MUST** be derived from generated date calendar payload keys, mapping CLDR `"gregorian"` to ECMA-402 `"gregory"` and appending `"iso8601"` only when Gregorian data is present. It must be sorted, unique, and must not copy a broader host runtime calendar list.
3a. Manual conformance fixtures must cover fallback for well-formed unsupported explicit calendar options and unsupported `-u-ca-*` locale extensions. Error fixtures cover malformed calendar syntax.
4. `SupportedLocalesOf` **MUST** filter only by supported base locale and localeMatcher, preserving well-formed requested Unicode extensions such as `en-US-u-ca-buddhist`.
5. `ResolvedOptions().Calendar` **MUST** return the `ca` value of the resolved locale record; the current active data can only be `"gregory"` or `"iso8601"`.

### 3.3 Calendar data access

**MUST** Rules:

1. Calendar data (era / month / weekday / dayPeriod name) **MUST** be accessed through [SPEC 32 §Calendar Data](./32-datetimeformat-tz.md#calendar-data); `internal/cldr/date` owns the generated date payload and accessors.
2. The `New` period **MUST** first resolve the public `locale.Locale` into `date.Locale`, and then pull active Gregorian data once through `internal/cldr/date.GregorianFor(cldrLoc)` and freeze it into the `DateTimeFormat` internal slot.
3. Expanding beyond Gregorian / ISO-8601 **MUST** add the generated calendar payload, calendar-specific local-time projection, pattern / part behavior, and generated-reference or native fixtures in the same change before adding the calendar to `SupportedCalendars()`.

---

## 4. Time zone processing

### 4.1 TimeZone option analysis

**MUST** Rules:

1. `Options.TimeZone` input parameter **MUST** accept three forms:
   - IANA canonical name:`"America/New_York"`
- IANA link (backwards compatible):`"US/Eastern"` - **MUST** resolve to canonical via [SPEC 32 §CanonicalLink](./32-datetimeformat-tz.md#canonicallink)
- UTC offset string:`"+05:30"` / `"-08:00"` / `"+00:00"` -- **MUST** be parsed via [SPEC 32 §ParseOffsetString](./32-datetimeformat-tz.md#parseoffsetstring)
2. **MUST** call `time.LoadLocation` or `internal/tz.Resolve` when `New`, cache `*time.Location` to the internal slot; **It is forbidden** to parse the time zone name during the `Format` call.
3. Parsing failure **MUST** return `ErrInvalidOption` wrapped error, the message contains the timezone string.
4. `Options{TimeZone: nil}` has the same semantics as the unpassed option: the host default is used through the ECMA-402 `SystemTimeZoneIdentifier()` branch. `Options{TimeZone: gointl.String("")}` is an explicit empty identifier and **MUST** return an error matching `gointl.ErrUnsupportedOption`.
5. UTC offset form `*time.Location` **MUST** never be DST (direct fixed-offset zone), aligned with ECMA-402 offset time-zone semantics.
6. UTC offset strings must accept hours `00` through `23` with minutes `00` through `59`; `+23:59` and `-23:59` are valid, while `+24:00` and `-24:00` are unsupported. Negative zero offsets canonicalize to `+00:00`.
7. Manual conformance fixtures must keep invalid IANA zones and unsupported offset forms in the `ErrUnsupportedOption` path instead of falling back to the default time zone.

> **Why**: The `Format` path is redone every time `time.LoadLocation` is a significant performance loss (~10 μs file search each time, even `time/tzdata` has to take the parsing path). Time zones are materialized in `New`, allowing hot-path benchmark telemetry to reflect formatting rather than location loading.
>
> **Rejected**:
> - Let `Format(t)` directly use `t.Location()` to determine the display time zone - deviating from ECMA-402 fixed `[[TimeZone]]`.
> - `Format` path accepts `*time.Location` parameter - violates ECMA-402 single-parameter `format(date)` model.

### 4.2 Format No check in period TZ

**MUST** Rules:

1. `Format` / `FormatToParts` paths **must not** call `time.LoadLocation` / `internal/tz.Resolve`.
2. The `Format` path **must not** parse `metaZones.json` or do file I/O; the time zone display name (`shortGeneric` / `longGeneric`, etc.) is obtained through the generated `internal/cldr/timezone.TimeZoneDisplayName(cldrLoc, zone, form, isDST, instant, offsetMs)` table lookup and GMT fallback.

> **Why**: hot path should not be used for time zone parsing or runtime JSON; CLDR metazone/exemplar city data has been compiled into a Go table during codegen, and only memory table lookup and necessary GMT fallback are performed during the call period.

---

## 5. Formatting main process (PartitionDateTimePattern)

`Format` and `FormatToParts` share the main process `PartitionDateTimePattern`, defined in `formatjs/.../PartitionDateTimePattern.ts` + `ToLocalTime.ts` + `FormatDateTimePattern.ts`:

```text
1. instant := t.UTC().UnixMilli()
2. localTime := ToLocalTime(instant, calendar, location)
   // localTime = {era, year, month, day, weekday, hour, minute, second, ms, dst, offset}
3. program := f.pattern.program // Selected and compiled once in New
4. parts := []Part{}
5. for token in program:
       if token is literal: parts += {Type: "literal", Value: token.text}
       else: parts += FormatField(localTime, token, dataLocale, numberingSystem)
6. return parts
```

**MUST** Rules:

1. `Format` **MUST** be equivalent to `strings.Join(part.Value for part in FormatToParts(t), "")` - the output string bytes of the two are equal, as asserted by the conformance test.
2. `time.Time` input **MUST** be rounded and converted through the cached `*time.Location` before any calendar fields are read.
3. ECMA-402 `ToLocalTime` semantics **MUST** be centralized behind one local-time projection path. Active `gregory` and `iso8601` use Gregorian fields, including ECMA-402 BCE display-year conversion (`year <= 0` formats as `1 - year`); future calendars must extend that projection instead of scattering calendar conditionals through pattern code.
4. Pattern token → Field formatted lookup table **MUST** pass [SPEC 31 §Skeleton character table](./31-datetimeformat-skeleton.md).
5. The `Part.Type` output by `FormatField` **MUST** qualify ECMA-402 §15.5.1 Table 9 for a total of 15 spec strings: `era | year | month | day | hour | minute | second | weekday | dayPeriod | timeZoneName | literal | fractionalSecond | relatedYear | yearName | unknown`. The option, resolved property, and pattern field remain `fractionalSecondDigits`; only the emitted part type is `fractionalSecond`, exposed as `PartFractionalSecond`. The AM/PM mark is triggered by the token `a/b/B` inside the pattern, but the output part type is still `"dayPeriod"` (spec §15.5.4), and `"ampm"` must not be emitted directly because it is not an ECMA-402 part type. `relatedYear / yearName / unknown` will not be emitted in Gregorian-only scope, but the constants must exist so consumer switches can stay exhaustive. **It is forbidden** to emit option names or non-spec strings such as `fractionalSecondDigits`, `hour24`, `hour11`, `dayperiod`, or `ampm` as part types.
6. `DateTimeFormat` **MUST** cache `selectedPattern` and compile its endpoint, date, time, interval, and distinguishing fallback programs in `New`. `Format`, `FormatToParts`, `FormatRange`, and `FormatRangeToParts` execute those immutable programs; they must not tokenize patterns, repeat skeleton lookup, adjust fields, or construct fallback patterns on the hot path. The selected record owns effective date/time formats, date+time interpolation, and the compiled range record used by both range sinks.
7. `localeMatcher` and `formatMatcher` **MUST** be kept separate: `localeMatcher` only affects the CLDR data locale selection; `formatMatcher` only affects the component options → pattern selection. **BANNED** Substituting locale fallback results for formatMatcher decisions.
8. Component-style `ResolvedOptions` **MUST** be projected from the effective selected patterns after matcher adjustments and appended fields, not copied from the caller's requested option bag. Fields absent from the effective pattern remain nil; hour-cycle fields are absent when the selected pattern has no hour.

```go
type Part struct {
Type string // Strict enumeration, not open
    Value string
}
```

> **Why**: `Part.Type` is the alignment key of the conformance fixture; opening the enumeration (allowing arbitrary strings) will make the fixture impossible to mechanically compare. `selectedPattern` makes the pattern spine of constructor-eager explicit to prevent style, component, and fallback branches from spreading in the hot path.
>
> **Rejected**: `type PartType int` enumeration constant - one more mapping step when comparing with the string field of the JSON fixture, and the ECMA-402 string value becomes less readable.

---

## 6. FormatRange / FormatRangeToParts <a id="rangekind"></a>

**MUST** Rules:

1. `New` **MUST** compile the constructor-resolved interval and fallback data into one immutable `rangePatternRecord`. `FormatRange` and `FormatRangeToParts` only project localized endpoint fields through that record; they do not repeat skeleton lookup, field adjustment, or fallback pattern construction.
   ```text
   record := compile(selectedEffectivePattern, intervalFormats, intervalFallback)
   relation := record.compare(localStart, localEnd)
   if relation.equal: return format(start)
   if relation.pattern != nil: return apply(relation.pattern, start, end)
   return apply(intervalFallback, distinguishingEndpointPattern, start, end)
   ```
2. `intervalFormats` and `intervalFormatFallback` data **MUST** be obtained from the constructor-resolved `internal/cldr/date.GregorianFor(loc)` data (CLDR `ca-gregorian.json` `dateTimeFormats.intervalFormats`).
3. `FormatRangeToParts` **MUST** return the `[]Part` element with the `Source` field (`"startRange" | "endRange" | "shared"`); the `Part` type is expanded to:
   ```go
   type RangePart struct {
Type string // Same as Part.Type
       Value  string
       Source string  // "startRange" | "endRange" | "shared"
   }
   ```
4. `Source` string constant **MUST** reuse `internal/ecma402.RangeKind` (same as [SPEC 20 §FormatRange](./20-numberformat.md#5-formatrange--formatrangetoparts) `RangeKind`) to avoid spelling drift.
5. The input of `start.After(end)` must not return an error, nor must the two ends be swapped or `~` added; use the input parameter order according to ECMA-402 `PartitionDateTimeRangePattern`.
6. Equality is defined by the localized semantic fields present in the effective selected range record, not by instant equality and not by rendered label equality. Lower fields after the record's last present field are irrelevant; distinct narrow month labels remain distinct semantic months; fractional seconds compare at the resolved precision; repeated wall times in a DST fold may collapse. Equal selected records return `Format(start)` with shared parts.
7. Generated and manual range fixtures must cover date, time, date-time, reversed, semantic-equal, flexible/standard day-period, fractional-precision, and cross-date fallback behavior while retaining CLDR literal spacing instead of handwritten separators.
8. Accepted DateTimeFormat range, range-part-source, time-zone, or resolved-options divergences **MUST** carry a `native_witness` entry in `datetimeformat/testdata/divergences.md` that points to a same-package native fixture. If that native fixture also differs from generated CLDR output, retain it as its own accepted divergence instead of hiding it behind the generated-reference entry.
9. Interval tokenization, repeated-field occurrence counting, endpoint selection, and literal `source` classification **MUST** compile through one package-private execution-step path. `FormatRange` appends values and `FormatRangeToParts` materializes records from the same relation and pattern decisions; neither sink may maintain an independent start/end or source traversal.
10. The text path must not allocate a `[]RangePart` merely to share decisions. The cross-locale date/time/dateTime/fallback/reversed/equal matrix must prove `FormatRange(start, end) == concat(FormatRangeToParts(start, end).Value)` and legal range sources.
11. When a time-only or incomplete date-time pattern crosses a local date boundary, fallback endpoints **MUST** add enough omitted date fields to distinguish the endpoints. Larger differences retain missing smaller fields in day → month → year → era order. The fallback is constructor-compiled and shared by text and parts; callers do not configure it.

> **Why**: `intervalFormats` three-stage fallback is a high-risk conformance area; range output must be kept under fixture evidence rather than prose.
>
> **Rejected**: Abstract `CollapseRange[T]` is shared with NumberFormat - the Part field is different (see [SPEC 20 §FormatRange](./20-numberformat.md#5-formatrange--formatrangetoparts) for details).

---

## 7. Error model

**MUST** Rules:

1. The construction error must match root sentinel:
   - `gointl.ErrInvalidOption`
   - `gointl.ErrUnsupportedOption` for unsupported time-zone requests and other backend refusals; well-formed unsupported calendar requests are fallback inputs, not errors
2. Construction-time errors must have a `*gointl.Error` structured context, including field names, user values, locale and expected-value guidance.
3. **BANNED** `panic` any user path.
4. `time.Time` cannot represent JavaScript invalid Date; don’t invent invalid-date fallback for Go zero values or ordinary representable times.

```go
// Error example (signature)
err := ecma402.UnsupportedOptionErrorExpected("datetimeformat", "timeZone",
    tz, loc.String(), timeZoneExpected) // public callers match gointl.ErrUnsupportedOption
```

---

## 8. Performance telemetry

Benchmark numbers guide profiling and prioritization; they do not override ECMA-402 correctness or act as standalone merge blockers(SPEC 71).

**MUST** Rules:

1. `BenchmarkDateTimeFormat_New` and cached date/time benchmarks stay in `task bench` telemetry.
2. `Format` hot-path allocation counts are tracked with `b.ReportAllocs()`.
3. Performance work must not change skeleton matching, time-zone canonicalization, calendar support, or parts semantics.

> **Why**: The messageformat-go `:datetime` function may take N times `Format`;< 2 μs for each message to preserve the message layer SLA.

---

## 9. Forbidden

- **BANNED** Do not invent additional errors when the `Format` / `FormatToParts` / `FormatRange` path returns option error; Go typed input cannot express JavaScript invalid Date.
- **DOWN** Calling `time.LoadLocation` on the `Format` path - `*time.Location` must be cached when `New` is used.
- **BANNED** Check CLDR time zone display name in `Format` path - `metaZones` data must be materialized at `New` time.
- **NO** Generating calendar pattern data (including Buddhist year offsets) outside of the Gregorian pattern payload before the §3.1 removal path is complete; this is a current implementation boundary, not a permanent range shrinkage.
- **Banned** Follow the translate-agent mode (directly use `t.Location()` to determine the display time zone) - deviate from ECMA-402.
- **FORBIDDEN** Transfer instant via `t.Unix() * 1000` - accuracy loss (second level).
- **Disabled** Accepts `any` as input parameter or `*time.Time` - weakened type safety.
- **Disabled** `Format` does not strip the monotonic clock - `t.Round(0)` must be called immediately at method entry.
- **BANNED** Change `Part.Type` to `int` constant enum - one more mapping step when comparing against JSON conformance fixture string fields.
- **FORBIDDEN** Shares `CollapseRange` with NumberFormat - Part field is different.
- **BANNED** Introducing ICU C++ dependencies or cgo paths (SPEC 00 §1.1 non-target).

---

## 10. Acceptance Criteria

- [ ] `formatjs/packages/intl-datetimeformat/tests/format.test.ts` All fixtures in `datetimeformat/conformance_unified_test.go` pass (byte-equality).
- [ ] `formatjs/packages/intl-datetimeformat/tests/format-range.test.ts` All fixtures passed.
- [ ] `formatjs/packages/intl-datetimeformat/tests/offset-timezone.test.ts` All fixtures pass (`+05:30` / `-08:00` and other inputs).
- [ ] `go test -race ./datetimeformat/...` passed (including `TestDateTimeFormat_TimezoneContextPreservation`: the same `time.Time` has different output under different options `TimeZone`).
- [ ] `go test -race ./datetimeformat/...` passes (including `TestDateTimeFormat_MonotonicClockStripping`:`t.Round(0)` followed by multiple `Format(t)` bytes that are equal).
- [ ] `go test -race ./datetimeformat/...` passed (including `TestDateTimeFormat_ConcurrentFormat` 100 goroutine × 1000 calls).
- [ ] `datetimeformat/range_relation_test.go` proves selected-record semantic equality, flexible and standard day periods, resolved fractional precision, DST-fold local equality, and distinguishing cross-date fallbacks; joined range parts equal `FormatRange` bytes.
- [ ] `go vet ./datetimeformat/...` clean.
- [ ] `New(locale.List{loc}, Options{Calendar: gointl.String("buddhist")})` succeeds and `ResolvedOptions().Calendar == "gregory"` unless generated locale data adopts that calendar.
- [ ] `New(locale.List{loc}, Options{Calendar: gointl.String("")})` returns a wrapped error matching `gointl.ErrInvalidOption`.
- [ ] `New(locale.List{loc}, Options{NumberingSystem: gointl.String("")})` returns a wrapped error matching `gointl.ErrInvalidOption`.
- [ ] `New(locale.List{loc}, Options{TimeZone: gointl.String("Mars/Olympus")})` returns a wrapped error matching `gointl.ErrUnsupportedOption`.
- [ ] `New(locale.List{loc}, Options{TimeZone: gointl.String("")})` returns a wrapped error matching `gointl.ErrUnsupportedOption`.
- [ ] `New(locale.List{loc}, Options{TimeZone: gointl.String("US/Eastern")})` succeeded; `ResolvedOptions().TimeZone == "America/New_York"` (canonical name conversion).
- [ ] `New(locale.List{loc}, Options{TimeZone: gointl.String("+05:30")})` succeeded; `ResolvedOptions().TimeZone == "+05:30"` (offset string reserved).
- [ ] `New(locale.List{locEnUS}, Options{Hour12: gointl.Bool(false), Hour: gointl.String(string(NumericFieldStyle))})`: `ResolvedOptions().HourCycle != nil && *ResolvedOptions().HourCycle == "h23"` and `ResolvedOptions().Hour12 != nil && *ResolvedOptions().Hour12 == false`.
- [ ] `New(locale.List{locFrFR}, Options{Hour: gointl.String(string(NumericFieldStyle))})`: `ResolvedOptions().HourCycle != nil && *ResolvedOptions().HourCycle == "h23"` and `ResolvedOptions().Hour12 != nil && *ResolvedOptions().Hour12 == false` (French default h23).
- [ ] DateTimeFormat cached and constructor benchmarks appear in non-blocking `task bench` telemetry.
- [ ] Benchmark reports label DateTimeFormat as a per-surface package, not root facade cost.

---

## 11. References

### Primary

- `.references/formatjs/packages/intl-datetimeformat/core.ts` — public API, `tzData` injection path
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/InitializeDateTimeFormat.ts` — option pipeline
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/PartitionDateTimePattern.ts` — Main process
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts` — UTC → local time (including `+offset` and IANA, calendar=gregory invariant)
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/FormatDateTimePattern.ts` — Part.Type enumeration
- `.references/formatjs/packages/intl-datetimeformat/tests/` — main conformance source

- `.references/intl/intl.go` — translate-agent/intl(Go DateTimeFormat precedent, but ignore TimeZone, we don’t learn)
- `.references/ext/src/ecma402/{calendar,hour_cycle,time_zone}.{c,h}` — PHP/ICU identifier verification path

### Project Cross-References

- [SPEC 12 §Abstract Ops](./12-abstract-operations.md) — shared validators / pattern helpers / `ErrInvalidOption`
- [SPEC 10 §Locale structure](./10-locale.md) — `Locale.HourCycle()` / `Locale.Calendar()`
- [SPEC 31 §Skeleton](./31-datetimeformat-skeleton.md#skeleton) — Skeleton character table + BestFitFormatMatcher
- [SPEC 32 §TZ Data](./32-datetimeformat-tz.md#tz-data) — `time/tzdata` injection + metaZones
- [SPEC 32 §Calendar Data](./32-datetimeformat-tz.md#calendar-data) — Gregorian data schema
- [SPEC 50 §Schema](./50-cldr-data.md#schema) — `ca-gregorian.json` data shape
- [SPEC 60](./60-facade.md) — root namespace ownership; root `intl.FormatDate` one-shot helpers are outside the long-term public surface.
- [SPEC 71](./71-benchmark.md) — non-blocking performance telemetry
