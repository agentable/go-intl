package conformance

import "testing"

func TestFixtureSupportedLocalesOfFeature(t *testing.T) {
	t.Parallel()

	if fixture := (Fixture{Feature: FeatureSupportedLocalesOf}); !fixture.IsSupportedLocalesOf() {
		t.Fatalf("Fixture{%s}.IsSupportedLocalesOf() = false, want true", FeatureSupportedLocalesOf)
	}
	if fixture := (Fixture{Feature: "format"}); fixture.IsSupportedLocalesOf() {
		t.Fatal(`Fixture{Feature: "format"}.IsSupportedLocalesOf() = true, want false`)
	}
}

func TestFixtureRequiredExpected(t *testing.T) {
	t.Parallel()

	expected := "hello"
	if got := (Fixture{Expected: &expected}).RequiredExpected(t); got != expected {
		t.Fatalf("RequiredExpected() = %q, want %q", got, expected)
	}
}
