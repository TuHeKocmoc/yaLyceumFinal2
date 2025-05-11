package handler_test

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

const testUserID int64 = 999

func TestMain(m *testing.M) {
	os.Setenv("DB_PATH", ":memory:")

	if err := db.InitDB(); err != nil {
		log.Fatal("failed to init in-memory db:", err)
	}

	code := m.Run()
	os.Exit(code)
}

func clearDB() error {
	_, err := db.GlobalDB.Exec("DELETE FROM tasks;")
	if err != nil {
		return err
	}
	_, err = db.GlobalDB.Exec("DELETE FROM expressions;")
	if err != nil {
		return err
	}
	return err
}

func withTestUserID(r *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(r.Context(), handler.UserIDCtxKey, userID)
	return r.WithContext(ctx)
}

func TestHandleCreateExpression(t *testing.T) {
	if err := clearDB(); err != nil {
		t.Fatalf("clearDB error: %v", err)
	}

	body := `{"expression":"2+2*2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	req = withTestUserID(req, testUserID)

	w := httptest.NewRecorder()

	handler.HandleCreateExpression(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response error: %v", err)
	}
	if out.ID == "" {
		t.Error("expected non-empty ID in response")
	}

	expr, err := repository.GetExpressionByID(testUserID, out.ID)
	if err != nil {
		t.Fatalf("GetExpressionByID error: %v", err)
	}
	if expr == nil {
		t.Fatalf("expression not found in repository after creation")
	}
	if expr.Raw != "2+2*2" {
		t.Errorf("expected raw=2+2*2, got %q", expr.Raw)
	}
}

func TestHandleGetAllExpressions(t *testing.T) {
	if err := clearDB(); err != nil {
		t.Fatalf("clearDB error: %v", err)
	}

	_, _ = repository.CreateExpression("111+222", testUserID)
	_, _ = repository.CreateExpression("333+444", testUserID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	req = withTestUserID(req, testUserID)

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
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode json error: %v", err)
	}

	if len(out.Expressions) != 2 {
		t.Errorf("expected 2 expressions, got %d", len(out.Expressions))
	}
}

func TestHandleFrontAdd(t *testing.T) {
	if err := clearDB(); err != nil {
		t.Fatalf("clearDB error: %v", err)
	}

	form := "expression=2%2B2"
	req := httptest.NewRequest(http.MethodPost, "/front/add", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req = withTestUserID(req, testUserID)

	w := httptest.NewRecorder()

	handler.HandleFrontAdd(w, req)

	if status := w.Result().StatusCode; status < 300 || status > 399 {
		t.Fatalf("expected 3xx redirect, got %d", status)
	}

	exprs, err := repository.GetAllExpressions(testUserID)
	if err != nil {
		t.Fatalf("GetAllExpressions error: %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(exprs))
	}
	if exprs[0].Raw != "2+2" {
		t.Errorf("expected raw=2+2, got %q", exprs[0].Raw)
	}
}
