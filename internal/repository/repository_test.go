package repository_test

import (
	"os"
	"sync"
	"testing"
	"time"

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
	if err := repository.Reset(); err != nil {
		t.Fatalf("reset DB error: %v", err)
	}

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

func TestUpdateExpression_StatusAndResult(t *testing.T) {
	if err := repository.Reset(); err != nil {
		t.Fatalf("Reset error: %v", err)
	}

	expr, err := repository.CreateExpression("2+2", 123)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}
	if expr.ID == "" {
		t.Fatal("expr.ID is empty")
	}
	if expr.Status != model.StatusPending {
		t.Errorf("expected initial status=PENDING, got %s", expr.Status)
	}
	if expr.Result != nil {
		t.Errorf("expected initial result=nil, got %v", expr.Result)
	}

	expr.Status = model.StatusInProgress
	if err := repository.UpdateExpression(expr); err != nil {
		t.Fatalf("UpdateExpression error (IN_PROGRESS): %v", err)
	}

	expr2, err := repository.GetExpressionByID(expr.UserID, expr.ID)
	if err != nil {
		t.Fatalf("GetExpressionByID error: %v", err)
	}
	if expr2 == nil {
		t.Fatal("expr not found after update")
	}
	if expr2.Status != model.StatusInProgress {
		t.Errorf("expected status=IN_PROGRESS, got %s", expr2.Status)
	}
	if expr2.Result != nil {
		t.Errorf("expected result=nil, got %v", expr2.Result)
	}

	val := 4.0
	expr2.Result = &val
	expr2.Status = model.StatusDone
	if err := repository.UpdateExpression(expr2); err != nil {
		t.Fatalf("UpdateExpression error (DONE, result=4): %v", err)
	}

	expr3, err := repository.GetExpressionByID(expr2.UserID, expr2.ID)
	if err != nil {
		t.Fatalf("GetExpressionByID error: %v", err)
	}
	if expr3 == nil {
		t.Fatal("expr not found after second update")
	}
	if expr3.Status != model.StatusDone {
		t.Errorf("expected status=DONE, got %s", expr3.Status)
	}
	if expr3.Result == nil || *expr3.Result != 4.0 {
		t.Errorf("expected result=4.0, got %v", expr3.Result)
	}

	expr3.Result = nil
	if err := repository.UpdateExpression(expr3); err != nil {
		t.Fatalf("UpdateExpression error (result=nil): %v", err)
	}

	expr4, err := repository.GetExpressionByID(expr3.UserID, expr3.ID)
	if err != nil {
		t.Fatalf("GetExpressionByID error: %v", err)
	}
	if expr4 == nil {
		t.Fatal("expr not found after result=nil update")
	}

	if expr4.Result != nil {
		t.Errorf("expected result=nil, got %v", expr4.Result)
	}
	t.Logf("Expression ID=%s final status=%s, result=%v", expr4.ID, expr4.Status, expr4.Result)
}

func TestTasks_MultipleDependencies(t *testing.T) {
	if err := repository.Reset(); err != nil {
		t.Fatalf("Reset error: %v", err)
	}

	expr, err := repository.CreateExpression("dummy", 123)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	val := 1.0
	taskA, err := repository.CreateTaskWithArgs(expr.ID, "+", &val, nil, &val, nil)
	if err != nil {
		t.Fatalf("CreateTaskWithArgs(A) error: %v", err)
	}
	taskB, err := repository.CreateTaskWithArgs(expr.ID, "*", nil, &taskA.ID, &val, nil)
	if err != nil {
		t.Fatalf("CreateTaskWithArgs(B) error: %v", err)
	}
	taskC, err := repository.CreateTaskWithArgs(expr.ID, "-", nil, &taskB.ID, &val, nil)
	if err != nil {
		t.Fatalf("CreateTaskWithArgs(C) error: %v", err)
	}

	t.Logf("Created tasks: A=%d, B=%d, C=%d", taskA.ID, taskB.ID, taskC.ID)

	first, err := repository.GetNextWaitingTask()
	if err != nil {
		t.Fatalf("GetNextWaitingTask error: %v", err)
	}
	if first == nil {
		t.Fatal("expected a waiting task, got nil")
	}
	if first.ID != taskA.ID {
		t.Errorf("expected A=%d, got %d", taskA.ID, first.ID)
	}

	taskA.Status = model.TaskStatusDone
	if err := repository.UpdateTask(taskA); err != nil {
		t.Fatalf("UpdateTask(A) error: %v", err)
	}
	second, err := repository.GetNextWaitingTask()
	if err != nil {
		t.Fatalf("GetNextWaitingTask error: %v", err)
	}
	if second == nil {
		t.Fatal("expected a second task, got nil")
	}
	if second.ID != taskB.ID {
		t.Errorf("expected B=%d, got %d", taskB.ID, second.ID)
	}

	taskB.Status = model.TaskStatusDone
	if err := repository.UpdateTask(taskB); err != nil {
		t.Fatalf("UpdateTask(B) error: %v", err)
	}

	third, err := repository.GetNextWaitingTask()
	if err != nil {
		t.Fatalf("GetNextWaitingTask error: %v", err)
	}
	if third == nil {
		t.Fatal("expected a third task, got nil")
	}
	if third.ID != taskC.ID {
		t.Errorf("expected C=%d, got %d", taskC.ID, third.ID)
	}
}

func TestTasks_NoDependencies(t *testing.T) {
	if err := repository.Reset(); err != nil {
		t.Fatalf("Reset error: %v", err)
	}

	expr, err := repository.CreateExpression("dummy2", 999)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}
	val5 := 5.0
	task, err := repository.CreateTaskWithArgs(expr.ID, "+", &val5, nil, &val5, nil)
	if err != nil {
		t.Fatalf("CreateTaskWithArgs error: %v", err)
	}

	t.Logf("Created task ID=%d", task.ID)

	next, err := repository.GetNextWaitingTask()
	if err != nil {
		t.Fatalf("GetNextWaitingTask error: %v", err)
	}
	if next == nil {
		t.Fatal("expected a waiting task, got nil")
	}
	if next.ID != task.ID {
		t.Errorf("expected task ID=%d, got %d", task.ID, next.ID)
	}
}

func TestTasks_ConcurrentGetNextWaiting(t *testing.T) {
	if err := repository.Reset(); err != nil {
		t.Fatalf("Reset error: %v", err)
	}

	expr, err := repository.CreateExpression("dummy for concurrency", 123)
	if err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}

	numTasks := 10
	for i := 0; i < numTasks; i++ {
		val := float64(i)
		_, err := repository.CreateTaskWithArgs(expr.ID, "+", &val, nil, &val, nil)
		if err != nil {
			t.Fatalf("CreateTaskWithArgs error: %v", err)
		}
	}

	var wg sync.WaitGroup
	workers := 5
	doneTasksCh := make(chan int, numTasks)

	workerFunc := func(workerID int) {
		defer wg.Done()
		for {
			task, err := repository.GetNextWaitingTask()
			if err != nil {
				t.Logf("[Worker #%d] GetNextWaitingTask error: %v", workerID, err)
				time.Sleep(10 * time.Millisecond)
				continue
			}
			if task == nil {
				return
			}

			task.Status = model.TaskStatusDone
			if err := repository.UpdateTask(task); err != nil {
				t.Logf("[Worker #%d] UpdateTask error: %v", workerID, err)
				continue
			}

			doneTasksCh <- task.ID
			time.Sleep(5 * time.Millisecond)
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go workerFunc(i + 1)
	}

	wg.Wait()
	close(doneTasksCh)

	takenTasks := make(map[int]bool)
	for id := range doneTasksCh {
		if takenTasks[id] {
			t.Errorf("task ID=%d was taken more than once!", id)
		}
		takenTasks[id] = true
	}
	if len(takenTasks) != numTasks {
		t.Errorf("expected %d tasks done, got %d", numTasks, len(takenTasks))
	}

}
