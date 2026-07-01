package codegen

type canonicalTimeZoneLinkRecord struct {
	alias     string
	canonical string
}

// canonicalTimeZoneLinks is the generator-side owner of the tiny static IANA
// aliases the locale kernel emits. Timezone supported-value precomputation and
// the generated runtime CanonicalTimeZoneLink switch both consume this table.
var canonicalTimeZoneLinks = [...]canonicalTimeZoneLinkRecord{
	{alias: "US/Eastern", canonical: "America/New_York"},
	{alias: "US/Pacific", canonical: "America/Los_Angeles"},
	{alias: "Asia/Calcutta", canonical: "Asia/Kolkata"},
}

func canonicalTimeZoneLink(name string) string {
	for _, link := range canonicalTimeZoneLinks[:] {
		if link.alias == name {
			return link.canonical
		}
	}
	return name
}
