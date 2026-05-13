package cldr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read timeData.json: %w", err)
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
	out := make(map[string][]string, len(doc.Supplemental.TimeData))
	for region, data := range doc.Supplemental.TimeData {
		out[strings.ReplaceAll(region, "_", "-")] = hourCycles(data.Allowed)
	}
	return out, nil
}

func hourCycles(raw string) []string {
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
		}
	}
	return out
}

func loadWeekData(root string) (map[string]WeekData, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "weekData.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read weekData.json: %w", err)
	}
	var doc struct {
		Supplemental struct {
			WeekData struct {
				FirstDay     map[string]string `json:"firstDay"`
				WeekendStart map[string]string `json:"weekendStart"`
				WeekendEnd   map[string]string `json:"weekendEnd"`
				MinDays      map[string]string `json:"minDays"`
			} `json:"weekData"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse weekData.json: %w", err)
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
	worldMinDays, err := strconv.Atoi(doc.Supplemental.WeekData.MinDays["001"])
	if err != nil {
		return nil, fmt.Errorf("parse minDays %q: %w", doc.Supplemental.WeekData.MinDays["001"], err)
	}
	world := WeekData{
		FirstDay:     doc.Supplemental.WeekData.FirstDay["001"],
		WeekendStart: doc.Supplemental.WeekData.WeekendStart["001"],
		WeekendEnd:   doc.Supplemental.WeekData.WeekendEnd["001"],
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
		out[region] = record
	}
	return out, nil
}

func loadCalendarPreference(root string) (map[string][]string, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "calendarPreferenceData.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read calendarPreferenceData.json: %w", err)
	}
	var doc struct {
		Supplemental struct {
			CalendarPreferenceData map[string][]string `json:"calendarPreferenceData"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse calendarPreferenceData.json: %w", err)
	}
	return doc.Supplemental.CalendarPreferenceData, nil
}
