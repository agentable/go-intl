package cldr

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

var requiredPackages = [...]string{
	"cldr-core",
	"cldr-dates-full",
	"cldr-localenames-full",
	"cldr-misc-full",
	"cldr-numbers-full",
	"cldr-segments-full",
	"cldr-units-full",
}

func RequiredPackages() []string {
	return slices.Clone(requiredPackages[:])
}

func ResolveLocalDir(root string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("cldr-json root is required")
	}
	for _, name := range requiredPackages {
		info, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			return "", fmt.Errorf("missing %s under %s: %w", name, root, err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("%s under %s is not a directory", name, root)
		}
	}
	return root, nil
}
