package numberformat

import "testing"

func mustDecimalValue(t testing.TB, input string) Value {
	t.Helper()

	value, err := ParseValue(input)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
