package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

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

func TestHandleCreateExpression_InvalidExpression_WithAuth(t *testing.T) {
	repository.Reset()
	clearDB()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/register", handler.HandleRegister)
	mux.HandleFunc("/api/v1/login", handler.HandleLogin)

	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)),
	)

	srv := httptest.NewServer(mux)
	defer srv.Close()
	userLogin := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
	regBody := fmt.Sprintf(`{"login":"%s","password":"secret"}`, userLogin)
	regResp, err := http.Post(srv.URL+"/api/v1/register", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if regResp.StatusCode != http.StatusOK {
		t.Fatalf("register got status %d, want 200", regResp.StatusCode)
	}
	_ = regResp.Body.Close()

	loginResp, err := http.Post(srv.URL+"/api/v1/login", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login got status %d, want 200", loginResp.StatusCode)
	}
	var loginData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(loginResp.Body).Decode(&loginData)
	_ = loginResp.Body.Close()
	if loginData.Token == "" {
		t.Fatalf("empty token in login response")
	}
	t.Logf("Got token: %s", loginData.Token)

	badExpr := `{"expression":"2+a"}`

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/calculate", bytes.NewBufferString(badExpr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginData.Token)

	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/calculate error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}

	var msg []byte
	msg, _ = io.ReadAll(resp.Body)
	t.Logf("Got body: %s", string(msg))
}

func TestHandleCreateExpression_NoToken(t *testing.T) {
	repository.Reset()

	mux := http.NewServeMux()
	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)),
	)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	body := `{"expression":"2+2"}`
	resp, err := http.Post(srv.URL+"/api/v1/calculate", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/v1/calculate error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHandleCreateExpression_BadToken(t *testing.T) {
	repository.Reset()

	mux := http.NewServeMux()
	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)),
	)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	body := `{"expression":"2+2"}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/calculate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer not_a_valid_jwt")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHandleCreateExpression_EmptyExpression_WithAuth(t *testing.T) {
	repository.Reset()
	clearDB()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/register", handler.HandleRegister)
	mux.HandleFunc("/api/v1/login", handler.HandleLogin)

	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)),
	)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	userLogin := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
	regBody := fmt.Sprintf(`{"login":"%s","password":"secret"}`, userLogin)
	regResp, err := http.Post(srv.URL+"/api/v1/register", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if regResp.StatusCode != http.StatusOK {
		t.Fatalf("register got status %d, want 200", regResp.StatusCode)
	}
	_ = regResp.Body.Close()

	loginResp, err := http.Post(srv.URL+"/api/v1/login", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login got status %d, want 200", loginResp.StatusCode)
	}
	var loginData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(loginResp.Body).Decode(&loginData)
	_ = loginResp.Body.Close()
	if loginData.Token == "" {
		t.Fatalf("empty token in login response")
	}

	emptyExpr := `{"expression":""}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/calculate", strings.NewReader(emptyExpr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginData.Token)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/calculate error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}

	var respBody []byte
	respBody, _ = io.ReadAll(resp.Body)
	t.Logf("Got body: %s", string(respBody))
}

func TestHandleCreateExpression_MissingExpressionField(t *testing.T) {
	repository.Reset()
	clearDB()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", handler.HandleRegister)
	mux.HandleFunc("/api/v1/login", handler.HandleLogin)
	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)),
	)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	login := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
	regBody := fmt.Sprintf(`{"login":"%s","password":"secret"}`, login)
	regResp, err := http.Post(srv.URL+"/api/v1/register", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if regResp.StatusCode != http.StatusOK {
		t.Fatalf("register got status %d, want 200", regResp.StatusCode)
	}
	_ = regResp.Body.Close()

	loginResp, err := http.Post(srv.URL+"/api/v1/login", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login got status %d, want 200", loginResp.StatusCode)
	}
	var loginData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(loginResp.Body).Decode(&loginData)
	_ = loginResp.Body.Close()
	if loginData.Token == "" {
		t.Fatalf("empty token in login response")
	}
	missingExprJSON := `{}`

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/calculate", bytes.NewBufferString(missingExprJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginData.Token)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/calculate error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Got body: %s", string(respBody))
}
