package locale

import (
	"slices"
	"testing"
)

func TestParseList(t *testing.T) {
	t.Parallel()

	got, err := ParseList("en-us", "zh-Hans-CN-u-nu-latn")
	if err != nil {
		t.Fatalf("ParseList err = %v", err)
	}
	want := []string{"en-US", "zh-Hans-CN-u-nu-latn"}
	if len(got) != len(want) {
		t.Fatalf("ParseList() length = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].String() != want[i] {
			t.Fatalf("ParseList()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}

func TestParseListRejectsInvalidLocale(t *testing.T) {
	t.Parallel()

	if _, err := ParseList("en-US", ""); err == nil {
		t.Fatal("ParseList with invalid locale succeeded, want error")
	}
}

func TestListStringsPreservesListOrder(t *testing.T) {
	t.Parallel()

	enUS := parseLocaleForTest("en-us")
	zh := parseLocaleForTest("zh-Hans-CN-u-nu-latn")

	got := List{
		enUS,
		parseLocaleForTest("en-US"),
		zh,
		enUS,
	}.Strings()
	want := []string{"en-US", "en-US", "zh-Hans-CN-u-nu-latn", "en-US"}
	if !slices.Equal(got, want) {
		t.Fatalf("List.Strings() = %v, want %v", got, want)
	}
	got[0] = "mutated"
	nextWant := []string{"en-US", "zh-Hans-CN-u-nu-latn"}
	if next := (List{enUS, zh}).Strings(); !slices.Equal(next, nextWant) {
		t.Fatalf("List.Strings() after caller mutation = %v, want %v", next, nextWant)
	}
}
