package codegen

import (
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

type canonicalTimeZoneLinkRecord struct {
	alias     string
	canonical string
}

// legacyIANATimeZoneLinks holds tzdb backward links not present in CLDR
// supplemental aliases but already accepted by the runtime through Go tzdata.
var legacyIANATimeZoneLinks = [...]canonicalTimeZoneLinkRecord{
	{alias: "US/Eastern", canonical: "America/New_York"},
	{alias: "US/Pacific", canonical: "America/Los_Angeles"},
	{alias: "Asia/Calcutta", canonical: "Asia/Kolkata"},
}

func canonicalTimeZoneLinks(aliases []cldr.TimeZoneAlias) []canonicalTimeZoneLinkRecord {
	byAlias := make(map[string]string, len(aliases)+len(legacyIANATimeZoneLinks))
	for _, alias := range aliases {
		if alias.Alias == "" || alias.Canonical == "" || alias.Alias == alias.Canonical {
			continue
		}
		byAlias[alias.Alias] = alias.Canonical
	}
	for _, link := range legacyIANATimeZoneLinks[:] {
		if _, ok := byAlias[link.alias]; !ok {
			byAlias[link.alias] = link.canonical
		}
	}
	keys := slices.Sorted(maps.Keys(byAlias))
	out := make([]canonicalTimeZoneLinkRecord, len(keys))
	for i, alias := range keys {
		out[i] = canonicalTimeZoneLinkRecord{alias: alias, canonical: byAlias[alias]}
	}
	return out
}

func canonicalTimeZoneLink(name string, links []canonicalTimeZoneLinkRecord) string {
	for _, link := range links {
		if link.alias == name {
			return link.canonical
		}
	}
	return name
}
