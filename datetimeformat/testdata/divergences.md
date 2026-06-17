id: datetimeformat-formatjs-format-range-test-ts-013
source: formatjs:packages/intl-datetimeformat/tests/format-range.test.ts
owner: datetimeformat
status: accepted
reason: FormatJS fixture uses ASCII space before AM/PM. Node v26.0.0 with ICU 78.3 / CLDR 48.0 also uses ASCII spacing; go-intl follows its generated CLDR 48.1 date/time pattern data and emits U+202F narrow no-break space in this range path.
native_witness: datetimeformat-node-v26-range-am-pm-spacing-utc
review_after: 2026-09-30
removal_path: Resolve during the next CLDR/ICU data review by either aligning generated en date/time day-period spacing with the native engine contract or documenting the CLDR 48.1 data-version difference as a permanent implementation-defined boundary.

id: datetimeformat-formatjs-format-range-test-ts-014
source: formatjs:packages/intl-datetimeformat/tests/format-range.test.ts
owner: datetimeformat
status: accepted
reason: FormatJS marks the shared month prefix as startRange in this formatRangeToParts fixture. Node v26.0.0 marks the common month and literal prefix as shared; go-intl follows that native ECMA-402 range part source behavior.
native_witness: datetimeformat-node-v26-range-shared-month-prefix
review_after: 2026-09-30
removal_path: Resolve if the FormatJS fixture aligns with native range-source semantics, or keep accepted while the FormatJS source remains the only disagreeing reference.

id: datetimeformat-formatjs-index-test-ts-002
source: formatjs:packages/intl-datetimeformat/tests/index.test.ts
owner: datetimeformat
status: accepted
reason: FormatJS fixture uses ASCII space before AM/PM. Node v26.0.0 with ICU 78.3 / CLDR 48.0 also uses ASCII spacing; go-intl follows its generated CLDR 48.1 date/time pattern data and emits U+202F narrow no-break space in this range path.
native_witness: datetimeformat-node-v26-range-new-york-am-pm-spacing
review_after: 2026-09-30
removal_path: Resolve during the next CLDR/ICU data review by either aligning generated en date/time day-period spacing with the native engine contract or documenting the CLDR 48.1 data-version difference as a permanent implementation-defined boundary.

id: datetimeformat-node-v26-range-am-pm-spacing-utc
source: node:v26.0.0:datetimeformat:p4-deep-contract
owner: datetimeformat
status: accepted
reason: Node v26.0.0 with ICU 78.3 / CLDR 48.0 uses ASCII space before AM/PM in this native range output; go-intl follows generated CLDR 48.1 date/time pattern data and emits U+202F narrow no-break space in this range path.
native_witness: datetimeformat-node-v26-range-am-pm-spacing-utc
review_after: 2026-09-30
removal_path: Resolve during the next CLDR/ICU data review by either aligning generated en date/time day-period spacing with the native engine contract or documenting the CLDR 48.1 data-version difference as a permanent implementation-defined boundary.

id: datetimeformat-node-v26-range-new-york-am-pm-spacing
source: node:v26.0.0:datetimeformat:p4-deep-contract
owner: datetimeformat
status: accepted
reason: Node v26.0.0 with ICU 78.3 / CLDR 48.0 uses ASCII space before AM/PM in this native New York range output; go-intl follows generated CLDR 48.1 date/time pattern data and emits U+202F narrow no-break space in this range path.
native_witness: datetimeformat-node-v26-range-new-york-am-pm-spacing
review_after: 2026-09-30
removal_path: Resolve during the next CLDR/ICU data review by either aligning generated en date/time day-period spacing with the native engine contract or documenting the CLDR 48.1 data-version difference as a permanent implementation-defined boundary.
