package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/codegen"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
	"github.com/agentable/go-intl/tools/internal/localeprofile"
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
	profile, err := localeprofile.Read(cfg.ProfileFile)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "locale profile loaded", "locales", len(profile.Locales))

	if cfg.CLDRDir == "" {
		return fmt.Errorf("config: -cldr-dir is required (point at a local cldr-json checkout)")
	}
	source, err := cldr.LoadAll(ctx, cfg.CLDRDir, want, profile.Locales)
	if err != nil {
		return fmt.Errorf("load cldr-json: %w", err)
	}
	log.InfoContext(ctx, "cldr-json checkout validated", "dir", source.Root)
	manifest, err := buildManifest(cfg.VersionFile, cfg.ProfileFile, source.Root, want, profile)
	if err != nil {
		return err
	}

	locales := extract.ExtractLocales(source.Available)
	likely := extract.ExtractLikelySubtags(source.LikelySubtags)
	numbers := extract.ExtractNumbers(source.Numbers, profile.Locales)
	currencies := extract.ExtractCurrencies(source.CurrencyFractions, source.Currencies, profile.Locales)
	dates := extract.ExtractDates(source.Dates, profile.Locales)
	metazones := extract.ExtractMetazones(source.Metazones, profile.Locales)
	units := extract.ExtractUnits(source.Units, profile.Locales)
	listPatterns := extract.ExtractListPatterns(source.ListPatterns, profile.Locales)
	relativeTime := extract.ExtractRelativeTimeFields(source.RelativeTime, profile.Locales)
	displayNames := extract.ExtractDisplayNames(source.DisplayNames, profile.Locales)
	input := codegen.RuntimeInput{
		Manifest:      manifest,
		Locales:       locales,
		LikelySubtags: likely,
		Numbers:       numbers,
		Currencies:    currencies,
		Dates:         dates,
		Preferences:   source.Preference,
		Metazones:     metazones,
		Units:         units,
		ListPatterns:  listPatterns,
		RelativeTime:  relativeTime,
		DisplayNames:  displayNames,
	}
	if err := codegen.RenderRuntime(cfg.OutDir, input); err != nil {
		return fmt.Errorf("render runtime data: %w", err)
	}
	return nil
}

func buildManifest(versionFile, profileFile, cldrRoot string, versions cldr.Versions, profile localeprofile.Profile) (codegen.ManifestInput, error) {
	inputFiles := [...]struct {
		name string
		path string
	}{
		{name: "internal/cldr/VERSION", path: versionFile},
		{name: "tools/locale-profile.json", path: profileFile},
	}
	packages := cldr.RequiredPackages()
	hashes := make([]codegen.ManifestHash, len(inputFiles)+len(packages))
	for i, file := range inputFiles {
		hash, err := fileSHA256(file.path)
		if err != nil {
			return codegen.ManifestInput{}, err
		}
		hashes[i] = codegen.ManifestHash{Name: file.name, SHA256: hash}
	}
	for i, name := range packages {
		hash, err := fileSHA256(filepath.Join(cldrRoot, name, "package.json"))
		if err != nil {
			return codegen.ManifestInput{}, err
		}
		hashes[len(inputFiles)+i] = codegen.ManifestHash{Name: filepath.ToSlash(filepath.Join("cldr-json", name, "package.json")), SHA256: hash}
	}
	return codegen.ManifestInput{
		Generator:     "tools/gen-cldr",
		CLDR:          versions.CLDR,
		ICU:           versions.ICU,
		TZData:        versions.TZData,
		LocaleProfile: profile.Locales,
		InputHashes:   hashes,
	}, nil
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}
