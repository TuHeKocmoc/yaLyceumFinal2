package planner

import (
	"fmt"
	"strconv"
	"strings"

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
