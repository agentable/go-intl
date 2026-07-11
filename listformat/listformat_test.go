package listformat

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func stringPtr[T ~string](v T) *string {
	value := string(v)
	return &value
}

func TestListFormatResolvedOptionsDefaults(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got := format.ResolvedOptions()
	if got.Locale.String() != "en" {
		t.Fatalf("ResolvedOptions().Locale = %q, want en", got.Locale.String())
	}
	if got.Type != Conjunction {
		t.Fatalf("ResolvedOptions().Type = %q, want %q", got.Type, Conjunction)
	}
	if got.Style != LongStyle {
		t.Fatalf("ResolvedOptions().Style = %q, want %q", got.Style, LongStyle)
	}
}

func TestListFormatRejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	loc := intltest.Locale(t, "en-US")
	tests := []struct {
		name         string
		opts         Options
		wantName     string
		wantValue    string
		wantExpected string
	}{
		{
			name:         "locale matcher",
			opts:         Options{LocaleMatcher: stringPtr("bad")},
			wantName:     "localeMatcher",
			wantValue:    "bad",
			wantExpected: `one of "lookup", "best fit"`,
		},
		{
			name:         "explicit empty locale matcher",
			opts:         Options{LocaleMatcher: stringPtr("")},
			wantName:     "localeMatcher",
			wantExpected: `one of "lookup", "best fit"`,
		},
		{
			name:         "type",
			opts:         Options{Type: stringPtr("bad")},
			wantName:     "type",
			wantValue:    "bad",
			wantExpected: `one of "conjunction", "disjunction", "unit"`,
		},
		{
			name:         "explicit empty type",
			opts:         Options{Type: stringPtr("")},
			wantName:     "type",
			wantExpected: `one of "conjunction", "disjunction", "unit"`,
		},
		{
			name:         "style",
			opts:         Options{Style: stringPtr("bad")},
			wantName:     "style",
			wantValue:    "bad",
			wantExpected: `one of "long", "short", "narrow"`,
		},
		{
			name:         "explicit empty style",
			opts:         Options{Style: stringPtr("")},
			wantName:     "style",
			wantExpected: `one of "long", "short", "narrow"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "listformat", intlerr.InvalidOption, tc.wantName, tc.wantValue, loc.String())
			testcontract.AssertOptionExpected(t, err, tc.wantExpected)
		})
	}
}

func TestListFormatFormatConjunctionLong(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	if got := format.Format([]string{"A", "B", "C"}); got != "A, B, and C" {
		t.Fatalf("Format(A, B, C) = %q, want %q", got, "A, B, and C")
	}
}

func TestListFormatFormatBoundaryLengths(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name string
		list []string
		want string
	}{
		{name: "empty", list: nil, want: ""},
		{name: "single", list: []string{"A"}, want: "A"},
		{name: "pair", list: []string{"A", "B"}, want: "A and B"},
		{name: "many", list: []string{"A", "B", "C", "D"}, want: "A, B, C, and D"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := format.Format(tc.list); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.list, got, tc.want)
			}
		})
	}
}

func TestListFormatFormatTreatsElementPlaceholdersAsText(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	if got := format.Format([]string{"{1}", "B"}); got != "{1} and B" {
		t.Fatalf("Format({1}, B) = %q, want %q", got, "{1} and B")
	}
}

func TestListFormatFormatEqualsFormatToPartsJoin(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name string
		list []string
	}{
		{name: "empty", list: nil},
		{name: "single", list: []string{"A"}},
		{name: "pair", list: []string{"A", "B"}},
		{name: "many", list: []string{"A", "B", "C", "D"}},
		{name: "placeholder text", list: []string{"{1}", "B"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := format.Format(tc.list)
			if want := listPartsText(format.FormatToParts(tc.list)); got != want {
				t.Fatalf("Format(%v) = %q, want joined FormatToParts %q", tc.list, got, want)
			}
		})
	}
}

func TestListFormatFormatToPartsPair(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got := format.FormatToParts([]string{"A", "B"})
	want := []Part{
		{Type: PartElement, Value: "A"},
		{Type: PartLiteral, Value: " and "},
		{Type: PartElement, Value: "B"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(A, B) = %#v, want %#v", got, want)
	}
}

func TestListFormatFormatToPartsBoundaryLengths(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name string
		list []string
		want []Part
	}{
		{name: "empty", list: nil, want: []Part{}},
		{
			name: "single",
			list: []string{"A"},
			want: []Part{{Type: PartElement, Value: "A"}},
		},
		{
			name: "many",
			list: []string{"A", "B", "C"},
			want: []Part{
				{Type: PartElement, Value: "A"},
				{Type: PartLiteral, Value: ", "},
				{Type: PartElement, Value: "B"},
				{Type: PartLiteral, Value: ", and "},
				{Type: PartElement, Value: "C"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := format.FormatToParts(tc.list); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("FormatToParts(%v) = %#v, want %#v", tc.list, got, tc.want)
			}
		})
	}
}

func listPartsText(parts []Part) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "fr-FR"), intltest.Locale(t, "en-US"), intltest.Locale(t, "zh-Hans-CN")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf()", got, []string{"fr-FR", "en-US", "zh-Hans-CN"})
}

func TestSupportedLocalesOfRejectsExplicitEmptyLocaleMatcher(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "en-US")}
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "bogus", value: "bogus"},
		{name: "explicit empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(tc.value)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("SupportedLocalesOf() error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "listformat", intlerr.InvalidOption, "localeMatcher", tc.value, "en-US")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}
}
