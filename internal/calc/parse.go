package calc

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func FindSecondOccurence(expression string, char rune) int {
	n := 2
	for i, c := range expression {
		if c == char {
			n--
			if n == 0 {
				return i
			}
		}
	}
	return math.MaxInt64
}

func FindSecondParenthesis(expression string) (int, error) {
	counter := 0
	for i, c := range expression {
		if c == '(' {
			counter++
		}
		if c == ')' {
			if counter > 0 {
				counter--
			} else {
				return i, nil
			}
		}
	}
	return -1, errors.New("parenthesis were not closed: " + expression)
}

func FindBorders(expression string, index int) (int, int, error) {
	operators := "+-*/"
	left := 0
	right := len(expression)

	for i := index + 1; i < len(expression); i++ {
		if strings.Contains(operators, string(expression[i])) {
			if string(expression[i]) == "-" {
				if _, err := strconv.Atoi(string(expression[i-1])); err == nil {
					if _, err := strconv.Atoi(string(expression[i+1])); err == nil {

						right = i
						break
					}
				}
				if i != len(expression)-1 {
					if _, err := strconv.Atoi(string(expression[i+1])); err == nil {
						continue
					}
				}
			}
			right = i
			break
		}
	}

	for i := index - 1; i >= 0; i-- {
		if strings.Contains(operators, string(expression[i])) {
			if string(expression[i]) == "-" {
				if i != 0 {
					if strings.Contains(operators, string(expression[i+1])) {
						left = i - 1
						break
					}
				}
			}
			left = i + 1
			break
		}
	}

	if left != 0 {
		if string(expression[left-1]) == "-" {
			if left == 1 {
				left -= 1
			} else {
				if _, err := strconv.Atoi(string(expression[left-2])); err != nil {
					left -= 1
				}
			}
		}
	}
	return left, right, nil
}

func RemoveParentheses(expression string, index int) (string, error) {
	lp := strings.Index(expression, "(")
	if lp == -1 {
		_, err := FindSecondParenthesis(expression)
		if err != nil {
			return expression, nil
		}
		return "", errors.New("parenthesis were not opened: " + expression)
	}

	expression = strings.Replace(expression, "(", "", 1)

	rp, err := FindSecondParenthesis(expression)
	if err != nil {
		return "", err
	}
	expression = expression[:rp] + expression[rp+1:]

	lp = strings.Index(expression, "(")
	if lp != -1 {
		expression, err = RemoveParentheses(expression, lp)
		if err != nil {
			return "", err
		}
	}

	check := strings.Index(expression, ")")
	if check != -1 {
		return "", errors.New("parenthesis were not started: " + expression)
	}

	result, err := strconv.ParseFloat(expression[index:rp], 64)
	if err != nil {
		result, err = Compute(expression[index:rp])
		if err != nil {
			return "", err
		}
	}

	expression = expression[:index] + string(fmt.Sprintf("%f", result)) + expression[rp:]

	return expression, nil
}
