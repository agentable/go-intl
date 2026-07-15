package codegen

import (
	"strings"
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/tzdb"
)

func TestRenderTimeZoneRegistry(t *testing.T) {
	t.Parallel()

	src, err := renderTimeZoneRegistry(tzdb.Registry{
		Version: "2025b",
		SHA256:  strings.Repeat("a", 64),
		Records: []tzdb.Record{
			{Identifier: "America/New_York", Primary: "America/New_York"},
			{Identifier: "US/Eastern", Primary: "America/New_York"},
		},
		Regions: []tzdb.Region{{Code: "US", Zones: []string{"America/New_York"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertSourceContainsAll(t, "time-zone registry", string(src),
		"package tz",
		`timeZoneDataVersion = "2025b"`,
		`Identifier: "US/Eastern", Primary: "America/New_York"`,
		`region: "US", zones: []string{"America/New_York"}`,
	)
}
