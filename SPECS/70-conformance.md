# SPEC 70 — Conformance Test Strategy

> **Status:** Revised (2026-08-09)
> **Type:** Flow + Schema + Rule — defines how go-intl proves ECMA-402 observable behavior through generated-reference and native fixtures.
> **Authority:** ECMA-402 and native/reference fixtures define observable behavior. This spec records the conformance fixture format, fixture sources, divergence handling, correctness gates, and XFAIL discipline. Per-formatter SPECS (10/20/30/40/41/42/43/44) record their own option semantics; this spec records how those semantics are *verified* against the reference implementations.

---

## Overview

The "correctness" of go-intl active scope is defined by the **ECMA-402 specification + fixture-driven conformance tests**: the specification determines the semantic boundaries, fixtures mechanically or manually extract input/output pairs from reference tests or native Intl snapshots, and assert them one by one within the Go-side harness. This SPEC covers:

1. **fixture format record**: unified JSON schema, universal across formatters.
2. **fixture source**: generated-reference fixtures, native Intl smoke cases, and a small number of handwritten ECMA-402 edge use cases.
3. **divergence process**: known deviations are registered in `<package>/testdata/divergences.md`; silent skips are **FORBIDDEN**.
4. **CI gates + telemetry**: conformance tests (blocking), conformance audit (blocking), and performance telemetry (non-blocking report).
5. **XFAIL aging**: Each XFAIL must have an expiration date, and it will automatically fail when it expires.

> **Why**: ECMA-402 one-to-one alignment is the core commitment of SPECS/00 §1; fixtures are the execution mechanism for that commitment.
> **Rejected**: Relying only on handwritten unit tests to cover ECMA-402 behavior. The reference fixture corpus is broader than local handwritten tests and gives better drift detection.

Conformance policy follows SPEC 00's authority model: SPECS explain and gate implementation work, but fixtures prove observable ECMA-402 behavior. A missing implementation is an **implementation gap** in the owning SPEC, not an accepted divergence and not a reason to delete or weaken fixtures.

---

## 1. Fixture Format Record

### 1.1 JSON Schema

Each fixture **MUST** conform to the following schema (JSON object):

```json
{
  "id": "intl-relativetimeformat/en/second-past",
  "source": "formatjs:packages/intl-relativetimeformat/tests/index.test.ts",
  "locale": "en",
  "options": {"numeric": "always"},
  "input": {"value": -1, "unit": "second"},
  "expected": "1 second ago",
  "expectedParts": [{"type": "integer", "value": "1", "unit": "second"}, {"type": "literal", "value": " second ago"}]
}
```

### 1.2 Field Contract

| Field | Required | Type | Meaning |
|------|------|------|------|
| `id` | Required | string | Globally unique; generated FormatJS IDs use a readable source prefix and semantic fingerprint; manual and Node IDs retain their curated identities; ledgers reference this id |
| `source` | Required | string | `<source>:<path>` —— `formatjs:` or `manual:` |
| `locale` | Required | string | BCP 47 locale tag, the Go end uses `locale.Parse(string)` to parse at the fixture boundary |
| `options` | Required | object | ECMA-402 option; key name spec original text (camelCase) |
| `input` | Required | number / string / array / object / null | Numeric literal (`null` is the NumberFormat NaN bridge), string list (ListFormat), ISO-8601 string (DateTimeFormat), `{start, end}` object (Range), `{value, unit}` object (RelativeTimeFormat), duration record object (DurationFormat) |
| `expected` | optional | string | `Format` output; FormatToParts-only fixture can be omitted |
| `expectedOk` | Optional | boolean | DisplayNames `Of` result presence; false is an observation |
| `expectedLocales` | Optional | array | `SupportedLocalesOf` output; omitting means no verification supported locales |
| `expectedParts` | Optional | array | `FormatToParts` output; RelativeTimeFormat number parts can include `unit`; omitting means not checking parts |
| `expectedRange` | optional | string | `FormatRange` output |
| `expectedRangeParts` | optional | array | `FormatRangeToParts` output |
| `expectedResolvedOptions` | Optional | object | Complete resolved-options JSON record; property presence, JSON types, and values are compared |
| `errorCode` | Optional | string | Error case (replacing `expected`), corresponding to go-intl sentinel name |

**Rules**:

1. The fixture file **MUST** be a JSON array (multiple fixtures per file), with the path `<package>/testdata/conformance/<source>/<file>.json`.
2. `id` **MUST** be globally unique. `LoadFixtures` rejects duplicates within a package before ledger indexes or callbacks, identifying the ID and both source paths. `ValidateFixtureRoots` additionally rejects IDs shared across packages.
3. When `input` is a date, you must use ISO-8601 with timezone offset string (such as `"2020-01-01T00:00:00Z"`), and Go harness deserializes it through `time.Parse(time.RFC3339, ...)`. **FORBIDDEN** to use Unix epoch ms numbers; they are too ambiguous at the Go fixture boundary.
4. The error fixture **MUST** go through `errorCode` field, which is kept separately in `<package>/testdata/conformance/<source>/errors.json` and separated into files with the success fixture; Go harness goes through `errors.Is(err, sentinel)` verification.
5. `errors.json` **can only** contain error fixtures with `errorCode`; forward fixtures **are prohibited** from being put into `errors.json`.
6. The source directory where the fixture is located **MUST** be consistent with the `source` field prefix: `manual/` for `manual` or `manual:*`, `formatjs/` for `formatjs:*`, `node-*` for `node:*`.
7. The `options` field **MUST** maintain the original ECMA-402 spec naming (`maximumFractionDigits` instead of `MaximumFractionDigits`); the Go-side harness is mapped to a typed `Options` value when loading.
   Locale fixtures pass their input tag and typed options to `locale.New`
   before observing canonicalization, maximize/minimize, or locale info.
   Omitted pointer fields remain absent; explicit false and empty strings are
   preserved. Constructor option failures use `invalid_option`, while invalid
   input tags use `invalid_value`.
8. **It is prohibited** to embed JS functions, callbacks, and Date literals in fixtures; parts that cannot be mechanically extracted are classified according to the SPEC §2.4 process.
9. Loading rejects unknown record fields, non-object options, non-object
   resolved expectations, and records without an expectation or `errorCode`.
   Typed formatter input/options validation remains in the package adapter.
10. Omission means no observation. Empty text, false, empty expected arrays,
    and an empty resolved object remain declared observations through JSON
    encoding, loading, coverage, and native-witness validation. Explicit null
    is invalid for record fields other than the `input` bridge; it does not
    stand for an empty array or an omitted expectation.
11. Fixture shape errors fail the complete suite before any callback runs.
    Diagnostics identify the source file and field or JSON location; semantic
    validation also identifies the fixture when its ID is available.
12. Within the operation's valid input family, adapters execute their declared
    text, parts, range text, and range parts observations independently, without
    an unrelated text prerequisite. NumberFormat, DateTimeFormat, DurationFormat,
    DisplayNames, PluralRules, ListFormat, and RelativeTimeFormat also execute
    resolved-options snapshots independently; their resolved-only records do
    not parse a formatting input. Locale has no resolvedOptions operation.
    Empty expected arrays are asserted rather than skipped.
13. Resolved-options snapshots describe the complete object. Missing, extra,
    or unknown properties fail comparison; null differs from an absent
    property, and false, zero, and empty strings retain their JSON values.
    Object property order does not affect equality.

> **Why**: The unified schema is universal across formatters and lets the shared harness validate every active surface without formatter-specific fixture loaders.
> **Rejected**: Each formatter custom schema (NumberFormat `"value"` vs DateTimeFormat `"date"`) - 4 sets of harness, 4 sets of loaders, DRY violation.

### 1.3 Manifest

`<package>/testdata/conformance/<source>/MANIFEST.json` **SHOULD** record for each fixture file:

| Field | Meaning |
|------|------|
| `extractor_version` | Version tag of `tools/gen-fixtures-from-formatjs/` |
| `extracted_from` | Reference revision + test file path |
| `extracted_at` | ISO 8601 timestamp |
| `count` | Number of fixtures in this file |

**SHOULD** (not mandatory for active scope): manifest is missing by PR-CI lint verification. active scope can be replaced by git blame; consumer-driven expansion trigger condition = fixture PR number > 5/month.

> **Why**: The manifest is an audit trail - when a reference upgrade causes fixture drift, the manifest allows the reviewer to locate which fixture batch changed.
> **Rejected**: Mandatory manifests in active scope. Git history is sufficient for the first release, so a required manifest would be YAGNI until fixture volume or review workflow demands it.

---

## 2. Fixture Sources

### 2.1 Fixture source matrix

| Source | role | active scope requirement |
|--------|------|-------------|
| **Generated reference** | Main fixture source | Generated fixtures must pass; ungenerated test sources must enter `.skip-list.json` audit. The current on-disk source prefix and directory names are historical implementation details, not product authority. |
| **Native Intl** | Native-engine contract source | Each active constructor retains at least one native snapshot and one native constructor error/refusal fixture. Native fixtures also own observable option, parts, locale, and backend-capability boundaries that are not safely covered by generated extraction. |
| **Manual ECMA-402 edge cases** | Supplementary fixture source | Only for boundaries that cannot be mechanically extracted or are explicitly required by spec |

**Rules**:

1. The main generated fixture lane **MUST** be translated through the generated-fixture extractor; the normative interpretation rights still belong to the ECMA-402 spec.
2. Native fixtures **MUST** record the runtime major version. The current fixture `source` prefix is historical; it identifies the witness lane and version, not an external product authority.
3. Manual fixture **MUST** indicate that it corresponds to the ECMA-402 section or local SPEC rationale, and is marked with `source: "manual:<topic>"`.
4. Fixture sources **MUST** be marked via the `source` field; mixing sources into the same JSON file is prohibited.
5. `tools/.gen-versions` labels the generated-reference source as `formatjs=local-reference`; it does not pin a revision. Git's submodule commit identifies the repository's reference checkout, while the extractor accepts an arbitrary `-formatjs` directory without checking its revision. After changing the source checkout, rerun the extractor and review the generated fixture and `.skip-list.json` diff. The active Node witness separately enforces its exact version through `tools/conformance.ActiveNodeWitnessVersion`.

> **Why**: Keeping fixture provenance narrow reduces license and toolchain surface, and forces semantic interpretation back to ECMA-402 rather than to local taste.

### 2.2 Extraction tool

The generated-fixture extractor **Required**:

1. It is an independent Go module (independent `go.mod`), decoupled from the main module.
2. Enter: `.test.ts` source in the checked-in reference test tree. `__snapshots__/*.snap` is a reference output artifact; unless the extractor also has a paired source mapping that restores the input, neither the fixture nor `.skip-list.json` is entered.
3. Output: `<package>/testdata/conformance/formatjs/<source-slug>.json`; Each JSON file **MUST** contain only one `source` value to prevent the mixed-source gate of `tools/check-conformance` from failing.
4. Mechanical extraction **MUST** only cover assertion forms that can determine `{locale, options, input, expected}` losslessly. Currently active extractor supports:
   - `const nf = new NumberFormat("en", {...}); expect(nf.format(42)).toBe("42")`
   - `expect(new Intl.NumberFormat("en", {...}).format(42)).toEqual("42")`
   - `NumberFormat`'s `formatToParts` / `formatRange` / `formatRangeToParts` direct string or parts array assertions.
   - `formatjs:packages/intl-numberformat/tests/notation-compact-zh-TW.test.ts` compact decimal format assertions for `zh-TW` short and long compact display.
   - `const pr = new PluralRules("en", {...}); expect(pr.select(1)).toBe("one")`
   - `expect(new Intl.PluralRules("fr").select(1000000n)).toBe("many")`
   - Static numeric/BigInt literal assertion of `PluralRules.selectRange(start, end)`.
   - `const dtf = new DateTimeFormat("en-US", {...}); expect(dtf.format(date)).toBe("...")`
   - `Date` input in `DateTimeFormat.formatRange` / `formatRangeToParts` that is statically reducible to an RFC3339 string.
   - `toString` / `maximize` / `minimize` / canonicalization string assertions for `Intl.Locale` / `Intl.getCanonicalLocales`.
   - Static locale, options, string array input and string expectation assertions in `ListFormat.format`.
   - Static locale, options, numeric input, unit string and string expectation assertions in `RelativeTimeFormat.format`.
   - Static locale, options, duration object input and string expectation assertions in `DurationFormat.format`.
   - Simple JS object literal options: string / number / boolean values. The
     extractor must consume the complete options expression, including trivia
     and the closing object boundary. Duplicate keys retain their last literal
     value in source order. Unsupported values, spreads, computed keys, and
     other incompletely understood expressions do not produce partial options.
     A recognized but unsupported later constructor declaration blocks reuse
     of an earlier declaration with the same name. Such assertions remain in
     the existing source-owned extraction audit; supported assertions in the
     same source remain active.
   - PluralRules BigInt input is written to the fixture as a decimal string to avoid using float64 to carry integer semantics on the Go side.
5. The following `.test.ts` source **MUST** be written to `.skip-list.json`, each contains `source`, `category`, `route`, and `reason`, and silent discarding is prohibited:
- Table-driven arrays, callbacks, variable expected values, and other test shapes that cannot be restored statically without loss.
- Assertions that have been mechanically discovered but are outside the current generated fixture gate (e.g. locale/unit/currency-name behavior has not yet entered that gate).
- When only some assertions in the same source can be extracted losslessly, the source must still enter `.skip-list.json`, and the reason must explain why the remaining assertions do not enter the gate.
6. The error use case (`expect(...).toThrow(...)`) must be written into `errors.json` and separated from the success fixture.

The extractor scans each source once to identify code token positions, masking
comments without changing byte offsets and preserving quoted argument values.
Assertions, constructor declarations, and date declarations share that lexical
boundary. Text inside comments, strings, regex literals, and template text does
not create fixtures or extraction debt. Template substitutions are delimited but
are not mechanically extracted; assertion candidates inside them remain audit
items. Bare `expect` candidates allow whitespace and comment trivia before `(`.
Supported extraction is matched by candidate byte position, so repeated matches
cannot hide an uncovered assertion. These positions stay private to the tool.
Direct writes to a constructor variable invalidate its earlier configuration,
including chained assignments and writes inside template substitutions. Only an understood constructor
declaration establishes a new extractable binding; reassignment right-hand sides
are not evaluated. Earlier safe observations and independent literal constructors
remain extractable, and rejected assertions remain source-owned debt. Text in
comments, quoted strings, template text, and regex literals is not a write.

The scanner does not evaluate control flow or implement a complete JavaScript
parser. Unterminated literals/comments, excessive template nesting, and ambiguous
regex/division slashes stop extraction beyond the uncertain boundary and create a
source-owned skip reason with a byte offset. Safe earlier fixtures remain active.
A source with emitted fixtures and remaining debt is partial; a source with no
emitted fixtures is unsupported. This also applies when recognized assertions
are excluded by the generated-fixture support boundary.

Evidence: `tools/gen-fixtures-from-formatjs/source.go` and the extraction/import
regressions in `tools/gen-fixtures-from-formatjs/main_test.go` and
`options_test.go:TestConstructorAssignmentsRespectCodePositions`.

`tools/check-conformance/` is the single conformance audit CLI. It delegates to
the shared `tools/conformance` package and owns fixture schema validation,
XFAIL validation, skip-list validation, divergence/source integrity checks, and
coverage health output. It also owns the required native witness coverage matrix:
required native topics are enforced as fixture data, including one native
constructor error/refusal fixture for each active constructor. Intentional gaps
must carry a reason in the matrix. Formatter packages must not grow parallel
fixture, divergence, or native-witness validators.

`.skip-list.json` category value **MUST** be from the following collection:

| category | meaning |
|----------|------|
| `unsupported-extractor-shape` | The source has uncovered assertion candidates or a lexical diagnostic, and no fixtures can be emitted without loss |
| `partial-extraction` | Some assertions in the same source have generated fixtures, but the remaining assertions have not been covered |

The new category must be implemented in the same PR of extractor, `tools/check-conformance`, coverage report and this SPEC. Do not reserve a category that will not be recognized by the tool; the accepted product boundary will be owned SPEC + `testdata/divergences.md`, the implementation gap will be fixture failure / XFAIL, the missing reference will be tool error, not `.skip-list.json` debt.

`.skip-list.json` route value **MUST** be from the following collection:

| route | meaning |
|-------|---------|
| `extractor` | The source remains fixture debt owned by the extractor; deleting the skip entry requires generated fixtures or a narrower route. |
| `native-witness` | The source's observable behavior is intentionally covered by a native fixture; the entry must include `witness` pointing to that fixture id. |
| `not-applicable` | The source is outside the local ECMA-402 surface; the reason must name the boundary, not a missing implementation. |

**Rules**:

1. The extraction **MUST** be idempotent - running the same pinned reference twice must produce byte-identical output.
2. Manual editing of extracted generated-reference fixtures is **prohibited**; manual fixtures go to `<package>/testdata/conformance/manual/<file>.json` and are not confused with the generated-reference directory.
3. `.skip-list.json` is an extraction audit, not a test skip mechanism. Generated but failed fixtures must go through divergences.md or xfail.json; ungenerated sources can appear in the skip-list.
4. `tools/check-conformance` **MUST** verify fixture schema, XFAIL schema, skip-list schema, active skip-list categories, source uniqueness, and divergence-to-fixture consistency.
5. `task conformance:verify` **MUST** output coverage health: the number of fixture sources for each package, the number of manual / generated-reference / native fixtures, the number of active divergence, the number of xfail, and the skip-list category and route counts.
6. Package runners **MUST** call `tools/conformance.RunFixtures`, which loads fixtures, divergences, and XFAIL records into one validated suite before starting any fixture subtest. Missing package roots and malformed, expired, unknown, or mismatched ledger entries fail before a formatter callback can run.
7. `task conformance:verify` **MUST** validate the native witness matrix for every active package passed to `tools/check-conformance`; a missing required topic is a conformance audit failure, and an intentional gap without a reason is invalid.
8. Non-mechanizable fixtures, such as Date literals, callbacks, and complex error assertions, must be migrated manually; silently skipping is prohibited.
9. Removing a source from `.skip-list.json` **MUST** land a generated fixture file that records the exact `formatjs:` source path and observable expected output. Existing generated fixture files must remain byte-stable unless their source family is intentionally regenerated in the same change.

> **Why**: The extraction script is the only trusted bridge between reference tests and go-intl; idempotence ensures that diff is readable when upgrading references.
> **Rejected**: Blind AST full porting - Incomplete AST rules can generate fixtures that look formal but have incorrect input/options. Rather write complex sources into the source/reason skip-list than generate untrusted fixtures.

---

## 3. Divergence Handling

### 3.1 divergences.md file

`<package>/testdata/divergences.md` is required only when an active or resolved divergence entry exists for that package. Packages without accepted divergence history can omit this file; the tool treats missing files as having 0 active divergences. Each divergence entry **MUST** contain:

| Field | Meaning |
|------|------|
| `id` | One-to-one correspondence with fixture `id` field |
| `source` | Must be exactly the same as the `source` field of the corresponding fixture (`formatjs:` / `manual:` / `node:`, etc.) |
| `owner` | The formatter / data owner responsible for reviewing and deleting the divergence |
| `status` | Required; `accepted` for an active entry or `resolved` for validated history |
| `reason` | Why accept this difference (implementation-defined behavior / CLDR data version / reference fixture difference); must be sufficient to support human review |
| `native_witness` | Required for active `owner: datetimeformat` divergences; fixture `id` of a same-package `node:` fixture that records the native observable behavior |
| `review_after` | Next review anchor point (CLDR upgrade / Go 1.27 / Quarterly review date) |
| `removal_path` | The conditions or implementation paths that make this divergence disappear; if ECMA-402 is permanently implementation-defined, the retention conditions must be stated |

The record grammar is closed: only `id`, `source`, `owner`, `status`, `reason`,
`native_witness`, `review_after`, and `removal_path` are recognized.
`native_witness` is optional only when the active owner is not DateTimeFormat.
Blank lines and lines beginning with `#` are the only non-field lines; unknown
fields, duplicate fields, malformed non-empty lines, fields before `id`, and
trailing partial records are errors.

### 3.2 Handling Process

**Rules**:

1. CI fixture runners **MUST** compile accepted divergence and unexpired XFAIL IDs once while loading the suite, then skip assertions only for those validated IDs. Best-effort per-fixture file reads or swallowed ledger errors are forbidden.
2. Any failure that is not listed in divergences.md **MUST** block `task verify`.
3. Modifications to divergences.md **MUST** be explicitly approved by PR (reviewer ≥ 1 maintainer); automatic registration is **disallowed**.
4. The active `id` in divergences.md **MUST** find the corresponding entry in the fixture file; CI lint(`tools/check-conformance/`) checks the integrity.
5. The `source` of the active divergence must be exactly the same as the `source` of the fixture with the same ID.
6. Every record's `review_after` **MUST** be `YYYY-MM-DD`; `status` **MUST** be exactly `accepted` or `resolved`.
7. Duplicate divergence `id` values are prohibited.
8. Divergence entries are **FORBIDDEN** to be deleted; obsolete entries are marked with `status: resolved` and the audit trail is retained. Only empty placeholder files that never contained entries can be deleted.
9. Each divergence entry **MUST** contain a removal path. Long-term differences without removal paths must justify that ECMA-402 leaves the behavior implementation-defined.
10. Unimplemented functions are **FORBIDDEN** to be registered as accepted divergence; fixtures that can be generated must be kept in the fixture gate and explicitly stated through XFAIL, and sources that cannot be generated can only enter `.skip-list.json` according to the active extractor category. The narrowed product range must be written in owning SPEC and cannot be disguised as skip-list category.
11. Active `owner: datetimeformat` divergences **MUST** include `native_witness`, and `tools/check-conformance` must reject missing, unknown, non-native, expectation-free, or skipped witness fixtures. Range, time-zone, and resolved-options differences are too implementation-sensitive to live only in prose.
12. Resolved records remain history, but they receive the same closed-field, required-field, duplicate-ID, status, and date validation before being excluded from active fixture counting and matching.

> **Why**: divergences.md is the human review channel for spec interpretation rights; automatic registration will silently cover up "failed fixture" incidents.
> **Rejected**: Silent skips, such as moving a fixture to `testdata/skipped/`; this spec forbids that escape hatch.

### 3.3 Known divergence categories

The following categories **MUST** be visible (when present) in divergences.md:

| Category | Example |
|------|------|
| compact-threshold | Reference compact threshold differs from a locale data threshold |
| canonicalization | `en-arab-US` ⇒ `en-Arab-US` case rules (this class is forced to be aligned, and it is a bug if it occurs) |
| range-plural | `(1, 1.5)` ⇒ `'one'` / `'few'` depending on CLDR `pluralRanges.json` |
| hour12-default | `en-IN` default 12h vs 24h locale boundary |
| cldr-version-pin | go-intl pin CLDR 48.1.0 vs fixture source data baseline |

---

## 4. CI Gates and Performance Telemetry

`task verify` **MUST** be strung into correctness gates; performance remains visible through benchmark reports but is not a merge gate.

| Check | Function | Trigger command | PR behavior | main behavior |
|------|------|---------|---------|----------|
| **Gate 1: Conformance** | All non-divergence fixtures must pass | `task test`(`go test -race -p 1 ./...`) | block | block |
| **Gate 2: Conformance Audit** | fixture schema, XFAIL, skip-list, divergences.md consistent with fixture; none silently skip | `task conformance:verify` | block | block |
| **Performance Telemetry** | hot-path benchmark trend visibility | `task bench` / targeted `go test -bench` | non-blocking report | non-blocking artifact/report |

**Rules**:

1. Gate 1 + Gate 2 **MUST** block on all PRs; **override is prohibited (unless the maintainer explicitly approves and revise SPEC).
2. Benchmark numbers are non-blocking telemetry and do not determine the command's exit code. Execution failures, such as a failed build or benchmark, still return non-zero; a measured slowdown alone is not a merge blocker.
3. `task verify` **Required** serial Gate 1 → Gate 2 and other correctness/security checks; performance reports run separately when relevant.
4. Gate 1 or Gate 2 fails **MUST** return a non-zero exit code, and CI blocks accordingly.

> **Why**: correctness and divergence are repeatable contract judgments; benchmark is an environment-sensitive trend signal. Turning noisy performance numbers into red can induce bypassing ECMA-402 correctness or removing the root facade surface.
> **Rejected**: main/nightly performance block. The report only needs to retain evidence; whether to optimize or not is determined by SPEC 71's non-blocking telemetry policy and reviewer judgment.

---

## 5. XFAIL Discipline

XFAIL(expected failure)= fixture is known to fail but is allowed to pass CI. **MUST** have an expiration date.

### 5.1 XFAIL Registration

XFAIL entry **MUST** be located in `<package>/testdata/xfail.json`, field:

| Field | Meaning |
|------|------|
| `id` | fixture id |
| `reason` | Reason for failure (missing implementation / upstream bug / waiting for CLDR upgrade) |
| `expires_at` | ISO 8601 date; expired CI **Required** fail this fixture |
| `tracking_issue` | go-intl issue URL |

**Rules**:

1. Each XFAIL **MUST** have `expires_at`; **FORBIDDEN** permanent XFAIL (replace with divergences.md).
2. XFAIL `id` **MUST** be unique, and the corresponding entry must be found in the fixture file.
3. `expires_at` Default ≤ 90 days; extension **MUST** be explicitly approved by PR.
4. CI **MUST** automatically fail after XFAIL expires - implemented using fixture runner's "date > expires_at force assertion".
5. XFAIL **FORBIDDEN** is used for "I'm too lazy to fix" - only allowed: upstream reference/data bugs, tracked local issues with explicit priority, and cases waiting on SPEC revisions.

> **Why**: XFAIL aging is an anti-entropy mechanism - an XFAIL without an expiration date will become a "TODO comment that will never be repaired", and a year later the CI screen will be full of yellow.
> **Rejected**: Indefinite XFAIL - This turns a temporary exemption into long-term unvalidated behavior.

### 5.2 The difference between XFAIL and divergence

| Dimension | XFAIL | Divergence |
|------|-------|-----------|
| Nature | Known bugs / Not implemented | Known behavioral differences |
| Solution | Delete after repair | Long-term coexistence |
| Timeliness | Must have expiration date | Permanent (but requires review_after anchor) |
| CI Behavior | Skip Assertion | Skip Assertion |
| Source | go-intl’s own code defects | implementation-defined behavior / CLDR version difference |

---

## 6. Forbidden

- **Ban** silently skip failed fixtures (must go to divergences.md or xfail.json, both require PR approval).
- **BANNED** silently skip unextracted reference `.test.ts` source (must write `.skip-list.json` of `source` + `category` + `route` + `reason`); snapshot-only artifact is not source debt.
- **Disabled** fixture has no `id` or `source` field.
- **FORBIDDEN** Mixing generated-reference, native-witness, and manual sources within the same JSON file, even when the historical source prefixes differ only by version.
- **BANNED** Using Unix epoch ms numbers as DateTime input (requires ISO-8601 string).
- **FORBIDDEN** Delete the divergences.md historical entries (outdated entries are changed to `status: resolved` and retained); empty placeholder files do not need to be retained.
- **FORBIDDEN** XFAIL None `expires_at`.
- **BANNED** ICU4J from entering the CI toolchain (Java overhead does not match ROI).
- **FORBIDDEN** Introduce assertion libraries such as `stretchr/testify`; test stack limit stdlib `testing` + `google/go-cmp`.
- **BANNED** Introducing a snapshot framework (`gkampitakis/go-snaps` / `bradleyjkemp/cupaloy`, etc.) as an alternative to fixture JSON. Fixture JSON is sufficient and keeps review diffs explicit.
- **FORBIDDEN** Extracting manually edited script products (manual fixtures go to the `manual/` directory).
- **BANNED** `.skip-list.json` is passed as a pass by the fixture runner; it only audits extraction coverage.
- **FORBIDDEN** performance telemetry blocks PR or main; **FORBIDDEN** Gate 1/2 does not block in main (must block).
- **BANNED** Calling fixture loader from hot-path Go code (loader only in `_test.go` file).

---

## 7. Acceptance Ledger

SPEC 70 is accepted by the unified conformance gate plus the product workflow
rules in AGENTS/CLAUDE. The ledger separates repo-local evidence from
governance rules that cannot be proven by a Go test.

### Fixture Schema And Sources

| Contract | Evidence | Status |
|----------|----------|--------|
| Conformance fixture files are JSON arrays matching §1.1, with globally unique IDs and source-directory consistency. | `tools/conformance/fixtures.go`; `tools/conformance/skip_test.go`; `tools/check-conformance/main.go`; `task conformance:verify` | Satisfied |
| Error fixtures live in `errors.json` lanes and assert sentinel behavior through package runners. | `*/testdata/conformance/node-v26/errors.json`; package `conformance_unified_test.go` files | Satisfied |
| NumberFormat independently executes single/range text, parts, and resolved observations, including empty expected arrays; deliberately incorrect observations fail the actual callback. | `numberformat/conformance_observations_test.go`; `numberformat/conformance_unified_test.go` | Satisfied |
| PluralRules `selectRange` runtime `invalid_value` executes SelectRange; NaN fails, Infinity remains legal, and Select has no error lane. | `pluralrules/conformance_observations_test.go` | Satisfied |
| PluralRules executes complete resolved JSON without a select input, including mixed category/snapshot records; parts observations fail explicitly. | `pluralrules/conformance_observations_test.go` | Satisfied |
| ListFormat independently observes complete resolved JSON without decoding list input; unsupported range observations fail. | `listformat/conformance_observations_test.go` | Satisfied |
| RelativeTimeFormat executes complete resolved/text/parts independently; runtime error feature selects `format` or `formatToParts`, with unsupported observations rejected. | `relativetimeformat/conformance_observations_test.go` | Satisfied |
| DateTimeFormat, ListFormat, and RelativeTimeFormat execute parts observations without a text prerequisite and assert empty expected arrays. | Package `conformance_observations_test.go` and `conformance_unified_test.go` files | Satisfied |
| DurationFormat supportedLocalesOf fixtures call the static method directly, compare the complete locale list, and ignore formatting options. | `durationformat/conformance_observations_test.go` | Satisfied |
| DurationFormat runtime `invalid_value` lanes execute the method named by `feature` (`format` by default, or `formatToParts`); constructor `invalid_option` stays in New. | `durationformat/conformance_observations_test.go` | Satisfied |
| DurationFormat uses the shared validated suite, executes text/parts/resolved observations independently, and rejects malformed or expired ledgers before its callback. | `durationformat/conformance_observations_test.go`; `durationformat/conformance_unified_test.go` | Satisfied |
| DisplayNames compares complete resolved-options JSON snapshots, including omitted languageDisplay; the shared comparator preserves scalar types and zero values. | `displaynames/conformance_snapshots_test.go`; `internal/testcontract/resolved_json_test.go` | Satisfied |
| NumberFormat compares complete snapshots, including Boolean false grouping and absent fraction properties for significant-digit rounding. | `numberformat/conformance_snapshots_test.go`; `numberformat/testdata/conformance/manual/resolved-options.json` | Satisfied |
| DateTimeFormat compares complete snapshots, including hour-cycle and style properties and the absence of automatic dayPeriod. | `datetimeformat/conformance_snapshots_test.go`; existing Node resolved-options fixtures | Satisfied |
| DurationFormat compares complete snapshots and distinguishes omitted fractionalDigits from explicit zero; no formatter uses the former subset comparator. | `durationformat/conformance_snapshots_test.go`; `internal/testcontract/resolved_json.go` | Satisfied |
| `tools/gen-fixtures-from-formatjs/` is a standalone module and owns generated FormatJS lanes for currently extractable surfaces. | `tools/gen-fixtures-from-formatjs/go.mod`; `tools/gen-fixtures-from-formatjs/main.go`; generated `testdata/conformance/formatjs` fixtures | Satisfied |
| Native witness validation enforces required topics, constructor error/refusal coverage, and explicit intentional gaps. | `tools/conformance/node_witness.go`; `tools/conformance/node_witness_test.go`; `tools/conformance/product_contract_test.go`; `task conformance:verify` | Satisfied |
| `.skip-list.json` audits non-extracted and partially extracted reference sources with `source`, `category`, `route`, and `reason`. | `.skip-list.json`; `tools/conformance/coverage.go`; `tools/conformance/coverage_test.go` | Satisfied |
| The repository reference checkout is identified by its submodule commit; `formatjs=local-reference` is a source label, and arbitrary `-formatjs` directories are not revision-validated. Node witness regeneration enforces the exact active Node version. | Git submodule commit; `tools/.gen-versions`; `tools/gen-fixtures-from-formatjs/main.go:importNode`; `tools/conformance.ActiveNodeWitnessVersion` | Satisfied |

### Divergences And XFAIL

| Contract | Evidence | Status |
|----------|----------|--------|
| Packages with active or resolved divergence history keep `testdata/divergences.md`; packages without divergence history may omit it. | `datetimeformat/testdata/divergences.md`; `tools/conformance/divergences.go` | Satisfied |
| Divergence records use a closed, unique field set; malformed, duplicate, unknown, incomplete, or invalid-date records fail. Resolved history is fully validated before filtering, while active records also name an existing fixture and DateTimeFormat native witness when required. | `tools/conformance/divergences.go`; `tools/conformance/coverage_test.go`; `task conformance:verify` | Satisfied |
| XFAIL records require `id`, `reason`, `expires_at`, and `tracking_issue`; IDs must be unique, real, and unexpired. | `tools/conformance/xfail.go`; `tools/conformance/skip_test.go`; `tools/check-conformance/main_test.go` | Satisfied |
| `RunFixtures` validates the complete suite before callbacks; duplicate fixture IDs, malformed divergence/XFAIL state, expired entries, unknown IDs, and missing roots fail closed. | `tools/conformance/runner.go`; `tools/conformance/skip_test.go` subprocess witnesses | Satisfied |
| XFAIL growth must be reviewed toward zero as implementation matures. | `task conformance:verify` reports xfail totals; human review decides whether growth is justified | Governance |

### Gates And Drift Rules

| Contract | Evidence | Status |
|----------|----------|--------|
| `task verify` runs the conformance gate with the other correctness/security checks. | `Taskfile.yml` `verify` and `conformance:verify` tasks | Satisfied |
| PR/main CI runs conformance as a blocking correctness gate. | `.github/workflows/ci.yml` | Satisfied |
| Performance telemetry is non-blocking and separated from conformance correctness. | `SPECS/71-benchmark.md`; `Taskfile.yml` benchmark tasks | Satisfied |
| Generated fixtures must pass in their package runner except known divergence/XFAIL records. | Package `conformance_unified_test.go` files; `tools/conformance/RunFixtures` | Satisfied |
| Public API, option, resolved-option, part, and range-source changes must first locate the ECMA-402 owner or a narrow Go typed bridge. | `AGENTS.md`; `SPECS/72-operation-ledger.md`; relevant surface specs | Governance |
| Mechanically extractable reference cases enter fixtures; non-extractable sources enter `.skip-list.json`; failing generated fixtures are handled only by implementation repair, divergence records, or XFAIL. | `tools/conformance/coverage.go`; `.skip-list.json`; package fixture runners | Satisfied |

### Toolchain Constraints

| Contract | Evidence | Status |
|----------|----------|--------|
| Snapshot/assertion dependencies such as `testify`, `go-snaps`, `cupaloy`, and `goldie` are absent. | `go.mod` | Satisfied |
| No test-only dependency is required for conformance fixtures; the old `google/go-cmp` allowance is no longer part of the active toolchain. | `go.mod`; stdlib package tests | Satisfied |

### Open Governance Notes

The maintainer-review requirement for `divergences.md` changes cannot be
verified from repository-local Go tests; it belongs in branch protection or PR
review policy.

---

## References

- SPECS/00 §2.1(test fixture policy), §2.2(reference hygiene)
- SPEC 60 (Acceptance Criteria refers to this SPEC fixtures)
- SPEC 71(Benchmark Strategy & Performance Telemetry)
- `Taskfile.yml` `conformance:verify`
- `tools/check-conformance/`
- `tools/conformance/`
- `.references/formatjs/packages/intl-numberformat/tests/` (main extraction source)
- `.references/formatjs/packages/intl-datetimeformat/tests/` (main extraction source)
- `.references/formatjs/packages/intl-pluralrules/tests/` (main extraction source)
- `.references/formatjs/packages/intl-locale/tests/` (main extraction source)
- `.references/formatjs/packages/intl-listformat/tests/` (main extraction source)
- `.references/formatjs/packages/intl-relativetimeformat/tests/` (main extraction source)
- `.references/formatjs/packages/intl-durationformat/tests/` (main extraction source)

The active Node witness refresh validates its exact `nodeVersion` against the
version consumed by the conformance checker before writing fixtures; a different
Node release requires an explicit checker/provenance upgrade. Every generated
fixture group must be nonempty, and supported-values observations must include
calendar, currency, numberingSystem, timeZone, and unit. Empty observation values
remain valid. The importer stages the fixture records and reuses the conformance
loader's schema, source, error-file, and identity checks before changing any
official output; old files cannot fill gaps in the incoming payload. Node
publication replaces only the generator-owned exact files, preserving manual
witness files in the same directories and the other fixture lanes. The FormatJS
extractor prepares all owned directories and the skip-list on the output
filesystem, then checks the complete target collection and parent path types
before replacing any output. Extraction, encoding, staging, and preflight
failures preserve prior fixtures and skip-list. Both importers share one private
publication operation that retains the old target collection until replacement
succeeds. A detected replacement failure
restores touched targets, including restoring absence for a newly created target.
If recovery itself fails, the error preserves both causes and identifies retained
backups outside the disposable staging directory. This is finite failure recovery,
not crash atomicity or coordination between concurrent generators.
It does not own DisplayNames'
older FormatJS lane, manual fixtures,
Node witnesses, or divergence/XFAIL ledgers. Source-deleted/zero-assertion
lanes are removed on successful extraction.

## Executable witnesses and generator identity

`loadRunSuite` validates fixture and ledger structure, computes active skip
reasons, then checks witness references. Required-topic coverage, skip-list
native routes, and divergence native witnesses use that same state. A candidate
covered only by XFAIL/divergence skips does not satisfy coverage; another
eligible fixture can. Self/mutual skipped witnesses are rejected. Diagnostics
identify the topic/reference, ID and skip reason. Eligibility proves scheduling,
not execution success; actual formatter callbacks must still pass.
Evidence: `tools/conformance/node_witness_test.go`, `coverage_test.go`, and
`skip_test.go`.

DisplayNames dispatches static `supportedLocalesOf` before constructing an
instance, so Type may be omitted there. Instance text, ok and resolved-options
observations are all asserted. Unsupported features, parts/ranges (including
empty arrays), and incompatible observations mixed into error fixtures fail.
Resolved options may coexist with an Of error because construction succeeded
and those options are asserted. See `displaynames/conformance_observations_test.go`.

The published `tools/gen-fixtures-from-formatjs` module owns `node_script.go`
and directly invokes the chosen Node executable exactly once. It never searches
the caller's CWD for repository source or invokes `go run` at runtime.
`Taskfile.yml:conformance:witness` delegates to this command. Exact Node version
validation and staged fixture publication remain unchanged.

Generated FormatJS identity hashes deterministic JSON containing source path
(with `/` separators), locale, feature/operation, options, input and eight
observation-presence flags. A readable surface/source prefix precedes SHA-256.
Expected values, assertion position and file-wide ordinal are excluded. Integer
inputs retain their integer precision; floating JS literals retain the scanner's
Number conversion. Nil versus present-empty observation arrays are distinct.
Empty options and omitted options both represent the empty JS options object.
The same invocation with conflicting expected values fails with source and ID;
fully identical cases merge. Manual and native lanes are not reidentified.

The 2026-10-21 migration mapped all 226 old generated IDs by complete non-ID
payload, changed no expected values, and merged only these identical pairs:

- PluralRules index `009` / `043`: fr, empty options, 1000000 → many.
- NumberFormat misc `050` / `051`: en, signDisplay=exceptZero, 0 → 0.

All ledger bindings followed the same payload mapping. Source extraction is
repeatable; `main_test.go:TestFormatJSSemanticIdentity` protects insertion,
reordering/expectation independence and conflict detection; precision and
observation kinds are covered by `TestSemanticIdentityObservationsAndPrecision`.
