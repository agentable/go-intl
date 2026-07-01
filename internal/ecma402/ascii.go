package ecma402

const asciiCaseOffset byte = 'a' - 'A'

func isASCIIAlpha(value string) bool {
	for i := range len(value) {
		if !isASCIIAlphaByte(value[i]) {
			return false
		}
	}
	return true
}

func isASCIIAlphaByte(value byte) bool {
	return isASCIIUpperByte(value) || isASCIILowerByte(value)
}

func isASCIIUpperByte(value byte) bool {
	return value >= 'A' && value <= 'Z'
}

func isASCIILowerByte(value byte) bool {
	return value >= 'a' && value <= 'z'
}

func asciiUpper(value string) string {
	for i := range len(value) {
		if isASCIILowerByte(value[i]) {
			return asciiUpperFrom(value, i)
		}
	}
	return value
}

func asciiUpperFrom(value string, start int) string {
	out := []byte(value)
	for i := start; i < len(out); i++ {
		if isASCIILowerByte(out[i]) {
			out[i] -= asciiCaseOffset
		}
	}
	return string(out)
}

func asciiLower(value string) string {
	for i := range len(value) {
		if isASCIIUpperByte(value[i]) {
			return asciiLowerFrom(value, i)
		}
	}
	return value
}

func asciiLowerFrom(value string, start int) string {
	out := []byte(value)
	for i := start; i < len(out); i++ {
		if isASCIIUpperByte(out[i]) {
			out[i] += asciiCaseOffset
		}
	}
	return string(out)
}

func asciiTitle(value string) string {
	for i := range len(value) {
		if i == 0 {
			if isASCIILowerByte(value[i]) {
				return asciiTitleFrom(value)
			}
			continue
		}
		if isASCIIUpperByte(value[i]) {
			return asciiTitleFrom(value)
		}
	}
	return value
}

func asciiTitleFrom(value string) string {
	out := []byte(value)
	if len(out) > 0 && isASCIILowerByte(out[0]) {
		out[0] -= asciiCaseOffset
	}
	for i := 1; i < len(out); i++ {
		if isASCIIUpperByte(out[i]) {
			out[i] += asciiCaseOffset
		}
	}
	return string(out)
}
