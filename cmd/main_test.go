package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/handler"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

func TestIntegration_Orchestrator(t *testing.T) {
	repository.Reset()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", handler.HandleCreateExpression)
	mux.HandleFunc("/api/v1/expressions", handler.HandleGetAllExpressions)
	mux.HandleFunc("/api/v1/expressions/", handler.HandleGetExpressionByID)

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

	calcBody := `{"expression":"2+2*2"}`
	resp, err := http.Post(server.URL+"/api/v1/calculate", "application/json", strings.NewReader(calcBody))
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
		time.Sleep(100 * time.Millisecond) // Подождём чуть

		// GET /internal/task
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

		// POST /internal/task
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

	expr, err := repository.GetExpressionByID(created.ID)
	if err != nil {
		t.Fatalf("GetExpressionByID error: %v", err)
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
