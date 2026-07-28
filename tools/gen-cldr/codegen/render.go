package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

type RuntimeInput struct {
	Manifest         ManifestInput
	Locales          extract.Locales
	LikelySubtags    extract.LikelySubtags
	ScriptDirections map[string]bool
	Numbers          extract.Numbers
	Currencies       extract.CurrencyData
	Dates            extract.Dates
	Preferences      cldr.PreferenceData
	Metazones        extract.Metazones
	Units            extract.Units
	ListPatterns     extract.ListPatterns
	RelativeTime     extract.RelativeTimeFields
	DisplayNames     extract.DisplayNames
}

type generatedFile struct {
	name string
	src  []byte
}

const localeDataFile = "locale/data.go"

type localeKernelLeaf struct {
	name   string
	render func(RuntimeInput) ([]byte, error)
}

var localeKernelLeaves = [...]localeKernelLeaf{
	{name: "locale/manifest.go", render: func(in RuntimeInput) ([]byte, error) { return renderManifest(in.Manifest) }},
}

func RenderRuntime(outDir string, input RuntimeInput) error {
	files, err := renderRuntimeFiles(input)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(outDir, file.name)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
		if err := os.WriteFile(path, file.src, 0o666); err != nil {
			return fmt.Errorf("write %s: %w", file.name, err)
		}
	}
	return nil
}

func renderCLDRFileError(name string, err error) error {
	return fmt.Errorf("render internal/cldr/%s: %w", name, err)
}

// renderRuntimeFiles emits the generated CLDR data. The root internal/cldr
// package is retired: every domain (including the locale kernel) owns its data
// directly, so the generator no longer writes any package cldr file. The root
// string table is gone too — each domain carries its own private _data table.
func renderRuntimeFiles(input RuntimeInput) ([]generatedFile, error) {
	leafFiles, err := renderLeafCLDRFiles(input)
	if err != nil {
		return nil, err
	}
	domainFiles, err := renderDomainFiles(input)
	if err != nil {
		return nil, err
	}
	files := make([]generatedFile, len(leafFiles)+len(domainFiles))
	copy(files, leafFiles)
	copy(files[len(leafFiles):], domainFiles)
	return files, nil
}

// renderLeafCLDRFiles renders the locale kernel package (cldrlocale). The kernel
// follows the same const-only-payload form as the leaf domains: a generated
// data.go holds _data plus the kernel blobs, and the hand-written decode.go and
// accessors.go expand and query them. The registry data and likely-subtags bomb
// that formerly lived in literal locales.go / likely_subtags.go / preference.go /
// strings.go are dissolved into that single payload; the kernel _data holds only
// the kernel's own strings.
//
// The data manifest (CLDR/ICU/tzdata pin and locale profile) is owned by the
// kernel too, so manifest.go is rendered here with the kernel package name.
func renderLeafCLDRFiles(input RuntimeInput) ([]generatedFile, error) {
	files := make([]generatedFile, 1+len(localeKernelLeaves))

	localeData, err := encodeLocaleKernel(input, NewStringTable())
	if err != nil {
		return nil, renderCLDRFileError(localeDataFile, err)
	}
	files[0] = generatedFile{name: localeDataFile, src: localeData}

	for i, leaf := range localeKernelLeaves[:] {
		src, err := leaf.render(input)
		if err != nil {
			return nil, renderCLDRFileError(leaf.name, err)
		}
		src, err = replaceGeneratedPackage(src, "cldrlocale")
		if err != nil {
			return nil, renderCLDRFileError(leaf.name, err)
		}
		files[i+1] = generatedFile{name: leaf.name, src: src}
	}
	return files, nil
}

func localeKernelFileNames() []string {
	names := make([]string, 1+len(localeKernelLeaves))
	names[0] = localeDataFile
	for i, leaf := range localeKernelLeaves[:] {
		names[i+1] = leaf.name
	}
	return names
}

func runtimeFileNames() []string {
	kernelNames := localeKernelFileNames()
	names := make([]string, len(kernelNames)+len(domains))
	copy(names, kernelNames)
	for i, d := range domains[:] {
		names[len(kernelNames)+i] = d.payloadFile()
	}
	return names
}

func replaceGeneratedPackage(src []byte, packageName string) ([]byte, error) {
	replaced := bytes.Replace(src, []byte("package cldr\n"), []byte("package "+packageName+"\n"), 1)
	return FormatFile(replaced)
}
