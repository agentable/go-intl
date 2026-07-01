package cldr

import (
	"fmt"
	"strings"

	pluralop "github.com/agentable/go-intl/internal/plural"
)

func pluralCategoryFromField(field, prefix string) (string, bool, error) {
	category, ok := strings.CutPrefix(field, prefix)
	if !ok {
		return "", false, nil
	}
	if _, ok := pluralop.ParseCategory(category); !ok {
		return "", false, fmt.Errorf("invalid plural category %q", category)
	}
	return category, true, nil
}
