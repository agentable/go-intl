package main

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// jsSource keeps byte offsets stable while excluding non-code match starts.
// Quoted values stay intact for the existing literal parsers.
type jsSource struct {
	text   string
	code   []bool
	opaque [][2]int
	issue  string
	at     int
}

type jsScanner struct {
	source jsSource
	raw    string
	masked []byte
	i      int
}

func scanJSSource(text string) jsSource {
	s := jsScanner{raw: text, masked: []byte(text), source: jsSource{code: make([]bool, len(text))}}
	s.code(false, 0)
	s.source.text = string(s.masked)
	return s.source
}

// code only balances braces to delimit template substitutions. It does not
// evaluate expressions or infer control flow. Ambiguous slash goals fail closed.
func (s *jsScanner) code(substitution bool, depth int) {
	braces := 0
	previous := ""
	for s.i < len(s.raw) && s.source.issue == "" {
		start := s.i
		c := s.raw[s.i]
		if width := jsWhitespaceWidth(s.raw[s.i:]); width > 0 {
			for j := range width {
				if s.masked[s.i+j] != '\n' && s.masked[s.i+j] != '\r' {
					s.masked[s.i+j] = ' '
				}
			}
			s.i += width
			continue
		}
		switch {
		case strings.HasPrefix(s.raw[s.i:], "//"):
			s.i += 2
			for s.i < len(s.raw) && !jsLineTerminator(s.raw[s.i:]) {
				s.i++
			}
			s.maskComment(start)
			continue
		case strings.HasPrefix(s.raw[s.i:], "/*"):
			end := strings.Index(s.raw[s.i+2:], "*/")
			if end < 0 {
				s.fail(start, "unterminated block comment")
				return
			}
			s.i += end + 4
			s.maskComment(start)
			continue
		case c == '\'' || c == '"':
			s.quoted(c)
			previous = "value"
			continue
		case c == '`':
			s.template(depth)
			previous = "value"
			continue
		case c == '/':
			if previous == ")" || previous == "}" {
				s.fail(start, "ambiguous regular-expression or division slash")
				return
			}
			if previous != "value" && previous != "]" {
				s.regex()
				previous = "value"
				continue
			}
		case c == '}':
			if braces == 0 && substitution {
				s.i++
				return
			}
			braces--
		case c == '{':
			braces++
		}
		s.source.code[start] = previous != "."
		s.i++
		if isJSIdentifierByte(c) {
			for s.i < len(s.raw) && isJSIdentifierByte(s.raw[s.i]) && jsWhitespaceWidth(s.raw[s.i:]) == 0 {
				s.i++
			}
			switch s.raw[start:s.i] {
			case "return", "throw", "case", "delete", "void", "typeof", "instanceof", "in", "of", "yield", "await", "else", "do", "break", "continue", "debugger":
				previous = "operator"
			default:
				previous = "value"
			}
		} else {
			previous = s.raw[start:s.i]
		}
	}
	if substitution && s.source.issue == "" {
		s.fail(s.i, "unterminated template substitution")
	}
}

func (s *jsScanner) quoted(quote byte) {
	start := s.i
	s.i++
	for s.i < len(s.raw) {
		c := s.raw[s.i]
		s.i++
		switch {
		case c == '\\' && s.i < len(s.raw):
			s.i++
		case c == quote:
			return
		case c == '\n' || c == '\r':
			s.fail(start, "unterminated quoted string")
			return
		}
	}
	s.fail(start, "unterminated quoted string")
}

func (s *jsScanner) template(depth int) {
	start := s.i
	if depth >= 64 {
		s.fail(start, "template nesting exceeds extraction limit")
		return
	}
	s.i++
	for s.i < len(s.raw) && s.source.issue == "" {
		c := s.raw[s.i]
		s.i++
		switch c {
		case '\\':
			if s.i < len(s.raw) {
				s.i++
			}
		case '`':
			s.source.opaque = append(s.source.opaque, [2]int{start, s.i})
			return
		case '$':
			if s.i < len(s.raw) && s.raw[s.i] == '{' {
				s.i++
				s.code(true, depth+1)
			}
		}
	}
	if s.source.issue == "" {
		s.fail(start, "unterminated template string")
	}
}

func (s *jsScanner) regex() {
	start := s.i
	class := false
	s.i++
	for s.i < len(s.raw) {
		c := s.raw[s.i]
		s.i++
		switch c {
		case '\\':
			if s.i < len(s.raw) {
				s.i++
			}
		case '[':
			class = true
		case ']':
			class = false
		case '/':
			if !class {
				for s.i < len(s.raw) && isJSIdentifierByte(s.raw[s.i]) && jsWhitespaceWidth(s.raw[s.i:]) == 0 {
					s.i++
				}
				s.source.opaque = append(s.source.opaque, [2]int{start, s.i})
				return
			}
		case '\n', '\r':
			s.fail(start, "unterminated regular expression")
			return
		}
	}
	s.fail(start, "unterminated regular expression")
}

func (s *jsScanner) maskComment(start int) {
	for j := start; j < s.i; j++ {
		if s.masked[j] != '\n' && s.masked[j] != '\r' {
			s.masked[j] = ' '
		}
	}
}

func (s *jsScanner) fail(at int, reason string) {
	s.source.at, s.source.issue = at, reason
}

func jsWhitespaceWidth(text string) int {
	if text[0] < utf8.RuneSelf {
		if strings.IndexByte(" \t\r\n\v\f", text[0]) >= 0 {
			return 1
		}
		return 0
	}
	r, width := utf8.DecodeRuneInString(text)
	if r == '\uFEFF' || r == '\u2028' || r == '\u2029' || unicode.Is(unicode.Zs, r) {
		return width
	}
	return 0
}

func jsLineTerminator(text string) bool {
	return text[0] == '\n' || text[0] == '\r' || strings.HasPrefix(text, "\u2028") || strings.HasPrefix(text, "\u2029")
}

func isJSIdentifierByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '$' || c >= 0x80
}

func (s jsSource) matches(re *regexp.Regexp) [][]int {
	return slices.DeleteFunc(s.codeMatches(re), func(match []int) bool {
		return s.overlapsOpaque(match[0], match[1])
	})
}

// codeMatches retains declarations whose values cannot be extracted: a later
// unsupported declaration must still block reuse of an earlier configuration.
func (s jsSource) codeMatches(re *regexp.Regexp) [][]int {
	var matches [][]int
	for _, match := range re.FindAllStringSubmatchIndex(s.text, -1) {
		if !s.code[match[0]] || s.issue != "" && match[1] > s.at {
			continue
		}
		if !s.overlapsOpaque(match[0], match[0]+1) {
			matches = append(matches, match)
		}
	}
	return matches
}

func (s jsSource) overlapsOpaque(start, end int) bool {
	for _, span := range s.opaque {
		if start < span[1] && end > span[0] {
			return true
		}
	}
	return false
}

var jsExpectRE = regexp.MustCompile(`expect\s*\(`)

func (s jsSource) candidates() map[int]bool {
	positions := map[int]bool{}
	for _, match := range jsExpectRE.FindAllStringIndex(s.text, -1) {
		if s.code[match[0]] {
			positions[match[0]] = true
		}
	}
	return positions
}
