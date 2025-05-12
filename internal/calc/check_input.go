package calc

import (
	"regexp"
	"strings"
	"unicode"
)

func removeSpaces(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func CheckInput(s string) bool {
	s = removeSpaces(s)
	if s == "" {
		return false
	}

	re := regexp.MustCompile(`^[0-9+\-*/().]+$`)
	if !re.MatchString(s) {
		return false
	}

	var stack int
	for _, r := range s {
		if r == '(' {
			stack++
		} else if r == ')' {
			stack--
			if stack < 0 {
				return false
			}
		}
	}
	if stack != 0 {
		return false
	}
	last := s[len(s)-1]
	return !strings.ContainsRune("+-*/", rune(last))
}
