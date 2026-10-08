# go-intl/x — Experimental modules

`github.com/agentable/go-intl/x` is the experimental satellite module of
[go-intl](https://github.com/agentable/go-intl). It hosts packages that are
useful today but do not belong to the main module's public surface:

- Go bridges for **in-flight TC39 proposals** (Stage 1–3), whose API shape can
  still change.
- **Non-surface exposure of finished-spec machinery** (ECMA-402 abstract
  operations) for CLDR-driven consumers outside `Intl` constructors.

## Stability

The module stays **v0 with no compatibility promise**. Any release may add,
change, or remove packages and APIs. Pin an exact version and read the release
notes before upgrading.

The main module is unaffected: every exported symbol there still maps to a
native ECMA-402 owner. Nothing in `x/` relaxes that rule.

## Versioning

`x/` joins the coordinated go-intl release set: it is tagged `x/vX.Y.Z` at the
same commit and with the same version as the root module and the `tools/*`
modules, and its `go.mod` requires the matching root-module version.

## Graduation

A package leaves `x/` when its specification target finalizes and the
implementation is complete:

- A TC39 proposal that reaches Stage 4 may be promoted into the main module as
  a constructor package with a root namespace alias; the `x/` package is then
  deprecated and removed in a later v0 release.
- A package that exposes finished-spec machinery (such as locale matching) may
  graduate the same way if its proposal finalizes, or remain in `x/`
  long-term.

## Packages

| Package | Use |
|---------|-----|
| `github.com/agentable/go-intl/x/localematcher` | Match requested locales against an application's available locales with the same ECMA-402 lookup/best-fit algorithms and CLDR data that back the `Intl` constructors. Follows the Stage 1 [`Intl.LocaleMatcher`](https://github.com/tc39/proposal-intl-localematcher) proposal. |

### localematcher example

```go
package main

import (
	"fmt"

	"github.com/agentable/go-intl/x/localematcher"
)

func main() {
	// One-shot, mirrors Intl.LocaleMatcher.match.
	result := localematcher.Match(
		[]string{"zh-HK"},
		[]string{"en", "zh-Hans", "zh-Hant", "ja"},
		"en",
		localematcher.AlgorithmBestFit,
	)
	fmt.Println(result.Locale) // zh-Hant

	// Compiled once, reused per message key in an i18n catalog.
	matcher := localematcher.New([]string{"en", "zh-Hans", "zh-Hant", "ja"})
	fmt.Println(matcher.Match([]string{"zh-SG"}, "en", localematcher.AlgorithmBestFit).Locale) // zh-Hans
}
```

## Governance

`SPECS/80-x-modules.md` in the main repository owns the module charter:
admission criteria, versioning, graduation, and removal rules.
