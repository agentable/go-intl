package localematcher

import (
	"testing"

	"github.com/agentable/go-intl/internal/localeid"
)

func TestUnicodeExtensionValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		extension string
		key       string
		want      string
	}{
		{name: "missing prefix", extension: "u-ca-gregory", key: "ca"},
		{name: "missing key", extension: "-u-ca-gregory", key: "nu"},
		{name: "boolean keyword", extension: "-u-kn-ca-gregory", key: "kn", want: "true"},
		{name: "multi part value", extension: "-u-ca-islamic-civil-nu-arab", key: "ca", want: "islamic-civil"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := UnicodeExtensionValue(tc.extension, tc.key); got != tc.want {
				t.Fatalf("UnicodeExtensionValue(%q, %q) = %q, want %q", tc.extension, tc.key, got, tc.want)
			}
		})
	}
}

func TestInsertUnicodeExtensionAndCanonicalize(t *testing.T) {
	t.Parallel()

	if got := InsertUnicodeExtensionAndCanonicalize("en", nil); got != "en" {
		t.Fatalf("InsertUnicodeExtensionAndCanonicalize(en, nil) = %q, want en", got)
	}
	keywords := []localeid.UnicodeKeyword{
		{Key: "nu", Value: "arab"},
		{Key: "kn", Value: "true"},
		{Key: "ca", Value: "islamic-civil"},
	}
	got := InsertUnicodeExtensionAndCanonicalize("en", keywords)
	want := "en-u-ca-islamic-civil-kn-nu-arab"
	if got != want {
		t.Fatalf("InsertUnicodeExtensionAndCanonicalize() = %q, want %q", got, want)
	}
	got = InsertUnicodeExtensionAndCanonicalize("en-x-private", keywords)
	want = "en-u-ca-islamic-civil-kn-nu-arab-x-private"
	if got != want {
		t.Fatalf("InsertUnicodeExtensionAndCanonicalize(private) = %q, want %q", got, want)
	}
}
