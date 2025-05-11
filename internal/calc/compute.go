package calc

import (
	"errors"
	"strconv"
	"strings"
)

func Compute(expression string) (float64, error) {
	var firstnum []rune
	var secondnum []rune
	var operator string
	firstNumEnded := false
	operators := "+-*/"

	for i, char := range expression {
		if strings.Contains(operators, string(char)) {
			temp := operator
			operator = string(char)
			if (operator == "-" && i == 0) || firstNumEnded {
				operator = temp
			} else {
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
