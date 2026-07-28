package tz

import (
	"fmt"
	"sync"
	"time"
)

type ZoneInfo struct {
	Name     string
	OffsetMs int64
	IsDST    bool
	Abbrv    string
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
	record, ok := LookupIdentifier(name)
	if !ok {
		return nil, unsupportedTimeZone(name)
	}
	canonical := record.Primary
	return cachedLocation(canonical, func() (*time.Location, error) {
		loc, err := time.LoadLocation(canonical)
		if err != nil {
			return nil, fmt.Errorf("tz: load transitions for registered identifier %q: %w", canonical, err)
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
	record, ok := LookupIdentifier(name)
	if !ok {
		return name
	}
	return record.Primary
}

func LookupAt(loc *time.Location, t time.Time) ZoneInfo {
	local := t.In(loc)
	name, offset := local.Zone()
	return ZoneInfo{Name: loc.String(), OffsetMs: int64(offset) * 1000, IsDST: local.IsDST(), Abbrv: name}
}
