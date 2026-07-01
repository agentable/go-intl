package localematcher

import "github.com/agentable/go-intl/internal/localeid"

func UnicodeExtensionValue(extension, key string) string {
	ext, err := localeid.ParseUnicodeExtension(extension)
	if err != nil {
		return ""
	}
	return ext.ValueForKey(key)
}

func InsertUnicodeExtensionAndCanonicalize(loc string, keywords []localeid.UnicodeKeyword) string {
	if len(keywords) == 0 {
		return loc
	}
	return localeid.InsertUnicodeExtension(loc, nil, keywords)
}
