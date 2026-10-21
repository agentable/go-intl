package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/codegen"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
	"github.com/agentable/go-intl/tools/gen-cldr/tzdb"
	"github.com/agentable/go-intl/tools/internal/localeprofile"
)

// Config controls a generator run.
type Config struct {
	CLDRDir          string
	OutDir           string
	LocaleIDOut      string
	LocaleMatcherOut string
	TimeZoneOut      string
	TZDataLock       string
	TZDataArchive    string
	VersionFile      string
	ProfileFile      string
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
	if cfg.TZDataLock == "" {
		return fmt.Errorf("config: -tzdata-lock is required to generate time-zone display data")
	}
	pin, err := tzdb.ReadPin(cfg.TZDataLock)
	if err != nil {
		return err
	}
	if pin.Version != want.TZData {
		return fmt.Errorf("tzdb pin version %q, want VERSION tzdata %q", pin.Version, want.TZData)
	}
	if cfg.TZDataArchive == "" {
		cfg.TZDataArchive = filepath.Join(filepath.Dir(cfg.TZDataLock), ".tzdata", "tzdata"+pin.Version+".tar.gz")
	}
	aliases, err := tzdb.LoadCLDRTimeZoneAliases(filepath.Join(source.Root, "cldr-bcp47", "bcp47", "timezone.json"))
	if err != nil {
		return err
	}
	registry, err := tzdb.LoadArchive(cfg.TZDataArchive, pin, aliases.Primary)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "tzdb identity source validated", "version", registry.Version, "identifiers", len(registry.Records), "regions", len(registry.Regions))
	manifestExtras := []manifestInputFile{
		{name: "tools/gen-cldr/tzdata.json", path: cfg.TZDataLock},
		{name: "iana/tzdata" + pin.Version + ".tar.gz", path: cfg.TZDataArchive},
	}
	for _, name := range slices.Sorted(maps.Keys(source.NumberRootHashes)) {
		manifestExtras = append(manifestExtras, manifestInputFile{name: name, sha256: source.NumberRootHashes[name]})
	}
	manifest, err := buildManifest(cfg.VersionFile, cfg.ProfileFile, source.Root, want, profile, manifestExtras...)
	if err != nil {
		return err
	}

	locales := extract.ExtractLocales(source.Available)
	likely, err := extract.ExtractLikelySubtags(source.LikelySubtags)
	if err != nil {
		return fmt.Errorf("extract likely subtags: %w", err)
	}
	numbers := extract.ExtractNumbers(source.Numbers, source.Available)
	currencies := extract.ExtractCurrencies(source.CurrencyFractions, source.Currencies, source.Available)
	dates := extract.ExtractDates(source.Dates, source.Available)
	metazones := extract.ExtractMetazones(source.Metazones, source.Available)
	displayKeys, err := timeZoneDisplayKeys(metazones, registry, aliases.Display)
	if err != nil {
		return err
	}
	units := extract.ExtractUnits(source.Units, source.Available)
	listPatterns := extract.ExtractListPatterns(source.ListPatterns, source.Available)
	relativeTime := extract.ExtractRelativeTimeFields(source.RelativeTime, source.Available)
	displayNames := extract.ExtractDisplayNames(source.DisplayNames, source.Available)
	input := codegen.RuntimeInput{
		Manifest:            manifest,
		Locales:             locales,
		LikelySubtags:       likely,
		ScriptDirections:    source.ScriptDirections,
		Numbers:             numbers,
		Currencies:          currencies,
		Dates:               dates,
		Preferences:         source.Preference,
		Metazones:           metazones,
		TimeZoneDisplayKeys: displayKeys,
		Units:               units,
		ListPatterns:        listPatterns,
		RelativeTime:        relativeTime,
		DisplayNames:        displayNames,
	}
	if err := validateProfileRoutes(source.Root, profile.Locales, input); err != nil {
		return err
	}
	if err := codegen.RenderRuntime(cfg.OutDir, input); err != nil {
		return fmt.Errorf("render runtime data: %w", err)
	}
	if cfg.LocaleIDOut != "" {
		if err := codegen.RenderUnicodeTypeAliases(cfg.LocaleIDOut, source.UnicodeTypeAliases, source.IdentityAliases); err != nil {
			return fmt.Errorf("render Unicode type aliases: %w", err)
		}
	}
	if cfg.LocaleMatcherOut != "" {
		if err := codegen.RenderLanguageMatchingProfile(cfg.LocaleMatcherOut, source.LanguageMatching); err != nil {
			return fmt.Errorf("render language matching profile: %w", err)
		}
	}
	if cfg.TimeZoneOut != "" {
		if err := codegen.RenderTimeZoneRegistry(cfg.TimeZoneOut, registry); err != nil {
			return fmt.Errorf("render time-zone registry: %w", err)
		}
	}
	return nil
}

func validateProfileRoutes(root string, locales []string, input codegen.RuntimeInput) error {
	for _, locale := range locales {
		for _, domain := range []struct {
			name  string
			found bool
		}{
			{"number", hasLocalePayload(input.Numbers, locale)},
			{"currency", hasLocalePayload(input.Currencies.Currencies, locale)},
			{"date", hasLocalePayload(input.Dates, locale)},
			{"timezone", hasLocalePayload(input.Metazones.Formats, locale)},
			{"unit", hasLocalePayload(input.Units, locale)},
			{"list", hasLocalePayload(input.ListPatterns, locale)},
			{"relativetime", hasLocalePayload(input.RelativeTime, locale)},
			{"displaynames", hasLocalePayload(input.DisplayNames, locale)},
		} {
			if !domain.found {
				return fmt.Errorf("locale profile %q: %s has no loaded lookup payload in %s", locale, domain.name, root)
			}
		}
	}
	return nil
}

func hasLocalePayload[T any](payload map[string]T, locale string) bool {
	for locale != "" {
		if _, ok := payload[locale]; ok {
			return true
		}
		index := strings.LastIndexByte(locale, '-')
		if index < 0 {
			return false
		}
		locale = locale[:index]
	}
	return false
}

type manifestInputFile struct {
	name   string
	path   string
	sha256 string
}

func buildManifest(versionFile, profileFile, cldrRoot string, versions cldr.Versions, profile localeprofile.Profile, extras ...manifestInputFile) (codegen.ManifestInput, error) {
	inputFiles := []manifestInputFile{
		{name: "internal/cldr/VERSION", path: versionFile},
		{name: "tools/locale-profile.json", path: profileFile},
	}
	inputFiles = append(inputFiles, extras...)
	packages := cldr.RequiredPackages()
	hashes := make([]codegen.ManifestHash, len(inputFiles)+len(packages))
	for i, file := range inputFiles {
		hash := file.sha256
		if hash == "" {
			var err error
			hash, err = fileSHA256(file.path)
			if err != nil {
				return codegen.ManifestInput{}, err
			}
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
