package pattern

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestPartition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want Pattern
	}{
		{
			name: "literal placeholder literal",
			in:   "AA{0}BB",
			want: Pattern{
				{Type: Literal, Value: "AA"},
				{Type: "0"},
				{Type: Literal, Value: "BB"},
			},
		},
		{
			name: "adjacent placeholders",
			in:   "{0}{1}",
			want: Pattern{
				{Type: "0"},
				{Type: "1"},
			},
		},
		{
			name: "literal only",
			in:   "literal",
			want: Pattern{{Type: Literal, Value: "literal"}},
		},
		{
			name: "empty",
			in:   "",
			want: Pattern{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Partition(tc.in)
			if err != nil {
				t.Fatalf("Partition(%q) error = %v", tc.in, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Partition(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

func TestPartitionRejectsUnmatchedPlaceholder(t *testing.T) {
	t.Parallel()

	_, err := Partition("AA{0BB")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Partition() error = %v, want ErrInvalid", err)
	}
}

func TestFormatIndexed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     string
		values []string
		want   string
	}{
		{name: "replaces indexed placeholders", in: "{1} at {0}", values: []string{"9:00", "May 8"}, want: "May 8 at 9:00"},
		{name: "preserves placeholders inside values", in: "{0} and {1}", values: []string{"{1}", "B"}, want: "{1} and B"},
		{name: "preserves unknown placeholder", in: "{name} {0}", values: []string{"A"}, want: "{name} A"},
		{name: "preserves out of range placeholder", in: "{0} {2}", values: []string{"A"}, want: "A {2}"},
		{name: "preserves invalid pattern", in: "A {0", values: []string{"B"}, want: "A {0"},
		{name: "preserves original text when later placeholder is invalid", in: "{0} A {1", values: []string{"B"}, want: "{0} A {1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := FormatIndexed(tc.in, tc.values...); got != tc.want {
				t.Fatalf("FormatIndexed(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func FuzzPartitionRoundTrip(f *testing.F) {
	for _, seed := range []string{
		"",
		"literal",
		"{0}",
		"{0} and {1}",
		"{name} {0}",
		"{0}{1}{2}",
		"{{0}}",
		"A {0",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 256 {
			t.Skip("bounded fuzz input")
		}

		parts, err := Partition(text)
		if err != nil {
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("Partition(%q) error = %v, want ErrInvalid", text, err)
			}
			if got := FormatIndexed(text, "A", "B"); got != text {
				t.Fatalf("FormatIndexed(%q) = %q, want original invalid pattern", text, got)
			}
			return
		}

		var rebuilt strings.Builder
		for _, part := range parts {
			switch part.Type {
			case Literal:
				if part.Value == "" {
					t.Fatalf("Partition(%q) emitted empty literal part", text)
				}
				rebuilt.WriteString(part.Value)
			default:
				rebuilt.WriteByte('{')
				rebuilt.WriteString(part.Type)
				rebuilt.WriteByte('}')
			}
		}
		if got := rebuilt.String(); got != text {
			t.Fatalf("Partition(%q) rebuild = %q", text, got)
		}
	})
}

func FuzzFormatIndexedMatchesPartition(f *testing.F) {
	seeds := []struct {
		text   string
		first  string
		second string
	}{
		{text: "{1} at {0}", first: "9:00", second: "May 8"},
		{text: "{0} and {1}", first: "{1}", second: "B"},
		{text: "{name} {0}", first: "A", second: "B"},
		{text: "{00} {01}", first: "A", second: "B"},
		{text: "A {0", first: "B", second: "C"},
	}
	for _, seed := range seeds {
		f.Add(seed.text, seed.first, seed.second)
	}

	f.Fuzz(func(t *testing.T, text, first, second string) {
		if len(text) > 256 || len(first) > 128 || len(second) > 128 {
			t.Skip("bounded fuzz input")
		}

		values := []string{first, second}
		parts, err := Partition(text)
		if err != nil {
			if got := FormatIndexed(text, values...); got != text {
				t.Fatalf("FormatIndexed(%q) = %q, want original invalid pattern", text, got)
			}
			return
		}

		var want strings.Builder
		for _, part := range parts {
			switch part.Type {
			case Literal:
				want.WriteString(part.Value)
			default:
				value, ok := indexedValue(part.Type, values)
				if !ok {
					want.WriteByte('{')
					want.WriteString(part.Type)
					want.WriteByte('}')
					continue
				}
				want.WriteString(value)
			}
		}
		if got := FormatIndexed(text, values...); got != want.String() {
			t.Fatalf("FormatIndexed(%q) = %q, want %q", text, got, want.String())
		}
	})
}
