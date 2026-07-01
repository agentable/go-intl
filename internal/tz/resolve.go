package tz

import (
	"sync"
	"time"

	cldrtimezone "github.com/agentable/go-intl/internal/cldr/timezone"
)

type ZoneInfo struct {
	Name     string
	OffsetMs int64
	IsDST    bool
	Abbrv    string
	Metazone string
}

// locationCache memoizes successful canonical-name → *time.Location lookups.
// Both IANA names and canonical offset strings share the cache; the two key
// spaces are disjoint because offset keys begin with '+' or '-'.
var locationCache sync.Map

func Resolve(name string) (*time.Location, error) {
	if isOffsetName(name) {
		offsetMs, canonical, err := parseFixedOffset(name)
		if err != nil {
			return nil, err
		}
		return cachedLocation(canonical, func() (*time.Location, error) {
			return time.FixedZone(canonical, int(offsetMs/1000)), nil
		})
	}
	canonical := CanonicalLink(name)
	return cachedLocation(canonical, func() (*time.Location, error) {
		loc, err := time.LoadLocation(canonical)
		if err != nil {
			return nil, unsupportedTimeZone(name)
		}
		return loc, nil
	})
}

func cachedLocation(name string, load func() (*time.Location, error)) (*time.Location, error) {
	if v, ok := locationCache.Load(name); ok {
		return v.(*time.Location), nil
	}
	loc, err := load()
	if err != nil {
		return nil, err
	}
	actual, _ := locationCache.LoadOrStore(name, loc)
	return actual.(*time.Location), nil
}

func CanonicalLink(name string) string {
	return cldrtimezone.CanonicalTimeZoneLink(name)
}

func LookupAt(loc *time.Location, t time.Time) ZoneInfo {
	local := t.In(loc)
	name, offset := local.Zone()
	return ZoneInfo{Name: loc.String(), OffsetMs: int64(offset) * 1000, IsDST: isDST(loc, local), Abbrv: name}
}

func isDST(loc *time.Location, t time.Time) bool {
	_, januaryOffset := time.Date(t.Year(), time.January, 1, 12, 0, 0, 0, loc).Zone()
	_, julyOffset := time.Date(t.Year(), time.July, 1, 12, 0, 0, 0, loc).Zone()
	if januaryOffset == julyOffset {
		return false
	}
	_, offset := t.Zone()
	return offset == max(januaryOffset, julyOffset)
}
