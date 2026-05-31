# SPEC 50 — CLDR Data & Codegen

> **Status:** Revised (2026-05-31)
> **Priority:** High (all formatter data bottom layer; blocking SPEC 10 / 20 / 30 / 31 / 32 / 40)
> **Authority:** CLDR / ICU / tzdata release data are the upstream authorities. This SPEC records the `internal/cldr/` package structure, version peg, active scope locale scope, `tools/gen-cldr/` code generator schema.

---

## Overview

`internal/cldr/` is the data bottom layer of go-intl CLDR-backed formatter. It compiles the locale-aware table of [Unicode CLDR](https://cldr.unicode.org/) (numeric symbols, numeric time separators, currency precision, date mode, plurality rules, list mode, relative time field, duration unit mode, time zone display name, likely subtags, and reference locale matching data) into Go literals during the generation period, and consumes them through the O(1) accessor for the formatter package at runtime. The runtime algorithm boundaries for `collator`, `segmenter`, and active best-fit locale matching are defined by their respective SPECs. `tools/gen-cldr/` is the code generator for this package and is maintained as an independent Go module.

This SPEC definition: data source selection (direct consumption `unicode-org/cldr-json`), embedding strategy (Go literal, **not** use `//go:embed *.json`), active locale scope (single `locales` collection, CLDR-backed surface payload coverage and accessor-derived supported set), version pinning (`cldr=48.1.0` / `icu=78` / `tzdata=2025b`), default data portrait (104 locale curated default), generator location (`tools/gen-cldr/` independent `go.mod`), file layout and accessor interface, upgrade process.

---

## CLDR Pin Rationale <a id="cldr-pin-rationale"></a>

The current active data baseline is fixed at `cldr=48.1.0` / `icu=78` / `tzdata=2025b`, and is exposed to tests and data contracts by `internal/cldr/manifest.go`. The three must move together as a conformance baseline: the CLDR JSON provides the locale payload, the ICU version represents the pattern matching and reference behavior context, and the tzdata version determines the IANA zone/link and display name boundaries.

Version upgrades must not be mixed with ordinary code changes as an opportunistic dependency refresh. Upgrades must also update `internal/cldr/VERSION`, rerun `task data`, pass `task data:check` / `task conformance:verify` / `task verify`, and document user-observable data differences in this SPEC or release notes. `tools/gen-cldr` The first line of the generated file must reference this section to ensure that any generated CLDR payload can be traced back to this set of maintenance rules.

---

## 1. Data Source

### 1.1 Selection

CLDR data **MUST** directly consume the npm image package (`cldr-bcp47` / `cldr-core` / `cldr-dates-full` / `cldr-localenames-full` / `cldr-misc-full` / `cldr-numbers-full` / `cldr-segments-full` / `cldr-units-full`) of [`unicode-org/cldr-json`](https://github.com/unicode-org/cldr-json), which has the same origin as FormatJS.

> **Why**:
> 1. **Same source** - locked to the same CLDR version as FormatJS, conformance failure must be caused by code differences rather than data differences, and the debugging path is clear.
> 2. **Stable** - `cldr-json` npm package release rhythm is consistent with the CLDR version, twice a year (spring/autumn).
> 3. **Shape matching** - JSON directly connects to Go `encoding/json` (only used during the generation period), no LDML XML parser is required.
>
> **Rejected**:
> - **`golang.org/x/text/cldr`**: The data version does not constitute the same conformance baseline as the CLDR pinned version of go-intl/formatjs; its design goal is `x/text` internal tool consumption and not exposed to production code; the data form is Go struct, not JSON, and the transformation cost is high.
> - **FormatJS intermediate products** (`@formatjs_generated/cldr.locale/`, etc.): It is a TS source + Bazel compiled product, **not sent to npm** (FormatJS `knowledge-base/001-repo-layout.md` expressly states "Generated files are compiled and packaged, not checked into git"); to use it, you must first run the FormatJS Bazel pipeline in CI, and the operation and maintenance cost is unacceptable.
> - **ICU CGO binding** (`goccy/go-icu` / `goodsign/icu4go`): The former GitHub 404 does not exist, the latter has no active maintenance; CGO conflicts with SPEC 00 §1.1 "Does not depend on ICU C/C++".

### 1.2 Extraction scope (active scope) <a id="schema"></a>

Each `internal/cldr/<file>.go` corresponds to an extraction entry:

| `internal/cldr/<file>.go` | CLDR source | Extract fields |
|---------------------------|-----------|---------|
| `numbers.go` | `cldr-numbers-full/main/<locale>/numbers.json` | symbols(decimal/group/percent/plus/minus/NaN/Infinity/timeSeparator), decimalFormats, percentFormats, currencyFormats, scientificFormats, numberingSystems |
| `dates.go` | `cldr-dates-full/main/<locale>/ca-gregorian.json` | era / month / weekday / quarter / dayPeriod name (stand-alone × format, wide / abbreviated / narrow), dateFormats, timeFormats, dateTimeFormats, availableFormats, intervalFormats |
| `metazones.go` | `cldr-core/supplemental/metaZones.json` + `cldr-dates-full/main/<locale>/timeZoneNames.json` | zone → metazone mapping, metazone display name (long / short × generic / standard / daylight), exemplarCity |
| `timezones.go` | `cldr-core/supplemental/metaZones.json` + IANA `backward` | IANA link → canonical zone mapping |
| `collations.go` | `cldr-bcp47/bcp47/collation.json` | canonical collation candidate identifiers, excluding deprecated and non-sort internal values; not a root support promise until `collator` can apply them |
| `plural/*.go` | `cldr-core/supplemental/plurals.json` + `ordinals.json` + `pluralRanges.json` | cardinal rules, ordinal rules, pluralRanges (this file is output by SPEC 40 codegen; this SPEC only agrees on the file location) |
| `preference.go` | `cldr-core/supplemental/timeData.json` + `weekData.json` + `calendarPreferenceData.json` | hourCycle preference, firstDay / weekend / minDays, calendar preference for each region |
| `likely_subtags.go` | `cldr-core/supplemental/likelySubtags.json` | language → maximized (script, region) mapping |
| `locale_matching.go` | `cldr-core/supplemental/languageMatching.json` | reference paradigmLocales, matchVariables, and distance table accessors for future best-fit expansion; active SPEC 11 matching does not import this package |
| `regions.go` | `cldr-core/supplemental/territoryContainment.json` | matchVariables area expansion ($enUS / $cnsar / $americas / $maghreb, etc.) |
| `currencies.go` | `cldr-numbers-full/main/<locale>/currencies.json` + `cldr-core/supplemental/currencyData.json` | Currency display name (long / short / narrow), plural form, defaultFractionDigits, cashDigits, rounding |
| `units.go` | `cldr-units-full/main/<locale>/units.json` | NumberFormat sanctioned unit plural mode, DurationFormat duration unit mode (long / short / narrow), compoundUnitPatterns |
| `list_patterns.go` | `cldr-misc-full/main/<locale>/listPatterns.json` | ListFormat pair/start/middle/end pattern of `conjunction` / `disjunction` / `unit` × `long` / `short` / `narrow` |
| `relative_time.go` | `cldr-dates-full/main/<locale>/dateFields.json` | long/short/narrow relative and relativeTime pattern of RelativeTimeFormat year/quarter/month/week/day/hour/minute/second |
| `locales.go` | `cldr-core/availableLocales.json` | Supported locale list + `AvailableLocales()` accessor |
| `supported.go` | Derived from generated runtime maps, generated compact supported-value indexes, and generator constants | formatter-specific supported locale list + CLDR-backed inputs of root supported-value accessors |
| `manifest.go` | `internal/cldr/VERSION` + `tools/locale-profile.json` + CLDR package metadata | generator name, CLDR / ICU / tzdata pins, active locale profile, input file SHA-256 hash |
| `strings.go` | (derived) | Shared deduplication string table (`const _data string`) |

### 1.3 Locale profile schema

`tools/locale-profile.json` is a maintenance record of generated CLDR payload coverage. The contract of a world-class Intl library is not "all constructors share the same supported-locale answer", but "any constructor only claims support if the real data or algorithm suffices". For CLDR-backed surface, the locale in the profile must generate the corresponding payload; for non-CLDR runtime engines such as `collator` / `segmenter`, the supported set is defined separately by the corresponding engine SPEC.

| Key | Consumed by | Required |
|-----|-------------|----------|
| `locales` | CLDR-backed surface(date / number / plural / list / relativetime / duration / displaynames / unit / currency / time zone display name) generates the corresponding runtime payload for each locale in the list | **Yes, and unique** |

MUST rules:

1. Profile JSON **only** contains a key of `locales`. Adding a new CLDR-backed surface must not introduce a new profile key - the new surface generates the payload from the same locale profile.
2. `tools/gen-cldr` and `tools/gen-plural-rules` **MUST** use strict JSON decoder to read the profile. Unknown keys, multiple top-level JSON values, and empty profiles are all generation errors; `task data:contract` must verify that the profile in the warehouse is still the schema.
3. A CLDR-backed formatter `SupportedLocalesOf` **MUST** be derived through the generated supported-locale accessor and may not directly read the `locales` profile or `AvailableLocales()`. This ensures that `SupportedLocalesOf` is consistent with the actual payload.
4. Each CLDR-backed generated supported-locale accessor must reflect the active generated payload and be verified as a subset of `Manifest().LocaleProfile` by `task data:contract`; each locale in the profile must be able to fall to the real payload locale through ECMA-402 lookup. Duplicate zone alias data must not be generated for the sake of "looking consistent".
5. Non-CLDR runtime engines (`collator` supported by `golang.org/x/text/collate`, `segmenter` supported by `github.com/rivo/uniseg`) must not automatically inherit the `locales` profile; they must maintain a narrower set that can be truly supported by the algorithm.
6. Change the `locales` collection = change the conformance surface of the library. Each change must go through `task data` rebirth + `task verify` full return.

> **Why**: The early 7 surface-specific keys (`formatterLocales` / `numberLocales` / `pluralLocales` / `listLocales` / `relativeTimeLocales` / `durationLocales` / `displayNamesLocales`) exist for binary size optimization; the cost is that contributors must understand the subset rules and default chain every time a new CLDR-backed surface is added, and generate "`hi` with plural But there is no semi-supported surprise like "duration". Actual measurement shows that the cost of folding to a single `locales` key is that the binary delta changes from 3.2 MB → 9 MB (acceptable), in exchange for the concept 7 → 1. The supported set of the constructor layer must still be derived from the actual payload / engine capabilities to avoid over-claiming.
>
> **Rejected**:
> - Multiple surface-specific keys (original solution): Optimize the wrong object. The binary size is acceptable; the mental cost of contributor + the user's surprise is unacceptable.
> - 2-tier(`locales` + `richLocales`): There is still CLDR payload "semi-supported" semantics, but the 7 types of semi-support are converged to 2 types - treating the symptoms but not the root cause.
> - per-locale tag tag (`{"en": "full", "kk": "plural-only"}`): Adds new JSON form with higher complexity.

### 1.4 Identifier source decision

| Identifier | Source | Remarks |
|--------|------|------|
| Currency (ISO 4217 encoding + precision) | CLDR `currencyData.json` | **Do not** introduce an independent ISO 4217 table (to avoid dual-source synchronization); **Do not** introduce `bojanz/currency` (self-maintaining CLDR derived table, version independent, drifting with `internal/cldr/VERSION`) |
| Time zone (IANA zone) | `time/tzdata` (transition table) + CLDR `metaZones.json` (display name) | tzdata is injected by SPEC 32; this SPEC only generates `metazones.go` display name table |
| Sorting collation candidate identifiers | CLDR BCP47 `collation.json` | Generation filter deprecated, `ducet`, `search`, `standard`; active root support narrowed by Collator backend capabilities by `internal/collation` |
| Unit identifier (sanctioned list) | ECMA-402 spec hardcode into `internal/ecma402/numberformat/constants.go` | spec list authority; CLDR provides schema but not sanctioned list |

> **Why**: Currency precision belongs to the pinned CLDR data baseline; `bojanz/currency` carries a separately maintained CLDR-derived table and would introduce "two copies of CLDR data" that must be verified independently. Spec-sanctioned units are normative, and any CLDR-driven detection may deviate from the spec.
>
> **Rejected**:
> - Introduce ISO 4217 static table (inconsistent with CLDR / ICU / FormatJS, conformance reverse divergence appears).
> - `bojanz/currency` as an ECMA-402 data source because it splits the data version path.
> - CLDR detects the list of sanctioned units (spec is authoritative, CLDR table is wider).

---

## 2. Embedding Strategy

### 2.1 Go literal, not `//go:embed *.json`

CLDR data **MUST** be compiled into Go literals (constant strings, map literals, slice literals), generated by `tools/gen-cldr/`. **BANNED**:

- `//go:embed *.json` + runtime `encoding/json.Unmarshal`.
- `//go:embed *.bin` + custom binary decoder (only considered for consumer-driven expansion).
- Network pull or file I/O on startup.

> **Why**:
> 1. **Zero cost at runtime** - Go literals in `.rodata` segment, O(1) access; JSON parse 100 locales × several data files ≈ 100–500 ms. Unacceptable cold start overhead (CLI / Lambda / short connection service is especially sensitive).
> 2. **Remove JSON dependency** - Do not introduce `encoding/json` during runtime (only used during generation); reduce binary and dependency graphs.
> 3. **Go ecological precedent** - `golang.org/x/text/internal/data`, `x/text/currency`, `translate-agent/intl` all go this way.
> 4. **Controllable size** - Actual measurement `translate-agent/intl/internal/cldr/data.go` single file is 400 KB / 3203 lines, covering hundreds of locale era/month/weekday; after string deduplication, binary is 30–50% smaller than JSON.
> 5. **Clear runtime contract** - CLDR data is generated from Go source; hot path does not do file I/O and does not parse JSON.
>
> **Rejected**:
> - **`//go:embed *.json` + runtime deserialization**: cold start overhead + JSON/reflection dependencies + larger size.
> - **Custom binary + decoder** (FormatJS uses base36 + `|` in tz_data to separate): active scope is not necessary; only consider (consumer-driven expansion) when the binary size is overwhelming (such as WASM).
> - **Runtime network pull** (hosted on cdn/s3): Go binary monolithic philosophy + conflicting requirements for offline deployment.

### 2.2 Shared string table

In order to reduce the binary size, you should use a single `const _data string` giant string + `[start:end]` index access deduplication mode:

```go
// internal/cldr/strings.go (signature; generated by codegen)
package cldr

// Output by tools/gen-cldr/codegen/stringtable.go.
const _data = "" +
    "JanuaryFebruaryMarchAprilMayJuneJulyAugustSeptemberOctober..." +
// (Splicing era/month/weekday/dayPeriod/... of all locales)

// Get the corresponding string through [start:end] (zero allocation).
type sliceRef struct {
    start, length uint32
}
```

> **Why**: `translate-agent/intl/internal/cldr/data.go` is verified to be feasible; Go compiler puts `const _data` directly into `.rodata`; `_data[start:end]` is a string slice operation, with zero allocation.
>
> **Rejected**:
> - Each field has an independent `var s = "..."`: duplicate strings cannot be deduplicated (`"January"` is repeated about 10 times in 100 locales); binary increases by about 30%.
> - `var stringTable = []string{...}`: slice header overhead; `s[i]` is indirect access and loses constant folding.

### 2.3 Volume estimation and target

| Data type | Range | Estimated volume |
|---------|------|---------|
| era / month / weekday name | translate-agent/intl coverage (~100 locale × Gregorian) | 400 KB |
| dayPeriod / availableFormats / intervalFormats | + dateStyle/timeStyle and skeleton matrices | +300 KB |
| metaZones(zone → metazone → display name) | ~430 zones × ~50 display name locale | +200 KB |
| number symbols / currency / unit pattern | 30–50 line pattern strings per locale | +250 KB |
| currency display names + plurals | 200 currency × ~40 locale | +200 KB |
| compact / scientific digital pattern | small set per locale | +50 KB |
| plurals compiled product (SPEC 40 codegen) | 200 locale × ~10 lines of Go code | +50 KB |
| **Total active scope (~104 locale × full surface, measured on 2026-05-16)** | | **~9 MB (binary delta when full formatter is referenced)** |

`task build:size` runs `tools/sizecheck` and reports separate deltas for the empty baseline, `internal/cldr.AvailableLocales()`, the root facade, each direct formatter import, and all formatters together. It is not currently used as a CI gate, but SPEC 50 §3 is concerned about the "full data but linker can be cut" contract - the root facade cost is the aggregate namespace cost, and the direct formatter cost is the real signal of a single surface consumer.

### 2.4 Single file vs divided files

CLDR data **MUST** be generated in separate files according to §1.2 table (one `.go` file for each category). Each locale can be further divided into independent `data_<locale>.go` (such as `data_en_US.go`) to facilitate Go parallel compilation.

**Current status (2026-05-16)**: After S1 single profile + 104 locale, `internal/cldr/` single package ~234K lines of Go source, `metazones.go` single file ~38K lines, `units.go` ~91K lines, `displaynames/data.go` ~143K lines. The Go compiler is single-threaded in a single package, and the measured cold compilation time is ~5 min (`-race` ~10 min); warm cache < 1s.

**Unmade optimization (explicit placeholder)**: Split a single package into per-surface or per-locale sub-packages, and let Go compile across packages in parallel. This is the only way to reduce cold compilation from minutes to seconds (the single-thread limit of the Go compiler is the language ceiling, not a codegen problem). This split will change the import path and accessor boundary of `internal/cldr/`, which requires a separate SPEC revision + PR; this SPEC only leaves placeholders.

> **Why**:
> 1. A 400 KB single file of translate-agent/intl takes about 1 second to compile in `go build`; 104 locale × full surface single package is measured in ~5 minutes.
> 2. After sub-packaging (not just sub-files), `go build` compiles cross-packages in parallel, which can reduce 5 min → 30s.
> 3. The IDE jump can locate the specific constants of a specific locale.
>
> **Rejected**:
> - **Single `data.go` accommodates all locales**: Time-consuming compilation + slow IDE loading.
> - **Only divided into files but not into packages** (current status): The Go compiler is single-threaded in a single package, and the number of files does not solve the root cause.

---

## 3. active scope Locale Scope

### 3.1 Curated default: 104 modern tier 1 / 2 locale

active scope **MUST** fully embed all 104 CLDR modern tier 1 / 2 locales in the `tools/locale-profile.json` `locales` list. This list is the current default product portrait, not a temporary seed, nor a Node full ICU commitment. The specific list is maintained by this file, `internal/cldr.Manifest()` and `task data:contract` lock the version number and locale count. Default profile change = conformance surface change, evidence must be provided. **BANNED**:

- per-locale tree-shaking (FormatJS `__addLocaleData` style).
- sub-locale lazy loading (file I/O pulls `zh-Hans` data during runtime).
- build tag split (intl_full / intl_minimal, **consumer-driven expansion** will only be considered).

Candidate list (maintained by `tools/locale-profile.json` during the implementation period; **SPEC does not fix the specific list**, only gives directions):

```text
# Must contain (active scope strong constraint)
en, en-US, en-GB, en-CA, en-AU
zh, zh-Hans, zh-Hans-CN, zh-Hant, zh-Hant-TW, zh-Hant-HK
ja, ja-JP
ko, ko-KR
fr, fr-FR, fr-CA
de, de-DE, de-AT
es, es-ES, es-419, es-MX
pt, pt-BR, pt-PT
ru, ru-RU
ar, ar-SA, ar-EG
hi, hi-IN
it, it-IT
nl, nl-NL
pl, pl-PL
tr, tr-TR
th, th-TH
vi, vi-VN
id, id-ID

# Should be included (covering Tier 1 area)
sv, da, nb, fi, cs, hu, ro, sk, uk, el, he, fa,
ms, bn, ta, te, ml, mr, gu, kn,
... (104 in total, subject to tools/locale-profile.json)
```

### 3.2 Volume Constraints

The generated code size should remain reviewable. `task build:size` reports the current profile, binary delta and warm-cache build duration; `task build:size:cold` first cleans the Go build cache, and then reports the size / compile-time evidence of the same harness. Both do not block CI, but any `tools/locale-profile.json` expansion PR **MUST** post both outputs, and pass both `task data:contract` and `task conformance:verify`. Reports must label the root facade, direct formatter, CLDR-only, and all-formatters harness respectively. Lightweight entry points (e.g. just calling `internal/cldr.AvailableLocales`) must not pull in formatter heavy data due to the same package init; large CLDR maps must be initialized via on-demand loaders or narrower package boundaries.

> **Why**:
> 1. **Pay for use** - active scope does not require a per-locale selector, but the lightweight API should not pay for the full table of number/date/time-zone; retain single binary deployment according to the lazy table of the surface, and let Go linker trim the unreferenced loader.
> 2. **Go does not have dynamic registration such as `__addLocaleData`** - Implementing this mode requires `init()` + global variable map, which violates SPEC 00 §1 "no implicit state".
> 3. **Go dead-code elimination** - Go linker can eliminate unreferenced loader functions, but package-level map initialization is the root path; the generator must avoid heavy data appearing in package-level initialization expressions.
> 4. **Single binary deployment** - No need to deploy sidecar data files; Lambda / Docker friendly.
>
> **Rejected**:
> - **per-locale subpackage** (`numberformat/locale-data/zh.js` style): Go does not have `__addLocaleData`; import relationships between subpackages pollute the dependency graph.
> - **active scope introduces build tag**(`intl_full` / `intl_minimal`): increases generator complexity; main consumer has no requirement.

### 3.3 consumer-driven expansion placeholder

consumer-driven expansion **can** introduce build tag classification, but the default strategy is an excellent curated profile, not a configuration matrix. New build profiles are only allowed when multiple hosts have clear and repeated requirements that cannot be served by 104-locale default. Baseline reference (2026-05-16 S1 actual measurement):

- Default (no tag): 104 locale × full surface, lightweight entry ~270 KB delta, full formatter reference ~9 MB delta.
- `intl_full`: All ~500 locale × full surface, estimated full formatter reference ~40 MB (not yet measured; based on linear extrapolation).
- `intl_minimal`: only `root` + `en`, estimated full formatter reference ~500 KB delta (not yet measured).

Implementation path: generator outputs `data_default.go` / `data_full.go` / `data_minimal.go`, each adding `//go:build` header. Any selectable profile PR must include size, cold compile, conformance, supported-locale behavior evidence, and explain why the default profile cannot meet the requirements. **This SPEC does not enforce consumer-driven expansion time**; SPEC 80 is a placeholder.

> **Also pay attention**: See §2.4 about `internal/cldr/` single-package cold compilation time. Build tag hierarchically solves the "volume" dimension but not the "cold compilation" dimension; true cold compilation and compression requires package splitting. The two tasks can be advanced independently.

---

## 4. Version Pinning <a id="version-pin"></a>

### 4.1 Decision (Close SPEC 00 §8 Q4)

`internal/cldr/VERSION` **MUST** store three version numbers in a single file:

```text
cldr=48.1.0
icu=78
tzdata=2025b
```

> **Why**:
> 1. **Synchronization with FormatJS** —— FormatJS current main branch `package.json` lock `cldr-*: 48.1.0` (corresponding to ICU 78); go-intl "Be consistent with FormatJS byte level" is the primary goal of SPEC 00 §1. Pinning earlier versions = institutionalizing reverse divergence of "we are right, FormatJS is wrong".
> 2. **Update Window** - CLDR 48 / ICU 78 was released on 2025-10 and has been stable for 6+ months by the time of public release.
> 3. **tzdata 2025b** - Go 1.26.3 built-in tzdata version; tzdata release at the same time as CLDR/ICU pinning.
>
> **Rejected**:
> - **CLDR 47 / ICU 76** (2024-10): One version later than FormatJS, reverse diverge.
> - **CLDR 44 / ICU 74** (2023-11): Two years behind, high failure rate with FormatJS byte equality.
> - **Follow `golang.org/x/text/cldr` version**: The data baseline is not controlled by go-intl, and the conformance goal will be handed over to the external release rhythm.

### 4.2 Upgrade process

CLDR / ICU / tzdata Any version change **required**:

1. Update three lines of `internal/cldr/VERSION`.
2. Synchronously update 8 `cldr-*` dependent versions in `tools/gen-cldr/.cldr-json/package.json`, and run `npm install --package-lock-only --prefix tools/gen-cldr/.cldr-json` to regenerate the lockfile.
3. Run `task data` to regenerate `internal/cldr/*.go`. `data:preflight` catches VERSION ↔ package.json drift before npm ci and gives an executable repair prompt.
4. Review the field-level diff with the previous version through `git diff internal/cldr/`.
5. CI on main **MUST** block unreviewed data increments (diff appears during PR review).
6. Synchronously update conformance fixture(`<package>/testdata/conformance/`) if necessary.

> **Why**: CLDR upgrades often introduce hundreds of row-level data changes (new locale, new dayPeriod, new currency); silent replacement will make a PR change unreviewable. `data:preflight` Early fail + `git diff` + CI block forces review and converges the "drift → repair" link into an executable instruction.

### 4.3 Hash consistency check

The version number in `internal/cldr/VERSION` must correspond to the data source encoded in `internal/cldr/*.go`. CI **MUST** pass `task data:check` verification (regenerated with git file byte-equal).

> **Why**: To prevent manually changing `internal/cldr/*.go` and forgetting to update `VERSION`; and vice versa.

### 4.4 Generated manifest

`tools/gen-cldr` **MUST** generate `internal/cldr/manifest.go`. The manifest is an internal audit fact, not a root public API, and the content must be stable and not contain local absolute paths:

- `Generator`: current generator name (`tools/gen-cldr`).
- `CLDR` / `ICU` / `TZData`: identical to `internal/cldr/VERSION`.
- `LocaleProfile`: Normalized `tools/locale-profile.json` locale list.
- `InputHashes`: SHA-256 of `internal/cldr/VERSION`, `tools/locale-profile.json`, 8 `cldr-*` package metadata.

`internal/cldr.Manifest()` **MUST** return slice clone to avoid internal testing or future tools from accidentally changing the global generated state. The byte-equal comparison of `task data:check` must cover the `manifest.go`; manifest drift is the data source drift.

### 4.5 CLDR data package convention <a id="cldr-data-package-convention"></a>

Every file generated by `tools/gen-cldr` **must** begin with the Go generated-file
line that points back to §4, followed by three stable metadata comments:

1. `Source`: the pinned CLDR / ICU / tzdata versions and the manifest hash
   location.
2. `Generated`: the reproducibility contract; it must never include timestamps,
   hostnames, usernames, absolute paths, or other machine-local values.
3. `Schema`: this section anchor, which records the generated package layout.

Generated data subpackages follow one convention: `accessors.go` contains the
runtime accessors, and one or more axis-specific data files contain embedded
CLDR payload. Generator tools produce both; do not hand-edit them. `go/format`
is the final formatting step, so import grouping and literal layout must be
deterministic across repeated generator runs.

---

## 5. Generator (`tools/gen-cldr/`) <a id="codegen"></a>

### 5.1 Independent Go module

`tools/gen-cldr/` **MUST** be an independent Go module (independent `go.mod`) and **not** pollute the main module dependency graph.

```text
go-intl/
├── go.mod # Main module, minimal runtime dependencies: x/text + apd + tzdata + go-cmp
└── tools/
    └── gen-cldr/
├── go.mod # Independent module; can rely on encoding/json and third-party parsing tools during generation
├── main.go # cmd entry; CLI flags / config
        ├── run.go      # generator orchestration
        ├── locale_list.go
        ├── extract/
        │   ├── dates.go
        │   ├── numbers.go
        │   ├── metazones.go
        │   ├── likely_subtags.go
│ ├── matching.go # SPEC 11 collaboration
        │   └── locales.go
        ├── codegen/
│ ├── stringtable.go # Shared deduplication string table generation
│ ├── format.go # gofmt output
│ ├── golang_literal.go # Go literal serialization (supports map / slice / struct)
        │   ├── render.go         # consumer-driven expansion render orchestration
        │   ├── dates.go
        │   ├── numbers.go
        │   ├── metazones.go
        │   └── matching.go
        └── cldr/
├── fetch.go # Pull the cldr-json npm package (or GitHub release)
            ├── source.go         # JSON source loader
├── version.go # Verify that the cldr-json version is consistent with internal/cldr/VERSION
            ├── dates.go
            ├── numbers.go
            ├── metazones.go
            ├── units.go
            ├── matching.go
            └── preference.go
```

> **Why**:
> 1. **Dependency isolation** - The main module does not introduce `encoding/json` heavily used, third-party CLDR JSON parsing tool (if any); keep the runtime dependency graph clean.
> 2. **Parallel evolution** - generator upgrade and library code upgrade can be PRed independently; CI only runs generator tests on generator PR.
> 3. **YAGNI** - generator is a one-way tool (CLDR JSON → Go literal); it does not need to be consumed externally, and independent modules do not increase the mental burden of Go users.
>
> **Rejected**:
> - **Same main module** (`internal/gen/`): the solution of translate-agent/intl, but go-intl generator requires more dependencies (JSON parser / CLI / fetch), which will pollute the main module.
> - **Independent warehouse** (`agentable/go-intl-gen-cldr`): The synchronized PR flow is complicated (three warehouses: SPEC + data + library); the CLDR upgrade path is fragmented.

### 5.2 Codegen tool stack

`tools/gen-cldr/codegen/` **MUST** remain stdlib-only: output constructed with deterministic strings/literals, finally formatted with `go/format`. **BANNED** `dave/jennifer` or other codegen frameworks.

```go
// tools/gen-cldr/codegen/golang_literal.go (signature; not complete implementation)
package codegen

import (
    "go/format"
    "io"
)

// EmitLiteral serializes any Go value into a legal Go source literal (map / slice / struct).
// The output goes to go/format.Source to ensure that gofmt is consistent.
func EmitLiteral(w io.Writer, value any) error

// FormatFile formats the complete Go source file into gofmt source code.
func FormatFile(src []byte) ([]byte, error)
```

> **Why**:
> 1. **stdlib is sufficient** - the current generation is deterministic Go literals and small accessor functions; string construction + `go/format` is more straightforward than the AST/template framework.
> 2. **`dave/jennifer` stasis** - No new commits after 2024-09; new code should not introduce stale code-generation tooling.
> 3. **Fewer concepts, less drift** - The generator only needs to describe how the data is placed on the disk, and does not need to maintain an additional set of Go AST DSL.
>
> **Rejected**:
> - `dave/jennifer`: See above.
> - `agentable/gendog`: an internal framework that can be evaluated later if generator complexity demands it, but the active codegen path only requires literal string concatenation plus `go fmt`; the stdlib is enough.

### 5.3 Generator entrance

```go
// tools/gen-cldr/main.go (signature)
package main

type Config struct {
CLDRDir string // Local cldr-json decompression root directory
OutDir string // internal/cldr path (default ../../internal/cldr)
Version string // CLDR version number (read and verify from internal/cldr/VERSION)
}

type Extractor interface {
    Name() string
    Extract(ctx context.Context, src *cldr.Source) (codegen.Module, error)
}

func Run(ctx context.Context, cfg Config, extractors []Extractor) error
```

Calling form: `task data` (exposed in Taskfile, not triggered by default in `task verify` to avoid CI network dependence). CI is regenerated from `task data:check` to `.tmp/cldr-check/` and compared with `internal/cldr/` byte.

### 5.4 Input and output contract

- **Input**: `tools/gen-cldr/.cldr-json/package.json` + `package-lock.json` (pinned 8 `cldr-*@48.1.0`, sha512 integrity hash written to lockfile). `task data:fetch` goes to `npm ci`, which fails in advance and does not fall back to `npm install` lock-free parsing.
- **Output**: `internal/cldr/*.go` (single directory, not written to other locations in the warehouse).
- **Side effects**: `internal/cldr/VERSION` is **not** modified by the generator; the generator is triggered by manual updates; CLDR version changes must be synchronized with `package.json` / `package-lock.json` (refer to §4.2 Upgrade process).

> **Why**: lockfile moves the "data source contract" from the Taskfile inline command to the declarative manifest; `npm ci` refuses to install when sha512 does not match, to prevent "silent diverge" caused by the replacement of the registry-side tarball. `task data:fetch` uses the `sources` / `generates` that comes with Task to compare the last modification time to implement the local cache of "no reloading if the lockfile is not changed".

### 5.5 Locale list extraction

`tools/gen-cldr/extract/locales.go` is responsible for merging CLDR available locale and active allowlist into generation locale records.

**MUST** Rules:

1. The generated locale tag list **MUST** be sorted deterministically, and `und` **MUST** be located at position 0. The stability of `Locale(0)` / `Undefined` depends on this position.
2. For collection operations such as sorting, key collection, and move-to-front, stdlib `maps` / `slices` helper must be used first; the introduction of private collection helpers or third-party collection libraries is prohibited.
3. `und` First position adjustment **MUST** be done explicitly after sorting and cannot rely on map iteration order, allowlist input order, or CLDR JSON original order.

> **Why**: When locale index is written into generated Go data, any non-deterministic order will cause byte-diff noise and even destroy the sentinel semantics of `cldr.Undefined`. stdlib `maps` / `slices` already covers this requirement, and there is no need to maintain custom relocation logic.

---

## 6. Data Access API

### 6.1 Maintained Records

`internal/cldr/` is the data access point for all formatter packets; the accessor **MUST** O(1) lookup the table, **not** allocate. The formatter package does not directly read the `internal/cldr/*.go` file-level var, and must use the accessor function.

```go
// internal/cldr/cldr.go (signature; accessor subset)
package cldr

// Locale is an opaque locale handle; internally is the dataLocale index.
type Locale uint16

// ResolveLocale resolves language.Tag into dataLocale (BCP 47 → internal index).
// The second return value false indicates that the tag is not in the CLDR data set (at this time the caller takes SPEC 11 best-fit fallback).
func ResolveLocale(tag language.Tag) (Locale, bool)

// AvailableLocales returns CLDR availableLocales universe.
func AvailableLocales() []string
// formatter-specific packages return actual generated payload locale lists (SPEC 11 ResolveLocale input parameter).
func number.SupportedLocales() []string
func date.SupportedLocales() []string
func plural.SupportedLocales() []string
func list.SupportedLocales() []string
func relativetime.SupportedLocales() []string
func displaynames.SupportedLocales() []string

// CLDR-backed Intl.supportedValuesOf data accessor (SPEC 60 consumed through root supported.go).
// Collation is the exception: CLDR exposes candidate identifiers here, while
// root supportedValuesOf("collation") must use internal/collation so it only
// advertises values the active Collator can apply.
func SupportedCalendars() []string
func SupportedCollations() []string
func SupportedCurrencies() []string
func SupportedNumberingSystems() []string
func SupportedTimeZones() []string

`SupportedCalendars()` is generated from date calendar payload keys, maps CLDR `"gregorian"` to ECMA-402 `"gregory"`, and appends `"iso8601"` only when Gregorian data exists. Supported-value accessors must use generated compact indexes rather than scanning heavy formatter payload maps on first call. They must not mirror Node's broader lists until matching formatter semantics and fixtures are active.

// Numeric symbols and patterns
type NumberSymbols struct {
    Decimal, Group, Percent, Plus, Minus, NaN, Infinity, ApproxSign string
}
func (l Locale) NumberSymbols(numberingSystem string) NumberSymbols
func (l Locale) DecimalPattern(ns string) string
func (l Locale) PercentPattern(ns string) string
func (l Locale) CurrencyPattern(ns, sign string) string
func (l Locale) CompactPattern(ns, display string, exp int, plural string) string

// ECMA-402 sanctioned simple unit identifiers for Intl.supportedValuesOf("unit")
// come from internal/ecma402. Do not generate a forwarding supported-values
// wrapper that reconnects CLDR data and ECMA-402 constants.

// Date / Time (SPEC 30 / 31 / 32 consumption)
type CalendarNames struct{ Eras, Months, Weekdays, DayPeriods, Quarters []string }
func (l Locale) CalendarNames(calendar, width, context string) CalendarNames
func (l Locale) DateFormat(style string) string
func (l Locale) TimeFormat(style string) string
func (l Locale) AvailableFormats() map[string]string  // skeleton -> pattern
func (l Locale) IntervalFormats() IntervalFormats

// Time zone (SPEC 32 consumption)
func ZoneToMetazone(zone string) string
func (l Locale) MetazoneName(metazone, kind string) string  // kind: "long-generic" / "short-standard" / ...

// Currency (SPEC 20 consumption)
func CurrencyDigits(code string) (defaultDigits, cashDigits int, rounding int)
func (l Locale) CurrencyDisplayName(code, plural string) string

// Complex number (SPEC 40 consumption) - output by SPEC 40 codegen
func (l Locale) Cardinal(operand Operand) Form
func (l Locale) Ordinal(operand Operand) Form

// Locale Preference (SPEC 10 GetCalendars and other consumption)
func (l Locale) HourCyclePreference() []string
func (l Locale) FirstDayOfWeek() time.Weekday
func (l Locale) Weekend() []time.Weekday
func (l Locale) MinimalDaysInFirstWeek() int
func (l Locale) CalendarPreference() []string

// Likely Subtags(SPEC 10 Maximize / Minimize consumption)
func MaximizeSubtags(language, script, region string) (lang, scr, reg string, ok bool)
func MinimizeSubtags(language, script, region string) (lang, scr, reg string, ok bool)

// Reference locale matching data (not an active SPEC 11 package dependency)
func MatchingDistance(desired, supported string) int
func ParadigmLocales() []string
func MatchVariables() map[string][]string
```

**MUST** Rules:

1. `AvailableLocales()` represents the CLDR availability universe; formatter locale matching **MUST** use a formatter-specific list, i.e. `internal/cldr/number.SupportedLocales()`, `internal/cldr/date.SupportedLocales()`, `internal/cldr/plural.SupportedLocales()`, `internal/cldr/list.SupportedLocales()`, `internal/cldr/relativetime.SupportedLocales()`, or `internal/cldr/displaynames.SupportedLocales()`.
2. Formatter-specific `SupportedLocales()` **MUST** be derived from generated runtime data maps, and each tag can find non-`Undefined` locales through `ResolveLocale`.
3. `internal/localematcher` **MUST NOT** import `internal/cldr`; formatter constructors inject generated supported-locale slices, maximizers, and locale-data providers into matcher APIs.
3. `internal/cldrmatch` **FORBIDDEN** from maintaining a handwritten locale list; after adding formatter data, the corresponding supported-locale accessor must be added to the CLDR generator and derived from it.
4. `internal/cldr.SupportedCollations()` **MUST** generate candidate identifiers from `cldr-bcp47`. Temporary lists such as `emoji` / `eor` must not be handwritten at runtime; root `SupportedCollations()` ** must not** directly consume this candidate list and must reflect the active Collator backend truth through `internal/collation.SupportedCollations()`.

> **Why**: `availableLocales.json` may contain some locales for which the formatter has not yet generated complete numbers/date data. The supported locale list must come from the actual generated data, otherwise `ResolveLocale` can hit a locale without formatter payload, causing the fallback behavior to drift from FormatJS.

### 6.2 accessor allocation constraints

Hot path accessors **MUST** zero-allocate after data loading is complete (returning `string` is a string slice into `_data`, returning a read-only map/slice reference). Large tables are allowed to be initialized via `sync.Once` on first access; this cold start cost must not occur in lightweight entries that do not use the corresponding formatter surface.

Root supported-value accessors are lightweight semantic indexes. Currency, calendar, numbering-system, and time-zone candidates are computed during `tools/gen-cldr` generation from the same runtime payload sources that prove support, then emitted as compact sorted slices. Runtime code may canonicalize the small generated candidate slice (for example time-zone links), but it must not load date, currency-name, number-symbol, or metazone display-name payloads merely to answer `Intl.supportedValuesOf`.

Use `go test -benchmem -run=^$ -bench=BenchmarkAccessor ./internal/cldr/` as non-blocking telemetry for accessor hot paths.

> **Why**: hot path performance. NumberFormat / DateTimeFormat reads ~10 CLDR fields in the `Format` call; repeating the read phase with each allocation amplifies GC pressure, but SPEC 71 keeps this as telemetry rather than a merge gate.
>
> **Rejected**:
> - Return copy-on-read (deep copy): Each `Format` call generates GC pressure.
> - Interface return (`type DataReader interface { ... }`): Interface method calling is slower than direct method dispatch and has no benefit.

### 6.3 Public Visibility

`internal/cldr/` Force private via `internal/` path segment. The formatter public package does not re-export `cldr.Locale` / `cldr.NumberSymbols` and other types; it is consumed indirectly through the `internal/ecma402` abstraction layer.

CLDR / ICU / tzdata pins MUST remain in `internal/cldr.Version()` and `internal/cldr/VERSION` for internal testing, auditing, and reporting use; the root package MUST NOT expose `intl.Version()` or other diagnostic APIs, as the ECMA-402 `Intl` namespace has no corresponding member.

---

## 7. Testing

### 7.1 Generator self-test

`tools/gen-cldr/` **MUST** come with `gen_test.go`, which asserts basic contracts such as "the era/month field of en/zh/ar is not empty" in the generated results:

```go
// tools/gen-cldr/gen_test.go (signature)
func TestExtractDates_BasicLocales(t *testing.T) {
    t.Parallel()
// Run extract and assert that the month name length of en / zh-Hans / ar >= 12, etc.
}
```

### 7.2 Snapshot test

`internal/cldr/snapshot_test.go` **MUST** compare the `internal/cldr/*.go` in the current git with the regenerated output, CI rejects the difference (to avoid manual modification):

```go
// internal/cldr/snapshot_test.go (signature)
func TestGenerated_NoDrift(t *testing.T) {
    t.Parallel()
// Adjust tools/gen-cldr to regenerate to the temporary directory; byte-compare and files in git
}
```

### 7.3 Conformance fixture integration

`internal/cldr/` The correctness of data access **MUST** be indirectly verified through the conformance test of the downstream formatter (SPEC 70). This SPEC does not duplicate fixture design.

---

## Forbidden

- **`//go:embed *.json` + runtime deserialization**: Violation of §2.1 embedding policy.
- ✅ Do: `tools/gen-cldr/` is serialized to a Go literal during generation.
  - ❌ Don't: `//go:embed numbers.json` + `var _ = json.Unmarshal(numbersJSON, ...)`.

- **Runtime network pull CLDR data** (from CDN/S3): breaks offline deployment, violates SPEC 00 §1 single binary philosophy.
- ✅ Do: Full embed (active scope)/ build tag split (consumer-driven expansion).
  - ❌ Don't: `http.Get("https://cdn.example/cldr-data/...")`.

- **Deserialization on startup** (including JSON parse in `init()`): Breaks cold start times.
- ✅ Do: Go literals are accessed in `.rodata` section, O(1).
  - ❌ Don't: `func init() { json.Unmarshal(_embedded, &globalData) }`.

- **sub-locale lazy loading** (runtime file I/O pulls `zh-Hans` data): Violates SPEC 00 §5.3 "Locale data is universal in active scope".
- ✅ Do: Fully embedded.
  - ❌ Don't: `func loadLocaleData(loc string) { os.Open(...) }`.

- **`golang.org/x/text/cldr`**: The data baseline is not controlled by go-intl, and may diverge with FormatJS.
- ✅ Do: Direct consumption `unicode-org/cldr-json`.
- ❌ Don't: `import "golang.org/x/text/cldr"` (in `internal/cldr/` and `tools/gen-cldr/`).

- **FormatJS intermediate product** (`@formatjs_generated/*`): If it is not sent to npm, operation and maintenance is not feasible.
- ✅ Do: Directly consume upstream `cldr-json`.
- ❌ Don't: `npm install @formatjs_generated/...` in `tools/gen-cldr/`.

- **`dave/jennifer` codegen**: stalled after 2024-09; codegen framework is not required at current scale.
- ✅ Do: stdlib JSON reading + deterministic string output + `go/format`.
- ❌ Don't: `import "github.com/dave/jennifer/jen"`(in `tools/gen-cldr/`).

- **`bojanz/currency` as ECMA-402 data source**: self-maintaining CLDR derived table, drifting with `internal/cldr/VERSION`.
- ✅ Do: CLDR `currencyData.json` directly generates `internal/cldr/currencies.go`.
- ❌ Don't: `import "github.com/bojanz/currency"` (in `numberformat/` implementation).

- **per-locale tree-shaking in active scope** (FormatJS `__addLocaleData` style): No Go equivalent mechanism + no need for main consumer.
- ✅ Do: Full embedding (active scope); build tag classification (consumer-driven expansion).
- ❌ Don't: active scope introduces `init()` to register the global map.

- **`internal/cldr/*.go` Manual modification without updating `VERSION`**: Destruction of auditable data source.
- ✅ Do: `VERSION` + `task data` two-step process, CI checks byte-equal through `task data:check`.
- ❌ Don't: Edit `internal/cldr/numbers.go` directly.

- **The generator is located in the main module**(`internal/gen/`): polluting the main module dependency graph.
- ✅ Do: `tools/gen-cldr/` independent of `go.mod`.
- ❌ Don't: Put the generator file in `internal/gen/`.

---

## Acceptance Criteria

### Data source

- [ ] The `internal/cldr/VERSION` file exists and the content is `cldr=48.1.0\nicu=78\ntzdata=2025b\n` (three lines).
- [ ] `tools/gen-cldr/cldr/version.go` verifies the `package.json` version of the local `cldr-json` directory = the `cldr=` value in `VERSION` at startup. If it is inconsistent, fail.
- [ ] `internal/cldr/` does not contain a `*.json` file; the `//go:embed` command does not appear (grep `go:embed` in `internal/cldr/*.go` returns 0).

### Package structure

- [ ] The `internal/cldr/` file list is consistent with the §1.2 table (`numbers.go` / `dates.go` / `metazones.go` / `timezones.go` / `manifest.go` / `plurals.go` / `plural/*.go` / `preference.go` / `likely_subtags.go` / `locale_matching.go` / `regions.go` / `currencies.go` / `units.go` / `locales.go` / `supported.go` / `strings.go` etc.).
- [ ] `internal/cldr/` forces the Go toolchain to reject external imports via the `internal/` path segment.
- [ ] `internal/cldr.Version()` returns the contents of `VERSION`; the root package does not expose version diagnostic functions.
- [ ] `internal/cldr.Manifest()` returns generator, version pin, locale profile and input SHA-256, and returns slice clone.

### Generator

- [ ] `tools/gen-cldr/go.mod` exists (standalone Go module).
- [ ] The `tools/gen-cldr/main.go` entry exists; `task data` is exposed in `Taskfile.yml` and triggers the generator.
- [ ] `tools/gen-cldr/codegen/` only import stdlib, use deterministic string output + `go/format`; **not** import `github.com/dave/jennifer` or other codegen frameworks.
- [ ] `tools/gen-cldr/extract/locales.go` The output locale tag list is sorted deterministically, and `und` is always at position 0.
- [ ] Partial verification of `tools/gen-cldr/` is performed from the nested module root directory: `cd tools/gen-cldr && go test ./...` and `go vet ./...`; the `./tools/gen-cldr/...` pattern of the root module is not accepted as a validation entry.

### Embedding and volume

- [ ] `internal/cldr/strings.go` contains a single `const _data string` shared deduplicated string table; accessors are accessed via `[start:end]` slices.
- [ ] `task build:size` reports growth of the root facade, direct formatter, CLDR-only, and all-formatters binary relative to an empty baseline.
- [ ] active scope generates corresponding payloads for all CLDR-backed surfaces for each tag in the `locales` list of `tools/locale-profile.json`. `SupportedLocalesOf` on CLDR-backed formatters must be derived from the actual payload accessor; non-CLDR-backed surfaces must document and test their own engine-specific supported set.

### Upgrade process

- [ ] `task data` regenerates `internal/cldr/*.go` (no git diff after any run, except `VERSION` has been changed first).
- [ ] `git diff internal/cldr/` Outputs a field-level diff with the last commit (for PR review).
- [ ] `task data:check`(CI) check `internal/cldr/*.go` is inconsistent with the regenerated result byte-equal, fail.
- [ ] CI blocks data increment PR on main until reviewer approves.

### Accessor

- [ ] §6.1 All the accessors listed in §6.1 are declared (`ResolveLocale` / `NumberSymbols` / `CalendarNames` / `IntervalFormats` / `ZoneToMetazone` / `MetazoneName` / `CurrencyDigits` / `MaximizeSubtags` / `MatchingDistance`, etc.).
- [ ] Each formatter-specific `SupportedLocales()` is non-empty, does not contain `und`, and each tag has a corresponding generated payload.
- [ ] `SupportedCalendars()` / `SupportedCurrencies()` / `SupportedNumberingSystems()` / `SupportedTimeZones()` returns canonical, sorted, unique values, and the data comes from generated runtime maps or ECMA-402 generator constants; `internal/cldr.SupportedCollations()` returns canonical, sorted, unique CLDR candidate values only.
- [ ] `SupportedCalendars()` derives from generated date calendar payloads, maps CLDR `"gregorian"` to ECMA-402 `"gregory"`, and contains `iso8601` when Gregorian data exists; `SupportedNumberingSystems()` contains the full table of ECMA-402 simple digit numbering systems, even if the current profile does not generate the corresponding CLDR symbol payload.
- [ ] `go test -benchmem -run=^$ -bench=BenchmarkAccessor ./internal/cldr/` reports accessor allocation telemetry.

### Test

- [ ] `tools/gen-cldr/gen_test.go` passes; Asserts that the era / month field of the base locale (en / zh-Hans / ar) is not empty.
- [ ] `internal/cldr/snapshot_test.go` passed; the CI verification generated result is consistent with git.
- [ ] FormatJS `tests/likely-subtags.test.ts` / `minimize.test.ts` All fixtures passed in `internal/cldr/likely_subtags_test.go`.
- [ ] `internal/cldr/locale_matching_test.go` proves generated `MatchingDistance`, `ParadigmLocales`, and `MatchVariables` accessors remain deterministic. SPEC 11 best-fit tests do not depend on these accessors unless a future spec revision adopts generated CLDR distance matching.

---

## References

### Specification

- [Unicode CLDR](https://cldr.unicode.org/) —— Data source.
- [CLDR JSON Distribution](https://github.com/unicode-org/cldr-json) —— `cldr-bcp47` / `cldr-core` / `cldr-dates-full` / `cldr-numbers-full` / `cldr-units-full` and other npm packages.
- [ECMA-402 §6 — Locale and Currency Identifiers](https://tc39.es/ecma402/#locale-and-currency-identifiers) — Identifier specification.

### Reference implementations

- `.references/formatjs/package.json` -- `cldr-*: 48.1.0` Lock evidence.
- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` —— CLDR JSON import list and extraction logic.
- `.references/formatjs/packages/intl-numberformat/scripts/extract-currencies.ts` —— currency extraction.
- `.references/formatjs/knowledge-base/001-repo-layout.md` -- `@formatjs_generated/*` package architecture.
- `.references/intl/internal/cldr/data.go` - 400 KB / 3203 lines Go literal precedent (translate-agent/intl).
- `.references/intl/internal/gen/main.go` - text/template template generator precedent.

### Cross-SPEC

- [SPEC 00 §5.3 — Data strategy](./00-vision-and-scope.md#53-data-strategy)
- [SPEC 10 §Maximize / Minimize](./10-locale.md) - Consumes `MaximizeSubtags` / `MinimizeSubtags`.
- [SPEC 11 §BestFitMatcher](./11-locale-matching.md) - Active matcher receives supported locales and maximizers from formatter constructors; generated `MatchingDistance` / `ParadigmLocales` / `MatchVariables` remain reference data for a future explicit matcher revision.
- [SPEC 20 §Currency Data](./20-numberformat.md) - consumes `CurrencyDigits` / `CurrencyDisplayName`.
- [SPEC 30 §DateTimeFormat Core](./30-datetimeformat.md) - Consumes `CalendarNames` / `DateFormat` / `TimeFormat` / `AvailableFormats` / `IntervalFormats`.
- [SPEC 31 §Skeleton Resolution](./31-datetimeformat-skeleton.md) - Consumes `AvailableFormats` / `IntervalFormats`.
- [SPEC 32 §TimeZone & Calendar Data](./32-datetimeformat-tz.md) - consumes `metazones.go`, integrated with `time/tzdata`.
- [SPEC 40 §PluralRules](./40-pluralrules.md) - `plurals.go` and `plural/*.go` are output to this directory by the SPEC 40 codegen.

---

> This SPEC is the maintenance record of the underlying CLDR data. Version nail changes trigger SPEC revision; active scope locale inventory changes are maintained through `tools/gen-cldr/locale_list.go` (do not trigger SPEC revision), and volume changes are assisted through `task build:size` review.
