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
	re := regexp.MustCompile(`^[0-9+\-/*().]+$`) // проверяет что даны только цифры и мат символы
	return re.MatchString(removeSpaces(s))
}
