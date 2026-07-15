// Command gen-cldr compiles unicode-org/cldr-json into Go literal source
// under internal/cldr/. Run from any working directory; the generator
// resolves paths relative to the -out flag.
//
// See SPECS/50-cldr-data.md and tools/gen-cldr/README.md for the full
// upgrade flow (task data / task data:check).
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	var cfg Config
	flag.StringVar(&cfg.CLDRDir, "cldr-dir", "", "path to a local cldr-json checkout (npm node_modules root)")
	flag.StringVar(&cfg.OutDir, "out", "../../internal/cldr", "output directory (internal/cldr)")
	flag.StringVar(&cfg.LocaleIDOut, "localeid-out", "../../internal/localeid/unicode_alias_data.go", "output file for generated Unicode type aliases")
	flag.StringVar(&cfg.LocaleMatcherOut, "localematcher-out", "../../internal/localematcher/profile_data.go", "output file for generated language matching profile")
	flag.StringVar(&cfg.TimeZoneOut, "timezone-out", "", "output file for generated IANA time-zone registry")
	flag.StringVar(&cfg.TZDataLock, "tzdata-lock", "", "path to the pinned tzdata lock JSON")
	flag.StringVar(&cfg.TZDataArchive, "tzdata-archive", "", "path to the pinned IANA tzdata tarball")
	flag.StringVar(&cfg.VersionFile, "version-file", "", "path to internal/cldr/VERSION (defaults to <out>/VERSION)")
	flag.StringVar(&cfg.ProfileFile, "profile", "", "path to tools/locale-profile.json")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := Run(context.Background(), cfg, logger); err != nil {
		fmt.Fprintln(os.Stderr, "gen-cldr:", err)
		os.Exit(1)
	}
}
