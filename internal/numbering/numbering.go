package numbering

import "strings"

// SimpleNumberingSystems is ECMA-402 Table 28, AvailableCanonicalNumberingSystems.
var SimpleNumberingSystems = []string{
	"adlm", "ahom", "arab", "arabext", "bali", "beng", "bhks", "brah",
	"cakm", "cham", "deva", "diak", "fullwide", "gara", "gong", "gonm",
	"gujr", "gukh", "guru", "hanidec", "hmng", "hmnp", "java", "kali",
	"kawi", "khmr", "knda", "krai", "lana", "lanatham", "laoo", "latn",
	"lepc", "limb", "mathbold", "mathdbl", "mathmono", "mathsanb",
	"mathsans", "mlym", "modi", "mong", "mroo", "mtei", "mymr",
	"mymrepka", "mymrpao", "mymrshan", "mymrtlng", "nagm", "newa",
	"nkoo", "olck", "onao", "orya", "osma", "outlined", "rohg", "saur",
	"segment", "shrd", "sind", "sinh", "sora", "sund", "sunu", "takr",
	"talu", "tamldec", "telu", "thai", "tibt", "tirh", "tnsa", "tols",
	"vaii", "wara", "wcho",
}

var digitZeroByNumberingSystem = map[string]rune{
	"adlm": 0x1E950, "ahom": 0x11730, "arab": 0x0660, "arabext": 0x06F0,
	"bali": 0x1B50, "beng": 0x09E6, "bhks": 0x11C50, "brah": 0x11066,
	"cakm": 0x11136, "cham": 0xAA50, "deva": 0x0966, "diak": 0x11950,
	"fullwide": 0xFF10, "gara": 0x10D40, "gong": 0x11DA0, "gonm": 0x11D50,
	"gujr": 0x0AE6, "gukh": 0x16130, "guru": 0x0A66, "hmng": 0x16B50,
	"hmnp": 0x1E140, "java": 0xA9D0, "kali": 0xA900, "kawi": 0x11F50,
	"khmr": 0x17E0, "knda": 0x0CE6, "krai": 0x16D70, "lana": 0x1A80,
	"lanatham": 0x1A90, "laoo": 0x0ED0, "latn": 0x0030, "lepc": 0x1C40,
	"limb": 0x1946, "mathbold": 0x1D7CE, "mathdbl": 0x1D7D8,
	"mathmono": 0x1D7F6, "mathsanb": 0x1D7EC, "mathsans": 0x1D7E2,
	"mlym": 0x0D66, "modi": 0x11650, "mong": 0x1810, "mroo": 0x16A60,
	"mtei": 0xABF0, "mymr": 0x1040, "mymrepka": 0x116DA,
	"mymrpao": 0x116D0, "mymrshan": 0x1090, "mymrtlng": 0xA9F0,
	"nagm": 0x1E4F0, "newa": 0x11450, "nkoo": 0x07C0, "olck": 0x1C50,
	"onao": 0x1E5F1, "orya": 0x0B66, "osma": 0x104A0, "outlined": 0x1CCF0,
	"rohg": 0x10D30, "saur": 0xA8D0, "segment": 0x1FBF0, "shrd": 0x111D0,
	"sind": 0x112F0, "sinh": 0x0DE6, "sora": 0x110F0, "sund": 0x1BB0,
	"sunu": 0x11BF0, "takr": 0x116C0, "talu": 0x19D0, "tamldec": 0x0BE6,
	"telu": 0x0C66, "thai": 0x0E50, "tibt": 0x0F20, "tirh": 0x114D0,
	"tnsa": 0x16AC0, "tols": 0x11DE0, "vaii": 0xA620, "wara": 0x118E0,
	"wcho": 0x1E2F0,
}

var hanidecDigits = [...]string{"〇", "一", "二", "三", "四", "五", "六", "七", "八", "九"}

// LocalizeDigits replaces ASCII decimal digits with the ECMA-402 simple digit
// set for numberingSystem. Unsupported systems are left unchanged.
func LocalizeDigits(s, numberingSystem string) string {
	if numberingSystem == "" || numberingSystem == "latn" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	changed := false
	for _, r := range s {
		if r < '0' || r > '9' {
			b.WriteRune(r)
			continue
		}
		changed = true
		digit := int(r - '0')
		if numberingSystem == "hanidec" {
			b.WriteString(hanidecDigits[digit])
			continue
		}
		zero, ok := digitZeroByNumberingSystem[numberingSystem]
		if !ok {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(zero + rune(digit))
	}
	if !changed {
		return s
	}
	return b.String()
}
