package cachekey

import "testing"

func TestAppendFragmentsAndJoin(t *testing.T) {
	t.Parallel()

	parts := make([]string, 0, 4)
	parts = AppendString(parts, "style", "currency", "decimal")
	parts = AppendString(parts, "currency", "", "")
	parts = AppendInt(parts, "minimumIntegerDigits", 2, 1)
	parts = AppendInt(parts, "maximumFractionDigits", 3, 3)
	parts = AppendBool(parts, "hour12", false)

	if got, want := Join(parts), "style=currency;minimumIntegerDigits=2;hour12=false"; got != want {
		t.Fatalf("Join(parts) = %q, want %q", got, want)
	}
}

func TestAppendNonEmptyString(t *testing.T) {
	t.Parallel()

	parts := make([]string, 0, 2)
	parts = AppendNonEmptyString(parts, "dateStyle", "medium")
	parts = AppendNonEmptyString(parts, "timeStyle", "")

	if got, want := Join(parts), "dateStyle=medium"; got != want {
		t.Fatalf("Join(parts) = %q, want %q", got, want)
	}
}
