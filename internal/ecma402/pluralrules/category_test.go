package ecma402pr

import "testing"

func TestCategoryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cat  Category
		want string
	}{
		{Zero, "zero"},
		{One, "one"},
		{Two, "two"},
		{Few, "few"},
		{Many, "many"},
		{Other, "other"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			if got := tc.cat.String(); got != tc.want {
				t.Fatalf("Category.String() = %q, want %q", got, tc.want)
			}
		})
	}
}
