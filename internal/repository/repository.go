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
		if task.Status != model.TaskStatusWaiting {
			continue
		}

		if task.Arg1TaskID != nil {
			depTask := tasksMap[*task.Arg1TaskID]
			if depTask == nil || depTask.Status != model.TaskStatusDone {
				continue
			}
		}

		if task.Arg2TaskID != nil {
			depTask := tasksMap[*task.Arg2TaskID]
			if depTask == nil || depTask.Status != model.TaskStatusDone {
				continue
			}
		}

		return task, nil
	}

	return nil, nil
}

func CreateTaskWithArgs(
	expressionID string,
	op string,
	arg1Value *float64, arg1TaskID *int,
	arg2Value *float64, arg2TaskID *int,
) (*model.Task, error) {

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

		Arg1Value:  arg1Value,
		Arg1TaskID: arg1TaskID,

		Arg2Value:  arg2Value,
		Arg2TaskID: arg2TaskID,

		Status: model.TaskStatusWaiting,
	}
	tasksMap[taskID] = task

	expr.Tasks = append(expr.Tasks, taskID)
	expressionsMap[expressionID] = expr

	return task, nil
}
