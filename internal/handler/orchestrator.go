package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/calc"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/planner"
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

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req requestExpression
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] expression = %q", req.Expression)

	if !calc.CheckInput(req.Expression) {
		http.Error(w, "expression is not valid", http.StatusUnprocessableEntity)
		return
	}

	expr, err := repository.CreateExpression(req.Expression, userID)
	if err != nil {
		http.Error(w, "cannot create expression", http.StatusInternalServerError)
		return
	}

	if expr.Raw != "" {
		finalTaskID, err := planner.PlanTasksWithNestedParen(expr.ID, expr.Raw)
		if err != nil {
			expr.Status = model.StatusError
			_ = repository.UpdateExpression(expr)
			http.Error(w, "cannot plan tasks: "+err.Error(), http.StatusUnprocessableEntity)
			log.Printf("[DEBUG] PlanTasksWithNestedParen error: %v", err)
			return
		}

		expr.Status = model.StatusInProgress
		expr.FinalTaskID = finalTaskID
		_ = repository.UpdateExpression(expr)
	}

	resp := responseCreateExpression{ID: expr.ID}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

type responseExpressionsList struct {
	Expressions []*model.Expression `json:"expressions"`
}

func HandleGetAllExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	exprs, err := repository.GetAllExpressions(userID)
	if err != nil {
		http.Error(w, "failed to get expressions", http.StatusInternalServerError)
		log.Printf("[DEBUG] GetAllExpressions error: %v", err)
		return
	}

	resp := responseExpressionsList{Expressions: exprs}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type responseSingleExpression struct {
	Expression *model.Expression `json:"expression"`
}

func HandleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	id := parts[3]

	expr, err := repository.GetExpressionByID(userID, id)
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

type responseTask struct {
	Task struct {
		ID            int         `json:"id"`
		Arg1          interface{} `json:"arg1"`
		Arg2          interface{} `json:"arg2"`
		Operation     string      `json:"operation"`
		OperationTime int         `json:"operation_time"`
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
	if task.Op == "FULL" {
		expr, err := repository.GetExpressionByIDForTask(task.ExpressionID)
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
	} else {
		arg1 = fetchArgumentValue(task.Arg1Value, task.Arg1TaskID)
		arg2 = fetchArgumentValue(task.Arg2Value, task.Arg2TaskID)
	}

	var resp responseTask
	resp.Task.ID = task.ID
	resp.Task.Arg1 = arg1
	resp.Task.Arg2 = arg2
	resp.Task.Operation = task.Op
	resp.Task.OperationTime = operationTime

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func fetchArgumentValue(argValue *float64, argTaskID *int) interface{} {
	if argValue != nil {
		return *argValue
	}
	if argTaskID != nil {
		depTask, _ := repository.GetTaskByID(*argTaskID)
		if depTask != nil && depTask.Result != nil {
			return *depTask.Result
		}
	}
	return 0.0
}

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
	taskID := int(idInt64)

	task, err := repository.GetTaskByID(taskID)
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

	expr, err := repository.GetExpressionByIDForTask(task.ExpressionID)
	if err != nil || expr == nil {
		http.Error(w, "expression not found", http.StatusInternalServerError)
		return
	}

	if task.Op == "FULL" {
		expr.Result = &resultFloat64
		expr.Status = model.StatusDone
		_ = repository.UpdateExpression(expr)
	} else {
		done, lastTaskResult, err := checkAllTasksDone(expr.ID)
		if err != nil {
			http.Error(w, "cannot check tasks", http.StatusInternalServerError)
			return
		}
		if done {
			expr.Result = lastTaskResult
			expr.Status = model.StatusDone
			if err := repository.UpdateExpression(expr); err != nil {
				http.Error(w, "failed to update expression", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func checkAllTasksDone(exprID string) (bool, *float64, error) {
	tasks, err := repository.GetTasksByExpressionID(exprID)
	if err != nil {
		return false, nil, err
	}
	if len(tasks) == 0 {
		return true, nil, nil
	}

	allDone := true
	var lastResult *float64
	var lastTaskID int

	for _, t := range tasks {
		if t.Status != model.TaskStatusDone {
			allDone = false
		}
		if t.ID > lastTaskID && t.Result != nil {
			lastTaskID = t.ID
			lastResult = t.Result
		}
	}
	return allDone, lastResult, nil
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
