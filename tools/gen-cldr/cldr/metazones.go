package cldr

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type Metazones struct {
	ZoneToMetazones map[string][]MetazonePeriod
	Names           map[string]map[string]MetazoneNames
	ZoneNames       map[string]map[string]MetazoneNames
	ExemplarCities  map[string]map[string]string
	Formats         map[string]TimeZoneFormats
}

type MetazonePeriod struct {
	Metazone string
	Start    int64
	End      int64
}

type MetazoneNames struct {
	LongGeneric, LongStandard, LongDaylight    string
	ShortGeneric, ShortStandard, ShortDaylight string
}

type TimeZoneFormats struct {
	GMTFormat     string
	GMTZeroFormat string
	HourFormat    string
}

type metazoneUse struct {
	UsesMetazone struct {
		Mzone string `json:"_mzone"`
		From  string `json:"_from"`
		To    string `json:"_to"`
	} `json:"usesMetazone"`
}

type metazoneHistory []metazoneUse

func (h *metazoneHistory) UnmarshalJSON(data []byte) error {
	var many []metazoneUse
	if err := json.Unmarshal(data, &many); err == nil {
		*h = many
		return nil
	}
	var one metazoneUse
	if err := json.Unmarshal(data, &one); err != nil {
		return err
	}
	*h = []metazoneUse{one}
	return nil
}

func loadMetazones(root string, locales []string) (Metazones, error) {
	zones, err := loadZoneToMetazone(root)
	if err != nil {
		return Metazones{}, err
	}
	names, zoneNames, cities, formats, err := loadTimeZoneNames(root, locales)
	if err != nil {
		return Metazones{}, err
	}
	return Metazones{ZoneToMetazones: zones, Names: names, ZoneNames: zoneNames, ExemplarCities: cities, Formats: formats}, nil
}

func loadZoneToMetazone(root string) (map[string][]MetazonePeriod, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "metaZones.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read metaZones.json: %w", err)
	}
	var doc struct {
		Supplemental struct {
			MetaZones struct {
				MetazoneInfo struct {
					Timezone map[string]map[string]metazoneHistory `json:"timezone"`
				} `json:"metazoneInfo"`
			} `json:"metaZones"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse metaZones.json: %w", err)
	}
	out := make(map[string][]MetazonePeriod)
	for _, area := range slices.Sorted(maps.Keys(doc.Supplemental.MetaZones.MetazoneInfo.Timezone)) {
		zones := doc.Supplemental.MetaZones.MetazoneInfo.Timezone[area]
		for _, zone := range slices.Sorted(maps.Keys(zones)) {
			history := zones[zone]
			name := area + "/" + zone
			for _, item := range history {
				metazone := item.UsesMetazone.Mzone
				if metazone == "" {
					continue
				}
				start, err := parseMetazoneBoundary(item.UsesMetazone.From, -1<<63)
				if err != nil {
					return nil, fmt.Errorf("parse metazone from %q for %s: %w", item.UsesMetazone.From, name, err)
				}
				end, err := parseMetazoneBoundary(item.UsesMetazone.To, 1<<63-1)
				if err != nil {
					return nil, fmt.Errorf("parse metazone to %q for %s: %w", item.UsesMetazone.To, name, err)
				}
				out[name] = append(out[name], MetazonePeriod{Metazone: metazone, Start: start, End: end})
			}
		}
	}
	return out, nil
}

func parseMetazoneBoundary(value string, fallback int64) (int64, error) {
	if value == "" {
		return fallback, nil
	}
	t, err := time.Parse("2006-01-02 15:04", value)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

func loadTimeZoneNames(root string, locales []string) (map[string]map[string]MetazoneNames, map[string]map[string]MetazoneNames, map[string]map[string]string, map[string]TimeZoneFormats, error) {
	names := make(map[string]map[string]MetazoneNames)
	zoneNames := make(map[string]map[string]MetazoneNames)
	cities := make(map[string]map[string]string)
	formats := make(map[string]TimeZoneFormats)
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		path := filepath.Join(root, "cldr-dates-full", "main", locale, "timeZoneNames.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc struct {
			Main map[string]struct {
				Dates struct {
					TimeZoneNames struct {
						GMTFormat     string `json:"gmtFormat"`
						GMTZeroFormat string `json:"gmtZeroFormat"`
						HourFormat    string `json:"hourFormat"`
						Metazone      map[string]struct {
							Long  metazoneWidth `json:"long"`
							Short metazoneWidth `json:"short"`
						} `json:"metazone"`
						Zone map[string]map[string]struct {
							ExemplarCity string        `json:"exemplarCity"`
							Long         metazoneWidth `json:"long"`
							Short        metazoneWidth `json:"short"`
						} `json:"zone"`
					} `json:"timeZoneNames"`
				} `json:"dates"`
			} `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("parse %s: %w", path, err)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, nil, nil, nil, fmt.Errorf("timeZoneNames body missing for %s", locale)
		}
		formats[locale] = TimeZoneFormats{
			GMTFormat:     body.Dates.TimeZoneNames.GMTFormat,
			GMTZeroFormat: body.Dates.TimeZoneNames.GMTZeroFormat,
			HourFormat:    body.Dates.TimeZoneNames.HourFormat,
		}
		localeNames := make(map[string]MetazoneNames)
		for _, metazone := range slices.Sorted(maps.Keys(body.Dates.TimeZoneNames.Metazone)) {
			value := body.Dates.TimeZoneNames.Metazone[metazone]
			localeNames[metazone] = MetazoneNames{
				LongGeneric:   value.Long.Generic,
				LongStandard:  value.Long.Standard,
				LongDaylight:  value.Long.Daylight,
				ShortGeneric:  value.Short.Generic,
				ShortStandard: value.Short.Standard,
				ShortDaylight: value.Short.Daylight,
			}
		}
		if len(localeNames) > 0 {
			names[locale] = localeNames
		}
		localeZoneNames := make(map[string]MetazoneNames)
		localeCities := make(map[string]string)
		for area, byZone := range body.Dates.TimeZoneNames.Zone {
			for zone, value := range byZone {
				key := area + "/" + zone
				names := MetazoneNames{
					LongGeneric:   value.Long.Generic,
					LongStandard:  value.Long.Standard,
					LongDaylight:  value.Long.Daylight,
					ShortGeneric:  value.Short.Generic,
					ShortStandard: value.Short.Standard,
					ShortDaylight: value.Short.Daylight,
				}
				if names != (MetazoneNames{}) {
					localeZoneNames[key] = names
				}
				if value.ExemplarCity != "" {
					localeCities[key] = value.ExemplarCity
				}
			}
		}
		if len(localeZoneNames) > 0 {
			zoneNames[locale] = localeZoneNames
		}
		if len(localeCities) > 0 {
			cities[locale] = localeCities
		}
	}
	return names, zoneNames, cities, formats, nil
}

type metazoneWidth struct {
	Generic  string `json:"generic"`
	Standard string `json:"standard"`
	Daylight string `json:"daylight"`
}
