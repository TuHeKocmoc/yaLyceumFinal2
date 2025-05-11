package calc

import (
	"bytes"
	"math"
	"testing"
)

func TestFindSecondOccurence(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		char       rune
		want       int
	}{
		{
			name:       "No occurrences",
			expression: "12345",
			char:       '+',
			want:       math.MaxInt64,
		},
		{
			name:       "One occurrence",
			expression: "1+2",
			char:       '+',
			want:       math.MaxInt64,
		},
		{
			name:       "Two occurrences",
			expression: "1+2+3",
			char:       '+',
			want:       3,
		},
		{
			name:       "More than two occurrences",
			expression: "1-2-3-4",
			char:       '-',
			want:       3,
		},
		{
			name:       "Empty string",
			expression: "",
			char:       '-',
			want:       math.MaxInt64,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FindSecondOccurence(tc.expression, tc.char)
			if got != tc.want {
				t.Errorf("FindSecondOccurence(%q, %q) = %d; want %d",
					tc.expression, string(tc.char), got, tc.want)
			}
		})
	}
}

func TestFindBorders(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		index      int
		wantLeft   int
		wantRight  int
		wantErr    error
	}{
		{
			name:       "Simple plus",
			expression: "3+5",
			index:      1,
			wantLeft:   0,
			wantRight:  3,
			wantErr:    nil,
		},
		{
			name:       "Operator at the end",
			expression: "123+",
			index:      3,
			wantLeft:   0,
			wantRight:  4,
			wantErr:    nil,
		},
		{
			name:       "Negative number in expression",
			expression: "10*-2",
			index:      2,
			wantLeft:   0,
			wantRight:  5,
			wantErr:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			left, right, err := FindBorders(tc.expression, tc.index)
			if (err != nil) && tc.wantErr == nil {
				t.Fatalf("FindBorders(%q, %d) got err = %v, want no error",
					tc.expression, tc.index, err)
			}
			if err == nil && tc.wantErr != nil {
				t.Fatalf("FindBorders(%q, %d) got no error, want %v",
					tc.expression, tc.index, tc.wantErr)
			}
			if left != tc.wantLeft || right != tc.wantRight {
				t.Errorf("FindBorders(%q, %d) = (%d, %d); want (%d, %d)",
					tc.expression, tc.index, left, right, tc.wantLeft, tc.wantRight)
			}
		})
	}
}

func TestFindSecondParenthesis(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantIndex  int
		wantErr    bool
	}{
		{
			name:       "No parentheses",
			expression: "123+456",
			wantIndex:  -1,
			wantErr:    true,
		},
		{
			name:       "Extra closing parenthesis",
			expression: "1+2)",
			wantIndex:  3,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FindSecondParenthesis(tc.expression)
			if (err != nil) != tc.wantErr {
				t.Fatalf("FindSecondParenthesis(%q) err=%v, wantErr=%v",
					tc.expression, err, tc.wantErr)
			}
			if got != tc.wantIndex && !tc.wantErr {
				t.Errorf("FindSecondParenthesis(%q) = %d; want %d",
					tc.expression, got, tc.wantIndex)
			}
		})
	}
}

func TestCompute(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       float64
		wantErr    bool
	}{
		{
			name:       "Simple plus",
			expression: "3+5",
			want:       8,
			wantErr:    false,
		},
		{
			name:       "Simple minus",
			expression: "10-3",
			want:       7,
			wantErr:    false,
		},
		{
			name:       "Simple multiply",
			expression: "2*3",
			want:       6,
			wantErr:    false,
		},
		{
			name:       "Simple division",
			expression: "10/2",
			want:       5,
			wantErr:    false,
		},
		{
			name:       "Division by zero",
			expression: "10/0",
			want:       0,
			wantErr:    true,
		},
		{
			name:       "Negative number, plus",
			expression: "-2+5",
			want:       3,
			wantErr:    false,
		},
		{
			name:       "Invalid expression (no operator)",
			expression: "123",
			want:       0,
			wantErr:    true,
		},
		{
			name:       "Invalid expression (incorrect syntax)",
			expression: "+2",
			want:       0,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Compute(tc.expression)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Compute(%q) err=%v, wantErr=%v",
					tc.expression, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("Compute(%q) = %v, want %v",
					tc.expression, got, tc.want)
			}
		})
	}
}

func TestRemoveParentheses(t *testing.T) {
	tests := []struct {
		name         string
		expression   string
		index        int
		wantContains string
		wantErr      bool
	}{
		{
			name:         "Simple parentheses",
			expression:   "(1+2)",
			index:        0,
			wantContains: "3.000000",
			wantErr:      false,
		},
		{
			name:         "Nested parentheses",
			expression:   "((1+2))",
			index:        0,
			wantContains: "3.000000",
			wantErr:      false,
		},
		{
			name:         "Unmatched parentheses",
			expression:   "(1+2",
			index:        0,
			wantContains: "",
			wantErr:      true,
		},
		{
			name:         "Expression inside parentheses with minus",
			expression:   "(-2+5)",
			index:        0,
			wantContains: "3.000000",
			wantErr:      false,
		},
		{
			name:         "Extra closing parenthesis",
			expression:   "1+2)",
			index:        1,
			wantContains: "",
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RemoveParentheses(tc.expression, tc.index)
			if (err != nil) != tc.wantErr {
				t.Fatalf("RemoveParentheses(%q, %d) err=%v, wantErr=%v",
					tc.expression, tc.index, err, tc.wantErr)
			}
			if !tc.wantErr && !stringsContains(got, tc.wantContains) {
				t.Errorf("RemoveParentheses(%q, %d) = %q; want substring %q",
					tc.expression, tc.index, got, tc.wantContains)
			}
		})
	}
}

func TestCalc(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       float64
		wantErr    bool
	}{
		{
			name:       "Simple expression +",
			expression: "1+2",
			want:       3,
			wantErr:    false,
		},
		{
			name:       "Expression with parentheses",
			expression: "(2+3)*4",
			want:       20,
			wantErr:    false,
		},
		{
			name:       "Nested parentheses",
			expression: "((1+2)*3)",
			want:       9,
			wantErr:    false,
		},
		{
			name:       "Division by zero in parentheses",
			expression: "(10/0)",
			want:       0,
			wantErr:    true,
		},
		{
			name:       "Multiple operators in a row with negative number",
			expression: "2--2",
			want:       4,
			wantErr:    false,
		},
		{
			name:       "Complex expression",
			expression: "3+5*2-8/4",
			want:       11,
			wantErr:    false,
		},
		{
			name:       "Incorrect expression (unsupported sequence)",
			expression: "abc+123",
			want:       0,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Calc(tc.expression)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Calc(%q) err=%v, wantErr=%v",
					tc.expression, err, tc.wantErr)
			}
			if !tc.wantErr && !floatEquals(got, tc.want) {
				t.Errorf("Calc(%q) = %v, want %v",
					tc.expression, got, tc.want)
			}
		})
	}
}

func stringsContains(haystack, needle string) bool {
	return bytes.Contains([]byte(haystack), []byte(needle))
}

func floatEquals(a, b float64) bool {
	eps := 1e-9
	return (a-b) < eps && (b-a) < eps
}
