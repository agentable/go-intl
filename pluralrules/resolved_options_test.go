package pluralrules

import (
	"encoding/json/v2"
	"regexp"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

// TestResolvedOptionsJSONKeyOrder pins the marshaled key order to the ECMA-402
// §16.4.5 "Resolved Options of PluralRules Instances" table, so the JSON shape
// matches Node's Intl.PluralRules.prototype.resolvedOptions() exactly.
func TestResolvedOptionsJSONKeyOrder(t *testing.T) {
	t.Parallel()

	notation := string(CompactNotation)
	compactDisplay := string(ShortCompactDisplay)
	minFrac := 1
	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Notation:              &notation,
		CompactDisplay:        &compactDisplay,
		MinimumFractionDigits: &minFrac,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(rules.ResolvedOptions())
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"locale",
		"type",
		"notation",
		"compactDisplay",
		"minimumIntegerDigits",
		"minimumFractionDigits",
		"maximumFractionDigits",
		"pluralCategories",
		"roundingIncrement",
		"roundingMode",
		"roundingPriority",
		"trailingZeroDisplay",
	}

	keyRe := regexp.MustCompile(`"([a-zA-Z]+)":`)
	matches := keyRe.FindAllStringSubmatch(string(data), -1)
	got := make([]string, 0, len(matches))
	for _, m := range matches {
		got = append(got, m[1])
	}

	// Filter got to only the keys we assert order for (all present keys must
	// appear in the spec order; optional absent keys like significant digits
	// are simply skipped in `want`).
	wantIndex := make(map[string]int, len(want))
	for i, k := range want {
		wantIndex[k] = i
	}
	last := -1
	for _, k := range got {
		idx, ok := wantIndex[k]
		if !ok {
			t.Fatalf("unexpected resolvedOptions key %q in %s", k, data)
		}
		if idx <= last {
			t.Errorf("resolvedOptions key %q out of ECMA-402 order\n got: %v\nwant order: %v", k, got, want)
		}
		last = idx
	}
}
