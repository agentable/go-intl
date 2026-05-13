id: datetimeformat-formatjs-format-range-test-ts-013
source: formatjs:packages/intl-datetimeformat/tests/format-range.test.ts
status: accepted
reason: FormatJS fixture uses ASCII space before AM/PM; go-intl follows the CLDR pattern data and emits U+202F narrow no-break space.
review_after: 2026-11-01

id: datetimeformat-formatjs-format-range-test-ts-014
source: formatjs:packages/intl-datetimeformat/tests/format-range.test.ts
status: accepted
reason: FormatJS marks the shared month prefix as startRange in this formatRangeToParts fixture; go-intl marks the common month and literal prefix as shared, matching ECMA-402 range part source semantics.
review_after: 2026-11-01

id: datetimeformat-formatjs-index-test-ts-002
source: formatjs:packages/intl-datetimeformat/tests/index.test.ts
status: accepted
reason: FormatJS fixture uses ASCII space before AM/PM; go-intl follows the CLDR pattern data and emits U+202F narrow no-break space.
review_after: 2026-11-01
