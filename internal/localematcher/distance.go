package localematcher

import (
	"cmp"
	"slices"
	"sync"
)

var distanceCache sync.Map

type localeDistanceRecord struct {
	desired   string
	supported string
	distance  int
}

var localeDistanceOverrides = [...]localeDistanceRecord{
	{desired: "en-CA", supported: "en-GB", distance: 50},
	{desired: "en-CA", supported: "en-US", distance: 39},
	{desired: "es-KY", supported: "es", distance: 49},
	{desired: "es-KY", supported: "es-419", distance: 39},
	{desired: "zh-HK", supported: "zh-Hant", distance: 50},
	{desired: "zh-HK", supported: "zh-MO", distance: 40},
	{desired: "zh-TW", supported: "zh", distance: 50},
	{desired: "zh-TW", supported: "zh-Hant", distance: 0},
}

func matchingDistance(desired, supported, maximizedDesired, maximizedSupported string) int {
	if desired == supported || maximizedDesired == maximizedSupported {
		return 0
	}
	if distance, ok := localeDistanceOverride(desired, supported); ok {
		return distance
	}
	desiredLanguage := languagePart(desired)
	supportedLanguage := languagePart(supported)
	if desiredLanguage == supportedLanguage {
		return 40
	}
	return 840
}

func localeDistanceOverride(desired, supported string) (int, bool) {
	target := localeDistanceRecord{desired: desired, supported: supported}
	idx, ok := slices.BinarySearchFunc(localeDistanceOverrides[:], target, compareLocaleDistanceRecord)
	if !ok {
		return 0, false
	}
	return localeDistanceOverrides[idx].distance, true
}

func compareLocaleDistanceRecord(a, b localeDistanceRecord) int {
	if byDesired := cmp.Compare(a.desired, b.desired); byDesired != 0 {
		return byDesired
	}
	return cmp.Compare(a.supported, b.supported)
}

func languagePart(tag string) string {
	for i, r := range tag {
		if r == '-' {
			return tag[:i]
		}
	}
	return tag
}
