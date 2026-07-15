package tz

import (
	"slices"
	"sync"
)

// IdentifierRecord is one immutable ECMA-402 time-zone identifier record.
type IdentifierRecord struct {
	Identifier string
	Primary    string
}

type timeZoneRegionRecord struct {
	region string
	zones  []string
}

var timeZoneIdentifierIndex = sync.OnceValue(func() map[string]int {
	index := make(map[string]int, len(timeZoneIdentifierRecords))
	for i, record := range &timeZoneIdentifierRecords {
		folded, _ := foldASCIIIdentifier(record.Identifier)
		index[folded] = i
	}
	return index
})

// LookupIdentifier performs the ECMA-402 ASCII-case-insensitive identifier
// lookup and returns source casing plus the stable primary identifier.
func LookupIdentifier(name string) (IdentifierRecord, bool) {
	folded, ascii := foldASCIIIdentifier(name)
	if !ascii {
		return IdentifierRecord{}, false
	}
	idx, ok := timeZoneIdentifierIndex()[folded]
	if !ok {
		return IdentifierRecord{}, false
	}
	return timeZoneIdentifierRecords[idx], true
}

func foldASCIIIdentifier(name string) (string, bool) {
	hasUpper := false
	for i := range len(name) {
		b := name[i]
		if b >= 0x80 {
			return "", false
		}
		hasUpper = hasUpper || b >= 'A' && b <= 'Z'
	}
	if !hasUpper {
		return name, true
	}
	folded := []byte(name)
	for i, b := range folded {
		if b >= 'A' && b <= 'Z' {
			folded[i] = b + ('a' - 'A')
		}
	}
	return string(folded), true
}

// SupportedTimeZones returns the sorted primary identifier projection.
func SupportedTimeZones() []string {
	out := make([]string, 0, len(timeZoneIdentifierRecords))
	for _, record := range &timeZoneIdentifierRecords {
		if record.Identifier == record.Primary {
			out = append(out, record.Identifier)
		}
	}
	return out
}

// TimeZonesForRegion returns the sorted primary identifiers assigned by IANA
// zone.tab to an ISO 3166-1 alpha-2 region.
func TimeZonesForRegion(region string) []string {
	for _, record := range &timeZoneRegionRecords {
		if record.region == region {
			return slices.Clone(record.zones)
		}
	}
	return nil
}
