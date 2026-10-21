# French compact reference review

Reviewed 2026-09-25 against FormatJS revision
`4bfe08a527d96a511e816a12d896e0ed0548d66e` and the current Go worktree.

| Fixture suffix | Input | FormatJS assertion | go-intl / native witness | Decision |
|---|---:|---|---|---|
| `index-test-ts-040` | 1200000 | `other`, `index.test.ts:236` | `many` | Reference environment difference |
| `index-test-ts-041` | 234500000 | `other`, `index.test.ts:237` | `many` | Reference environment difference |
| `index-test-ts-044` | 1200000 | `other`, `index.test.ts:252` | `many` | Same difference, second source assertion |

All three use locale `fr` and `{notation: "compact"}` without other options.
These were the original generated assertions at the reviewed revision.
They are no longer present after the 2026-10-03 source refresh.

## Why the assertions differ

The reference tests explicitly describe an environment **without NumberFormat
locale data**. Their imports install Locale and canonical-locale polyfills plus
PluralRules test data; the Bazel test target has no NumberFormat polyfill/data
dependency:

- `.references/formatjs/packages/intl-pluralrules/tests/index.test.ts:1` and
  its compact tests at lines 222–253.
- `.references/formatjs/packages/intl-pluralrules/BUILD.bazel`:
  `intl-pluralrules_test` and generated `tests-locale-data-*` targets.
- `.references/formatjs/packages/ecma402-abstract/PluralRules/InitializePluralRules.ts`:
  compact data is populated only from a polyfilled
  `Intl.NumberFormat.localeData`; native Intl does not expose that property.
  The reference also resolves digit options with `standard` notation.
- `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePlural.ts`:
  formats the **source** value, then leaves the compact exponent at zero when
  NumberFormat data is absent. It does not select from a compact display decimal.

With `e=0`, neither source integer is divisible by one million, so the French
rule returns `other`. In go-intl, `pluralrules/pluralrules.go` loads compact
patterns from generated NumberFormat data in the constructor. Both inputs use
the million exponent `e=6`; the pinned French rule's `e != 0..5` branch returns
`many`, independent of these inputs' rounding details. The rule comes from
`tools/gen-cldr/.cldr-json/node_modules/cldr-core/supplemental/plurals.json`
(CLDR 48.1.0), generated into `internal/cldr/plural/cardinal_rules.go`.

The selected contract follows `ResolvePlural` / `PluralRuleSelect` in
`.references/ecma402/spec/pluralrules.html`: notation participates in category
selection. A successful go-intl compact constructor owns its required data;
it does not reproduce the reference test's optional-data fallback. This is an
intentional reference-environment difference, not missing product functionality.
The previous XFAIL reason about compact display decimals was incorrect.

## Independent witness

On 2026-09-25, Node 26.8.1 / V8 14.6.202.34-node.28 / ICU 78.3 / CLDR 48.0 /
tz 2026a returned `many` for both inputs. `resolvedOptions()` reported
`notation: compact`, `compactDisplay: short`, significant digits 1–2 and
`roundingPriority: morePrecision`; the option was not silently ignored.

```js
const rules = new Intl.PluralRules('fr', {notation: 'compact'});
console.log(process.versions);
console.log(rules.resolvedOptions());
for (const input of [1200000, 234500000]) {
  console.log(input, rules.select(input));
}
```

The exact-version fixtures are in
`pluralrules/testdata/conformance/node-v26/compact-review.json`; they supplement
the existing v26.0.0 witnesses and are not labelled as a v26.0.0 replay. FormatJS
was inspected at the pinned source revision, not executed as a built polyfill:
its generated test locale modules are absent from the source checkout.

## Revisit

Keep the three original assertions under explicit accepted divergences while
their test setup deliberately lacks compact data. Recheck on a FormatJS
reference update or by 2026-12-31. Resolve each record if its assertion/setup
aligns with complete compact data; reconsider the product behavior if the
ECMA-402 operation or pinned CLDR rule changes. Never change the extracted
`other` assertions or suppress the native `many` witnesses to make a gate pass.

## Source refresh on 2026-10-03

The current `.references/formatjs/packages/intl-pluralrules/tests/index.test.ts`
now expects `many` for compact selection without ambient NumberFormat data and
adds rounded-magnitude cases. The three original assertions no longer exist;
their ledger records are resolved history, not exemptions for newly numbered
fixtures. Native compact witnesses retain their original version and values.

## Rounded-source determinism

The compact exponent is selected from the integer magnitude of the rounded,
unscaled `FormatNumericToString` result. It is not recomputed by independently
rounding a scaled raw input. Thus French short/long compact values 999499.6 and
999500.4 at maximumFractionDigits=0 both classify as `other`, and their ranges
are order-independent. 999999.6 rounds across the million boundary and selects
`many`. `pluralrules/range_test.go:TestCompactRoundedSourceDeterminesCategory`
records this policy independently of host Node versions, which can disagree at
these thresholds. Existing native snapshots are retained as historical evidence.
