# SPEC 31 — DateTimeFormat Skeleton & BestFitFormatMatcher

> **Status:** Draft (2026-05-08)
> **Owner:** `internal/ecma402/datetimeformat/skeleton.go` + `internal/ecma402/datetimeformat/matcher.go`
> **Reference contract:** `formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts` + `BestFitFormatMatcher.ts` + `BasicFormatMatcher.ts`

## Overview

Define ECMA-402 option → CLDR skeleton (such as `yMMMd`) → pattern (such as `MMM d, y`) three-stage pipeline is implemented in the algorithm layer of go-intl:

1. **Skeleton parsing** (`internal/ecma402/datetimeformat/skeleton.go`): LDML TR35 character → `Intl.DateTimeFormatOptions` field mapping, stateless algorithm.
2. **BestFitFormatMatcher / BasicFormatMatcher**(`internal/ecma402/datetimeformat/matcher.go`): Separate Basic and BestFit scoring policies; only BestFit adjusts the selected fields to the caller’s requested widths.
3. **Data access boundary**: Algorithm layer prohibits import `internal/cldr`, data slices are injected by the caller of [SPEC 30 §Calendar data access](./30-datetimeformat.md).

This SPEC does not redefine:
- `DateTimeFormat` Public API → see [SPEC 30 §Public API](./30-datetimeformat.md)
- CLDR `availableFormats` / `intervalFormats` data schema → see [SPEC 50 §Schema](./50-cldr-data.md#schema)
- `Locale.HourCycle()` linkage → see [SPEC 30 §HourCycle linkage](./30-datetimeformat.md)

---

## 1. Hierarchical contract <a id="skeleton"></a>

### 1.1 Boundary between algorithm layer and data layer

```text
internal/ecma402/datetimeformat/
├── skeleton.go # Algorithm + types: DATE_TIME_REGEX, Parse, applySkeletonToken, patternHasTimeZoneName, the Formats struct
└── matcher.go # Algorithm: BasicFormatMatcher, BestFitFormatMatcher, AdjustFieldTypes

internal/cldr/date/
├── data.go      # Generated const blobs: per-locale availableFormats / intervalFormats / formats payload
├── decode.go    # Lazy decode into runtime records
└── accessors.go # GregorianFor, DayPeriodFor, SupportedLocales, SupportedCalendars
```

**MUST** Rules:

1. The `internal/ecma402/datetimeformat/` algorithm layer **MUST** be stateless (no package-level mutable var, no `init()` side effects).
2. `internal/ecma402/datetimeformat/` **Must not** directly import `internal/cldr`; CLDR data slices are injected by the caller (`datetimeformat/` package) through parameters when `New`.
3. `internal/cldr/date/data.go` is a per-locale generated const-blob payload, produced by codegen and decoded through `internal/cldr/date`; handwritten mapping tables are prohibited.
4. The `Formats` structure (per-skeleton pattern table) in `skeleton.go` must be aligned one-to-one with the `Formats` fields output by generated-reference `parseDateTimeSkeleton`, plus the go-intl-only `PatternHasTimeZoneName` flag derived from the selected pattern.

> **Why**: Upgrading LDML does not trigger data regeneration (the algorithm is stable); upgrading CLDR does not trigger algorithm modification (the data is stable); the conformance test can independently replace the algorithm or data for difference.
>
> **Rejected**: The algorithm and data are merged into `internal/cldr` - Upgrade path coupling, algorithm testing requires mock CLDR data.
>
> **Rejected**: translate-agent style "exhaustive branch" (16 `seqEraYearMonthDay` functions) - `Intl.DateTimeFormatOptions` has a total of 11 mutually exclusive/optional fields, binomial expansion 2^11 + dateStyle/timeStyle 5×5 is not feasible.

---

## 2. Skeleton character table <a id="character-table"></a>

### 2.1 Character table source

The character table source **MUST** be an LDML TR35 Date Field Symbol Table. `DATE_TIME_REGEX` of Generated references are the authoritative porting entry:

```text
/(?:[Eec]{1,6}|G{1,5}|[Qq]{1,5}|(?:[yYur]+|U{1,5})|[ML]{1,5}|d{1,2}|D{1,3}|F{1}|[abB]{1,5}|[hkHK]{1,2}|w{1,2}|W{1}|m{1,2}|s{1,2}|[zZOvVxX]{1,4})(?=([^']*'[^']*')*[^']*$)/g
```

**MUST** Rules:

1. `internal/ecma402/datetimeformat/skeleton.go` **MUST** transplant the above regular rules and all branch semantics.
2. Character → option field mapping (`matchSkeletonPattern`) **MUST** be aligned with Generated reference 1:1:

| LDML characters | Corresponding to Options field | Field value based on character length |
|-----------|------------------|------------------|
| `G G GG GGG` | `Era` | `short`(`G/GG/GGG`) / `long`(`GGGG`) / `narrow`(`GGGGG`) |
| `y yy yyyy ...` | `Year` | `2-digit`(`yy`) / `numeric`(other) |
| `M MM MMM MMMM MMMMM` | `Month` | `numeric/2-digit/short/long/narrow` |
| `L LL LLL LLLL LLLLL` | `Month`(stand-alone) | Same as above |
| `d dd` | `Day` | `numeric/2-digit` |
| `E EE EEE EEEE EEEEE EEEEEE` | `Weekday` | `short` × 3 / `long` / `narrow` / `short` |
| `c cc ccc cccc ccccc cccccc` | `Weekday`(stand-alone) | Same as E |
| `h hh / H HH / k kk / K KK` | `Hour` + `HourCycle` | `numeric/2-digit`;hourCycle is determined by characters (h12/h23/h24/h11) |
| `m mm` | `Minute` | `numeric/2-digit` |
| `s ss` | `Second` | `numeric/2-digit` |
| `S SS SSS` | `FractionalSecondDigits` | 1/2/3 |
| `a aa aaa aaaa aaaaa / b bb ...` | Automatic day-period tokens; no resolved `DayPeriod` property | `short`/`long`/`narrow` token width |
| `B BB BBB BBBB BBBBB` | `DayPeriod`(flexible) | `short`/`long`/`narrow` |
| `z zz zzz zzzz` | `TimeZoneName` | `short` / `long`(`zzzz`) |
| `Z ZZ ZZZ ZZZZ` | `TimeZoneName` | `shortOffset` / `longOffset`(`ZZZZ`) |
| `O OOOO` | `TimeZoneName` | `shortOffset` / `longOffset` |
| `v vvvv` | `TimeZoneName` | `shortGeneric` / `longGeneric` |
| `V VV VVV VVVV` | `TimeZoneName` | IANA name / city / generic fallback; mapped to `TimeZoneNameShortGeneric` / `TimeZoneNameLongGeneric` |
| `X XX XXX XXXX XXXXX` | `TimeZoneName` | offset variants |
| `Y u U r` | `Year` | LDML extended year (`Y` week-of-year, `u` extended, `U` cyclic name, `r` related); the output under active Gregorian path is numeric year |
| `Q q` | (recognized, no field) | LDML quarter(formatting / standalone). ECMA-402 does not expose `quarter` options; parsers MUST recognize these tokens to avoid silent discarding, but currently `Formats` has no `Quarter` field |

3. `Parse(skeleton string, pattern string, hour12 *bool, hourCycle HourCycle) Formats` **MUST** return the `Formats{...}` structure (corresponding to ECMA-402 §13.1.2):
   ```go
   type Formats struct {
       Pattern                string // selected active pattern string (including literal quotes)
       Skeleton               string // original skeleton
       PatternHasTimeZoneName bool   // set when the selected pattern contains a time-zone field (z/Z/O/v/V/X); go-intl-only
       Era                    FieldStyle
       Year                   NumericStyle
       Month                  FieldStyle
       Day                    NumericStyle
       Weekday                FieldStyle
       Hour                   NumericStyle
       HourCycle              HourCycle
       Minute                 NumericStyle
       Second                 NumericStyle
       DayPeriod              FieldStyle
       FractionalSecondDigits int
       TimeZoneName           TimeZoneName
   }
   ```
   The constructor resolves the selected skeleton to a complete pattern in the requested 12/24-hour family before replacing hour characters with the exact resolved cycle. Basic selects its component record without hour12 scoring; when its clock family differs, `basicHourCyclePattern` resolves that selected record’s fields against the locale’s candidates in the target family. This realization uses the selected components, not the original caller options, and carries the variant’s interval lookup skeleton. Interval programs use the same cycle; cross-date fallbacks reuse the effective endpoint time pattern.

   `go-intl` does not retain a dormant `Pattern12` slot in the skeleton candidate. ECMA-402 and FormatJS model an optional 12-hour pattern record; this implementation stores the selected active `Pattern` plus `HourCycle`, and carries no field unless runtime code consumes it.

> **Why**: The LDML character table is the consistent point across ICU/CLDR implementations; Generated reference has completed authoritative transplantation on the TS side, and mechanical translation on the Go side can ensure byte equality.
>
> **Rejected**: Redesign character → field table - inconsistent with Generated reference, which means breaking byte equality.

### 2.2 Literal and escaping

**MUST** Rules:

1. Substrings surrounded by single quotes in Skeleton **MUST** be treated as literal, skipping the character table lookup (aligned with LDML TR35 quote semantics). Example: `'Year'` in `'Year' yyyy` is literal `"Year"`.
2. Double single quotes `''` **MUST** be output as single quote characters `'` (aligned LDML escape).
3. `processDateTimePattern(rawPattern string, result *Formats)` **MUST** handle literals correctly to prevent literal characters from triggering `matchSkeletonPattern`.

---

## 3. BestFitFormatMatcher Rating

### 3.1 Best-fit scoring policy

ECMA-402 defines BestFitFormatMatcher as implementation-defined. `Match`
selects with `scoreBestFitFormat`, then adjusts field widths. Its current policy
retains the hour-cycle mismatch penalty, a fixed 15 penalty for differing
numeric widths, and a fixed 15 penalty for numeric/text style changes. Other
style widths use directional penalties. These are local best-fit choices, not
requirements imposed by BasicFormatMatcher or ICU.

Both matchers share candidate selection (highest score, first candidate on a
tie), field-presence penalties (missing 120, extra 20), fractional-width scoring,
and the specified time-zone-name fallback scores. Width penalties are 3/6 for
one/two steps wider and 6/8 for one/two steps narrower; offset fallback costs 1.
The numeric/text and numeric-width best-fit penalty does not enter Basic.

### 3.2 adjustFieldTypes post-processing

**MUST** Rules:

1. After selecting the best `Formats`, `adjustFieldTypes(format, options)` must be called to modify the character length of the corresponding field in `format.Pattern` according to the field value of `options`. Example: `format.Year = numeric` but `options.Year = 2-digit`, replace `y` in pattern with `yy`.
2. Replacement rules **MUST** be consistent with generated-reference `BestFitFormatMatcher.ts`: alphabetic pattern can be adjusted to the requested width, but numeric pattern must not be forcibly changed to alphabetic month/day-period form. Example: Chinese `yMEd` patterns keep the numeric month token even when the localized pattern contains year/month/day markers even if `month: "long"` is requested, aligned with Generated reference.
2a. `AdjustFieldTypes` **MUST** skip `minute` and `second`: it does not rewrite their widths, matching the reference `BestFitFormatMatcher` which `continue`s past minute/second ("Don't mess with minute/second"). Rewriting them corrupts the interval (`FormatRange`) path, which parses the pattern as the skeleton; for example `FormatRange(09:05 → 09:07)` en-US `{hour, minute: numeric}` must render `"9:05 – 9:07 AM"`, not `"9:5 – 9:7 AM"`. `FractionalSecondDigits` width is still adjusted.
2b. Cached fractional candidates remain numbering-system-neutral. In `New`, the freshly compiled endpoint, interval, and cross-date fallback separator tokens consume `internal/cldr/number.Locale.NumberSymbols(resolvedNumberingSystem).Decimal`; fractional parts contain digits only. Missing symbol rows follow that accessor’s same-locale default fallback. French arab therefore retains the pinned French decimal comma; extending number-symbol data is outside this rule.

    Evidence: `datetimeformat/fraction_separator_test.go` covers French,
    ar-EG latn/arab, intervals, cross-date fallback, and cache isolation.
    `datetimeformat/pattern_program.go` owns the private compiled tokens;
    `.references/node/deps/icu-small/source/i18n/dtptngen.cpp` supplies the
    localized fractional-separator precedent.
3. Pattern scanning of `adjustFieldTypes` must maintain ASCII byte level: LDML pattern field characters are all ASCII, and field membership can use `strings.IndexByte` and other stdlib byte helpers; it is forbidden to change to rune/regex scanning.
4. Pattern scanning loop **MUST** retain explicit index advancement, because quoted literal and repeated field width will skip multi-byte segments at one time; it is prohibited to hide index jumps in the loop body after mechanically changing to `for range len(pattern)`.
5. After `adjustFieldTypes`, `format.Pattern` is the final pattern string and can be directly sent to `FormatDateTimePattern`.
6. Month width adjustment preserves the selected LDML field family: `M` uses formatting names and `L` uses stand-alone names. The compiled endpoint and interval programs keep that field identity through `Format`, `FormatToParts`, `FormatRange`, and `FormatRangeToParts`; ru/pl month-only and month-with-day cases test the distinction.

> **Why**: `adjustFieldTypes` is the equivalent of ICU `DateTimePatternGenerator::adjustFieldTypes`; without post-processing, the best-fit pattern cannot be adapted to the user's precise length requirements.
> ASCII byte scanning keeps the algorithm aligned with LDML field grammar while avoiding a private one-off membership helper.

### 3.3 BasicFormatMatcher

`MatchBasic` uses the deterministic ECMA-402 BasicFormatMatcher algorithm:

- Score only the date-time component table; hour12/hourCycle do not affect ranking.
- Compare `[2-digit, numeric, narrow, short, long]` indexes directionally,
  clamping candidate-minus-requested to [-2, 2]. Fractional seconds use [1, 2, 3].
- Apply the normative timeZoneName fallback rules and field-presence penalties.
- Preserve first-candidate ties and do not adjust the selected field widths.

The constructor realizes the selected skeleton’s full clock-family variant and then applies the exact resolved hour cycle. Merely replacing `H` with `h` cannot provide localized AM/PM placement. The existing best-fit pattern materializer is reused for that variant; it does not rescore the original caller’s Basic request.

Basic matching retains an explicitly selected flexible day-period (`B`) pattern
and its interval skeleton across hour-cycle changes; only its hour fields are
rewritten. Automatic AM/PM (`a`/`b`) patterns still select a complete clock-family
alternative. `TestBasicHourCyclePreservesFlexibleDayPeriod` verifies h23/h24
text, resolved fields, and same-day/cross-date ranges.

For current Arabic candidates, month=numeric/day=2-digit gives `d\u200f/M`
(score -3) precedence over `dd\u200f/MM` (score -6). For 2026-01-03 UTC,
`formatMatcher: "basic"` therefore emits `3\u200f/1`; resolved components and
parts describe the chosen pattern. Default best-fit keeps its separate policy.

Evidence: `matcher_test.go:TestMatchBasicUsesDirectionalWidthPenalties`,
`TestMatchBasicIgnoresHour12Preference`, and
`datetimeformat/datetimeformat_test.go:TestDateTimeFormatBasicMatcherNumericWidths`,
and `datetimeformat/hour_cycle_test.go:TestBasicHourCyclePatterns` (localized
AM/PM positions, literals, parts, and same-day/cross-date ranges).


---

## 4. Skeleton → Pattern data access

### 4.1 Data source

**MUST** Rules:

1. Skeleton → Pattern mapping **MUST** go to the `dateTimeFormats.availableFormats` node of CLDR `cldr-dates-full/main/<locale>/ca-gregorian.json`.
2. Range pattern selection **MUST** go to the `dateTimeFormats.intervalFormats` node of the same file.
3. `dateFormat / timeFormat / dateTimeFormat`’s `full / long / medium / short` **MUST** go `dates.calendars.gregorian.dateFormats` / `timeFormats` / `dateTimeFormats`.
4. **Disable** handwritten skeleton → pattern mapping table; **Prohibit** skipping codegen through `// hard-coded` annotation.

> **Why**: CLDR is upstream authoritative data; handwritten tables will drift when CLDR is upgraded, and ~100 locale × dozens of skeletons cannot be maintained manually.

### 4.2 Data slice injection

**MUST** Rules:

1. The Skeleton algorithm layer **MUST** accept `Formats[]` slices as input parameters; **forbid** the algorithm layer to check CLDR by itself.
2. Call flow (`datetimeformat.New` internal):
   ```text
   1. gregorian   := internal/cldr/date.GregorianFor(loc)
   2. cldrFormats := gregorian.AvailableFormats
   3. result      := skeleton.Match(options, cldrFormats)  // BestFitFormatMatcher
   4. f.pattern   := result.Pattern // Cache to DateTimeFormat slot
   ```
3. The `Formats[]` slice **MUST** have been parsed by `parseDateTimeSkeleton` (at codegen time, or cached after lazy parsing in the `internal/cldr/date` accessor), and the algorithm layer receives the parsed `Formats` structure.

```go
// Algorithm layer signature (no implementation)
package skeleton

func Match(opts Options, formats []Formats) Formats           // BestFitFormatMatcher
func MatchBasic(opts Options, formats []Formats) Formats      // BasicFormatMatcher
func Parse(skeleton string, pattern string, hour12 *bool, hourCycle HourCycle) Formats // Separate skeleton string parsing
```

> **Why**: The algorithm layer is stateless and can be tested independently; data slices are injected by the caller, in line with CLAUDE.md "no production Go code beyond signatures" and the KISS principle.

### 4.3 Generation-time executable grammar

`tools/gen-cldr` **MUST** reject malformed patterns before they enter the
const payload. Validation has two consumer-aware layers:

1. Date/time style patterns, interval patterns, and `availableFormats`
   skeletons whose fields are executable by the current runtime validate LDML
   quote balance, active field symbols, and supported widths. Retained
   quarter/week skeletons that no public path can select remain source data;
   the validator must not pretend the runtime executes them.
2. Date-time combinations, date-time-at combinations, interval fallback, and
   the active `AppendItems["Timezone"]` consumer require exactly one `{0}` and
   one `{1}` and reject every other placeholder. Other retained append-item
   records may contain at most one `{2}` because CLDR uses it for the field
   display name, but unmatched braces, duplicates, and unknown indexes are
   always errors.

Errors must include source path, locale, calendar, and the relevant
skeleton/style/field key. Runtime parsers consume generated patterns as already
validated data and do not grow a second defensive grammar.

---

## 5. ICU Skeleton string syntax compatible

**MUST** Rules:

1. The private `parseDateTimeSkeleton` **MUST** consume ICU skeleton syntax such as `"yMMMd"` and `"hms"` for generated candidates and internal matching. Public `Options` accepts ECMA-402 components and styles, not skeleton strings.
2. Compatibility range: LDML TR35 §2.6.1 Date Field Symbol Table all characters.
3. `parseDateTimeSkeleton` **must not** change the field character order; the output `Formats.Skeleton` field retains the original value of the input string.

> **Why**: skeleton syntax is a standardized interoperability boundary. Silently accepting an incompatible dialect would make patterns non-portable and errors unpredictable.

---

## 6. Performance

**MUST** Rules:

1. `BenchmarkSkeleton_Match` and `BenchmarkSkeleton_Parse` remain benchmark telemetry.
2. `Match` must avoid unnecessary candidate slice copies.
3. Performance work must not change skeleton parsing semantics or matcher scoring.

> **Why**: skeleton parsing + matching is constructor work. Benchmarks keep it visible without turning noisy machine-local numbers into merge gates(SPEC 71).

---

## 7. Forbidden

- **FORBIDDEN** `internal/ecma402/datetimeformat/` algorithm layer import `internal/cldr` - data must be injected by the caller.
- **FORBIDDEN** Handwritten skeleton → pattern mapping table (per-locale) - codegen is required.
- **BANNED** translate-agent style "exhaustive 16 seq* functions" - does not support arbitrary skeleton strings.
- **Disabled** Calling ICU `DateTimePatternGenerator` (no cgo / no ICU).
- Basic scoring follows ECMA-402 directly. Changes to the independent best-fit policy require observable acceptance evidence, not permission from a reference implementation.
- **Disabled** Skip `adjustFieldTypes` post-processing - even if best-fit hit is accurate, length adjustment still needs to be done.
- **BANNED** The algorithm layer uses `init()` or package-level mutable var (must be stateless).
- **BANNED** `parseDateTimeSkeleton` changes LDML character semantics (field values per character length must match Generated reference 1:1).
- **DOWN** Calling `Match` or `parseDateTimeSkeleton` on a `Format` path - the pattern must be cached on `New`.

---

## 8. Acceptance Criteria

- [ ] `formatjs/packages/intl-datetimeformat/tests/skeleton-resolution.test.ts` All fixtures in `internal/ecma402/datetimeformat/skeleton_test.go` pass (byte-equality).
- [ ] `formatjs/packages/intl-datetimeformat/tests/best-fit-format-matcher.test.ts` All fixtures passed in `internal/ecma402/datetimeformat/matcher_test.go`.
- [ ] `formatjs/packages/intl-datetimeformat/tests/basic-format-matcher.test.ts` All fixtures passed.
- [ ] `parseDateTimeSkeleton("yMMMd")` returns `Formats{Year:numeric, Month:short, Day:numeric, Skeleton:"yMMMd"}`.
- [ ] `parseDateTimeSkeleton("hms")` returns `Formats{Hour:numeric, HourCycle:h12, Minute:numeric, Second:numeric}`.
- [ ] `parseDateTimeSkeleton("HHmmss")` returns `Formats{Hour:2-digit, HourCycle:h23, Minute:2-digit, Second:2-digit}`.
- [ ] `Match(opts={Year:numeric, Month:long}, formats=...)` hits `"MMMM y"` under `en-US` but not `"y"` or `"MMMM"` (complete fields take precedence).
- [ ] `Match` When testing the algorithm layer independently, there is no need to import `internal/cldr` (import path check passed).
- [ ] `internal/ecma402/datetimeformat/` package-level `init()` count = 0(`grep -c "func init" *.go == 0`).
- [ ] `adjustFieldTypes(format={Year:numeric, Pattern:"y"}, opts={Year:"2-digit"})` modifies `format.Pattern` to `"yy"`.
- [ ] Basic uses normative directional scoring without hour12 preference; BestFit retains independent policy and width adjustment. Public Arabic numeric-width regression and both text/parts projections pass.
- [ ] `BenchmarkSkeleton_Match` appears in non-blocking benchmark telemetry.
- [ ] `BenchmarkSkeleton_Parse` appears in non-blocking benchmark telemetry.
- [ ] LDML literal test: `parseDateTimeSkeleton("'Year:' y")` output `Formats.Pattern` contains literal `"Year:"`, and `Year:numeric`.
- [ ] LDML escape test: `parseDateTimeSkeleton("'it''s' y")` literal part outputs `"it's"`.
- [ ] Generator tests reject unbalanced quotes, unknown executable fields, unsupported widths, malformed indexed placeholders, duplicate `{0}`/`{1}`, and forbidden `{2}` with source context.
- [ ] Current pinned data regenerates byte-identically; inactive quarter/week skeletons are not falsely advertised as executable, and only inactive append-item records may retain one `{2}`.

---

## 9. References

### Primary

- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts` — `DATE_TIME_REGEX`, `matchSkeletonPattern`, `parseDateTimeSkeleton`, `processDateTimePattern`
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/BestFitFormatMatcher.ts` — `bestFitFormatMatcherScore`, `adjustFieldTypes`, scoring constant
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/BasicFormatMatcher.ts` — `BasicFormatMatcher` algorithm
- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` — CLDR `availableFormats` Extraction path (data generation reference)
- `.references/formatjs/packages/intl-datetimeformat/tests/skeleton-resolution.test.ts` — main conformance fixture
- `.references/formatjs/packages/intl-datetimeformat/tests/best-fit-format-matcher.test.ts` — matcher fixture
- `.references/formatjs/packages/intl-datetimeformat/tests/basic-format-matcher.test.ts` — basic matcher fixture

### Secondary

- LDML TR35 §2.6.1 Date Field Symbol Table — Character table authoritative document (Unicode standard)
- ICU4J `com.ibm.icu.text.DateTimePatternGenerator` — `adjustFieldTypes` C++/Java equivalent

### Project Cross-References

- [SPEC 30 §Public API](./30-datetimeformat.md) — `DateTimeFormat.New` caller
- [SPEC 30 §HourCycle Linkage](./30-datetimeformat.md) — `Hour12 / HourCycle / Locale.HourCycle()` Analysis
- [SPEC 32 §Calendar Data](./32-datetimeformat-tz.md#calendar-data) — `internal/cldr` Gregorian Data Access Layer
- [SPEC 50 §Schema](./50-cldr-data.md#schema) — `availableFormats` / `intervalFormats` JSON schema
