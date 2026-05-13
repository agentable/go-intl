package main

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/codegen"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// Config controls a generator run.
type Config struct {
	CLDRDir     string
	OutDir      string
	VersionFile string
	ProfileFile string
}

// Run drives the CLDR extract and emit pipeline.
func Run(ctx context.Context, cfg Config, log *slog.Logger) error {
	if cfg.OutDir == "" {
		return fmt.Errorf("config: OutDir is required")
	}
	if cfg.VersionFile == "" {
		cfg.VersionFile = filepath.Join(cfg.OutDir, "VERSION")
	}
	if cfg.ProfileFile == "" {
		return fmt.Errorf("config: -profile is required")
	}

	want, err := cldr.ReadVersionFile(cfg.VersionFile)
	if err != nil {
		return fmt.Errorf("read pinned version: %w", err)
	}
	log.InfoContext(ctx, "pinned versions", "cldr", want.CLDR, "icu", want.ICU, "tzdata", want.TZData)
	profile, err := readLocaleProfile(cfg.ProfileFile)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "locale profile loaded", "locales", len(profile.Locales), "formatterLocales", len(profile.FormatterLocales), "numberLocales", len(profile.NumberLocales), "pluralLocales", len(profile.PluralLocales))

	if cfg.CLDRDir == "" {
		return fmt.Errorf("config: -cldr-dir is required (point at a local cldr-json checkout)")
	}
	source, err := cldr.LoadAll(ctx, cfg.CLDRDir, want, profile.Locales)
	if err != nil {
		return fmt.Errorf("load cldr-json: %w", err)
	}
	log.InfoContext(ctx, "cldr-json checkout validated", "dir", source.Root)

	locales := extract.ExtractLocales(source.Available, profile.Locales)
	likely := extract.ExtractLikelySubtags(source.LikelySubtags)
	numbers := extract.ExtractNumbers(source.Numbers, profile.NumberLocales)
	currencies := extract.ExtractCurrencies(source.CurrencyFractions, source.Currencies, profile.NumberLocales)
	matching := extract.ExtractLocaleMatching(source.LanguageMatching, source.Regions)
	dates := extract.ExtractDates(source.Dates, profile.FormatterLocales)
	preference := extract.ExtractPreference(source.Preference)
	metazones := extract.ExtractMetazones(source.Metazones, profile.FormatterLocales)
	units := extract.ExtractUnits(source.Units, profile.FormatterLocales)
	if err := codegen.RenderPhase3(cfg.OutDir, codegen.Phase3Data{Locales: locales, Likely: likely, Numbers: numbers, Currencies: currencies, Collations: source.Collations, Matching: matching, Dates: dates, Preference: preference, Metazones: metazones, Units: units}); err != nil {
		return fmt.Errorf("render phase 3 data: %w", err)
	}
	return nil
}
