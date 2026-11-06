# Experimental locale matching

`github.com/agentable/go-intl/x/localematcher` provides application resource
matching based on the official TC39 Stage 1
[`Intl.LocaleMatcher` draft](https://github.com/tc39/proposal-intl-localematcher/blob/c35e7b759f698b16e9d73d8e98647207e35a14ce/spec.emu).
It ships in the **root go-intl module**, at the root module's version. The API
may change as the proposal evolves; pin a version before upgrading.

```go
available, err := locale.ParseList("en", "zh-Hans", "zh-Hant", "ja")
if err != nil {
    return err
}
matcher, err := localematcher.New(available, localematcher.Options{})
if err != nil {
    return err
}
requested, err := locale.ParseList("zh-HK")
if err != nil {
    return err
}
if matched, ok := matcher.Match(requested); ok {
    fmt.Println(matched.String()) // zh-Hant
}
```

Import `locale` and `x/localematcher` from `github.com/agentable/go-intl`.
`Options{}` selects best fit; use `Algorithm: gointl.String("lookup")` for
lookup. Available locales containing Unicode `u` extensions are rejected.
Matching returns only a canonical locale from your available set. It adds no
parent resources or aliases to that set, drops requested Unicode extensions,
and leaves default fallback to your application.

There is no separate `x/go.mod`, release tag, internal-type re-export, or root
namespace alias. Existing formatter constructors keep their own available-data
expansion and default-locale behavior. See [SPEC 80](../SPECS/80-x-modules.md)
for the draft contract and verification evidence.
