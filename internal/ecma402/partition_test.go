package ecma402_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestPartitionPattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pattern string
		want    ecma402.Pattern
	}{
		{
			name:    "literal then placeholder then literal",
			pattern: "AA{0}BB",
			want: ecma402.Pattern{
				{Type: "literal", Value: "AA"},
				{Type: "0", Value: ""},
				{Type: "literal", Value: "BB"},
			},
		},
		{
			name:    "leading placeholder",
			pattern: "{0} BB",
			want: ecma402.Pattern{
				{Type: "0", Value: ""},
				{Type: "literal", Value: " BB"},
			},
		},
		{
			name:    "trailing placeholder",
			pattern: "AA {0}",
			want: ecma402.Pattern{
				{Type: "literal", Value: "AA "},
				{Type: "0", Value: ""},
			},
		},
		{
			name:    "empty pattern",
			pattern: "",
			want:    ecma402.Pattern{},
		},
		{
			name:    "no placeholders",
			pattern: "literal text",
			want: ecma402.Pattern{
				{Type: "literal", Value: "literal text"},
			},
		},
		{
			name:    "adjacent placeholders",
			pattern: "{a}{b}",
			want: ecma402.Pattern{
				{Type: "a", Value: ""},
				{Type: "b", Value: ""},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ecma402.PartitionPattern(tc.pattern)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestPartitionPattern_Unmatched(t *testing.T) {
	t.Parallel()
	_, err := ecma402.PartitionPattern("AA{0BB")
	if !errors.Is(err, ecma402.ErrInvalidOption) {
		t.Fatalf("err = %v, want errors.Is(ErrInvalidOption)", err)
	}
}
