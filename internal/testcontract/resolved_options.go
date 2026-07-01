package testcontract

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

type ResolvedOptionsFixture map[string]json.RawMessage

func ExpectedResolvedOptions(t testing.TB, fixture conformance.Fixture) ResolvedOptionsFixture {
	t.Helper()

	var values ResolvedOptionsFixture
	if err := json.Unmarshal(fixture.ExpectedResolved, &values); err != nil {
		t.Fatal(err)
	}
	return values
}

func AssertResolvedString(t testing.TB, values ResolvedOptionsFixture, name, got string) {
	t.Helper()

	assertResolvedValue(t, values, name, got, formatResolvedString[string])
}

func AssertResolvedOptionalString[T ~string](t testing.TB, values ResolvedOptionsFixture, name string, got *T) {
	t.Helper()

	assertResolvedOptional(t, values, name, got, formatResolvedString[T])
}

func AssertResolvedInt(t testing.TB, values ResolvedOptionsFixture, name string, got int) {
	t.Helper()

	assertResolvedValue(t, values, name, got, strconv.Itoa)
}

func AssertResolvedBool(t testing.TB, values ResolvedOptionsFixture, name string, got bool) {
	t.Helper()

	assertResolvedValue(t, values, name, got, strconv.FormatBool)
}

func AssertResolvedOptionalInt(t testing.TB, values ResolvedOptionsFixture, name string, got *int) {
	t.Helper()

	assertResolvedOptional(t, values, name, got, strconv.Itoa)
}

func assertResolvedValue[T comparable](
	t testing.TB,
	values ResolvedOptionsFixture,
	name string,
	got T,
	format func(T) string,
) {
	t.Helper()

	raw, ok := values[name]
	if !ok {
		return
	}
	var want T
	decodeExpectedResolved(t, name, raw, &want)
	if got != want {
		t.Fatalf("ResolvedOptions().%s = %s, want %s", name, format(got), format(want))
	}
}

func assertResolvedOptional[T comparable](
	t testing.TB,
	values ResolvedOptionsFixture,
	name string,
	got *T,
	format func(T) string,
) {
	t.Helper()

	raw, ok := values[name]
	if !ok {
		return
	}
	if string(raw) == "null" {
		if got != nil {
			t.Fatalf("ResolvedOptions().%s = %s, want omitted", name, format(*got))
		}
		return
	}
	var want T
	decodeExpectedResolved(t, name, raw, &want)
	if got == nil {
		t.Fatalf("ResolvedOptions().%s omitted, want %s", name, format(want))
		return
	}
	if *got != want {
		t.Fatalf("ResolvedOptions().%s = %s, want %s", name, format(*got), format(want))
	}
}

func formatResolvedString[T ~string](value T) string {
	return strconv.Quote(string(value))
}

func decodeExpectedResolved(t testing.TB, name string, raw json.RawMessage, out any) {
	t.Helper()

	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("expectedResolvedOptions.%s = %s: %v", name, raw, err)
	}
}
