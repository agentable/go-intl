package cldr

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentable/go-intl/tools/internal/datapin"
)

func ResolveLocalDir(root string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("cldr-json root is required")
	}
	for _, name := range datapin.RequiredCLDRPackages() {
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
