package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/calc"
)

// /internal/task (GET)
type GetTaskResponse struct {
	Task struct {
		ID            int         `json:"id"`
		Arg1          interface{} `json:"arg1"`
		Arg2          interface{} `json:"arg2"`
		Operation     string      `json:"operation"`
		OperationTime int         `json:"operation_time"`
		ExpressionID  string      `json:"expression_id,omitempty"`
	} `json:"task"`
}

// POST /internal/task
type PostTaskRequest struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

var (
	orchestratorURL = "http://localhost:8080"
	computingPower  = 1
)

func main() {
	cpStr := os.Getenv("COMPUTING_POWER")
	if cpStr == "" {
		cpStr = "1"
	}
	cp, err := strconv.Atoi(cpStr)
	if err != nil {
		log.Printf("Invalid COMPUTING_POWER=%s, use 1 by default\n", cpStr)
		cp = 1
	}
	computingPower = cp

	urlFromEnv := os.Getenv("ORCHESTRATOR_URL")
	if urlFromEnv != "" {
		orchestratorURL = urlFromEnv
	}

	log.Printf("[AGENT] Starting with %d workers. Orchestrator = %s\n", computingPower, orchestratorURL)

	for i := 0; i < computingPower; i++ {
		go worker(i)
	}

	select {}
}

// 1) GET /internal/task
// 2) 404 -- sleep 2 secs
// 3) 200 --  POST result
func worker(workerID int) {
	log.Printf("[Worker #%d] started", workerID)

	for {
		task, err := getTaskFromOrchestrator()
		if err != nil {
			log.Printf("[Worker #%d] getTask error: %v", workerID, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if task == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("[Worker #%d] got task ID=%d, op=%s, arg1=%.2f, arg2=%.2f, opTime=%d",
			workerID,
			task.Task.ID,        // <-- task.Task.ID
			task.Task.Operation, // ...
			task.Task.Arg1,
			task.Task.Arg2,
			task.Task.OperationTime,
		)

		time.Sleep(time.Duration(task.Task.OperationTime) * time.Millisecond)

		resultValue, err := compute(task.Task.Arg1, task.Task.Arg2, task.Task.Operation)
		if err != nil {
			log.Printf("[Worker #%d] compute error: %v", workerID, err)
			time.Sleep(time.Second)
			continue
		}

		err = postResultToOrchestrator(task.Task.ID, resultValue)
		if err != nil {
			log.Printf("[Worker #%d] postResult error: %v", workerID, err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("[Worker #%d] done task ID=%d, result=%.2f", workerID, task.Task.ID, resultValue)
	}
}

// GET /internal/task
//
//	404 -> (nil, nil), err -> (nil, err)
func getTaskFromOrchestrator() (*GetTaskResponse, error) {
	url := orchestratorURL + "/internal/task"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http GET error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var out GetTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	return &out, nil
}

// POST /internal/task
func postResultToOrchestrator(taskID int, result float64) error {
	url := orchestratorURL + "/internal/task"

	bodyStruct := PostTaskRequest{
		ID:     taskID,
		Result: result,
	}
	bodyBytes, err := json.Marshal(bodyStruct)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("http POST error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status from orchestrator: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %w", err)
	}

	// {"status": "ok"}
	var resultJSON struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &resultJSON); err != nil {
		return fmt.Errorf("unmarshal error: %w (body=%q)", err, string(respBody))
	}

	if resultJSON.Status != "ok" {
		return fmt.Errorf("unexpected status field: %q (body=%q)", resultJSON.Status, string(respBody))
	}

	return nil
}

// compute "arg1 op arg2"
func compute(a, b interface{}, op string) (float64, error) {
	switch op {
	case "FULL":
		// Если это "FULL" — мы ожидаем, что a.(string) содержит полное выражение
		exprStr, ok := a.(string)
		if !ok {
			return 0, fmt.Errorf("compute FULL: Arg1 is not a string")
		}
		// Вызвать ваш парсер/калькулятор
		return calc.Calc(exprStr)

	case "+", "-", "*", "/":
		// Здесь a, b — должны быть float64
		fa, ok := a.(float64)
		if !ok {
			return 0, fmt.Errorf("arg1 is not a float64")
		}
		fb, ok := b.(float64)
		if !ok {
			return 0, fmt.Errorf("arg2 is not a float64")
		}
		return computeArith(fa, fb, op)

	default:
		return 0, fmt.Errorf("unknown operation: %s", op)
	}
}

func computeArith(a, b float64, op string) (float64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	}
	return 0, fmt.Errorf("unknown op: %s", op)
}
