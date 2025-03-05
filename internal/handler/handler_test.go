package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

// Тестируем POST /api/v1/calculate
func TestHandleCreateExpression(t *testing.T) {
	repository.Reset() // очищаем карты

	body := `{"expression":"2+2*2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.HandleCreateExpression(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var out struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out.ID == "" {
		t.Error("expected non-empty ID in response")
	}

	expr, _ := repository.GetExpressionByID(out.ID)
	if expr == nil {
		t.Fatalf("expression not found in repository after creation")
	}
	if expr.Raw != "2+2*2" {
		t.Errorf("expected raw=2+2*2, got %q", expr.Raw)
	}
}

// Тестируем GET /api/v1/expressions
func TestHandleGetAllExpressions(t *testing.T) {
	repository.Reset()

	_, _ = repository.CreateExpression("111+222")
	_, _ = repository.CreateExpression("333+444")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	w := httptest.NewRecorder()

	handler.HandleGetAllExpressions(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Result().StatusCode)
	}

	var out struct {
		Expressions []struct {
			ID     string   `json:"id"`
			Status string   `json:"status"`
			Result *float64 `json:"result"`
		} `json:"expressions"`
	}
	_ = json.NewDecoder(w.Body).Decode(&out)

	if len(out.Expressions) != 2 {
		t.Errorf("expected 2 expressions, got %d", len(out.Expressions))
	}

}

// POST /front/add
func TestHandleFrontAdd(t *testing.T) {
	repository.Reset()

	form := "expression=2%2B2"
	req := httptest.NewRequest(http.MethodPost, "/front/add", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	handler.HandleFrontAdd(w, req)

	if status := w.Result().StatusCode; status < 300 || status > 399 {
		t.Fatalf("expected 3xx redirect, got %d", status)
	}

	exprs, _ := repository.GetAllExpressions()
	if len(exprs) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(exprs))
	}
	if exprs[0].Raw != "2+2" {
		t.Errorf("expected raw=2+2, got %q", exprs[0].Raw)
	}
}
