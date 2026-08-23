package cldr

import (
	"encoding/json/v2"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const worldRegion = "001"

type PreferenceData struct {
	HourCycle map[string][]string
	Week      map[string]WeekData
	Calendar  map[string][]string
}

type WeekData struct {
	FirstDay     string
	WeekendStart string
	WeekendEnd   string
	MinDays      int
}

type rawWeekData struct {
	FirstDay     map[string]string `json:"firstDay"`
	WeekendStart map[string]string `json:"weekendStart"`
	WeekendEnd   map[string]string `json:"weekendEnd"`
	MinDays      map[string]string `json:"minDays"`
}

func loadPreferenceData(root string) (PreferenceData, error) {
	hour, err := loadHourCyclePreference(root)
	if err != nil {
		return PreferenceData{}, err
	}
	week, err := loadWeekData(root)
	if err != nil {
		return PreferenceData{}, err
	}
	calendar, err := loadCalendarPreference(root)
	if err != nil {
		return PreferenceData{}, err
	}
	return PreferenceData{HourCycle: hour, Week: week, Calendar: calendar}, nil
}

func loadHourCyclePreference(root string) (map[string][]string, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "timeData.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			TimeData map[string]struct {
				Allowed string `json:"_allowed"`
			} `json:"timeData"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse timeData.json: %w", err)
	}
	if doc.Supplemental.TimeData == nil {
		return nil, fmt.Errorf("expected supplemental timeData map")
	}
	world, ok := doc.Supplemental.TimeData[worldRegion]
	if !ok || world.Allowed == "" {
		return nil, fmt.Errorf("expected supplemental timeData %s default", worldRegion)
	}
	out := make(map[string][]string, len(doc.Supplemental.TimeData))
	for region, data := range doc.Supplemental.TimeData {
		cycles, err := hourCycles(data.Allowed)
		if err != nil {
			return nil, fmt.Errorf("parse hour cycles for %s: %w", region, err)
		}
		out[preferenceRegionKey(region)] = cycles
	}
	return out, nil
}

func hourCycles(raw string) ([]string, error) {
	out := make([]string, 0, 4)
	for _, token := range strings.Fields(raw) {
		switch token {
		case "h":
			out = append(out, "h12")
		case "H":
			out = append(out, "h23")
		case "K":
			out = append(out, "h11")
		case "k":
			out = append(out, "h24")
		case "hb", "hB":
			// CLDR timeData may include day-period hour symbols; ECMA-402
			// hourCycle records only the base h/H/K/k cycle.
		default:
			return nil, fmt.Errorf("invalid hour cycle token %q", token)
		}
	}
	return out, nil
}

func loadWeekData(root string) (map[string]WeekData, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "weekData.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			WeekData rawWeekData `json:"weekData"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse weekData.json: %w", err)
	}
	if err := validateWeekDataShape(doc.Supplemental.WeekData); err != nil {
		return nil, err
	}
	regions := make(map[string]bool)
	for region := range doc.Supplemental.WeekData.FirstDay {
		regions[region] = true
	}
	for region := range doc.Supplemental.WeekData.WeekendStart {
		regions[region] = true
	}
	for region := range doc.Supplemental.WeekData.WeekendEnd {
		regions[region] = true
	}
	for region := range doc.Supplemental.WeekData.MinDays {
		regions[region] = true
	}
	worldMinDays, err := strconv.Atoi(doc.Supplemental.WeekData.MinDays[worldRegion])
	if err != nil {
		return nil, fmt.Errorf("parse minDays %q: %w", doc.Supplemental.WeekData.MinDays[worldRegion], err)
	}
	world := WeekData{
		FirstDay:     doc.Supplemental.WeekData.FirstDay[worldRegion],
		WeekendStart: doc.Supplemental.WeekData.WeekendStart[worldRegion],
		WeekendEnd:   doc.Supplemental.WeekData.WeekendEnd[worldRegion],
		MinDays:      worldMinDays,
	}
	out := make(map[string]WeekData, len(regions))
	for region := range regions {
		record := world
		if day := doc.Supplemental.WeekData.FirstDay[region]; day != "" {
			record.FirstDay = day
		}
		if day := doc.Supplemental.WeekData.WeekendStart[region]; day != "" {
			record.WeekendStart = day
		}
		if day := doc.Supplemental.WeekData.WeekendEnd[region]; day != "" {
			record.WeekendEnd = day
		}
		if rawDays := doc.Supplemental.WeekData.MinDays[region]; rawDays != "" {
			days, err := strconv.Atoi(rawDays)
			if err != nil {
				return nil, fmt.Errorf("parse minDays %q: %w", rawDays, err)
			}
			record.MinDays = days
		}
		out[preferenceRegionKey(region)] = record
	}
	return out, nil
}

func validateWeekDataShape(data rawWeekData) error {
	if err := validateWeekDataMap("firstDay", data.FirstDay); err != nil {
		return err
	}
	if err := validateWeekDataMap("weekendStart", data.WeekendStart); err != nil {
		return err
	}
	if err := validateWeekDataMap("weekendEnd", data.WeekendEnd); err != nil {
		return err
	}
	if err := validateWeekDataMap("minDays", data.MinDays); err != nil {
		return err
	}
	return nil
}

func validateWeekDataMap(name string, values map[string]string) error {
	if values == nil {
		return fmt.Errorf("expected supplemental weekData %s map", name)
	}
	if values[worldRegion] == "" {
		return fmt.Errorf("expected supplemental weekData %s %s default", name, worldRegion)
	}
	return nil
}

func loadCalendarPreference(root string) (map[string][]string, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "calendarPreferenceData.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			CalendarPreferenceData map[string][]string `json:"calendarPreferenceData"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse calendarPreferenceData.json: %w", err)
	}
	if doc.Supplemental.CalendarPreferenceData == nil {
		return nil, fmt.Errorf("expected supplemental calendarPreferenceData map")
	}
	if calendars, ok := doc.Supplemental.CalendarPreferenceData[worldRegion]; !ok || len(calendars) == 0 {
		return nil, fmt.Errorf("expected supplemental calendarPreferenceData %s default", worldRegion)
	}
	out := make(map[string][]string, len(doc.Supplemental.CalendarPreferenceData))
	for region, calendars := range doc.Supplemental.CalendarPreferenceData {
		out[preferenceRegionKey(region)] = calendars
	}
	return out, nil
}

func preferenceRegionKey(region string) string {
	return strings.ReplaceAll(region, "_", "-")
}
