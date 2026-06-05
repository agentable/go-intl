package codegen

import (
	"maps"
	"slices"
)

func sortedLocaleKeys[T any](values map[string]T) []string {
	return slices.Sorted(maps.Keys(values))
}
