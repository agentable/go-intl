package locale

import "testing"

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

func TestListCanonicalizationDedupesAndPreservesOrder(t *testing.T) {
	t.Parallel()

	enUS := MustParse("en-us")
	zh := MustParse("zh-Hans-CN-u-nu-latn")

	got := canonicalizeList(List{
		enUS,
		MustParse("en-US"),
		zh,
		enUS,
	})
	want := List{enUS, zh}
	if len(got) != len(want) {
		t.Fatalf("canonicalizeList() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].String() != want[i].String() {
			t.Fatalf("canonicalizeList()[%d] = %q, want %q", i, got[i].String(), want[i].String())
		}
	}
}

func TestListCanonicalizationClonesInput(t *testing.T) {
	t.Parallel()

	input := List{MustParse("en-US")}
	got := canonicalizeList(input)
	input[0] = MustParse("fr")
	if got[0].String() != "en-US" {
		t.Fatalf("canonicalizeList() shares input backing array, got %q", got[0].String())
	}
}

func TestListStringsUsesCanonicalOrder(t *testing.T) {
	t.Parallel()

	list := List{
		MustParse("en-US"),
		MustParse("en-us"),
		MustParse("fr"),
	}
	got := list.Strings()
	want := []string{"en-US", "fr"}
	if len(got) != len(want) {
		t.Fatalf("List.Strings() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("List.Strings()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
