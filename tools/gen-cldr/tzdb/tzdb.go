// Package tzdb loads the pinned IANA tzdb source used to generate ECMA-402
// time-zone identifier records.
package tzdb

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
)

type Pin struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
	License string `json:"license"`
}

type Record struct {
	Identifier string
	Primary    string
}

type Region struct {
	Code  string
	Zones []string
}

type Registry struct {
	Version string
	SHA256  string
	Records []Record
	Regions []Region
}

func ReadPin(path string) (Pin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Pin{}, fmt.Errorf("read tzdb pin %s: %w", path, err)
	}
	var pin Pin
	if err := json.Unmarshal(data, &pin); err != nil {
		return Pin{}, fmt.Errorf("parse tzdb pin %s: %w", path, err)
	}
	if pin.Version == "" || pin.URL == "" || len(pin.SHA256) != 64 || pin.License != "public-domain" {
		return Pin{}, fmt.Errorf("parse tzdb pin %s: incomplete or unsupported pin", path)
	}
	return pin, nil
}

func LoadArchive(path string, pin Pin, primaryAliases map[string]string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, fmt.Errorf("read tzdb archive %s: %w", path, err)
	}
	sum := fmt.Sprintf("%x", sha256.Sum256(data))
	if sum != pin.SHA256 {
		return Registry{}, fmt.Errorf("tzdb archive %s sha256 %s, want %s", path, sum, pin.SHA256)
	}
	files, err := readArchive(data)
	if err != nil {
		return Registry{}, fmt.Errorf("read tzdb archive %s: %w", path, err)
	}
	version := strings.TrimSpace(files["version"])
	if version != pin.Version {
		return Registry{}, fmt.Errorf("tzdb archive version %q, want %q", version, pin.Version)
	}
	registry, err := buildRegistry(files, primaryAliases)
	if err != nil {
		return Registry{}, fmt.Errorf("build tzdb %s registry: %w", pin.Version, err)
	}
	registry.Version = pin.Version
	registry.SHA256 = pin.SHA256
	return registry, nil
}

func LoadCLDRPrimaryAliases(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read CLDR timezone aliases %s: %w", path, err)
	}
	var doc struct {
		Keyword struct {
			Unicode struct {
				TimeZones map[string]jsontext.Value `json:"tz"`
			} `json:"u"`
		} `json:"keyword"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse CLDR timezone aliases %s: %w", path, err)
	}
	if len(doc.Keyword.Unicode.TimeZones) == 0 {
		return nil, fmt.Errorf("parse CLDR timezone aliases %s: keyword.u.tz missing", path)
	}
	aliases := map[string]string{}
	for key, raw := range doc.Keyword.Unicode.TimeZones {
		if strings.HasPrefix(key, "_") || len(raw) == 0 || raw[0] != '{' {
			continue
		}
		var metadata struct {
			Alias string `json:"_alias"`
			IANA  string `json:"_iana"`
		}
		if err := json.Unmarshal(raw, &metadata); err != nil {
			return nil, fmt.Errorf("parse CLDR timezone alias %s: %w", key, err)
		}
		names := strings.Fields(metadata.Alias)
		if len(names) == 0 {
			continue
		}
		primary := metadata.IANA
		if primary == "" {
			primary = names[0]
		}
		for _, name := range names {
			if previous, ok := aliases[name]; ok && previous != primary {
				return nil, fmt.Errorf("CLDR timezone alias %q maps to both %q and %q", name, previous, primary)
			}
			aliases[name] = primary
		}
	}
	return aliases, nil
}

var sourceFiles = [...]string{
	"africa", "antarctica", "asia", "australasia", "europe",
	"northamerica", "southamerica", "etcetera", "factory", "backward",
}

func readArchive(data []byte) (map[string]string, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	wanted := map[string]bool{"version": true, "zone.tab": true, "backzone": true, "LICENSE": true}
	for _, name := range sourceFiles {
		wanted[name] = true
	}
	files := make(map[string]string, len(wanted))
	r := tar.NewReader(gz)
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if !wanted[header.Name] {
			continue
		}
		content, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		files[header.Name] = string(content)
	}
	for name := range wanted {
		if files[name] == "" {
			return nil, fmt.Errorf("required file %s missing", name)
		}
	}
	if !strings.Contains(files["LICENSE"], "public domain") {
		return nil, fmt.Errorf("LICENSE does not declare tzdb source public domain")
	}
	return files, nil
}

func buildRegistry(files map[string]string, cldrPrimary map[string]string) (Registry, error) {
	zones := map[string]struct{}{}
	links := map[string]string{}
	for _, source := range sourceFiles {
		if err := parseDefinitions(source, files[source], zones, links); err != nil {
			return Registry{}, err
		}
	}
	zoneRegions, regionZones, zoneTabIDs, err := parseZoneTable(files["zone.tab"])
	if err != nil {
		return Registry{}, err
	}
	overrides, err := parseBackzoneOverrides(files["backzone"])
	if err != nil {
		return Registry{}, err
	}

	identifiers := slices.Sorted(maps.Keys(zones))
	identifiers = slices.AppendSeq(identifiers, maps.Keys(links))
	slices.Sort(identifiers)
	identifiers = slices.Compact(identifiers)
	known := make(map[string]struct{}, len(identifiers))
	folded := make(map[string]string, len(identifiers))
	for _, identifier := range identifiers {
		known[identifier] = struct{}{}
		lower, ok := foldASCIIIdentifier(identifier)
		if !ok {
			return Registry{}, fmt.Errorf("identifier %q contains non-ASCII bytes", identifier)
		}
		if previous, ok := folded[lower]; ok {
			return Registry{}, fmt.Errorf("identifiers %q and %q collide under ASCII case folding", previous, identifier)
		}
		folded[lower] = identifier
	}
	for alias, primary := range cldrPrimary {
		if _, ok := known[alias]; !ok {
			continue
		}
		if _, ok := known[primary]; !ok {
			if primary == "Etc/Unknown" {
				continue
			}
			return Registry{}, fmt.Errorf("CLDR timezone primary %q for %q is absent from IANA tzdb", primary, alias)
		}
	}
	for alias := range links {
		if _, err := resolveLink(alias, links, zones); err != nil {
			return Registry{}, err
		}
	}

	primaryByID := make(map[string]string, len(identifiers))
	for _, identifier := range identifiers {
		primary := identifier
		if identifier == "UTC" {
			primary = "UTC"
		} else if _, isZone := zones[identifier]; isZone {
			if identifier == "Etc/UTC" || identifier == "Etc/GMT" {
				primary = "UTC"
			}
		} else if zoneTabIDs[identifier] {
			primary = identifier
		} else {
			resolved, err := resolveLink(identifier, links, zones)
			if err != nil {
				return Registry{}, err
			}
			if resolved == "Etc/UTC" || resolved == "Etc/GMT" {
				primary = "UTC"
			} else if canonical := cldrPrimary[identifier]; canonical != "" {
				primary = canonical
			} else if override := overrides[identifier]; override != "" {
				primary = override
			} else {
				primary = resolved
			}
		}
		if _, ok := known[primary]; !ok {
			return Registry{}, fmt.Errorf("identifier %q has unknown primary %q", identifier, primary)
		}
		primaryByID[identifier] = primary
	}
	for identifier, primary := range primaryByID {
		seen := map[string]bool{identifier: true}
		for primaryByID[primary] != primary {
			if seen[primary] {
				return Registry{}, fmt.Errorf("primary cycle resolving %q at %q", identifier, primary)
			}
			seen[primary] = true
			primary = primaryByID[primary]
		}
		primaryByID[identifier] = primary
	}

	records := make([]Record, len(identifiers))
	for i, identifier := range identifiers {
		records[i] = Record{Identifier: identifier, Primary: primaryByID[identifier]}
	}
	regions := make([]Region, 0, len(regionZones))
	for _, code := range slices.Sorted(maps.Keys(regionZones)) {
		regionZones[code] = sortedUnique(regionZones[code])
		for _, zone := range regionZones[code] {
			if primaryByID[zone] != zone {
				return Registry{}, fmt.Errorf("zone.tab region %s identifier %q is not primary", code, zone)
			}
			if !slices.Contains(zoneRegions[zone], code) {
				return Registry{}, fmt.Errorf("zone.tab region metadata for %q is inconsistent", zone)
			}
		}
		regions = append(regions, Region{Code: code, Zones: regionZones[code]})
	}
	return Registry{Records: records, Regions: regions}, nil
}

func parseDefinitions(source, data string, zones map[string]struct{}, links map[string]string) error {
	lineNumber := 0
	for line := range strings.SplitSeq(data, "\n") {
		lineNumber++
		line, _, _ = strings.Cut(line, "#")
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "Zone":
			if len(fields) < 2 {
				return fmt.Errorf("%s:%d malformed Zone", source, lineNumber)
			}
			zones[fields[1]] = struct{}{}
		case "Link":
			if len(fields) < 3 {
				return fmt.Errorf("%s:%d malformed Link", source, lineNumber)
			}
			if previous, ok := links[fields[2]]; ok && previous != fields[1] {
				return fmt.Errorf("%s:%d Link %q targets both %q and %q", source, lineNumber, fields[2], previous, fields[1])
			}
			links[fields[2]] = fields[1]
		}
	}
	return nil
}

func parseZoneTable(data string) (map[string][]string, map[string][]string, map[string]bool, error) {
	zoneRegions := map[string][]string{}
	regionZones := map[string][]string{}
	zoneTabIDs := map[string]bool{}
	lineNumber := 0
	for line := range strings.SplitSeq(data, "\n") {
		lineNumber++
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 || len(fields[0]) != 2 {
			return nil, nil, nil, fmt.Errorf("zone.tab:%d malformed row", lineNumber)
		}
		region, zone := fields[0], fields[2]
		zoneRegions[zone] = append(zoneRegions[zone], region)
		regionZones[region] = append(regionZones[region], zone)
		zoneTabIDs[zone] = true
	}
	return zoneRegions, regionZones, zoneTabIDs, nil
}

func parseBackzoneOverrides(data string) (map[string]string, error) {
	overrides := map[string]string{}
	lineNumber := 0
	for line := range strings.SplitSeq(data, "\n") {
		lineNumber++
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "#PACKRATLIST zone.tab "); ok {
			line = rest
		} else {
			line, _, _ = strings.Cut(line, "#")
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "Link" {
			continue
		}
		if len(fields) < 3 {
			return nil, fmt.Errorf("backzone:%d malformed Link", lineNumber)
		}
		alias, target := fields[2], fields[1]
		if previous, ok := overrides[alias]; ok && previous != target {
			return nil, fmt.Errorf("backzone:%d override %q targets both %q and %q", lineNumber, alias, previous, target)
		}
		overrides[alias] = target
	}
	return overrides, nil
}

func resolveLink(identifier string, links map[string]string, zones map[string]struct{}) (string, error) {
	seen := map[string]bool{}
	current := identifier
	for {
		if seen[current] {
			return "", fmt.Errorf("Link cycle resolving %q at %q", identifier, current)
		}
		seen[current] = true
		if _, ok := zones[current]; ok {
			return current, nil
		}
		next, ok := links[current]
		if !ok {
			return "", fmt.Errorf("Link %q resolves to unknown identifier %q", identifier, current)
		}
		current = next
	}
}

func sortedUnique(values []string) []string {
	out := slices.Clone(values)
	slices.Sort(out)
	return slices.Compact(out)
}

func foldASCIIIdentifier(identifier string) (string, bool) {
	for i := range len(identifier) {
		if identifier[i] >= 0x80 {
			return "", false
		}
	}
	folded := []byte(identifier)
	for i, b := range folded {
		if b >= 'A' && b <= 'Z' {
			folded[i] = b + ('a' - 'A')
		}
	}
	return string(folded), true
}
