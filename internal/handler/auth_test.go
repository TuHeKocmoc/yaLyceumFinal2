package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
)

func TestGetExpression_AnotherUser_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", handler.HandleRegister)
	mux.HandleFunc("/api/v1/login", handler.HandleLogin)
	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)))
	mux.Handle("/api/v1/expressions/",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleGetExpressionByID)))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	regA := `{"login":"userA","password":"passA"}`
	respRegA, err := http.Post(srv.URL+"/api/v1/register", "application/json", strings.NewReader(regA))
	if err != nil {
		t.Fatalf("userA register error: %v", err)
	}
	if respRegA.StatusCode != http.StatusOK {
		t.Fatalf("userA register got %d, want 200", respRegA.StatusCode)
	}
	_ = respRegA.Body.Close()

	loginA := `{"login":"userA","password":"passA"}`
	respLoginA, err := http.Post(srv.URL+"/api/v1/login", "application/json", strings.NewReader(loginA))
	if err != nil {
		t.Fatalf("userA login error: %v", err)
	}
	if respLoginA.StatusCode != http.StatusOK {
		t.Fatalf("userA login got %d, want 200", respLoginA.StatusCode)
	}
	var tokenA struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(respLoginA.Body).Decode(&tokenA)
	_ = respLoginA.Body.Close()

	if tokenA.Token == "" {
		t.Fatalf("userA empty token")
	}

	calcBody := `{"expression":"2+2"}`
	reqA, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/calculate", strings.NewReader(calcBody))
	reqA.Header.Set("Content-Type", "application/json")
	reqA.Header.Set("Authorization", "Bearer "+tokenA.Token)

	client := &http.Client{}
	respCalcA, err := client.Do(reqA)
	if err != nil {
		t.Fatalf("userA POST /api/v1/calculate error: %v", err)
	}
	if respCalcA.StatusCode != http.StatusCreated {
		t.Fatalf("userA calculate got %d, want 201", respCalcA.StatusCode)
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(respCalcA.Body).Decode(&created)
	_ = respCalcA.Body.Close()

	if created.ID == "" {
		t.Fatalf("no expression ID returned for userA")
	}

	regB := `{"login":"userB","password":"passB"}`
	respRegB, err := http.Post(srv.URL+"/api/v1/register", "application/json", strings.NewReader(regB))
	if err != nil {
		t.Fatalf("userB register error: %v", err)
	}
	if respRegB.StatusCode != http.StatusOK {
		t.Fatalf("userB register got %d, want 200", respRegB.StatusCode)
	}
	_ = respRegB.Body.Close()

	loginB := `{"login":"userB","password":"passB"}`
	respLoginB, err := http.Post(srv.URL+"/api/v1/login", "application/json", strings.NewReader(loginB))
	if err != nil {
		t.Fatalf("userB login error: %v", err)
	}
	if respLoginB.StatusCode != http.StatusOK {
		t.Fatalf("userB login got %d, want 200", respLoginB.StatusCode)
	}
	var tokenB struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(respLoginB.Body).Decode(&tokenB)
	_ = respLoginB.Body.Close()
	if tokenB.Token == "" {
		t.Fatalf("userB empty token")
	}

	urlExpr := fmt.Sprintf("%s/api/v1/expressions/%s", srv.URL, created.ID)
	reqB, _ := http.NewRequest(http.MethodGet, urlExpr, nil)
	reqB.Header.Set("Authorization", "Bearer "+tokenB.Token)

	respExprB, err := client.Do(reqB)
	if err != nil {
		t.Fatalf("userB GET expression error: %v", err)
	}
	defer respExprB.Body.Close()

	if respExprB.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", respExprB.StatusCode)
	}
}

func TestNoToken_AuthEndpoint(t *testing.T) {
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

func TestLogin_WrongPassword(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", handler.HandleRegister)
	mux.HandleFunc("/api/v1/login", handler.HandleLogin)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	regBody := `{"login":"testuser","password":"goodpass"}`
	regResp, err := http.Post(srv.URL+"/api/v1/register", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if regResp.StatusCode != http.StatusOK {
		t.Fatalf("register got %d, want 200", regResp.StatusCode)
	}
	_ = regResp.Body.Close()

	badLoginBody := `{"login":"testuser","password":"wrongpass"}`
	badLoginResp, err := http.Post(srv.URL+"/api/v1/login", "application/json", strings.NewReader(badLoginBody))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	defer badLoginResp.Body.Close()

	if badLoginResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", badLoginResp.StatusCode)
	}
}
