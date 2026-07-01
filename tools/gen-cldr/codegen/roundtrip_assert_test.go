package codegen

import "testing"

func assertStringSliceEqual(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s len = %d, want %d", name, len(got), len(want))
	}
	checkStringSliceEqual(t, name, got, want)
}

func checkStringSliceEqual(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s len = %d, want %d", name, len(got), len(want))
		return
	}
	for i, wantValue := range want {
		if got[i] != wantValue {
			t.Errorf("%s[%d] = %q, want %q", name, i, got[i], wantValue)
		}
	}
}
