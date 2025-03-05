package repository

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
)

var (
	mu sync.Mutex

	expressionsMap = make(map[string]*model.Expression)

	tasksMap = make(map[int]*model.Task)

	taskAutoID = 0
)

func CreateExpression(rawExpr string) (*model.Expression, error) {
	mu.Lock()
	defer mu.Unlock()

	id := uuid.New().String()

	expr := &model.Expression{
		ID:     id,
		Raw:    rawExpr,
		Status: model.StatusPending,
	}
	expressionsMap[id] = expr

	return expr, nil
}

func GetExpressionByID(id string) (*model.Expression, error) {
	mu.Lock()
	defer mu.Unlock()

	expr, ok := expressionsMap[id]
	if !ok {
		return nil, nil
	}
	return expr, nil
}

func GetAllExpressions() ([]*model.Expression, error) {
	mu.Lock()
	defer mu.Unlock()

	result := make([]*model.Expression, 0, len(expressionsMap))
	for _, expr := range expressionsMap {
		result = append(result, expr)
	}
	return result, nil
}

func UpdateExpression(e *model.Expression) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := expressionsMap[e.ID]; !ok {
		return errors.New("expression not found")
	}
	expressionsMap[e.ID] = e
	return nil
}

func CreateTask(expressionID string, op string, arg1, arg2 *float64) (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	expr, ok := expressionsMap[expressionID]
	if !ok {
		return nil, fmt.Errorf("no expression with id=%s", expressionID)
	}

	taskAutoID++
	taskID := taskAutoID

	task := &model.Task{
		ID:           taskID,
		ExpressionID: expressionID,
		Op:           op,
		Arg1:         arg1,
		Arg2:         arg2,
		Status:       model.TaskStatusWaiting,
	}
	tasksMap[taskID] = task

	expr.Tasks = append(expr.Tasks, taskID)
	expressionsMap[expressionID] = expr

	return task, nil
}

func GetTaskByID(id int) (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	t, ok := tasksMap[id]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func UpdateTask(t *model.Task) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := tasksMap[t.ID]; !ok {
		return errors.New("task not found")
	}
	tasksMap[t.ID] = t
	return nil
}

func GetNextWaitingTask() (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	for _, task := range tasksMap {
		if task.Status == model.TaskStatusWaiting {
			return task, nil
		}
	}
	return nil, nil
}
