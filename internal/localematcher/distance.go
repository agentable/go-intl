package localematcher

import "sync"

var distanceCache sync.Map

var fixtureDistances = map[[2]string]int{
	{"zh-TW", "zh-Hant"}: 0,
	{"zh-TW", "zh"}:      50,
	{"zh-HK", "zh-MO"}:   40,
	{"zh-HK", "zh-Hant"}: 50,
	{"en-CA", "en-US"}:   39,
	{"en-CA", "en-GB"}:   50,
	{"es-KY", "es-419"}:  39,
	{"es-KY", "es"}:      49,
}

func matchingDistance(desired, supported, maximizedDesired, maximizedSupported string) int {
	if desired == supported || maximizedDesired == maximizedSupported {
		return 0
	}
	if distance, ok := fixtureDistance(desired, supported); ok {
		return distance
	}
	desiredLanguage := languagePart(desired)
	supportedLanguage := languagePart(supported)
	if desiredLanguage == supportedLanguage {
		return 40
	}
	return 840
}

func fixtureDistance(desired, supported string) (int, bool) {
	distance, ok := fixtureDistances[[2]string{desired, supported}]
	return distance, ok
}

func languagePart(tag string) string {
	for i, r := range tag {
		if r == '-' {
			return tag[:i]
		}
	}
	return tag
}
