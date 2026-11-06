# Unicode identity and the x/text boundary

Decision (2026-10-21): narrow `golang.org/x/text/language.Tag` to an explicit
interoperability bridge. It is not the storage or syntax authority for Locale.
The current v0.43.0 tables cannot represent the complete ECMA-402 input domain.

| Input | Required identity | Existing x/text-backed Parse | Decision |
| --- | --- | --- | --- |
| `und-Gara` | `und-Gara` | rejected | accept grammatical unknown script |
| `sjc` | `sjc` | rejected | accept grammatical unknown language |
| `xx-INVALID` | `xx-invalid` | rejected | accept unknown language/variant |
| `abcdefgh` | `abcdefgh` | rejected | accept 5–8 letter language |
| `en-foobar` | `en-foobar` | rejected | accept unknown variant |
| `qaa` | `qaa` | accepted | preserve |
| `no`, `nb` | distinct identities | accepted | preserve |
| `en-Zzzz-ZZ` | explicit unknown script/region | accepted | preserve presence |
| `und-Armn-SU` | `und-Armn-AM` | Parse special case only | generated alias + likely region, including options |
| duplicate variant/singleton | invalid value | rejected | preserve rejection before canonicalization |

`internal/localeid` owns a canonical string tag, the existing ASCII grammar and
Unicode-extension parser, and pinned CLDR alias replacement. Language/script/
region/variant identity is independent of available formatter data. Locale and
DisplayNames consume this owner; DisplayNames still rejects extensions.
The existing CLDR maximizer is supplied for multi-target territory selection,
so no second likely-subtag table or formatter-specific alias path is introduced.

`Locale.Tag() (language.Tag, error)` projects the **complete** Locale identity,
including Unicode extensions. It succeeds only when converting the returned
Tag back through Locale preserves that identity; unrepresentable or lossy
conversion returns an error matching `ErrInvalidValue` with the original value.
`FromTag(tag, Options{})` remains the explicit input bridge. `String`/`Parse`
always provide the complete identity round trip. Callers formatting values do
not need to call `Tag` at all. This intentionally replaces the old infallible,
Unicode-extension-dropping bridge; there is no compatibility alias.

References: `.references/ecma402/spec/locales-currencies-tz.html`
(`IsWellFormedLanguageTag`, `CanonicalizeUnicodeLocaleId`),
`.references/ecma402/spec/locale.html` (`UpdateLanguageId`),
`.references/formatjs/packages/intl-getcanonicallocales/parser.ts` and
`canonicalizer.ts`, `.references/node/deps/icu-small/source/common/locid.cpp`
(`AliasReplacer`). Regression ownership: `locale/language_identity_test.go`,
`locale/construct_test.go`, `internal/localeid/localeid_test.go` and
`internal/ecma402/displaynames_test.go`.
