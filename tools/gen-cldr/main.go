// Command gen-cldr compiles unicode-org/cldr-json into Go literal source
// under internal/cldr/. Run from any working directory; the generator
// resolves paths relative to the -out flag.
//
// See SPECS/50-cldr-data.md and tools/gen-cldr/README.md for the full
// upgrade flow (task data:update / data:diff / data:verify).
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
	flag.StringVar(&cfg.VersionFile, "version-file", "", "path to internal/cldr/VERSION (defaults to <out>/VERSION)")
	flag.StringVar(&cfg.ProfileFile, "profile", "", "path to tools/locale-profile.json")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := Run(context.Background(), cfg, logger); err != nil {
		fmt.Fprintln(os.Stderr, "gen-cldr:", err)
		os.Exit(1)
	}
}
