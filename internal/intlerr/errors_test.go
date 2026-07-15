package intlerr

import (
	"errors"
	"strings"
	"testing"
)

var errDetailSentinel = errors.New("sentinel")

func TestNewWrapsErrorAndCarriesContext(t *testing.T) {
	t.Parallel()

	err := New(InvalidOption, "numberformat", "currency", "US", "en-US", errDetailSentinel)
	if !errors.Is(err, errDetailSentinel) {
		t.Fatalf("New() error = %v, want wrapped error", err)
	}
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("New() error = %T, want *Error", err)
	}
	if detail.Kind != InvalidOption || detail.Owner != "numberformat" || detail.Name != "currency" || detail.Value != "US" || detail.Locale != "en-US" {
		t.Fatalf("Error = kind %q owner %q name %q value %q locale %q, want invalidOption numberformat currency US en-US",
			detail.Kind, detail.Owner, detail.Name, detail.Value, detail.Locale)
	}
	if !errors.Is(detail.Err, errDetailSentinel) {
		t.Fatalf("Error.Err = %v, want wrapped sentinel", detail.Err)
	}
	if !errors.Is(detail.Err, ErrInvalidOption) {
		t.Fatalf("Error.Err = %v, want kind sentinel", detail.Err)
	}
	got := err.Error()
	for _, want := range []string{
		`numberformat: invalid option currency "US"`,
		`for locale "en-US"`,
		`expected a supported value for option "currency"`,
		`; got "US"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("New() text = %q, want fragment %q", got, want)
		}
	}
}

func TestNewAddsKindSentinelToUnderlyingCause(t *testing.T) {
	t.Parallel()

	err := New(InvalidValue, "pluralrules", "value", "NaN", "", errDetailSentinel)
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("New() error = %T, want *Error", err)
	}
	if !errors.Is(detail.Err, ErrInvalidValue) {
		t.Fatalf("Error.Err = %v, want kind sentinel", detail.Err)
	}
	if !errors.Is(detail.Err, errDetailSentinel) {
		t.Fatalf("Error.Err = %v, want underlying cause", detail.Err)
	}
}

func TestNewExpectedUsesCallerGuidance(t *testing.T) {
	t.Parallel()

	err := NewExpected(InvalidValue, "durationformat", "seconds", "abc", "", "a finite numeric string", ErrInvalidValue)
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("NewExpected() error = %T, want *Error", err)
	}
	if detail.Expected != "a finite numeric string" {
		t.Fatalf("Error.Expected = %q, want caller guidance", detail.Expected)
	}
	got := err.Error()
	for _, want := range []string{`expected a finite numeric string`, `; got "abc"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("NewExpected() text = %q, want fragment %q", got, want)
		}
	}
}

func TestNewInvalidValueExpectedUsesCallerGuidance(t *testing.T) {
	t.Parallel()

	err := NewInvalidValueExpected("pluralrules", "value", "NaN", "", "a finite numeric value", errDetailSentinel)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("NewInvalidValueExpected() error = %v, want ErrInvalidValue", err)
	}
	if !errors.Is(err, errDetailSentinel) {
		t.Fatalf("NewInvalidValueExpected() error = %v, want cause", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("NewInvalidValueExpected() error = %T, want *Error", err)
	}
	if detail.Expected != "a finite numeric value" {
		t.Fatalf("Error.Expected = %q, want caller guidance", detail.Expected)
	}
}

func TestNewInvalidCodeExpectedUsesCallerGuidance(t *testing.T) {
	t.Parallel()

	err := NewInvalidCodeExpected("displaynames", "region", "USA!", "", "a two-letter ASCII or three-digit region code", errDetailSentinel)
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("NewInvalidCodeExpected() error = %v, want ErrInvalidCode", err)
	}
	if !errors.Is(err, errDetailSentinel) {
		t.Fatalf("NewInvalidCodeExpected() error = %v, want cause", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("NewInvalidCodeExpected() error = %T, want *Error", err)
	}
	if detail.Expected != "a two-letter ASCII or three-digit region code" {
		t.Fatalf("Error.Expected = %q, want caller guidance", detail.Expected)
	}
}

func TestNewUnsupportedOptionExpectedUsesCallerGuidance(t *testing.T) {
	t.Parallel()

	err := NewUnsupportedOptionExpected("collator", "caseFirst", "upper", "en", `"false"`, errDetailSentinel)
	if !errors.Is(err, ErrUnsupportedOption) {
		t.Fatalf("NewUnsupportedOptionExpected() error = %v, want ErrUnsupportedOption", err)
	}
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("NewUnsupportedOptionExpected() error = %v, want errors.ErrUnsupported", err)
	}
	if !errors.Is(err, errDetailSentinel) {
		t.Fatalf("NewUnsupportedOptionExpected() error = %v, want cause", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("NewUnsupportedOptionExpected() error = %T, want *Error", err)
	}
	if detail.Expected != `"false"` {
		t.Fatalf("Error.Expected = %q, want caller guidance", detail.Expected)
	}
}

func TestErrorKindSentinelMatching(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		kind        ErrorKind
		want        error
		unsupported bool
	}{
		{name: "invalid option", kind: InvalidOption, want: ErrInvalidOption},
		{name: "unsupported option", kind: UnsupportedOption, want: ErrUnsupportedOption, unsupported: true},
		{name: "invalid value", kind: InvalidValue, want: ErrInvalidValue},
		{name: "invalid code", kind: InvalidCode, want: ErrInvalidCode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := New(tc.kind, "owner", "name", "value", "", nil)
			if !errors.Is(err, tc.want) {
				t.Fatalf("New(%s) error = %v, want errors.Is sentinel", tc.kind, err)
			}
			if got := errors.Is(err, errors.ErrUnsupported); got != tc.unsupported {
				t.Fatalf("New(%s) errors.Is(errors.ErrUnsupported) = %t, want %t", tc.kind, got, tc.unsupported)
			}
		})
	}
}

func TestNamedConstructorsWrapSentinelsAndContext(t *testing.T) {
	t.Parallel()

	cause := errors.New("cause")
	tests := []struct {
		name         string
		makeErr      func() error
		wantKind     ErrorKind
		want         error
		unsupported  bool
		wantCause    bool
		wantExpected string
	}{
		{
			name: "invalid option",
			makeErr: func() error {
				return NewInvalidOptionExpected("owner", "name", "value", "locale", `"supported"`, cause)
			},
			wantKind:     InvalidOption,
			want:         ErrInvalidOption,
			wantCause:    true,
			wantExpected: `"supported"`,
		},
		{
			name: "unsupported option",
			makeErr: func() error {
				return NewUnsupportedOptionExpected("owner", "name", "value", "locale", `"supported"`, cause)
			},
			wantKind:     UnsupportedOption,
			want:         ErrUnsupportedOption,
			unsupported:  true,
			wantCause:    true,
			wantExpected: `"supported"`,
		},
		{
			name: "invalid value",
			makeErr: func() error {
				return NewInvalidValueExpected("owner", "name", "value", "locale", "a finite value", cause)
			},
			wantKind:     InvalidValue,
			want:         ErrInvalidValue,
			wantCause:    true,
			wantExpected: "a finite value",
		},
		{
			name: "invalid code",
			makeErr: func() error {
				return NewInvalidCodeExpected("owner", "name", "value", "locale", "a valid code", cause)
			},
			wantKind:     InvalidCode,
			want:         ErrInvalidCode,
			wantCause:    true,
			wantExpected: "a valid code",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.makeErr()
			if !errors.Is(err, tc.want) {
				t.Fatalf("%s error = %v, want sentinel %v", tc.name, err, tc.want)
			}
			if got := errors.Is(err, errors.ErrUnsupported); got != tc.unsupported {
				t.Fatalf("%s errors.Is(errors.ErrUnsupported) = %t, want %t", tc.name, got, tc.unsupported)
			}
			if tc.wantCause && !errors.Is(err, cause) {
				t.Fatalf("%s error = %v, want cause", tc.name, err)
			}
			detail, ok := errors.AsType[*Error](err)
			if !ok {
				t.Fatalf("%s error = %T, want *Error", tc.name, err)
			}
			if detail.Kind != tc.wantKind || detail.Owner != "owner" || detail.Name != "name" || detail.Value != "value" || detail.Locale != "locale" {
				t.Fatalf("%s detail = %+v, want kind %s owner/name/value/locale", tc.name, detail, tc.wantKind)
			}
			if tc.wantExpected != "" && detail.Expected != tc.wantExpected {
				t.Fatalf("%s detail.Expected = %q, want %q", tc.name, detail.Expected, tc.wantExpected)
			}
		})
	}
}

func TestUnsupportedKindWithCustomCauseMatchesStdlibUnsupported(t *testing.T) {
	t.Parallel()

	err := New(UnsupportedOption, "collator", "usage", "search", "en", errDetailSentinel)
	if !errors.Is(err, ErrUnsupportedOption) {
		t.Fatalf("New(UnsupportedOption) error = %v, want ErrUnsupportedOption", err)
	}
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("New(UnsupportedOption) error = %v, want errors.ErrUnsupported", err)
	}
}

func TestDefaultExpectedGuidanceByKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind ErrorKind
		want string
	}{
		{name: "invalid option", kind: InvalidOption, want: "a supported option value"},
		{name: "unsupported option", kind: UnsupportedOption, want: "an implementation-supported option value"},
		{name: "invalid value", kind: InvalidValue, want: "a well-formed Intl value"},
		{name: "invalid code", kind: InvalidCode, want: "a well-formed code"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := New(tc.kind, "owner", "", "", "", nil)
			detail, ok := errors.AsType[*Error](err)
			if !ok {
				t.Fatalf("New(%s) error = %T, want *Error", tc.kind, err)
			}
			if detail.Expected != tc.want {
				t.Fatalf("Error.Expected = %q, want %q", detail.Expected, tc.want)
			}
		})
	}
}

func TestCustomKindHasStableFallbackBehavior(t *testing.T) {
	t.Parallel()

	err := NewExpected(ErrorKind("future"), "owner", "", "", "", "", nil)
	if errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("custom kind matched errors.ErrUnsupported")
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("NewExpected(custom) error = %T, want *Error", err)
	}
	if detail.Expected != "" {
		t.Fatalf("Error.Expected = %q, want empty fallback", detail.Expected)
	}
	if got := err.Error(); got == "" {
		t.Fatal("custom error text is empty")
	}
}

func TestSentinelErrorsMatchStructuredTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sentinel error
		kind     ErrorKind
	}{
		{name: "invalid option", sentinel: ErrInvalidOption, kind: InvalidOption},
		{name: "unsupported option", sentinel: ErrUnsupportedOption, kind: UnsupportedOption},
		{name: "invalid value", sentinel: ErrInvalidValue, kind: InvalidValue},
		{name: "invalid code", sentinel: ErrInvalidCode, kind: InvalidCode},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.sentinel.Error(); got == "" {
				t.Fatal("sentinel error text is empty")
			}
			if !errors.Is(tc.sentinel, &Error{Kind: tc.kind}) {
				t.Fatalf("%v should match structured target kind %s", tc.sentinel, tc.kind)
			}
			if errors.Is(tc.sentinel, errDetailSentinel) {
				t.Fatalf("%v unexpectedly matched unrelated sentinel", tc.sentinel)
			}
		})
	}
}
