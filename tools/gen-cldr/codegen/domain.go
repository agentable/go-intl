package codegen

// domain describes one CLDR semantic domain that is emitted as a self-contained
// package of const-only payload (data.go) plus hand-written decode.go and
// accessors.go. The registry is the single source of truth that the generator,
// the round-trip gate, and the shape gate all derive their expectations from:
// adding a domain is adding a row here, not scattering functions across files.
//
// Fields are kept minimal on purpose. The package name is the path segment
// under internal/cldr/; the payload file is always <pkg>/data.go by convention,
// so it is derived rather than stored. emit renders that const-only payload
// from the typed extract input.
type domain struct {
	// pkg is the package directory under internal/cldr/ (also the payload path
	// stem: internal/cldr/<pkg>/data.go).
	pkg string
	// emit renders the const-only payload (data.go) for this domain. A fresh
	// per-domain StringTable is created by the caller so the _data table holds
	// only this domain's strings.
	emit func(input RuntimeInput, table *StringTable) ([]byte, error)
}

// domains is the domain registry. New domains are migrated in one row at a
// time as the literal renderers retire.
var domains = [...]domain{
	{pkg: "unit", emit: encodeUnits},
	{pkg: "date", emit: encodeDates},
	{pkg: "relativetime", emit: encodeRelativeTime},
	{pkg: "currency", emit: encodeCurrencies},
	{pkg: "number", emit: encodeNumbers},
	{pkg: "list", emit: encodeList},
	{pkg: "timezone", emit: encodeTimezone},
	{pkg: "displaynames", emit: encodeDisplayNames},
}

// payloadFile returns the internal/cldr-relative payload file for a domain.
func (d domain) payloadFile() string {
	return d.pkg + "/data.go"
}

// renderDomainFiles renders the const-only payload (data.go) for every
// registered domain. Each domain gets a fresh StringTable so its _data holds
// only its own strings — the representation invariant that dissolves the shared
// global string table.
func renderDomainFiles(input RuntimeInput) ([]generatedFile, error) {
	files := make([]generatedFile, len(domains))
	for i, d := range domains[:] {
		table := NewStringTable()
		src, err := d.emit(input, table)
		name := d.payloadFile()
		if err != nil {
			return nil, renderCLDRFileError(name, err)
		}
		files[i] = generatedFile{name: name, src: src}
	}
	return files, nil
}
