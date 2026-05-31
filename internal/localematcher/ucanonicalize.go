package localematcher

import (
	"github.com/agentable/go-intl/internal/localeid"
)

type keyword struct {
	key   string
	value string
}

func UnicodeExtensionValue(extension, key string) string {
	ext, err := localeid.ParseUnicodeExtension(extension)
	if err != nil {
		return ""
	}
	return ext.ValueForKey(key)
}

func InsertUnicodeExtensionAndCanonicalize(loc string, keywords []keyword) string {
	if len(keywords) == 0 {
		return loc
	}
	out := make([]localeid.UnicodeKeyword, 0, len(keywords))
	for _, kw := range keywords {
		out = append(out, localeid.UnicodeKeyword{Key: kw.key, Value: kw.value})
	}
	return localeid.InsertUnicodeExtension(loc, nil, out)
}
