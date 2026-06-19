id: datetimeformat-formatjs-format-range-test-ts-014
source: formatjs:packages/intl-datetimeformat/tests/format-range.test.ts
owner: datetimeformat
status: accepted
reason: FormatJS marks the shared month prefix as startRange in this formatRangeToParts fixture. Node v26.0.0 marks the common month and literal prefix as shared; go-intl follows that native ECMA-402 range part source behavior.
native_witness: datetimeformat-node-v26-range-shared-month-prefix
review_after: 2026-09-30
removal_path: Resolve if the FormatJS fixture aligns with native range-source semantics, or keep accepted while the FormatJS source remains the only disagreeing reference.

id: datetimeformat-formatjs-index-test-ts-000
source: formatjs:packages/intl-datetimeformat/tests/index.test.ts
owner: datetimeformat
status: accepted
reason: FormatJS expects U+202F narrow no-break spacing before AM/PM with a time-zone name. Node v26.0.0 uses ASCII space; go-intl follows the native witness for day-period and time-zone-name spacing.
native_witness: datetimeformat-node-v26-day-period-time-zone-name-spacing
review_after: 2026-09-30
removal_path: Remove this divergence if FormatJS aligns with native Intl spacing, or if a future Node witness changes the native output.
