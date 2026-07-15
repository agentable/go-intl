package cldr

import "testing"

func TestValidateExecutableDatePattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{name: "fields and escaped apostrophe", pattern: "EEEE, MMMM d, y 'at' h:mm:ss a 'o''clock' zzzz"},
		{name: "literal apostrophe", pattern: "h 'o''clock'"},
		{name: "offset widths", pattern: "O OOOO X XXXX"},
		{name: "unterminated quote", pattern: "y 'at", wantErr: true},
		{name: "unknown field", pattern: "y Q", wantErr: true},
		{name: "unsupported month width", pattern: "MMMMMM", wantErr: true},
		{name: "unsupported offset width", pattern: "OO", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateExecutableDatePattern(tc.pattern)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateExecutableDatePattern(%q) error = %v, wantErr %t", tc.pattern, err, tc.wantErr)
			}
		})
	}
}

func TestExecutableDateSkeletonBoundary(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		skeleton string
		want     bool
	}{
		{skeleton: "yMMMd", want: true},
		{skeleton: "Hmsvvvv", want: true},
		{skeleton: "MMMMW-count-one", want: false},
		{skeleton: "yQQQ", want: false},
	} {
		if got := isExecutableDateSkeleton(tc.skeleton); got != tc.want {
			t.Errorf("isExecutableDateSkeleton(%q) = %t, want %t", tc.skeleton, got, tc.want)
		}
	}
}

func TestValidateIndexedTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		value      string
		allowThird bool
		wantErr    bool
	}{
		{name: "pair", value: "{1} at {0}"},
		{name: "append field name", value: "{0} ({2}: {1})", allowThird: true},
		{name: "missing", value: "{0}", wantErr: true},
		{name: "duplicate", value: "{0} {0} {1}", wantErr: true},
		{name: "unknown", value: "{0} {3} {1}", allowThird: true, wantErr: true},
		{name: "third rejected", value: "{0} {2} {1}", wantErr: true},
		{name: "third duplicate", value: "{0} {2} {2} {1}", allowThird: true, wantErr: true},
		{name: "unmatched open", value: "{0} {1", wantErr: true},
		{name: "unmatched close", value: "{0} {1}}", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateIndexedTemplate(tc.value, tc.allowThird)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateIndexedTemplate(%q, %t) error = %v, wantErr %t", tc.value, tc.allowThird, err, tc.wantErr)
			}
		})
	}
}
