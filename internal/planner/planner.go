package planner

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

func PlanTasks(expressionID string, raw string) (int, error) {
	expression := strings.ReplaceAll(raw, " ", "")

	for {
		i := findFirstMulDiv(expression)
		if i == -1 {
			break
		}
		leftVal, leftTaskID, startPos, err := findOperandLeft(expression, i)
		if err != nil {
			return 0, err
		}
		rightVal, rightTaskID, endPos, err := findOperandRight(expression, i)
		if err != nil {
			return 0, err
		}
		op := string(expression[i])
		task, err := repository.CreateTaskWithArgs(
			expressionID,
			op,
			leftVal, leftTaskID,
			rightVal, rightTaskID,
		)
		if err != nil {
			return 0, err
		}

		newPart := fmt.Sprintf("T%d", task.ID)
		expression = expression[:startPos] + newPart + expression[endPos:]
	}

	for {
		j := findFirstAddSub(expression)
		if j == -1 {
			break
		}
		leftVal, leftTaskID, startPos, err := findOperandLeft(expression, j)
		if err != nil {
			return 0, err
		}
		rightVal, rightTaskID, endPos, err := findOperandRight(expression, j)
		if err != nil {
			return 0, err
		}
		op := string(expression[j])
		task, err := repository.CreateTaskWithArgs(
			expressionID,
			op,
			leftVal, leftTaskID,
			rightVal, rightTaskID,
		)
		if err != nil {
			return 0, err
		}

		newPart := fmt.Sprintf("T%d", task.ID)
		expression = expression[:startPos] + newPart + expression[endPos:]
	}

	if strings.HasPrefix(expression, "T") {
		tidStr := expression[1:]
		tid, err := strconv.Atoi(tidStr)
		if err != nil {
			return 0, fmt.Errorf("cannot parse final task id: %s", expression)
		}
		return tid, nil
	}

	val, err := strconv.ParseFloat(expression, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse expression: %s", expression)
	}
	t, err := repository.CreateTaskWithArgs(
		expressionID,
		"+",
		&val, nil,
		nil, nil,
	)
	if err != nil {
		return 0, err
	}
	return t.ID, nil
}

func findFirstMulDiv(expr string) int {
	for i, c := range expr {
		if c == '*' || c == '/' {
			return i
		}
	}
	return -1
}
func findFirstAddSub(expr string) int {
	for i, c := range expr {
		if c == '+' || c == '-' {
			return i
		}
	}
	return -1
}

// (value *float64, taskID *int, startPos, err).
func findOperandLeft(expr string, opPos int) (*float64, *int, int, error) {
	start := opPos - 1
	for start >= 0 && !strings.ContainsRune("+-*/", rune(expr[start])) {
		start--
	}
	leftStr := expr[start+1 : opPos]
	startPos := start + 1
	if strings.HasPrefix(leftStr, "T") {
		tidStr := leftStr[1:]
		tid, err := strconv.Atoi(tidStr)
		if err != nil {
			return nil, nil, 0, err
		}
		return nil, &tid, startPos, nil
	}
	val, err := strconv.ParseFloat(leftStr, 64)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("invalid left operand: %s", leftStr)
	}
	return &val, nil, startPos, nil
}

func findOperandRight(expr string, opPos int) (*float64, *int, int, error) {
	end := opPos + 1
	for end < len(expr) && !strings.ContainsRune("+-*/", rune(expr[end])) {
		end++
	}
	rightStr := expr[opPos+1 : end]
	endPos := end
	if strings.HasPrefix(rightStr, "T") {
		tidStr := rightStr[1:]
		tid, err := strconv.Atoi(tidStr)
		if err != nil {
			return nil, nil, 0, err
		}
		return nil, &tid, endPos, nil
	}
	val, err := strconv.ParseFloat(rightStr, 64)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("invalid right operand: %s", rightStr)
	}
	return &val, nil, endPos, nil
}

func PlanTasksWithParen(exprID, raw string) (int, error) {
	expression := removeSpaces(raw)

	for {
		lp := findInnerParentheses(expression)
		if lp == -1 {
			break // нет ( ) больше
		}

		rp, err := findMatchingParen(expression, lp)
		if err != nil {
			return 0, fmt.Errorf("unmatched parentheses: %v", err)
		}

		subStr := expression[lp+1 : rp]
		subTaskID, err := PlanTasks(exprID, subStr)
		if err != nil {
			return 0, fmt.Errorf("cannot plan sub-expression: %v", err)
		}

		newPart := fmt.Sprintf("T%d", subTaskID)
		expression = expression[:lp] + newPart + expression[rp+1:]
	}

	finalTaskID, err := PlanTasks(exprID, expression)
	if err != nil {
		return 0, err
	}
	return finalTaskID, nil
}

func findInnerParentheses(expr string) int {
	depth := 0
	var candidate = -1
	for i, ch := range expr {
		if ch == '(' {
			depth++
			if depth == 1 {
				candidate = i
			}
		} else if ch == ')' {
			depth--
		}
	}
	return candidate
}

func findMatchingParen(expr string, lp int) (int, error) {
	depth := 0
	for i := lp; i < len(expr); i++ {
		if expr[i] == '(' {
			depth++
		} else if expr[i] == ')' {
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return -1, fmt.Errorf("no matching closing parenthesis")
}

func removeSpaces(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
