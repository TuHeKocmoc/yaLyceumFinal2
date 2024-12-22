package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ================ Тесты для FindSecondOccurence ================

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
			want:       9223372036854775807, // math.MaxInt64
		},
		{
			name:       "One occurrence",
			expression: "1+2",
			char:       '+',
			want:       9223372036854775807,
		},
		{
			name:       "Two occurrences",
			expression: "1+2+3",
			char:       '+',
			want:       3, // индекс второго знака '+' (с 0 начинается)
		},
		{
			name:       "More than two occurrences",
			expression: "1-2-3-4",
			char:       '-',
			want:       3, // индекс второго '-' (с 0 начинается)
		},
		{
			name:       "Empty string",
			expression: "",
			char:       '-',
			want:       9223372036854775807,
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

// ================ Тесты для FindBorders ================

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
			index:      1, // индекс '+'
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
			wantErr:    nil, // хоть это и странное выражение, но формально ошибка не возвращается
		},
		{
			name:       "Negative number in expression",
			expression: "10*-2",
			index:      2, // индекс '*'
			wantLeft:   0,
			wantRight:  5,
			wantErr:    nil,
		},
		{
			name:       "Minus operator in the beginning",
			expression: "-2+3",
			index:      2, // индекс '+'
			wantLeft:   0,
			wantRight:  4,
			wantErr:    nil,
		},
		{
			name:       "Multiple operators",
			expression: "10+2*3",
			index:      3, // индекс '+'
			wantLeft:   3,
			wantRight:  4,
			wantErr:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			left, right, err := FindBorders(tc.expression, tc.index)
			if err != nil && tc.wantErr == nil {
				t.Fatalf("FindBorders(%q, %d) unexpected error: %v",
					tc.expression, tc.index, err)
			}
			if err == nil && tc.wantErr != nil {
				t.Fatalf("FindBorders(%q, %d) expected error: %v, got nil",
					tc.expression, tc.index, tc.wantErr)
			}
			if left != tc.wantLeft || right != tc.wantRight {
				t.Errorf("FindBorders(%q, %d) = (%d, %d); want (%d, %d)",
					tc.expression, tc.index, left, right, tc.wantLeft, tc.wantRight)
			}
		})
	}
}

// ================ Тесты для FindSecondParenthesis ================

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
			wantErr:    true, // т.к. «не закрылись скобки» или нет ')'
		},
		{
			name:       "Extra closing parenthesis",
			expression: "1+2)",
			wantIndex:  3,
			wantErr:    false, // по коду, если встретил ')', а counter==0, то возвращаем i
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FindSecondParenthesis(tc.expression)
			if (err != nil) != tc.wantErr {
				t.Fatalf("FindSecondParenthesis(%q) error = %v, wantErr %v",
					tc.expression, err, tc.wantErr)
			}
			if got != tc.wantIndex && !tc.wantErr {
				t.Errorf("FindSecondParenthesis(%q) = %d; want %d",
					tc.expression, got, tc.wantIndex)
			}
		})
	}
}

// ================ Тесты для Compute ================

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
				t.Fatalf("Compute(%q) error = %v, wantErr %v",
					tc.expression, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("Compute(%q) = %v, want %v",
					tc.expression, got, tc.want)
			}
		})
	}
}

// ================ Тесты для RemoveParentheses ================

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
			wantContains: "3.000000", // 1+2 должно вычислиться
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RemoveParentheses(tc.expression, tc.index)
			if (err != nil) != tc.wantErr {
				t.Fatalf("RemoveParentheses(%q, %d) error = %v, wantErr %v",
					tc.expression, tc.index, err, tc.wantErr)
			}
			if !tc.wantErr && !stringsContains(got, tc.wantContains) {
				t.Errorf("RemoveParentheses(%q, %d) got = %q; want substring %q",
					tc.expression, tc.index, got, tc.wantContains)
			}
		})
	}
}

// небольшая утилита, чтобы проверять, содержится ли подстрока
func stringsContains(haystack, needle string) bool {
	return bytes.Contains([]byte(haystack), []byte(needle))
}

// ================ Тесты для Calc ================

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
			expression: "(2+3)*4", // = 5*4 = 20
			want:       20,
			wantErr:    false,
		},
		{
			name:       "Nested parentheses",
			expression: "((1+2)*3)", // = 3*3 = 9
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
			expression: "2--2", // 2 - (-2) = 4
			want:       4,
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
				t.Fatalf("Calc(%q) error = %v, wantErr %v",
					tc.expression, err, tc.wantErr)
			}
			if !tc.wantErr && !floatEquals(got, tc.want) {
				t.Errorf("Calc(%q) = %v, want %v",
					tc.expression, got, tc.want)
			}
		})
	}
}

// сравнение с учётом возможных плавающих точек
func floatEquals(a, b float64) bool {
	eps := 1e-9
	return (a-b) < eps && (b-a) < eps
}

// ================ Тесты для checkInput ================

func TestCheckInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "Valid with digits and operators",
			input: "123+45-(6*7)/8",
			want:  true,
		},
		{
			name:  "Invalid with letter",
			input: "1+2a",
			want:  false,
		},
		{
			name:  "Empty string",
			input: "",
			want:  false,
		},
		{
			name:  "Contains space",
			input: "1 + 2",
			want:  false,
		},
		{
			name:  "Valid parentheses",
			input: "(1)-(2)",
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkInput(tc.input)
			if got != tc.want {
				t.Errorf("checkInput(%q) = %v, want %v",
					tc.input, got, tc.want)
			}
		})
	}
}

// ================ Тесты для CalcHandler ================

func TestCalcHandler(t *testing.T) {
	// Вспомогательная функция для отправки запроса
	sendRequest := func(method, url string, body []byte) (*httptest.ResponseRecorder, error) {
		req := httptest.NewRequest(method, url, bytes.NewReader(body))
		w := httptest.NewRecorder()

		// Вызов нашего хендлера
		CalcHandler(w, req)
		return w, nil
	}

	t.Run("Valid JSON, valid expression", func(t *testing.T) {
		requestBody := Request{Expression: "2+2"}
		jsonBody, _ := json.Marshal(requestBody)

		w, _ := sendRequest("POST", "/api/v1/calculate", jsonBody)
		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("Status = %d, want %d", w.Result().StatusCode, http.StatusOK)
		}
		var out Output
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("unmarshal error: %v, body=%s", err, w.Body.String())
		}
		// Проверяем, что вернулся корректный результат (4)
		if !floatEquals(out.Result, 4) {
			t.Errorf("Result = %f, want %f", out.Result, 4.0)
		}
	})

	t.Run("Valid JSON, invalid expression (буква)", func(t *testing.T) {
		requestBody := Request{Expression: "2+2a"}
		jsonBody, _ := json.Marshal(requestBody)

		w, _ := sendRequest("POST", "/api/v1/calculate", jsonBody)
		if w.Result().StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("Status = %d, want %d", w.Result().StatusCode, http.StatusUnprocessableEntity)
		}
		var errResp Err
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if errResp.Error != "Expression is not valid" {
			t.Errorf("Error = %q, want %q", errResp.Error, "Expression is not valid")
		}
	})

	t.Run("Valid JSON, leads to internal error (деление на ноль)", func(t *testing.T) {
		requestBody := Request{Expression: "10/0"}
		jsonBody, _ := json.Marshal(requestBody)

		w, _ := sendRequest("POST", "/api/v1/calculate", jsonBody)
		if w.Result().StatusCode != http.StatusInternalServerError {
			t.Fatalf("Status = %d, want %d", w.Result().StatusCode, http.StatusInternalServerError)
		}
		var errResp Err
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if errResp.Error != "Internal server error" {
			t.Errorf("Error = %q, want %q", errResp.Error, "Internal server error")
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		// Преднамеренно ломаем JSON
		invalidBody := []byte(`{ expression: 2+2 }`)

		w, _ := sendRequest("POST", "/api/v1/calculate", invalidBody)
		if w.Result().StatusCode != http.StatusInternalServerError {
			t.Fatalf("Status = %d, want %d", w.Result().StatusCode, http.StatusInternalServerError)
		}
		var errResp Err
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if errResp.Error != "Internal server error" {
			t.Errorf("Error = %q, want %q", errResp.Error, "Internal server error")
		}
	})
}
