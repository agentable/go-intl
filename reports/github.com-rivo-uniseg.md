# github.com/rivo/uniseg Segmenter Backend Report

Dependency: `github.com/rivo/uniseg` v0.4.7

## Trigger

M5 evaluates whether the active `segmenter` backend can support Node v26 `Intl.Segmenter` word-boundary behavior for dictionary and CJK-tailored locales such as Thai, Japanese, and Traditional Chinese.

## Expected Behavior

Before `go-intl` advertises a Segmenter locale, `SupportedLocalesOf` should mean the active backend can produce trustworthy grapheme, word, and sentence boundaries for that locale. For word segmentation, Thai requires dictionary segmentation and Japanese / Chinese require CJK-tailored boundaries.

## Actual Behavior

`uniseg` provides Unicode Standard Annex #29 grapheme, word, and sentence boundary algorithms, but the active integration does not provide dictionary or CJK tailoring. It can truthfully support the existing allowlist covered by Node fixtures, but it should not advertise `ja`, `th`, `zh`, `zh-Hans`, or `zh-Hant`.

## Evidence

- Passing Node v26 fixtures in `segmenter/testdata/conformance/node-v26/locale-contract.json` cover every locale currently returned by `internal/segmentation.SupportedLocales()`.
- XFAIL Node v26 fixtures in `segmenter/testdata/conformance/node-v26/tailored-locale-contract.json` capture Thai, Japanese, and Traditional Chinese word-boundary behavior that must be satisfied before those locales are advertised.
- `internal/segmentation.SupportedLocales()` remains an explicit allowlist and returns a snapshot.

## Suggested Workaround

Keep dictionary and CJK locales out of `SupportedLocalesOf` until a backend with tailored segmentation is selected or generated. Expand the allowlist only in the same change that makes the tailored Node fixtures pass and records any size/performance impact.
