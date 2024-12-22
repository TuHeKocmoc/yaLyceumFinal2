package calc

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Основная функция
func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")

	// fmt.Println("expression with deleted whitespaces: ", expression)

	for strings.Contains(expression, "(") {
		lp := strings.Index(expression, "(")
		var err error
		expression, err = RemoveParentheses(expression, lp)
		if err != nil {
			return 0.0, err
		}
	}

	// fmt.Println("got expression with removed parenthesis: ", expression)

	for strings.Contains(expression, "/") || strings.Contains(expression, "*") {
		index_mul := strings.Index(expression, "*")
		index_div := strings.Index(expression, "/")
		left := -1
		right := -1
		var err error

		if index_div == -1 {
			index_div = math.MaxInt64
		} else {
			index_mul = math.MaxInt64
		}

		// fmt.Println("indexes:", index_mul, index_div)

		if index_div < index_mul {
			// fmt.Println("sent expression with div op at: ", expression, index_div)
			left, right, err = FindBorders(expression, index_div)
			if err != nil {
				return 0.0, err
			}
		} else {
			// fmt.Println("sent expression with mul op at: ", expression, index_div)
			left, right, err = FindBorders(expression, index_mul)
			if err != nil {
				return 0.0, err
			}
		}

		// fmt.Println("computation expression in borders ", expression[left:right])
		result, err := Compute(expression[left:right])
		if err != nil {
			return 0.0, err
		}

		// fmt.Println("expression before computation: ", expression)
		expression = expression[:left] + string(fmt.Sprintf("%f", result)) + expression[right:]
		// fmt.Println("made new expression after computation: ", expression)

	}

	if len(expression) > 1 {
		if string(expression[0]) == "-" && string(expression[1]) == "-" {
			expression = strings.Replace(expression, "-", "", 2)
		}
	}

	for strings.Contains(expression, "+") || (strings.Contains(expression, "-") && (strings.Index(expression, "-") != 0 || strings.Count(expression, "-") != 1)) {
		index_add := strings.Index(expression, "+")
		index_sub := strings.Index(expression, "-")
		if index_sub == 0 {
			index_sub = FindSecondOccurence(expression, '-')
		}

		left := -1
		right := -1
		var err error

		if index_add == -1 {
			index_add = math.MaxInt64
		} else {
			index_sub = math.MaxInt64
		}

		// fmt.Println("indexes:", index_add, index_sub)

		if index_sub < index_add || index_add == -1 {
			// fmt.Println("sent expression with op sub at: ", expression, index_sub)
			left, right, err = FindBorders(expression, index_sub)
			if err != nil {
				return 0.0, err
			}
		} else {
			// fmt.Println("sent expression with op pl at: ", expression, index_add)
			left, right, err = FindBorders(expression, index_add)
			if err != nil {
				return 0.0, err
			}
		}

		// fmt.Println("computation expression in borders: ", expression[left:right])
		result, err := Compute(expression[left:right])
		if err != nil {
			return 0.0, err
		}

		// fmt.Println("expression before computation: ", expression)
		expression = expression[:left] + string(fmt.Sprintf("%f", result)) + expression[right:]
		// fmt.Println("made new expression after computation: ", expression)

	}

	// fmt.Println("ready to answer: ", expression)

	res, err := strconv.ParseFloat(expression, 64)
	if err != nil {
		return 0.0, err
	}
	return res, nil
}
