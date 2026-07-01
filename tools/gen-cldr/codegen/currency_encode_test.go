package codegen

import (
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestAppendCurrencyFraction(t *testing.T) {
	t.Parallel()

	var e blobEncoder
	appendCurrencyFraction(&e, cldr.CurrencyFraction{
		Digits:     2,
		CashDigits: 3,
		Rounding:   4,
	})

	assertBytesEqual(t, "appendCurrencyFraction() bytes", e.bytes(), []byte{2, 3, 4})
}

func TestAppendCurrencyNames(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	appendCurrencyNames(&e, cldr.CurrencyNames{
		Display:   map[string]string{"one": "a"},
		Canonical: "bb",
		Symbol:    "ccc",
		Narrow:    "dddd",
	}, table)

	assertBytesEqual(t, "appendCurrencyNames() bytes", e.bytes(), []byte{1, 0, 3, 3, 1, 4, 2, 6, 3, 9, 4})
}
