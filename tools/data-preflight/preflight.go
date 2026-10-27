package main

import (
	"encoding/json/v2"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/internal/datapin"
)

type preflightConfig struct {
	versionFile  string
	packageFile  string
	tzLockFile   string
	goUpdateFile string
}

type dataPins struct {
	cldr   string
	tzdata string
}

type tzDataLock = datapin.TZData

func checkDataPins(config preflightConfig) error {
	pins, err := readVersionPins(config.versionFile)
	if err != nil {
		return err
	}
	if err := checkCLDRPackagePins(config.packageFile, pins.cldr); err != nil {
		return err
	}
	if pins.tzdata == "" {
		return fmt.Errorf("%s: missing tzdata version pin", config.versionFile)
	}
	lock, err := readTZDataLock(config.tzLockFile)
	if err != nil {
		return err
	}
	if lock.Version != pins.tzdata {
		return fmt.Errorf("%s: tzdata version mismatch: expected %s from %s, got %s", config.tzLockFile, pins.tzdata, config.versionFile, lock.Version)
	}
	goVersion, err := readGoTZDataVersion(config.goUpdateFile)
	if err != nil {
		return err
	}
	identity, err := parseTZDataVersion(pins.tzdata)
	if err != nil {
		return fmt.Errorf("%s: invalid tzdata version %q: %w", config.versionFile, pins.tzdata, err)
	}
	transition, err := parseTZDataVersion(goVersion)
	if err != nil {
		return fmt.Errorf("%s: invalid DATA version %q: %w", config.goUpdateFile, goVersion, err)
	}
	if transition.less(identity) {
		return fmt.Errorf("%s: Go transition tzdata %s is older than identity tzdata %s", config.goUpdateFile, goVersion, pins.tzdata)
	}
	return nil
}

type tzDataVersion struct {
	year    int
	release string
}

func readGoTZDataVersion(path string) (string, error) {
	data, err := readPreflightFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	version := ""
	for lineNumber, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		value, ok := strings.CutPrefix(line, "DATA=")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return "", fmt.Errorf("%s:%d: empty DATA version", path, lineNumber+1)
		}
		if version != "" {
			return "", fmt.Errorf("%s:%d: duplicate DATA version", path, lineNumber+1)
		}
		version = value
	}
	if version == "" {
		return "", fmt.Errorf("%s: missing DATA version", path)
	}
	return version, nil
}

func parseTZDataVersion(version string) (tzDataVersion, error) {
	if len(version) < 5 {
		return tzDataVersion{}, fmt.Errorf("expected YYYY plus release suffix")
	}
	year, err := strconv.Atoi(version[:4])
	if err != nil {
		return tzDataVersion{}, fmt.Errorf("invalid year: %w", err)
	}
	release := version[4:]
	for _, r := range release {
		if r < 'a' || r > 'z' {
			return tzDataVersion{}, fmt.Errorf("invalid release suffix %q", release)
		}
	}
	return tzDataVersion{year: year, release: release}, nil
}

func (v tzDataVersion) less(other tzDataVersion) bool {
	if v.year != other.year {
		return v.year < other.year
	}
	return v.release < other.release
}

func readTZDataLock(path string) (tzDataLock, error) { return datapin.ReadTZData(path) }

func checkCLDRPackagePins(path, want string) error {
	data, err := readPreflightFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	var manifest struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	for _, name := range datapin.RequiredCLDRPackages() {
		if manifest.Dependencies[name] == "" {
			return fmt.Errorf("%s: missing %s package pin", path, name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(manifest.Dependencies)) {
		if !strings.HasPrefix(name, "cldr-") {
			continue
		}
		if got := manifest.Dependencies[name]; got != want {
			return fmt.Errorf("%s: package %s version mismatch: expected %s, got %s", path, name, want, got)
		}
	}
	return nil
}

func readVersionPins(path string) (dataPins, error) {
	pins, err := datapin.ReadVersions(path)
	return dataPins{cldr: pins.CLDR, tzdata: pins.TZData}, err
}

func readPreflightFile(path string) ([]byte, error) {
	// Paths are fixed by the maintainer-run preflight command or isolated test fixtures.
	return os.ReadFile(path) // #nosec G304 -- reading the selected preflight input is the tool's purpose.
}
