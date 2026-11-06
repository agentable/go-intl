# Scientific and engineering plural operands

Reviewed 2026-10-15 with Go 1.27.2, generated CLDR 48.1.0, and host
Node v26.10.0 / ICU 78.3 / CLDR 48.0. These host observations are not a refresh
of the repository's active Node v26.0.0 fixture lane.

## Decision

Use the decimal string returned by source-value `FormatNumericToString` for
N/I/V/W/F/T, including its option-induced zeros. Determine C/E with the existing
scientific/engineering exponent operation on that rounded source value. Do not
replace the decimal string with a mantissa, or round that mantissa and restore
its magnitude for classification. Special values keep their existing handling.

This follows ECMA-402 `ResolvePlural`, which passes the source formatted string
to `PluralRuleSelect`. The latter's notation-dependent classification is
implementation-defined. Our CLDR implementation retains the scientific exponent
instead of copying FormatJS's exponent-zero choice. Compact selection is a
separate, unchanged contract described in `compact-review.md`.

## Finite evidence matrix

All rows use cardinal rules, halfExpand, and the shown min/max fraction digits.
`both` means scientific and engineering produced the same observations.
Source, mantissa and candidate categories were calculated with the current
Go digit engine and generated rules. Node categories were executed independently.
The FormatJS column is derived from its source algorithm (source string, E=0),
using these CLDR rules; it is not a claim that its polyfill was executed.

| Locale / input | Notation | min/max | Source string | Notation mantissa / E | Chosen source N / I / V / E | Chosen category | Node category | FormatJS-derived category |
|---|---|---|---|---|---|---|---|---|
| en / 1000 | both | 0/3 | 1000 | 1 / 3 | 1000 / 1000 / 0 / 3 | other | other | other |
| fr / 12345678 | scientific | 0/3 | 12345678 | 1.235 / 7 | 12345678 / 12345678 / 0 / 7 | many | many | other |
| fr / 12345678 | engineering | 0/3 | 12345678 | 12.346 / 6 | 12345678 / 12345678 / 0 / 6 | many | many | other |
| ca / 12345678 | scientific | 0/3 | 12345678 | 1.235 / 7 | 12345678 / 12345678 / 0 / 7 | many | many | other |
| ca / 12345678 | engineering | 0/3 | 12345678 | 12.346 / 6 | 12345678 / 12345678 / 0 / 6 | many | many | other |
| ca / 999999.6 | both | 0/0 | 1000000 | 1 / 6 | 1000000 / 1000000 / 0 / 6 | many | many | many |
| en / 1 | both | 2/3 | 1.00 | 1.00 / 0 | 1.00 / 1 / 2 / 0 | other | other | other |
| ru / 1001 | both | 0/0 | 1001 | 1 / 3 | 1001 / 1001 / 0 / 3 | one | many | one |
| en / 0.001 | both | 2/3 | 0.001 | 1.00 / -3 | 0.001 / 0 / 3 / -3 | other | other | other |
| ar / 0.001 | both | 0/0 | 0 | 1 / -3 | 0 / 0 / 0 / 0 | zero | other | zero |

The prior Go path used the mantissa directly: en 1000 became `one`, French
scientific 12345678 became `one`, and the carry case became `one`. Restoring
the rounded mantissa fixes those examples but does not preserve the source
rounding boundary. It changes ru 1001 to N/I=1000 (`many`); for en 0.001 it
produces 0.00100 with V=5/F=100 instead of V=3/F=1 (both T=1); for ar 0.001
it retains N=0.001 instead of the source-rounded zero. These are the reasons
not to reuse NumberFormat's style-operand restoration for PluralRules.

## Sources and reproduction

- `.references/ecma402/spec/pluralrules.html:ResolvePlural/PluralRuleSelect`
  at `730ca2b`: source formatting precedes implementation-defined classification.
- `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePlural.ts`
  at `0f12c63`: source decimal string, scientific/engineering exponent zero.
- `.references/node/deps/v8/src/objects/js-plural-rules.cc:JSPluralRules::ResolvePlural`
  and `.references/node/deps/icu-small/source/i18n/number_decimalquantity.cpp:getPluralOperand/fractionCount`
  at `c883ccf3b2`: native classification consumes ICU FormattedNumber and restores
  its exponent for N/I and fraction counts.
- `internal/ecma402/numberformat/exponent.go:ScientificExponent` owns exponent
  carry handling; `internal/plural/plural.go:GetOperands` owns decimal operands.
- `numberformat/affix.go:pluralFormattedWithExponent` implements the different
  NumberFormat style contract; it is not needed by this source-string path.

For each Node row, run `new Intl.PluralRules(locale, {notation,
minimumFractionDigits: min, maximumFractionDigits: max}).select(value)`.
No generated native fixture or divergence allowlist is edited to establish
this decision. Public regression tests and operand tests must prove the chosen
source contract independently of the host output.
