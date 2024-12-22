package calc

import (
	"errors"
	"fmt"
	"math"
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
