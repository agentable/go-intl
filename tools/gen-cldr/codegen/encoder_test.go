package codegen

import (
	"errors"
	"testing"
)

func TestAppendLocaleDeltaRecordsSortsByKernelIndex(t *testing.T) {
	t.Parallel()

	var e blobEncoder
	var got []string
	err := e.appendLocaleDeltaRecords(
		[]string{"fr", "en", "ar"},
		map[string]uint64{"und": 0, "en": 1, "ar": 2, "fr": 3},
		func(locale string) {
			got = append(got, locale)
		},
	)
	if err != nil {
		t.Fatalf("appendLocaleDeltaRecords() error = %v", err)
	}
	assertStringSliceEqual(t, "appendLocaleDeltaRecords() order", got, []string{"en", "ar", "fr"})
	assertBytesEqual(t, "appendLocaleDeltaRecords() bytes", e.bytes(), []byte{3, 1, 1, 1})
}

func TestAppendLocaleDeltaRecordsReturnsErrorsForMissingKernelLocale(t *testing.T) {
	t.Parallel()

	var e blobEncoder
	called := false
	err := e.appendLocaleDeltaRecords(
		[]string{"fr"},
		map[string]uint64{"und": 0, "en": 1},
		func(locale string) {
			called = true
		},
	)
	assertErrorContains(t, "appendLocaleDeltaRecords()", err, `locale "fr" missing from kernel locale registry`)
	if called {
		t.Fatal("appendLocaleDeltaRecords() called encode after missing locale error")
	}
}

func TestAppendUint64DeltaSliceSortsByKey(t *testing.T) {
	t.Parallel()

	type deltaValue struct {
		key   uint64
		value string
	}
	values := []deltaValue{
		{key: 5, value: "b"},
		{key: 2, value: "a"},
		{key: 5, value: "c"},
	}
	var e blobEncoder
	var got []string
	err := appendUint64DeltaSlice(&e, values, func(value deltaValue) (uint64, error) {
		return value.key, nil
	}, func(value deltaValue) {
		got = append(got, value.value)
	})
	if err != nil {
		t.Fatalf("appendUint64DeltaSlice() error = %v", err)
	}
	assertStringSliceEqual(t, "appendUint64DeltaSlice() encoded values", got, []string{"a", "b", "c"})
	assertBytesEqual(t, "appendUint64DeltaSlice() bytes", e.bytes(), []byte{3, 2, 3, 0})
}

func TestAppendUint64DeltaSliceReturnsKeyErrorsBeforeEncoding(t *testing.T) {
	t.Parallel()

	errKey := errors.New("bad key")
	values := []uint64{3, 2}
	var e blobEncoder
	called := false
	err := appendUint64DeltaSlice(&e, values, func(value uint64) (uint64, error) {
		if value == 2 {
			return 0, errKey
		}
		return value, nil
	}, func(uint64) {
		called = true
	})
	if !errors.Is(err, errKey) {
		t.Fatalf("appendUint64DeltaSlice() error = %v, want %v", err, errKey)
	}
	if called {
		t.Fatal("appendUint64DeltaSlice() encoded values after key error")
	}
	assertBytesEqual(t, "appendUint64DeltaSlice() bytes", e.bytes(), nil)
}

func TestAppendDeltaReturnsErrorsForRegressingKeys(t *testing.T) {
	t.Parallel()

	var e blobEncoder
	if err := e.appendDelta(3); err != nil {
		t.Fatalf("appendDelta(3) error = %v", err)
	}
	err := e.appendDelta(2)
	assertErrorContains(t, "appendDelta(2)", err, "delta stream regressed: key 2 follows 3")
}

func TestAppendStringRefKeyMap(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	appendStringRefKeyMap(&e, map[string]uint64{
		"b": 9,
		"a": 7,
	}, table, func(value uint64) {
		e.appendUvarint(value)
	})

	assertBytesEqual(t, "appendStringRefKeyMap() bytes", e.bytes(), []byte{2, 0, 1, 7, 1, 1, 9})
}

func TestAppendStringRefSlice(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	e.appendStringRefSlice([]string{"aa", "b"}, table)

	assertBytesEqual(t, "appendStringRefSlice() bytes", e.bytes(), []byte{2, 0, 2, 2, 1})
}

func TestAppendStringRefMap(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	e.appendStringRefMap(map[string]string{
		"b": "B",
		"a": "A",
	}, table)

	assertBytesEqual(t, "appendStringRefMap() bytes", e.bytes(), []byte{2, 0, 1, 1, 1, 2, 1, 3, 1})
}

func TestAppendCountedSlice(t *testing.T) {
	t.Parallel()

	var e blobEncoder
	appendCountedSlice(&e, []uint64{2, 5}, func(value uint64) {
		e.appendUvarint(value)
	})

	assertBytesEqual(t, "appendCountedSlice() bytes", e.bytes(), []byte{2, 2, 5})
}
