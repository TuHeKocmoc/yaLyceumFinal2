package repository_test

import (
	"os"
	"testing"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

func TestMain(m *testing.M) {
	os.Setenv("DB_PATH", ":memory:")
	if err := db.InitDB(); err != nil {
		panic("failed to init in-memory db: " + err.Error())
	}

	code := m.Run()

	os.Exit(code)
}

const testUserID int64 = 999

func TestCreateExpression(t *testing.T) {
	raw := "2+2"
	expr, err := repository.CreateExpression(raw, testUserID)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}
	if expr.ID == "" {
		t.Errorf("expected expr.ID not empty")
	}
	if expr.Raw != raw {
		t.Errorf("unexpected Raw, got %q, want %q", expr.Raw, raw)
	}
	if expr.UserID != testUserID {
		t.Errorf("expected UserID=%d, got %d", testUserID, expr.UserID)
	}
}

func TestCreateTaskWithArgs(t *testing.T) {
	expr, err := repository.CreateExpression("dummy", testUserID)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	val2 := 2.0
	t1, err := repository.CreateTaskWithArgs(expr.ID, "+", &val2, nil, &val2, nil)
	if err != nil {
		t.Fatalf("CreateTaskWithArgs error: %v", err)
	}
	if t1.ID <= 0 {
		t.Errorf("expected t1.ID > 0, got %d", t1.ID)
	}
	if t1.Arg1Value == nil || *t1.Arg1Value != 2.0 {
		t.Errorf("Arg1Value is not 2.0, got %v", t1.Arg1Value)
	}
	if t1.ExpressionID != expr.ID {
		t.Errorf("task ExpressionID mismatch, got %q want %q", t1.ExpressionID, expr.ID)
	}
}

func TestGetNextWaitingTask(t *testing.T) {
	expr, err := repository.CreateExpression("dummy", testUserID)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	val2 := 2.0
	t1, err := repository.CreateTaskWithArgs(expr.ID, "*", &val2, nil, &val2, nil)
	if err != nil {
		t.Fatalf("CreateTaskWithArgs t1 error: %v", err)
	}
	t2, err := repository.CreateTaskWithArgs(expr.ID, "+", nil, &t1.ID, &val2, nil)
	if err != nil {
		t.Fatalf("CreateTaskWithArgs t2 error: %v", err)
	}

	t.Logf("Created tasks: t1=%d, t2=%d", t1.ID, t2.ID)

	task, err := repository.GetNextWaitingTask()
	if err != nil {
		t.Fatalf("GetNextWaitingTask error: %v", err)
	}
	if task == nil {
		t.Fatal("expected a waiting task, got nil")
	}
	t.Logf("First waiting task: %d (status=%s)", task.ID, task.Status)

	if task.ID != t1.ID {
		t.Errorf("expected t1 to be returned first, got taskID=%d", task.ID)
	}

	t1.Status = model.TaskStatusDone
	if err := repository.UpdateTask(t1); err != nil {
		t.Fatalf("UpdateTask(t1) error: %v", err)
	}

	task2, err := repository.GetNextWaitingTask()
	if err != nil {
		t.Fatalf("GetNextWaitingTask error: %v", err)
	}
	if task2 == nil {
		t.Fatal("expected a second task, got nil")
	}
	t.Logf("Second waiting task: %d (status=%s)", task2.ID, task2.Status)

	if task2.ID != t2.ID {
		t.Errorf("expected t2, got taskID=%d", task2.ID)
	}
}
