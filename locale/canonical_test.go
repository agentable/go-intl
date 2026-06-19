package locale

import "testing"

func TestMaximize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "en", want: "en-Latn-US"},
		{in: "zh", want: "zh-Hans-CN"},
		{in: "zh-Hant", want: "zh-Hant-TW"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := parseLocaleForTest(tc.in).Maximize().String(); got != tc.want {
				t.Fatalf("Maximize() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMinimize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "en-Latn-US", want: "en"},
		{in: "zh-Hans-CN", want: "zh"},
		{in: "zh-Hant-TW", want: "zh-Hant"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := parseLocaleForTest(tc.in).Minimize().String(); got != tc.want {
				t.Fatalf("Minimize() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMaximizePreservesExtensions(t *testing.T) {
	t.Parallel()

	loc := parseLocaleForTest("zh-u-hc-h23-ca-chinese")
	got := loc.Maximize()
	if got.String() != "zh-Hans-CN-u-ca-chinese-hc-h23" {
		t.Fatalf("Maximize() = %q", got.String())
	}
	if got.Calendar() != "chinese" || got.HourCycle() != "h23" {
		t.Fatalf("extensions = %#v", got)
	}
}
