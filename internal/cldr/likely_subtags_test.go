package cldr

import "testing"

func TestMaximizeSubtags(t *testing.T) {
	t.Parallel()

	lang, script, region, ok := MaximizeSubtags("zh", "", "")
	if !ok {
		t.Fatal("MaximizeSubtags(zh) ok=false")
	}
	if lang != "zh" || script != "Hans" || region != "CN" {
		t.Fatalf("MaximizeSubtags(zh) = %q, %q, %q", lang, script, region)
	}
}

func TestMinimizeSubtags(t *testing.T) {
	t.Parallel()

	lang, script, region, ok := MinimizeSubtags("zh", "Hans", "CN")
	if !ok {
		t.Fatal("MinimizeSubtags(zh-Hans-CN) ok=false")
	}
	if lang != "zh" || script != "" || region != "" {
		t.Fatalf("MinimizeSubtags(zh-Hans-CN) = %q, %q, %q", lang, script, region)
	}
}
