package segmenter_test

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/segmenter"
)

// Example demonstrates Intl.Segmenter.prototype.segment from ECMA-402.
func Example() {
	words, err := segmenter.New(mustLocaleList("en"), segmenter.Options{
		Granularity: segmenter.WordGranularity,
	})
	if err != nil {
		panic(err)
	}

	for segment := range words.Segment("Hello, world!").All() {
		if segment.IsWordLike {
			fmt.Println(segment.Segment)
		}
	}

	// Output:
	// Hello
	// world
}

// Example_options demonstrates Intl.Segmenter constructor options from ECMA-402.
func Example_options() {
	sentences, err := segmenter.New(mustLocaleList("en"), segmenter.Options{
		Granularity: segmenter.SentenceGranularity,
	})
	if err != nil {
		panic(err)
	}

	for segment := range sentences.Segment("Hello. Goodbye.").All() {
		fmt.Printf("%q\n", segment.Segment)
	}

	// Output:
	// "Hello. "
	// "Goodbye."
}

// ExampleSegments_Containing demonstrates Intl.Segments.prototype.containing from ECMA-402.
func ExampleSegments_Containing() {
	words, err := segmenter.New(mustLocaleList("en"), segmenter.Options{
		Granularity: segmenter.WordGranularity,
	})
	if err != nil {
		panic(err)
	}

	segment, ok := words.Segment("Hello, world!").Containing(8)
	fmt.Println(segment.Segment, ok)

	// Output:
	// world true
}

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
