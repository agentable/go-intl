# Code Surface-Lock Pruning

Detection ran against all Go test files with the three Go signal families. The
shape family returned two textual false positives. The dependency/boundary
family returned 17 hits. The error-mechanics family returned 438 hits, of which
419 are `filepath.Join` or `strings.Join` call sites. The rows below classify
the complete detected set before deletion.

| File:line | Test/check | Signal | Claimed invariant | Observable contract | Classification | Follow-up owner |
|---|---|---|---|---|---|---|
| `internal/tz/resolve_test.go:231` | host-dependent skip message | `does not expose` text | None | Test behavior is unrelated to exported-member shape | `false-positive` | None |
| `internal/ecma402/datetimeformat/skeleton_test.go:256` | quarter-support comment | `does not expose` text | None | Test behavior is unrelated to exported-member shape | `false-positive` | None |
| `listformat/listformat_test.go:119,210` | list boundary-length tests | `Boundary` in test name | Empty, singleton, pair, and many-element formatting and parts | Public `Format` and `FormatToParts` results at list-length transitions | `keep-runtime` | `listformat` |
| `tools/gen-fixtures-from-formatjs/main_test.go:15` | `TestRunImportsSyntheticNodeWitnessFixtures` | `Imports` in test name | End-to-end Node witness import | Generated fixture files contain the witnessed outputs and omit the skip list | `keep-runtime` | fixture importer |
| `tools/gen-fixtures-from-formatjs/main_test.go:435,482,504,555,583,604,626,652,674,696,2186` | FormatJS import tests and one matched importer call | `Import` token | Fixture extraction, filtering, skip recording, stale-file cleanup, errors, and generated outputs | Filesystem outputs and errors produced by the importer | `keep-runtime` | fixture importer |
| `tools/gen-fixtures-from-formatjs/main_test.go:913` | `TestFormatJSSurfaceRoutesPreserveImportOrder` | `Import` in test name | Exact membership, order, and fields of two private route slices | None; output fixtures and sorted skip records are covered by end-to-end importer tests, and no production caller mutates the private slices | `delete-surface` | None; delete its exclusive assertion helper |
| `internal/cldr/displaynames/displaynames_test.go:78` | language-region fallback boundary | `Boundary` in test name | Boolean fallback changes lookup success and returned text | `Of` returns `English (QQ)` only when code fallback is enabled | `keep-runtime` | display-name lookup |
| `tools/gen-cldr/cldr/date_pattern_test.go:33` | executable skeleton boundary | `Boundary` in test name | Supported fields pass and unsupported week/quarter fields fail | Generator acceptance result for concrete skeleton inputs | `keep-runtime` | CLDR generator |
| `relativetimeformat/relativetimeformat_test.go:478`; `durationformat/durationformat_test.go:672`; `listformat/listformat_test.go:161`; `datetimeformat/datetimeformat_test.go:642,1775,1935`; `numberformat/format_test.go:50`; `numberformat/range_test.go:161` | formatter text/parts consistency tests | `Join` in test name | ECMA-402 string methods are concatenations of their parts methods, including ranges and fallback paths | Public formatted text equals the values returned by the public parts APIs | `keep-runtime` | owning formatter packages |
| `internal/localeid/localeid_test.go:12,19,20` | locale ID parts and join | `Join` token | Locale identity reconstruction ignores empty components and preserves extension parts | Concrete internal runtime result used by locale serialization | `keep-runtime` | locale ID kernel |
| `tools/gen-cldr/dates_test.go:101` | literal-chunk comment | `Join` text | None | Comment describes fixture construction; it asserts no error mechanics | `false-positive` | None |
| `intl_test.go:125`; `internal/ecma402/errors_test.go:10,39`; `internal/intlerr/errors_test.go:48,143,171,314` | structured error and sentinel classification tests | `Sentinel` / `WrapsSentinel` in test names | Public and internal constructors expose documented custom kind, context, cause, and `errors.Is`/`As` behavior | Callers can classify actual formatter failures and inspect structured error detail | `keep-runtime` | error contract owners |
| 419 call sites in the 42 files listed below | path/string construction inside test setup and assertions | `Join` token | None | Ordinary runtime path/text construction; these do not test `errors.Join` or API shape | `false-positive` | None |

## Mechanical `Join` False Positives

The final grouped row covers every detector hit whose matched expression is
`filepath.Join` or `strings.Join`. Counts by file are retained so re-detection
is auditable without turning 419 setup expressions into individual policy rows:

```text
  3 internal/cldr/locale/snapshot_test.go
  1 internal/localeid/localeid_test.go
  6 tools/check-conformance/main_test.go
  3 tools/check-generated-data/compare_test.go
 24 tools/conformance/coverage_test.go
  2 tools/conformance/node_witness_test.go
  2 tools/conformance/product_contract_test.go
  9 tools/conformance/skip_test.go
  5 tools/data-preflight/preflight_test.go
  3 tools/gen-cldr/cldr/bcp47_test.go
  4 tools/gen-cldr/cldr/dates_test.go
 15 tools/gen-cldr/cldr/displaynames_test.go
  1 tools/gen-cldr/cldr/fetch_test.go
  4 tools/gen-cldr/cldr/files_test.go
  5 tools/gen-cldr/cldr/language_matching_test.go
  2 tools/gen-cldr/cldr/list_patterns_test.go
  6 tools/gen-cldr/cldr/metazones_test.go
 12 tools/gen-cldr/cldr/numbers_test.go
 12 tools/gen-cldr/cldr/preference_test.go
  5 tools/gen-cldr/cldr/relative_time_test.go
  2 tools/gen-cldr/cldr/script_metadata_test.go
  7 tools/gen-cldr/cldr/source_test.go
 10 tools/gen-cldr/cldr/units_test.go
  5 tools/gen-cldr/cldr/version_test.go
  2 tools/gen-cldr/codegen/render_test.go
  3 tools/gen-cldr/codegen/roundtrip_source_test.go
  1 tools/gen-cldr/codegen/unit_roundtrip_test.go
 27 tools/gen-cldr/dates_test.go
  9 tools/gen-cldr/list_patterns_test.go
 34 tools/gen-cldr/locales_test.go
 14 tools/gen-cldr/numbers_test.go
  4 tools/gen-cldr/regenerate_idempotent_test.go
  9 tools/gen-cldr/relative_time_test.go
 11 tools/gen-cldr/run_test.go
  6 tools/gen-cldr/source_test.go
 16 tools/gen-cldr/timezones_test.go
 14 tools/gen-cldr/units_test.go
 72 tools/gen-fixtures-from-formatjs/main_test.go
 25 tools/gen-plural-rules/main_test.go
  2 tools/internal/localeprofile/profile_test.go
  2 tools/node-witness/main_test.go
 20 tools/sizecheck/main_test.go
```

## Pruning Verdict

Delete only the private route-table inventory and its exclusive assertion
helper. The conformance contract is owned by generated fixture contents,
skip-list records, source integrity, idempotence, and sorted audit output, all
of which retain executable coverage. No production change or replacement
policy gate is justified by this pruning verdict.
