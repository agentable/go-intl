# SPEC 32 — DateTimeFormat TimeZone & Calendar Data

> **Status:** Draft (2026-05-08)
> **Owner:** `internal/cldr/timezone/` + `internal/cldr/date/` + `internal/cldr/locale/timezones.go` + `internal/tz/`
> **Reference contract:** `formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` + `time/tzdata` + CLDR `cldr-core/supplemental/metaZones.json` + `cldr-dates-full/main/<locale>/timeZoneNames.json` + `cldr-dates-full/main/<locale>/ca-gregorian.json`

## Overview

Define the two types of underlying data required by DateTimeFormat:

1. **TimeZone data**: IANA tzdata injection strategy, `*time.Location` parsing, UTC offset string parsing, CLDR `metaZones` → display name mapping.
2. **Calendar data** (active scope only Gregorian): era / month / weekday / dayPeriod name, formats, intervalFormats, availableFormats schema and access.

This SPEC does not redefine:
- `DateTimeFormat.New` call flow → see [SPEC 30 §Public API](./30-datetimeformat.md)
- Skeleton character table and BestFitFormatMatcher → see [SPEC 31 §Skeleton character table](./31-datetimeformat-skeleton.md)
- CLDR full data generation pipeline → see [SPEC 50 §Codegen](./50-cldr-data.md#codegen)
- Version pinned global strategy → See [SPEC 50 §Version Pin](./50-cldr-data.md#version-pin)

---

## 1. TimeZone data strategy <a id="tz-data"></a>

### 1.1 Data source decision

**MUST** Rules:

1. IANA tzdata **MUST** be injected via the Go official `_ "time/tzdata"` blank import.
2. **It is forbidden** to copy the tzif file or the transition data of `cldr-core/supplemental/timezone.json` to the warehouse.
3. **FORBIDDEN** host-dependent `/usr/share/zoneinfo` (`time.LoadLocation` default behavior): deployment to Alpine/scratch containers will fail.
4. **FORBIDDEN** from transplanting FormatJS's `tz_data.tar.gz` pipeline (zdump + Docker + base36 encoding) - Go already has `time/tzdata`, and there is no profit in reinventing the wheel.

> **Why**:
> 1. `time/tzdata` is maintained by the Go team and has the same origin as IANA tzdata. ~450 KB increments added to the binary are acceptable (Go binary is usually 10–50 MB).
> 2. ECMA-402 requires three types of input to coexist: `"Etc/UTC"` / `"America/New_York"` / `"+05:30"`; the first two types go to `time.LoadLocation`, and the third type parses by itself.
> 3. SPEC 00 §5.4 has stated that "we want deterministic output" - the system zoneinfo is not deterministic in different deployment environments.
>
> **Rejected**:
> - **Dependent system zoneinfo**: Alpine/scratch container has no files, prod/dev failure mode is inconsistent.
> - **Transplanting FormatJS tz_data**: CI is highly complex (Docker is required to run zic), and the packed binary format is difficult to mechanically translate on the Go side.
> - **Automatic zone-transition table codegen**:`time/tzdata` is already the minimum trusted source and has no profit.

### 1.2 Injection location

**MUST** Rules:

1. Blank import **MUST** be located in `internal/tz/tzdata.go` (single file):
   ```go
   package tz

   import _ "time/tzdata"
   ```
2. **Disable** letting the top-level `intl` package or formatter package do this import - focus on `internal/tz/` for auditing purposes.
3. `internal/tz/` **MUST** be used as the only time zone related entry; the `datetimeformat/` package is called through the `internal/tz` accessor.

### 1.3 *time.Location analysis

**MUST** Rules:

1. `internal/tz.Resolve(name string) (*time.Location, error)` **MUST** be implemented:
- Supports IANA canonical name(`"America/New_York"`).
- Supports IANA link(`"US/Eastern"` → `"America/New_York"`).
- Supports UTC offset string(`"+05:30"` / `"-08:00"` / `"+00:00"` / `"+14:00"`).
2. Formatter boundary failure **MUST** return a wrapped error matching `gointl.ErrUnsupportedOption`, and the message contains the input string.
3. UTC offset string parsing **MUST** be implemented in `internal/tz.ParseOffsetString(s string) (int64, error)`, returning milliseconds east of UTC; **FORBIDDEN** to pass regular fallback to `time.FixedZone` (must explicitly check the range `[-14:00, +14:00]`, consistent with ECMA-402 §13.1.3). <a id="parseoffsetstring"></a>

```go
// Interface form (signature, no implementation)
package tz

func Resolve(name string) (*time.Location, error)
func Default() (string, *time.Location)
func ParseOffsetString(s string) (int64, error)        // ms east of UTC
func CanonicalLink(name string) string                 // "US/Eastern" → "America/New_York"
func LookupAt(loc *time.Location, t time.Time) ZoneInfo

type ZoneInfo struct {
    Name     string  // canonical IANA name
    OffsetMs int64
    IsDST    bool
    Abbrv    string  // e.g. "EST" / "EDT"
    Metazone string  // e.g. "America_Eastern", from CLDR metazoneInfo
}
```

`Default` is the single internal owner for the DateTimeFormat default time-zone provider. It canonicalizes IANA links and offset names the same way as explicit `Options.TimeZone`; tests may replace the provider with `OverrideDefaultForTest`, but no public diagnostic or configuration API is exposed.

### 1.4 CanonicalLink table <a id="canonicallink"></a>

**MUST** Rules:

1. The IANA link → canonical mapping **MUST** be extracted from the `cldr-core/supplemental/metaZones.json` `metazoneInfo.timezone` node + IANA `backward` file through codegen to generate Go literal.
2. **Disabled** runtime parsing of IANA `backward` files; **Disabled** `//go:embed backward` string file + parsing.
3. The canonical-link table storage path **MUST** be `internal/cldr/locale/timezones.go` (generated locale-kernel data).
4. Form:
   ```go
   var canonicalTimeZoneLinks = map[string]string{
       "US/Eastern":    "America/New_York",
       "US/Pacific":    "America/Los_Angeles",
       "Asia/Calcutta": "Asia/Kolkata",
       // ...
   }
   ```
5. `CanonicalLink(name)` **MUST** look up the table first; if it cannot be found, it will be returned as is (assuming that the input is canonical).

### 1.5 metaZones three-segment mapping <a id="metazones"></a>

The localized display name of DateTimeFormat `timeZoneName: short | long | shortGeneric | longGeneric` is CLDR `metaZones` three-segment mapping:

```text
zone (e.g. "America/New_York")
  → metazoneInfo (zone → metazone, time-bounded)
    → metaZoneNames (locale × metazone → display name strings)
      → exemplarCity (locale × zone → city name fallback)
```

**MUST** Rules:

1. The three-segment mapping **MUST** be output by codegen into the `internal/cldr/timezone` domain package.
2. **Disable** runtime JSON parsing (`//go:embed metaZones.json` + `encoding/json`); **Disable** deserialization at startup.
3. Schema of three-segment table:
   ```go
// generated_metazones.go (fragment, signature)

// Phase 1:zone → []metazone period (metazone can change over time)
   type MetazonePeriod struct {
       Metazone string
Start int64 // unix milliseconds, MinInt64 means -∞
End int64 //, MaxInt64 means +∞
   }
   var zoneToMetazones = map[string][]MetazonePeriod{
       "America/New_York": {{"America_Eastern", math.MinInt64, math.MaxInt64}},
       // ...
   }

// Stage 2: locale × metazone × form → display name
   type MetazoneNames struct {
       LongGeneric, LongStandard, LongDaylight string
       ShortGeneric, ShortStandard, ShortDaylight string
   }
   var metazoneNamesByLocale = map[Locale]map[string]MetazoneNames{
       localeIndex["en-US"]: {
           "America_Eastern": {
               LongGeneric: "Eastern Time", LongStandard: "Eastern Standard Time", LongDaylight: "Eastern Daylight Time",
               ShortGeneric: "ET", ShortStandard: "EST", ShortDaylight: "EDT",
           },
       },
       // ...
   }

// Stage 3: locale × zone → exemplar city
   var exemplarCitiesByLocale = map[Locale]map[string]string{
       localeIndex["en-US"]: {"America/New_York": "New York"},
       // ...
   }
   ```
4. accessor `internal/cldr/timezone.TimeZoneDisplayName(loc timezone.Locale, zone string, form TimeZoneName, isDST bool, instant int64, offsetMs int64) string` **MUST** implement ECMA-402 §13.1.5 fallback chain:
   ```text
1. Check metazoneNames[locale][metazone(zone, instant)][form] → Return if hit
2. Check exemplarCities[locale][zone] → Return generic fallback("Time in {city}") if hit
3. Fall back to GMT offset string ("GMT-05:00")
   ```
5. The metazone selection **MUST** support time sensitivity (`instant` decides which `MetazonePeriod` to take) - Example: `Europe/Moscow` has different metazones around 2011.

> **Why**: `metaZones` three-segment mapping is a difficulty in conformance testing (FormatJS `tests/timezone-name.test.ts` 30+ fixture); any runtime analysis will introduce cold start delay and binary size overhead.
>
> **Rejected**: Deserializing JSON on startup - conflicts with SPEC 50 §"no runtime file I/O for CLDR".

### 1.6 tzdata version pinned <a id="tzdata-version"></a>

**MUST** Rules:

1. The tzdata version number **MUST** be written into the `internal/cldr/VERSION` single file, as a `tzdata=2025b` line (the same file as `cldr=` / `icu=`).
2. CI **MUST** verify that the embedded IANA version of `time/tzdata` is consistent with the `VERSION` file; inconsistency is a block.
3. The tzdata version and the CLDR/ICU version **MUST** be bumped at the same time; bumping either one independently is a PR block.
4. Inconsistencies between tzdata and metaZones **MUST** be authoritative (behavioral correctness > display name consistency); disagreements are logged to `divergences.md`.

> **Why**: tzdata and CLDR are released independently (IANA multiple times a year, CLDR twice a year); version inconsistency will cause `LookupAt(loc, t)` and metaZones data to have a corner case of "the moment should be in metazone X, but tzdata has changed". Pin the same file + CI check to avoid drift.
>
> **Rejected**: Let tzdata and CLDR bump independently - the cost of time zone data drift is much higher than the cost of "unified bump once".

---

## 2. Calendar data <a id="calendar-data"></a>

### 2.1 Active scope

**MUST** Rules:

1. Active scope calendar pattern data currently only generates Gregorian. `SupportedCalendars()` **MUST** derive values from generated calendar payload keys, expose CLDR `"gregorian"` as ECMA-402 `"gregory"`, and add ECMA-402 required `"iso8601"` only while Gregorian data is present; other well-formed calendar requests return errors matching `gointl.ErrUnsupportedOption` during construction and must not silently fall back to Gregorian behavior.
2. The codegen tool (`tools/gen-cldr/`) **MUST** skip non-Gregorian pattern payload until the formatter has a real local-time projection, pattern / part behavior, and fixtures for that calendar. If non-Gregorian payload is generated without a consumer, build or tests must fail.
3. Calendar data storage path **MUST** be the `internal/cldr/date` domain package, with generated const blobs in `internal/cldr/date/data.go`.
4. The consumer-driven expansion calendar pattern can only be added when there is consumer demand; when adding, the corresponding generated schema and test will be introduced, and the `//go:build future` placeholder file shall not be reserved in the active scope.

> **Why**:
> 1. SPEC 30 §3 has declared that calendar support is advertised only when generated data and local-time semantics are both real.
> 2. The active scope placeholder file has no runtime consumer, which will expand the audit area and create unused code.
> 3. The future calendar directory and schema should be deduced from the actual implementation and should not be frozen at the current stage.

### 2.2 Gregorian Schema

Gregorian data schema in `internal/cldr/date` must be aligned to CLDR `cldr-dates-full/main/<locale>/ca-gregorian.json`:

```go
// Signature, no implementation
package date

type Gregorian struct {
// Source: dates.calendars.gregorian.eras.eraNames / eraAbbr / eraNarrow
    Eras struct {
        Wide   [2]string // ["Before Christ", "Anno Domini"]
        Abbr   [2]string // ["BC", "AD"]
        Narrow [2]string // ["B", "A"]
    }

// Source: dates.calendars.gregorian.months.format.{wide,abbreviated,narrow}
    Months struct {
        Wide       [12]string // ["January", "February", ...]
        Abbr       [12]string // ["Jan", "Feb", ...]
        Narrow     [12]string // ["J", "F", ...]
        StandWide  [12]string // stand-alone form
        StandAbbr  [12]string
        StandNarrow [12]string
    }

// Source: dates.calendars.gregorian.days.format.{wide,abbreviated,narrow,short}
    Weekdays struct {
        Wide        [7]string // ["Sunday", ..., "Saturday"]
        Abbr        [7]string // ["Sun", ..., "Sat"]
        Narrow      [7]string // ["S", "M", ...]
        Short       [7]string // ["Su", "Mo", ...]
        StandWide   [7]string
        StandAbbr   [7]string
        StandNarrow [7]string
        StandShort  [7]string
    }

// Source: dates.calendars.gregorian.dayPeriods.format.{wide,abbreviated,narrow}
    DayPeriods struct {
        AM struct{ Wide, Abbr, Narrow string }
        PM struct{ Wide, Abbr, Narrow string }
// flexible day periods (b / B characters, §2.5)
        Flex map[string]struct{ Wide, Abbr, Narrow string }
// For example keys: "morning1", "afternoon1", "evening1", "night1", "noon", "midnight"
    }

// Source: dates.calendars.gregorian.dateFormats / timeFormats / dateTimeFormats
    DateFormats     [4]string // [full, long, medium, short]
    TimeFormats     [4]string
DateTimeFormats [4]string // Combination pattern, such as "{1} {0}" or "{1}, {0}"

// Source: dates.calendars.gregorian.dateTimeFormats.availableFormats
    AvailableFormats map[string]string // skeleton → pattern

// Source: dates.calendars.gregorian.dateTimeFormats.intervalFormats
    IntervalFormats map[string]map[string]string // skeleton → field → pattern
    IntervalFallback string                      // intervalFormatFallback

// Source: dates.calendars.gregorian.dateTimeFormats.appendItems
    AppendItems map[string]string // field → "{0} ({1})" pattern
}
```

**MUST** Rules:

1. All fields **MUST** be extracted and generated by codegen from CLDR JSON; **handwriting is prohibited**.
2. The data form **MUST** be Go literal (`var gregorian_en_US = Gregorian{...}`); **FORBIDDEN** `//go:embed *.json` + runtime deserialization.
3. The accessor `internal/cldr/date.GregorianFor(loc date.Locale) Gregorian` **MUST** receive the parsed date-domain locale handle; the matching from public `locale.Locale` to `date.Locale` is completed by locale negotiation + `internal/cldr/date.ResolveLocale(tag string)` during the formatter construction period.
4. `dayPeriods.format.flex` **MUST** be generated in full (`morning1` / `afternoon1` / `evening1` / `night1` / `noon` / `midnight`); **FORBIDDEN** only generates `noon` / `midnight` (FormatJS current limitation).

> **Why**:
> 1. Go literals = compile-time verification + no runtime JSON/file I/O; large tables are initialized through on-demand loaders to avoid paying for unused formatter data in lightweight calling paths.
> 2. Flex day periods are a necessary condition for the correct output of `b` / `B` characters (`zh-CN` early morning/morning/morning/noon/afternoon/night), and fallback to AM/PM means byte equality is destroyed.

### 2.3 dayPeriodRules <a id="dayperiodrules"></a>

flex day periods boundary data comes from CLDR `cldr-core/supplemental/dayPeriods.json` `dayPeriodRules` nodes (different for each locale):

**MUST** Rules:

1. `dayPeriodRules` **MUST** be fully embedded into the `internal/cldr/date` generated payload:
   ```go
// internal/cldr/date/data.go (conceptual fragment)
   type DayPeriodRange struct {
From, To time.Duration // Time offset from 00:00 on the current day
       Type     string         // "morning1" / "afternoon1" / ...
   }
   var dayPeriodRules = map[string][]DayPeriodRange{
       "en": {
           {0 * time.Hour, 5 * time.Hour, "night1"},
           {5 * time.Hour, 12 * time.Hour, "morning1"},
           {12 * time.Hour, 12 * time.Hour, "noon"},     // exact 12:00
           {12 * time.Hour, 18 * time.Hour, "afternoon1"},
           {18 * time.Hour, 21 * time.Hour, "evening1"},
           {21 * time.Hour, 24 * time.Hour, "night1"},
       },
       "zh": { /* ... */ },
   }
   ```
2. `internal/cldr/date.DayPeriodFor(loc date.Locale, hour, minute int) string` **MUST** return `"morning1"` / `"noon"` and other type names; the caller then checks `Gregorian.DayPeriods.Flex[type]`.
3. **Disabled** only hardcoded `en` boundaries; **Required** full codegen.

---

## 3. ResolvedOptions.TimeZone behavior

**MUST** rules (corresponding to [SPEC 30 §ResolvedOptions](./30-datetimeformat.md#23-resolvedoptions)):

1. The `ResolvedOptions().TimeZone` return value **MUST** be the IANA canonical name (converted for link) or UTC offset string (reserved for `+HH:MM` input).
2. The conversion process **MUST** be:
   ```text
1. resolveTZ := internal/tz.Resolve(input) // Accept link / canonical / offset
   2. canonical := internal/tz.CanonicalLink(input)
3. f.timeZoneCanonical := canonical // Cache to DateTimeFormat slot
   4. ResolvedOptions().TimeZone == canonical
   ```
3. UTC offset string input **MUST** retain the original sign and padding (`"+05:30"` does not become `"05:30"`).

---

## 4. Performance

**MUST** Rules:

1. `BenchmarkTZ_Resolve` remains trend telemetry for `time.LoadLocation` and canonicalization cost.
2. `internal/cldr/date.GregorianFor(loc)` **MUST** be based on the parsed date-domain locale handle, and the BCP 47 string must not be parsed during the call period.
3. `BenchmarkCLDR_TimeZoneDisplayName` remains trend telemetry for generated metazone lookup cost.

---

## 5. Forbidden

- **FORBIDDEN** Copying tzif files to the repository - must be injected via `_ "time/tzdata"`.
- **BANNED** Depends on system `/usr/share/zoneinfo` (`time.LoadLocation` default behavior, Alpine container has no files).
- **NO** Porting the FormatJS `tz_data.tar.gz` pipeline - Go already has `time/tzdata`.
- **Disabled** Runtime JSON parsing `metaZones.json` - Must codegen output Go literal.
- **BANNED** `//go:embed metaZones.json` + `encoding/json` paths - Conflicts with SPEC 50 "no runtime file I/O".
- **FORBIDDEN** advertising or generating active non-Gregorian calendar data, including Buddhist placeholders, without formatter local-time projection, pattern / part behavior, and conformance fixtures in the same change.
- **Disabled** The tzdata version is bumped independently from the CLDR / ICU version - must be the same as the `internal/cldr/VERSION` file, CI verification.
- **BANNED** `dayPeriodRules` only generates `en` - must fully codegen all active scope locales.
- **BANNED** `dayPeriodRules` only covers `noon` / `midnight` - must be full flex(morning1/afternoon1/evening1/night1).
- **BANNED** `Format` paths calling `internal/tz.Resolve` or `time.LoadLocation` -- must cache `*time.Location` when `New` does (SPEC 30 §4.2).
- **Disabled** UTC offset input skips range checking - must be `[-14:00, +14:00]`.
- **BANNED** The IANA link table (`canonicalTimeZoneLinks`) is generated by runtime parsing of the `backward` file - codegen is required.

---

## 6. Acceptance Criteria

- [ ] `internal/tz/tzdata.go` contains only one line `import _ "time/tzdata"`; no other declarations.
- [ ] `internal/tz.Resolve("America/New_York")` returns non-empty `*time.Location`, offset including DST (-04:00 in summer, -05:00 in winter).
- [ ] `internal/tz.Resolve("US/Eastern")` returns the equivalent (canonicalization) of `Resolve("America/New_York")`.
- [ ] `internal/tz.Resolve("+05:30")` returns fixed-offset Location, DST is always false.
- [ ] `internal/tz.Resolve("+15:00")` returns internal `ErrUnsupportedTimeZone` (out of ±14:00 range), DateTimeFormat bounds mapped to `gointl.ErrUnsupportedOption`.
- [ ] `internal/tz.Resolve("Mars/Olympus")` returns internal `ErrUnsupportedTimeZone` wrapped error, the message contains `"Mars/Olympus"`.
- [ ] `internal/tz.CanonicalLink("US/Eastern") == "America/New_York"`.
- [ ] `internal/tz.CanonicalLink("America/New_York") == "America/New_York"`(canonical returned as is).
- [ ] `internal/tz.ParseOffsetString("+05:30") == 5*3600*1000 + 30*60*1000`(ms east of UTC).
- [ ] `internal/tz.ParseOffsetString("-08:00") == -8*3600*1000`.
- [ ] The `internal/cldr/VERSION` file contains three lines: `cldr=` / `icu=` / `tzdata=`, and the CI verification hash is consistent.
- [ ] After `internal/cldr/date.ResolveLocale("en-US")` succeeds, `internal/cldr/date.GregorianFor(loc)` returns the complete `Gregorian` structure, all fields are non-null.
- [ ] After `internal/cldr/date.ResolveLocale("zh-Hans-CN")` succeeds, `internal/cldr/date.GregorianFor(loc).DayPeriods.Flex` contains all 6 keys of `"morning1" / "afternoon1" / "evening1" / "night1" / "noon" / "midnight"`.
- [ ] `internal/cldr/date.DayPeriodFor(enUSDateLocale, 5, 0) == "morning1"`.
- [ ] `internal/cldr/date.DayPeriodFor(zhHansCNDateLocale, 13, 0) == "afternoon1"`.
- [ ] `internal/cldr/timezone.TimeZoneDisplayName(enUSTimeZoneLocale, "America/New_York", TimeZoneNameLongGeneric, false, time.Now().UnixMilli(), -5*3600*1000) == "Eastern Time"`.
- [ ] `internal/cldr/timezone.TimeZoneDisplayName(enUSTimeZoneLocale, "America/New_York", TimeZoneNameShort, false, instant, -5*3600*1000) == "EST"`.
- [ ] `internal/cldr/timezone.TimeZoneDisplayName(enUSTimeZoneLocale, "Mars/Olympus", TimeZoneNameShortGeneric, false, instant, -5*3600*1000) == "GMT-05:00"`.
- [ ] `internal/cldr/timezone/data.go` is generated, and the file header contains `// Code generated by tools/gen-cldr; DO NOT EDIT.`.
- [ ] `internal/cldr/date/data.go` is generated, the file header contains `// Code generated by tools/gen-cldr; DO NOT EDIT.`; there is no `internal/cldr/calendars/*` active scope placeholder file.
- [ ] The codegen tool currently reads `ca-gregorian.json`, the date-domain active scope only contains the `"gregorian"` calendar, and `SupportedCalendars()` derives `gregory` / `iso8601` from that generated payload rather than a hand-written list.
- [ ] `formatjs/packages/intl-datetimeformat/tests/timezone-name.test.ts` All fixtures passed in `datetimeformat/conformance_unified_test.go`.
- [ ] `formatjs/packages/intl-datetimeformat/tests/day-period.test.ts` All fixtures passed (including `b` / `B` flex day period).
- [ ] `BenchmarkTZ_Resolve` appears in non-blocking benchmark telemetry.
- [ ] `BenchmarkCLDR_TimeZoneDisplayName` appears in non-blocking benchmark telemetry.

---

## 7. References

### Primary

- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` — CLDR `metaZones` / `timeZoneNames` extraction path (L137-200)
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts` — UTC offset string parsing (`OFFSET_TIMEZONE_FORMAT_REGEX`,L17-18,45-74) and `getApplicableZoneData` branch (L96-100)
- `cldr-core/supplemental/metaZones.json` — zone → metazone mapping, metazoneInfo
- `cldr-dates-full/main/<locale>/timeZoneNames.json` — locale × metazone display name + exemplarCity
- `cldr-dates-full/main/<locale>/ca-gregorian.json` — Gregorian calendar full data
- `cldr-core/supplemental/dayPeriods.json` — `dayPeriodRules` boundary

### Secondary

- Go `time/tzdata` package documentation - Embedded IANA data
- Go `time.LoadLocation` Documentation — Parsing Paths
- IANA `backward` document — link List of authoritative sources
- Upstream ICU tzdata — ICU tzdata (for referees)
- `.references/intl/intl.go` — translate-agent/intl does not implement time zone (counterexample)

### Project Cross-References

- [SPEC 30 §Public API](./30-datetimeformat.md) — `DateTimeFormat.New` caller
- [SPEC 30 §Time Zone Processing](./30-datetimeformat.md) — Format does not check TZ rules during the period
- [SPEC 31 §Skeleton character table](./31-datetimeformat-skeleton.md) — `z/Z/O/v/V/X` character → `TimeZoneName` mapping
- [SPEC 50 §Codegen](./50-cldr-data.md#codegen) — `tools/gen-cldr` generator architecture
- [SPEC 50 §Version Pin](./50-cldr-data.md#version-pin) — `internal/cldr/VERSION` file structure
