package cldr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
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

type timeZoneNamesBody struct {
	Dates *timeZoneNamesDates `json:"dates"`
}

type timeZoneNamesDates struct {
	TimeZoneNames *timeZoneNamesPayload `json:"timeZoneNames"`
}

type timeZoneNamesPayload struct {
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
}

const (
	openMetazoneStart = -1 << 63
	openMetazoneEnd   = 1<<63 - 1
)

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
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			MetaZones struct {
				MetazoneInfo struct {
					Timezone map[string]json.RawMessage `json:"timezone"`
				} `json:"metazoneInfo"`
			} `json:"metaZones"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse metaZones.json: %w", err)
	}
	if doc.Supplemental.MetaZones.MetazoneInfo.Timezone == nil {
		return nil, fmt.Errorf("expected supplemental metaZones metazoneInfo timezone map")
	}
	out := make(map[string][]MetazonePeriod)
	for _, area := range slices.Sorted(maps.Keys(doc.Supplemental.MetaZones.MetazoneInfo.Timezone)) {
		if err := appendZoneMetazones(out, []string{area}, doc.Supplemental.MetaZones.MetazoneInfo.Timezone[area]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func appendZoneMetazones(out map[string][]MetazonePeriod, path []string, raw json.RawMessage) error {
	raw = bytes.TrimSpace(raw)
	name := strings.Join(path, "/")
	if len(raw) == 0 {
		return fmt.Errorf("expected metazone history for %s", name)
	}
	switch raw[0] {
	case '[':
		var history metazoneHistory
		if err := json.Unmarshal(raw, &history); err != nil {
			return fmt.Errorf("parse metazone history for %s: %w", name, err)
		}
		return appendMetazoneHistory(out, name, history)
	case '{':
		var node map[string]json.RawMessage
		if err := json.Unmarshal(raw, &node); err != nil {
			return fmt.Errorf("parse metazone node for %s: %w", name, err)
		}
		if _, ok := node["usesMetazone"]; ok {
			var history metazoneHistory
			if err := json.Unmarshal(raw, &history); err != nil {
				return fmt.Errorf("parse metazone history for %s: %w", name, err)
			}
			return appendMetazoneHistory(out, name, history)
		}
		if len(node) == 0 {
			return fmt.Errorf("expected metazone history or nested timezone map for %s", name)
		}
		for _, segment := range slices.Sorted(maps.Keys(node)) {
			nextPath := append(slices.Clone(path), segment)
			if err := appendZoneMetazones(out, nextPath, node[segment]); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("expected metazone history or nested timezone map for %s", name)
	}
}

func appendMetazoneHistory(out map[string][]MetazonePeriod, name string, history metazoneHistory) error {
	if len(history) == 0 {
		return fmt.Errorf("metazone history empty for %s", name)
	}
	for _, item := range history {
		metazone := item.UsesMetazone.Mzone
		if metazone == "" {
			return fmt.Errorf("metazone missing for %s", name)
		}
		start, err := parseMetazoneBoundary(item.UsesMetazone.From, openMetazoneStart)
		if err != nil {
			return fmt.Errorf("parse metazone from %q for %s: %w", item.UsesMetazone.From, name, err)
		}
		end, err := parseMetazoneBoundary(item.UsesMetazone.To, openMetazoneEnd)
		if err != nil {
			return fmt.Errorf("parse metazone to %q for %s: %w", item.UsesMetazone.To, name, err)
		}
		out[name] = append(out[name], MetazonePeriod{Metazone: metazone, Start: start, End: end})
	}
	return nil
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
		if locale == undefinedLocale {
			continue
		}
		path := filepath.Join(root, "cldr-dates-full", "main", locale, "timeZoneNames.json")
		raw, ok, err := readOptionalFile(path)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if !ok {
			continue
		}
		var doc struct {
			Main map[string]timeZoneNamesBody `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("parse %s: %w", path, err)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, nil, nil, nil, fmt.Errorf("timeZoneNames body missing for %s", locale)
		}
		timeZoneNames, err := requiredTimeZoneNames(locale, body)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		formats[locale] = TimeZoneFormats{
			GMTFormat:     timeZoneNames.GMTFormat,
			GMTZeroFormat: timeZoneNames.GMTZeroFormat,
			HourFormat:    timeZoneNames.HourFormat,
		}
		localeNames := make(map[string]MetazoneNames)
		for _, metazone := range slices.Sorted(maps.Keys(timeZoneNames.Metazone)) {
			value := timeZoneNames.Metazone[metazone]
			localeNames[metazone] = metazoneNamesFromWidths(value.Long, value.Short)
		}
		if len(localeNames) > 0 {
			names[locale] = localeNames
		}
		localeZoneNames := make(map[string]MetazoneNames)
		localeCities := make(map[string]string)
		for area, byZone := range timeZoneNames.Zone {
			for zone, value := range byZone {
				key := area + "/" + zone
				names := metazoneNamesFromWidths(value.Long, value.Short)
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

func requiredTimeZoneNames(locale string, body timeZoneNamesBody) (timeZoneNamesPayload, error) {
	if body.Dates == nil {
		return timeZoneNamesPayload{}, fmt.Errorf("timeZoneNames dates missing for %s", locale)
	}
	if body.Dates.TimeZoneNames == nil {
		return timeZoneNamesPayload{}, fmt.Errorf("timeZoneNames data missing for %s", locale)
	}
	timeZoneNames := *body.Dates.TimeZoneNames
	switch {
	case timeZoneNames.GMTFormat == "":
		return timeZoneNamesPayload{}, fmt.Errorf("timeZoneNames gmtFormat missing for %s", locale)
	case timeZoneNames.GMTZeroFormat == "":
		return timeZoneNamesPayload{}, fmt.Errorf("timeZoneNames gmtZeroFormat missing for %s", locale)
	case timeZoneNames.HourFormat == "":
		return timeZoneNamesPayload{}, fmt.Errorf("timeZoneNames hourFormat missing for %s", locale)
	}
	return timeZoneNames, nil
}

type metazoneWidth struct {
	Generic  string `json:"generic"`
	Standard string `json:"standard"`
	Daylight string `json:"daylight"`
}

func metazoneNamesFromWidths(long, short metazoneWidth) MetazoneNames {
	return MetazoneNames{
		LongGeneric:   long.Generic,
		LongStandard:  long.Standard,
		LongDaylight:  long.Daylight,
		ShortGeneric:  short.Generic,
		ShortStandard: short.Standard,
		ShortDaylight: short.Daylight,
	}
}
