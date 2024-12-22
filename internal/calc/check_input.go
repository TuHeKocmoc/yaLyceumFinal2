package calc

import "regexp"

func CheckInput(s string) bool {
	re := regexp.MustCompile(`^[0-9+\-/*().]+$`) // проверяет что даны только цифры и мат символы
	if !re.MatchString(s) {
		return false
	} // проверка на валидность, отсутствие лишних символов
	return true
}
