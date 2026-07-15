package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentable/go-intl/tools/gen-cldr/tzdb"
)

func RenderTimeZoneRegistry(path string, registry tzdb.Registry) error {
	src, err := renderTimeZoneRegistry(registry)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, src, 0o666); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func renderTimeZoneRegistry(registry tzdb.Registry) ([]byte, error) {
	var body bytes.Buffer
	if _, err := fmt.Fprintf(&body, "package tz\n\nconst timeZoneDataVersion = %q\nconst timeZoneDataSHA256 = %q\n\n", registry.Version, registry.SHA256); err != nil {
		return nil, err
	}
	if _, err := body.WriteString("var timeZoneIdentifierRecords = [...]IdentifierRecord{\n"); err != nil {
		return nil, err
	}
	for _, record := range registry.Records {
		if _, err := fmt.Fprintf(&body, "\t{Identifier: %q, Primary: %q},\n", record.Identifier, record.Primary); err != nil {
			return nil, err
		}
	}
	if _, err := body.WriteString("}\n\nvar timeZoneRegionRecords = [...]timeZoneRegionRecord{\n"); err != nil {
		return nil, err
	}
	for _, region := range registry.Regions {
		if _, err := fmt.Fprintf(&body, "\t{region: %q, zones: %#v},\n", region.Code, region.Zones); err != nil {
			return nil, err
		}
	}
	if _, err := body.WriteString("}\n"); err != nil {
		return nil, err
	}
	return FormatFile(body.Bytes())
}
