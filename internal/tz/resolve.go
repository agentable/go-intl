package tz

import (
	"fmt"
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
	if name != "" && (name[0] == '+' || name[0] == '-') {
		offsetMs, err := ParseOffsetString(name)
		if err != nil {
			return nil, err
		}
		canonical, err := CanonicalOffsetString(name)
		if err != nil {
			return nil, err
		}
		if v, ok := locationCache.Load(canonical); ok {
			return v.(*time.Location), nil
		}
		loc := time.FixedZone(canonical, int(offsetMs/1000))
		actual, _ := locationCache.LoadOrStore(canonical, loc)
		return actual.(*time.Location), nil
	}
	canonical := CanonicalLink(name)
	if v, ok := locationCache.Load(canonical); ok {
		return v.(*time.Location), nil
	}
	loc, err := time.LoadLocation(canonical)
	if err != nil {
		return nil, fmt.Errorf("tz: unsupported time zone %q: %w", name, ErrUnsupportedTimeZone)
	}
	actual, _ := locationCache.LoadOrStore(canonical, loc)
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
