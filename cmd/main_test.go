package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

func TestMain(m *testing.M) {
	os.Setenv("DB_PATH", ":memory:")

	if err := db.InitDB(); err != nil {
		log.Fatal("failed to init db:", err)
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
	return err
}

func TestIntegration_Orchestrator(t *testing.T) {
	if err := clearDB(); err != nil {
		t.Fatalf("clearDB error: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/register", handler.HandleRegister)
	mux.HandleFunc("/api/v1/login", handler.HandleLogin)

	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)))
	mux.Handle("/api/v1/expressions",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleGetAllExpressions)))
	mux.Handle("/api/v1/expressions/",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleGetExpressionByID)))

	mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.HandleGetTask(w, r)
		} else if r.Method == http.MethodPost {
			handler.HandlePostTaskResult(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	registerBody := `{"login":"testuser","password":"123"}`
	regResp, err := http.Post(server.URL+"/api/v1/register", "application/json", strings.NewReader(registerBody))
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if regResp.StatusCode != http.StatusOK {
		t.Fatalf("register got status %d, want 200", regResp.StatusCode)
	}
	_ = regResp.Body.Close()

	loginBody := `{"login":"testuser","password":"123"}`
	loginResp, err := http.Post(server.URL+"/api/v1/login", "application/json", strings.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login got status %d, want 200", loginResp.StatusCode)
	}
	var tokenData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(loginResp.Body).Decode(&tokenData)
	_ = loginResp.Body.Close()
	if tokenData.Token == "" {
		t.Fatalf("empty token after login")
	}
	t.Logf("Got token: %s", tokenData.Token)

	calcBody := `{"expression":"2+2*2"}`
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/calculate", strings.NewReader(calcBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenData.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/calculate error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status: got %d, want 201", resp.StatusCode)
	}

	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	if created.ID == "" {
		t.Fatal("expected non-empty ID in response")
	}
	t.Logf("Created expression with ID=%s", created.ID)

	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)

		getResp, err := http.Get(server.URL + "/internal/task")
		if err != nil {
			t.Fatalf("GET /internal/task error: %v", err)
		}
		if getResp.StatusCode == http.StatusNotFound {
			_ = getResp.Body.Close()
			t.Log("No tasks available - maybe done.")
			break
		}
		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status from GET /internal/task: %d", getResp.StatusCode)
		}
		var taskResp struct {
			Task struct {
				ID            int     `json:"id"`
				Arg1          float64 `json:"arg1"`
				Arg2          float64 `json:"arg2"`
				Operation     string  `json:"operation"`
				OperationTime int     `json:"operation_time"`
			} `json:"task"`
		}
		_ = json.NewDecoder(getResp.Body).Decode(&taskResp)
		getResp.Body.Close()
		t.Logf("[AGENT] Got task: ID=%d, op=%s, arg1=%.2f, arg2=%.2f",
			taskResp.Task.ID, taskResp.Task.Operation, taskResp.Task.Arg1, taskResp.Task.Arg2)

		var resVal float64
		switch taskResp.Task.Operation {
		case "+":
			resVal = taskResp.Task.Arg1 + taskResp.Task.Arg2
		case "-":
			resVal = taskResp.Task.Arg1 - taskResp.Task.Arg2
		case "*":
			resVal = taskResp.Task.Arg1 * taskResp.Task.Arg2
		case "/":
			if taskResp.Task.Arg2 == 0 {
				resVal = 0
			} else {
				resVal = taskResp.Task.Arg1 / taskResp.Task.Arg2
			}
		case "FULL":
			resVal = 6
		default:
			t.Logf("Unknown operation: %s", taskResp.Task.Operation)
			resVal = 0
		}

		bodyReq, _ := json.Marshal(map[string]interface{}{
			"id":     taskResp.Task.ID,
			"result": resVal,
		})
		postResp, err := http.Post(server.URL+"/internal/task", "application/json", bytes.NewReader(bodyReq))
		if err != nil {
			t.Fatalf("POST /internal/task error: %v", err)
		}
		if postResp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status from POST /internal/task: %d", postResp.StatusCode)
		}
		_ = postResp.Body.Close()
		t.Logf("[AGENT] posted result=%.2f for taskID=%d", resVal, taskResp.Task.ID)
	}

	expr, err := repository.GetExpressionByIDNoUserCheck(created.ID)
	if err != nil {
		t.Fatalf("GetExpressionByIDNoUserCheck error: %v", err)
	}
	if expr == nil {
		t.Fatal("expression not found in repository")
	}
	t.Logf("Expression status=%s, result=%v", expr.Status, expr.Result)
	if expr.Status != model.StatusDone {
		t.Errorf("expected status DONE, got %s", expr.Status)
	}
	if expr.Result == nil || *expr.Result != 6.0 {
		t.Errorf("expected result=6, got %v", expr.Result)
	}
}

func TestIntegration_ComplexExpressions(t *testing.T) {
	if err := clearDB(); err != nil {
		t.Fatalf("clearDB error: %v", err)
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/register", handler.HandleRegister)
	mux.HandleFunc("/api/v1/login", handler.HandleLogin)

	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)),
	)

	mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.HandleGetTask(w, r)
		} else if r.Method == http.MethodPost {
			handler.HandlePostTaskResult(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	regBody := `{"login":"complex_user","password":"secret123"}`
	respReg, err := http.Post(server.URL+"/api/v1/register", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if respReg.StatusCode != http.StatusOK {
		t.Fatalf("register got status %d, want 200", respReg.StatusCode)
	}
	_ = respReg.Body.Close()

	loginBody := `{"login":"complex_user","password":"secret123"}`
	respLogin, err := http.Post(server.URL+"/api/v1/login", "application/json", strings.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if respLogin.StatusCode != http.StatusOK {
		t.Fatalf("login got status %d, want 200", respLogin.StatusCode)
	}
	var loginData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(respLogin.Body).Decode(&loginData)
	_ = respLogin.Body.Close()
	if loginData.Token == "" {
		t.Fatalf("empty token after login")
	}

	httpClient := &http.Client{}

	tests := []struct {
		name       string
		expression string
		wantStatus string
		wantResult *float64
	}{
		{
			name:       "DoubleParen",
			expression: "((1+2)*3)",
			wantStatus: model.StatusDone,
			wantResult: floatPtr(9),
		},
		{
			name:       "ParenProduct",
			expression: "(2+3)*(4-1)",
			wantStatus: model.StatusDone,
			wantResult: floatPtr(15),
		},
		{
			name:       "UnaryMinus1",
			expression: "-2+3",
			wantStatus: model.StatusDone,
			wantResult: floatPtr(1),
		},
		{
			name:       "UnaryMinus2",
			expression: "2+-3",
			wantStatus: model.StatusDone,
			wantResult: floatPtr(-1),
		},
		{
			name:       "UnaryMinusParen",
			expression: "-(2+3)",
			wantStatus: model.StatusDone,
			wantResult: floatPtr(-5),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calcReqBody := fmt.Sprintf(`{"expression":"%s"}`, tc.expression)

			req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/calculate", bytes.NewBufferString(calcReqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+loginData.Token)

			respCalc, err := httpClient.Do(req)
			if err != nil {
				t.Fatalf("POST /api/v1/calculate error: %v", err)
			}
			if respCalc.StatusCode != http.StatusCreated {
				t.Fatalf("unexpected status from calculate: got %d, want 201", respCalc.StatusCode)
			}

			var createdExpr struct {
				ID string `json:"id"`
			}
			_ = json.NewDecoder(respCalc.Body).Decode(&createdExpr)
			_ = respCalc.Body.Close()

			if createdExpr.ID == "" {
				t.Fatalf("no expression ID returned for expression %q", tc.expression)
			}

			for i := 0; i < 10; i++ {
				time.Sleep(100 * time.Millisecond)

				getResp, err := http.Get(server.URL + "/internal/task")
				if err != nil {
					t.Fatalf("GET /internal/task error: %v", err)
				}
				if getResp.StatusCode == http.StatusNotFound {
					_ = getResp.Body.Close()
					break
				}
				if getResp.StatusCode != http.StatusOK {
					t.Fatalf("GET /internal/task status=%d", getResp.StatusCode)
				}

				var taskResp struct {
					Task struct {
						ID            int     `json:"id"`
						Arg1          float64 `json:"arg1"`
						Arg2          float64 `json:"arg2"`
						Operation     string  `json:"operation"`
						OperationTime int     `json:"operation_time"`
					} `json:"task"`
				}
				_ = json.NewDecoder(getResp.Body).Decode(&taskResp)
				_ = getResp.Body.Close()

				resVal := computeStub(taskResp.Task.Arg1, taskResp.Task.Arg2, taskResp.Task.Operation)
				bodyReq, _ := json.Marshal(map[string]interface{}{
					"id":     taskResp.Task.ID,
					"result": resVal,
				})
				postResp, err := http.Post(server.URL+"/internal/task", "application/json", bytes.NewReader(bodyReq))
				if err != nil {
					t.Fatalf("POST /internal/task error: %v", err)
				}
				if postResp.StatusCode != http.StatusOK {
					t.Fatalf("unexpected status from POST /internal/task: %d", postResp.StatusCode)
				}
				_ = postResp.Body.Close()
			}

			expr, err := repository.GetExpressionByIDNoUserCheck(createdExpr.ID)
			if err != nil {
				t.Fatalf("GetExpressionByIDNoUserCheck error: %v", err)
			}
			if expr == nil {
				t.Fatalf("expression not found")
			}

			if expr.Status != tc.wantStatus {
				t.Errorf("expression status = %s, want %s (expr=%q)", expr.Status, tc.wantStatus, tc.expression)
			}
			if tc.wantStatus == model.StatusDone {
				if expr.Result == nil {
					t.Errorf("wantResult=%v, got nil for expr=%q", tc.wantResult, tc.expression)
				} else {
					gotRes := *expr.Result
					if tc.wantResult != nil && gotRes != *tc.wantResult {
						t.Errorf("got result=%.2f, want=%.2f for expr=%q", gotRes, *tc.wantResult, tc.expression)
					}
				}
			} else if tc.wantStatus == model.StatusError {
				if expr.Result != nil {
					t.Errorf("status=ERROR, but result=%.2f != nil for expr=%q", *expr.Result, tc.expression)
				}
			}
		})
	}
}

func computeStub(a, b float64, op string) float64 {
	switch op {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		if b == 0 {
			return 0
		}
		return a / b
	case "FULL":
		return 42
	default:
		return 0
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestMultiUserExpressions(t *testing.T) {
	err := clearDB()
	if err != nil {
		t.Fatalf("clearDB error: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/register", handler.HandleRegister)
	mux.HandleFunc("/api/v1/login", handler.HandleLogin)

	mux.Handle("/api/v1/calculate",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleCreateExpression)),
	)
	mux.Handle("/api/v1/expressions",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleGetAllExpressions)))
	mux.Handle("/api/v1/expressions/",
		handler.AuthMiddleware(http.HandlerFunc(handler.HandleGetExpressionByID)))

	server := httptest.NewServer(mux)
	defer server.Close()

	tokenA := registerAndLogin(t, server.URL, "userA", "passA")

	exprA := createExpression(t, server.URL, tokenA, "2+2")
	t.Logf("UserA created expression: %s", exprA)

	tokenB := registerAndLogin(t, server.URL, "userB", "passB")

	getExprResp, err := doAuthorizedGet(t, server.URL+"/api/v1/expressions/"+exprA, tokenB)
	if err != nil {
		t.Fatalf("B GET exprA error: %v", err)
	}
	defer getExprResp.Body.Close()

	if getExprResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for B getting exprA, got %d", getExprResp.StatusCode)
	}

	exprB := createExpression(t, server.URL, tokenB, "10/2")
	t.Logf("UserB created expression: %s", exprB)

	getExprResp2, err := doAuthorizedGet(t, server.URL+"/api/v1/expressions/"+exprB, tokenA)
	if err != nil {
		t.Fatalf("A GET exprB error: %v", err)
	}
	defer getExprResp2.Body.Close()
	if getExprResp2.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for A getting exprB, got %d", getExprResp2.StatusCode)
	}

	t.Log("TestMultiUserExpressions passed - each user sees only their own expressions.")
}

func registerAndLogin(t *testing.T, baseURL, login, pass string) string {
	regBody := `{"login":"` + login + `","password":"` + pass + `"}`
	regResp, err := http.Post(baseURL+"/api/v1/register", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("register %s error: %v", login, err)
	}
	if regResp.StatusCode != http.StatusOK && regResp.StatusCode != http.StatusConflict {
		t.Fatalf("register %s got %d, want 200 or 409", login, regResp.StatusCode)
	}
	_ = regResp.Body.Close()

	loginBody := `{"login":"` + login + `","password":"` + pass + `"}`
	loginResp, err := http.Post(baseURL+"/api/v1/login", "application/json", strings.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login %s error: %v", login, err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login %s got %d, want 200", login, loginResp.StatusCode)
	}
	var tokenData struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(loginResp.Body).Decode(&tokenData)
	_ = loginResp.Body.Close()
	if tokenData.Token == "" {
		t.Fatalf("empty token for user %s", login)
	}
	return tokenData.Token
}

func createExpression(t *testing.T, baseURL, token, expr string) string {
	calcReqBody := `{"expression":"` + expr + `"}`
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/calculate", bytes.NewBufferString(calcReqBody))
	if err != nil {
		t.Fatalf("createExpression new request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("createExpression do error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createExpression got %d, want 201", resp.StatusCode)
	}

	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&created)
	if created.ID == "" {
		t.Fatalf("empty ID in createExpression response")
	}
	return created.ID
}

func doAuthorizedGet(t *testing.T, url, token string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return http.DefaultClient.Do(req)
}
