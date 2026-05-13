package tz

import (
	"os"
	"strings"
	"time"
)

// Default returns the host default time zone snapshot used by DateTimeFormat
// construction when no explicit timeZone option is provided.
func Default() (string, *time.Location) {
	return defaultLocation(time.Local)
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
			return canonicalName(name), loc
		}
	}
	return "UTC", time.UTC
}

func canonicalLocation(name string, fallback *time.Location) (string, *time.Location) {
	if strings.HasPrefix(name, "+") || strings.HasPrefix(name, "-") {
		if loc, err := Resolve(name); err == nil {
			return name, loc
		}
		return name, fallback
	}
	canonical := canonicalName(name)
	if loc, err := Resolve(canonical); err == nil {
		return canonical, loc
	}
	return name, fallback
}

func canonicalName(name string) string {
	if strings.HasPrefix(name, "+") || strings.HasPrefix(name, "-") {
		return name
	}
	return CanonicalLink(name)
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
