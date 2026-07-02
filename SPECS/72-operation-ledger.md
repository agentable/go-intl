# 72 - Operation Ledger

Status: active
Owns: mapping public `go-intl` surfaces to ECMA-402 owners, local implementation files, and verification gates.

This ledger is the maintenance index for public API truthfulness. Package SPECS own detailed semantics; this file answers one narrower question: "why does this exported surface exist, and where is its observable contract tested?"

The ledger is a live truth table, not a compatibility defense. It records the current or same-change public surface and its ECMA-402 owner; it must not preserve historical names merely because callers may exist, and it must not list planned symbols that are not implemented in the same branch.

---

## 1. Rules

1. Every exported constructor, namespace function, method, option helper, and record family must map to an ECMA-402 owner or an explicit Go typed bridge.
2. Bridge overload families may be grouped when they share one ECMA-402 operation and differ only by Go input type.
3. A public surface with no ECMA-402 owner, no typed bridge rationale, and no verification entry must be deleted or moved to `internal`.
4. A row cannot justify a local API by historical convenience; if ECMA-402 ownership is weak, the owning SPEC must either name the Go typed bridge or remove the surface.
5. Rows for implementation gaps must point to the owning SPEC's support-tier section; the ledger itself does not accept divergences.
6. `SupportedLocalesOf` methods must use `internal/ecma402.SupportedLocalesOf`, which parses `localeMatcher`, returns formatter-owned `OptionError` values for invalid matcher input, and delegates to `internal/ecma402.SupportedLocales`.
7. Constructor and `SupportedLocalesOf` entrypoints must receive a single typed `Options` value; option count is enforced by the public Go signature.
8. `localeMatcher` parsing must use `internal/ecma402.LocaleMatcherAlgorithm` through `internal/ecma402.ResolveConstructorLocale`, `internal/ecma402.SupportedLocalesOf`, or a formatter-owned pre-resolution check; constructor validation should use `internal/ecma402.LocaleMatcherOptionInput` when the public options bag preserves omitted versus explicit empty input.
8a. Production code may enter `internal/localematcher.ResolveLocale` only through `internal/ecma402.ResolveConstructorLocale`; formatter constructors must not grow private negotiation copies.
9. String-backed enum validation must use `internal/ecma402.InvalidStringOption` unless the option needs package-specific unsupported-state handling.
10. Integer range validation must use `internal/ecma402.InvalidIntegerOption`; cross-field constraints stay in the owning formatter package.
11. Public record structs that cross host/API boundaries must marshal with ECMA-402 field names and omission semantics; do not introduce parallel `map[string]any` record APIs.
12. Every partial capability must have a capability-slice row before the public API can advertise it. The row must name the native owner, supported slice, refused slice, sentinel, truthful supported-set rule, verification, review date, and exit path.
13. Every supported-value or supported-locale expansion must land with executable owner proof: one accepted advertised case, one refused unsupported case when the domain has a known unsupported slice, resolved-options proof where the option is observable, and CLDR/backend/native-witness evidence for the advertised behavior.

> **Why**: The stable API is small enough to review by ledger. Keeping this as prose plus file references is clearer than generating a brittle public-symbol manifest.
>
> **Rejected**: Keeping this only in a temporary implementation plan. The ledger is a durable contract artifact.

---

## 2. Namespace and Locale

| Surface | ECMA-402 owner / operation | Go entrypoints | Verification |
|---------|----------------------------|----------------|--------------|
| Root canonicalization | `Intl.getCanonicalLocales` | `intl.go`, `GetCanonicalLocales` | `intl_test.go`, locale conformance fixtures |
| Root supported values | `Intl.supportedValuesOf` | `supported.go`, typed supported-value accessors | `supported_test.go`, generated CLDR/tz data contract |
| Root constructor aliases | `Intl.<Constructor>` namespace properties | `intl.go` type aliases | `intl_test.go`, SPEC 60 |
| Structured error details | Go error bridge for ECMA-402 `RangeError` / `TypeError`-equivalent failures | `errors.go`, `internal/intlerr/errors.go`, `internal/ecma402/errors.go`, `ErrorKind`, `Error`, root category sentinels | `internal/intlerr/errors_test.go`, `internal/ecma402/errors_test.go`, constructor option tests, `intl_test.go` invalid-key and public error-text tests |
| ECMA-402 record JSON | Host-boundary bridge for `resolvedOptions()`, parts, range parts, locale info, segments, and duration records | package `ResolvedOptions` structs, package `Part` / `RangePart` structs, `locale.WeekInfo`, `locale.TextInfo`, `segmenter.Segment`, `durationformat.Duration` | `record_json_test.go` |
| Locale construction | `Intl.Locale` constructor plus Go `language.Tag` bridge | `locale/locale.go`, `locale.New`, `locale.FromTag`, `locale.Parse` | `locale/construct_test.go`, `locale/locale_test.go`, `locale/conformance_unified_test.go`, `locale/list_test.go` |
| Locale identifier subtag grammar | Unicode language identifier grammar used by `Intl.Locale`, locale info subdivision prefixes, locale matching aliases, CLDR accessors, and code canonicalization | `internal/localeid` subtag validators for language, script, region, and variant subtags | `internal/localeid/localeid_test.go`, `locale/locale_test.go`, `locale/info_test.go`, `internal/ecma402/displaynames_test.go`, `internal/localematcher/compiled_test.go`, `internal/cldr/date/date_test.go`, `internal/cldr/displaynames/displaynames_test.go` |
| Locale option presence fields | `Intl.Locale(tag, options)` option bag typed bridge | `locale.Options` string pointers, `Numeric *bool`, root `gointl.String` / `gointl.Bool` | `locale/construct_test.go`, `locale/options_ptr_test.go`, `locale/locale_test.go` |
| Locale string identity | `Intl.Locale.prototype.toString`, `baseName` | `Locale.String`, `BaseName`, `MarshalText`, `UnmarshalText` | `locale/locale_test.go`, `locale/extension_test.go` |
| Locale accessors | `Intl.Locale.prototype` getters | `Calendar`, `Collation`, `HourCycle`, `CaseFirst`, `Numeric`, `NumberingSystem`, `FirstDayOfWeek`, `Language`, `Script`, `Region`, `Variants` | `locale/construct_test.go`, `locale/extension_test.go` |
| Locale maximize/minimize | `maximize`, `minimize` | `locale/canonical.go` | `locale/canonical_test.go`, CLDR likely-subtag tests |
| Locale info methods | `getCalendars`, `getCollations`, `getHourCycles`, `getNumberingSystems`, `getTimeZones`, `getWeekInfo`, `getTextInfo` | `locale/info.go` | `locale/info_test.go`, `locale/info_ownership_test.go` |
| Locale list bridge | `CanonicalizeLocaleList` typed bridge | `locale/list.go`, `locale.ParseList`, `internal/ecma402.CanonicalLocaleList` | `locale/list_test.go`, `internal/ecma402/locale_list_test.go`, formatter constructor tests |

---

## 3. Formatter Constructors

| Surface | ECMA-402 owner / operation | Go entrypoints | Verification |
|---------|----------------------------|----------------|--------------|
| Constructor options object | `new Intl.<Constructor>(locales, options?)` | package `New(locales locale.List, opts Options)` | package constructor and resolved-options tests |
| Locale negotiation | `ResolveLocale`, relevant extension keys | package constructors through `internal/ecma402.ResolveConstructorLocale`; `internal/localematcher` owns the algorithm only | package resolved-options tests, `internal/ecma402/locale_matcher_test.go`, `internal/ecma402/locale_list_test.go`, `internal/localematcher/*_test.go` |
| Supported locales | `Intl.<Constructor>.supportedLocalesOf` | package `SupportedLocalesOf`, `internal/ecma402.SupportedLocalesOf` / `SupportedLocales` | package `TestSupportedLocalesOf`, `internal/localematcher/filter_test.go`, derived available-locale alias tests |
| Resolved options | `resolvedOptions()` | package `ResolvedOptions` methods | package resolved-options tests |

---

## 4. Number and Plural

| Surface | ECMA-402 owner / operation | Go entrypoints | Verification |
|---------|----------------------------|----------------|--------------|
| Number formatting | `Intl.NumberFormat.prototype.format` | `numberformat/format.go`, `Value` constructors, `Format` | `numberformat/format_test.go`, conformance fixtures |
| Number parts | `formatToParts` | `FormatToParts` | `numberformat/format_test.go`, conformance fixtures |
| Number ranges | `formatRange`, `formatRangeToParts` | `numberformat/range.go`, `FormatRange`, `FormatRangeToParts` | `numberformat/range_test.go`, invalid-value tests, conformance fixtures |
| Number option presence fields | ECMA-402 option bag typed bridge | `LocaleMatcher *string`, `NumberingSystem *string`, `Style *string`, `Currency *string`, `CurrencyDisplay *string`, `CurrencySign *string`, `Unit *string`, `UnitDisplay *string`, `Notation *string`, `CompactDisplay *string`, `UseGrouping *string`, `SignDisplay *string`, `RoundingMode *string`, `RoundingPriority *string`, `TrailingZeroDisplay *string`, `*int` digit fields, root `gointl.String` / `gointl.Int` | `numberformat/resolved_options_test.go`, option error tests, conformance fixtures, pointer copy tests |
| Number resolved-option presence fields | `Intl.NumberFormat.prototype.resolvedOptions()` branch properties | `ResolvedOptions` pointer fields for `Currency`, `CurrencyDisplay`, `CurrencySign`, `Unit`, `UnitDisplay`, `CompactDisplay` | `numberformat/resolved_options_test.go`, `record_json_test.go`, conformance fixtures |
| Plural selection | `Intl.PluralRules.prototype.select` | `pluralrules/pluralrules.go`, `Value` constructors, `Select` | `pluralrules/pluralrules_test.go`, conformance fixtures |
| Plural range selection | `selectRange` | `SelectRange` | `pluralrules/range_test.go`, conformance fixtures |
| Plural option presence fields | ECMA-402 option bag typed bridge plus shared digit options typed bridge | `LocaleMatcher *string`, `Type *string`, `Notation *string`, `CompactDisplay *string`, `RoundingMode *string`, `RoundingPriority *string`, `TrailingZeroDisplay *string`, `*int` digit fields, root `gointl.String` / `gointl.Int` | `pluralrules/options_test.go`, pointer copy tests, conformance fixtures |
| Plural resolved-option presence fields | `Intl.PluralRules.prototype.resolvedOptions()` branch properties | `ResolvedOptions` pointer fields for digit branches and `CompactDisplay` | `pluralrules/options_test.go`, `record_json_test.go`, conformance fixtures |

---

## 5. Date, List, Relative Time, Duration

| Surface | ECMA-402 owner / operation | Go entrypoints | Verification |
|---------|----------------------------|----------------|--------------|
| Date formatting | `Intl.DateTimeFormat.prototype.format` | `datetimeformat/datetimeformat.go`, `Format` | `datetimeformat/datetimeformat_test.go`, conformance fixtures |
| Date parts | `formatToParts` | `datetimeformat/parts.go` | `datetimeformat/datetimeformat_test.go`, conformance fixtures |
| Date ranges | `formatRange`, `formatRangeToParts` | `datetimeformat/range.go` | `datetimeformat/datetimeformat_test.go`, invalid-range tests, conformance fixtures |
| Date option presence fields | ECMA-402 option bag typed bridge | `LocaleMatcher *string`, `Calendar *string`, `NumberingSystem *string`, `FormatMatcher *string`, `TimeZone *string`, `TimeZoneName *string`, component/style/hour-cycle `*string` fields, `Hour12 *bool`, `FractionalSecondDigits *int`, root `gointl.Bool` / `gointl.Int` / `gointl.String` | `datetimeformat/datetimeformat_test.go`, conformance fixtures, pointer copy tests |
| Date resolved-option presence fields | `Intl.DateTimeFormat.prototype.resolvedOptions()` branch properties | `ResolvedOptions` pointer fields for `HourCycle`, `Hour12`, component fields, `FractionalSecondDigits`, `TimeZoneName`, `DateStyle`, `TimeStyle` | `datetimeformat/datetimeformat_test.go`, `datetimeformat/resolved_options_snapshot_test.go`, `record_json_test.go`, conformance fixtures |
| List formatting | `Intl.ListFormat.prototype.format` | `listformat/format.go`, `Format` | `listformat/listformat_test.go`, conformance fixtures |
| List parts | `formatToParts` | `FormatToParts` | `listformat/listformat_test.go` |
| Relative time formatting | `Intl.RelativeTimeFormat.prototype.format` | `relativetimeformat/format.go`, `Value` bridge + `Format`; `LocaleMatcher *string`, `NumberingSystem *string`, `Style *string`, and `Numeric *string` option presence bridges | `relativetimeformat/relativetimeformat_test.go`, conformance fixtures |
| Relative time parts | `formatToParts` | `Value` bridge + `FormatToParts` | `relativetimeformat/parts_test.go`, conformance fixtures |
| Duration formatting | `Intl.DurationFormat.prototype.format` | `durationformat/format.go`, `Format` | `durationformat/durationformat_test.go`, conformance fixtures |
| Duration parts | `formatToParts` | `FormatToParts` | `durationformat/parts_test.go`, conformance fixtures |
| Duration option presence fields | ECMA-402 option bag typed bridge | `LocaleMatcher *string`, `NumberingSystem *string`, `Style *string`, per-unit style/display `*string` fields, `FractionalDigits *int`, root `gointl.String` / `gointl.Int` | `durationformat/durationformat_test.go`, conformance fixtures, pointer copy tests |
| Duration resolved fractional-digits presence | `Intl.DurationFormat.prototype.resolvedOptions()` option-present `fractionalDigits` property | `durationformat.ResolvedOptions.FractionalDigits` as a pointer field independent of `style` | `record_json_test.go`, `durationformat/durationformat_test.go` |

---

## 6. Names, Collation, Segmentation

| Surface | ECMA-402 owner / operation | Go entrypoints | Verification |
|---------|----------------------------|----------------|--------------|
| Display names lookup | `Intl.DisplayNames.prototype.of` | `displaynames/displaynames.go`, `Of(code) (string, bool, error)` | `displaynames/displaynames_test.go`, `displaynames/validate_test.go` |
| Display names code canonicalization | `CanonicalCodeForDisplayNames` | `internal/ecma402/displaynames.go` | `displaynames/validate_test.go` |
| Display names option presence fields | ECMA-402 option bag typed bridge | `LocaleMatcher *string`, `Type *string`, `Style *string`, `Fallback *string`, `LanguageDisplay *string`, root `gointl.String` | `displaynames/displaynames_test.go`, conformance fixtures |
| Collation compare | `Intl.Collator.prototype.compare` | `collator/collator.go`, `Compare` | `collator/collator_test.go`, `collator/search_sensitivity_test.go` |
| Collator option presence fields | ECMA-402 option bag typed bridge | `LocaleMatcher *string`, `Usage *string`, `Sensitivity *string`, `CaseFirst *string`, `Numeric *bool`, `IgnorePunctuation *bool`, `Collation *string`, root `gointl.Bool` / `gointl.String` | `collator/collator_test.go`, conformance fixtures, pointer copy tests |
| Collator resolved collation presence | `Intl.Collator.prototype.resolvedOptions()` always-present `collation` property | `collator.ResolvedOptions.Collation` as a non-omitempty string defaulting to `"default"` | `record_json_test.go`, `collator/collator_test.go` |
| Segmentation view | `Intl.Segmenter.prototype.segment` | `segmenter/segmenter.go`, `Segment` | `segmenter/segmenter_test.go` |
| Segment iteration | `Segments` iterable | `Segments.All() iter.Seq[Segment]` | `segmenter/segmenter_test.go` |
| Segment containing lookup | `Segments.prototype.containing` | `Containing`, `ContainingByte` | `segmenter/segmenter_test.go` |

---

## 7. Capability Slice Ledger

This table records partial capabilities that are intentionally narrower than the
full ECMA-402 surface. It does not grant permission to ship approximations. Each
row points to the owning SPEC section that holds the detailed support tier,
rationale, `review_after`, and removal path.

| Capability slice | Native owner | Supported slice | Refused slice | Sentinel / supported-set rule | Verification | Exit path |
|------------------|--------------|-----------------|---------------|-------------------------------|--------------|-----------|
| DateTimeFormat calendar data | `Intl.DateTimeFormat` calendar option and `ResolveLocale` `ca` key | Gregorian / ISO-8601 observable formatting backed by generated data | Malformed calendar syntax; well-formed unsupported calendars fall back through `ResolveLocale` | Malformed calendar requests match `gointl.ErrInvalidOption`; `SupportedCalendars()` and date data accessors expose only generated supported calendars plus ECMA-402 required `iso8601` | SPEC 30 §3.1, `intl_test.go` calendar supported-values test, `datetimeformat/*_test.go` | Generate non-Gregorian payloads, implement calendar arithmetic equivalent to ECMA-402 `ToLocalTime`, add Generated reference/native fixtures, then expand `SupportedCalendars()`; review_after 2026-09-30 |
| Collator search usage | `Intl.Collator` `usage = "search"` | `usage = "sort"` with active `x/text/collate` behavior | Explicit `SearchUsage` | `gointl.ErrUnsupportedOption`; `SupportedLocalesOf` remains a locale capability check and must not imply search tailoring | SPEC 45 §1.1, `collator/search_sensitivity_test.go`, `collator/collator_test.go` | Identify a CLDR/x/text-backed search tailoring path, add native comparison fixtures, then accept `SearchUsage`; review_after 2026-09-30 |
| Collator case-first tailoring | `Intl.Collator` `caseFirst` and locale `kf` key | Default / `false` case-first behavior | Explicit `caseFirst=upper|lower`; locale `kf=upper|lower` falls back through `ResolveLocale` | Explicit unsupported caseFirst matches `gointl.ErrUnsupportedOption`; `ResolvedOptions().CaseFirst` reports only behavior the backend applies | SPEC 45 §1.1, `collator/collator_test.go`, `collator/testdata/xfail.json` | Add backend support or dependency report, then verify resolved options and ordering fixtures; review_after 2026-09-30 |
| Collator collation tailoring | `Intl.Collator` `collation` option and locale `co` key | Default collation behavior plus locale-scoped backend specializations such as German `phonebk` | Malformed collation syntax; well-formed unsupported collations fall back through `ResolveLocale`; Node v26 reflects explicit option values in `ResolvedOptions().Locale`, while ECMA-402 and FormatJS keep option overrides out of the locale tag | Malformed collation requests match `gointl.ErrInvalidOption`; `SupportedCollations()` advertises only values extracted from active `x/text/collate.Supported()` extension tags, excluding `default`, `standard`, and `search` | SPEC 45 §1.1, `intl_test.go` collation supported-values test, `collator/collator_test.go`, `collator/testdata/conformance/*` | Keep the explicit-option native witness XFAIL unless the normative resolver semantics change; review_after 2026-09-30 |
| Segmenter dictionary/CJK tailoring | `Intl.Segmenter` locale-sensitive word and sentence segmentation | Verified UAX #29 behavior for locales in `internal/segmentation.SupportedLocales()` | Dictionary/CJK-tailored locales such as `ja`, `km`, `lo`, `my`, `th`, `zh`, `zh-Hans`, `zh-Hant` | Do not advertise through `segmenter.SupportedLocalesOf`; constructors may still resolve to an available default locale | SPEC 46 §2, `internal/segmentation/accessors_test.go`, `segmenter/segmenter_test.go` | Add or select a backend with dictionary/CJK tailoring, generate native engine fixtures, then expand `internal/segmentation.SupportedLocales()`; review_after 2026-09-30 |

---

## 8. Acceptance Criteria

- [ ] New exported surfaces update this ledger in the same change.
- [ ] Deleted public surfaces remove their ledger row and any README/SPEC references.
- [ ] Rows for narrowed implementation gaps link to the owning SPEC section with current behavior, `review_after`, and removal path.
- [ ] New partial capabilities update the capability-slice ledger before they are advertised by public supported-locale or supported-value APIs.
- [ ] Supported-locale and supported-value expansions include accepted, refused, resolved-options, and CLDR/backend/native-witness proof at the owner package or conformance layer.
- [ ] Public supported-locale and supported-value APIs do not advertise refused capability slices.
- [ ] `task lint` and `task test` pass after ledger-affecting code changes.
- [ ] `task conformance:verify` passes when fixture or divergence references change.
