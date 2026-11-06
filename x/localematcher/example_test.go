package localematcher_test

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/x/localematcher"
)

func ExampleNew() {
	available, err := locale.ParseList("en", "zh-Hans", "zh-Hant", "ja")
	if err != nil {
		panic(err)
	}
	matcher, err := localematcher.New(available, localematcher.Options{})
	if err != nil {
		panic(err)
	}
	for _, tag := range []string{"zh-HK", "zh-SG", "ko"} {
		requested, err := locale.ParseList(tag)
		if err != nil {
			panic(err)
		}
		matched, ok := matcher.Match(requested)
		if !ok {
			fmt.Println("no match")
			continue
		}
		fmt.Println(matched.String())
	}
	// Output:
	// zh-Hant
	// zh-Hans
	// no match
}
