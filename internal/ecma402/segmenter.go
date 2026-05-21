package ecma402

// UTF16CodeUnitCount returns the number of UTF-16 code units needed to encode s.
func UTF16CodeUnitCount(s string) int {
	n := 0
	for _, r := range s {
		if r <= 0xFFFF {
			n++
			continue
		}
		n += 2
	}
	return n
}
