package ecma402_test

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestCanonicalCodeForDisplayNamesCanonicalizesValidCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  string
		code string
		want string
	}{
		{name: "language tag", typ: "language", code: "EN-us", want: "en-US"},
		{name: "structurally valid unknown language tag", typ: "language", code: "xx-CN", want: "xx-CN"},
		{name: "unknown language script region tag", typ: "language", code: "xx-hANs-sG", want: "xx-Hans-SG"},
		{name: "language variant", typ: "language", code: "en-Latn-US-variant", want: "en-Latn-US-variant"},
		{name: "deprecated language canonicalized", typ: "language", code: "iw", want: "he"},
		{name: "alpha region", typ: "region", code: "us", want: "US"},
		{name: "numeric region", typ: "region", code: "419", want: "419"},
		{name: "script", typ: "script", code: "lAtN", want: "Latn"},
		{name: "calendar", typ: "calendar", code: "ISLAMIC-UMALQURA", want: "islamic-umalqura"},
		{name: "date-time field", typ: "dateTimeField", code: "timeZoneName", want: "timeZoneName"},
		{name: "currency", typ: "currency", code: "usd", want: "USD"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ecma402.CanonicalCodeForDisplayNames(tc.typ, tc.code)
			if err != nil {
				t.Fatalf("CanonicalCodeForDisplayNames(%q, %q) error = %v, want nil", tc.typ, tc.code, err)
			}
			if got != tc.want {
				t.Fatalf("CanonicalCodeForDisplayNames(%q, %q) = %q, want %q", tc.typ, tc.code, got, tc.want)
			}
		})
	}
}

func TestCanonicalCodeForDisplayNamesRejectsInvalidCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  string
		code string
	}{
		{name: "language", typ: "language", code: ""},
		{name: "language private use", typ: "language", code: "x-private"},
		{name: "language private use suffix", typ: "language", code: "en-x-private"},
		{name: "language extension", typ: "language", code: "en-u-ca-gregory"},
		{name: "language grandfathered", typ: "language", code: "i-klingon"},
		{name: "language extlang", typ: "language", code: "zh-cmn"},
		{name: "language four-letter primary", typ: "language", code: "abcd"},
		{name: "region", typ: "region", code: "U1"},
		{name: "script", typ: "script", code: "lat1"},
		{name: "calendar", typ: "calendar", code: "gregory_foo"},
		{name: "date-time field", typ: "dateTimeField", code: "century"},
		{name: "currency", typ: "currency", code: "US1"},
		{name: "unknown type", typ: "unknown", code: "US"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ecma402.CanonicalCodeForDisplayNames(tc.typ, tc.code)
			if !errors.Is(err, ecma402.ErrInvalidOption) {
				t.Fatalf("CanonicalCodeForDisplayNames(%q, %q) error = %v, want ErrInvalidOption", tc.typ, tc.code, err)
			}
		})
	}
}
