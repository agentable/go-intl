# SPEC 00 — Vision, Scope, and Architecture

> **Status:** Active contract (2026-05-20)
> **Audience:** Maintainers and contributors of `go-intl` and its primary consumers (`messageformat-go`, `go-test`, and JS host integrations).
> **Authority:** ECMA-402 is the behavioral authority. This document is the project-level memory for what `go-intl` is trying to be; focused SPECS (`10-locale.md`, `20-numberformat.md`, …) own individual surfaces and must change when ECMA-402 or the correctness target proves them too narrow.

---

## 1. Vision

`go-intl` is a Go implementation of the **ECMA-402 Internationalization API**. It exposes the same active surface that JavaScript developers know as the `Intl` namespace plus `Intl.Locale`, `Intl.NumberFormat`, `Intl.DateTimeFormat`, `Intl.PluralRules`, `Intl.ListFormat`, `Intl.RelativeTimeFormat`, `Intl.DurationFormat`, `Intl.DisplayNames`, `Intl.Collator`, and `Intl.Segmenter`. Public API shape, option names, option values, resolved options, parts, range sources, and error boundaries are governed by ECMA-402 first; FormatJS is the readable polyfill reference and `.references/node/` is the native-engine behavior reference.

The library exists because the Go ecosystem has no equivalent today:

- Go stdlib provides Unicode primitives, but no locale-aware formatters.
- `golang.org/x/text` provides Unicode/CLDR building blocks (`language.Tag`, `message`, `number`, `collate`, `feature/plural`) but is not ECMA-402, has gaps (no `DateTimeFormat`-equivalent, no resolved-options model, no `formatToParts`), and does not aim for ECMA-402 output parity.
- Existing Go libraries (`translate-agent/intl`, `messageformat-go`) cover slices of the surface and would otherwise each reinvent the same locale-aware primitives.

**The mission:** give Agentable Go libraries — and any Go consumer who wants native JavaScript `Intl` semantics in Go — one shared, ECMA-402-faithful, CLDR-driven formatting layer.

The design posture is restraint first: the public API should feel inevitable, not configurable. Implementation complexity stays private; users get one obvious construction path, truthful supported sets, precise errors, and examples whose shape teaches the model without extra ceremony.

### 1.1 Non-goals

- **Not a loose Go reinterpretation.** We mirror ECMA-402's public concepts and observable semantics. Go may use typed `Options`, `time.Time`, and returned `error` values, but only as type-system bridges for the same constructors, methods, options, resolved options, and error conditions.
- **Not a superset of `golang.org/x/text`.** We *use* `language.Tag` and CLDR helpers from `x/text` where they fit; we do not re-export the rest of `x/text` under our name.
- **Not an ICU port.** We do not link against ICU C/C++ libraries. CLDR data ships embedded as Go source.
- **Not a translation system.** Message localization (string catalogs, plural-aware templates) is `messageformat-go`'s job. `go-intl` provides primitives that `messageformat-go` calls.

---

## 2. Reference Implementation Policy

The authoritative reference is **ECMA-402**, vendored at `.references/ecma402/spec/`. FormatJS and other vendored references are subordinate implementation guides.

### 2.1 Authority model

SPECS are durable project memory, not a higher authority than ECMA-402. They describe the target, the current implementation tier, and the acceptance gates. They must not preserve a local behavior merely because existing code or historical tests already do it.

Support tiers:

| Tier | Meaning | Requirement |
|------|---------|-------------|
| Complete | The surface matches ECMA-402 observable behavior for the advertised data/locale set. | Covered by conformance fixtures and normal tests. |
| Narrowed implementation gap | The surface intentionally refuses or withholds unsupported behavior to avoid false support. | Owning SPEC must name current behavior, rationale, `review_after`, and removal path. |
| Accepted divergence | A generated or native fixture exists and go-intl intentionally differs from the reference for an implementation-defined or data-version reason. | Must be in `testdata/divergences.md` with owner, reason, review date, and removal path. |

> **Why**: The cleanest API is the one that tells the truth. A narrowed surface is acceptable only when it prevents false correctness and has an exit path; it is not a permanent product philosophy.
>
> **Rejected**: Treating local SPECS as immutable once implemented. That turns documentation into back-compat pressure and lets accidental limitations masquerade as design.

| Subject | Reference path | Use |
|---------|----------------|-----|
| Intl namespace | `.references/ecma402/spec/intl.html` | Root package contract: `Intl` is not a constructor, constructor properties, `getCanonicalLocales`, `supportedValuesOf` |
| Locale and locale negotiation | `.references/ecma402/spec/locale.html`, `.references/ecma402/spec/negotiation.html` | `Intl.Locale`, `CanonicalizeLocaleList`, `ResolveOptions`, `ResolveLocale`, `FilterLocales`, `localeMatcher` |
| `Intl.NumberFormat` | `.references/ecma402/spec/numberformat.html` | Constructor, options, resolved options, parts, ranges, digit rounding, `FormatNumericToString` |
| `Intl.DateTimeFormat` | `.references/ecma402/spec/datetimeformat.html` | Constructor, date/time options, system time zone default, parts, ranges |
| `Intl.PluralRules` | `.references/ecma402/spec/pluralrules.html` | Constructor, digit options, `select`, `selectRange`, plural category resolution |
| `Intl.ListFormat` | `.references/ecma402/spec/listformat.html` | Constructor, `type` / `style`, list pattern partitioning, `format`, `formatToParts` |
| `Intl.RelativeTimeFormat` | `.references/ecma402/spec/relativetimeformat.html` | Constructor, `style` / `numeric`, unit validation, numeric auto literals, parts |
| `Intl.DurationFormat` | `.references/ecma402/spec/durationformat.html` | Constructor, unit options, digital formatting, duration validation, parts |

**FormatJS implementation references:**

| Subject | Reference path | Use |
|---------|----------------|-----|
| Abstract operations | `.references/formatjs/packages/ecma402-abstract/` | Behavior reference for production-used algorithms (`PartitionPattern`, digit rounding, skeleton matching, plural operands, …) |
| Top-level helpers | `.references/formatjs/packages/intl/` | Non-normative helper reference only; it must not override ECMA-402 `Intl` namespace shape |
| `Intl.Locale` | `.references/formatjs/packages/intl-locale/` | Locale parsing, canonicalization, `maximize`/`minimize`, preference data |
| `Intl.NumberFormat` | `.references/formatjs/packages/intl-numberformat/` | Decimal, percent, currency, unit, scientific, compact notation |
| `Intl.DateTimeFormat` | `.references/formatjs/packages/intl-datetimeformat/` | Date/time formatting, `formatRange`, time-zone handling |
| `Intl.PluralRules` | `.references/formatjs/packages/intl-pluralrules/` | Cardinal, ordinal, `select`, `selectRange` |
| `Intl.ListFormat` | `.references/formatjs/packages/intl-listformat/` | List pattern selection, aliases, supported locales |
| `Intl.RelativeTimeFormat` | `.references/formatjs/packages/intl-relativetimeformat/` | Relative field lookup, numeric auto behavior, supported locales |
| `Intl.DurationFormat` | `.references/formatjs/packages/intl-durationformat/` | Duration unit option resolution, partitioning, digital separators, supported locales |

**Secondary references:**

| Subject | Path | Use |
|---------|------|-----|
| Native Node/V8 behavior | `.references/node/` | ICU-backed tiebreaker for implementation-defined output, edge cases, time zones, and Node Intl snapshots |
| ECMA-402 scope check | `.references/ext/` (PHP `ext/intl`) | Validates the size and shape of a full ECMA-402 surface |
| CLDR-driven Go pattern | `.references/intl/` (`translate-agent/intl`) | A working Go example of embedding CLDR data and using `language.Tag` |

### 2.2 Reference hygiene

Only reference trees with compatible root licensing and direct project value belong under `.references/`. Clean out references whose root project license is GPL, AGPL, or LGPL; do not remove MIT/BSD/ECMA-compatible references merely because their trees contain third-party license notices. `.references/node/` remains the Node/V8 reference source; do not copy GPL-licensed vendored files from it into go-intl source or fixtures.

**Authority rules:**

1. **ECMA-402 is the primary reference.** If `.references/ecma402/spec/` conflicts with a local SPEC, README example, existing test, or FormatJS helper API, update the local artifact first.
2. **Local SPECS are subordinate memory.** They may constrain implementation order and support tiers, but they must not redefine ECMA-402 semantics.
3. **Implementation gaps need an exit.** Any retained gap must state current behavior, rationale, `review_after`, and the concrete removal path in the owning SPEC.
4. **FormatJS is the primary readable implementation reference** because it is TypeScript and a faithful spec polyfill.
5. **`.references/node/` breaks native-engine ties** when ECMA-402 leaves behavior implementation-defined or when Node/V8/ICU output is the observable compatibility target.
6. **CLDR is the data oracle.** When tables disagree, trust the CLDR version pinned in `internal/cldr/VERSION` and document the divergence.
7. **No local convenience beats native ownership.** Go typed bridges are allowed, but helper shape, cache knobs, or historical ergonomics must not override the native `Intl` owner model.
8. **No import-cost shortcut beats the `Intl` namespace.** The root package represents ECMA-402 `Intl`; active constructor aliases are the current Go bridge for constructor properties such as `Intl.NumberFormat`. Measure and document aggregate facade cost separately instead of deleting constructor properties to make dependency reports smaller.

### 2.3 Test fixture policy

We pull **language-agnostic input/output pairs** from three sources and run them through one Go harness per package that asserts byte-equality:

| Source | Path | Format | Used for |
|--------|------|--------|----------|
| `FormatJS` polyfill tests | `.references/formatjs/packages/<polyfill>/tests/*.test.ts` + locale-data JSON | TypeScript (Vitest) `describe`/`it` blocks with inline expectations | Primary conformance — every public formatter must pass these |
| Node localization snapshots | `.references/node/` | JSON snapshots extracted from Node Intl behavior | Cross-validation for native V8/ICU output near the pinned CLDR/ICU version |

**Porting flow:**

1. Extract the assertion table into JSON under `<package>/testdata/conformance/<source>/<file>.json`.
2. Record locale, options, input, and expected output.
3. Run through a Go harness (`<package>/conformance_unified_test.go`) that asserts byte-equality.

**Divergence handling:**

When a generated or native fixture exists and go-intl knowingly differs, record it in `<package>/testdata/divergences.md` with: the source, the case, owner, our output, the reference output, reason, `review_after`, and removal path. A missing implementation is not a divergence by itself; it belongs in the owning SPEC as a narrowed implementation gap until the fixture can be enabled.

### 2.4 Specification ownership map

Each concept has exactly one owner. Cross-links may explain dependencies, but they must not restate another spec's rules.

| Spec | Owns |
|------|------|
| [`10-locale.md`](./10-locale.md) | `locale.Locale`, BCP 47 parsing, Unicode extension fields, maximize/minimize, locale getters |
| [`11-locale-matching.md`](./11-locale-matching.md) | lookup and best-fit matching, `ResolveLocale`, supported-locales filtering |
| [`12-abstract-operations.md`](./12-abstract-operations.md) | ECMA-402 option-shape abstract operations, internal slot conventions, root structured error details, abstract error boundaries |
| [`20-numberformat.md`](./20-numberformat.md) | `numberformat` public API, option resolution, format parts, ranges, compact/scientific/unit/currency behavior |
| [`21-number-math.md`](./21-number-math.md) | decimal backend, `ToIntlMathematicalValue`, rounding modes, rounding increments, mathematical value bridge |
| [`30-datetimeformat.md`](./30-datetimeformat.md) | `datetimeformat` public API, resolved options, date/time parts, range behavior |
| [`31-datetimeformat-skeleton.md`](./31-datetimeformat-skeleton.md) | skeleton parsing, best-fit format matching, pattern scoring |
| [`32-datetimeformat-tz.md`](./32-datetimeformat-tz.md) | time-zone resolution, canonical links, Gregorian calendar names, metazone display data |
| [`40-pluralrules.md`](./40-pluralrules.md) | `pluralrules` public API, operands, CLDR plural rule codegen, `selectRange` |
| [`41-listformat.md`](./41-listformat.md) | `listformat` public API, list pattern data, format parts, supported locales |
| [`42-relativetimeformat.md`](./42-relativetimeformat.md) | `relativetimeformat` public API, relative field data, NumberFormat/PluralRules composition |
| [`43-durationformat.md`](./43-durationformat.md) | `durationformat` public API, duration records, unit options, NumberFormat/ListFormat composition |
| [`44-displaynames.md`](./44-displaynames.md) | `displaynames` public API, type/style/fallback semantics, CLDR localenames data |
| [`45-collator.md`](./45-collator.md) | `collator` public API, x/text/collate binding, sensitivity/caseFirst/numeric mapping |
| [`46-segmenter.md`](./46-segmenter.md) | `segmenter` public API, UAX #29 binding via `uniseg`, byte-offset bridge |
| [`50-cldr-data.md`](./50-cldr-data.md) | CLDR version pins, generated data layout, generator architecture, runtime data access API |
| [`60-facade.md`](./60-facade.md) | root `go-intl` namespace, static common Intl functions, active constructor aliases, forbidden one-shot helpers |
| [`61-messageformat-integration.md`](./61-messageformat-integration.md) | `messageformat-go` dependency direction and formatter adapter contract |
| [`70-conformance.md`](./70-conformance.md) | fixture format, fixture sources, divergences, conformance gates |
| [`71-benchmark.md`](./71-benchmark.md) | benchmark layout, non-blocking performance telemetry, benchstat workflow |
| [`72-operation-ledger.md`](./72-operation-ledger.md) | public surface to ECMA-402 owner, implementation, and verification ledger |
| [`73-json-records.md`](./73-json-records.md) | JSON field names and presence policy for resolved options, parts, segment records, and locale info |

> **Why**: The project has enough surface area that "read SPEC 00" is no longer precise enough. The map routes every design question to one owner and prevents duplicate mini-specs in CLAUDE.md, README.md, tests, or source comments.
>
> **Rejected**: A generated index or test-enforced spec layout. The spec set is small and intentional; forcing layout through code would turn documentation discipline into runtime-adjacent machinery.

---

## 3. Public Surface

The maintained surface is the minimum viable API needed by the primary consumers:

| Package | Type / function | Mirrors |
|---------|-----------------|---------|
| `github.com/agentable/go-intl/locale` | `Locale`, `Parse`, `MustParse`, `New`, `(Locale).Maximize`, `(Locale).Minimize`, locale info getters | `Intl.Locale` |
| `github.com/agentable/go-intl/numberformat` | `NumberFormat`, `New`, `SupportedLocalesOf`, `Value` constructors, `Format`, parts/range methods, `.ResolvedOptions` | `Intl.NumberFormat` |
| `github.com/agentable/go-intl/datetimeformat` | `DateTimeFormat`, `New`, `SupportedLocalesOf`, `(*DateTimeFormat).Format`, `.FormatToParts`, `.FormatRange`, `.FormatRangeToParts`, `.ResolvedOptions` | `Intl.DateTimeFormat` |
| `github.com/agentable/go-intl/pluralrules` | `PluralRules`, `New`, `SupportedLocalesOf`, `Value` constructors, `Select`, `SelectRange`, `.ResolvedOptions` | `Intl.PluralRules` |
| `github.com/agentable/go-intl/listformat` | `ListFormat`, `New`, `SupportedLocalesOf`, `.Format`, `.FormatToParts`, `.ResolvedOptions` | `Intl.ListFormat` |
| `github.com/agentable/go-intl/relativetimeformat` | `RelativeTimeFormat`, `New`, `SupportedLocalesOf`, typed `Format*`, typed parts methods, `.ResolvedOptions` | `Intl.RelativeTimeFormat` |
| `github.com/agentable/go-intl/durationformat` | `DurationFormat`, `New`, `SupportedLocalesOf`, `.Format`, `.FormatToParts`, `.ResolvedOptions` | `Intl.DurationFormat` |
| `github.com/agentable/go-intl/displaynames` | `DisplayNames`, `New`, `SupportedLocalesOf`, `.Of`, `.ResolvedOptions` | `Intl.DisplayNames` |
| `github.com/agentable/go-intl/collator` | `Collator`, `New`, `SupportedLocalesOf`, `.Compare`, `.ResolvedOptions` | `Intl.Collator` |
| `github.com/agentable/go-intl/segmenter` | `Segmenter`, `New`, `SupportedLocalesOf`, `.Segment`, `.ResolvedOptions` | `Intl.Segmenter` |
| `github.com/agentable/go-intl` (root) | `GetCanonicalLocales`, root supported-value accessors, active constructor type aliases, `ErrorKind`, `Error`, root category sentinels | `Intl` namespace object, constructor-property bridge, plus Go error bridge for ECMA-402 `RangeError` / `TypeError`-equivalent failures |

The root package mirrors the JavaScript `Intl` namespace object as closely as Go allows. JavaScript `Intl` is not a constructor and has no per-locale instance state; therefore root `New`, root typed one-shot helpers, and root cache controls are outside the long-term public surface. JavaScript `Intl` also exposes constructor properties; root type aliases are the current Go bridge for that shape and must not be removed merely to reduce aggregate root import cost. Detailed formatter options live in their formatter packages.

Every exported API must have an identified native JavaScript owner before it is added. Go-only typed bridges are allowed only when they preserve a native operation's semantics while replacing JavaScript dynamic values with Go types; they must not widen the observable Intl surface.

Public ECMA-402 record structs (`ResolvedOptions`, parts, range parts, locale info, segments, and duration records) must marshal to JSON with the same camelCase field names that the corresponding JavaScript record exposes. Omitted JavaScript properties are represented by nil pointer fields plus `omitempty`, not by ambiguous zero values. This JSON shape is a host-boundary bridge, not permission to replace typed Go APIs with `map[string]any`.

### 3.1 Outside the active surface

All ten ECMA-402 active constructors are implemented. Future ECMA-402 additions (any new `Intl.*` constructor that ships in a later edition) join the active surface only after a new SPEC plus package implementation lands.

---

## 4. Locale Model

### 4.1 Decision

`go-intl` defines its own immutable `Locale` type that stores a `golang.org/x/text/language.Tag` plus ECMA-402 locale extension state. We do **not** alias `language.Tag` directly because ECMA-402 carries data that BCP 47 does not surface as first-class properties (`calendar`, `collation`, `hourCycle`, `caseFirst`, `numeric`, `numberingSystem`, week info).

```go
// Conceptual shape; final field set defined in SPEC 10.
type Locale struct {
    tag language.Tag
    ext localeExtensions
}
```

### 4.2 Rationale

- `language.Tag` already implements BCP 47 parsing, canonicalization, and matcher infrastructure used across the Go ecosystem (`x/text/message`, `x/text/currency`, `messageformat-go`'s plural module, `translate-agent/intl`). Reusing it preserves interop and avoids re-parsing locale strings at every boundary.
- ECMA-402 exposes these values through read-only `Intl.Locale.prototype` accessors (`baseName`, `calendar`, `collation`, `hourCycle`, `caseFirst`, `numeric`, `numberingSystem`, etc.). Go must provide methods with the same meaning, not exported mutable fields.
- A value type keeps `Locale` cheap to pass while preventing callers from constructing invalid internal state through struct literals.

### 4.3 String round-trip

`(Locale).String()` returns the canonical BCP 47 representation including Unicode `-u-` extensions for whichever fields are set. `Parse(s string) (Locale, error)` is the inverse and accepts both raw BCP 47 tags and tags with Unicode extensions.

---

## 5. Architecture

### 5.1 Package layout

```
go-intl/
├── intl.go                 # Top-level Intl namespace: GetCanonicalLocales and active aliases
├── supported.go            # Root Intl.supportedValuesOf typed accessors
├── doc.go                  # Package overview
├── errors.go               # Root namespace structured errors and category sentinels
├── locale/                 # Intl.Locale
├── numberformat/           # Intl.NumberFormat
├── datetimeformat/         # Intl.DateTimeFormat
├── pluralrules/            # Intl.PluralRules
├── listformat/             # Intl.ListFormat
├── relativetimeformat/     # Intl.RelativeTimeFormat
├── durationformat/         # Intl.DurationFormat
├── displaynames/           # Intl.DisplayNames
├── collator/               # Intl.Collator
├── segmenter/              # Intl.Segmenter
└── internal/
    ├── ecma402/            # Production-used abstract operations, constructor-locale wrapper, validators
    ├── ecma402/numberformat/    # digit option resolution and FormatNumericToString
    ├── ecma402/datetimeformat/  # skeleton parser and DateTimeFormat pattern matcher
    ├── ecma402/pluralrules/     # operands and plural categories
    ├── localematcher/      # ResolveLocale, lookup, best-fit, supported-locale filtering
    ├── cldrmatch/          # formatter-family data-locale resolution over generated CLDR subsets
    ├── decimal/            # apd-backed Decimal for ToIntlMathematicalValue
    ├── intlerr/            # Cycle-free implementation backing root error aliases
    ├── collation/          # Collator backend capability metadata
    ├── cldr/               # Generated CLDR data split by domain package
    │   ├── codec/          # generated-blob decoding helpers
    │   ├── locale/         # locale kernel, likely subtags, preferences, profile gates
    │   ├── number/         # number symbols, compact tables, numbering-system locale data
    │   ├── date/           # Gregorian date names, skeleton/range patterns, calendars
    │   ├── currency/       # currency digits and display data
    │   ├── unit/           # unit pattern payloads
    │   ├── list/           # list pattern payloads
    │   ├── relativetime/   # relative-time field payloads
    │   ├── displaynames/   # locale/currency/language/region/script display names
    │   ├── plural/         # supported plural locales
    │   └── timezone/       # generated CLDR time-zone display/link data
    └── tz/                 # IANA timezone data + offset arithmetic (DateTimeFormat dep)
```

### 5.2 Layering rules

1. **Formatter packages depend on `internal/*`; `internal/*` never depends on public formatter packages.** Structured error construction lives in `internal/intlerr` so formatter packages can build root-category errors without importing the root facade and creating a cycle.
2. **`internal/cldr/<domain>` packages are the only places that touch generated data tables.** Every other package goes through narrow domain accessors. The retired root `internal/cldr` package is not a data owner.
3. **`internal/ecma402` keeps production-used abstract algorithms.** Go does not expose arbitrary JavaScript objects as options, but typed `Options` must preserve ECMA-402 option coercion, defaulting, validation, and resolved-state semantics.
4. **The root namespace is thin and native-shaped.** It exposes ECMA-402 common namespace functions, active constructor aliases for constructor-property parity, and root error bridges. It contains no formatting logic of its own.

### 5.3 Data strategy

CLDR data is **compiled into Go source** via a generator under `tools/gen-cldr/`. The runtime reads generated Go literals from domain packages under `internal/cldr/<domain>/`; it does not embed CLDR JSON, parse JSON at startup, or perform file I/O on formatting paths. The durable layout is SPEC 50's per-domain package model: each formatter imports only the generated payload families it can observe, and the retired root `internal/cldr` package must not return as a data owner.

Decisions:

- **No runtime JSON parsing.** All decoding happens at generation time.
- **Single CLDR profile, honest supported sets.** `tools/locale-profile.json` lists the generated CLDR payload target for number, date, plural, list, relative time, duration, display names, units, currency, and time-zone display. CLDR-backed constructors derive `SupportedLocalesOf` from the generated payloads they consume; non-CLDR engines such as `collator` and `segmenter` own narrower engine-specific supported sets. No constructor may advertise a locale before its backing data or boundary algorithm can support it. See SPEC 50 §1.3 and SPEC 46 §5.
- **CLDR version is pinned.** The pinned version lives in `internal/cldr/VERSION` and is referenced by the generator. Changing it is a SPEC-affecting decision.

### 5.4 Time-zone strategy

`Intl.DateTimeFormat` requires IANA time-zone data and DST offset arithmetic:

- `internal/tz/tzdata.go` blank-imports `time/tzdata` so `time.LoadLocation` has deterministic IANA data even on minimal deploy images.
- The pinned tzdata version is recorded alongside CLDR / ICU in `internal/cldr/VERSION`.
- Canonical-name resolution (`US/Eastern` → `America/New_York`) goes through generated CLDR canonical-link tables.

---

## 6. Consumer Contract

### 6.1 `messageformat-go` (primary consumer)

`messageformat-go`'s `pkg/functions/` package today implements its own number/date/currency/unit formatting. The intended migration is:

1. `messageformat-go` imports `go-intl` and rewrites its formatter-related built-in functions (`:integer`, `:number`, `:currency`, `:percent`, `:unit`, `:offset`, `:date`, `:datetime`, `:time`) as adapters that delegate to `go-intl`.
2. `MessageFunctionContext.Locales()` (a `[]string`) is converted to `locale.List` via `locale.Parse`.

**API stability requirement:** any change to `go-intl`'s formatter constructors or `ResolvedOptions` shape after `messageformat-go` integrates is a breaking change and requires a major version bump.

### 6.2 `go-test` (secondary consumer)

`go-test` already pulls `messageformat-go` transitively. Direct use of `go-intl` is optional and limited to locale-aware display in assertion output.

### 6.3 `go-humanize` (non-consumer, by design)

`go-humanize` is deliberately English-only and will not adopt `go-intl`. We preserve room for a future `humanize/i18n` shim, but it is not an active deliverable.

### 6.4 JS host integrations

If a JS host integration exposes `globalThis.Intl`, `go-intl` is the backing implementation. The active API must remain expressible as a JS host binding while preserving typed Go entrypoints: record structs marshal with ECMA-402 field names, constructor errors expose `*gointl.Error`, supported sets refuse unsupported tailoring honestly, and range methods preserve caller-provided order even when `start > end`.

The maintained host consumer profile covers supported-locale exclusions, root supported-value include/exclude boundaries, and reversed range behavior for NumberFormat, DateTimeFormat, and PluralRules. Changes that break that profile are public contract changes and must update the owning SPEC before code or fixtures move.

---

## 7. Active Boundaries

`go-intl` grows only when a real consumer or ECMA-402 correctness gap justifies the next step. The active constructor surface is exactly the ten implemented ECMA-402 constructors plus the root `Intl` namespace:

- `locale`
- `numberformat`
- `datetimeformat`
- `pluralrules`
- `listformat`
- `relativetimeformat`
- `durationformat`
- `displaynames`
- `collator`
- `segmenter`
- root `go-intl`

Until a new ECMA-402 edition expands the surface:

- new Intl families require their own SPEC before implementation;
- optimization work must preserve the public API and byte-equal output;
- data-size work must keep runtime CLDR data embedded in generated Go source.

---

## 8. Design Decisions

The formerly open questions are resolved by focused SPECS:

1. **Decimal type** — decided in `SPEC 21 — Number Math & Decimal`.
2. **Functional options vs. config struct** — decided in `SPEC 20 — NumberFormat`.
3. **`Intl.Locale.weekInfo` / `calendars` / related getters** — decided in `SPEC 10 — Locale`.
4. **CLDR / ICU / tzdata pins** — recorded in `internal/cldr/VERSION` and governed by `SPEC 50 — CLDR Data & Codegen`.
5. **Best-fit matcher** — decided in `SPEC 11 — Locale Matching`.

---

## 9. Definition of Done for this SPEC

- [x] Vision, scope, and non-goals stated.
- [x] Reference-implementation policy and test-fixture policy defined.
- [x] Public surface enumerated.
- [x] Locale model decided (wrap `language.Tag`).
- [x] Package layout and layering rules defined.
- [x] Consumer contract recorded for `messageformat-go`, `go-test`, `go-humanize`, and JS host integrations.
- [x] Active boundaries stated.
- [x] ECMA-402-over-SPECS authority model stated.
- [x] Design decisions linked to focused SPECS.
