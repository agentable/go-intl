package main

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

func TestNodeWitnessPreservesObservationPresence(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	witness := syntheticNodeWitness(t)
	witness.NumberFormatSmoke = []fixture{
		{ID: "numberformat-node-v26-empty", Source: "node:" + conformance.ActiveNodeWitnessVersion + ":numberformat", Locale: "en", Options: map[string]any{}, Input: 1,
			Expected: new(""), ExpectedOK: new(false), ExpectedLocales: []string{}, ExpectedParts: []conformance.Part{}, ExpectedRange: new(""), ExpectedRangeParts: []conformance.RangePart{}, ExpectedResolved: map[string]any{}},
		{ID: "numberformat-node-v26-absent", Source: "node:" + conformance.ActiveNodeWitnessVersion + ":numberformat", Locale: "en", Options: map[string]any{}, Input: 1, Expected: new("1")},
	}
	witness.SupportedValues.Values["currency"] = []string{}
	nodePath := writeNodeWitness(t, witness)
	if err := run([]string{"-node", nodePath, "-out", root}); err != nil {
		t.Fatal(err)
	}
	loaded, err := conformance.LoadFixtures(filepath.Join(root, "numberformat"))
	if err != nil {
		t.Fatal(err)
	}
	var empty, absent conformance.Fixture
	for _, fixture := range loaded {
		switch fixture.ID {
		case "numberformat-node-v26-empty":
			empty = fixture
		case "numberformat-node-v26-absent":
			absent = fixture
		}
	}
	if empty.Expected == nil || *empty.Expected != "" || empty.ExpectedOK == nil || *empty.ExpectedOK ||
		empty.ExpectedLocales == nil || empty.ExpectedParts == nil || empty.ExpectedRange == nil ||
		*empty.ExpectedRange != "" || empty.ExpectedRangeParts == nil || string(empty.ExpectedResolved) != "{}" {
		t.Fatalf("native pipeline lost an empty observation: %+v", empty)
	}
	if absent.Expected == nil || *absent.Expected != "1" || absent.ExpectedOK != nil ||
		absent.ExpectedLocales != nil || absent.ExpectedParts != nil || absent.ExpectedRange != nil ||
		absent.ExpectedRangeParts != nil || absent.ExpectedResolved != nil {
		t.Fatalf("native pipeline invented an absent observation: %+v", absent)
	}
}

func TestWriteFixturesPreservesEmptyExpectations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field string
		value string
		set   func(*fixture)
	}{
		{"expected", `""`, func(f *fixture) { f.Expected = new("") }},
		{"expectedOk", `false`, func(f *fixture) { f.ExpectedOK = new(false) }},
		{"expectedLocales", `[]`, func(f *fixture) { f.ExpectedLocales = []string{} }},
		{"expectedParts", `[]`, func(f *fixture) { f.ExpectedParts = []conformance.Part{} }},
		{"expectedRange", `""`, func(f *fixture) { f.ExpectedRange = new("") }},
		{"expectedRangeParts", `[]`, func(f *fixture) { f.ExpectedRangeParts = []conformance.RangePart{} }},
		{"expectedResolvedOptions", `{}`, func(f *fixture) { f.ExpectedResolved = map[string]any{} }},
	}
	for _, tc := range tests {
		t.Run(tc.field, func(t *testing.T) {
			t.Parallel()

			root := filepath.Join(t.TempDir(), "numberformat")
			path := filepath.Join(root, "testdata", "conformance", "manual", "empty.json")
			f := fixture{ID: "empty", Source: "manual:presence", Locale: "en", Options: map[string]any{}, Input: 1}
			tc.set(&f)
			if err := writeJSON(path, []fixture{f}); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var records []map[string]jsontext.Value
			if err := json.Unmarshal(data, &records); err != nil {
				t.Fatal(err)
			}
			if got := string(records[0][tc.field]); got != tc.value {
				t.Fatalf("written %s = %q, want %q", tc.field, got, tc.value)
			}
			loaded, err := conformance.LoadFixtures(root)
			if err != nil {
				t.Fatalf("LoadFixtures() rejected written observation: %v", err)
			}
			if len(loaded) != 1 {
				t.Fatalf("LoadFixtures() returned %d fixtures, want 1", len(loaded))
			}
		})
	}
}
