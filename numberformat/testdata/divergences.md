# Accepted reference differences
# Original FormatJS expectations remain unchanged. These entries skip those callbacks; independent manual adopted-number-patterns.json records execute the library contract.

id: numberformat-formatjs-misc-test-ts-e01356b640dd68c3651b243cef6bc555647f2cf43b4425bd6c220a91a7b5d781
source: formatjs:packages/intl-numberformat/tests/misc.test.ts
owner: numberformat
status: accepted
reason: Reviewed 2026-10-04. en-BS, input 10000, compact percent: FormatJS expects 1M. ECMA PartitionNumberPattern requires the percentSign part. Generated number payloads omit en-BS; NumberFormat/New uses the shared generated written-new best-fit matcher, selects en-GB, and uses pinned en-GB decimal compact m plus the percent pattern. Adopted output is 1m% with percentSign, proved by numberformat-manual-adopted-compact-percent (complete text/parts/resolved JSON). Host Node v26.10.0 / ICU 78.3 / CLDR 48.0 gives en-BS 1M% and en-GB 1m% but marks % as unit; text agreement does not imply part-type agreement. Sources: numberformat/numberformat.go, internal/ecma402/constructor_locale.go, internal/localematcher/compiled.go, tools/locale-profile.json, tools/gen-cldr/.cldr-json/node_modules/cldr-numbers-full/main/en-GB/numbers.json, .references/ecma402/spec/numberformat.html (PartitionNumberPattern).
review_after: 2026-12-31
removal_path: Re-evaluate this exact input if the reference expectation, cited normative contract, or pinned data changes; retain the independent complete public assertions and resolve only when the observations agree.

id: numberformat-formatjs-misc-test-ts-9cbd1d9e79d927b2a1df62aea18511bcc15ea9cc3dc7334a57089233dda91145
source: formatjs:packages/intl-numberformat/tests/misc.test.ts
owner: numberformat
status: accepted
reason: Reviewed 2026-10-04. en compact short, fraction digits 0..2, input 10000000000000000: FormatJS expects 10000T. ECMA [[UseGrouping]] makes positions and presence implementation-defined. The displayed four-digit mantissa groups under pinned en/min2, producing 10,000T. Host Node v26.10.0 / ICU 78.3 / CLDR 48.0 agrees with complete text/parts/resolved JSON. Independent adopted fixture: numberformat-manual-adopted-compact-grouping-10000. Sources: numberformat/notation.go, numberformat/finite.go, .references/ecma402/spec/numberformat.html ([[UseGrouping]]), .references/node/deps/icu-small/source/i18n/number_compact.cpp.
review_after: 2026-12-31
removal_path: Re-evaluate this exact input if the reference expectation, cited normative contract, or pinned data changes; retain the independent complete public assertions and resolve only when the observations agree.

id: numberformat-formatjs-misc-test-ts-6274faf88b574c4356a39a694e4d2324b0d3a555db56033722f3fe720302024f
source: formatjs:packages/intl-numberformat/tests/misc.test.ts
owner: numberformat
status: accepted
reason: Reviewed 2026-10-04. en compact short, fraction digits 0..2, input 55000000000000000: FormatJS expects 55000T. ECMA [[UseGrouping]] makes positions and presence implementation-defined. The displayed five-digit mantissa groups under pinned en/min2, producing 55,000T. Host Node v26.10.0 / ICU 78.3 / CLDR 48.0 agrees with complete text/parts/resolved JSON. Independent adopted fixture: numberformat-manual-adopted-compact-grouping-55000. Sources: numberformat/notation.go, numberformat/finite.go, .references/ecma402/spec/numberformat.html ([[UseGrouping]]), .references/node/deps/icu-small/source/i18n/number_compact.cpp.
review_after: 2026-12-31
removal_path: Re-evaluate this exact input if the reference expectation, cited normative contract, or pinned data changes; retain the independent complete public assertions and resolve only when the observations agree.

id: numberformat-formatjs-misc-test-ts-dc3f2dd7e146f1d083ab6ad725d28421874941cc9bd3c29d3ffcb29ff00f0766
source: formatjs:packages/intl-numberformat/tests/misc.test.ts
owner: numberformat
status: accepted
reason: Reviewed 2026-10-04. pt-PT EUR suffix, signDisplay always, range 2.9 to 3.1: FormatJS expects +2,90 - 3,10 NBSP EUR. Adopted text keeps + on both endpoints and shares the currency suffix with pinned CLDR 48.1.0 separator. ECMA CollapseNumberRange explicitly permits retaining endpoint affixes. Independent fixture numberformat-manual-adopted-signed-range-positive asserts complete text/range parts/sources/resolved JSON: signs are startRange/endRange and currency/NBSP are shared. Host Node v26.10.0 / ICU 78.3 / CLDR 48.0 shares one +; its legal attribution is a different strategy. Original FormatJS fixture only observes text. Sources: numberformat/range.go, .references/ecma402/spec/numberformat.html (CollapseNumberRange), .references/formatjs/packages/ecma402-abstract/NumberFormat/CollapseNumberRange.ts.
review_after: 2026-12-31
removal_path: Re-evaluate this exact input if the reference expectation, cited normative contract, or pinned data changes; retain the independent complete public assertions and resolve only when the observations agree.

id: numberformat-formatjs-misc-test-ts-e3cc5894b4f5da76aae48d71707a416424697bd476806d00177f5de4aa6b1ad0
source: formatjs:packages/intl-numberformat/tests/misc.test.ts
owner: numberformat
status: accepted
reason: Reviewed 2026-10-04. pt-PT EUR suffix, signDisplay always, range -3.1 to -2.9: FormatJS expects -3,10 - 2,90 NBSP EUR. Adopted text keeps - on both endpoints and shares the currency suffix with pinned CLDR 48.1.0 separator. ECMA CollapseNumberRange explicitly permits retaining endpoint affixes. Independent fixture numberformat-manual-adopted-signed-range-negative asserts complete text/range parts/sources/resolved JSON: signs are startRange/endRange and currency/NBSP are shared. Host Node v26.10.0 / ICU 78.3 / CLDR 48.0 shares one -; its legal attribution is a different strategy. Original FormatJS fixture only observes text. Sources: numberformat/range.go, .references/ecma402/spec/numberformat.html (CollapseNumberRange), .references/formatjs/packages/ecma402-abstract/NumberFormat/CollapseNumberRange.ts.
review_after: 2026-12-31
removal_path: Re-evaluate this exact input if the reference expectation, cited normative contract, or pinned data changes; retain the independent complete public assertions and resolve only when the observations agree.
