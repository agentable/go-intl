package codegen

import (
	"strconv"
	"strings"
	"testing"
)

func TestRenderPayloadFileEmitsDataAndBlobs(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	table.Add("January")
	table.Add("February")
	binaryBlob := string([]byte{0, 'A', '\n', '"'})

	got, err := renderPayloadFile("number", table,
		payloadBlob{name: "_symbolsBlob", bytes: []byte(binaryBlob)},
		payloadBlob{name: "_patternsBlob", bytes: []byte("pattern")},
	)
	if err != nil {
		t.Fatalf("renderPayloadFile() error = %v", err)
	}
	src := string(got)
	assertSourceContainsAll(t, "renderPayloadFile() output", src,
		generatedHeader,
		"package number",
		"const _data =",
		strconv.Quote("JanuaryFebruary"),
		"const _symbolsBlob =",
		strconv.Quote(binaryBlob),
		"const _patternsBlob =",
		strconv.Quote("pattern"),
	)

	dataIndex := strings.Index(src, "const _data")
	symbolsIndex := strings.Index(src, "const _symbolsBlob")
	patternsIndex := strings.Index(src, "const _patternsBlob")
	if dataIndex < 0 || symbolsIndex < 0 || patternsIndex < 0 {
		t.Fatalf("renderPayloadFile() output missing expected consts:\n%s", src)
	}
	if !(dataIndex < symbolsIndex && symbolsIndex < patternsIndex) {
		t.Fatalf("renderPayloadFile() const order = data:%d symbols:%d patterns:%d, want data before symbols before patterns", dataIndex, symbolsIndex, patternsIndex)
	}
}

func TestRenderPayloadFileReturnsFormatError(t *testing.T) {
	t.Parallel()

	if _, err := renderPayloadFile("not-valid-package", NewStringTable()); err == nil {
		t.Fatal("renderPayloadFile() succeeded, want error")
	}
}
