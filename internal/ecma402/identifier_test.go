package ecma402_test

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/testcontract"
)

func TestIsWellFormedCurrencyCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"USD", true},
		{"usd", true},
		{"EuR", true},
		{"US", false},
		{"USDA", false},
		{"US1", false},
		{"", false},
		{"Ü€$", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsWellFormedCurrencyCode(tc.in); got != tc.want {
				t.Errorf("IsWellFormedCurrencyCode(%q) = %v, want %v",
					tc.in, got, tc.want)
			}
		})
	}
}

func TestCanonicalCurrencyCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want string
	}{
		{"USD", "USD"},
		{"usd", "USD"},
		{"EuR", "EUR"},
		{"US1", "US1"},
		{"円円円", "円円円"},
		{"", ""},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.CanonicalCurrencyCode(tc.in); got != tc.want {
				t.Fatalf("CanonicalCurrencyCode(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestApplyCurrencyCodeOptionInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       *string
		initial     string
		want        string
		wantPresent bool
	}{
		{name: "nil keeps default", initial: "USD", want: "USD"},
		{name: "canonicalizes valid input", input: stringPtr("usd"), want: "USD", wantPresent: true},
		{name: "canonicalizes malformed ASCII input", input: stringPtr("u1d"), want: "U1D", wantPresent: true},
		{name: "preserves explicit empty input", input: stringPtr(""), wantPresent: true},
		{name: "preserves non ASCII input", input: stringPtr("円円円"), want: "円円円", wantPresent: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.initial
			var present bool
			ecma402.ApplyCurrencyCodeOptionInput(&got, &present, tc.input)
			if got != tc.want || present != tc.wantPresent {
				t.Fatalf("ApplyCurrencyCodeOptionInput() = %q/%v, want %q/%v",
					got, present, tc.want, tc.wantPresent)
			}
		})
	}
}

func TestIdentifierOptionErrorsCarryExpectedGuidance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "unicode type",
			err:  ecma402.InvalidUnicodeTypeOptionError("datetimeformat", "numberingSystem", "bad!", "en-US"),
			want: "a Unicode locale extension type",
		},
		{
			name: "currency code",
			err:  ecma402.InvalidCurrencyCodeOptionError("numberformat", "currency", "US1", "en-US"),
			want: "a three-letter ASCII currency code",
		},
		{
			name: "unit identifier",
			err:  ecma402.InvalidUnitIdentifierOptionError("numberformat", "unit", "METER", "en-US"),
			want: "a sanctioned unit identifier or <unit>-per-<unit> compound",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !errors.Is(tc.err, ecma402.ErrInvalidOption) {
				t.Fatalf("error = %v, want ErrInvalidOption", tc.err)
			}
			detail, ok := errors.AsType[*ecma402.Error](tc.err)
			if !ok {
				t.Fatalf("error = %T, want *ecma402.Error", tc.err)
			}
			if detail.Expected != tc.want {
				t.Fatalf("Expected = %q, want %q", detail.Expected, tc.want)
			}
		})
	}
}

func TestIsWellFormedUnicodeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want bool
	}{
		{"gregory", true},
		{"islamic-umalqura", true},
		{"arab", true},
		{"ARAB", true},
		{"gregorian", false},
		{"a", false},
		{"ab", false},
		{"abcdefghi", false},
		{"islamic_", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsWellFormedUnicodeType(tc.in); got != tc.want {
				t.Fatalf("IsWellFormedUnicodeType(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestCanonicalUnicodeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{in: "arab", want: "arab", wantOK: true},
		{in: "ARAB-LATN", want: "arab-latn", wantOK: true},
		{in: "gregorian"},
		{in: "\u212Aarab"},
		{in: "a"},
		{in: ""},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			got, ok := ecma402.CanonicalUnicodeType(tc.in)
			if ok != tc.wantOK || got != tc.want {
				t.Fatalf("CanonicalUnicodeType(%q) = %q, %v; want %q, %v",
					tc.in, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestApplyUnicodeTypeOptionInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       *string
		initial     string
		want        string
		wantPresent bool
	}{
		{name: "nil keeps default", initial: "latn", want: "latn"},
		{name: "canonicalizes valid input", input: stringPtr("PHONEBK"), want: "phonebk", wantPresent: true},
		{name: "preserves invalid input", input: stringPtr("bad!"), want: "bad!", wantPresent: true},
		{name: "preserves explicit empty input", input: stringPtr(""), wantPresent: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.initial
			var present bool
			ecma402.ApplyUnicodeTypeOptionInput(&got, &present, tc.input)
			if got != tc.want || present != tc.wantPresent {
				t.Fatalf("ApplyUnicodeTypeOptionInput() = %q/%v, want %q/%v",
					got, present, tc.want, tc.wantPresent)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}

func TestValidateUnicodeTypeOption(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "omitted"},
		{name: "valid", value: "arab"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ecma402.ValidateUnicodeTypeOption("numberformat", "numberingSystem", tc.value, "en-US"); err != nil {
				t.Fatalf("ValidateUnicodeTypeOption() error = %v, want nil", err)
			}
		})
	}

	err := ecma402.ValidateUnicodeTypeOption("numberformat", "numberingSystem", "bad!", "en-US")
	if !errors.Is(err, ecma402.ErrInvalidOption) {
		t.Fatalf("ValidateUnicodeTypeOption() error = %v, want ErrInvalidOption", err)
	}
	detail, ok := errors.AsType[*ecma402.Error](err)
	if !ok {
		t.Fatalf("ValidateUnicodeTypeOption() error = %T, want *ecma402.Error", err)
	}
	if detail.Owner != "numberformat" || detail.Name != "numberingSystem" || detail.Value != "bad!" || detail.Locale != "en-US" {
		t.Fatalf("ValidateUnicodeTypeOption() detail = %+v, want numberformat numberingSystem bad! en-US", detail)
	}
}

func TestValidateUnicodeTypeOptionInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		present bool
		wantErr bool
	}{
		{name: "omitted empty"},
		{name: "omitted valid still validates", value: "latn"},
		{name: "present valid", value: "latn", present: true},
		{name: "present empty", present: true, wantErr: true},
		{name: "present malformed", value: "bad!", present: true, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ecma402.ValidateUnicodeTypeOptionInput("numberformat", "numberingSystem", tc.value, "en-US", tc.present)
			if tc.wantErr {
				if !errors.Is(err, ecma402.ErrInvalidOption) {
					t.Fatalf("ValidateUnicodeTypeOptionInput() error = %v, want ErrInvalidOption", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateUnicodeTypeOptionInput() error = %v, want nil", err)
			}
		})
	}
}

func TestLocalizeDigits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		numberingSystem string
		want            string
	}{
		{name: "arab", numberingSystem: "arab", want: "١٬٢٣٤٫٥"},
		{name: "fullwide", numberingSystem: "fullwide", want: "１٬２３４٫５"},
		{name: "hanidec", numberingSystem: "hanidec", want: "一٬二三四٫五"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.LocalizeDigits("1٬234٫5", tc.numberingSystem); got != tc.want {
				t.Fatalf("LocalizeDigits() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsSanctionedSimpleUnitIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"meter", true},
		{"kilometer", true},
		{"mile-scandinavian", true},
		{"percent", true},
		{"celsius", true},
		{"Meter", false}, // case-sensitive at this layer
		{"length-meter", false},
		{"foo", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsSanctionedSimpleUnitIdentifier(tc.in); got != tc.want {
				t.Errorf("IsSanctionedSimpleUnitIdentifier(%q) = %v, want %v",
					tc.in, got, tc.want)
			}
		})
	}
}

func TestIsWellFormedUnitIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"meter", true},
		{"meter-per-second", true},
		{"microsecond", true},
		{"nanosecond-per-second", true},
		{"kilometer-per-hour", true},
		{"foot-per-foot", true},
		{"METER", false},
		{"meter-PER-second", false},
		{"foo", false},
		{"meter-per-foo", false},
		{"foo-per-meter", false},
		{"meter-per-meter-per-second", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsWellFormedUnitIdentifier(tc.in); got != tc.want {
				t.Errorf("IsWellFormedUnitIdentifier(%q) = %v, want %v",
					tc.in, got, tc.want)
			}
		})
	}
}

func TestSanctionedUnitsCount(t *testing.T) {
	t.Parallel()
	if got := len(ecma402.SanctionedUnitIdentifiers()); got != 45 {
		t.Errorf("SanctionedUnitIdentifiers() count = %d, want 45", got)
	}
}

func TestSanctionedSimpleUnitIdentifiers(t *testing.T) {
	t.Parallel()

	got := ecma402.SanctionedSimpleUnitIdentifiers()
	namespaced := ecma402.SanctionedUnitIdentifiers()
	if len(got) != len(namespaced) {
		t.Fatalf("SanctionedSimpleUnitIdentifiers() length = %d, want %d", len(got), len(namespaced))
	}
	testcontract.AssertStringSliceSortedUnique(t, "SanctionedSimpleUnitIdentifiers", got)
	testcontract.AssertStringSliceContainsAll(t, "SanctionedSimpleUnitIdentifiers()", got,
		"celsius", "meter", "mile-scandinavian", "percent",
	)

	testcontract.AssertStringSliceReturnsCopy(t, "SanctionedSimpleUnitIdentifiers", ecma402.SanctionedSimpleUnitIdentifiers)
}

func TestSanctionedUnitIdentifiersReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SanctionedUnitIdentifiers", ecma402.SanctionedUnitIdentifiers)
}
