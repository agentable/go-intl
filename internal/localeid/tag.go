package localeid

import (
	"strings"

	"golang.org/x/text/language"
)

func Parts(tag language.Tag) (lang, script, region string) {
	base, scr, reg := tag.Raw()
	lang = base.String()
	if !scr.IsPrivateUse() {
		if s := scr.String(); s != "Zzzz" {
			script = s
		}
	}
	if !reg.IsPrivateUse() {
		if r := reg.String(); r != "ZZ" {
			region = r
		}
	}
	return lang, script, region
}

func Join(lang, script, region string) string {
	parts := []string{lang}
	if script != "" {
		parts = append(parts, script)
	}
	if region != "" {
		parts = append(parts, region)
	}
	return strings.Join(parts, "-")
}
