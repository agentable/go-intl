package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentable/go-intl/tools/internal/localeprofile"
)

const outputDir = ".tmp/sizecheck"

var (
	errMissingDataVersion = errors.New("missing cldr/icu/tzdata pin")
)

type binaryCase struct {
	name   string
	source string
}

type result struct {
	name     string
	bytes    int64
	duration time.Duration
}

type dataProfile struct {
	CLDR        string
	ICU         string
	TZData      string
	LocaleCount int
	Cold        bool
}

func main() {
	os.Exit(runMain(os.Args[1:], os.Stdout, os.Stderr))
}

func runMain(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("sizecheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	cold := fs.Bool("cold", false, "clear the Go build cache before measuring compile time")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		if _, err := fmt.Fprintf(stderr, "unknown argument %q\n", fs.Arg(0)); err != nil {
			return 1
		}
		return 2
	}
	if err := runWithDependencies(context.Background(), outputDir, runOptions{Root: ".", Cold: *cold}, defaultRunDependencies(stdout)); err != nil {
		if _, printErr := fmt.Fprintln(stderr, err); printErr != nil {
			return 1
		}
		return 1
	}
	return 0
}

type runOptions struct {
	Root string
	Cold bool
}

type runDependencies struct {
	cases           []binaryCase
	cleanBuildCache func(context.Context) error
	stdout          io.Writer
}

func defaultRunDependencies(stdout io.Writer) runDependencies {
	return runDependencies{
		cases:           sizeCases(),
		cleanBuildCache: cleanBuildCache,
		stdout:          stdout,
	}
}

func run(ctx context.Context, dir string, opts runOptions) error {
	return runWithDependencies(ctx, dir, opts, defaultRunDependencies(os.Stdout))
}

func runWithDependencies(ctx context.Context, dir string, opts runOptions, deps runDependencies) error {
	if deps.cleanBuildCache == nil {
		deps.cleanBuildCache = cleanBuildCache
	}
	if deps.stdout == nil {
		deps.stdout = os.Stdout
	}
	if deps.cases == nil {
		deps.cases = sizeCases()
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	profile, err := readDataProfile(opts.Root)
	if err != nil {
		return err
	}
	profile.Cold = opts.Cold
	if opts.Cold {
		if err := deps.cleanBuildCache(ctx); err != nil {
			return err
		}
	}
	caseDir := filepath.Join(dir, "cases")
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		return err
	}

	results := make([]result, len(deps.cases))
	for i, tc := range deps.cases {
		pkgDir, err := writeCase(caseDir, tc)
		if err != nil {
			return err
		}
		result, err := buildCase(ctx, tc.name, pkgDir, filepath.Join(binDir, tc.name))
		if err != nil {
			return err
		}
		results[i] = result
	}
	if _, err := fmt.Fprint(deps.stdout, formatResults(profile, results)); err != nil {
		return err
	}
	return nil
}

type commandContextFunc func(context.Context, string, ...string) *exec.Cmd

func cleanBuildCache(ctx context.Context) error {
	return cleanBuildCacheWithCommand(ctx, exec.CommandContext)
}

func cleanBuildCacheWithCommand(ctx context.Context, command commandContextFunc) error {
	cmd := command(ctx, "go", "clean", "-cache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clean build cache: %w\n%s", err, output)
	}
	return nil
}

func writeCase(root string, tc binaryCase) (string, error) {
	dir := filepath.Join(root, tc.name)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(tc.source)+"\n"), 0o600); err != nil {
		return "", err
	}
	return dir, nil
}

func buildCase(ctx context.Context, name, pkgDir, out string) (result, error) {
	start := time.Now()
	cmd := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", out, "./"+filepath.ToSlash(pkgDir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result{}, fmt.Errorf("build %s: %w\n%s", name, err, output)
	}
	info, err := os.Stat(out)
	if err != nil {
		return result{}, err
	}
	return result{name: name, bytes: info.Size(), duration: time.Since(start).Round(time.Millisecond)}, nil
}

func formatResults(profile dataProfile, results []result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "data-profile: locales=%d cldr=%s icu=%s tzdata=%s cold=%t\n", profile.LocaleCount, profile.CLDR, profile.ICU, profile.TZData, profile.Cold)
	var baseline int64
	for _, r := range results {
		if r.name == "empty" {
			baseline = r.bytes
			break
		}
	}
	fmt.Fprintf(&b, "%-22s %12s %12s %12s\n", "case", "bytes", "delta", "build")
	for _, r := range results {
		fmt.Fprintf(&b, "%-22s %12d %12d %12s\n", r.name, r.bytes, r.bytes-baseline, r.duration)
	}
	return b.String()
}

func readDataProfile(root string) (dataProfile, error) {
	version, err := readVersionFile(filepath.Join(root, "internal", "cldr", "VERSION"))
	if err != nil {
		return dataProfile{}, err
	}
	count, err := readLocaleProfileCount(filepath.Join(root, "tools", "locale-profile.json"))
	if err != nil {
		return dataProfile{}, err
	}
	version.LocaleCount = count
	return version, nil
}

func readVersionFile(path string) (dataProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return dataProfile{}, fmt.Errorf("read version file %s: %w", path, err)
	}
	var profile dataProfile
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		switch key {
		case "cldr":
			profile.CLDR = value
		case "icu":
			profile.ICU = value
		case "tzdata":
			profile.TZData = value
		}
	}
	if profile.CLDR == "" || profile.ICU == "" || profile.TZData == "" {
		return dataProfile{}, fmt.Errorf("read version file %s: %w", path, errMissingDataVersion)
	}
	return profile, nil
}

func readLocaleProfileCount(path string) (int, error) {
	profile, err := localeprofile.Read(path)
	if err != nil {
		return 0, err
	}
	return len(profile.Locales), nil
}

func sizeCases() []binaryCase {
	return []binaryCase{
		{
			name: "empty",
			source: `package main

func main() {}`,
		},
		{
			name: "cldr-available",
			source: `package main

import cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"

func main() {
	_ = cldrlocale.AvailableLocales()
}`,
		},
		{
			name: "root-facade",
			source: `package main

import intl "github.com/agentable/go-intl"

func main() {
	_ = intl.SupportedCalendars()
}`,
		},
		directFormatterCase("numberformat", "numberformat"),
		directFormatterCase("datetimeformat", "datetimeformat"),
		directFormatterCase("pluralrules", "pluralrules"),
		directFormatterCase("listformat", "listformat"),
		directFormatterCase("relativetimeformat", "relativetimeformat"),
		directFormatterCase("durationformat", "durationformat"),
		directFormatterCase("displaynames", "displaynames"),
		{
			name: "all-formatters",
			source: `package main

import (
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
)

func main() {
	_, _ = numberformat.New(nil, numberformat.Options{})
	_, _ = datetimeformat.New(nil, datetimeformat.Options{})
	_, _ = pluralrules.New(nil, pluralrules.Options{})
	_, _ = listformat.New(nil, listformat.Options{})
	_, _ = relativetimeformat.New(nil, relativetimeformat.Options{})
	_, _ = durationformat.New(nil, durationformat.Options{})
	_, _ = displaynames.New(nil, displaynames.Options{})
}`,
		},
	}
}

func directFormatterCase(name, pkg string) binaryCase {
	return binaryCase{
		name: name,
		source: fmt.Sprintf(`package main

import "github.com/agentable/go-intl/%s"

func main() {
	_, _ = %s.New(nil, %s.Options{})
}`, pkg, pkg, pkg),
	}
}
