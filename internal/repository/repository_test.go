package repository_test

import (
	"testing"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

func TestCreateExpression(t *testing.T) {
	repository.Reset()

	expr, err := repository.CreateExpression("2+2")
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}
	if expr.ID == "" {
		t.Errorf("expected expr.ID not empty")
	}
	if expr.Raw != "2+2" {
		t.Errorf("unexpected Raw, got %q", expr.Raw)
	}
}

func TestCreateTaskWithArgs(t *testing.T) {
	repository.Reset()

	expr, _ := repository.CreateExpression("dummy")
	val2 := 2.0
	t1, err := repository.CreateTaskWithArgs(expr.ID, "+", &val2, nil, &val2, nil)
	if err != nil {
		t.Fatalf("CreateTaskWithArgs error: %v", err)
	}
	if t1.ID <= 0 {
		t.Errorf("expected t1.ID > 0")
	}
	if t1.Arg1Value == nil || *t1.Arg1Value != 2.0 {
		t.Errorf("Arg1Value is not 2.0")
	}
}

func TestGetNextWaitingTask(t *testing.T) {
	repository.Reset()

	expr, _ := repository.CreateExpression("dummy")
	val2 := 2.0
	t1, _ := repository.CreateTaskWithArgs(expr.ID, "*", &val2, nil, &val2, nil)
	t2, _ := repository.CreateTaskWithArgs(expr.ID, "+", nil, &t1.ID, &val2, nil)

	t.Logf("Created tasks: t1=%d, t2=%d", t1.ID, t2.ID)
	// t1 - WAITING, t2 - WAITING, но у t2 Arg1TaskID=t1 => не READY
	task, err := repository.GetNextWaitingTask()
	if err != nil {
		t.Fatalf("GetNextWaitingTask error: %v", err)
	}

	if task == nil {
		t.Fatal("expected a waiting task")
	}
	t.Logf("First waiting task: %d (status=%s)", task.ID, task.Status)

	if task.ID != t1.ID {
		t.Errorf("expected t1 to be returned first, got taskID=%d", task.ID)
	}
	// представим, что мы помечаем t1 DONE
	t1.Status = model.TaskStatusDone
	_ = repository.UpdateTask(t1)

	task2, _ := repository.GetNextWaitingTask()
	if task2 == nil {
		t.Fatal("expected a second task")
	}
	t.Logf("Second waiting task: %d (status=%s)", task2.ID, task2.Status)
	if task2.ID != t2.ID {
		t.Errorf("expected t2, got %d", task2.ID)
	}
}
