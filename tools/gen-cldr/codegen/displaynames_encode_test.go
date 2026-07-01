package codegen

import (
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestEncodeLanguageDisplayNames(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	encodeLanguageDisplayNames(&e, cldr.DisplayNames{
		Languages: cldr.LanguageDisplay{
			Dialect: cldr.StyledNames{
				Long:   map[string]string{"aa": "A"},
				Short:  map[string]string{"bb": "B"},
				Narrow: map[string]string{"cc": "C"},
			},
			Standard: cldr.StyledNames{
				Long:   map[string]string{"dd": "D"},
				Short:  map[string]string{"ee": "E"},
				Narrow: map[string]string{"ff": "F"},
			},
		},
		LocalePattern: "{0} ({1})",
	}, table)

	want := []byte{
		1, 0, 2, 2, 1,
		1, 3, 2, 5, 1,
		1, 6, 2, 8, 1,
		1, 9, 2, 11, 1,
		1, 12, 2, 14, 1,
		1, 15, 2, 17, 1,
		18, 9,
	}
	assertBytesEqual(t, "encodeLanguageDisplayNames() bytes", e.bytes(), want)
}

func TestEncodeStyledNames(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	encodeStyledNames(&e, cldr.StyledNames{
		Long:   map[string]string{"aa": "A"},
		Short:  map[string]string{"bb": "B"},
		Narrow: map[string]string{"cc": "C"},
	}, table)

	want := []byte{
		1, 0, 2, 2, 1,
		1, 3, 2, 5, 1,
		1, 6, 2, 8, 1,
	}
	assertBytesEqual(t, "encodeStyledNames() bytes", e.bytes(), want)
}
