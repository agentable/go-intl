# messageformat-go Dependency Issue Report

Use this template when `kaptinlin/messageformat-go` exposes a go-intl bug, limitation, or ECMA-402 interpretation gap. Follow CLAUDE.md Dependency Issue Reporting and SPEC 61 before adding adapter workarounds.

## Open issues

_No open issues._

## Template

```markdown
### <short title>

- status: open
- dependency: kaptinlin/messageformat-go <version>
- go-intl version: <version or commit>
- problem: <one paragraph describing the issue>
- trigger: <minimal input and options>
- expected: <expected output with ECMA-402 and formatjs reference>
- actual: <actual output, error, or stack trace>
- workaround: <caller-side temporary workaround, if any>
- upstream issue: <go-intl issue URL, if opened>
```

Workarounds must not be implemented as go-intl reimplementations in messageformat-go. If messageformat-go needs behavior from go-intl, report it here instead of forking, silently skipping, or reimplementing ECMA-402 formatter logic in an adapter.
