package main

import (
	"bytes"
	"fmt"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/internal/generateddata"
)

func compareGeneratedData(committedRoot, generatedRoot string) error {
	committed, err := generateddata.Read(committedRoot)
	if err != nil {
		return err
	}
	generated, err := generateddata.Read(generatedRoot)
	if err != nil {
		return err
	}
	for _, path := range slices.Sorted(maps.Keys(committed)) {
		fresh, ok := generated[path]
		if !ok {
			return fmt.Errorf("generated path %s is missing from fresh output", path)
		}
		if !bytes.Equal(committed[path], fresh) {
			return fmt.Errorf("generated path %s differs from fresh output", path)
		}
	}
	for _, path := range slices.Sorted(maps.Keys(generated)) {
		if _, ok := committed[path]; !ok {
			return fmt.Errorf("fresh generated path %s is missing from the committed tree", path)
		}
	}
	return nil
}
