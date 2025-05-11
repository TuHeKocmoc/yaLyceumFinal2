package main_test

import (
	"bytes"
	"encoding/json"
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
