package tz

import (
	"os"
	"strings"
	"sync"
	"time"
)

var defaultProvider = struct {
	sync.RWMutex
	location func() *time.Location
}{location: func() *time.Location { return time.Local }}

// Default returns the host default time zone snapshot used by DateTimeFormat
// construction when no explicit timeZone option is provided.
func Default() (string, *time.Location) {
	defaultProvider.RLock()
	location := defaultProvider.location
	defaultProvider.RUnlock()
	return defaultLocation(location())
}

// OverrideDefaultForTest replaces the default time-zone provider for tests and
// returns a restore function.
func OverrideDefaultForTest(name string) (func(), error) {
	loc, err := Resolve(name)
	if err != nil {
		return nil, err
	}
	defaultProvider.Lock()
	previous := defaultProvider.location
	defaultProvider.location = func() *time.Location { return loc }
	defaultProvider.Unlock()
	return func() {
		defaultProvider.Lock()
		defaultProvider.location = previous
		defaultProvider.Unlock()
	}, nil
}

func defaultLocation(local *time.Location) (string, *time.Location) {
	if local == nil {
		return "UTC", time.UTC
	}
	if name := local.String(); name != "" && name != "Local" {
		return canonicalLocation(name, local)
	}
	if name := localtimeLinkName(); name != "" {
		if loc, err := Resolve(name); err == nil {
			return CanonicalLink(name), loc
		}
	}
	return "UTC", time.UTC
}

func canonicalLocation(name string, fallback *time.Location) (string, *time.Location) {
	if isOffsetName(name) {
		if loc, err := Resolve(name); err == nil {
			canonical := loc.String()
			return canonical, loc
		}
		return name, fallback
	}
	canonical := CanonicalLink(name)
	if loc, err := Resolve(canonical); err == nil {
		return canonical, loc
	}
	return name, fallback
}

func localtimeLinkName() string {
	link, err := os.Readlink("/etc/localtime")
	if err != nil {
		return ""
	}
	const marker = "zoneinfo/"
	idx := strings.LastIndex(link, marker)
	if idx < 0 {
		return ""
	}
	return strings.TrimPrefix(link[idx+len(marker):], "/")
}
