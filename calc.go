package main

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Найти индекс второго появление символа в строке
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

// Эта функция получает выражение и индекс оператора,
// исходя из этого она возвращает границы выражения вида [число] (оператор) [число]
func FindBorders(expression string, index int) (int, int, error) {
	operators := "+-*/"
	left := 0
	right := len(expression)
	// fmt.Println("got expression to find border: ", expression)
	// fmt.Println("got also index:", index, string(expression[index]))

	for i := index + 1; i < len(expression); i++ {
		if strings.Contains(operators, string(expression[i])) {
			if string(expression[i]) == "-" {
				if _, err := strconv.Atoi(string(expression[i-1])); err == nil {
					if _, err := strconv.Atoi(string(expression[i+1])); err == nil {
						// fmt.Println("skipping due to special case 2")
						right = i
						break
					}
				}
				// fmt.Println("first check done, i:", i+1, string(expression[i+1]))
				if i != len(expression)-1 {
					// fmt.Println("second check done")
					if _, err := strconv.Atoi(string(expression[i+1])); err == nil {
						// fmt.Println("skipping due to special case")
						continue
					}
				}
			}
			// fmt.Println("current sign:", string(expression[i]))
			right = i
			// fmt.Println("border was found using usual case")
			break
		}
	}
	// fmt.Println("found right border: ", expression[:right], "was: ", expression)

	for i := index - 1; i >= 0; i-- {
		if strings.Contains(operators, string(expression[i])) {
			if string(expression[i]) == "-" {
				if i != 0 {
					if strings.Contains(operators, string(expression[i+1])) {
						left = i - 1
						// fmt.Println("border was found using special case")
						break
					}
				}
			}
			left = i + 1
			// fmt.Println("border was found using usual case")
			break
		}
	}
	// fmt.Println("found left border: ", expression[left:right])

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
	// fmt.Println("found new left border: ", expression[left:right])

	return left, right, nil
}

// Эта функция ищет закрывающую скобку в строке, strings.Index() не подходит,
// т.к. она не может нормально обрабатывать случаи вложенных скобок
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

// Этой функции на вход должно даться выражение типа [число] (оператор) [число],
// она его посчитает и вернет float64
func Compute(expression string) (float64, error) {
	var firstnum []rune
	var secondnum []rune
	var operator string
	firstNumEnded := false
	operators := "+-*/"

	// fmt.Println("got expression to compute: ", expression)

	for i, char := range expression {
		if strings.Contains(operators, string(char)) {
			temp := operator
			operator = string(char)
			if (operator == "-" && i == 0) || firstNumEnded == true {
				// fmt.Println("got in special case with i:", i, string(expression[i]))
				// fmt.Println(operator == "-" && i == 0)
				operator = temp
			} else {
				// fmt.Println("got op and firstnum: ", string(firstnum), operator)
				firstNumEnded = true
				continue
			}
		}
		if !firstNumEnded {
			firstnum = append(firstnum, char)
		} else {
			secondnum = append(secondnum, char)
		}
	}

	// fmt.Println("got first_num:", string(firstnum))
	// fmt.Println("got second num:", string(secondnum))
	// fmt.Println("got op:", operator)

	if len(firstnum) == 0 || len(secondnum) == 0 || len(operator) == 0 {
		return 0.0, errors.New("incorrect expression" + expression)
	}

	str_firstnum := string(firstnum)
	float_firstnum, err := strconv.ParseFloat(str_firstnum, 64)
	if err != nil {
		return 0.0, errors.New("incorrect number: " + str_firstnum)
	}

	str_secondnum := string(secondnum)
	float_secondnum, err := strconv.ParseFloat(string(secondnum), 64)
	if err != nil {
		return 0.0, errors.New("incorrect number: " + str_secondnum)
	}

	if operator == "+" {
		return float_firstnum + float_secondnum, nil
	} else if operator == "-" {
		return float_firstnum - float_secondnum, nil
	} else if operator == "*" {
		return float_firstnum * float_secondnum, nil
	} else if operator == "/" {
		if float_secondnum == 0 {
			return 0.0, errors.New("division by zero")
		}
		return float_firstnum / float_secondnum, nil
	}

	return 0.0, errors.New("incorrect operator: " + operator)
}

// Эта функция рекурсивно удаляет все скобки из последовательности, а точнее -- раскрывает их
func RemoveParentheses(expression string, index int) (string, error) {
	// fmt.Println("got expression to remove parenthesis: ", expression)
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
	// fmt.Println("expression with removed parenthesis: ", expression)

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

	// fmt.Println("computed expression after removing parenthesis: ", expression)
	expression = expression[:index] + string(fmt.Sprintf("%f", result)) + expression[rp:]
	// fmt.Println("made new expression after removing parenthesis: ", expression)

	return expression, nil
}

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

func checkInput(s string) bool {
	re := regexp.MustCompile(`^[0-9+\-/*().]+$`) // проверяет что даны только цифры и мат символы
	if !re.MatchString(s) {
		return false
	} // проверка на валидность, отсутствие лишних символов
	return true
}
