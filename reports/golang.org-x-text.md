# golang.org/x/text Collation Backend Report

Dependency: `golang.org/x/text` v0.37.0

## Trigger

M5 evaluates whether the active `collator` backend can support the ECMA-402 `Intl.Collator` behavior that Node v26 advertises for `usage=search`, `caseFirst=upper|lower`, and explicit collation tailorings such as German phonebook.

## Expected Behavior

Before `go-intl` advertises or accepts a Collator option, the backend must apply the observable behavior behind that option:

- `usage=search` must use search collation data rather than pretending sort collation is search collation.
- `caseFirst=upper|lower` must affect ordering and resolved options.
- explicit `collation` / `co` values must apply the requested CLDR tailoring.

## Actual Behavior

The active `x/text/collate` integration can cover base sort behavior, numeric comparison, sensitivity, and alternate-shifted punctuation handling. It does not expose a stable public path in this project for ECMA-402 search tailoring, case-first direction, or arbitrary explicit collation tailorings.

## Evidence

- Passing Node v26 fixtures cover behavior the backend can truthfully perform, including numeric ordering and Swedish default collation ordering.
- XFAIL Node v26 fixtures in `collator/testdata/conformance/node-v26/options.json` capture unsupported `usage=search`, `caseFirst`, and German phonebook contracts before any backend expansion is accepted.
- `SPECS/45-collator.md` keeps these rows in a narrowed implementation gap rather than advertising false support.

## Suggested Workaround

Keep the current constructor rejections and withhold `Intl.supportedValuesOf("collation")` values until either `x/text/collate` can be mapped to those ECMA-402 semantics or a replacement/generated backend proves the behavior with fixtures, size evidence, and benchmark evidence.
