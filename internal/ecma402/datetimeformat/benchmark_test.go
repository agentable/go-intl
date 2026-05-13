package ecma402dtf

import "testing"

func BenchmarkSkeleton_Parse(b *testing.B) {
	for b.Loop() {
		_ = Parse("yMMMd", "MMM d, y", nil, "")
	}
}

func BenchmarkSkeleton_Match(b *testing.B) {
	formats := []Formats{
		Parse("y", "y", nil, ""),
		Parse("M", "M", nil, ""),
		Parse("d", "d", nil, ""),
		Parse("yMd", "M/d/y", nil, ""),
		Parse("yMMMd", "MMM d, y", nil, ""),
		Parse("Hm", "HH:mm", nil, ""),
		Parse("Hms", "HH:mm:ss", nil, ""),
		Parse("yMMMMEEEEd", "EEEE, MMMM d, y", nil, ""),
	}
	opts := Options{Year: NumericNumeric, Month: FieldShort, Day: NumericNumeric}

	b.ReportAllocs()
	for b.Loop() {
		_ = Match(opts, formats)
	}
}
