# SPEC 70 — Conformance Test Strategy

> **Status:** Revised (2026-05-20)
> **Type:** Flow + Schema + Rule — defines how go-intl proves ECMA-402 observable behavior through FormatJS-derived and native-Intl fixtures.
> **Authority:** ECMA-402 and native/reference fixtures define observable behavior. This spec records the conformance fixture format, fixture sources, divergence handling, correctness gates, and XFAIL discipline. Per-formatter SPECS (10/20/30/40/41/42/43/44/45/46) record their own option semantics; this spec records how those semantics are *verified* against the reference implementations.

---

## Overview

The "correctness" of go-intl active scope is defined by the **ECMA-402 specification + fixture-driven conformance tests**: the specification determines the semantic boundaries, fixtures mechanically or manually extract input/output pairs from FormatJS tests or Node/V8 native Intl snapshots, and assert them one by one within the Go-side harness. This SPEC covers:

1. **fixture format record**: unified JSON schema, universal across formatters.
2. **fixture source**: FormatJS tests, Node/V8 native Intl smoke cases and a small number of handwritten ECMA-402 edge use cases.
3. **divergence process**: known deviations are registered in `<package>/testdata/divergences.md`; silent skips are **FORBIDDEN**.
4. **CI gates + telemetry**: conformance tests (blocking), conformance audit (blocking), and performance telemetry (non-blocking report).
5. **XFAIL aging**: Each XFAIL must have an expiration date, and it will automatically fail when it expires.

> **Why**: ECMA-402 one-to-one alignment is the core commitment of SPECS/00 §1; fixtures are the execution mechanism for that commitment.
> **Rejected**: Relying only on handwritten unit tests to cover ECMA-402 behavior. FormatJS has accumulated a broad fixture corpus, and handwritten tests would not catch up with that coverage.

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
| `id` | Required | string | Globally unique; stable slug, it is recommended to include formatter, source, feature / case index; divergences.md is referenced by this id |
| `source` | Required | string | `<source>:<path>` —— `formatjs:` or `manual:` |
| `locale` | Required | string | BCP 47 locale tag, the Go end uses `locale.Parse(string)` to parse at the fixture boundary |
| `options` | Required | object | ECMA-402 option; key name spec original text (camelCase) |
| `input` | Required | number / string / array / object | Numeric literal, string list (ListFormat), ISO-8601 string (DateTimeFormat), `{start, end}` object (Range), `{value, unit}` object (RelativeTimeFormat), duration record object (DurationFormat) |
| `expected` | optional | string | `Format` output; FormatToParts-only fixture can be omitted |
| `expectedLocales` | Optional | array | `SupportedLocalesOf` output; omitting means no verification supported locales |
| `expectedParts` | Optional | array | `FormatToParts` output; RelativeTimeFormat number parts can include `unit`; omitting means not checking parts |
| `expectedRange` | optional | string | `FormatRange` output |
| `expectedRangeParts` | optional | array | `FormatRangeToParts` output |
| `errorCode` | Optional | string | Error case (replacing `expected`), corresponding to go-intl sentinel name |

**Rules**:

1. The fixture file **MUST** be a JSON array (multiple fixtures per file), with the path `<package>/testdata/conformance/<source>/<file>.json`.
2. `id` **MUST** be globally unique; when violated, CI will directly block (`tools/check-conformance/` verification).
3. When `input` is a date, you must use ISO-8601 with timezone offset string (such as `"2020-01-01T00:00:00Z"`), and Go harness deserializes it through `time.Parse(time.RFC3339, ...)`. **FORBIDDEN** to use Unix epoch ms numbers (FormatJS `Date.now()` style), too ambiguous.
4. The error fixture **MUST** go through `errorCode` field, which is kept separately in `<package>/testdata/conformance/<source>/errors.json` and separated into files with the success fixture; Go harness goes through `errors.Is(err, sentinel)` verification.
5. `errors.json` **can only** contain error fixtures with `errorCode`; forward fixtures **are prohibited** from being put into `errors.json`.
6. The source directory where the fixture is located **MUST** be consistent with the `source` field prefix: `manual/` for `manual` or `manual:*`, `formatjs/` for `formatjs:*`, `node-*` for `node:*`.
7. The `options` field **MUST** maintain the original ECMA-402 spec naming (`maximumFractionDigits` instead of `MaximumFractionDigits`); the Go-side harness is mapped to a typed `Options` value when loading.
8. **It is prohibited** to embed JS functions, callbacks, and Date literals in fixtures; parts that cannot be mechanically extracted are classified according to the SPEC §2.4 process.

> **Why**: The unified schema is universal across formatters, and the harness can be shared; a schema is also the contract for future integration testing of messageformat-go.
> **Rejected**: Each formatter custom schema (NumberFormat `"value"` vs DateTimeFormat `"date"`) - 4 sets of harness, 4 sets of loaders, DRY violation.

### 1.3 Manifest

`<package>/testdata/conformance/<source>/MANIFEST.json` **SHOULD** record for each fixture file:

| Field | Meaning |
|------|------|
| `extractor_version` | Version tag of `tools/gen-fixtures-from-formatjs/` |
| `extracted_from` | FormatJS submit SHA + test file path |
| `extracted_at` | ISO 8601 timestamp |
| `count` | Number of fixtures in this file |

**SHOULD** (not mandatory for active scope): manifest is missing by PR-CI lint verification. active scope can be replaced by git blame; consumer-driven expansion trigger condition = fixture PR number > 5/month.

> **Why**: The manifest is an audit trail - when FormatJS upgrade causes fixture drift, the manifest allows the reviewer to locate "which PR was introduced" at a glance.
> **Rejected**: Mandatory manifests in active scope. Git history is sufficient for the first release, so a required manifest would be YAGNI until fixture volume or review workflow demands it.

---

## 2. Fixture Sources

### 2.1 Fixture source matrix

| Source | role | active scope requirement |
|--------|------|-------------|
| **FormatJS Vitest** | **Main fixture source** | Generated fixtures must pass; ungenerated test sources must enter `.skip-list.json` audit |
| **node / V8 native Intl** | Native-engine contract source | Each active constructor retains at least one native snapshot; FormatJS extractor has not yet covered or does not have a readable polyfill surface that cannot stay in smoke:Collator must record search / caseFirst / collation option contracts before backend acceptance, Segmenter must retain word and sentence contracts for each advertised locale |
| **manual ECMA-402 edge cases** | Supplementary fixture source | Only for FormatJS/Node boundaries that cannot be mechanically extracted or explicitly required by spec |

**Rules**:

1. The main fixture **MUST** come from FormatJS Vitest, translated through `tools/gen-fixtures-from-formatjs/`; the normative interpretation rights still belong to the ECMA-402 spec.
2. Node/V8 fixtures **MUST** record the Node major version and mark it with `source: "node:<version>:<surface>"`; it is a native behavior tiebreaker, not a source of new semantics to bypass ECMA-402.
3. Manual fixture **MUST** indicate that it corresponds to the ECMA-402 section or local SPEC rationale, and is marked with `source: "manual:<topic>"`.
4. Fixture sources **MUST** be marked via the `source` field; mixing sources into the same JSON file is prohibited.
5. The FormatJS version **MUST** be pinned to `tools/.gen-versions`; the current value is the local `.references/formatjs` reference. After upgrading FormatJS or refreshing submodule, you must rerun extractor and check the generated fixture and `.skip-list.json` diff.

> **Why**: FormatJS is the maintained TypeScript ECMA-402 polyfill and the only vendored implementation reference. Keeping fixture provenance narrow reduces license and toolchain surface.

### 2.2 Extraction tool

`tools/gen-fixtures-from-formatjs/` **Required**:

1. It is an independent Go module (independent `go.mod`), decoupled from the main module.
2. Enter: `.test.ts` source in FormatJS `packages/<polyfill>/tests/`. `__snapshots__/*.snap` is a reference output artifact; unless the extractor also has a paired source mapping that restores the input, neither the fixture nor `.skip-list.json` is entered.
3. Output: `<package>/testdata/conformance/formatjs/<source-slug>.json`; Each JSON file **MUST** contain only one `source` value to prevent the mixed-source gate of `tools/check-conformance` from failing.
4. Mechanical extraction **MUST** only cover assertion forms that can determine `{locale, options, input, expected}` losslessly. Currently active extractor supports:
   - `const nf = new NumberFormat("en", {...}); expect(nf.format(42)).toBe("42")`
   - `expect(new Intl.NumberFormat("en", {...}).format(42)).toEqual("42")`
- `NumberFormat`'s `formatToParts` / `formatRange` / `formatRangeToParts` direct string or parts array assertions.
   - `const pr = new PluralRules("en", {...}); expect(pr.select(1)).toBe("one")`
   - `expect(new Intl.PluralRules("fr").select(1000000n)).toBe("many")`
- Static numeric/BigInt literal assertion of `PluralRules.selectRange(start, end)`.
   - `const dtf = new DateTimeFormat("en-US", {...}); expect(dtf.format(date)).toBe("...")`
- `Date` input in `DateTimeFormat.formatRange` / `formatRangeToParts` that is statically reducible to an RFC3339 string.
- `toString` / `maximize` / `minimize` / canonicalization string assertions for `Intl.Locale` / `Intl.getCanonicalLocales`.
- Static locale, options, string array input and string expectation assertions in `ListFormat.format`.
- Static locale, options, numeric input, unit string and string expectation assertions in `RelativeTimeFormat.format`.
- Static locale, options, duration object input and string expectation assertions in `DurationFormat.format`.
- Simple JS object literal options: string / number / boolean values.
- PluralRules BigInt input is written to the fixture as a decimal string to avoid using float64 to carry integer semantics on the Go side.
5. The following `.test.ts` source **MUST** be written to `.skip-list.json`, each contains `source`, `category` and `reason`, and silent discarding is prohibited:
- `it.each(table)` / `tests` arrays/callbacks/variable expected values and other Vitest shapes that cannot be restored statically without loss.
- Assertions that have been mechanically discovered but are outside the current generated fixture gate (e.g. locale/unit/compact/currency-name/selectRange behavior has not yet entered that gate).
- When only some assertions in the same source can be extracted losslessly, the source must still enter `.skip-list.json`, and the reason must explain why the remaining assertions do not enter the gate.
6. The error use case (`expect(...).toThrow(...)`) must be written into `errors.json` and separated from the success fixture.

`tools/check-conformance/` is the single conformance audit CLI. It delegates to
the shared `tools/conformance` package and owns fixture schema validation,
XFAIL validation, skip-list validation, divergence/source integrity checks, and
coverage health output. Formatter packages must not grow parallel fixture or
divergence validators.

`.skip-list.json` category value **MUST** be from the following collection:

| category | meaning |
|----------|------|
| `unsupported-extractor-shape` | The source file contains `expect(...)`, but the current extractor cannot restore it statically without loss |
| `partial-extraction` | Some assertions in the same source have generated fixtures, but the remaining assertions have not been covered |

The new category must be implemented in the same PR of extractor, `tools/check-conformance`, coverage report and this SPEC. Do not reserve a category that will not be recognized by the tool; the accepted product boundary will be owned SPEC + `testdata/divergences.md`, the implementation gap will be fixture failure / XFAIL, the missing reference will be tool error, not `.skip-list.json` debt.

**Rules**:

1. The extraction **MUST** be idempotent - the same as running FormatJS commit twice to produce byte-identical.
2. Manual editing of the extracted product is **prohibited**; manual fixtures go to `<package>/testdata/conformance/manual/<file>.json` and are **not** confused with the formatjs/ directory.
3. `.skip-list.json` is an extraction audit, not a test skip mechanism. Generated but failed fixtures must go through divergences.md or xfail.json; ungenerated sources can appear in the skip-list.
4. `tools/check-conformance` **MUST** verify fixture schema, XFAIL schema, skip-list schema, active skip-list categories, source uniqueness, and divergence-to-fixture consistency.
5. `task conformance:verify` **MUST** output coverage health: the number of fixture sources for each package, the number of manual / FormatJS / node fixtures, the number of active divergence, the number of xfail, and the skip-list category count.
6. Non-mechanizable fixtures, such as Date literals, callbacks, and complex error assertions, must be migrated manually; silently skipping is prohibited.

> **Why**: The extraction script is the only trusted bridge between FormatJS and go-intl; idempotence ensures that diff is readable when upgrading FormatJS.
> **Rejected**: Blind AST full porting - Incomplete AST rules can generate fixtures that look formal but have incorrect input/options. Rather write complex sources into the source/reason skip-list than generate untrusted fixtures.

### 2.3 Consumer Profile Fixtures

Consumer profiles live under `testdata/consumer/<consumer>/intl-profile.json`. They are cross-surface compatibility fixtures, not a replacement for formatter conformance fixtures. Use them when a real host or adapter depends on a small set of observable boundaries that span packages.

Active profile:

| Consumer | Fixture | Runner | Scope |
|----------|---------|--------|-------|
| JS host integration | host consumer profile | `consumer_profile_test.go` | Supported-set exclusions, root supported-value include/exclude boundaries, and reversed range behavior for NumberFormat, DateTimeFormat, and PluralRules |

Rules:

1. Consumer profiles must be JSON objects loaded only from `_test.go` files.
2. They may assert cross-package host contracts; they must not duplicate broad ECMA-402 output matrices that belong in formatter conformance fixtures.
3. They may assert that unsupported capabilities stay unadvertised, such as collator tailoring, unsupported DateTimeFormat calendars, or Segmenter dictionary/CJK locales.
4. They must preserve caller-provided range order. Reversed ranges are valid inputs for NumberFormat, DateTimeFormat, and PluralRules unless the owning formatter SPEC says otherwise.
5. A profile change that weakens support boundaries or output behavior must update the owning SPEC first.
6. Consumer profiles must not force public `map[string]any` APIs. Host-boundary needs are served by typed Go APIs plus JSON-marshallable ECMA-402 records.

> **Why**: Host adapters need a thin contract that cuts across root supported values, constructor supported locales, and range behavior. A profile keeps that dependency visible without turning README examples or per-package fixtures into consumer-specific code.
> **Rejected**: Baking consumer behavior into production adapters or README prose only. Tests must protect the host contract, and SPECS must explain why the protection exists.

---

## 3. Divergence Handling

### 3.1 divergences.md file

`<package>/testdata/divergences.md` is required only when an active or resolved divergence entry exists for that package. Packages without accepted divergence history can omit this file; the tool treats missing files as having 0 active divergences. Each divergence entry **MUST** contain:

| Field | Meaning |
|------|------|
| `id` | One-to-one correspondence with fixture `id` field |
| `source` | Must be exactly the same as the `source` field of the corresponding fixture (`formatjs:` / `manual:` / `node:`, etc.) |
| `owner` | The formatter / data owner responsible for reviewing and deleting the divergence |
| `status` | Optional; active entry is empty or `accepted`, historical entry is `resolved` |
| `reason` | Why accept this difference (implementation-defined behavior / CLDR data version / FormatJS difference); must be sufficient to support human review |
| `review_after` | Next review anchor point (CLDR upgrade / Go 1.27 / Quarterly review date) |
| `removal_path` | The conditions or implementation paths that make this divergence disappear; if ECMA-402 is permanently implementation-defined, the retention conditions must be stated |

Optional evidence fields such as `our`, `reference`, `category` can be retained, but CI audit mandatory fields are `id`, `source`, `owner`, `reason`, `review_after`, and `removal_path`.

### 3.2 Handling Process

**Rules**:

1. CI fixture runners **MUST** read the `id` list from divergences.md and skip the assertion for matching accepted divergences.
2. Any failure that is not listed in divergences.md **MUST** block `task verify`.
3. Modifications to divergences.md **MUST** be explicitly approved by PR (reviewer ≥ 1 maintainer); automatic registration is **disallowed**.
4. The active `id` in divergences.md **MUST** find the corresponding entry in the fixture file; CI lint(`tools/check-conformance/`) checks the integrity.
5. The `source` of the active divergence must be exactly the same as the `source` of the fixture with the same ID.
6. `review_after` of active divergence **MUST** be `YYYY-MM-DD`; `status` **can only** be empty, `accepted` or `resolved`.
7. Duplicate divergence `id` values are prohibited.
8. Divergence entries are **FORBIDDEN** to be deleted; obsolete entries are marked with `status: resolved` and the audit trail is retained. Only empty placeholder files that never contained entries can be deleted.
9. Each divergence entry **MUST** contain a removal path. Long-term differences without removal paths must justify that ECMA-402 leaves the behavior implementation-defined.
10. Unimplemented functions are **FORBIDDEN** to be registered as accepted divergence; fixtures that can be generated must be kept in the fixture gate and explicitly stated through XFAIL, and sources that cannot be generated can only enter `.skip-list.json` according to the active extractor category. The narrowed product range must be written in owning SPEC and cannot be disguised as skip-list category.

> **Why**: divergences.md is the human review channel for spec interpretation rights; automatic registration will silently cover up "failed fixture" incidents.
> **Rejected**: Silent skips, such as moving a fixture to `testdata/skipped/`; this spec forbids that escape hatch.

### 3.3 Known divergence categories

The following categories **MUST** be visible (when present) in divergences.md:

| Category | Example |
|------|------|
| compact-threshold | FormatJS `1.2K` starting point is different from a certain locale ICU threshold |
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
2. Performance telemetry **MUST** return non-zero exit code alone; benchmark numbers cannot be used as standalone merge blocker.
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
5. XFAIL **FORBIDDEN** is used for "I'm too lazy to fix" - only allowed: upstream bugs to be fixed (FormatJS/CLDR), issues in this warehouse that have been opened but have low priority, and depend on SPEC revisions.

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
- **BANNED** silently skip unextracted FormatJS `.test.ts` source (must write `.skip-list.json` of `source` + `category` + `reason`); snapshot-only artifact is not source debt.
- **Disabled** fixture has no `id` or `source` field.
- **FORBIDDEN** Mixing `formatjs:` / `node:` / `icu4j:` sources within the same JSON file.
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

## 7. Acceptance Criteria

### Fixture Schema

- [ ] `<package>/testdata/conformance/<source>/*.json` exists; each file is a JSON array conforming to the SPEC §1.1 schema.
- [ ] `tools/check-conformance/` check `id` is globally unique; violates block CI.
- [ ] The error case is located independently in `errors.json` and passes the `errors.Is` verification.
- [ ] `tools/check-conformance/` Verify that the fixture `source` field is consistent with the source directory.

### Fixture Sources

- [ ] `tools/gen-fixtures-from-formatjs/` is a standalone Go module; currently generated gate outputs at least NumberFormat format / parts / range, DateTimeFormat format / range, PluralRules select / selectRange, Locale canonicalization, ListFormat format, RelativeTimeFormat format, and DurationFormat format fixtures. DisplayNames must retain at least Node/V8 smoke fixtures; Collator and Segmenter must retain product-contract Node/V8 fixtures until FormatJS or native extractor covers a more complete test matrix.
- [ ] The root directory `.skip-list.json` exists, each record contains `source`, `category` and `reason`, and covers FormatJS `.test.ts` sources / partial sources that cannot be mechanically extracted or exceed the current generated gate.
- [ ] FormatJS reference is pinned to `tools/.gen-versions`; CI check exists.
- [ ] A host consumer profile is loaded by a root consumer-profile test and covers supported-set boundaries plus reversed ranges.

### Divergences

- [ ] Formatter packages containing active or resolved divergence history retain `<package>/testdata/divergences.md`; packages without divergence history may omit this file.
- [ ] `tools/check-conformance/` verifies active divergence `id` can be located in the fixture, `source` is consistent with the fixture, `status` / `review_after` is legal; violates block CI.
- [ ] divergences.md modification PR requires ≥ 1 maintainer review on GitHub.

### CI Gates

- [ ] `task verify` serially executes Gate 1 → Gate 2 and other correctness/security checks, failing short circuit.
- [ ] Gate 1 + Gate 2 blocks both PR and main.
- [ ] Performance telemetry can comment or upload artifacts, but cannot be blocked individually based on benchmark numbers.
- [ ] The testdata output by `tools/gen-fixtures-from-formatjs/` passes 100% in the corresponding formatter package (except for known divergence and XFAIL).

### XFAIL

- [ ] `<package>/testdata/xfail.json` schema verification passed (each contains `id`, `reason`, `expires_at`, `tracking_issue`).
- [ ] XFAIL `id` is unique and points to a real fixture; violates block CI.
- [ ] CI automatically fails the corresponding fixture after `expires_at` expires.
- [ ] The total number of XFAILs does not grow in the main branch (monthly review; active scope should → 0 when nearing completion).

### SPEC / Code Drift Checklist

- [ ] Before adding or modifying any exported API, first locate the owner (`Intl` / `Intl.Locale` / `Intl.NumberFormat` / `Intl.DateTimeFormat` / `Intl.PluralRules` / `Intl.ListFormat` / `Intl.RelativeTimeFormat` / `Intl.DurationFormat`) in ECMA-402 or specify it as Go typed bridge; without owner, you are not allowed to enter the public surface.
- [ ] When adding or modifying option / resolved option / part type / range source, check the ECMA-402 field name, allowed value, default value and error boundary simultaneously.
- [ ] When local SPEC conflicts with ECMA-402, change SPEC first and then change the code or fixture.
- [ ] Reference cases that can be mechanically extracted in FormatJS must enter the fixture; non-extractable sources must enter `.skip-list.json` with category.
- [ ] When a generated fixture fails, it can only be processed through implementation repair, `testdata/divergences.md` or `testdata/xfail.json`, and must not be removed from the fixture or written to skip-list.

### Tool chain constraints

- [ ] `go.mod` does not contain `stretchr/testify`, `gkampitakis/go-snaps`, `bradleyjkemp/cupaloy`, `sebdah/goldie`.
- [ ] `go.mod` test depends only on `google/go-cmp` (SPEC 70 only test-only direct dependency).

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
