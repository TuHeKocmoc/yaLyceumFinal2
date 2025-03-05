package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/calc" // если хотим переиспользовать CheckInput
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

var (
	additionTime       int
	subtractionTime    int
	multiplicationTime int
	divisionTime       int
	fullTime           int
)

func init() {
	additionTime = getEnvAsInt("TIME_ADDITION_MS", 1000)
	subtractionTime = getEnvAsInt("TIME_SUBTRACTION_MS", 1200)
	multiplicationTime = getEnvAsInt("TIME_MULTIPLICATIONS_MS", 2000)
	divisionTime = getEnvAsInt("TIME_DIVISIONS_MS", 2500)
	fullTime = getEnvAsInt("TIME_FULL_MS", 3000)
}

func getEnvAsInt(name string, defaultVal int) int {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}

// 1) POST /api/v1/calculate
// { "expression": "2+2*2" }
type requestExpression struct {
	Expression string `json:"expression"`
}
type responseCreateExpression struct {
	ID string `json:"id"`
}

func HandleCreateExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req requestExpression
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if !calc.CheckInput(req.Expression) {
		http.Error(w, "expression is not valid", http.StatusUnprocessableEntity)
		return
	}

	expr, err := repository.CreateExpression(req.Expression)
	if err != nil {
		http.Error(w, "cannot create expression", http.StatusInternalServerError)
		return
	}

	if expr.Raw != "" {
		_, err = repository.CreateTask(expr.ID, "FULL", nil, nil)
		if err != nil {
			http.Error(w, "cannot create task", http.StatusInternalServerError)
			return
		}
		expr.Status = model.StatusInProgress
		_ = repository.UpdateExpression(expr)
	}

	resp := responseCreateExpression{ID: expr.ID}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// 2) GET /api/v1/expressions
type responseExpressionsList struct {
	Expressions []*model.Expression `json:"expressions"`
}

func HandleGetAllExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	exprs, err := repository.GetAllExpressions()
	if err != nil {
		http.Error(w, "failed to get expressions", http.StatusInternalServerError)
		return
	}

	resp := responseExpressionsList{Expressions: exprs}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// 3) GET /api/v1/expressions/:id
type responseSingleExpression struct {
	Expression *model.Expression `json:"expression"`
}

func HandleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL e.g. api/v1/expressions/123e4567-e89b-12d3-a456-426614174000
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	id := parts[3]

	expr, err := repository.GetExpressionByID(id)
	if err != nil {
		http.Error(w, "error in repository", http.StatusInternalServerError)
		return
	}
	if expr == nil {
		http.Error(w, "expression not found", http.StatusNotFound)
		return
	}

	resp := responseSingleExpression{Expression: expr}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// 4) GET /internal/task
type responseTask struct {
	Task struct {
		ID            int         `json:"id"`
		Arg1          interface{} `json:"arg1"`
		Arg2          interface{} `json:"arg2"`
		Operation     string      `json:"operation"`
		OperationTime int         `json:"operation_time"` // из переменных окружения
	} `json:"task"`
}

func HandleGetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	task, err := repository.GetNextWaitingTask()
	if err != nil {
		http.Error(w, "failed to get task", http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, "no task available", http.StatusNotFound)
		return
	}
	task.Status = model.TaskStatusInProgress
	if err := repository.UpdateTask(task); err != nil {
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	operationTime := getOperationTime(task.Op)

	var arg1, arg2 interface{}
	arg1 = 0.0
	arg2 = 0.0

	if task.Op == "FULL" {
		expr, err := repository.GetExpressionByID(task.ExpressionID)
		if err != nil {
			http.Error(w, "failed to get expression", http.StatusInternalServerError)
			return
		}
		if expr == nil {
			http.Error(w, "expression not found", http.StatusNotFound)
			return
		}
		arg1 = expr.Raw
		arg2 = nil
	}

	resp := responseTask{}
	resp.Task.ID = task.ID
	resp.Task.Arg1 = arg1
	resp.Task.Arg2 = arg2
	resp.Task.Operation = task.Op
	resp.Task.OperationTime = operationTime

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// 5) POST /internal/task
// { "id": 101, "result": 6.0 }
type rawTaskResult struct {
	ID     json.Number `json:"id"`
	Result json.Number `json:"result"`
}

func HandlePostTaskResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var raw rawTaskResult
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "invalid json syntax", http.StatusBadRequest)
		return
	}

	idInt64, err := raw.ID.Int64()
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusUnprocessableEntity)
		return
	}
	resultFloat64, err := raw.Result.Float64()
	if err != nil {
		http.Error(w, "result must be a float", http.StatusUnprocessableEntity)
		return
	}

	id := int(idInt64)

	task, err := repository.GetTaskByID(id)
	if err != nil {
		http.Error(w, "error in repository", http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if task.Status != model.TaskStatusInProgress {
		http.Error(w, "task not in progress", http.StatusBadRequest)
		return
	}

	task.Status = model.TaskStatusDone
	task.Result = &resultFloat64
	if err := repository.UpdateTask(task); err != nil {
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	expr, err := repository.GetExpressionByID(task.ExpressionID)
	if err != nil || expr == nil {
		http.Error(w, "expression not found", http.StatusInternalServerError)
		return
	}

	if task.Op == "FULL" {
		expr.Result = &resultFloat64
		expr.Status = model.StatusDone
		_ = repository.UpdateExpression(expr)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func getOperationTime(op string) int {
	switch op {
	case "+", "ADD":
		return additionTime
	case "-", "SUB":
		return subtractionTime
	case "*", "MUL":
		return multiplicationTime
	case "/", "DIV":
		return divisionTime
	case "FULL":
		return fullTime
	default:
		return additionTime
	}
}
