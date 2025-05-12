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

func removeSpaces(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func PlanTasksWithNestedParen(exprID, raw string) (int, error) {
	expression := rewriteUnaryMinuses(removeSpaces(raw))

	for strings.Contains(expression, "(") {
		lp, rp := findDeepestParenPair(expression)
		if lp == -1 || rp == -1 {
			return 0, fmt.Errorf("unmatched parentheses in expression: %s", expression)
		}

		subStr := expression[lp+1 : rp]
		subTaskID, err := PlanTasks(exprID, subStr)
		if err != nil {
			return 0, fmt.Errorf("error planning sub-expression %q: %v", subStr, err)
		}

		newPart := fmt.Sprintf("T%d", subTaskID)
		expression = expression[:lp] + newPart + expression[rp+1:]
	}

	return PlanTasks(exprID, expression)
}

func findDeepestParenPair(s string) (int, int) {
	maxDepth := 0
	curDepth := 0
	lpCandidate := -1
	lpResult := -1
	rpResult := -1

	for i, ch := range s {
		if ch == '(' {
			curDepth++
			if curDepth > maxDepth {
				maxDepth = curDepth
				lpCandidate = i
			} else if curDepth == maxDepth {
				lpCandidate = i
			}
		} else if ch == ')' {
			if curDepth == maxDepth {
				lpResult = lpCandidate
				rpResult = i
			}
			curDepth--
		}
	}
	return lpResult, rpResult
}

func rewriteUnaryMinuses(expr string) string {
	exprRunes := []rune(expr)
	var result []rune

	i := 0
	for i < len(exprRunes) {
		ch := exprRunes[i]

		if ch == '-' {
			if isUnaryMinus(result) {
				result = append(result, '0', '-')
				i++
				startIdx := i

				for i < len(exprRunes) {
					if isOperator(exprRunes[i]) && exprRunes[i] != '.' {
						break
					}
					if unicode.IsSpace(exprRunes[i]) {
						break
					}
					i++
				}
				sub := exprRunes[startIdx:i]
				result = append(result, sub...)
				continue
			} else {
				result = append(result, '-')
				i++
				continue
			}
		} else {
			result = append(result, ch)
			i++
		}
	}
	return string(result)
}

func isUnaryMinus(resultSoFar []rune) bool {
	if len(resultSoFar) == 0 {
		return true
	}
	prev := resultSoFar[len(resultSoFar)-1]
	if isOperator([]rune{prev}) ||
		prev == '(' {
		return true
	}
	return false
}

func isOperator(r interface{}) bool {
	switch x := r.(type) {
	case rune:
		return x == '+' || x == '-' || x == '*' || x == '/'
	case []rune:
		if len(x) == 1 {
			return isOperator(x[0])
		}
	}
	return false
}
